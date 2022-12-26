package utils

import (
	"fmt"

	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/IBM/volume-group-operator/controllers/volumegroup"
	grpcClient "github.com/IBM/volume-group-operator/pkg/client"
	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ModifyVolumeGroup(logger logr.Logger, client client.Client, vg *volumegroupv1.VolumeGroup,
	vgClient grpcClient.VolumeGroup) error {
	params, err := generateModifyVolumeGroupParams(logger, client, vg, vgClient)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf(messages.ModifyVolumeGroup, params.VolumeGroupID, params.VolumeIds))
	volumeGroupRequest := volumegroup.NewVolumeGroupRequest(params)
	modifyVolumeGroupResponse := volumeGroupRequest.Modify()
	responseError := modifyVolumeGroupResponse.Error
	if responseError != nil {
		logger.Error(responseError, fmt.Sprintf(messages.FailedToModifyVolumeGroup, vg.Namespace, vg.Name))
		return responseError
	}
	logger.Info(fmt.Sprintf(messages.ModifiedVolumeGroup, params.VolumeGroupID))
	return nil
}
func generateModifyVolumeGroupParams(logger logr.Logger, client client.Client,
	vg *volumegroupv1.VolumeGroup, vgClient grpcClient.VolumeGroup) (volumegroup.CommonRequestParameters, error) {
	vgId, err := getVgId(logger, client, vg)
	if err != nil {
		return volumegroup.CommonRequestParameters{}, err
	}
	volumeIds, err := getPVCListVolumeIds(logger, client, vg.Status.PVCList)
	if err != nil {
		return volumegroup.CommonRequestParameters{}, err
	}
	secrets, err := getSecrets(logger, client, vg)
	if err != nil {
		return volumegroup.CommonRequestParameters{}, err
	}

	return volumegroup.CommonRequestParameters{
		Secrets:       secrets,
		VolumeGroup:   vgClient,
		VolumeGroupID: vgId,
		VolumeIds:     volumeIds,
	}, nil
}
func getSecrets(logger logr.Logger, client client.Client, vg *volumegroupv1.VolumeGroup) (map[string]string, error) {
	vgc, err := GetVolumeGroupClass(client, logger, *vg.Spec.VolumeGroupClassName)
	if err != nil {
		return nil, err
	}
	secrets, err := GetSecretDataFromClass(client, vgc, logger, vg)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}
