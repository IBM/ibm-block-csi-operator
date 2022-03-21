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
	clustersyncer "github.com/IBM/ibm-block-csi-operator/controllers/syncer"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"
)

var (
	relativeCrPath = "../config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml"
	nodeContainerName = clustersyncer.NodeContainerName
	controllerContainerName = clustersyncer.ControllerContainerName
)

type crYamlConfig struct {
	ApiVersion string
	Kind string
	Metadata struct {
		Namespace string
	}
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

func GetNamespaceFromCrFile() string {
	var c crYamlConfig

	return c.getCrYaml().Metadata.Namespace
}

func GetApiVersionFromCrFile() string {
	var c crYamlConfig

	return c.getCrYaml().ApiVersion
}

func GetKindFromCrFile() string {
	var c crYamlConfig

	return c.getCrYaml().Kind
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

func GetIBMBlockCsiSpec(containersImages map[string]string) csiv1.IBMBlockCSISpec {
	var spec csiv1.IBMBlockCSISpec
	spec.Controller = addControllerToIBMSpec(spec, containersImages)
	spec.Node = addNodeToIBMSpec(spec, containersImages)
	spec.Sidecars = addSidecarsToIBMSpec(spec, containersImages)
	return spec
}

func addControllerToIBMSpec(spec csiv1.IBMBlockCSISpec, containersImages map[string]string) csiv1.IBMBlockCSIControllerSpec {
	var controllerSpec csiv1.IBMBlockCSIControllerSpec
	controllerSpec.Repository = strings.Split(containersImages[controllerContainerName], ":")[0]
	controllerSpec.Tag = strings.Split(containersImages[controllerContainerName], ":")[1]
	return controllerSpec
}

func addNodeToIBMSpec(spec csiv1.IBMBlockCSISpec, containersImages map[string]string) csiv1.IBMBlockCSINodeSpec {
	var nodeSpec csiv1.IBMBlockCSINodeSpec
	nodeSpec.Repository = strings.Split(containersImages[nodeContainerName], ":")[0]
	nodeSpec.Tag = strings.Split(containersImages[nodeContainerName], ":")[1]
	return nodeSpec
}

func addSidecarsToIBMSpec(spec csiv1.IBMBlockCSISpec, containersImages map[string]string) []csiv1.CSISidecar {
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
