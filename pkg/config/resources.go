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
	"fmt"
	"strings"
)

// ResourceName is the type for aliasing resources that will be created.
type ResourceName string

func (rn ResourceName) String() string {
	return string(rn)
}

const (
	CSIController                         ResourceName = "csi-controller"
	CSINode                               ResourceName = "csi-node"
	HostDefiner                           ResourceName = "hostdefiner"
	NodeAgent                             ResourceName = "ibm-node-agent"
	CSIControllerServiceAccount           ResourceName = "csi-controller-sa"
	CSINodeServiceAccount                 ResourceName = "csi-node-sa"
	HostDefinerServiceAccount             ResourceName = "hostdefiner-sa"
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
	CSIVolumeGroupClusterRole             ResourceName = "csi-volume-group-clusterrole"
	CSIVolumeGroupClusterRoleBinding      ResourceName = "csi-volume-group-clusterrolebinding"
	CSIControllerSCCClusterRole           ResourceName = "csi-controller-scc-clusterrole"
	CSIControllerSCCClusterRoleBinding    ResourceName = "csi-controller-scc-clusterrolebinding"
	CSINodeSCCClusterRole                 ResourceName = "csi-node-scc-clusterrole"
	CSINodeSCCClusterRoleBinding          ResourceName = "csi-node-scc-clusterrolebinding"
	HostDefinerClusterRole                ResourceName = "hostdefiner-clusterrole"
	HostDefinerClusterRoleBinding         ResourceName = "hostdefiner-clusterrolebinding"
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
		if strings.Contains(name.String(), "hostdefiner") {
			return fmt.Sprintf("%s-%s", driverName, name.String())
		}
		return fmt.Sprintf("%s-%s", driverName, name)
	}
}

func GetControllerPodName(statefulSetName string) string {
	return fmt.Sprintf("%s-0", statefulSetName)
}
