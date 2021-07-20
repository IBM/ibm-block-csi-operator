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

package volumeattachment

import (
	"context"
	"os"

	"github.com/pkg/errors"

	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/node/client"
	"github.com/IBM/ibm-block-csi-operator/pkg/storageagent"
	"github.com/IBM/ibm-block-csi-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileVolumeAttachment) getNodeAddresses(nodeName string) ([]string, error) {
	log.Info("Retrieving Node info")

	node := &corev1.Node{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      nodeName,
		Namespace: "",
	}, node)
	if err != nil {
		log.Error(err, "Failed to get Node", "name", nodeName)
		return nil, err
	}
	addrs := util.GetNodeAddresses(node)

	log.Info("Found node addresses", "addresses", addrs)
	return addrs, nil
}

func (r *ReconcileVolumeAttachment) loginIscsiTargets(arrayAddr, user, password, nodeName string) error {
	log.Info("Starting to login iscsi targets")
	addrs, err := r.getNodeAddresses(nodeName)
	if err != nil {
		return err
	}

	port := os.Getenv(config.ENVIscsiAgentPort)
	if port == "" {
		return errors.Errorf("env %s is not set", config.ENVIscsiAgentPort)
	}

	log.Info("Checking if node accessable", "addresses", addrs)
	addr := util.TestConnectivity(addrs, port)
	if addr == "" {
		return errors.New("No node address is available to connect")
	}

	targets, err := getIscsiTargetsFromArray(arrayAddr, user, password)
	if err != nil {
		return err
	}

	c := client.NewNodeClient(addr+":"+port, log)
	return c.IscsiLogin(targets)
}

func getIscsiTargetsFromArray(arrayAddr, user, password string) ([]string, error) {
	log.Info("Retrieving iscsi targets from array", "array", arrayAddr)
	c := storageagent.NewStorageClient(arrayAddr, user, password, log)
	targets, err := c.ListIscsiTargets()
	if err != nil {
		log.Error(err, "Failed to get iscsi targets from array")
		return nil, err
	}

	ts := []string{}
	for _, t := range targets {
		ts = append(ts, t.Address)
	}
	log.Info("Found iscsi targets", "targets", ts)
	return ts, nil
}
