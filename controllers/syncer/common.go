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
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var defaultAnnotations = []string{
	"productID",
	"productName",
	"productVersion",
}

func ensureAnnotations(templateObjectMeta *metav1.ObjectMeta, objectMeta *metav1.ObjectMeta, annotations labels.Set) {
	for _, s := range defaultAnnotations {
		templateObjectMeta.Annotations[s] = annotations[s]
		objectMeta.Annotations[s] = annotations[s]
	}
}

func GetConfigMap(configMapName string) (map[string]string) {
	config, err := rest.InClusterConfig()
	if err != nil {
		//os.Exit(3)
		return nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		//os.Exit(4)
		return nil
	}

	// Get namespace (default if not specified)
	namespace := "default"
	// Try to get namespace from service account
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		namespace = string(data)
	}

	// Retrieve ConfigMap
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	return configMap.Data
}

// GetConfigMapValue extracts a specific value from a ConfigMap data map.
func GetConfigMapValue(configMapData map[string]string, key string, defaultvalue string) (string, bool) {
	if configMapData == nil {
		return defaultvalue, false
	}

	value, exists := configMapData[key]
	if !exists {
		return defaultvalue, false
	}

	return value, true
}

// ensureResourcesFromConfigMap reads the memoryRequirements key from the given ConfigMap
// (falling back to defaultRequirements if absent) and returns a populated ResourceRequirements.
// defaultRequirements and the ConfigMap value must be a comma-separated string of exactly
// four values: cpuRequests,cpuLimits,memoryRequests,memoryLimits.
func ensureResourcesFromConfigMap(configMapName, defaultRequirements string) corev1.ResourceRequirements {
	configMapData := GetConfigMap(configMapName)
	value, _ := GetConfigMapValue(configMapData, "memoryRequirements", defaultRequirements)
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		log.Fatalf("memoryRequirements in %s is not valid (expected 4 comma-separated values: cpuReq,cpuLim,memReq,memLim)", configMapName)
	}
	return ensureResources(parts[0], parts[1], parts[2], parts[3])
}

// svcSshKeepaliveSettings maps a ConfigMap key to the controller env var that carries the optional
// SSH keepalive setting to the driver, with the maximal allowed value. The driver does not validate
// the values. The unit is part of each name (Seconds, Ms), except for the probe count.
// Driver defaults:
//
//	svcSshKeepaliveIntervalSeconds - default 15
//	svcTcpKeepidleSeconds          - default 15
//	svcTcpKeepintvlSeconds         - default 5
//	svcTcpKeepcnt                  - number of probes, default 3
//	svcTcpUserTimeoutMs            - default 30000
var svcSshKeepaliveSettings = []struct {
	configMapKey string
	envVarName   string
	maxValue     uint64
}{
	{"svcSshKeepaliveIntervalSeconds", "SVC_SSH_KEEPALIVE_INTERVAL_SECONDS", 3600},
	{"svcTcpKeepidleSeconds", "SVC_TCP_KEEPIDLE_SECONDS", 32767},
	{"svcTcpKeepintvlSeconds", "SVC_TCP_KEEPINTVL_SECONDS", 32767},
	{"svcTcpKeepcnt", "SVC_TCP_KEEPCNT", 127},
	{"svcTcpUserTimeoutMs", "SVC_TCP_USER_TIMEOUT_MS", 600000},
}

// getSvcSshKeepaliveEnvVars returns the optional SSH keepalive env vars that are set to a valid
// value in the given ConfigMap data: a positive integer, not above the maximum of the key. A key
// that is missing, empty or not valid is left out with a warning, so the driver keeps its default for it.
func getSvcSshKeepaliveEnvVars(nodeConfigMapData map[string]string) []corev1.EnvVar {
	envVars := []corev1.EnvVar{}
	for _, setting := range svcSshKeepaliveSettings {
		value, exists := GetConfigMapValue(nodeConfigMapData, setting.configMapKey, "")
		if !exists || value == "" {
			continue
		}
		number, err := strconv.ParseUint(value, 10, 32)
		if err != nil || number == 0 || number > setting.maxValue {
			log.Printf("Ignoring invalid value %q of %s, a positive integer up to %d is required. The default is used.",
				value, setting.configMapKey, setting.maxValue)
			continue
		}
		envVars = append(envVars, corev1.EnvVar{Name: setting.envVarName, Value: value})
	}
	return envVars
}
