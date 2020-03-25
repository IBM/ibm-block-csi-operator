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
	"strings"

	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	corev1 "k8s.io/api/core/v1"
)

// SetDefaults set defaults if omitted in spec, returns true means CR should be updated on cluster.
// Replace it with kubernetes native default setter when it is available.
// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#defaulting
func (c *IBMBlockCSI) SetDefaults(platform string) bool {
	// repository is mandatory, tag is optional, but if repository is not set
	// and tag is set, the tag will be overrided to the default one.

	changed := c.clearPreviousVersion(platform)

	// if controller is empty
	if c.Spec.Controller.Repository == "" {
		regAndTag := strings.Split(c.GetDefaultImageByPlatformAndName(platform, config.ControllerImage), ":")
		c.Spec.Controller.Repository = regAndTag[0]
		c.Spec.Controller.Tag = regAndTag[1]

		changed = true
	}

	// if node is empty
	if c.Spec.Node.Repository == "" {
		regAndTag := strings.Split(c.GetDefaultImageByPlatformAndName(platform, config.NodeImage), ":")
		c.Spec.Node.Repository = regAndTag[0]
		c.Spec.Node.Tag = regAndTag[1]

		changed = true
	}

	if sidecarChanged := c.clearSidecarPreviousVersion(platform); sidecarChanged {
		changed = sidecarChanged
	}
	if sidecarChanged := c.setDefualtSidecars(platform); sidecarChanged {
		changed = sidecarChanged
	}

	c.setDefaultSliceFields()

	return changed
}
func (c *IBMBlockCSI) setDefaultSliceFields() {
	if c.Spec.ImagePullSecrets == nil {
		c.Spec.ImagePullSecrets = []string{}
	}
	if c.Spec.Controller.Tolerations == nil {
		c.Spec.Controller.Tolerations = []corev1.Toleration{}
	}
	if c.Spec.Node.Tolerations == nil {
		c.Spec.Node.Tolerations = []corev1.Toleration{}
	}
}

// clearPreviousVersion clears the previous version of controller and node
// images to empty value, so that they will be updated to the latest version.
// For example: If current controller image is ibmcom/ibm-block-csi-controller-driver:1.0.0,
// it will be cleared and updated to ibmcom/ibm-block-csi-controller-driver:1.1.0 in setDefault().
func (c *IBMBlockCSI) clearPreviousVersion(platform string) bool {
	changed := false

	// if controller is a replace version
	if config.GetReplaceVersions(platform, config.ControllerImage).Has(c.GetCSIControllerImage()) {
		c.Spec.Controller.Repository = ""
		c.Spec.Controller.Tag = ""
		changed = true
	}

	// if node is a replace version
	if config.GetReplaceVersions(platform, config.NodeImage).Has(c.GetCSINodeImage()) {
		c.Spec.Node.Repository = ""
		c.Spec.Node.Tag = ""
		changed = true
	}

	return changed
}

func (c *IBMBlockCSI) clearSidecarPreviousVersion(platform string) bool {
	changed := false
	var updated []csiv1.CSISidecar

	for _, sidecar := range c.Spec.Sidecars {
		if config.GetReplaceVersions(platform, sidecar.Name).Has(c.GetSidecarImageByName(sidecar.Name)) {
			sidecar.Repository = ""
			sidecar.Tag = ""
			changed = true
		}
		updated = append(updated, sidecar)
	}
	c.Spec.Sidecars = updated
	return changed
}

func (c *IBMBlockCSI) setDefualtSidecars(platform string) bool {
	changed := false
	var sidecars []csiv1.CSISidecar

	for _, name := range c.GetSidecarNames() {
		sidecar := c.GetSidecarByName(name)
		if sidecar == nil || sidecar.Repository == "" {
			regAndTag := strings.Split(c.GetDefaultImageByPlatformAndName(platform, name), ":")
			sidecar = &csiv1.CSISidecar{
				Name:            name,
				Repository:      regAndTag[0],
				Tag:             regAndTag[1],
				ImagePullPolicy: corev1.PullIfNotPresent,
			}
			changed = true
		}
		sidecars = append(sidecars, *sidecar)

	}
	c.Spec.Sidecars = sidecars
	return changed
}
