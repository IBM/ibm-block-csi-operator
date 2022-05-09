/**
 * Copyright 2022 IBM Corp.
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

package hostdefinition

import (
	"path"

	corev1 "k8s.io/api/core/v1"

	"github.com/IBM/ibm-block-csi-operator/pkg/config"
)

func (c *HostDefinition) SetDefaults() bool {

	c.setDefaultForNilSliceFields()

	if c.isUnofficialRepo(c.Spec.HostDefinition.Repository) {
		return false
	}
	return c.setDefaults()
}

func (c *HostDefinition) setDefaultForNilSliceFields() {
	if c.Spec.ImagePullSecrets == nil {
		c.Spec.ImagePullSecrets = []string{}
	}
	if c.Spec.HostDefinition.Tolerations == nil {
		c.Spec.HostDefinition.Tolerations = []corev1.Toleration{}
	}
}

func (c *HostDefinition) isUnofficialRepo(repo string) bool {
	if repo != "" {
		var registryUsername = path.Dir(repo)
		if !config.OfficialRegistriesUsernames.Has(registryUsername) {
			return true
		}
	}
	return false
}

func (c *HostDefinition) setDefaults() bool {
	var changed = false

	if c.Spec.HostDefinition.Repository != config.DefaultHostDefinitionCr.Spec.HostDefinition.Repository ||
		c.Spec.HostDefinition.Tag != config.DefaultHostDefinitionCr.Spec.HostDefinition.Tag {
		c.Spec.HostDefinition.Repository = config.DefaultHostDefinitionCr.Spec.HostDefinition.Repository
		c.Spec.HostDefinition.Tag = config.DefaultHostDefinitionCr.Spec.HostDefinition.Tag

		changed = true
	}

	return changed
}
