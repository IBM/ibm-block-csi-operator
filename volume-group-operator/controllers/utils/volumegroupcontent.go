package utils

import (
	"context"
	"fmt"

	csiv1 "github.com/IBM/volume-group-operator/api/v1"
	vgerrors "github.com/IBM/volume-group-operator/pkg/errors"
	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetVGC(logger logr.Logger, client client.Client, vg *csiv1.VolumeGroup) (*csiv1.VolumeGroupContent, error) {
	logger.Info(fmt.Sprintf(messages.GetVolumeGroupContentOfVolumeGroup, vg.Name, vg.Namespace))
	vgc := &csiv1.VolumeGroupContent{}
	VGCName := *vg.Spec.Source.VolumeGroupContentName
	namespacedVGC := types.NamespacedName{Name: VGCName, Namespace: vg.Namespace}
	err := client.Get(context.TODO(), namespacedVGC, vgc)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, &vgerrors.PersistentVolumeDoesNotExist{VGCName, vg.Namespace, err.Error()}
		}

		return nil, err
	}

	return vgc, nil
}

func RemovePVFromVGC(logger logr.Logger, client client.Client, pv *corev1.PersistentVolume, vgc *csiv1.VolumeGroupContent) error {
	logger.Info(fmt.Sprintf(messages.RemovePersistentVolumeFromVolumeGroupContent,
		pv.Namespace, pv.Name, vgc.Namespace, vgc.Name))
	vgc.Status.PVList = removePersistentVolumeFromVolumeGroupContentPVList(pv, vgc.Status.PVList)
	err := client.Status().Update(context.TODO(), vgc)
	if err != nil {
		logger.Error(err, fmt.Sprintf(messages.FailedToRemovePersistentVolumeFromVolumeGroupContent,
			pv.Namespace, pv.Name, vgc.Namespace, vgc.Name))
		return err
	}
	logger.Info(fmt.Sprintf(messages.RemovedPersistentVolumeFromVolumeGroupContent,
		pv.Namespace, pv.Name, vgc.Namespace, vgc.Name))
	return nil
}

func removePersistentVolumeFromVolumeGroupContentPVList(pv *corev1.PersistentVolume,
	pvListInVGC []corev1.PersistentVolume) []corev1.PersistentVolume {
	for index, pvcFromList := range pvListInVGC {
		if pvcFromList.Name == pv.Name && pvcFromList.Namespace == pv.Namespace {
			pvListInVGC = removeByIndexFromPersistentVolumeList(pvListInVGC, index)
			return pvListInVGC
		}
	}
	return pvListInVGC
}
