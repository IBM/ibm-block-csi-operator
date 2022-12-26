/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package messages

var (
	ReconcilePersistentVolumeClaim                   = "Reconciling PersistentVolumeClaim"
	RequestName                                      = "Request.Name"
	RequestNamespace                                 = "Request.Namespace"
	UnableToCreatePVCController                      = "Unable to create persistentvolumeclaim controller"
	PersistentVolumeClaimNotFound                    = "%s/%s persistentVolumeClaim not found"
	ListVolumeGroups                                 = "Listing volumeGroups"
	CheckIfPersistentVolumeClaimMatchesVolumeGroup   = "Checking if %s/%s persistentVolumeClaim is matches %s/%s volumeGroup"
	PersistentVolumeClaimMatchedToVolumeGroup        = "%s/%s persistentVolumeClaim is matched with %s/%s volumeGroup"
	PersistentVolumeClaimNotMatchedToVolumeGroup     = "%s/%s persistentVolumeClaim is not matched with %s/%s volumeGroup"
	RemovePersistentVolumeClaimFromVolumeGroup       = "Removing %s/%s persistentVolumeClaim from %s/%s volumeGroup"
	RemovedPersistentVolumeClaimFromVolumeGroup      = "Successfully removed %s/%s persistentVolumeClaim from %s/%s volumeGroup"
	PersistentVolumeClaimDoesNotHavePersistentVolume = "PersistentVolumeClaim does not Have persistentVolume"
	GetPersistentVolumeOfPersistentVolumeClaim       = "Get matching persistentVolume from %s/%s persistentVolumeClaim"
	GetVolumeGroupContentOfVolumeGroup               = "Get matching volumeGroupContent from %s/%s VolumeGroup"
	RemovePersistentVolumeFromVolumeGroupContent     = "Removing %s persistentVolume from %s/%s volumeGroupContent"
	RemovedPersistentVolumeFromVolumeGroupContent    = "Successfully removed %s persistentVolume from %s/%s volumeGroupContent"
	FailedToModifyVolumeGroup                        = "Failed to modify %s/%s volumeGroup"
	AddPersistentVolumeClaimToVolumeGroup            = "Adding %s/%s persistentVolumeClaim to %s/%s volumeGroup"
	AddedPersistentVolumeClaimToVolumeGroup          = "Successfully added %s/%s persistentVolumeClaim to %s/%s volumeGroup"
	AddPersistentVolumeToVolumeGroupContent          = "Adding %s persistentVolume to %s/%s volumeGroup"
	AddedPersistentVolumeToVolumeGroupContent        = "Successfully added %s persistentVolume to %s/%s volumeGroupContent"
	RemoveVolumeFromVolumeGroup                      = "Removing volume of persistentVolumeClaim from %s/%s volumeGroup"
	RemovedVolumeFromVolumeGroup                     = "Successfully Removed volume of persistentVolumeClaim from %s/%s volumeGroup"
	ModifyVolumeGroup                                = "Modifying %s volumeGroupID with %v volumeIDs"
	ModifiedVolumeGroup                              = "Successfully modified %s volumeGroupID"
	CreateEventForNamespacedObject                   = "Creating event for %s/%s %s, with [%s] message"
	EventCreated                                     = "Successfully Created  %s/%s event"
	UpdateVolumeGroupStatus                          = "Updating status of %s/%s volumeGroup"
	GetPersistentVolumeClaim                         = "Getting %s/%s persistentVolumeClaim"
	GetPersistentVolume                              = "Getting %s persistentVolume"
	AddVolumeToVolumeGroup                           = "Adding volume of persistentVolumeClaim to %s/%s volumeGroup"
	AddedVolumeToVolumeGroup                         = "Successfully added volume of persistentVolumeClaim to %s/%s volumeGroup"
	PersistentVolumeClaimIsNotInBoundPhase           = "PersistentVolumeClaim is not in bound phase, stopping the reconcile, when it will be in bound phase, reconcile will continue"
	StorageClassHasVGParameter                       = "StorageClass %s contain parameter volume_group for claim %s/%s. volumegroup feature is not supported"
)
