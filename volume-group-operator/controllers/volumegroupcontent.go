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
func (r *VolumeGroupReconciler) createVolumeGroupContent(logger logr.Logger, instance *volumegroupv1.VolumeGroup, vgcObj *volumegroupv1.VolumeGroupContent) error {
	err := r.Client.Create(context.TODO(), vgcObj, nil)
	if err != nil {
		logger.Error(err, "VolumeGroupContent not found", "VolumeGroupContent Name")
		return err
	}

	// Validate PVC in bound state
	if *vgcObj.Status.Ready != true {
		return fmt.Errorf("VolumeGroupContentSource %q is not Ready", instance.Name)
	}

	return nil
}

func (r *VolumeGroupReconciler) generateVolumeGroupContent(instance *volumegroupv1.VolumeGroup, vgcObj *volumegroupv1.VolumeGroupClass, resp *volumegroup.Response, secretName string, secretNamespace string, groupCreationTime *metav1.Time, ready *bool) *volumegroupv1.VolumeGroupContent {
	return &volumegroupv1.VolumeGroupContent{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", instance.Name, "content"),
		},
		Spec:   r.generateVolumeGroupContentSpec(instance, vgcObj, resp, secretName, secretNamespace),
		Status: r.generateVolumeGroupContentStatus(groupCreationTime, ready),
	}
}

func (r *VolumeGroupReconciler) generateVolumeGroupContentStatus(groupCreationTime *metav1.Time, ready *bool) volumegroupv1.VolumeGroupContentStatus {
	return volumegroupv1.VolumeGroupContentStatus{
		GroupCreationTime: groupCreationTime,
		PVList:            []corev1.PersistentVolume{},
		Ready:             ready,
	}
}

func (r *VolumeGroupReconciler) generateVolumeGroupContentSpec(instance *volumegroupv1.VolumeGroup, vgcObj *volumegroupv1.VolumeGroupClass, resp *volumegroup.Response, secretName string, secretNamespace string) volumegroupv1.VolumeGroupContentSpec {
	return volumegroupv1.VolumeGroupContentSpec{
		VolumeGroupClassName: instance.Spec.VolumeGroupClassName,
		VolumeGroupRef:       r.generateObjectReference(instance),
		Source:               r.generateVolumeGroupContentSource(vgcObj, resp),
		VolumeGroupSecretRef: r.generateSecretReference(secretName, secretNamespace),
	}
}

func (r *VolumeGroupReconciler) generateObjectReference(instance *volumegroupv1.VolumeGroup) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		Kind:            instance.Kind,
		Namespace:       instance.Namespace,
		Name:            instance.Name,
		UID:             instance.UID,
		APIVersion:      instance.APIVersion,
		ResourceVersion: instance.ResourceVersion,
	}
}

func (r *VolumeGroupReconciler) generateSecretReference(secretName string, secretNamespace string) *corev1.SecretReference {
	return &corev1.SecretReference{
		Name:      secretName,
		Namespace: secretNamespace,
	}
}

func (r *VolumeGroupReconciler) generateVolumeGroupContentSource(vgcObj *volumegroupv1.VolumeGroupClass, resp *volumegroup.Response) *volumegroupv1.VolumeGroupContentSource {
	return &volumegroupv1.VolumeGroupContentSource{
		Driver:                vgcObj.Driver,
		VolumeGroupHandle:     resp.Response.VolumeGroup.VolumeGroupId,
		VolumeGroupAttributes: resp.Response.VolumeGroup.VolumeGroupContext,
	}
}
