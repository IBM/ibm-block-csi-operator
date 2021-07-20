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
	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// Config is the wrapper for csiv1.Config type
type Config struct {
	*csiv1.Config
}

// New returns a wrapper for csiv1.Config
func New(c *csiv1.Config) *Config {
	return &Config{
		Config: c,
	}
}

// Unwrap returns the csiv1.Config object
func (c *Config) Unwrap() *csiv1.Config {
	return c.Config
}

func (c *Config) GetNodeAgentPodLabels() labels.Set {
	return labels.Set{
		"app.kubernetes.io/name": "ibm-node-agent",
	}
}

func (c *Config) GetNodeAgentImage() string {
	if c.Spec.NodeAgent.Tag == "" {
		return c.Spec.NodeAgent.Repository
	}
	return c.Spec.NodeAgent.Repository + ":" + c.Spec.NodeAgent.Tag
}
