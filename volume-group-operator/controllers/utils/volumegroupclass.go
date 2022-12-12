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
	"fmt"
	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
)

func (r *ControllerUtils) GetVolumeGroupClass(logger logr.Logger, vgcName string) (*volumegroupv1.VolumeGroupClass, error) {
	vgcObjl := &volumegroupv1.VolumeGroupClassList{}
	err := r.Client.List(context.TODO(), vgcObjl, &client.ListOptions{Raw: &metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name=%s", vgcName)}})
	if err != nil {
		logger.Error(err, "Got an unexpected error while fetching VolumeGroupClass", "VolumeGroupClass", vgcName)
		return nil, err
	}
	var Ierr error
	items := vgcObjl.Items
	if len(items) != 1 {
		if len(items) == 0 {
			Ierr = fmt.Errorf("VolumeGroupClass %s not found", vgcName)
			return nil, Ierr
		}
		Ierr = fmt.Errorf("got an unexpected amount of object while fetching VolumeGroupClass %s", vgcName)
		return nil, Ierr
	}
	vgcObj := &items[0]
	return vgcObj, nil
}
