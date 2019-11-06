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

type DriverPhase string

const (
	DriverPhaseNone     DriverPhase = ""
	DriverPhaseCreating DriverPhase = "Creating"
	DriverPhaseRunning  DriverPhase = "Running"
	DriverPhaseFailed   DriverPhase = "Failed"
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

// IBMBlockCSISpec defines the desired state of IBMBlockCSI
// +k8s:openapi-gen=true
type IBMBlockCSISpec struct {
	Controller IBMBlockCSIControllerSpec `json:"controller"`
	Node       IBMBlockCSINodeSpec       `json:"node"`

	// +listType=set
	// +kubebuilder:validation:Optional
	Sidecars []CSISidecar `json:"sidecars"`

	// +listType=set
	// +kubebuilder:validation:Optional
	ImagePullSecrets []string `json:"imagePullSecrets"`
}

// seems not work in this way, need to figure out why
//// IBMBlockCSIComponentSpec defines the desired state of IBMBlockCSIController
//// +k8s:openapi-gen=true
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
// +k8s:openapi-gen=true
type IBMBlockCSIControllerSpec struct {
	// BlockCSIComponent `json:"blockCSIComponent"`

	Repository string `json:"repository"`
	Tag        string `json:"tag"`

	// +kubebuilder:validation:Optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`

	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity"`

	// +listType=set
	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations"`
}

// IBMBlockCSINodeSpec defines the desired state of IBMBlockCSINode
// +k8s:openapi-gen=true
type IBMBlockCSINodeSpec struct {
	// BlockCSIComponent `json:"blockCSIComponent"`

	Repository string `json:"repository"`
	Tag        string `json:"tag"`

	// +kubebuilder:validation:Optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`

	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity"`

	// +listType=set
	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations"`
}

// IBMBlockCSIStatus defines the observed state of IBMBlockCSI
// +k8s:openapi-gen=true
type IBMBlockCSIStatus struct {
	// Phase is the driver running phase
	Phase           DriverPhase `json:"phase"`
	ControllerReady bool        `json:"controllerReady"`
	NodeReady       bool        `json:"nodeReady"`

	// Version is the current driver version
	Version string `json:"version"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IBMBlockCSI is the Schema for the ibmblockcsis API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=ibmblockcsis,scope=Namespaced,shortName=ibc
type IBMBlockCSI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IBMBlockCSISpec   `json:"spec,omitempty"`
	Status IBMBlockCSIStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IBMBlockCSIList contains a list of IBMBlockCSI
type IBMBlockCSIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IBMBlockCSI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IBMBlockCSI{}, &IBMBlockCSIList{})
}
