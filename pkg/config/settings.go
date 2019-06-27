package config

const (
	ClusterDriverRegistrarImage = "quay.io/k8scsi/csi-cluster-driver-registrar:v1.0.1"
	CSIProvisionerImage         = "quay.io/k8scsi/csi-provisioner:v1.1.1"
	CSIAttacherImage            = "quay.io/k8scsi/csi-attacher:v1.1.1"
	CSILivenessProbeImage       = "quay.io/k8scsi/livenessprobe:v1.1.0"

	ControllerTag = "1.0.0"
	NodeTag       = "1.0.0"
)
