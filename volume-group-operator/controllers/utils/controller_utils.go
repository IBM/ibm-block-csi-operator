package utils

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ControllerUtils contains helper methods
type ControllerUtils struct {
	client.Client
}

func (r *ControllerUtils) updateObject(updateObject client.Object) error {
	if err := r.Client.Update(context.TODO(), updateObject); err != nil {
		return fmt.Errorf("failed to update %s (%s/%s) %w", updateObject.GetObjectKind(), updateObject.GetNamespace(), updateObject.GetName(), err)
	}
	return nil
}
