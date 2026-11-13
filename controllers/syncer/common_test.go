/**
 * Copyright 2026 IBM Corp.
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

package syncer

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestGetSvcSshKeepaliveEnvVars(t *testing.T) {
	tests := []struct {
		name          string
		configMapData map[string]string
		expected      []corev1.EnvVar
	}{
		{
			name:          "no config map",
			configMapData: nil,
			expected:      []corev1.EnvVar{},
		},
		{
			name:          "config map without keepalive keys",
			configMapData: map[string]string{"svcSshPort": "22", "workersLimit": "10"},
			expected:      []corev1.EnvVar{},
		},
		{
			name: "all keys set",
			configMapData: map[string]string{
				"svcSshKeepaliveInterval": "20",
				"svcTcpKeepidle":          "25",
				"svcTcpKeepintvl":         "6",
				"svcTcpKeepcnt":           "4",
				"svcTcpUserTimeout":       "45000",
			},
			expected: []corev1.EnvVar{
				{Name: "SVC_SSH_KEEPALIVE_INTERVAL", Value: "20"},
				{Name: "SVC_TCP_KEEPIDLE", Value: "25"},
				{Name: "SVC_TCP_KEEPINTVL", Value: "6"},
				{Name: "SVC_TCP_KEEPCNT", Value: "4"},
				{Name: "SVC_TCP_USER_TIMEOUT", Value: "45000"},
			},
		},
		{
			name:          "only one key set",
			configMapData: map[string]string{"svcTcpUserTimeout": "60000"},
			expected:      []corev1.EnvVar{{Name: "SVC_TCP_USER_TIMEOUT", Value: "60000"}},
		},
		{
			name: "invalid and empty values are left out",
			configMapData: map[string]string{
				"svcSshKeepaliveInterval": "abc",
				"svcTcpKeepidle":          "0",
				"svcTcpKeepintvl":         "-5",
				"svcTcpKeepcnt":           "2.5",
				"svcTcpUserTimeout":       "",
			},
			expected: []corev1.EnvVar{},
		},
		{
			name: "valid values are kept next to invalid ones",
			configMapData: map[string]string{
				"svcTcpKeepidle": "abc",
				"svcTcpKeepcnt":  "5",
			},
			expected: []corev1.EnvVar{{Name: "SVC_TCP_KEEPCNT", Value: "5"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := getSvcSshKeepaliveEnvVars(test.configMapData)
			if !reflect.DeepEqual(test.expected, actual) {
				t.Errorf("expected %v, got %v", test.expected, actual)
			}
		})
	}
}
