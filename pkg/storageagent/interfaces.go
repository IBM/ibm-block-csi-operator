package storageagent

import pb "github.com/IBM/ibm-block-csi-driver-operator/pkg/storageagent/storageagent"

type StorageClient interface {
	CreateHost(name string, iscsiPorts, fcPorts []string) error
	ListIscsiTargets() ([]*pb.IscsiTarget, error)
}
