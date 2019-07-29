/**
 * Copyright 2019 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package node

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	csiv1 "github.com/IBM/ibm-block-csi-driver-operator/pkg/apis/csi/v1"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/config"
)

var log = logf.Log.WithName("node_syncer")

func Sync(c client.Client) {
	log.Info("Start to sync node info")
	hostname, err := getHostname()
	if err != nil {
		log.Error(err, "Failed to get hostname of the node")
		return
	}

	updateIscsi := false
	updateFc := false
	iqns, err := GetNodeIscsiIQNs()
	if err == nil {
		updateIscsi = true
		log.Info("Got iscsi initiators", "iqns", iqns)
	} else {
		log.Info("Iscsi initiator is not configured well", "err", err.Error())
	}
	wwpns, err := GetNodeFcWWPNs()
	if err == nil {
		updateFc = true
		log.Info("Got fc ports", "wwpns", wwpns)
	} else {
		log.Info("Fc port is not configured well", "err", err.Error())
	}

	// don't use a csiv1.NodeInfo here,because typed-client always get records
	// from cache, while in node-syncer, we never load the cache.
	found := &unstructured.Unstructured{}
	found.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   config.APIGroup,
		Kind:    reflect.TypeOf(csiv1.NodeInfo{}).Name(),
		Version: config.APIVersion,
	})

	created := false
	err = c.Get(context.TODO(), types.NamespacedName{
		Name:      hostname,
		Namespace: "", // it is a cluster scope resource
	}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new NodeInfo", "Name", hostname)
		nodeInfo := newNodeInfo(hostname)
		err = c.Create(context.TODO(), nodeInfo)
		if err != nil {
			log.Error(err, "Failed to create NodeInfo", "Name", hostname)
			return
		}
		created = true
	} else if err != nil {
		log.Error(err, "Failed to get NodeInfo", "Name", hostname)
		return
	}
	// update status
	if created {
		// get again after creation
		err := c.Get(context.TODO(), types.NamespacedName{
			Name:      hostname,
			Namespace: "", // it is a cluster scope resource
		}, found)
		if err != nil {
			log.Error(err, "Failed to get NodeInfo after creation", "Name", hostname)
			return
		}
	}

	if updateIscsi {
		//found.Status.Iqns = iqns
		unstructured.SetNestedStringSlice(found.Object, iqns, "status", "iqns")
	}
	if updateFc {
		//found.Status.Wwpns = wwpns
		unstructured.SetNestedStringSlice(found.Object, wwpns, "status", "wwpns")
	}
	log.Info("Updating NodeInfo", "Name", hostname)
	err = c.Status().Update(context.TODO(), found)
	if err != nil {
		log.Error(err, "Failed to update NodeInfo", "Name", hostname)
		return
	}

	log.Info("Finished to sync node info")
}

func newNodeInfo(hostname string) *csiv1.NodeInfo {
	return &csiv1.NodeInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name: hostname,
		},
	}
}

func getHostname() (string, error) {
	name := os.Getenv("NODE_NAME")
	if name == "" {
		return "", fmt.Errorf("ENV NODE_NAME is not set")
	}
	return name, nil
}
