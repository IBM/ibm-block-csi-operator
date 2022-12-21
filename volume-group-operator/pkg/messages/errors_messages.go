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
	MatchingLabelsAndLabelSelectorFailed                 = "Could not check if labels are matched with labelSelector, got %s"
	FailedToRemovePersistentVolumeClaimFromVolumeGroup   = "Could not remove %s/%s persistentVolumeClaim from %s/%s volumeGroup"
	PersistentVolumeDoesNotExist                         = "%s/%s persistentVolume does not exist"
	UnExpectedPersistentVolumeClaimError                 = "Got an unexpected error while fetching PersistentVolumeClaim"
	FailedToRemovePersistentVolumeFromVolumeGroupContent = "Could not remove %s/%s persistentVolume from %s/%s volumeGroupContent"
	FaildToCreateEvent                                   = "Failed to create %s/%s event"
)
