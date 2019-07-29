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

// Add a field here if it never changes, if it changes over time, put it to settings.go
const (
	APIGroup          = "csi.ibm.com"
	APIVersion        = "v1"
	Name              = "ibm-block-csi-operator"
	DriverName        = "ibm-block-csi-driver"
	ProductName       = "ibm-block-csi"
	DeployPath        = "/deploy"
	ENVIscsiAgentPort = "ISCSI_AGENT_PORT"
	ENVEndpoint       = "ENDPOINT"

	ControllerRepository = "ibmcom/ibm-block-csi-controller-driver"
	NodeRepository       = "ibmcom/ibm-block-csi-node-driver"

	ControllerSocketVolumeMountPath                       = "/var/lib/csi/sockets/pluginproxy/"
	NodeSocketVolumeMountPath                             = "/csi"
	ControllerLivenessProbeContainerSocketVolumeMountPath = "/csi"
	ControllerSocketPath                                  = "/var/lib/csi/sockets/pluginproxy/csi.sock"
	NodeSocketPath                                        = "/csi/csi.sock"
	NodeRegistrarSocketPath                               = "/var/lib/kubelet/plugins/ibm-block-csi-driver/csi.sock"
	CSIEndpoint                                           = "unix:///var/lib/csi/sockets/pluginproxy/csi.sock"
	CSINodeEndpoint                                       = "unix:///csi/csi.sock"
)
