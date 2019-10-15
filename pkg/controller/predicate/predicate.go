/**
 * Copyright 2019 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package predicate

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var log = logf.Log.WithName("predicate").WithName("eventFilters")

// CreateDeletePredicate implements a default predicate on resource creation or deletion events
type CreateDeletePredicate struct {
	predicate.Funcs
}

// no watch for update events
func (p CreateDeletePredicate) Update(e event.UpdateEvent) bool {
	return false
}

// no watch for generic events
func (p CreateDeletePredicate) Generic(e event.GenericEvent) bool {
	return false
}

// CreatePredicate implements a default predicate on resource creation events
type CreatePredicate struct {
	predicate.Funcs
}

// no watch for update events
func (p CreatePredicate) Update(e event.UpdateEvent) bool {
	return false
}

// no watch for update events
func (p CreatePredicate) Delete(e event.DeleteEvent) bool {
	return false
}

// no watch for generic events
func (p CreatePredicate) Generic(e event.GenericEvent) bool {
	return false
}

// NodePredicate implements a predicate for node controller
// It only watches for create events which have Status.Addresses
// and update events which change Status.Addresses
// and delete events
type NodePredicate struct {
	predicate.ResourceVersionChangedPredicate
}

func (p NodePredicate) Create(e event.CreateEvent) bool {
	node, ok := e.Object.(*corev1.Node)
	if !ok {
		log.Error(nil, "New runtime object is not a node", "event", e)
		return false
	}
	return len(node.Status.Addresses) > 0
}

func (p NodePredicate) Update(e event.UpdateEvent) bool {
	if !p.ResourceVersionChangedPredicate.Update(e) {
		return false
	}

	oldNode, ok := e.ObjectOld.(*corev1.Node)
	if !ok {
		log.Error(nil, "Old runtime object is not a node", "event", e)
		return false
	}
	newNode, ok := e.ObjectNew.(*corev1.Node)
	if !ok {
		log.Error(nil, "New runtime object is not a node", "event", e)
		return false
	}

	return !reflect.DeepEqual(oldNode.Status.Addresses, newNode.Status.Addresses)
}

// no watch for generic events
func (p NodePredicate) Generic(e event.GenericEvent) bool {
	return false
}
