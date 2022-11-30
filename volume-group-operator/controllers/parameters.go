/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// VolumeGroup Parameters prefixed with volumeGroupParameterPrefix are not passed through
	// to the driver on RPC calls. Instead, these are the parameters used by the
	// operator to get the required object from kubernetes and pass it to the
	// Driver.
	volumeGroupParameterPrefix = "volumegroup.storage.ibm.io/"

	prefixedVolumeGroupSecretNameKey      = volumeGroupParameterPrefix + "secret-name"      // name key for secret
	prefixedVolumeGroupSecretNamespaceKey = volumeGroupParameterPrefix + "secret-namespace" // namespace key secret
)

// filterPrefixedParameters removes all the reserved keys from the
// volumegroupclass which are matching the prefix.
func filterPrefixedParameters(prefix string, param map[string]string) map[string]string {
	newParam := map[string]string{}
	for k, v := range param {
		if !strings.HasPrefix(k, prefix) {
			newParam[k] = v
		}
	}

	return newParam
}

// validatePrefixParameters checks for unknown reserved keys in parameters and
// empty values for reserved keys.
func validatePrefixedParameters(param map[string]string) error {
	for k, v := range param {
		if strings.HasPrefix(k, volumeGroupParameterPrefix) {
			switch k {
			case prefixedVolumeGroupSecretNameKey:
				if v == "" {
					return errors.New("secret name cannot be empty")
				}
			case prefixedVolumeGroupSecretNamespaceKey:
				if v == "" {
					return errors.New("secret namespace cannot be empty")
				}
			// keep adding known prefixes to this list.
			default:

				return fmt.Errorf("found unknown parameter key %q with reserved prefix %s", k, volumeGroupParameterPrefix)
			}
		}
	}

	return nil
}