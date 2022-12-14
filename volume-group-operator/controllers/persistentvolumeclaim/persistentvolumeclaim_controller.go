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
	"context"

	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type PersistentVolumeClaimWatcher struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

func (r *PersistentVolumeClaimWatcher) Reconcile(_ context.Context, req reconcile.Request) (result reconcile.Result, err error) {
	result = reconcile.Result{}
	reqLogger := r.Log.WithValues(messages.RequestNamespace, req.Namespace, messages.RequestName, req.Name)
	reqLogger.Info(messages.ReconcilePersistentVolumeClaim)
	_, err = r.getPersistentVolumeClaim(reqLogger, req)
	if err != nil {
		if errors.IsNotFound(err) {
			return result, nil
		}
		return result, err
	}

	return result, nil
}

func (r PersistentVolumeClaimWatcher) getPersistentVolumeClaim(logger logr.Logger,
	req reconcile.Request) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, pvc)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, messages.PersistentVolumeClaimNotFound, persistentVolumeClaim, pvc)
		} else {
			logger.Error(err, messages.UnExpectedPersistentVolumeClaimError, persistentVolumeClaim, pvc)
		}

		return nil, err
	}

	return pvc, nil
}

func (r *PersistentVolumeClaimWatcher) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.LabelChangedPredicate{}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PersistentVolumeClaim{}, builder.WithPredicates(pvcPredicate)).
		WithEventFilter(pred).Complete(r)
}
