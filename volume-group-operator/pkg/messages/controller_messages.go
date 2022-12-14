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
	RemovedPersistentVolumeClaimFromVolumeGroup    = "Successfully removed %s/%s persistentVolumeClaim from %s/%s volumeGroup"
)