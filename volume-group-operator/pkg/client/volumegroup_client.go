/*
Copyright 2021.

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

package client

import (
	"context"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
)

type volumeGroupClient struct {
	client  csi.ControllerClient
	timeout time.Duration
}

// VolumeGroup holds the methods required for volume group.
type VolumeGroup interface {
	// CreateVolumeGroup RPC call to create the volume group.
	CreateVolumeGroup(name string, secrets, parameters map[string]string) (*csi.CreateVolumeGroupResponse, error)
	// DeleteVolumeGroup RPC call to delete the volume group.
	DeleteVolumeGroup(volumeGroupId string, secrets map[string]string) (*csi.DeleteVolumeGroupResponse, error)
	// ModifyVolumeGroupMembership RPC call to modify the volume group.
	ModifyVolumeGroupMembership(volumeGroupId string, volumeIds []string, secrets map[string]string) (*csi.ModifyVolumeGroupMembershipResponse, error)
}

// NewVolumeGroupClient returns VolumeGroup interface which has the RPC
// calls for volume group.
func NewVolumeGroupClient(cc *grpc.ClientConn, timeout time.Duration) VolumeGroup {
	return &volumeGroupClient{client: csi.NewControllerClient(cc), timeout: timeout}
}

// CreateVolumeGroup RPC call to create the volume group.
func (rc *volumeGroupClient) CreateVolumeGroup(name string, secrets, parameters map[string]string) (*csi.CreateVolumeGroupResponse, error) {
	req := &csi.CreateVolumeGroupRequest{
		Name:       name,
		Parameters: parameters,
		Secrets:    secrets,
	}

	createCtx, cancel := context.WithTimeout(context.Background(), rc.timeout)
	defer cancel()
	resp, err := rc.client.CreateVolumeGroup(createCtx, req)

	return resp, err
}

// DeleteVolumeGroup RPC call to delete the volume group.
func (rc *volumeGroupClient) DeleteVolumeGroup(volumeGroupId string, secrets map[string]string) (*csi.DeleteVolumeGroupResponse, error) {
	req := &csi.DeleteVolumeGroupRequest{
		VolumeGroupId: volumeGroupId,
		Secrets:       secrets,
	}

	createCtx, cancel := context.WithTimeout(context.Background(), rc.timeout)
	defer cancel()
	resp, err := rc.client.DeleteVolumeGroup(createCtx, req)

	return resp, err
}

// ModifyVolumeGroupMembership RPC call to modify the volume group.
func (rc *volumeGroupClient) ModifyVolumeGroupMembership(volumeGroupId string, volumeIds []string, secrets map[string]string) (*csi.ModifyVolumeGroupMembershipResponse, error) {
	req := &csi.ModifyVolumeGroupMembershipRequest{
		VolumeGroupId: volumeGroupId,
		VolumeIds:     volumeIds,
		Secrets:       secrets,
	}

	createCtx, cancel := context.WithTimeout(context.Background(), rc.timeout)
	defer cancel()
	resp, err := rc.client.ModifyVolumeGroupMembership(createCtx, req)

	return resp, err
}
