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

package utils

import (
	"context"
	"fmt"

	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
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

func GetVGList(logger logr.Logger, client client.Client) (volumegroupv1.VolumeGroupList, error) {
	logger.Info(messages.ListVolumeGroups)
	vg := &volumegroupv1.VolumeGroupList{}
	err := client.List(context.TODO(), vg)
	if err != nil {
		return volumegroupv1.VolumeGroupList{}, err
	}
	return *vg, nil
}

func IsPVCMatchesVG(logger logr.Logger, client client.Client,
	pvc *corev1.PersistentVolumeClaim, vg volumegroupv1.VolumeGroup) (bool, error) {

	logger.Info(fmt.Sprintf(messages.CheckIfPersistentVolumeClaimMatchesVolumeGroup,
		pvc.Namespace, pvc.Name, vg.Namespace, vg.Name))
	areLabelsMatchLabelSelector, err := areLabelsMatchLabelSelector(
		client, pvc.ObjectMeta.Labels, *vg.Spec.Source.Selector)

	if areLabelsMatchLabelSelector {
		logger.Info(fmt.Sprintf(messages.PersistentVolumeClaimMatchedToVolumeGroup,
			pvc.Namespace, pvc.Name, vg.Namespace, vg.Name))
		return true, err
	} else {
		logger.Info(fmt.Sprintf(messages.PersistentVolumeClaimNotMatchedToVolumeGroup,
			pvc.Namespace, pvc.Name, vg.Namespace, vg.Name))
		return false, err
	}
}

func IsPVCPartOfVG(pvc *corev1.PersistentVolumeClaim, pvcList []corev1.PersistentVolumeClaim) bool {
	for _, pvcFromList := range pvcList {
		if pvcFromList.Name == pvc.Name && pvcFromList.Namespace == pvc.Namespace {
			return true
		}
	}
	return false
}

func RemovePVCFromVG(logger logr.Logger, client client.Client, pvc *corev1.PersistentVolumeClaim, vg *volumegroupv1.VolumeGroup) error {
	logger.Info(fmt.Sprintf(messages.RemovePersistentVolumeClaimFromVolumeGroup,
		pvc.Namespace, pvc.Name, vg.Namespace, vg.Name))
	vg.Status.PVCList = RemoveFromPVCList(pvc, vg.Status.PVCList)
	err := client.Status().Update(context.TODO(), vg)
	if err != nil {
		logger.Error(err, fmt.Sprintf(messages.FailedToRemovePersistentVolumeClaimFromVolumeGroup,
			pvc.Namespace, pvc.Name, vg.Namespace, vg.Name))
		return err
	}
	logger.Info(fmt.Sprintf(messages.RemovedPersistentVolumeClaimFromVolumeGroup,
		pvc.Namespace, pvc.Name, vg.Namespace, vg.Name))
	return nil
}

func RemoveFromPVCList(pvc *corev1.PersistentVolumeClaim, pvcList []corev1.PersistentVolumeClaim) []corev1.PersistentVolumeClaim {
	for index, pvcFromList := range pvcList {
		if pvcFromList.Name == pvc.Name && pvcFromList.Namespace == pvc.Namespace {
			pvcList = removeByIndexFromPVCList(pvcList, index)
			return pvcList
		}
	}
	return pvcList
}

func getVgId(logger logr.Logger, client client.Client, vg *volumegroupv1.VolumeGroup) (string, error) {
	vgc, err := GetVolumeGroupContent(client, logger, vg)
	if err != nil {
		return "", err
	}
	return string(vgc.Spec.Source.VolumeGroupHandle), nil
}
