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
	pvName := pvc.Spec.VolumeName
	if pvName == "" {
		logger.Info(messages.PersistentVolumeClaimDoesNotHavePersistentVolume)
		return nil, nil
	}

	namespacedPV := types.NamespacedName{Name: pvName, Namespace: pvc.Namespace}
	pv := &corev1.PersistentVolume{}
	err := client.Get(context.TODO(), namespacedPV, pv)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, &vgerrors.PersistentVolumeDoesNotExist{pvName, pvc.Namespace, err.Error()}
		}
		return nil, err
	}
	return pv, nil
}
