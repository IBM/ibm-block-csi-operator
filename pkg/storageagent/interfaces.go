package storageagent

type StorageClient interface {
	CreateHost(name string, iscsiPorts, fcPorts []string) error
}
