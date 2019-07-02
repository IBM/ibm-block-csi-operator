package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IBMBlockCSISpec defines the desired state of IBMBlockCSI
// +k8s:openapi-gen=true
type IBMBlockCSISpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
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
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Ready bool `json:"ready"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IBMBlockCSI is the Schema for the ibmblockcsis API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
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
