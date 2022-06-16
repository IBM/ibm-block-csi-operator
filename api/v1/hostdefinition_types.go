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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HostDefinitionSpec defines the desired state of HostDefinition
type HostDefinitionSpec struct {
	HostDefinition IBMBlockCSIHostDefinitionSpec `json:"hostDefinition"`
}

// IBMBlockCSIHostDefinitionSpec defines the observed state of HostDefinition
type IBMBlockCSIHostDefinitionSpec struct {
	ManagementAddress string `json:"managementAddress"`
	HostNameInStorage string `json:"hostNameInStorage"`

	// +kubebuilder:validation:Optional
	SecretName string `json:"secretName"`
	// +kubebuilder:validation:Optional
	SecretNamespace string `json:"secretNamespace"`
	// +kubebuilder:validation:Optional
	ConnectvityType string `json:"connectvityType"`
	// +kubebuilder:validation:Optional
	ConnectivityPorts ConnectivityPorts `json:"connectivityPorts"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	RetryVerifying bool `json:"retryVerifying"`
}

// ConnectivityPorts defines the ports of the hostDefinition
type ConnectivityPorts struct {
	// +kubebuilder:validation:Optional
	Iscsi []string `json:"iscsi"`
	// +kubebuilder:validation:Optional
	Fc []string `json:"fc"`
	// +kubebuilder:validation:Optional
	Nvme []string `json:"nvme"`
}

// HostDefinitionStatus defines the observed state of HostDefinition
type HostDefinitionStatus struct {
	Phase DriverPhase `json:"phase"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// HostDefinition is the Schema for the hostdefinitions API
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Storage",type=string,JSONPath=`.spec.hostDefinition.storageServer`
// +kubebuilder:printcolumn:name="Host",type=string,JSONPath=`.spec.hostDefinition.hostNameInStorage`
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
