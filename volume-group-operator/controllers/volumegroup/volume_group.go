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

package volumegroup

import (
	"github.com/IBM/volume-group-operator/pkg/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VolumeGroup struct {
	Params CommonRequestParameters
}

type Response struct {
	Response interface{}
	Error    error
}

type CommonRequestParameters struct {
	Name          string
	VolumeGroupID string
	VolumeIds     []string
	Parameters    map[string]string
	Secrets       map[string]string
	VolumeGroup   client.VolumeGroup
}

func (r *VolumeGroup) Create() *Response {
	resp, err := r.Params.VolumeGroup.CreateVolumeGroup(
		r.Params.Name,
		r.Params.Secrets,
		r.Params.Parameters,
	)

	return &Response{Response: resp, Error: err}
}

func (r *VolumeGroup) Delete() *Response {
	resp, err := r.Params.VolumeGroup.DeleteVolumeGroup(
		r.Params.VolumeGroupID,
		r.Params.Secrets,
	)

	return &Response{Response: resp, Error: err}
}

func (r *VolumeGroup) Modify() *Response {
	resp, err := r.Params.VolumeGroup.ModifyVolumeGroupMembership(
		r.Params.VolumeGroupID,
		r.Params.VolumeIds,
		r.Params.Secrets,
	)

	return &Response{Response: resp, Error: err}
}

func (r *Response) HasKnownGRPCError(knownErrors []codes.Code) bool {
	if r.Error == nil {
		return false
	}

	s, ok := status.FromError(r.Error)
	if !ok {
		return false
	}

	for _, e := range knownErrors {
		if s.Code() == e {
			return true
		}
	}

	return false
}

func GetMessageFromError(err error) string {
	s, ok := status.FromError(err)
	if !ok {
		return err.Error()
	}

	return s.Message()
}