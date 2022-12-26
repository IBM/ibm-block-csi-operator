package utils

import (
	"context"
	"fmt"

	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/go-logr/logr"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getStorageClassProvisioner(logger logr.Logger, client client.Client, scName string) (string, error) {
	sc, err := getStorageClass(logger, client, scName)
	if err != nil {
		return "", err
	}
	return sc.Provisioner, nil
}

func getStorageClass(logger logr.Logger, client client.Client, scName string) (*storagev1.StorageClass, error) {
	sc := &storagev1.StorageClass{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: scName}, sc)
	if err != nil {
		logger.Error(err, fmt.Sprintf(messages.FailedToGetStorageClass, scName))
		return nil, err
	}
	return sc, nil
}
