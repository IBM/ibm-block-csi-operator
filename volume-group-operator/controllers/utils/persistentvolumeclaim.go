package utils

import (
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetPVCListVolumeIds(logger logr.Logger, client client.Client, pvcList []corev1.PersistentVolumeClaim) ([]string, error) {
	volumeIds := []string{}
	for _, pvc := range pvcList {
		pv, err := GetPVFromPVC(logger, client, &pvc)
		if err != nil {
			return nil, err
		}
		volumeIds = append(volumeIds, string(pv.Spec.CSI.VolumeHandle))
	}
	return volumeIds, nil
}
