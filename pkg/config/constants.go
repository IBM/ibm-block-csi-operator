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
	APIGroup    = "csi.ibm.com"
	APIVersion  = "v1"
	Name        = "ibm-block-csi-operator"
	DriverName  = "block.csi.ibm.com"
	ProductName = "ibm-block-csi-driver"
	DeployPath  = "/deploy"
	Masterlabel = "node-role.kubernetes.io/master"

	ControllerRepository = "ibmcom/ibm-block-csi-controller-driver"
	NodeRepository       = "ibmcom/ibm-block-csi-node-driver"
	NodeAgentRepository  = "ibmcom/ibm-node-agent"

	ENVIscsiAgentPort = "ISCSI_AGENT_PORT"
	ENVEndpoint       = "ENDPOINT"
	ENVNodeName       = "NODE_NAME"
	ENVKubeVersion    = "KUBE_VERSION"

	CSINodeDriverRegistrar = "csi-node-driver-registrar"
	CSIProvisioner         = "csi-provisioner"
	CSIAttacher            = "csi-attacher"
	CSISnapshotter         = "csi-snapshotter"
	LivenessProbe          = "livenessprobe"

	ControllerSocketVolumeMountPath                       = "/var/lib/csi/sockets/pluginproxy/"
	NodeSocketVolumeMountPath                             = "/csi"
	ControllerLivenessProbeContainerSocketVolumeMountPath = "/csi"
	ControllerSocketPath                                  = "/var/lib/csi/sockets/pluginproxy/csi.sock"
	NodeSocketPath                                        = "/csi/csi.sock"
	NodeRegistrarSocketPath                               = "/var/lib/kubelet/plugins/block.csi.ibm.com/csi.sock"
	CSIEndpoint                                           = "unix:///var/lib/csi/sockets/pluginproxy/csi.sock"
	CSINodeEndpoint                                       = "unix:///csi/csi.sock"
)
