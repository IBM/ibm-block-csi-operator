package controllers

import (
	"context"
	"fmt"
	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/IBM/volume-group-operator/controllers/volumegroup"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// getVolumeGroupContent get VolumeGroupContent object from the request.
func (r VolumeGroupReconciler) getVolumeGroupContent(logger logr.Logger, instance *volumegroupv1.VolumeGroup) (*volumegroupv1.VolumeGroupContent, error) {
	VGC := &volumegroupv1.VolumeGroupContent{}
	VolumeGroupContentName := *instance.Spec.Source.VolumeGroupContentName
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: VolumeGroupContentName, Namespace: instance.Namespace}, VGC)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "VolumeGroupContent not found", "VolumeGroupContent Name", VolumeGroupContentName)
		}

		return nil, err
	}

	return VGC, nil
}

// createVolumeGroupContent saves VolumeGroupContentSource on cluster.
func (r *VolumeGroupReconciler) createVolumeGroupContent(logger logr.Logger, instance *volumegroupv1.VolumeGroup, vgcObj *volumegroupv1.VolumeGroupContent) error {
	err := r.Client.Create(context.TODO(), vgcObj)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Info("VolumeGroupContent is already exists")
			return nil
		}
		logger.Error(err, "VolumeGroupContent creation failed", "VolumeGroupContent Name")
		return err
	}

	return nil
}

func (r *VolumeGroupReconciler) generateVolumeGroupContent(instance *volumegroupv1.VolumeGroup, vgcObj *volumegroupv1.VolumeGroupClass, resp *volumegroup.Response, secretName string, secretNamespace string, groupCreationTime *metav1.Time, ready *bool) *volumegroupv1.VolumeGroupContent {
	return &volumegroupv1.VolumeGroupContent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", instance.Name, "content"),
			Namespace: instance.Namespace,
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
	CreateVolumeGroupResponse := resp.Response.(*csi.CreateVolumeGroupResponse)
	return &volumegroupv1.VolumeGroupContentSource{
		Driver:                vgcObj.Driver,
		VolumeGroupHandle:     CreateVolumeGroupResponse.VolumeGroup.VolumeGroupId,
		VolumeGroupAttributes: CreateVolumeGroupResponse.VolumeGroup.VolumeGroupContext,
	}
}
