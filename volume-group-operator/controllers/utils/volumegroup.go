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

	csiv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetVGList(logger logr.Logger, client client.Client) (csiv1.VolumeGroupList, error) {
	logger.Info(messages.ListVolumeGroups)
	vg := &csiv1.VolumeGroupList{}
	err := client.List(context.TODO(), vg)
	if err != nil {
		return csiv1.VolumeGroupList{}, err
	}
	return *vg, nil
}

func IsPVCMatchesVG(logger logr.Logger, client client.Client,
	pvc *corev1.PersistentVolumeClaim, vg csiv1.VolumeGroup) (bool, error) {

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

func IsPvcPartOfVG(pvcName string, pvcListInVG []corev1.PersistentVolumeClaim) bool {
	for _, pvc := range pvcListInVG {
		if pvc.Name == pvcName {
			return true
		}
	}
	return false
}
