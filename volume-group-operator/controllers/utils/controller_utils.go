package utils

import "sigs.k8s.io/controller-runtime/pkg/client"

// ControllerUtils contains helper methods
type ControllerUtils struct {
	client.Client
}
