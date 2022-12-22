/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/IBM/volume-group-operator/controllers/utils"
	"github.com/IBM/volume-group-operator/pkg/config"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	grpcClient "github.com/IBM/volume-group-operator/pkg/client"
)

const (
	VolumeGroup               = "VolumeGroup"
	VolumeGroupClass          = "VolumeGroupClass"
	VolumeGroupContent        = "VolumeGroupContent"
)

// VolumeGroupReconciler reconciles a VolumeGroup object
type VolumeGroupReconciler struct {
	client.Client
	Utils             utils.ControllerUtils
	Log               logr.Logger
	Scheme            *runtime.Scheme
	DriverConfig      *config.DriverConfig
	GRPCClient        *grpcClient.Client
	VolumeGroupClient grpcClient.VolumeGroup
}

//+kubebuilder:rbac:groups=csi.ibm.com,resources=volumegroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csi.ibm.com,resources=volumegroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csi.ibm.com,resources=volumegroups/finalizers,verbs=update
//+kubebuilder:rbac:groups=csi.ibm.com,resources=volumegroupclasses,verbs=get;list;watch
//+kubebuilder:rbac:groups=csi.ibm.com,resources=volumegroupcontents,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims/finalizers,verbs=update

func (r *VolumeGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeGroupReconciler) SetupWithManager(mgr ctrl.Manager, cfg *config.DriverConfig) error {
	logger := r.Log.WithName("SetupWithManager")
	err := r.waitForCrds(logger)
	if err != nil {
		r.Log.Error(err, "failed to wait for crds")

		return err
	}
	pred := predicate.GenerationChangedPredicate{}

	r.VolumeGroupClient = grpcClient.NewVolumeGroupClient(r.GRPCClient.Client, cfg.RPCTimeout)

	return ctrl.NewControllerManagedBy(mgr).
		For(&volumegroupv1.VolumeGroup{}).
		WithEventFilter(pred).Complete(r)
}

func (r *VolumeGroupReconciler) waitForCrds(logger logr.Logger) error {
	err := r.waitForVolumeGroupResource(logger, VolumeGroup)
	if err != nil {
		logger.Error(err, "failed to wait for VolumeGroup CRD")

		return err
	}

	err = r.waitForVolumeGroupResource(logger, VolumeGroupClass)
	if err != nil {
		logger.Error(err, "failed to wait for VolumeGroupClass CRD")

		return err
	}

	err = r.waitForVolumeGroupResource(logger, VolumeGroupContent)
	if err != nil {
		logger.Error(err, "failed to wait for VolumeGroupContent CRD")

		return err
	}

	return nil
}

func (r *VolumeGroupReconciler) waitForVolumeGroupResource(logger logr.Logger, resourceName string) error {
	unstructuredResource := &unstructured.UnstructuredList{}
	unstructuredResource.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   volumegroupv1.GroupVersion.Group,
		Kind:    resourceName,
		Version: volumegroupv1.GroupVersion.Version,
	})
	for {
		err := r.Client.List(context.TODO(), unstructuredResource)
		if err == nil {
			return nil
		}
		// return errors other than NoMatch
		if !meta.IsNoMatchError(err) {
			logger.Error(err, "got an unexpected error while waiting for resource", "Resource", resourceName)

			return err
		}
		logger.Info("resource does not exist", "Resource", resourceName)
		time.Sleep(5 * time.Second)
	}
}
