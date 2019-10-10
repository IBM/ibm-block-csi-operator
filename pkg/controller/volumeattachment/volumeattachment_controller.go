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
	"time"

	// b64 "encoding/base64"
	"fmt"

	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	"github.com/IBM/ibm-block-csi-operator/pkg/controller/predicate"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	controllerutil "github.com/IBM/ibm-block-csi-operator/pkg/controller/util"
	"github.com/IBM/ibm-block-csi-operator/pkg/storageagent"
)

var log = logf.Log.WithName("volumeattachment_controller")

// Add creates a new VolumeAttachment Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {

	if !controllerutil.IsDefineHostEnabled(mgr.GetClient()) {
		log.Info("Skip volumeattachment_controller")
		return nil
	}

	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileVolumeAttachment{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("volumeattachment-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource VolumeAttachment
	err = c.Watch(
		&source.Kind{Type: &storagev1.VolumeAttachment{}},
		&handler.EnqueueRequestForObject{},
		predicate.CreatePredicate{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileVolumeAttachment implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileVolumeAttachment{}

// ReconcileVolumeAttachment reconciles a VolumeAttachment object
type ReconcileVolumeAttachment struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a VolumeAttachment object and makes changes based on the state read
// and what is in the VolumeAttachment.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileVolumeAttachment) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling VolumeAttachment")

	// Fetch the VolumeAttachment instance
	volAtt := &storagev1.VolumeAttachment{}
	err := r.client.Get(context.TODO(), request.NamespacedName, volAtt)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if volAtt.Spec.Attacher != config.DriverName {
		reqLogger.Info("Not managed by current driver, skip")
		return reconcile.Result{}, nil
	}

	if !controllerutil.IsDefineHostEnabled(r.client) {
		reqLogger.Info("Skip reconciling VolumeAttachment")
		return reconcile.Result{}, nil
	}

	err = r.processVolumeAttachment(volAtt)
	if err != nil {
		reqLogger.Error(err, "failed to processVolumeAttachment", "Namespace", volAtt.Namespace, "Name", volAtt.Name)
		// don't add back to the queue immediately
		return reconcile.Result{RequeueAfter: time.Minute}, err
	}

	reqLogger.Info("Reconciled VolumeAttachment")
	return reconcile.Result{}, nil
}

func (r *ReconcileVolumeAttachment) processVolumeAttachment(volAtt *storagev1.VolumeAttachment) error {
	vaLogger := log.WithValues("Namespace", volAtt.Namespace, "Name", volAtt.Name)
	vaLogger.Info("Processing VolumeAttachment")

	// pvName is a string pointer
	pvName := volAtt.Spec.Source.PersistentVolumeName
	if pvName == nil || *pvName == "" {
		return fmt.Errorf("pv name not found in VolumeAttachment %s", volAtt.Name)
	}

	nodeName := volAtt.Spec.NodeName
	pv := &corev1.PersistentVolume{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      *pvName,
		Namespace: "",
	}, pv)
	if err != nil {
		if errors.IsNotFound(err) {
			vaLogger.Info("PersistentVolume not found", "Name", *pvName)
			return nil
		}
		return err
	}
	err = r.processPersistentVolume(pv, nodeName)
	if err != nil {
		return err
	}

	vaLogger.Info("Processed VolumeAttachment")
	return nil
}

func (r *ReconcileVolumeAttachment) processPersistentVolume(pv *corev1.PersistentVolume, nodeName string) error {
	pvLogger := log.WithValues("pvName", pv.Name)
	pvLogger.Info("Processing PersistentVolume")

	controllerPublishSecretRef := pv.Spec.CSI.ControllerPublishSecretRef
	if controllerPublishSecretRef == nil {
		return fmt.Errorf("controllerPublishSecretRef not found in PersistentVolume %s", pv.Name)
	}

	secret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      controllerPublishSecretRef.Name,
		Namespace: controllerPublishSecretRef.Namespace,
	}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			pvLogger.Info(
				"controllerPublishSecret not found",
				"Namespace", controllerPublishSecretRef.Namespace,
				"Name", controllerPublishSecretRef.Name)
			return nil
		}
		return err
	}
	err = r.processControllerPublishSecret(secret, nodeName)
	if err != nil {
		return err
	}

	pvLogger.Info("Processed PersistentVolume")
	return nil
}

func (r *ReconcileVolumeAttachment) processControllerPublishSecret(secret *corev1.Secret, nodeName string) error {
	sLogger := log.WithValues("secretNamespace", secret.Namespace, "secretName", secret.Name)
	sLogger.Info("Processing ControllerPublishSecret")

	nodeInfo := &csiv1.NodeInfo{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      nodeName,
		Namespace: "",
	}, nodeInfo)
	if err != nil {
		if errors.IsNotFound(err) {
			sLogger.Info("nodeInfo not found", "nodeName", nodeName)
			return nil
		}
		return err
	}

	//	arrayAddr, err := b64.StdEncoding.DecodeString(string(secret.Data["management_address"]))
	//	if err != nil {
	//		sLogger.Error(err, "Failed to decode storage address from secret")
	//		return err
	//	}
	//	username, err := b64.StdEncoding.DecodeString(string(secret.Data["username"]))
	//	if err != nil {
	//		sLogger.Error(err, "Failed to decode username from secret")
	//		return err
	//	}
	//	password, err := b64.StdEncoding.DecodeString(string(secret.Data["password"]))
	//	if err != nil {
	//		sLogger.Error(err, "Failed to decode password from secret")
	//		return err
	//	}

	arrayAddr := string(secret.Data["management_address"])

	definedArrays := nodeInfo.Status.DefinedOnStorages
	if definedArrays == nil {
		definedArrays = []string{}
	}
	for _, array := range definedArrays {
		if string(arrayAddr) == array {
			sLogger.Info("Host is already defined on storage", "Host", nodeName, "Storage", arrayAddr)
			return nil
		}
	}

	err = defineHostOnArray(
		string(arrayAddr),
		string(secret.Data["username"]),
		string(secret.Data["password"]),
		nodeName, nodeInfo.Status.Iqns, nodeInfo.Status.Wwpns)
	if err != nil {
		sLogger.Error(err, "Failed to define host on storage")
		return err
	}

	// skip iscsi login if it is a fc host
	if len(nodeInfo.Status.Wwpns) == 0 {
		err = r.loginIscsiTargets(
			string(arrayAddr),
			string(secret.Data["username"]),
			string(secret.Data["password"]),
			nodeName,
		)
		if err != nil {
			sLogger.Error(err, "Failed to login iscsi targets")
			return err
		}
	}

	definedArrays = append(definedArrays, string(arrayAddr))
	nodeInfo.Status.DefinedOnStorages = definedArrays

	sLogger.Info("Updating NodeInfo", "nodeName", nodeName)
	err = r.client.Status().Update(context.TODO(), nodeInfo)
	if err != nil {
		sLogger.Error(err, "Failed to update NodeInfo", "nodeName", nodeName)
		return err
	}
	sLogger.Info("Processed ControllerPublishSecret")
	return nil
}

func defineHostOnArray(arrayAddr, user, password, nodeName string, iscsiPorts, fcPorts []string) error {
	//client := arrayactions.NewSvcMediator(arrayAddr, user, password, log)
	client := storageagent.NewStorageClient(arrayAddr, user, password, log)
	return client.CreateHost(nodeName, iscsiPorts, fcPorts)
}
