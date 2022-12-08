package messages

var (
	ReconcilePersistentVolumeClaim       = "Reconciling PersistentVolumeClaim"
	RequestName                          = "Request.Name"
	RequestNamespace                     = "Request.Namespace"
	UnableToCreatePVCController          = "Unable to create persistentvolumeclaim controller"
	PersistentVolumeClaimNotFound        = "PersistentVolumeClaim not found"
	UnExpectedPersistentVolumeClaimError = "Got an unexpected error while fetching PersistentVolumeClaim"
)
