/**
 * Copyright 2022 IBM Corp.
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
	"strings"

	csiv1 "github.com/IBM/ibm-block-csi-operator/api/v1"
	clustersyncer "github.com/IBM/ibm-block-csi-operator/controllers/syncer"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"k8s.io/apimachinery/pkg/types"
)

var (
	nodeContainerName = clustersyncer.NodeContainerName
	controllerContainerName = clustersyncer.ControllerContainerName
	defaultControllerByName map[string]csiv1.IBMBlockCSIControllerSpec
	defaultNodeByName map[string]csiv1.IBMBlockCSINodeSpec
)
var err = config.LoadDefaultsOfIBMBlockCSI()

func GetImagesByName() map[string]string {
	containersImages := make(map[string]string)
	setDefaultControllerImageByName()
	setDefaultNodeImageByName()

	containersImages = addImagesByNameFromYaml(containersImages)
	return containersImages
}

func setDefaultControllerImageByName() {
	defaultControllerByName = make(map[string]csiv1.IBMBlockCSIControllerSpec)
	defaultControllerByName[controllerContainerName] = config.DefaultCr.Spec.Controller
}

func setDefaultNodeImageByName() {
	defaultNodeByName = make(map[string]csiv1.IBMBlockCSINodeSpec)
	defaultNodeByName[nodeContainerName] = config.DefaultCr.Spec.Node
}

func addImagesByNameFromYaml(containersImages map[string]string) map[string]string {
	containersImages = addSideCarsImagesToContainersImagesMap(containersImages, config.DefaultSidecarsByName)
	containersImages = addNodeImageToContainersImagesMap(containersImages, defaultNodeByName)
	containersImages = addControllerImageToContainersImagesMap(containersImages, defaultControllerByName)
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

func getImageFromRepositoryAndTag(containerRepository string, containerTag string) string {
	image := containerRepository + ":" + containerTag
	return image
}

func GetNamespaceFromCrFile() string {
	return config.DefaultCr.ObjectMeta.Namespace
}

func GetIBMBlockCSISpec(containersImages map[string]string) csiv1.IBMBlockCSISpec {
	var spec csiv1.IBMBlockCSISpec
	spec.Controller = addControllerToIBMSpec(containersImages)
	spec.Node = addNodeToIBMSpec(containersImages)
	spec.Sidecars = addSidecarsToIBMSpec(containersImages)
	return spec
}

func addControllerToIBMSpec(containersImages map[string]string) csiv1.IBMBlockCSIControllerSpec {
	var controllerSpec csiv1.IBMBlockCSIControllerSpec
	controllerSpec.Repository = strings.Split(containersImages[controllerContainerName], ":")[0]
	controllerSpec.Tag = strings.Split(containersImages[controllerContainerName], ":")[1]
	return controllerSpec
}

func addNodeToIBMSpec(containersImages map[string]string) csiv1.IBMBlockCSINodeSpec {
	var nodeSpec csiv1.IBMBlockCSINodeSpec
	nodeSpec.Repository = strings.Split(containersImages[nodeContainerName], ":")[0]
	nodeSpec.Tag = strings.Split(containersImages[nodeContainerName], ":")[1]
	return nodeSpec
}

func addSidecarsToIBMSpec(containersImages map[string]string) []csiv1.CSISidecar {
	var sidecars []csiv1.CSISidecar
	for containerName, imageName := range containersImages {
		if ! isControllerOrNode(containerName) {
			sidecars = append(sidecars, getSidecar(containerName, imageName))
		}
	}
	return sidecars
}

func isControllerOrNode(containerName string) bool {
	controllerAndNode := []string{controllerContainerName, nodeContainerName}
	for _, pluginName := range controllerAndNode {
		if pluginName == containerName {
			return true
		}
	}
	return false
}

func getSidecar(containerName string, imageName string) csiv1.CSISidecar {
	var sidecar csiv1.CSISidecar
	sidecar.Name = containerName
	sidecar.Repository = strings.Split(imageName, ":")[0]
	sidecar.Tag = strings.Split(imageName, ":")[1]
	sidecar.ImagePullPolicy = "IfNotPresent"
	return sidecar
}

func GetResourceKey(resourceName config.ResourceName, CSIObjectName string, CSIObjectNamespace string) types.NamespacedName{
	resourceKey := types.NamespacedName{
	  Name:      getResourceNameInCluster(resourceName, CSIObjectName),
	  Namespace: CSIObjectNamespace,
	}
	return resourceKey
}

func getResourceNameInCluster(resourceName config.ResourceName, CSIObjectName string) string{
	name := config.GetNameForResource(resourceName, CSIObjectName)
	if CSIObjectName == "" {
		name = resourceName.String()
	}
	return name
}
