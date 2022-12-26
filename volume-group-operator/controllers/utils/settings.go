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
	letterBytes                           = "0123456789abcdefghijklmnopqrstuvwxyz"
	volumeGroupController                 = "volumeGroupController"
	warningEventType                      = "Warning"
	normalEventType                       = "Normal"
	storageClassVGParameter               = "volume_group"
)
