package utils

import (
	"context"
	"fmt"

	vgerrors "github.com/IBM/volume-group-operator/pkg/errors"
	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetPVFromPVC(logger logr.Logger, client client.Client, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolume, error) {
	logger.Info(fmt.Sprintf(messages.GetPersistentVolumeOfPersistentVolumeClaim, pvc.Name, pvc.Namespace))
	pvName, err := getPersistentVolumeNameFromPersistentVolumeClaim(logger, client, pvc)
	if err != nil {
		return nil, err
	}

	pv, err := getPersistentVolume(logger, client, pvName)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, &vgerrors.PersistentVolumeDoesNotExist{pvName, pvc.Namespace, err.Error()}
		}
		return nil, err
	}
	return pv, nil
}

func getPersistentVolumeNameFromPersistentVolumeClaim(logger logr.Logger, client client.Client,
	pvc *corev1.PersistentVolumeClaim) (string, error) {
	pvName := pvc.Spec.VolumeName
	if pvName == "" {
		logger.Info(messages.PersistentVolumeClaimDoesNotHavePersistentVolume)
		pvc, err := getPersistentVolumeClaim(logger, client, pvc)
		if err != nil {
			return "", err
		}
		if pvc.Spec.VolumeName != "" {
			return pvc.Spec.VolumeName, nil
		}
		return "", nil
	}
	return pvName, nil
}

func getPersistentVolume(logger logr.Logger, client client.Client, pvName string) (*corev1.PersistentVolume, error) {
	logger.Info(fmt.Sprintf(messages.GetPersistentVolumeClaim, pvName))
	pv := &corev1.PersistentVolume{}
	namespacedPV := types.NamespacedName{Name: pvName}
	err := client.Get(context.TODO(), namespacedPV, pv)
	if err != nil {
		logger.Error(err, fmt.Sprintf(messages.FailedToGetPersistentVolume, pvName))
		return nil, err
	}
	return pv, nil
}
