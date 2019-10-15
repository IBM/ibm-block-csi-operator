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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DriverPhase string

const (
	DriverPhaseNone     DriverPhase = ""
	DriverPhaseCreating DriverPhase = "Creating"
	DriverPhaseRunning  DriverPhase = "Running"
	DriverPhaseFailed   DriverPhase = "Failed"
)

// IBMBlockCSISpec defines the desired state of IBMBlockCSI
// +k8s:openapi-gen=true
type IBMBlockCSISpec struct {
	Controller IBMBlockCSIControllerSpec `json:"controller"`
	Node       IBMBlockCSINodeSpec       `json:"node"`
}

// IBMBlockCSIControllerSpec defines the desired state of IBMBlockCSIController
// +k8s:openapi-gen=true
type IBMBlockCSIControllerSpec struct {
	// The repository of the controller image
	Repository string `json:"repository"`

	// The tag of the controller image
	Tag string `json:"tag"`
}

// IBMBlockCSINodeSpec defines the desired state of IBMBlockCSINode
// +k8s:openapi-gen=true
type IBMBlockCSINodeSpec struct {
	// The repository of the node image
	Repository string `json:"repository"`

	// The tag of the node image
	Tag string `json:"tag"`
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
// +kubebuilder:resource:path=ibmblockcsis,scope=Namespaced
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
