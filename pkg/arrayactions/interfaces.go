package arrayactions

type ArrayMediator interface {
	CreateHost(name string, iscsiPorts, fcPorts []string) error
}
