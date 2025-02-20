/**
 * Copyright 2025 IBM Corp.
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

package tests

import (
	csiv1 "github.com/IBM/ibm-block-csi-operator/api/v1"
	clustersyncer "github.com/IBM/ibm-block-csi-operator/controllers/syncer"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"k8s.io/apimachinery/pkg/types"
)

var (
	nodeContainerName        = clustersyncer.NodeContainerName
	controllerContainerName  = clustersyncer.ControllerContainerName
	hostDefinerContainerName = clustersyncer.HostDefinerContainerName
	controllerByName         map[string]csiv1.IBMBlockCSIControllerSpec
	nodeByName               map[string]csiv1.IBMBlockCSINodeSpec
	hostDefinerByName        map[string]csiv1.IBMBlockHostDefinerSpec
)

func GetHostDefinerImagesByName(hostDefinerCr csiv1.HostDefiner) map[string]string {
	containersImages := make(map[string]string)
	setHostDefinerDeploymentImageByName(hostDefinerCr)

	containersImages = addHostDefinerDeploymentImageToContainersImagesMap(containersImages, hostDefinerByName)
	return containersImages
}

func GetImagesByName(defaultCr csiv1.IBMBlockCSI, sidecarsByName map[string]csiv1.CSISidecar) map[string]string {
	containersImages := make(map[string]string)
	setControllerImageByName(defaultCr)
	setNodeImageByName(defaultCr)

	containersImages = addImagesByNameFromYaml(containersImages, sidecarsByName)
	return containersImages
}

func setControllerImageByName(defaultCr csiv1.IBMBlockCSI) {
	controllerByName = make(map[string]csiv1.IBMBlockCSIControllerSpec)
	controllerByName[controllerContainerName] = defaultCr.Spec.Controller
}

func setNodeImageByName(defaultCr csiv1.IBMBlockCSI) {
	nodeByName = make(map[string]csiv1.IBMBlockCSINodeSpec)
	nodeByName[nodeContainerName] = defaultCr.Spec.Node
}

func setHostDefinerDeploymentImageByName(hostDefinerCr csiv1.HostDefiner) {
	hostDefinerByName = make(map[string]csiv1.IBMBlockHostDefinerSpec)
	hostDefinerByName[hostDefinerContainerName] = hostDefinerCr.Spec.HostDefiner
}

func addImagesByNameFromYaml(containersImages map[string]string, sidecarsByName map[string]csiv1.CSISidecar) map[string]string {
	containersImages = addSideCarsImagesToContainersImagesMap(containersImages, sidecarsByName)
	containersImages = addNodeImageToContainersImagesMap(containersImages, nodeByName)
	containersImages = addControllerImageToContainersImagesMap(containersImages, controllerByName)
	return containersImages
}

func addSideCarsImagesToContainersImagesMap(containersImages map[string]string,
	sidecarsImagesByName map[string]csiv1.CSISidecar) map[string]string {
	for containerName, sidecar := range sidecarsImagesByName {
		containersImages[containerName] = getImageFromRepositoryAndTag(sidecar.Repository, sidecar.Tag)
	}
	return containersImages
}

func addNodeImageToContainersImagesMap(containersImages map[string]string,
	nodeImagesByName map[string]csiv1.IBMBlockCSINodeSpec) map[string]string {
	node := nodeImagesByName[nodeContainerName]
	containersImages[nodeContainerName] = getImageFromRepositoryAndTag(node.Repository, node.Tag)
	return containersImages
}

func addControllerImageToContainersImagesMap(containersImages map[string]string,
	controllerImagesByName map[string]csiv1.IBMBlockCSIControllerSpec) map[string]string {
	controller := controllerImagesByName[controllerContainerName]
	containersImages[controllerContainerName] = getImageFromRepositoryAndTag(controller.Repository, controller.Tag)
	return containersImages
}

func addHostDefinerDeploymentImageToContainersImagesMap(containersImages map[string]string,
	hostDefinerImagesByName map[string]csiv1.IBMBlockHostDefinerSpec) map[string]string {
	hostDefiner := hostDefinerImagesByName[hostDefinerContainerName]
	containersImages[hostDefinerContainerName] = getImageFromRepositoryAndTag(hostDefiner.Repository, hostDefiner.Tag)
	return containersImages
}

func getImageFromRepositoryAndTag(containerRepository string, containerTag string) string {
	image := containerRepository + ":" + containerTag
	return image
}

func GetResourceKey(resourceName config.ResourceName, CSIObjectName string, CSIObjectNamespace string) types.NamespacedName {
	resourceKey := types.NamespacedName{
		Name:      getResourceNameInCluster(resourceName, CSIObjectName),
		Namespace: CSIObjectNamespace,
	}
	return resourceKey
}

func getResourceNameInCluster(resourceName config.ResourceName, CSIObjectName string) string {
	name := config.GetNameForResource(resourceName, CSIObjectName)
	if CSIObjectName == "" {
		name = resourceName.String()
	}
	return name
}
