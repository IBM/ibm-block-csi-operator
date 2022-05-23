/**
 * Copyright 2022 IBM Corp.
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

package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/IBM/ibm-block-csi-operator/controllers/internal/controller_instance"
	"github.com/IBM/ibm-block-csi-operator/controllers/util"
	oconfig "github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ControllerHelper struct {
	client.Client
	Log logr.Logger
}

func NewControllerHelper(client client.Client, log logr.Logger) *ControllerHelper {
	return &ControllerHelper{
		Client: client,
		Log:    log,
	}
}

func (ch *ControllerHelper) DeleteClusterRoleBindings(clusterRoleBindings []*rbacv1.ClusterRoleBinding) error {
	logger := ch.Log.WithName("deleteClusterRoleBindings")
	for _, crb := range clusterRoleBindings {
		found, err := ch.getClusterRoleBinding(crb)
		if err != nil && errors.IsNotFound(err) {
			continue
		} else if err != nil {
			logger.Error(err, "failed to get ClusterRoleBinding", "Name", crb.GetName())
			return err
		} else {
			logger.Info("deleting ClusterRoleBinding", "Name", crb.GetName())
			if err := ch.Delete(context.TODO(), found); err != nil {
				logger.Error(err, "failed to delete ClusterRoleBinding", "Name", crb.GetName())
				return err
			}
		}
	}
	return nil
}

func (ch *ControllerHelper) ReconcileClusterRoleBinding(clusterRoleBindings []*rbacv1.ClusterRoleBinding) error {
	logger := ch.Log.WithValues("Resource Type", "ClusterRoleBinding")
	for _, crb := range clusterRoleBindings {
		_, err := ch.getClusterRoleBinding(crb)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating a new ClusterRoleBinding", "Name", crb.GetName())
			err = ch.Create(context.TODO(), crb)
			if err != nil {
				return err
			}
		} else if err != nil {
			logger.Error(err, "Failed to get ClusterRole", "Name", crb.GetName())
			return err
		} else {
			// Resource already exists - don't requeue
			//ch.Log.Info("Skip reconcile: ClusterRoleBinding already exists", "Name", crb.GetName())
		}
	}
	return nil
}

func (ch *ControllerHelper) getClusterRoleBinding(crb *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	found := &rbacv1.ClusterRoleBinding{}
	err := ch.Get(context.TODO(), types.NamespacedName{
		Name:      crb.Name,
		Namespace: crb.Namespace,
	}, found)
	return found, err
}

func (ch *ControllerHelper) DeleteClusterRoles(clusterRoles []*rbacv1.ClusterRole) error {
	logger := ch.Log.WithName("deleteClusterRoles")
	for _, cr := range clusterRoles {
		found, err := ch.getClusterRole(cr)
		if err != nil && errors.IsNotFound(err) {
			continue
		} else if err != nil {
			logger.Error(err, "failed to get ClusterRole", "Name", cr.GetName())
			return err
		} else {
			logger.Info("deleting ClusterRole", "Name", cr.GetName())
			if err := ch.Delete(context.TODO(), found); err != nil {
				logger.Error(err, "failed to delete ClusterRole", "Name", cr.GetName())
				return err
			}
		}
	}
	return nil
}

func (ch *ControllerHelper) ReconcileClusterRole(clusterRoles []*rbacv1.ClusterRole) error {
	logger := ch.Log.WithValues("Resource Type", "ClusterRole")
	for _, cr := range clusterRoles {
		_, err := ch.getClusterRole(cr)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating a new ClusterRole", "Name", cr.GetName())
			err = ch.Create(context.TODO(), cr)
			if err != nil {
				return err
			}
		} else if err != nil {
			logger.Error(err, "Failed to get ClusterRole", "Name", cr.GetName())
			return err
		} else {
			err = ch.Update(context.TODO(), cr)
			if err != nil {
				logger.Error(err, "Failed to update ClusterRole", "Name", cr.GetName())
				return err
			}
		}
	}
	return nil
}

func (ch *ControllerHelper) getClusterRole(cr *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error) {
	found := &rbacv1.ClusterRole{}
	err := ch.Get(context.TODO(), types.NamespacedName{
		Name:      cr.GetName(),
		Namespace: cr.GetNamespace(),
	}, found)
	return found, err
}

func (ch *ControllerHelper) HasFinalizer(instance controller_instance.Instance) (bool, error) {
	accessor, finalizerName, err := ch.getAccessorAndFinalizerName(instance)
	if err != nil {
		return false, err
	}

	return util.Contains(accessor.GetFinalizers(), finalizerName), nil
}

func (ch *ControllerHelper) AddFinalizerIfNotPresent(instance controller_instance.Instance,
	unwrappedInstance client.Object) error {
	logger := ch.Log.WithName("addFinalizerIfNotPresent")

	accessor, finalizerName, err := ch.getAccessorAndFinalizerName(instance)
	if err != nil {
		return err
	}

	if !util.Contains(accessor.GetFinalizers(), finalizerName) {
		logger.Info("adding", "finalizer", finalizerName, "on", accessor.GetName())
		accessor.SetFinalizers(append(accessor.GetFinalizers(), finalizerName))

		if err := ch.Update(context.TODO(), unwrappedInstance); err != nil {
			logger.Error(err, "failed to add", "finalizer", finalizerName, "on", accessor.GetName())
			return err
		}
	}
	return nil
}

func (ch *ControllerHelper) RemoveFinalizer(instance controller_instance.Instance,
	unwrappedInstance client.Object) error {
	logger := ch.Log.WithName("removeFinalizer")

	accessor, finalizerName, err := ch.getAccessorAndFinalizerName(instance)
	if err != nil {
		return err
	}

	accessor.SetFinalizers(util.Remove(accessor.GetFinalizers(), finalizerName))
	if err := ch.Update(context.TODO(), unwrappedInstance); err != nil {
		logger.Error(err, "failed to remove", "finalizer", finalizerName, "from", accessor.GetName())
		return err
	}
	return nil
}

func (ch *ControllerHelper) getAccessorAndFinalizerName(instance controller_instance.Instance) (metav1.Object, string, error) {
	logger := ch.Log.WithName("getAccessorAndFinalizerName")
	lowercaseKind := strings.ToLower(instance.GetObjectKind().GroupVersionKind().Kind)
	finalizerName := fmt.Sprintf("%s.%s", lowercaseKind, oconfig.APIGroup)

	accessor, err := meta.Accessor(instance)
	if err != nil {
		logger.Error(err, "failed to get meta information of instance")
		return nil, "", err
	}
	return accessor, finalizerName, nil
}
