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

func AddMatchingPVToMatchingVGC(logger logr.Logger, client client.Client,
	pvc *corev1.PersistentVolumeClaim, vg *volumegroupv1.VolumeGroup) error {
	pv, err := GetPVFromPVC(logger, client, pvc)
	if err != nil {
		return err
	}
	vgc, err := GetVolumeGroupContent(client, logger, vg)
	if err != nil {
		return err
	}

	if pv != nil {
		return addPVToVGC(logger, client, pv, vgc)
	}
	return nil
}

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

func CreateVolumeGroupContent(client client.Client, logger logr.Logger, vgcObj *volumegroupv1.VolumeGroupContent) error {
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

func UpdateVolumeGroupContentStatus(client client.Client, logger logr.Logger, vgc *volumegroupv1.VolumeGroupContent, groupCreationTime *metav1.Time, ready bool) error {
	updateVolumeGroupContentStatusFields(vgc, groupCreationTime, ready)
	if err := UpdateObjectStatus(client, vgc); err != nil {
		logger.Error(err, "failed to update status")
		return err
	}
	return nil
}

func updateVolumeGroupContentStatusFields(vgc *volumegroupv1.VolumeGroupContent, groupCreationTime *metav1.Time, ready bool) {
	vgc.Status.GroupCreationTime = groupCreationTime
	vgc.Status.Ready = &ready
}

func GenerateVolumeGroupContent(vgname string, instance *volumegroupv1.VolumeGroup, vgcObj *volumegroupv1.VolumeGroupClass, resp *volumegroup.Response, secretName string, secretNamespace string) *volumegroupv1.VolumeGroupContent {
	return &volumegroupv1.VolumeGroupContent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vgname,
			Namespace: instance.Namespace,
		},
		Spec: generateVolumeGroupContentSpec(instance, vgcObj, resp, secretName, secretNamespace),
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
	vgc.Status.PVList = removeFromPVList(pv, vgc.Status.PVList)
	err := client.Status().Update(context.TODO(), vgc)
	if err != nil {
		logger.Error(err, fmt.Sprintf(messages.FailedToRemovePersistentVolumeFromVolumeGroupContent,
			pv.Name, vgc.Namespace, vgc.Name))
		return err
	}
	logger.Info(fmt.Sprintf(messages.RemovedPersistentVolumeFromVolumeGroupContent,
		pv.Name, vgc.Namespace, vgc.Name))
	return nil
}

func removeFromPVList(pv *corev1.PersistentVolume, pvList []corev1.PersistentVolume) []corev1.PersistentVolume {
	for index, pvFromList := range pvList {
		if pvFromList.Name == pv.Name && pvFromList.Namespace == pv.Namespace {
			pvList = removeByIndexFromPersistentVolumeList(pvList, index)
			return pvList
		}
	}
	return pvList
}

func addPVToVGC(logger logr.Logger, client client.Client, pv *corev1.PersistentVolume,
	vgc *volumegroupv1.VolumeGroupContent) error {
	logger.Info(fmt.Sprintf(messages.AddPersistentVolumeToVolumeGroupContent,
		pv.Name, vgc.Namespace, vgc.Name))
	vgc.Status.PVList = appendPersistentVolume(vgc.Status.PVList, *pv)
	err := client.Status().Update(context.TODO(), vgc)
	if err != nil {
		logger.Error(err, fmt.Sprintf(messages.FailedToAddPersistentVolumeToVolumeGroupContent,
			pv.Name, vgc.Namespace, vgc.Name))
		return err
	}
	logger.Info(fmt.Sprintf(messages.AddedPersistentVolumeToVolumeGroupContent,
		pv.Name, vgc.Namespace, vgc.Name))
	return nil
}

func appendPersistentVolume(pvListInVGC []corev1.PersistentVolume, pv corev1.PersistentVolume) []corev1.PersistentVolume {
	for _, pvFromList := range pvListInVGC {
		if pvFromList.Name == pv.Name {
			return pvListInVGC
		}
	}
	pvListInVGC = append(pvListInVGC, pv)
	return pvListInVGC
}
