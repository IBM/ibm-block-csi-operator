package utils

import (
	"context"
	"fmt"

	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

func GetPersistentVolumeClaim(logger logr.Logger, client client.Client, name, namespace string) (*corev1.PersistentVolumeClaim, error) {
	logger.Info(fmt.Sprintf(messages.GetPersistentVolumeClaim, namespace, name))
	pvc := &corev1.PersistentVolumeClaim{}
	namespacedPVC := types.NamespacedName{Name: name, Namespace: namespace}
	err := client.Get(context.TODO(), namespacedPVC, pvc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, fmt.Sprintf(messages.PersistentVolumeClaimNotFound, namespace, name))
		} else {
			logger.Error(err, fmt.Sprintf(messages.UnExpectedPersistentVolumeClaimError, namespace, name))
		}
		return nil, err
	}
	return pvc, nil
}

func IsPVCCanBeAddedToVG(logger logr.Logger, client client.Client,
	pvc *corev1.PersistentVolumeClaim, vgs []volumegroupv1.VolumeGroup) error {
	vgsWithPVC := []string{}
	newVGsForPVC := []string{}
	for _, vg := range vgs {
		if IsPVCPartOfVG(pvc, vg.Status.PVCList) {
			vgsWithPVC = append(vgsWithPVC, vg.Name)
		} else if isPVCMatchesVG, _ := IsPVCMatchesVG(logger, client, pvc, vg); isPVCMatchesVG {
			newVGsForPVC = append(newVGsForPVC, vg.Name)
		}
	}
	return checkIfPVCCanBeAddedToVG(logger, pvc, vgsWithPVC, newVGsForPVC)
}

func checkIfPVCCanBeAddedToVG(logger logr.Logger, pvc *corev1.PersistentVolumeClaim,
	vgsWithPVC, newVGsForPVC []string) error {
	if len(vgsWithPVC) > 0 && len(newVGsForPVC) > 0 {
		message := fmt.Sprintf(messages.PersistentVolumeClaimIsAlreadyBelongToGroup, pvc.Namespace, pvc.Name, newVGsForPVC, vgsWithPVC)
		logger.Info(message)
		return fmt.Errorf(message)
	}
	if len(newVGsForPVC) > 1 {
		message := fmt.Sprintf(messages.PersistentVolumeClaimMatchedWithMultipleNewGroups, pvc.Namespace, pvc.Name, newVGsForPVC)
		logger.Info(message)
		return fmt.Errorf(message)
	}
	return nil
}
