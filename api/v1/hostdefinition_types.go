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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HostDefinitionSpec defines the desired state of HostDefinition
type HostDefinitionSpec struct {
	HostDefinition IBMBlockCSIHostDefinitionSpec `json:"hostDefinition"`

	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
}

// IBMBlockCSIHostDefinitionSpec defines the observed state of HostDefinition
type IBMBlockCSIHostDefinitionSpec struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`

	// +kubebuilder:validation:Optional
	Prefix string `json:"prefix"`

	// +kubebuilder:validation:Optional
	Connectivity string `json:"connectivity"`

	// +kubebuilder:validation:Optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`

	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// HostDefinitionStatus defines the observed state of HostDefinition
type HostDefinitionStatus struct {
	Phase               DriverPhase `json:"phase"`
	HostDefinitionReady bool        `json:"hostDefinitionReady"`

	// Version is the current driver version
	Version string `json:"version"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HostDefinition is the Schema for the hostdefinitions API
type HostDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HostDefinitionSpec   `json:"spec,omitempty"`
	Status HostDefinitionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HostDefinitionList contains a list of HostDefinition
type HostDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HostDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HostDefinition{}, &HostDefinitionList{})
}