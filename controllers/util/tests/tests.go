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

package tests

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	csiv1 "github.com/IBM/ibm-block-csi-operator/api/v1"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"
)

var (
	relativeCrPath = "../config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml"
	nodeContainerName = "ibm-block-csi-node"
	controllerContainerName = "ibm-block-csi-controller"
)

type crYamlConfig struct {
	Spec struct {
		Sidecars []imageProperties
		Controller imageProperties
		Node imageProperties
	}
}
type imageProperties struct {
	Name string `yaml:"name"`

	Repository string `yaml:"repository"`

	Tag string `yaml:"tag"`
}
 
func GetImagesByName() map[string]string {
	var c crYamlConfig
	containersImages := make(map[string]string)

	containersImages = addImagesByNameFromYaml(containersImages, c.getCrYaml())
	return containersImages
}

func (c *crYamlConfig) getCrYaml() *crYamlConfig {
	yamlFile, err := ioutil.ReadFile(relativeCrPath)
	if err != nil {
		fmt.Println("unable to read yaml file, error:", err)
	 	os.Exit(1)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		fmt.Println("unable to decodes yaml file, error:", err)
	 	os.Exit(1)
	}
	 return c
}
 
func addImagesByNameFromYaml(containersImages map[string]string, crYaml *crYamlConfig) map[string]string {
	containersImages = addSideCarsImagesToContainersImagesMap(containersImages, crYaml)
	containersImages = addNodeImageToContainersImagesMap(containersImages, crYaml)
	containersImages = addControllerImageToContainersImagesMap(containersImages, crYaml)
	return containersImages
}
 
func addSideCarsImagesToContainersImagesMap(containersImages map[string]string, crYaml *crYamlConfig) map[string]string {
	for _, sidecar := range crYaml.Spec.Sidecars {
		containersImages[sidecar.Name] = getImageWithoutForwardSlash(sidecar)
	}
	return containersImages
}

func addNodeImageToContainersImagesMap(containersImages map[string]string, crYaml *crYamlConfig) map[string]string {
	node := crYaml.Spec.Node
	containersImages[nodeContainerName] = getImageWithoutForwardSlash(node)
	return containersImages
}

func addControllerImageToContainersImagesMap(containersImages map[string]string, crYaml *crYamlConfig) map[string]string {
	controller := crYaml.Spec.Controller
	containersImages[controllerContainerName] = getImageWithoutForwardSlash(controller)
	return containersImages
}

func getImageWithoutForwardSlash(container imageProperties) string {
	image := container.Repository + ":" + container.Tag
	return strings.Replace(image, "/", "-", -1)
}

func GetIbmBlockCsiSpec(containersImages map[string]string) csiv1.IBMBlockCSISpec {
	var spec csiv1.IBMBlockCSISpec
	spec.Controller = addControllerToIbmSpec(spec, containersImages)
	spec.Node = addNodeToIbmSpec(spec, containersImages)
	spec.Sidecars = addSidecarsToIbmSpec(spec, containersImages)
	return spec
}

func addControllerToIbmSpec(spec csiv1.IBMBlockCSISpec, containersImages map[string]string) csiv1.IBMBlockCSIControllerSpec {
	var controllerSpec csiv1.IBMBlockCSIControllerSpec
	controllerSpec.Repository = strings.Split(containersImages[controllerContainerName], ":")[0]
	controllerSpec.Tag = strings.Split(containersImages[controllerContainerName], ":")[1]
	return controllerSpec
}

func addNodeToIbmSpec(spec csiv1.IBMBlockCSISpec, containersImages map[string]string) csiv1.IBMBlockCSINodeSpec {
	var nodeSpec csiv1.IBMBlockCSINodeSpec
	nodeSpec.Repository = strings.Split(containersImages[nodeContainerName], ":")[0]
	nodeSpec.Tag = strings.Split(containersImages[nodeContainerName], ":")[1]
	return nodeSpec
}

func addSidecarsToIbmSpec(spec csiv1.IBMBlockCSISpec, containersImages map[string]string) []csiv1.CSISidecar {
	var sidecars []csiv1.CSISidecar
	for containerName, imageName := range containersImages {
		if ! isControllerOrNode(containerName) {
			sidecars = append(sidecars, getSidecar(containerName, imageName))
		}
	}
	return sidecars
}

func isControllerOrNode(containerName string) bool {
	controllerAndNode := [2]string{controllerContainerName, nodeContainerName}
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

func GetResourceKey(resourceName config.ResourceName, csiObjectName string, csiObjectNamespace string) types.NamespacedName{
	resourceKey := types.NamespacedName{
	  Name:      getResourceNameInCluster(resourceName, csiObjectName),
	  Namespace: csiObjectNamespace,
	}
	return resourceKey
}

func getResourceNameInCluster(resourceName config.ResourceName, csiObjectName string) string{
	name := config.GetNameForResource(resourceName, csiObjectName)
	if csiObjectName == "" {
		name = resourceName.String()
	}
	return name
}