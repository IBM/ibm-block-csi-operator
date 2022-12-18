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
	"fmt"
	"github.com/IBM/volume-group-operator/controllers/utils"
	"github.com/IBM/volume-group-operator/controllers/volumegroup"
	"github.com/IBM/volume-group-operator/pkg/config"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	grpcClient "github.com/IBM/volume-group-operator/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	volumeGroupLAbelKey = "volumegroup"
	snapshotNamePrefix  = "volumegroup"
)

type VolumeGroupReconciler struct {
	client.Client
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
	logger := r.Log.WithValues("Request.Name", req.Name, "Request.Namespace", req.Namespace)

	instance := &volumegroupv1.VolumeGroup{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {

			logger.Info("VolumeGroup resource not found")

			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	vgcObj, err := utils.GetVolumeGroupClass(r.Client, logger, *instance.Spec.VolumeGroupClassName)
	if err != nil {
		uErr := r.updateVolumeGroupStatusError(instance, logger, err.Error())
		if uErr != nil {
			logger.Error(uErr, "failed to update volumeGroup status", "VGName", instance.Name)
		}

		return ctrl.Result{}, err
	}

	if r.DriverConfig.DriverName != vgcObj.Driver {
		return ctrl.Result{}, nil
	}

	err = utils.ValidatePrefixedParameters(vgcObj.Parameters)
	if err != nil {
		logger.Error(err, "failed to validate parameters of volumegroupClass", "VGClassName", vgcObj.Name)
		uErr := r.updateVolumeGroupStatusError(instance, logger, err.Error())
		if uErr != nil {
			logger.Error(uErr, "failed to update volumeGroup status", "VGName", instance.Name)
		}

		return ctrl.Result{}, err
	}
	parameters := utils.FilterPrefixedParameters(utils.VolumeGroupAsPrefix, vgcObj.Parameters)

	secretName := vgcObj.Parameters[utils.PrefixedVolumeGroupSecretNameKey]
	secretNamespace := vgcObj.Parameters[utils.PrefixedVolumeGroupSecretNamespaceKey]
	secret := make(map[string]string)
	if secretName != "" && secretNamespace != "" {
		secret, err = utils.GetSecretData(r.Client, logger, secretName, secretNamespace)
		if err != nil {
			uErr := r.updateVolumeGroupStatusError(instance, logger, err.Error())
			if uErr != nil {
				logger.Error(uErr, "failed to update volumeGroup status", "VGName", instance.Name)
			}

			return reconcile.Result{}, err
		}
	}

	if instance.GetDeletionTimestamp().IsZero() {
		if err = utils.AddFinalizerToVG(r.Client, logger, instance); err != nil {
			logger.Error(err, "Failed to add VolumeGroup finalizer")

			return reconcile.Result{}, err
		}

	} else {
		if utils.Contains(instance.GetFinalizers(), utils.VolumeGroupFinalizer) {
			volumeGroupContent, uErr := utils.GetVolumeGroupContent(r.Client, logger, instance)
			if uErr != nil {
				if !errors.IsNotFound(uErr) {
					logger.Error(uErr, "failed to get volumeGroupContent", *instance.Spec.Source.VolumeGroupContentName)
					return ctrl.Result{}, uErr
				}

			} else {
				volumeGroupId := volumeGroupContent.Spec.Source.VolumeGroupHandle
				if err = r.deleteVolumeGroup(logger, volumeGroupId, secret); err != nil {
					logger.Error(err, "failed to delete volume group")

					return ctrl.Result{}, err
				}
				if err = utils.RemoveFinalizerFromVGC(r.Client, logger, volumeGroupContent); err != nil {
					logger.Error(err, "Failed to remove volume group content finalizer")

					return reconcile.Result{}, err
				}
				if err = r.Client.Delete(context.TODO(), volumeGroupContent); err != nil {
					logger.Error(err, "Failed to remove volume group content finalizer")

					return reconcile.Result{}, err
				}
			}

			if err = utils.RemoveFinalizerFromVG(r.Client, logger, instance); err != nil {
				logger.Error(err, "Failed to remove volume group finalizer")

				return reconcile.Result{}, err
			}

		}

		logger.Info("volumeGroup object is terminated, skipping reconciliation")

		return ctrl.Result{}, nil
	}

	groupCreationTime := getCurrentTime()
	volumeGroupName, err := makeVolumeGroupName(snapshotNamePrefix, string(instance.UID))
	if err != nil {
		return reconcile.Result{}, err
	}

	createVolumeGroupResponse := r.createVolumeGroup(volumeGroupName, parameters, secret)
	if createVolumeGroupResponse.Error != nil {
		logger.Error(err, "failed to create volume group")
		msg := utils.GetMessageFromError(createVolumeGroupResponse.Error)
		uErr := r.updateVolumeGroupStatusError(instance, logger, msg)
		if uErr != nil {
			logger.Error(uErr, "failed to update volumeGroup status", "VGName", instance.Name)
		}

		return reconcile.Result{}, err
	}
	vgc := utils.GenerateVolumeGroupContent(volumeGroupName, instance, vgcObj, createVolumeGroupResponse, secretName, secretNamespace)

	if err = utils.CreateVolumeGroupContent(r.Client, logger, instance, vgc); err != nil {
		logger.Error(err, "failed to create volumeGroupContent", "VGCName", vgc.Name)
		return reconcile.Result{}, err
	}
	instance.Spec.Source = volumegroupv1.VolumeGroupSource{
		VolumeGroupContentName: &vgc.Name,
		Selector:               r.volumeGroupLabelSelector(instance),
	}

	if err = utils.UpdateObject(r.Client, instance); err != nil {
		return reconcile.Result{}, err
	}
	ready := true
	instance.Status = volumegroupv1.VolumeGroupStatus{
		BoundVolumeGroupContentName: &vgc.Name,
		GroupCreationTime:           groupCreationTime,
		Ready:                       &ready,
		Error:                       nil,
	}

	uErr := r.updateVolumeGroupStatus(instance, logger)
	if uErr != nil {
		logger.Error(uErr, "failed to update volumeGroup status", "VGName", instance.Name)
	}
	vgc, err = utils.GetVolumeGroupContent(r.Client, logger, instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err = utils.AddFinalizerToVGC(r.Client, logger, vgc); err != nil {
		logger.Error(err, "Failed to add VolumeGroup finalizer")

		return reconcile.Result{}, err
	}

	if err = utils.UpdateVolumeGroupStatus(r.Client, logger, vgc, groupCreationTime, &ready); err != nil {
		return reconcile.Result{}, err
	}
	//TODO CSI-4986 add all PVCs that have the VG label to VG

	return ctrl.Result{}, nil
}

func makeVolumeGroupName(prefix string, volumeGroupUID string) (string, error) {
	if len(volumeGroupUID) == 0 {
		return "", fmt.Errorf("Corrupted volumeGroup object, it is missing UID")
	}
	return fmt.Sprintf("%s-%s", prefix, volumeGroupUID), nil
}

func (r *VolumeGroupReconciler) volumeGroupLabelSelector(instance *volumegroupv1.VolumeGroup) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{volumeGroupLAbelKey: instance.Name},
	}
}

func (r *VolumeGroupReconciler) SetupWithManager(mgr ctrl.Manager, cfg *config.DriverConfig) error {
	pred := predicate.GenerationChangedPredicate{}

	r.VolumeGroupClient = grpcClient.NewVolumeGroupClient(r.GRPCClient.Client, cfg.RPCTimeout)

	return ctrl.NewControllerManagedBy(mgr).
		For(&volumegroupv1.VolumeGroup{}).
		WithEventFilter(pred).Complete(r)
}

func (r *VolumeGroupReconciler) updateVolumeGroupStatusError(
	instance *volumegroupv1.VolumeGroup,
	logger logr.Logger,
	message string) error {
	instance.Status.Error = &volumegroupv1.VolumeGroupError{Message: &message}
	err := r.updateVolumeGroupStatus(instance, logger)
	if err != nil {
		return err
	}

	return nil
}

func (r *VolumeGroupReconciler) updateVolumeGroupStatus(instance *volumegroupv1.VolumeGroup, logger logr.Logger) error {
	if err := utils.UpdateObjectStatus(r.Client, instance); err != nil {
		logger.Error(err, "failed to update status")

		return err
	}
	return nil
}

func (r *VolumeGroupReconciler) deleteVolumeGroup(logger logr.Logger, volumeGroupId string, secrets map[string]string) error {
	c := volumegroup.CommonRequestParameters{
		VolumeGroupID: volumeGroupId,
		Secrets:       secrets,
		VolumeGroup:   r.VolumeGroupClient,
	}

	volumeGroup := volumegroup.VolumeGroup{
		Params: c,
	}

	resp := volumeGroup.Delete()

	if resp.Error != nil {
		logger.Error(resp.Error, "failed to delete volume group")
		return resp.Error
	}

	return nil
}

func (r *VolumeGroupReconciler) createVolumeGroup(volumeGroupName string, parameters, secrets map[string]string) *volumegroup.Response {
	c := volumegroup.CommonRequestParameters{
		Name:        volumeGroupName,
		Parameters:  parameters,
		Secrets:     secrets,
		VolumeGroup: r.VolumeGroupClient,
	}

	volumeGroup := volumegroup.VolumeGroup{
		Params: c,
	}

	resp := volumeGroup.Create()

	return resp
}

func getCurrentTime() *metav1.Time {
	metav1NowTime := metav1.NewTime(time.Now())

	return &metav1NowTime
}
