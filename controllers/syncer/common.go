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
