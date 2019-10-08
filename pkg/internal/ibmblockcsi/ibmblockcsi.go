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
	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	csiversion "github.com/IBM/ibm-block-csi-operator/version"
	"k8s.io/apimachinery/pkg/labels"
)

// IBMBlockCSI is the wrapper for csiv1.IBMBlockCSI type
type IBMBlockCSI struct {
	*csiv1.IBMBlockCSI
	ServerVersion string
}

// New returns a wrapper for csiv1.IBMBlockCSI
func New(c *csiv1.IBMBlockCSI, serverVersion string) *IBMBlockCSI {
	return &IBMBlockCSI{
		IBMBlockCSI:   c,
		ServerVersion: serverVersion,
	}
}

// Unwrap returns the csiv1.IBMBlockCSI object
func (c *IBMBlockCSI) Unwrap() *csiv1.IBMBlockCSI {
	return c.IBMBlockCSI
}

// GetAnnotations returns all the annotations to be set on all resources
func (c *IBMBlockCSI) GetAnnotations() labels.Set {
	labels := labels.Set{
		"app.kubernetes.io/name":       config.ProductName,
		"app.kubernetes.io/instance":   c.Name,
		"app.kubernetes.io/version":    csiversion.Version,
		"app.kubernetes.io/managed-by": config.Name,
	}

	if c.Annotations != nil {
		for k, v := range c.Annotations {
			labels[k] = v
		}
	}

	return labels
}

func (c *IBMBlockCSI) GetComponentAnnotations(component string) labels.Set {
	return labels.Set{
		"app.kubernetes.io/component": component,
	}
}

func (c *IBMBlockCSI) GetCSIControllerComponentAnnotations() labels.Set {
	return c.GetComponentAnnotations(config.CSIController.String())
}

func (c *IBMBlockCSI) GetCSINodeComponentAnnotations() labels.Set {
	return c.GetComponentAnnotations(config.CSINode.String())
}

func (c *IBMBlockCSI) GetCSIControllerAnnotations() labels.Set {
	labels := c.GetLabels()
	for k, v := range c.GetCSIControllerComponentAnnotations() {
		labels[k] = v
	}
	return labels
}

func (c *IBMBlockCSI) GetCSINodeAnnotations() labels.Set {
	labels := c.GetLabels()
	for k, v := range c.GetCSINodeComponentAnnotations() {
		labels[k] = v
	}
	return labels
}

func (c *IBMBlockCSI) GetCSIControllerImage() string {
	if c.Spec.Controller.Tag == "" {
		return c.Spec.Controller.Repository
	}
	return c.Spec.Controller.Repository + ":" + c.Spec.Controller.Tag
}

func (c *IBMBlockCSI) GetCSINodeImage() string {
	if c.Spec.Node.Tag == "" {
		return c.Spec.Node.Repository
	}
	return c.Spec.Node.Repository + ":" + c.Spec.Node.Tag
}
