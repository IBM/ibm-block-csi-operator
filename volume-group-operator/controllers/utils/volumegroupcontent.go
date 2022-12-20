package utils

import (
	"context"
	"fmt"

	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/IBM/volume-group-operator/controllers/volumegroup"
	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetVolumeGroupContent(client client.Client, logger logr.Logger, vg *volumegroupv1.VolumeGroup) (*volumegroupv1.VolumeGroupContent, error) {
	logger.Info(fmt.Sprintf(messages.GetVolumeGroupContentOfVolumeGroup, vg.Name, vg.Namespace))
	vgc := &volumegroupv1.VolumeGroupContent{}
	VolumeGroupContentName := *vg.Spec.Source.VolumeGroupContentName
	namespacedVGC := types.NamespacedName{Name: VolumeGroupContentName, Namespace: vg.Namespace}
	err := client.Get(context.TODO(), namespacedVGC, vgc)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "VolumeGroupContent not found", "VolumeGroupContent Name", VolumeGroupContentName)
		}
		return nil, err
	}

	return vgc, nil
}

func CreateVolumeGroupContent(client client.Client, logger logr.Logger, instance *volumegroupv1.VolumeGroup, vgcObj *volumegroupv1.VolumeGroupContent) error {
	err := client.Create(context.TODO(), vgcObj)
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

func GenerateVolumeGroupContent(vgname string, instance *volumegroupv1.VolumeGroup, vgcObj *volumegroupv1.VolumeGroupClass, resp *volumegroup.Response) *volumegroupv1.VolumeGroupContent {
	secretName, secretNamespace := GetSecretNameAndNamespace(vgcObj)
	return &volumegroupv1.VolumeGroupContent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vgname,
			Namespace: instance.Namespace,
		},
		Spec: generateVolumeGroupContentSpec(instance, vgcObj, resp, secretName, secretNamespace),
	}
}

func UpdateVolumeGroupContentStatus(client client.Client, logger logr.Logger, vgc *volumegroupv1.VolumeGroupContent, groupCreationTime *metav1.Time, ready bool) error {
	vgc.Status = generateVolumeGroupContentStatus(vgc, groupCreationTime, ready)
	if err := UpdateObjectStatus(client, vgc); err != nil {
		logger.Error(err, "failed to update status")
		return err
	}
	return nil
}

func generateVolumeGroupContentStatus(vgc *volumegroupv1.VolumeGroupContent, groupCreationTime *metav1.Time, ready bool) volumegroupv1.VolumeGroupContentStatus {
	return volumegroupv1.VolumeGroupContentStatus{
		GroupCreationTime: groupCreationTime,
		PVList:            vgc.Status.PVList,
		Ready:             &ready,
	}
}

func generateVolumeGroupContentSpec(instance *volumegroupv1.VolumeGroup, vgcObj *volumegroupv1.VolumeGroupClass,
	resp *volumegroup.Response, secretName string, secretNamespace string) volumegroupv1.VolumeGroupContentSpec {
	return volumegroupv1.VolumeGroupContentSpec{
		VolumeGroupClassName: instance.Spec.VolumeGroupClassName,
		VolumeGroupRef:       generateObjectReference(instance),
		Source:               generateVolumeGroupContentSource(vgcObj, resp),
		VolumeGroupSecretRef: generateSecretReference(secretName, secretNamespace),
	}
}

func generateObjectReference(instance *volumegroupv1.VolumeGroup) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		Kind:            instance.Kind,
		Namespace:       instance.Namespace,
		Name:            instance.Name,
		UID:             instance.UID,
		APIVersion:      instance.APIVersion,
		ResourceVersion: instance.ResourceVersion,
	}
}

func generateSecretReference(secretName string, secretNamespace string) *corev1.SecretReference {
	return &corev1.SecretReference{
		Name:      secretName,
		Namespace: secretNamespace,
	}
}

func generateVolumeGroupContentSource(vgcObj *volumegroupv1.VolumeGroupClass, resp *volumegroup.Response) *volumegroupv1.VolumeGroupContentSource {
	CreateVolumeGroupResponse := resp.Response.(*csi.CreateVolumeGroupResponse)
	return &volumegroupv1.VolumeGroupContentSource{
		Driver:                vgcObj.Driver,
		VolumeGroupHandle:     CreateVolumeGroupResponse.VolumeGroup.VolumeGroupId,
		VolumeGroupAttributes: CreateVolumeGroupResponse.VolumeGroup.VolumeGroupContext,
	}
}

func RemovePVFromVGC(logger logr.Logger, client client.Client, pv *corev1.PersistentVolume, vgc *volumegroupv1.VolumeGroupContent) error {
	logger.Info(fmt.Sprintf(messages.RemovePersistentVolumeFromVolumeGroupContent,
		pv.Namespace, pv.Name, vgc.Namespace, vgc.Name))
	vgc.Status.PVList = removePersistentVolumeFromVolumeGroupContentPVList(pv, vgc.Status.PVList)
	err := client.Status().Update(context.TODO(), vgc)
	if err != nil {
		logger.Error(err, fmt.Sprintf(messages.FailedToRemovePersistentVolumeFromVolumeGroupContent,
			pv.Namespace, pv.Name, vgc.Namespace, vgc.Name))
		return err
	}
	logger.Info(fmt.Sprintf(messages.RemovedPersistentVolumeFromVolumeGroupContent,
		pv.Namespace, pv.Name, vgc.Namespace, vgc.Name))
	return nil
}

func removePersistentVolumeFromVolumeGroupContentPVList(pv *corev1.PersistentVolume,
	pvListInVGC []corev1.PersistentVolume) []corev1.PersistentVolume {
	for index, pvcFromList := range pvListInVGC {
		if pvcFromList.Name == pv.Name && pvcFromList.Namespace == pv.Namespace {
			pvListInVGC = removeByIndexFromPersistentVolumeList(pvListInVGC, index)
			return pvListInVGC
		}
	}
	return pvListInVGC
}
