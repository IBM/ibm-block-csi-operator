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

package utils

import (
	"context"
	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// getVolumeGroupClass get volume group class object from the subjected namespace and return the same.
func (r *ControllerUtils) getVolumeGroupClass(logger logr.Logger, vgcName string) (*volumegroupv1.VolumeGroupClass, error) {
	vgcObj := &volumegroupv1.VolumeGroupClass{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: vgcName, Namespace: "default"}, vgcObj)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "VolumeGroupClass not found", "VolumeGroupClass", vgcName)
		} else {
			logger.Error(err, "Got an unexpected error while fetching VolumeGroupClass", "VolumeGroupClass", vgcName)
		}

		return nil, err
	}

	return vgcObj, nil
}
