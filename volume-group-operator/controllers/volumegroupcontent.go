package controllers

import (
	"context"
	"fmt"
	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	volumegroup "github.com/IBM/volume-group-operator/controllers/volumegroup"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// getVolumeGroupContentSource get VolumeGroupContentSource object from the request.
func (r VolumeGroupReconciler) getVolumeGroupContentSource(logger logr.Logger, req types.NamespacedName) (*volumegroupv1.VolumeGroupContentSource, error) {
	VGC := &volumegroupv1.VolumeGroupContent{}
	err := r.Client.Get(context.TODO(), req, VGC)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "VolumeGroupContent not found", "VolumeGroupContent Name", req.Name)
		}

		return nil, err
	}

	// Validate PVC in bound state
	if *VGC.Status.Ready != true {
		return nil, fmt.Errorf("VolumeGroupContentSource %q is not Ready", req.Name)
	}

	return VGC.Spec.Source, nil
}

// createVolumeGroupContent saves VolumeGroupContentSource on cluster.
func (r *VolumeGroupReconciler) createVolumeGroupContent(logger logr.Logger, instance *volumegroupv1.VolumeGroup, vgcObj *volumegroupv1.VolumeGroupClass, resp *volumegroup.Response, secretName string, secretNamespace string, groupCreationTime *metav1.Time, ready *bool) error {
	VGC := &volumegroupv1.VolumeGroupContent{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", instance.Name, "content"),
		},
		Spec: volumegroupv1.VolumeGroupContentSpec{
			VolumeGroupClassName: instance.Spec.VolumeGroupClassName,
			VolumeGroupRef: &corev1.ObjectReference{
				Kind:            instance.Kind,
				Namespace:       instance.Namespace,
				Name:            instance.Name,
				UID:             instance.UID,
				APIVersion:      instance.APIVersion,
				ResourceVersion: instance.ResourceVersion,
			},
			Source: &volumegroupv1.VolumeGroupContentSource{
				Driver:                vgcObj.Driver,
				VolumeGroupHandle:     resp.VolumeGroup.volume_group_id,
				VolumeGroupAttributes: resp.VolumeGroup.volume_group_context,
			},
			VolumeGroupSecretRef: &corev1.SecretReference{
				Name:      secretName,
				Namespace: secretNamespace,
			},
		},
		Status: volumegroupv1.VolumeGroupContentStatus{
			GroupCreationTime: groupCreationTime,
			PVList:            []corev1.PersistentVolume{},
			Ready:             ready,
		},
	}

	err := r.Client.Create(context.TODO(), VGC, nil)
	if err != nil {
		logger.Error(err, "VolumeGroupContent not found", "VolumeGroupContent Name")
		return err
	}

	// Validate PVC in bound state
	if *VGC.Status.Ready != true {
		return fmt.Errorf("VolumeGroupContentSource %q is not Ready", instance.Name)
	}

	return nil
}
