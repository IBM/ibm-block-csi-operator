package utils

import (
	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HandleErrorMessage(logger logr.Logger, client client.Client, vg *volumegroupv1.VolumeGroup,
	err error, reason string) error {
	if err != nil {
		errorMessage := GetMessageFromError(err)
		uErr := UpdateVolumeGroupStatusError(client, vg, logger, errorMessage)
		if uErr != nil {
			return uErr
		}
		uErr = createNamespacedObjectErrorEvent(logger, client, vg, errorMessage, reason)
		if uErr != nil {
			return uErr
		}
		return err
	}
	return nil
}

func HandleSuccessMessage(logger logr.Logger, client client.Client, vg *volumegroupv1.VolumeGroup, message, reason string) error {
	err := UpdateVolumeGroupStatusError(client, vg, logger, "")
	if err != nil {
		return err
	}
	err = createSuccessVolumeGroupEvent(logger, client, vg, message, reason)
	if err != nil {
		return err
	}
	return nil
}

func HandlePVCErrorMessage(logger logr.Logger, client client.Client, pvc *corev1.PersistentVolumeClaim,
	err error, reason string) error {
	if err != nil {
		errorMessage := GetMessageFromError(err)
		uErr := createNamespacedObjectErrorEvent(logger, client, pvc, errorMessage, reason)
		if uErr != nil {
			return uErr
		}
		return err
	}
	return nil
}
