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

package persistentvolumeclaim

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	pvcPredicate = predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isLabelsChanged(e.ObjectOld, e.ObjectNew) || isPhaseChanged(e.ObjectOld, e.ObjectNew)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
	removingPVC = "removePVC"
	addingPVC   = "addPVC"
)

func isLabelsChanged(oldObject, newObject client.Object) bool {
	return !reflect.DeepEqual(oldObject.(*corev1.PersistentVolumeClaim).Labels,
		newObject.(*corev1.PersistentVolumeClaim).Labels)
}

func isPhaseChanged(oldObject, newObject client.Object) bool {
	return !reflect.DeepEqual(oldObject.(*corev1.PersistentVolumeClaim).Status.Phase,
		newObject.(*corev1.PersistentVolumeClaim).Status.Phase)
}
