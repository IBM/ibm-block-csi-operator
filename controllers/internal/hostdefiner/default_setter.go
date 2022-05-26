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

package hostdefiner

import (
	"path"

	corev1 "k8s.io/api/core/v1"

	"github.com/IBM/ibm-block-csi-operator/pkg/config"
)

func (c *HostDefiner) SetDefaults() bool {

	c.setDefaultForNilSliceFields()

	if c.isUnofficialRepo(c.Spec.HostDefiner.Repository) {
		return false
	}
	return c.setDefaults()
}

func (c *HostDefiner) setDefaultForNilSliceFields() {
	if c.Spec.ImagePullSecrets == nil {
		c.Spec.ImagePullSecrets = []string{}
	}
	if c.Spec.HostDefiner.Tolerations == nil {
		c.Spec.HostDefiner.Tolerations = []corev1.Toleration{}
	}
}

func (c *HostDefiner) isUnofficialRepo(repo string) bool {
	if repo != "" {
		var registryUsername = path.Dir(repo)
		if !config.OfficialRegistriesUsernames.Has(registryUsername) {
			return true
		}
	}
	return false
}

func (c *HostDefiner) setDefaults() bool {
	var changed = false

	if c.Spec.HostDefiner.Repository != config.DefaultHostDefinerCr.Spec.HostDefiner.Repository ||
		c.Spec.HostDefiner.Tag != config.DefaultHostDefinerCr.Spec.HostDefiner.Tag {
		c.Spec.HostDefiner.Repository = config.DefaultHostDefinerCr.Spec.HostDefiner.Repository
		c.Spec.HostDefiner.Tag = config.DefaultHostDefinerCr.Spec.HostDefiner.Tag

		changed = true
	}

	return changed
}
