package config

// Add a field here if it never changes, if it changes over time, put it to settings.go
const (
	APIGroup         = "csi.ibm.com"
	APIVersion       = "v1"
	Name             = "ibm-block-csi-operator"
	DriverName       = "ibm-block-csi-driver"
	ProductName      = "ibm-block-csi"
	DeployPath       = "/deploy"
	DefaultNamespace = "kube-system"
	DefaultLogLevel  = "DEBUG"
	ControllerUserID = int64(9999)

	ControllerRepository = "ibmcom/ibm-block-csi-controller-driver"
	NodeRepository       = "ibmcom/ibm-block-csi-node-driver"

	ControllerSocketVolumeMountPath                       = "/var/lib/csi/sockets/pluginproxy/"
	NodeSocketVolumeMountPath                             = "/csi"
	ControllerLivenessProbeContainerSocketVolumeMountPath = "/csi"
	ControllerSocketPath                                  = "/var/lib/csi/sockets/pluginproxy/csi.sock"
	NodeSocketPath                                        = "/csi/csi.sock"
	NodeRegistrarSocketPath                               = "/registration/csi.sock"
	CSIEndpoint                                           = "unix:///var/lib/csi/sockets/pluginproxy/csi.sock"
	CSINodeEndpoint                                       = "unix:///csi/csi.sock"
)
