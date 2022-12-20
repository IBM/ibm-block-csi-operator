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

package persistentvolumeclaim

import (
	"context"
	"fmt"

	csiv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/IBM/volume-group-operator/controllers/utils"
	"github.com/IBM/volume-group-operator/controllers/volumegroup"
	grpcClient "github.com/IBM/volume-group-operator/pkg/client"
	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type PersistentVolumeClaimWatcher struct {
	Client            client.Client
	Scheme            *runtime.Scheme
	Log               logr.Logger
	GRPCClient        *grpcClient.Client
	VolumeGroupClient grpcClient.VolumeGroup
}

func (r *PersistentVolumeClaimWatcher) Reconcile(_ context.Context, req reconcile.Request) (result reconcile.Result, err error) {
	result = reconcile.Result{}
	reqLogger := r.Log.WithValues(messages.RequestNamespace, req.Namespace, messages.RequestName, req.Name)
	reqLogger.Info(messages.ReconcilePersistentVolumeClaim)
	pvc, err := r.getPersistentVolumeClaim(reqLogger, req)
	if err != nil {
		if errors.IsNotFound(err) {
			return result, nil
		}
		return result, err
	}
	vgList, err := utils.GetVGList(reqLogger, r.Client)
	if err != nil {
		return result, err
	}

	err = r.removePersistentVolumeClaimFromVolumeGroupObjects(reqLogger, pvc, vgList)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r PersistentVolumeClaimWatcher) getPersistentVolumeClaim(logger logr.Logger,
	req reconcile.Request) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, pvc)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, messages.PersistentVolumeClaimNotFound, persistentVolumeClaim, pvc)
		} else {
			logger.Error(err, messages.UnExpectedPersistentVolumeClaimError, persistentVolumeClaim, pvc)
		}

		return nil, err
	}

	return pvc, nil
}

func (r PersistentVolumeClaimWatcher) removePersistentVolumeClaimFromVolumeGroupObjects(
	logger logr.Logger, pvc *corev1.PersistentVolumeClaim, vgList csiv1.VolumeGroupList) error {
	for _, vg := range vgList.Items {
		if !utils.IsPVCPartOfVG(pvc, vg.Status.PVCList) {
			continue
		}
		IsPVCMatchesVG, err := utils.IsPVCMatchesVG(logger, r.Client, pvc, vg)
		if err != nil {
			return r.updateVolumeGroupStatusError(logger, &vg, err)
		}

		if !IsPVCMatchesVG {
			err := r.removeVolumeFromVolumeGroup(logger, pvc, &vg)
			if err != nil {
				return err
			}
			err = r.removeVolumeFromPvcListAndPvList(logger, pvc, vg)
			if err != nil {
				return r.updateVolumeGroupStatusError(logger, &vg, err)
			}
		}
	}
	return nil
}

func (r PersistentVolumeClaimWatcher) removeVolumeFromVolumeGroup(logger logr.Logger,
	pvc *corev1.PersistentVolumeClaim, vg *csiv1.VolumeGroup) error {
	vg.Status.PVCList = utils.RemovePVCFromVGPVCList(pvc, vg.Status.PVCList)

	err := r.modifyVolumeGroup(logger, vg)
	return r.updateVolumeGroupStatusError(logger, vg, err)
}

func (r *PersistentVolumeClaimWatcher) modifyVolumeGroup(logger logr.Logger, vg *csiv1.VolumeGroup) error {
	params, err := r.generateModifyVolumeGroupParams(logger, vg)
	if err != nil {
		return err
	}
	volumeGroup := volumegroup.VolumeGroup{
		Params: params,
	}

	modifyVolumeGroupResponse := volumeGroup.Modify()
	responseError := modifyVolumeGroupResponse.Error
	if responseError != nil {
		logger.Error(responseError, fmt.Sprintf(messages.FailedToModifyVolumeGroup, vg.Namespace, vg.Name))
		return responseError
	}
	return nil
}
func (r *PersistentVolumeClaimWatcher) generateModifyVolumeGroupParams(logger logr.Logger,
	vg *csiv1.VolumeGroup) (volumegroup.CommonRequestParameters, error) {
	vgId, err := utils.GetVgId(logger, r.Client, vg)
	if err != nil {
		return volumegroup.CommonRequestParameters{}, err
	}
	volumeIds, err := utils.GetPVCListVolumeIds(logger, r.Client, vg.Status.PVCList)
	if err != nil {
		return volumegroup.CommonRequestParameters{}, err
	}
	secrets, err := r.getSecrets(logger, *vg.Spec.VolumeGroupClassName)
	if err != nil {
		return volumegroup.CommonRequestParameters{}, err
	}

	return volumegroup.CommonRequestParameters{
		Secrets:       secrets,
		VolumeGroup:   r.VolumeGroupClient,
		VolumeGroupID: vgId,
		VolumeIds:     volumeIds,
	}, nil
}
func (r *PersistentVolumeClaimWatcher) getSecrets(logger logr.Logger, vgcName string) (map[string]string, error) {
	vgc, err := utils.GetVolumeGroupClass(r.Client, logger, vgcName)
	if err != nil {
		return nil, err
	}
	secrets, err := utils.GetSecretDataFromVolumeGroupClass(r.Client, logger, vgc)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (r *PersistentVolumeClaimWatcher) updateVolumeGroupStatusError(logger logr.Logger,
	vg *csiv1.VolumeGroup, err error) error {
	if err != nil {
		msg := utils.GetMessageFromError(err)
		uErr := utils.UpdateVolumeGroupStatusError(r.Client, vg, logger, msg)
		if uErr != nil {
			return uErr
		}
	}
	return err
}

func (r PersistentVolumeClaimWatcher) removeVolumeFromPvcListAndPvList(logger logr.Logger,
	pvc *corev1.PersistentVolumeClaim, vg csiv1.VolumeGroup) error {
	err := utils.RemovePVCFromVG(logger, r.Client, pvc, &vg)
	if err != nil {
		return err
	}
	pv, err := utils.GetPVFromPVC(logger, r.Client, pvc)
	if err != nil {
		return err
	}
	vgc, err := utils.GetVolumeGroupContent(r.Client, logger, &vg)
	if err != nil {
		return err
	}

	if pv != nil {
		err = utils.RemovePVFromVGC(logger, r.Client, pv, vgc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PersistentVolumeClaimWatcher) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.LabelChangedPredicate{}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PersistentVolumeClaim{}, builder.WithPredicates(pvcPredicate)).
		WithEventFilter(pred).Complete(r)
}
