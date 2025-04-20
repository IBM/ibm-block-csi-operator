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

package crutils

import (
	"path"

	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	corev1 "k8s.io/api/core/v1"
)

// SetDefaults set defaults if omitted in spec, returns true means CR should be updated on cluster.
// Replace it with kubernetes native default setter when it is available.
// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#defaulting
func (c *IBMBlockCSI) SetDefaults() bool {

	c.setDefaultForNilSliceFields()

	if c.isAnyUnofficialRepo() {
		return false
	}
	return c.setDefaults()
}

func (c *IBMBlockCSI) isAnyUnofficialRepo() bool {

	if c.isUnofficialRepo(c.Spec.Controller.Repository) {
		return true
	}

	if c.isUnofficialRepo(c.Spec.Node.Repository) {
		return true
	}

	for _, sidecar := range c.Spec.Sidecars {
		if c.isUnofficialRepo(sidecar.Repository) {
			return true
		}
	}
	return false
}

func (c *IBMBlockCSI) isUnofficialRepo(repo string) bool {
	if repo != "" {
		var registryUsername = path.Dir(repo)
		if !config.OfficialRegistriesUsernames.Has(registryUsername) {
			return true
		}
	}
	return false
}

func (c *IBMBlockCSI) setDefaults() bool {
	var changed = false

	if c.Spec.Controller.Repository != config.DefaultIBMBlockCSICr.Spec.Controller.Repository ||
		c.Spec.Controller.Tag != config.DefaultIBMBlockCSICr.Spec.Controller.Tag {
		c.Spec.Controller.Repository = config.DefaultIBMBlockCSICr.Spec.Controller.Repository
		c.Spec.Controller.Tag = config.DefaultIBMBlockCSICr.Spec.Controller.Tag

		changed = true
	}

	if c.Spec.Node.Repository != config.DefaultIBMBlockCSICr.Spec.Node.Repository ||
		c.Spec.Node.Tag != config.DefaultIBMBlockCSICr.Spec.Node.Tag {
		c.Spec.Node.Repository = config.DefaultIBMBlockCSICr.Spec.Node.Repository
		c.Spec.Node.Tag = config.DefaultIBMBlockCSICr.Spec.Node.Tag

		changed = true
	}

	changed = c.setDefaultSidecars() || changed

	return changed
}

func (c *IBMBlockCSI) setDefaultForNilSliceFields() {
	if c.Spec.ImagePullSecrets == nil {
		c.Spec.ImagePullSecrets = []string{}
	}
	if c.Spec.Controller.Tolerations == nil {
		c.Spec.Controller.Tolerations = []corev1.Toleration{}
	}
	if c.Spec.Node.Tolerations == nil {
		c.Spec.Node.Tolerations = []corev1.Toleration{}
	}
	if c.Spec.Node.MemoryRequirements == nil {
		c.Spec.Node.MemoryRequirements = "40m,1000m,40Mi,400Mi"
	}
	if c.Spec.EnableCallHome == "" {
		c.Spec.EnableCallHome = "true"
	}
}

func (c *IBMBlockCSI) setDefaultSidecars() bool {
	var change = false
	var defaultSidecars = config.DefaultIBMBlockCSICr.Spec.Sidecars

	if len(defaultSidecars) == len(c.Spec.Sidecars) {
		for _, sidecar := range c.Spec.Sidecars {
			if defaultSidecar, found := config.DefaultSidecarsByName[sidecar.Name]; found {
				if sidecar != defaultSidecar {
					change = true
				}
			} else {
				change = true
			}
		}
	} else {
		change = true
	}

	if change {
		c.Spec.Sidecars = defaultSidecars
	}

	return change
}
