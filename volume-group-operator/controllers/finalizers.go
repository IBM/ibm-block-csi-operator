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

package controllers

import (
	"context"
	"fmt"

	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
)

const (
	volumeGroupFinalizer        = "volumegroup.storage.ibm.io"
	volumeGroupContentFinalizer = "volumegroup.storage.ibm.io/vgc-protection"
	pvcVolumeGroupFinalizer     = "volumegroup.storage.ibm.io/pvc-protection"
)

// addFinalizerToVG adds the VG finalizer on the VolumeGroup instance.
func (r *VolumeGroupReconciler) addFinalizerToVG(logger logr.Logger, vg *volumegroupv1.VolumeGroup,
) error {
	if !contains(vg.ObjectMeta.Finalizers, volumeGroupFinalizer) {
		logger.Info("adding finalizer to VolumeGroup object", "Finalizer", volumeGroupFinalizer)
		vg.ObjectMeta.Finalizers = append(vg.ObjectMeta.Finalizers, volumeGroupFinalizer)
		if err := r.Client.Update(context.TODO(), vg); err != nil {
			return fmt.Errorf("failed to add finalizer (%s) to volumeGroup resource"+
				" (%s/%s) %w",
				volumeGroupFinalizer, vg.Namespace, vg.Name, err)
		}
	}

	return nil
}

// addFinalizerToVGC adds the VG finalizer on the VolumeGroupContent instance.
func (r *VolumeGroupReconciler) addFinalizerToVGC(logger logr.Logger, vgc *volumegroupv1.VolumeGroupContent,
) error {
	if !contains(vgc.ObjectMeta.Finalizers, volumeGroupContentFinalizer) {
		logger.Info("adding finalizer to volumeGroupContent object", "Finalizer", volumeGroupContentFinalizer)
		vgc.ObjectMeta.Finalizers = append(vgc.ObjectMeta.Finalizers, volumeGroupContentFinalizer)
		if err := r.Client.Update(context.TODO(), vgc); err != nil {
			return fmt.Errorf("failed to add finalizer (%s) to volumeGroupContent resource"+
				" (%s/%s) %w",
				volumeGroupContentFinalizer, vgc.Namespace, vgc.Name, err)
		}
	}

	return nil
}

// removeFinalizerFromVG removes the VG finalizer from the VolumeGroup instance.
func (r *VolumeGroupReconciler) removeFinalizerFromVG(logger logr.Logger, vg *volumegroupv1.VolumeGroup) error {
	if contains(vg.ObjectMeta.Finalizers, volumeGroupFinalizer) {
		logger.Info("removing finalizer from VolumeGroup object", "Finalizer", volumeGroupFinalizer)
		vg.ObjectMeta.Finalizers = remove(vg.ObjectMeta.Finalizers, volumeGroupFinalizer)
		if err := r.Client.Update(context.TODO(), vg); err != nil {
			return fmt.Errorf("failed to remove finalizer (%s) from VolumeGroup resource"+
				" (%s/%s), %w",
				volumeGroupFinalizer, vg.Namespace, vg.Name, err)
		}
	}

	return nil
}

// removeFinalizerFromVGC removes the VG finalizer from the VolumeGroupContent instance.
func (r *VolumeGroupReconciler) removeFinalizerFromVGC(logger logr.Logger, vgc *volumegroupv1.VolumeGroupContent) error {
	if contains(vgc.ObjectMeta.Finalizers, volumeGroupContentFinalizer) {
		logger.Info("removing finalizer from VolumeGroupContent object", "Finalizer", volumeGroupContentFinalizer)
		vgc.ObjectMeta.Finalizers = remove(vgc.ObjectMeta.Finalizers, volumeGroupContentFinalizer)
		if err := r.Client.Update(context.TODO(), vgc); err != nil {
			return fmt.Errorf("failed to remove finalizer (%s) from VolumeGroupContent resource"+
				" (%s/%s), %w",
				volumeGroupContentFinalizer, vgc.Namespace, vgc.Name, err)
		}
	}

	return nil
}

// addFinalizerToPVC adds the VG finalizer on the PersistentVolumeClaim.
func (r *VolumeGroupReconciler) addFinalizerToPVC(logger logr.Logger, pvc *corev1.PersistentVolumeClaim) error {
	if !contains(pvc.ObjectMeta.Finalizers, pvcVolumeGroupFinalizer) {
		logger.Info("adding finalizer to PersistentVolumeClaim object", "Finalizer", pvcVolumeGroupFinalizer)
		pvc.ObjectMeta.Finalizers = append(pvc.ObjectMeta.Finalizers, pvcVolumeGroupFinalizer)
		if err := r.Client.Update(context.TODO(), pvc); err != nil {
			return fmt.Errorf("failed to add finalizer (%s) to PersistentVolumeClaim resource"+
				" (%s/%s) %w",
				pvcVolumeGroupFinalizer, pvc.Namespace, pvc.Name, err)
		}
	}

	return nil
}

// removeFinalizerFromPVC removes the VG finalizer on PersistentVolumeClaim.
func (r *VolumeGroupReconciler) removeFinalizerFromPVC(logger logr.Logger, pvc *corev1.PersistentVolumeClaim,
) error {
	if contains(pvc.ObjectMeta.Finalizers, pvcVolumeGroupFinalizer) {
		logger.Info("removing finalizer from PersistentVolumeClaim object", "Finalizer", pvcVolumeGroupFinalizer)
		pvc.ObjectMeta.Finalizers = remove(pvc.ObjectMeta.Finalizers, pvcVolumeGroupFinalizer)
		if err := r.Client.Update(context.TODO(), pvc); err != nil {
			return fmt.Errorf("failed to remove finalizer (%s) from PersistentVolumeClaim resource"+
				" (%s/%s), %w",
				pvcVolumeGroupFinalizer, pvc.Namespace, pvc.Name, err)
		}
	}

	return nil
}
