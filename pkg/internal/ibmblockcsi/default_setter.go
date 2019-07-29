/**
 * Copyright 2019 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ibmblockcsi

import (
	csiv1 "github.com/IBM/ibm-block-csi-driver-operator/pkg/apis/csi/v1"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/config"
)

// TODO: improve this function
// SetDefaults set defaults if omitted in spec, returns true means CR should be updated on cluster.
// Replace it with kubernetes native default setter when it is available.
// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#defaulting
func (c *IBMBlockCSI) SetDefaults() bool {
	changed := false

	if c.Spec == (csiv1.IBMBlockCSISpec{}) {
		c.Spec.Controller.Repository = config.ControllerRepository
		c.Spec.Controller.Tag = config.ControllerTag
		c.Spec.Node.Repository = config.NodeRepository
		c.Spec.Node.Tag = config.NodeTag

		changed = true
	}

	return changed
}
