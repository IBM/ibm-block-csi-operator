package config

// Add a field here if it never changes, if it changes over time, put it to settings.go
const (
	Name             = "ibm-block-csi-operator"
	ProductName      = "ibm-block-csi"
	DeployPath       = "/deploy"
	DefaultNamespace = "kube-system"
	DefaultLogLevel  = "DEBUG"
	ControllerUserID = int64(9999)

	ControllerRepository = "ibmcom/ibm-block-csi-controller-driver"
	NodeRepository       = "ibmcom/ibm-block-csi-node-driver"

	ControllerSocketVolumeMountPath                       = "/var/lib/csi/sockets/pluginproxy/"
	ControllerLivenessProbeContainerSocketVolumeMountPath = "/csi"
	ControllerSocketPath                                  = "/var/lib/csi/sockets/pluginproxy/csi.sock"
	CSIEndpoint                                           = "unix:///var/lib/csi/sockets/pluginproxy/csi.sock"
)
