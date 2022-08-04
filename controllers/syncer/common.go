/**
 * Copyright 2022 IBM Corp.
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
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	csiversion "github.com/IBM/ibm-block-csi-operator/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var defaultAnnotations = labels.Set{
	"productID":      config.ProductName,
	"productName":    config.ProductName,
	"productVersion": csiversion.Version,
}

func ensureAnnotations(templateObjectMeta *metav1.ObjectMeta, objectMeta *metav1.ObjectMeta,
	annotations labels.Set) {
	for k := range defaultAnnotations {
		templateObjectMeta.Annotations[k] = annotations[k]
		objectMeta.Annotations[k] = annotations[k]
	}
}
