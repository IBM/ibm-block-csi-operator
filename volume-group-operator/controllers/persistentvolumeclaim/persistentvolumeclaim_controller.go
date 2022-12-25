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
	grpcClient "github.com/IBM/volume-group-operator/pkg/client"
	"github.com/IBM/volume-group-operator/pkg/config"
	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type PersistentVolumeClaimWatcher struct {
	Client            client.Client
	Scheme            *runtime.Scheme
	Log               logr.Logger
	DriverConfig      *config.DriverConfig
	GRPCClient        *grpcClient.Client
	VolumeGroupClient grpcClient.VolumeGroup
}

func (r *PersistentVolumeClaimWatcher) Reconcile(_ context.Context, req reconcile.Request) (result reconcile.Result, err error) {
	result = reconcile.Result{}
	reqLogger := r.Log.WithValues(messages.RequestNamespace, req.Namespace, messages.RequestName, req.Name)
	reqLogger.Info(messages.ReconcilePersistentVolumeClaim)
	pvc, err := utils.GetPersistentVolumeClaim(reqLogger, r.Client, req.Name, req.Namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			return result, nil
		}
		return result, err
	}
	if pvc.Status.Phase != corev1.ClaimBound {
		reqLogger.Info(messages.PersistentVolumeClaimIsNotInBoundPhase)
		return result, nil
	}
	err = r.removePersistentVolumeClaimFromVolumeGroupObjects(reqLogger, pvc)
	if err != nil {
		return result, err
	}

	err = r.addPersistentVolumeClaimToVolumeGroupObjects(reqLogger, pvc)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r PersistentVolumeClaimWatcher) removePersistentVolumeClaimFromVolumeGroupObjects(
	logger logr.Logger, pvc *corev1.PersistentVolumeClaim) error {
	vgList, err := utils.GetVGList(logger, r.Client)
	if err != nil {
		return err
	}

	for _, vg := range vgList.Items {
		if !utils.IsPVCPartOfVG(pvc, vg.Status.PVCList) {
			continue
		}
		IsPVCMatchesVG, err := utils.IsPVCMatchesVG(logger, r.Client, pvc, vg)
		if err != nil {
			return utils.HandleErrorMessage(logger, r.Client, &vg, err, removingPVC)
		}

		if !IsPVCMatchesVG {
			err := r.removeVolumeFromVolumeGroup(logger, pvc, &vg)
			if err != nil {
				return utils.HandleErrorMessage(logger, r.Client, &vg, err, removingPVC)
			}
			err = r.removeVolumeFromPvcListAndPvList(logger, pvc, vg)
			if err != nil {
				return utils.HandleErrorMessage(logger, r.Client, &vg, err, removingPVC)
			}
			err = r.addSuccessRemoveEvent(logger, pvc, &vg)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r PersistentVolumeClaimWatcher) addSuccessRemoveEvent(logger logr.Logger,
	pvc *corev1.PersistentVolumeClaim, vg *csiv1.VolumeGroup) error {
	message := fmt.Sprintf(messages.RemovedPersistentVolumeClaimFromVolumeGroup,
		pvc.Namespace, pvc.Name, vg.Namespace, vg.Name)
	return utils.HandleSuccessMessage(logger, r.Client, vg, message, removingPVC)
}

func (r PersistentVolumeClaimWatcher) removeVolumeFromVolumeGroup(logger logr.Logger,
	pvc *corev1.PersistentVolumeClaim, vg *csiv1.VolumeGroup) error {
	logger.Info(fmt.Sprintf(messages.RemoveVolumeFromVolumeGroup, vg.Namespace, vg.Name))
	vg.Status.PVCList = utils.RemovePVCFromVGPVCList(pvc, vg.Status.PVCList)

	err := utils.ModifyVolumeGroup(logger, r.Client, vg, r.VolumeGroupClient)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf(messages.RemovedVolumeFromVolumeGroup, vg.Namespace, vg.Name))
	return nil
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

func (r PersistentVolumeClaimWatcher) addPersistentVolumeClaimToVolumeGroupObjects(
	logger logr.Logger, pvc *corev1.PersistentVolumeClaim) error {
	var err error
	vgList, err := utils.GetVGList(logger, r.Client)
	if err != nil {
		return err
	}
	err = r.isPVCCanBeAddedToVG(logger, pvc, vgList)
	if err != nil {
		return err
	}

	for _, vg := range vgList.Items {
		isPVCMatchesVG := false
		if !utils.IsPVCPartOfVG(pvc, vg.Status.PVCList){
			isPVCMatchesVG, err = utils.IsPVCMatchesVG(logger, r.Client, pvc, vg)
			if err != nil {
				return utils.HandleErrorMessage(logger, r.Client, &vg, err, addingPVC)
			}
		}

		if isPVCMatchesVG {
			err := r.addVolumeToVolumeGroup(logger, pvc, &vg)
			if err != nil {
				return utils.HandleErrorMessage(logger, r.Client, &vg, err, addingPVC)
			}
			err = r.addVolumeFromPvcListAndPvList(logger, pvc, &vg)
			return utils.HandleErrorMessage(logger, r.Client, &vg, err, addingPVC)
		}
	}
	return nil
}

func (r PersistentVolumeClaimWatcher) isPVCCanBeAddedToVG(logger logr.Logger, pvc *corev1.PersistentVolumeClaim,
	vgList csiv1.VolumeGroupList) error {
	if r.DriverConfig.MultipleVGsToPVC == "true" {
		return nil
	}
	err := utils.IsPVCCanBeAddedToVG(logger, r.Client, pvc, vgList.Items)
	return utils.HandlePVCErrorMessage(logger, r.Client, pvc, err, addingPVC)
}

func (r PersistentVolumeClaimWatcher) addVolumeToVolumeGroup(logger logr.Logger,
	pvc *corev1.PersistentVolumeClaim, vg *csiv1.VolumeGroup) error {
	logger.Info(fmt.Sprintf(messages.AddVolumeToVolumeGroup, vg.Namespace, vg.Name))
	vg.Status.PVCList = append(vg.Status.PVCList, *pvc)

	err := utils.ModifyVolumeGroup(logger, r.Client, vg, r.VolumeGroupClient)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf(messages.AddedVolumeToVolumeGroup, vg.Namespace, vg.Name))
	return nil
}

func (r PersistentVolumeClaimWatcher) addVolumeFromPvcListAndPvList(logger logr.Logger,
	pvc *corev1.PersistentVolumeClaim, vg *csiv1.VolumeGroup) error {
	err := utils.AddPVCToVG(logger, r.Client, pvc, vg)
	if err != nil {
		return err
	}

	err = r.addPVToVGC(logger, pvc, vg)
	if err != nil {
		return err
	}

	return r.addSuccessAddEvent(logger, pvc, vg)
}

func (r PersistentVolumeClaimWatcher) addPVToVGC(logger logr.Logger, pvc *corev1.PersistentVolumeClaim, vg *csiv1.VolumeGroup) error {
	pv, err := utils.GetPVFromPVC(logger, r.Client, pvc)
	if err != nil {
		return err
	}
	vgc, err := utils.GetVolumeGroupContent(r.Client, logger, vg)
	if err != nil {
		return err
	}

	if pv != nil {
		return utils.AddPVToVGC(logger, r.Client, pv, vgc)
	}
	return nil
}

func (r PersistentVolumeClaimWatcher) addSuccessAddEvent(logger logr.Logger,
	pvc *corev1.PersistentVolumeClaim, vg *csiv1.VolumeGroup) error {
	message := fmt.Sprintf(messages.AddedPersistentVolumeClaimToVolumeGroup, pvc.Namespace, pvc.Name, vg.Namespace, vg.Name)
	return utils.HandleSuccessMessage(logger, r.Client, vg, message, addingPVC)
}

func (r *PersistentVolumeClaimWatcher) SetupWithManager(mgr ctrl.Manager, cfg *config.DriverConfig) error {
	r.VolumeGroupClient = grpcClient.NewVolumeGroupClient(r.GRPCClient.Client, cfg.RPCTimeout)

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PersistentVolumeClaim{}, builder.WithPredicates(pvcPredicate)).
		Complete(r)
}
