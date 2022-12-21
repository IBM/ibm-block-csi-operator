package utils

import (
	"context"
	"fmt"

	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getPVCListVolumeIds(logger logr.Logger, client client.Client, pvcList []corev1.PersistentVolumeClaim) ([]string, error) {
	volumeIds := []string{}
	for _, pvc := range pvcList {
		pv, err := GetPVFromPVC(logger, client, &pvc)
		if err != nil {
			return nil, err
		}
		if pv != nil {
			volumeIds = append(volumeIds, string(pv.Spec.CSI.VolumeHandle))
		}
	}
	return volumeIds, nil
}

func getPersistentVolumeClaim(logger logr.Logger, client client.Client,
	pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	logger.Info(fmt.Sprintf(messages.GetPersistentVolumeClaim, pvc.Namespace, pvc.Name))
	newPVC := &corev1.PersistentVolumeClaim{}
	namespacedPVC := types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}
	err := client.Get(context.TODO(), namespacedPVC, newPVC)
	if err != nil {
		logger.Error(err, fmt.Sprintf(messages.FailedToGetPersistentVolumeClaim, pvc.Namespace, pvc.Name))
		return nil, err
	}
	return newPVC, nil
}
