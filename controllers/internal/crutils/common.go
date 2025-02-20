/**
 * Copyright 2025 IBM Corp.
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

package crutils

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Instance interface {
	GetLabels() labels.Set
	GetObjectKind() schema.ObjectKind
}

func getImagePullSecrets(imagePullSecrets []string) []corev1.LocalObjectReference {
	secrets := []corev1.LocalObjectReference{}
	if len(imagePullSecrets) > 0 {
		for _, secretName := range imagePullSecrets {
			secrets = append(secrets, corev1.LocalObjectReference{Name: secretName})
		}
	}
	return secrets
}
