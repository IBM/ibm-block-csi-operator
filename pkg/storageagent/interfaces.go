package storageagent

type StoragClient interface {
	CreateHost(name string, iscsiPorts, fcPorts []string) error
}
