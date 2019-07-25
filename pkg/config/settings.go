package config

const (
	ClusterDriverRegistrarImage = "quay.io/k8scsi/csi-cluster-driver-registrar:v1.0.1"
	NodeDriverRegistrarImage    = "quay.io/k8scsi/csi-node-driver-registrar:v1.0.2"
	CSIProvisionerImage         = "quay.io/k8scsi/csi-provisioner:v1.1.1"
	CSIAttacherImage            = "quay.io/k8scsi/csi-attacher:v1.1.1"
	CSIAttacherImage_1_13       = "quay.io/k8scsi/csi-attacher:v1.0.1" // for k8s 1.13
	CSILivenessProbeImage       = "quay.io/k8scsi/livenessprobe:v1.1.0"

	ControllerTag = "1.0.0"
	NodeTag       = "1.0.0"

	DefaultNamespace = "kube-system"
	DefaultLogLevel  = "DEBUG"
	ControllerUserID = int64(9999)
)
