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

package config

import (
	"context"
	"fmt"
	"reflect"
	"time"

	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	oconfig "github.com/IBM/ibm-block-csi-operator/pkg/config"
	configsyncer "github.com/IBM/ibm-block-csi-operator/pkg/controller/config/syncer"
	operatorconfig "github.com/IBM/ibm-block-csi-operator/pkg/internal/config"
	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ReconcileTime is the delay between reconciliations
const ReconcileTime = 30 * time.Second

var log = logf.Log.WithName("config_controller")

// Add creates a new Config Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {

	if !controllerutil.IsDefineHostEnabled(mgr.GetClient()) {
		log.Info("Skip config_controller")
		return nil
	}

	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileConfig{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetRecorder("controller_config"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("config-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Config
	err = c.Watch(&source.Kind{Type: &csiv1.Config{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	subresources := []runtime.Object{
		&appsv1.DaemonSet{},
	}

	for _, subresource := range subresources {
		err = c.Watch(&source.Kind{Type: subresource}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &csiv1.Config{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileConfig implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileConfig{}

// ReconcileConfig reconciles a Config object
type ReconcileConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a Config object and makes changes based on the state read
// and what is in the Config.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Config")

	// Fetch the Config instance
	instance := operatorconfig.New(&csiv1.Config{})
	err := r.client.Get(context.TODO(), request.NamespacedName, instance.Unwrap())
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

	r.scheme.Default(instance.Unwrap())
	changed := instance.SetDefaults()

	if err := instance.Validate(); err != nil {
		err = fmt.Errorf("wrong Config options: %v", err)
		return reconcile.Result{RequeueAfter: ReconcileTime}, err
	}

	// update CR if there was changes after defaulting
	if changed {
		err = r.client.Update(context.TODO(), instance.Unwrap())
		if err != nil {
			err = fmt.Errorf("failed to update Config: %v", err)
			return reconcile.Result{}, err
		}
	}

	status := *instance.Status.DeepCopy()
	defer func() {
		if !reflect.DeepEqual(status, instance.Status) {
			reqLogger.Info("updating Config status", "name", instance.Name)
			sErr := r.client.Status().Update(context.TODO(), instance.Unwrap())
			if sErr != nil {
				reqLogger.Error(sErr, "failed to update Config status", "name", instance.Name)
			}
		}
	}()

	// sync the node agent only if defineHost is enabled.
	if instance.Spec.DefineHost {
		// sync the resources which change over time
		reqLogger.Info("Reconciling node agent")
		nodeAgentSyncer := configsyncer.NewNodeAgentSyncer(r.client, r.scheme, instance)
		if err := syncer.Sync(context.TODO(), nodeAgentSyncer, r.recorder); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Reconciled node agent")
	}

	if err := r.updateStatus(instance); err != nil {
		return reconcile.Result{}, err
	}

	// Resource created successfully - don't requeue
	return reconcile.Result{}, nil
}

func (r *ReconcileConfig) updateStatus(instance *operatorconfig.Config) error {
	if !instance.Spec.DefineHost {
		instance.Status.NodeAgent.Phase = csiv1.NodeAgentPhaseDisabled
		return nil
	}
	nodeAgent := &appsv1.DaemonSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      oconfig.GetNameForResource(oconfig.NodeAgent, instance.Name),
		Namespace: instance.Namespace,
	}, nodeAgent)

	if err != nil {
		return err
	}

	phase := csiv1.NodeAgentPhaseCreating
	if nodeAgent.Status.DesiredNumberScheduled == nodeAgent.Status.NumberAvailable {
		phase = csiv1.NodeAgentPhaseRunning
	}
	instance.Status.NodeAgent.Phase = phase

	// no need to push to status to API Server here.
	return nil
}
