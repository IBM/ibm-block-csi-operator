package config

import "fmt"

// ResourceName is the type for aliasing resources that will be created.
type ResourceName string

func (rn ResourceName) String() string {
	return string(rn)
}

const (
	CSIController                            ResourceName = "csi-controller"
	CSINode                                  ResourceName = "csi-node"
	CSIControllerServiceAccount              ResourceName = "csi-controller-sa"
	ExternalProvisionerClusterRole           ResourceName = "external-provisioner-clusterrole"
	ExternalProvisionerClusterRoleBinding    ResourceName = "external-provisioner-clusterrolebinding"
	ExternalAttacherClusterRole              ResourceName = "external-attacher-clusterrole"
	ExternalAttacherClusterRoleBinding       ResourceName = "external-attacher-clusterrolebinding"
	ClusterDriverRegistrarClusterRole        ResourceName = "cluster-driver-registrar-clusterrole"
	ClusterDriverRegistrarClusterRoleBinding ResourceName = "cluster-driver-registrar-clusterrolebinding"
	ExternalSnapshotterClusterRole           ResourceName = "external-snapshotter-clusterrole"
	ExternalSnapshotterClusterRoleBinding    ResourceName = "external-snapshotter-clusterrolebinding"
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
