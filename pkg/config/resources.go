package config

import "fmt"

// ResourceName is the type for aliasing resources that will be created.
type ResourceName string

func (rn ResourceName) String() string {
	return string(rn)
}

const (
	CSIController               ResourceName = "csi-controller"
	CSINode                     ResourceName = "csi-node"
	CSIControllerServiceAccount ResourceName = "csi-controller-sa"
)

// GetNameForResource returns the name of a resource for a CSI driver
func GetNameForResource(name ResourceName, driverName string) string {
	switch name {
	case CSIController:
		return fmt.Sprintf("%s-controller", driverName)
	case CSINode:
		return fmt.Sprintf("%s-node", driverName)
	case CSIControllerServiceAccount:
		return fmt.Sprintf("%s-controller-sa", driverName)
	default:
		return fmt.Sprintf("%s-%s", driverName, name)
	}
}
