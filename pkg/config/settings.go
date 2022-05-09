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
	"io/ioutil"
	"os"

	v1 "github.com/IBM/ibm-block-csi-operator/api/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

const (
	EnvNameIBMBlockCSICrYaml    = "IBMBlockCSI_CR_YAML"
	EnvNameHostDefinitionCrYaml = "HostDefinition_CR_YAML"
	NodeAgentTag                = "1.0.0"

	DefaultLogLevel  = "DEBUG"
	ControllerUserID = int64(9999)

	NodeAgentPort = "10086"

	IBMRegistryUsername        = "ibmcom"
	K8SRegistryUsername        = "k8s.gcr.io/sig-storage"
	QuayRegistryUsername       = "quay.io/k8scsi"
	QuayAddonsRegistryUsername = "quay.io/csiaddons"
	RedHatRegistryUsername     = "registry.redhat.io/openshift4"
)

var DefaultCr v1.IBMBlockCSI

var DefaultHostDefinitionCr v1.HostDefinition

var DefaultSidecarsByName map[string]v1.CSISidecar

var OfficialRegistriesUsernames = sets.NewString(IBMRegistryUsername, K8SRegistryUsername,
	QuayRegistryUsername, QuayAddonsRegistryUsername,
	RedHatRegistryUsername)

func LoadDefaultsOfIBMBlockCSI() error {
	yamlFile, err := getCrYamlFile(EnvNameIBMBlockCSICrYaml)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, &DefaultCr)
	if err != nil {
		return fmt.Errorf("error unmarshaling yaml: %v", err)
	}

	DefaultSidecarsByName = make(map[string]v1.CSISidecar)

	for _, sidecar := range DefaultCr.Spec.Sidecars {
		DefaultSidecarsByName[sidecar.Name] = sidecar
	}

	return nil
}

func LoadDefaultsOfHostDefinition() error {
	yamlFile, err := getCrYamlFile(EnvNameHostDefinitionCrYaml)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, &DefaultHostDefinitionCr)
	if err != nil {
		return fmt.Errorf("error unmarshaling yaml: %v", err)
	}

	return nil
}

func getCrYamlFile(crPathEnvVariable string) ([]byte, error) {
	crYamlPath, err := getCrYamlPath(crPathEnvVariable)
	if err != nil {
		return []byte{}, err
	}

	yamlFile, err := ioutil.ReadFile(crYamlPath)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to read file %q: %v", yamlFile, err)
	}
	return yamlFile, nil
}

func getCrYamlPath(crPathEnvVariable string) (string, error) {
	crYamlPath := os.Getenv(crPathEnvVariable)

	if crYamlPath == "" {
		return "", fmt.Errorf("environment variable %q was not set", crPathEnvVariable)
	}
	return crYamlPath, nil
}
