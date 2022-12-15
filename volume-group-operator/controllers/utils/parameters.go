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

package utils

import (
	"errors"
	"fmt"
	"strings"
)

// FilterPrefixedParameters removes all the reserved keys from the
// volumegroupclass which are matching the prefix.
func FilterPrefixedParameters(prefix string, param map[string]string) map[string]string {
	newParam := map[string]string{}
	for k, v := range param {
		if !strings.HasPrefix(k, prefix) {
			newParam[k] = v
		}
	}

	return newParam
}

// ValidatePrefixedParameters checks for unknown reserved keys in parameters and
// empty values for reserved keys.
func ValidatePrefixedParameters(param map[string]string) error {
	for k, v := range param {
		if strings.HasPrefix(k, VolumeGroupAsPrefix) {
			switch k {
			case PrefixedVolumeGroupSecretNameKey:
				if v == "" {
					return errors.New("secret name cannot be empty")
				}
			case PrefixedVolumeGroupSecretNamespaceKey:
				if v == "" {
					return errors.New("secret namespace cannot be empty")
				}
			// keep adding known prefixes to this list.
			default:

				return fmt.Errorf("found unknown parameter key %q with reserved prefix %s", k, VolumeGroupAsPrefix)
			}
		}
	}

	return nil
}
