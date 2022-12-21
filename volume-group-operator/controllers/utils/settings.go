package utils

const (
	VolumeGroupNamePrefix                 = "volumegroup"
	volumeGroupGroupName                  = "volumegroup.storage.ibm.io"
	VolumeGroupFinalizer                  = volumeGroupGroupName
	VolumeGroupAsPrefix                   = volumeGroupGroupName + "/"
	volumeGroupContentFinalizer           = VolumeGroupAsPrefix + "vgc-protection"
	pvcVolumeGroupFinalizer               = VolumeGroupAsPrefix + "pvc-protection"
	PrefixedVolumeGroupSecretNameKey      = VolumeGroupAsPrefix + "secret-name"      // name key for secret
	PrefixedVolumeGroupSecretNamespaceKey = VolumeGroupAsPrefix + "secret-namespace" // namespace key secret
)
