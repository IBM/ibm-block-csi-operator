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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HostDefinerSpec defines the desired state of HostDefiner
type HostDefinerSpec struct {
	HostDefiner IBMBlockHostDefinerSpec `json:"hostDefiner"`

	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
}

// IBMBlockHostDefinerSpec defines the observed state of HostDefiner
type IBMBlockHostDefinerSpec struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`

	// +kubebuilder:validation:Optional
	Prefix string `json:"prefix"`
	// +kubebuilder:validation:Optional
	ConnectivityType string `json:"connectivityType"`
	// +kubebuilder:validation:Optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`
	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=true
	AllowDelete bool `json:"allowDelete,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	DynamicNodeLabeling bool `json:"dynamicNodeLabeling,omitempty"`
	// +kubebuilder:validation:Optional
	PortSet string `json:"portSet"`
}

// HostDefinerStatus defines the observed state of HostDefiner
type HostDefinerStatus struct {
	Phase            DriverPhase `json:"phase"`
	HostDefinerReady bool        `json:"hostDefinerReady"`

	// Version is the current driver version
	Version string `json:"version"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HostDefiner is the Schema for the hostdefiners API
type HostDefiner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HostDefinerSpec   `json:"spec,omitempty"`
	Status HostDefinerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HostDefinerList contains a list of HostDefiner
type HostDefinerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HostDefiner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HostDefiner{}, &HostDefinerList{})
}
