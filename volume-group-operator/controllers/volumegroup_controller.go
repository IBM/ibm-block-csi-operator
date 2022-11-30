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
	volumegroup "github.com/IBM/volume-group-operator/controllers/volumegroup"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/IBM/volume-group-operator/pkg/config"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VolumeGroupReconciler reconciles a VolumeGroup object
type VolumeGroupReconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	DriverConfig *config.DriverConfig
}

//+kubebuilder:rbac:groups=csi.ibm.com,resources=volumegroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csi.ibm.com,resources=volumegroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csi.ibm.com,resources=volumegroups/finalizers,verbs=update

func (r *VolumeGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("Request.Name", req.Name, "Request.Namespace", req.Namespace)

	instance := &volumegroupv1.VolumeGroup{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("VolumeGroup resource not found")

			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	vgcObj, err := r.getVolumeGroupClass(logger, *instance.Spec.VolumeGroupClassName)
	if err != nil {
		setFailureCondition(instance)
		uErr := r.updateVolumeGroupStatus(instance, logger, err.Error())
		if uErr != nil {
			logger.Error(uErr, "failed to update volumeReplication status", "VGName", instance.Name)
		}

		return ctrl.Result{}, err
	}

	if r.DriverConfig.DriverName != vgcObj.Driver {
		return ctrl.Result{}, nil
	}

	parameters := filterPrefixedParameters(volumeGroupParameterPrefix, vgcObj.Parameters)

	secretName := vgcObj.Parameters[prefixedVolumeGroupSecretNameKey]
	secretNamespace := vgcObj.Parameters[prefixedVolumeGroupSecretNamespaceKey]
	secret := make(map[string]string)
	if secretName != "" && secretNamespace != "" {
		secret, err = r.getSecret(logger, secretName, secretNamespace)
		if err != nil {
			setFailureCondition(instance)
			uErr := r.updateVolumeGroupStatus(instance, logger, err.Error())
			if uErr != nil {
				logger.Error(uErr, "failed to update volumeReplication status", "VGName", instance.Name)
			}

			return reconcile.Result{}, err
		}
	}

	var pvc *corev1.PersistentVolumeClaim

	volumeGroupContentSource, _ := r.getVolumeGroupContentSource(logger, req.NamespacedName)
	volumeGroupId := volumeGroupContentSource.VolumeGroupHandle

	// check if the object is being deleted
	if instance.GetDeletionTimestamp().IsZero() {
		if err = r.addFinalizerToVG(logger, instance); err != nil {
			logger.Error(err, "Failed to add VolumeReplication finalizer")

			return reconcile.Result{}, err
		}
		if err = r.addFinalizerToPVC(logger, pvc); err != nil {
			logger.Error(err, "Failed to add PersistentVolumeClaim finalizer")

			return reconcile.Result{}, err
		}
	} else {
		if contains(instance.GetFinalizers(), volumeGroupFinalizer) {
			if err = r.deleteVolumeGroup(logger, volumeGroupId, secret); err != nil {
				logger.Error(err, "failed to delete volume group")

				return ctrl.Result{}, err
			}
			if err = r.removeFinalizerFromPVC(logger, pvc); err != nil {
				logger.Error(err, "Failed to remove PersistentVolumeClaim finalizer")

				return reconcile.Result{}, err
			}

			// once all finalizers have been removed, the object will be
			// deleted
			if err = r.removeFinalizerFromVG(logger, instance); err != nil {
				logger.Error(err, "Failed to remove volume group finalizer")

				return reconcile.Result{}, err
			}
		}
		logger.Info("volumeGroup object is terminated, skipping reconciliation")

		return ctrl.Result{}, nil
	}

	groupCreationTime := getCurrentTime()
	instance.Status.GroupCreationTime = groupCreationTime
	if err = r.Client.Update(context.TODO(), instance); err != nil {
		logger.Error(err, "failed to update status")

		return reconcile.Result{}, err
	}
	volumeGroupName := "" //TODO
	// create volume group replication on every reconcile
	resp := r.createVolumeGroup(logger, volumeGroupName, parameters, secret)
	if resp.Error != nil {
		logger.Error(err, "failed to create volume group")
		setFailureCondition(instance)
		msg := volumegroup.GetMessageFromError(resp.Error)
		uErr := r.updateVolumeGroupStatus(instance, logger, msg)
		if uErr != nil {
			logger.Error(uErr, "failed to update volumeGroup status", "VGName", instance.Name)
		}

		return reconcile.Result{}, err
	}
	ready := true

	if err = r.createVolumeGroupContent(logger, instance, vgcObj, resp, secretName, secretNamespace, groupCreationTime, &ready); err != nil {
		logger.Error(err, "failed to update volumeGroup status", "VGName", instance.Name)
		return reconcile.Result{}, err
	}
	//TODO add all PVCs that have the VG label to VG

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&volumegroupv1.VolumeGroup{}).
		Complete(r)
}

func (r *VolumeGroupReconciler) updateVolumeGroupStatus(
	instance *volumegroupv1.VolumeGroup,
	logger logr.Logger,
	message string) error {
	instance.Status.Error.Message = &message
	if err := r.Client.Status().Update(context.TODO(), instance); err != nil {
		logger.Error(err, "failed to update status")

		return err
	}

	return nil
}

// deleteVolumeGroup defines and runs a set of tasks required to delete volume group.
func (r *VolumeGroupReconciler) deleteVolumeGroup(logger logr.Logger, volumeGroupId string, secrets map[string]string) error {
	c := volumegroup.CommonRequestParameters{
		VolumeGroupID: volumeGroupId,
		Secrets:       secrets,
	}

	volumeReplication := volumegroup.VolumeGroup{
		Params: c,
	}

	resp := volumeReplication.Delete()

	if resp.Error != nil {
		//if isKnownError := resp.HasKnownGRPCError(disableReplicationKnownErrors); isKnownError {
		//	logger.Info("volume not found", "volumeID", volumeID)
		//
		//	return nil
		//}
		//logger.Error(resp.Error, "failed to disable volume replication")

		return resp.Error
	}

	return nil
}

// createVolumeGroup defines and runs a set of tasks required to delete volume group.
func (r *VolumeGroupReconciler) createVolumeGroup(logger logr.Logger, volumeGroupName string,
	parameters, secrets map[string]string) *csi.VolumeGroupResponse {
	c := volumegroup.CommonRequestParameters{
		Name:       volumeGroupName,
		Parameters: parameters,
		Secrets:    secrets,
	}

	volumeReplication := volumegroup.VolumeGroup{
		Params: c,
	}

	resp := volumeReplication.Create()

	//if resp.Error != nil {
	//	if isKnownError := resp.HasKnownGRPCError(disableReplicationKnownErrors); isKnownError {
	//		logger.Info("volume not found", "volumeID", volumeID)
	//
	//		return nil
	//	}
	//	logger.Error(resp.Error, "failed to disable volume replication")
	//
	//	return resp.Error
	//}

	return resp
}

func getCurrentTime() *metav1.Time {
	metav1NowTime := metav1.NewTime(time.Now())

	return &metav1NowTime
}
