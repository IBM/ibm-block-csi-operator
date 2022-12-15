package utils

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateObject(client client.Client, updateObject client.Object) error {
	if err := client.Update(context.TODO(), updateObject); err != nil {
		return fmt.Errorf("failed to update %s (%s/%s) %w", updateObject.GetObjectKind(), updateObject.GetNamespace(), updateObject.GetName(), err)
	}
	return nil
}

func UpdateObjectStatus(client client.Client, updateObject client.Object) error {
	if err := client.Status().Update(context.TODO(), updateObject); err != nil {
		return fmt.Errorf("failed to update %s (%s/%s) status %w", updateObject.GetObjectKind(), updateObject.GetNamespace(), updateObject.GetName(), err)
	}
	return nil
}
