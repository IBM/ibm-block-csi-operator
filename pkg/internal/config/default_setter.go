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

package config

import (
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
)

// SetDefaults set defaults if omitted in spec, returns true means CR should be updated on cluster.
// Replace it with kubernetes native default setter when it is available.
// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#defaulting
func (c *Config) SetDefaults() bool {
	changed := false

	// if node agent repository is empty
	if c.Spec.NodeAgent.Repository == "" {
		// repository is mandatory, tag is optional, but if repository is not set
		// and tag is set, the tag will be overrided to the default one.
		c.Spec.NodeAgent.Repository = config.NodeAgentRepository
		c.Spec.NodeAgent.Tag = config.NodeAgentTag

		changed = true
	}

	// if node agent port is empty
	if c.Spec.NodeAgent.Port == "" {
		c.Spec.NodeAgent.Port = config.NodeAgentPort
		changed = true
	}

	return changed
}
