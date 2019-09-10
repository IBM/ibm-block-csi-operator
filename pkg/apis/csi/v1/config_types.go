package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeAgentPhase string

const (
	NodeAgentPhaseDisabled NodeAgentPhase = "Disabled"
	NodeAgentPhaseCreating NodeAgentPhase = "Creating"
	NodeAgentPhaseRunning  NodeAgentPhase = "Running"
	NodeAgentPhaseFailed   NodeAgentPhase = "Failed"
)

// ConfigSpec defines the desired state of Config
// +k8s:openapi-gen=true
type ConfigSpec struct {
	DefineHost bool          `json:"defineHost"`
	NodeAgent  NodeAgentSpec `json:"nodeAgent"`
}

// NodeAgentSpec defines the desired state of NodeAgent
// +k8s:openapi-gen=true
type NodeAgentSpec struct {
	// The repository of the node agent image
	Repository string `json:"repository"`

	// The tag of the node agent image
	Tag string `json:"tag"`

	// The port of the node agent server
	Port string `json:"port"`
}

// NodeAgentStatus defines the current state of NodeAgent
// +k8s:openapi-gen=true
type NodeAgentStatus struct {
	// Phase is the NodeAgent running phase
	Phase NodeAgentPhase `json:"phase"`
}

// ConfigStatus defines the observed state of Config
// +k8s:openapi-gen=true
type ConfigStatus struct {
	// Phase is the driver running phase
	NodeAgent NodeAgentStatus `json:"nodeAgent"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Config is the Schema for the configs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigSpec   `json:"spec,omitempty"`
	Status ConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigList contains a list of Config
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Config{}, &ConfigList{})
}
