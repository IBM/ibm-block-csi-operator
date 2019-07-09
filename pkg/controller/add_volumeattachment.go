package controller

import (
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/controller/volumeattachment"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, volumeattachment.Add)
}
