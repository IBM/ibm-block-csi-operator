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
				"svcSshKeepaliveIntervalSeconds": "20",
				"svcTcpKeepidleSeconds":          "25",
				"svcTcpKeepintvlSeconds":         "6",
				"svcTcpKeepcnt":                  "4",
				"svcTcpUserTimeoutMs":            "45000",
			},
			expected: []corev1.EnvVar{
				{Name: "SVC_SSH_KEEPALIVE_INTERVAL_SECONDS", Value: "20"},
				{Name: "SVC_TCP_KEEPIDLE_SECONDS", Value: "25"},
				{Name: "SVC_TCP_KEEPINTVL_SECONDS", Value: "6"},
				{Name: "SVC_TCP_KEEPCNT", Value: "4"},
				{Name: "SVC_TCP_USER_TIMEOUT_MS", Value: "45000"},
			},
		},
		{
			name:          "only one key set",
			configMapData: map[string]string{"svcTcpUserTimeoutMs": "60000"},
			expected:      []corev1.EnvVar{{Name: "SVC_TCP_USER_TIMEOUT_MS", Value: "60000"}},
		},
		{
			name: "invalid and empty values are left out",
			configMapData: map[string]string{
				"svcSshKeepaliveIntervalSeconds": "abc",
				"svcTcpKeepidleSeconds":          "0",
				"svcTcpKeepintvlSeconds":         "-5",
				"svcTcpKeepcnt":                  "2.5",
				"svcTcpUserTimeoutMs":            "",
			},
			expected: []corev1.EnvVar{},
		},
		{
			name: "valid values are kept next to invalid ones",
			configMapData: map[string]string{
				"svcTcpKeepidleSeconds": "abc",
				"svcTcpKeepcnt":         "5",
			},
			expected: []corev1.EnvVar{{Name: "SVC_TCP_KEEPCNT", Value: "5"}},
		},
		{
			name: "values that are not a strict positive integer are left out",
			configMapData: map[string]string{
				"svcSshKeepaliveIntervalSeconds": "+5",
				"svcTcpKeepidleSeconds":          " 5 ",
				"svcTcpKeepintvlSeconds":         "1_0",
				"svcTcpKeepcnt":                  "0x10",
				"svcTcpUserTimeoutMs":            "4294967296",
			},
			expected: []corev1.EnvVar{},
		},
		{
			name: "values above the maximum are left out",
			configMapData: map[string]string{
				"svcSshKeepaliveIntervalSeconds": "3601",
				"svcTcpKeepidleSeconds":          "32768",
				"svcTcpKeepintvlSeconds":         "32768",
				"svcTcpKeepcnt":                  "128",
				"svcTcpUserTimeoutMs":            "600001",
			},
			expected: []corev1.EnvVar{},
		},
		{
			name: "values at the maximum are kept",
			configMapData: map[string]string{
				"svcSshKeepaliveIntervalSeconds": "3600",
				"svcTcpKeepidleSeconds":          "32767",
				"svcTcpKeepintvlSeconds":         "32767",
				"svcTcpKeepcnt":                  "127",
				"svcTcpUserTimeoutMs":            "600000",
			},
			expected: []corev1.EnvVar{
				{Name: "SVC_SSH_KEEPALIVE_INTERVAL_SECONDS", Value: "3600"},
				{Name: "SVC_TCP_KEEPIDLE_SECONDS", Value: "32767"},
				{Name: "SVC_TCP_KEEPINTVL_SECONDS", Value: "32767"},
				{Name: "SVC_TCP_KEEPCNT", Value: "127"},
				{Name: "SVC_TCP_USER_TIMEOUT_MS", Value: "600000"},
			},
		},
		{
			name: "value of 1 is kept",
			configMapData: map[string]string{
				"svcTcpKeepcnt": "1",
			},
			expected: []corev1.EnvVar{{Name: "SVC_TCP_KEEPCNT", Value: "1"}},
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
