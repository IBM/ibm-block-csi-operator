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
	v1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"sigs.k8s.io/yaml"
)

const (
	EnvNameCrYaml string = "CR_YAML"
	NodeAgentTag  = "1.0.0"

	DefaultLogLevel  = "DEBUG"
	ControllerUserID = int64(9999)

	NodeAgentPort = "10086"
)

var DefaultCr v1.IBMBlockCSI

var DefaultSidecarsByName map[string]v1.CSISidecar

var OfficialRegistriesUsernames = sets.NewString("ibmcom", "k8s.gcr.io/sig-storage",
                                                 "quay.io/k8scsi", "registry.redhat.io/openshift4")

func LoadDefaultsOfIBMBlockCSI() error {
	crYamlPath := os.Getenv(EnvNameCrYaml)

	if crYamlPath == "" {
		return fmt.Errorf("environment variable %q was not set", EnvNameCrYaml)
	}

	yamlFile, err := ioutil.ReadFile(crYamlPath)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %v", yamlFile, err)
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
