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

import "fmt"

// ResourceName is the type for aliasing resources that will be created.
type ResourceName string

func (rn ResourceName) String() string {
	return string(rn)
}

const (
	CSIController                         ResourceName = "csi-controller"
	CSINode                               ResourceName = "csi-node"
	CallHome                              ResourceName = "call-home"
	NodeAgent                             ResourceName = "ibm-node-agent"
	CSIControllerServiceAccount           ResourceName = "csi-controller-sa"
	CSINodeServiceAccount                 ResourceName = "csi-node-sa"
	CallHomeServiceAccount                ResourceName = "call-home-sa"
	ExternalProvisionerClusterRole        ResourceName = "external-provisioner-clusterrole"
	ExternalProvisionerClusterRoleBinding ResourceName = "external-provisioner-clusterrolebinding"
	ExternalAttacherClusterRole           ResourceName = "external-attacher-clusterrole"
	ExternalAttacherClusterRoleBinding    ResourceName = "external-attacher-clusterrolebinding"
	ExternalSnapshotterClusterRole        ResourceName = "external-snapshotter-clusterrole"
	ExternalSnapshotterClusterRoleBinding ResourceName = "external-snapshotter-clusterrolebinding"
	ExternalResizerClusterRole            ResourceName = "external-resizer-clusterrole"
	ExternalResizerClusterRoleBinding     ResourceName = "external-resizer-clusterrolebinding"
	CSIAddonsReplicatorClusterRole        ResourceName = "csi-addons-replicator-clusterrole"
	CSIAddonsReplicatorClusterRoleBinding ResourceName = "csi-addons-replicator-clusterrolebinding"
	CSIControllerSCCClusterRole           ResourceName = "csi-controller-scc-clusterrole"
	CSIControllerSCCClusterRoleBinding    ResourceName = "csi-controller-scc-clusterrolebinding"
	CSINodeSCCClusterRole                 ResourceName = "csi-node-scc-clusterrole"
	CallHomeSCCClusterRole                ResourceName = "call-home-scc-clusterrole"
	CSINodeSCCClusterRoleBinding          ResourceName = "csi-node-scc-clusterrolebinding"
	CallHomeSCCClusterRoleBinding         ResourceName = "call-home-scc-clusterrolebinding"
)

// GetNameForResource returns the name of a resource for a CSI driver
func GetNameForResource(name ResourceName, driverName string) string {
	switch name {
	case CSIController:
		return fmt.Sprintf("%s-controller", driverName)
	case CSINode:
		return fmt.Sprintf("%s-node", driverName)
	case CSIControllerServiceAccount:
		return fmt.Sprintf("%s-controller-sa", driverName)
	case CSINodeServiceAccount:
		return fmt.Sprintf("%s-node-sa", driverName)
	default:
		return fmt.Sprintf("%s-%s", driverName, name)
	}
}
