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

package errors

import (
	"fmt"

	"github.com/IBM/volume-group-operator/pkg/messages"
)

type MatchingLabelsAndLabelSelectorError struct {
	ErrorMessage string
}

func (e *MatchingLabelsAndLabelSelectorError) Error() string {
	return fmt.Sprintf(messages.MatchingLabelsAndLabelSelectorFailed, e.ErrorMessage)
}

type PersistentVolumeDoesNotExist struct {
	PVName       string
	PVNamespace  string
	ErrorMessage string
}

func (e *PersistentVolumeDoesNotExist) Error() string {
	return fmt.Sprintf(messages.PersistentVolumeDoesNotExist, e.PVName, e.PVNamespace, e.ErrorMessage)
}
