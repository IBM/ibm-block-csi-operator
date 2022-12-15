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
	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
)

func (r *ControllerUtils) AddFinalizerToVG(logger logr.Logger, vg *volumegroupv1.VolumeGroup) error {
	if !Contains(vg.ObjectMeta.Finalizers, VolumeGroupFinalizer) {
		logger.Info("adding finalizer to VolumeGroup object", "Finalizer", VolumeGroupFinalizer)
		vg.ObjectMeta.Finalizers = append(vg.ObjectMeta.Finalizers, VolumeGroupFinalizer)
		if err := r.updateObject(vg); err != nil {
			logger.Error(err, "failed to add finalizer to volumeGroup resource", "finalizer", VolumeGroupFinalizer)
			return err
		}
	}

	return nil
}

func (r *ControllerUtils) AddFinalizerToVGC(logger logr.Logger, vgc *volumegroupv1.VolumeGroupContent) error {
	if !Contains(vgc.ObjectMeta.Finalizers, volumeGroupContentFinalizer) {
		logger.Info("adding finalizer to volumeGroupContent object", "Name", vgc.Name, "Finalizer", volumeGroupContentFinalizer)
		vgc.ObjectMeta.Finalizers = append(vgc.ObjectMeta.Finalizers, volumeGroupContentFinalizer)
		if err := r.updateObject(vgc); err != nil {
			logger.Error(err, "failed to add finalizer to volumeGroupContent resource", "finalizer", VolumeGroupFinalizer)
			return err
		}
	}

	return nil
}

func (r *ControllerUtils) RemoveFinalizerFromVG(logger logr.Logger, vg *volumegroupv1.VolumeGroup) error {
	if Contains(vg.ObjectMeta.Finalizers, VolumeGroupFinalizer) {
		logger.Info("removing finalizer from VolumeGroup object", "Finalizer", VolumeGroupFinalizer)
		vg.ObjectMeta.Finalizers = remove(vg.ObjectMeta.Finalizers, VolumeGroupFinalizer)
		if err := r.updateObject(vg); err != nil {
			logger.Error(err, "failed to remove finalizer to VolumeGroup resource", "finalizer", VolumeGroupFinalizer)
			return err
		}
	}

	return nil
}

func (r *ControllerUtils) RemoveFinalizerFromVGC(logger logr.Logger, vgc *volumegroupv1.VolumeGroupContent) error {
	if Contains(vgc.ObjectMeta.Finalizers, volumeGroupContentFinalizer) {
		logger.Info("removing finalizer from VolumeGroupContent object", "Name", vgc.Name, "Finalizer", volumeGroupContentFinalizer)
		vgc.ObjectMeta.Finalizers = remove(vgc.ObjectMeta.Finalizers, volumeGroupContentFinalizer)
		if err := r.updateObject(vgc); err != nil {
			logger.Error(err, "failed to remove finalizer to VolumeGroupContent resource", "finalizer", VolumeGroupFinalizer)
			return err
		}
	}

	return nil
}

func (r *ControllerUtils) AddFinalizerToPVC(logger logr.Logger, pvc *corev1.PersistentVolumeClaim) error {
	if !Contains(pvc.ObjectMeta.Finalizers, pvcVolumeGroupFinalizer) {
		logger.Info("adding finalizer to PersistentVolumeClaim object", "Namespace", pvc.Namespace, "Name", pvc.Name, "Finalizer", pvcVolumeGroupFinalizer)
		pvc.ObjectMeta.Finalizers = append(pvc.ObjectMeta.Finalizers, pvcVolumeGroupFinalizer)
		if err := r.updateObject(pvc); err != nil {
			logger.Error(err, "failed to add finalizer to PersistentVolumeClaim resource", "finalizer", VolumeGroupFinalizer)
			return err
		}
	}

	return nil
}

func (r *ControllerUtils) RemoveFinalizerFromPVC(logger logr.Logger, pvc *corev1.PersistentVolumeClaim) error {
	if Contains(pvc.ObjectMeta.Finalizers, pvcVolumeGroupFinalizer) {
		logger.Info("removing finalizer from PersistentVolumeClaim object", "Namespace", pvc.Namespace, "Name", pvc.Name, "Finalizer", pvcVolumeGroupFinalizer)
		pvc.ObjectMeta.Finalizers = remove(pvc.ObjectMeta.Finalizers, pvcVolumeGroupFinalizer)
		if err := r.updateObject(pvc); err != nil {
			logger.Error(err, "failed to remove finalizer to PersistentVolumeClaim resource", "finalizer", VolumeGroupFinalizer)
			return err
		}
	}

	return nil
}
