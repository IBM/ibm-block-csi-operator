package utils

import (
	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateVolumeGroupSourceContent(client client.Client, instance *volumegroupv1.VolumeGroup,
	vgc *volumegroupv1.VolumeGroupContent, logger logr.Logger) error {
	instance.Spec.Source.VolumeGroupContentName = &vgc.Name
	if err := UpdateObject(client, instance); err != nil {
		logger.Error(err, "failed to update status")
		return err
	}
	return nil
}

func updateVolumeGroupStatus(client client.Client, instance *volumegroupv1.VolumeGroup, logger logr.Logger) error {
	if err := UpdateObjectStatus(client, instance); err != nil {
		logger.Error(err, "failed to update status")

		return err
	}
	return nil
}

func UpdateVolumeGroupStatus(client client.Client, instance *volumegroupv1.VolumeGroup, vgc *volumegroupv1.VolumeGroupContent,
	groupCreationTime *metav1.Time, ready bool, logger logr.Logger) error {
	instance.Status.BoundVolumeGroupContentName = &vgc.Name
	instance.Status.GroupCreationTime = groupCreationTime
	instance.Status.Ready = &ready
	instance.Status.Error = nil

	return updateVolumeGroupStatus(client, instance, logger)
}

func UpdateVolumeGroupStatusError(client client.Client, instance *volumegroupv1.VolumeGroup, logger logr.Logger, message string) error {
	instance.Status.Error = &volumegroupv1.VolumeGroupError{Message: &message}
	err := updateVolumeGroupStatus(client, instance, logger)
	if err != nil {
		logger.Error(err, "failed to update volumeGroup status", "VGName", instance.Name)
		return err
	}

	return nil
}
