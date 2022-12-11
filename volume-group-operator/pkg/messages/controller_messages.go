package messages

var (
	ReconcilePersistentVolumeClaim                 = "Reconciling PersistentVolumeClaim"
	RequestName                                    = "Request.Name"
	RequestNamespace                               = "Request.Namespace"
	UnableToCreatePVCController                    = "Unable to create persistentvolumeclaim controller"
	PersistentVolumeClaimNotFound                  = "PersistentVolumeClaim not found"
	UnExpectedPersistentVolumeClaimError           = "Got an unexpected error while fetching PersistentVolumeClaim"
	ListVolumeGroups                               = "Listing volumeGroups"
	CheckIfPersistentVolumeClaimMatchesVolumeGroup = "Checking if %s/%s persistentVolumeClaim is matches %s/%s volumeGroup"
	PersistentVolumeClaimMatchedToVolumeGroup      = "%s/%s persistentVolumeClaim is matched with %s/%s volumeGroup"
	PersistentVolumeClaimNotMatchedToVolumeGroup   = "%s/%s persistentVolumeClaim is not matched with %s/%s volumeGroup"
	RemovePersistentVolumeClaimFromVolumeGroup     = "Removing %s/%s persistentVolumeClaim from %s/%s volumeGroup"
)
