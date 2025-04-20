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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CSISidecar struct {
	// The name of the csi sidecar image
	Name string `json:"name"`

	// The repository of the csi sidecar image
	Repository string `json:"repository"`

	// The tag of the csi sidecar image
	Tag string `json:"tag"`

	// The pullPolicy of the csi sidecar image
	// +kubebuilder:validation:Optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`
}

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IBMBlockCSISpec defines the desired state of IBMBlockCSI
type IBMBlockCSISpec struct {
	Controller IBMBlockCSIControllerSpec `json:"controller"`
	Node       IBMBlockCSINodeSpec       `json:"node"`

	// +kubebuilder:validation:Optional
	Sidecars []CSISidecar `json:"sidecars,omitempty"`

	// +kubebuilder:validation:Optional
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	HealthPort uint16 `json:"healthPort,omitempty"`

	// +kubebuilder:validation:Optional
	EnableCallHome string `json:"enableCallHome,omitempty"`

	// +kubebuilder:validation:Optional
	ODFVersionForCallHome string `json:"odfVersionForCallHome,omitempty"`
}

// seems not work in this way, need to figure out why
//// IBMBlockCSIComponentSpec defines the desired state of IBMBlockCSIController
//type BlockCSIComponent struct {
//	Repository string `json:"repository"`
//	Tag        string `json:"tag"`

//	// +kubebuilder:validation:Optional
//	PullPolicy string `json:"pullPolicy"`

//	// +kubebuilder:validation:Optional
//	Affinity *corev1.Affinity `json:"affinity"`

//	// +listType=set
//	// +kubebuilder:validation:Optional
//	Tolerations []corev1.Toleration `json:"tolerations"`
//}

// IBMBlockCSIControllerSpec defines the desired state of IBMBlockCSIController
type IBMBlockCSIControllerSpec struct {
	// BlockCSIComponent `json:"blockCSIComponent"`

	Repository string `json:"repository"`
	Tag        string `json:"tag"`

	// +kubebuilder:validation:Optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`

	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// IBMBlockCSINodeSpec defines the desired state of IBMBlockCSINode
type IBMBlockCSINodeSpec struct {
	// BlockCSIComponent `json:"blockCSIComponent"`

	Repository string `json:"repository"`
	Tag        string `json:"tag"`

	// +kubebuilder:validation:Optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`

	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	MemoryRequirements string `json:"memoryRequirements"`

	WorkersLimit uint16 `json:"workersLimit,omitempty"`
}

// IBMBlockCSIStatus defines the observed state of IBMBlockCSI
type IBMBlockCSIStatus struct {
	// Phase is the driver running phase
	Phase           DriverPhase `json:"phase"`
	ControllerReady bool        `json:"controllerReady"`
	NodeReady       bool        `json:"nodeReady"`

	// Version is the current driver version
	Version string `json:"version"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IBMBlockCSI is the Schema for the ibmblockcsis API
// +kubebuilder:resource:path=ibmblockcsis,scope=Namespaced,shortName=ibc
type IBMBlockCSI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IBMBlockCSISpec   `json:"spec,omitempty"`
	Status IBMBlockCSIStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IBMBlockCSIList contains a list of IBMBlockCSI
type IBMBlockCSIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IBMBlockCSI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IBMBlockCSI{}, &IBMBlockCSIList{})
}
