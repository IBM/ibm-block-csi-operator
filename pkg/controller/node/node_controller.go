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
	"os"
	"time"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/controller/predicate"
	controllerutil "github.com/IBM/ibm-block-csi-operator/pkg/controller/util"
	nodeclient "github.com/IBM/ibm-block-csi-operator/pkg/node/client"
	pb "github.com/IBM/ibm-block-csi-operator/pkg/node/nodeagent"
	"github.com/IBM/ibm-block-csi-operator/pkg/util"
)

var log = logf.Log.WithName("node_controller")

// Add creates a new Node Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {

	if !controllerutil.IsDefineHostEnabled(mgr.GetClient()) {
		log.Info("Skip node_controller")
		return nil
	}

	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNode{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("node-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for node create and delete event.
	err = c.Watch(
		&source.Kind{Type: &corev1.Node{}},
		&handler.EnqueueRequestForObject{},
		predicate.NodePredicate{},
	)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileNode implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNode{}

// ReconcileNode reconciles a Node object
type ReconcileNode struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Node object and update the NodeInfo accordingly.
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNode) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Name", request.Name)
	reqLogger.Info("Reconciling Node")

	// Fetch the Node instance
	node := &corev1.Node{}
	err := r.client.Get(context.TODO(), request.NamespacedName, node)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if isMasterNode(node) {
		// skip master node
		return reconcile.Result{}, nil
	}

	if !controllerutil.IsDefineHostEnabled(r.client) {
		reqLogger.Info("Skip reconciling Node")
		return reconcile.Result{}, nil
	}

	if !controllerutil.IsNodeAgentReady(r.client) {
		reqLogger.Info("Node Agent is not ready, try it later")
		return reconcile.Result{RequeueAfter: 4 * time.Second}, nil
	}

	err = r.processNodeInfo(node)
	if err != nil {
		reqLogger.Error(err, "failed to processNodeInfo")
		// don't add back to the queue immediately
		return reconcile.Result{RequeueAfter: time.Minute}, err
	}

	reqLogger.Info("Reconciled Node")
	return reconcile.Result{}, nil
}

func (r *ReconcileNode) processNodeInfo(node *corev1.Node) error {
	log.Info("Processing NodeInfo", "name", node.GetName())
	nodeInfoPB, err := getNodeInfoFromNode(node)
	if err != nil {
		return err
	}

	log.Info("Getting NodeInfo", "Name", node.GetName())
	found := &csiv1.NodeInfo{}
	created := false
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      node.GetName(),
		Namespace: "", // it is a cluster scope resource
	}, found)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new NodeInfo", "Name", node.GetName())
		nodeInfo := newNodeInfo(node.GetName())
		err = r.client.Create(context.TODO(), nodeInfo)
		if err != nil {
			log.Error(err, "Failed to create NodeInfo", "Name", node.GetName())
			return err
		}
		created = true
	} else if err != nil {
		log.Error(err, "Failed to get NodeInfo", "Name", node.GetName())
		return err
	}
	// update status
	if created {
		// get again after creation
		// retry in case the new object is not ready yet.
		var err error
		for i := 0; i < 3; i++ {
			err = r.client.Get(context.TODO(), types.NamespacedName{
				Name:      node.GetName(),
				Namespace: "", // it is a cluster scope resource
			}, found)
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}
		if err != nil {
			log.Error(err, "Failed to get NodeInfo after creation", "Name", node.GetName())
			return err
		}

	}

	iqns := nodeInfoPB.GetIqns()
	if iqns == nil {
		iqns = []string{}
	}
	wwpns := nodeInfoPB.GetWwpns()
	if wwpns == nil {
		wwpns = []string{}
	}
	definedOnStorages := found.Status.DefinedOnStorages
	if definedOnStorages == nil {
		definedOnStorages = []string{}
	}

	found.Status.Iqns = iqns
	found.Status.Wwpns = wwpns
	found.Status.DefinedOnStorages = definedOnStorages
	log.Info("Updating NodeInfo", "Name", node.GetName())
	err = r.client.Status().Update(context.TODO(), found)
	if err != nil {
		log.Error(err, "Failed to update NodeInfo", "Name", node.GetName())
		return err
	}
	log.Info("Processed NodeInfo", "name", node.GetName())
	return nil
}

func getNodeInfoFromNode(node *corev1.Node) (*pb.Node, error) {
	log.Info("Getting NodeInfo from node", "name", node.GetName())
	addrs := util.GetNodeAddresses(node)

	port := os.Getenv(config.ENVIscsiAgentPort)
	if port == "" {
		return nil, errors.Errorf("ENV %s is not set", config.ENVIscsiAgentPort)
	}

	log.Info("Checking if node accessable", "addresses", addrs)
	addr := util.TestConnectivity(addrs, port)
	if addr == "" {
		return nil, errors.New("No node address is available to connect")
	}

	c := nodeclient.NewNodeClient(addr+":"+port, log)
	return c.GetNodeInfo(node.GetName())
}

func newNodeInfo(nodeName string) *csiv1.NodeInfo {
	return &csiv1.NodeInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
		},
	}
}

func isMasterNode(node *corev1.Node) bool {
	labels := node.GetLabels()
	if labels == nil {
		return false
	}
	_, ok := labels[config.Masterlabel]
	return ok
}
