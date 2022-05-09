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

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/IBM/ibm-block-csi-operator/controllers/internal/hostdefinition"
	clustersyncer "github.com/IBM/ibm-block-csi-operator/controllers/syncer"
	"github.com/IBM/ibm-block-csi-operator/controllers/util"
	oconfig "github.com/IBM/ibm-block-csi-operator/pkg/config"
	oversion "github.com/IBM/ibm-block-csi-operator/version"
	"github.com/go-logr/logr"
	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	csiv1 "github.com/IBM/ibm-block-csi-operator/api/v1"
)

var hostDefinitionLog = logf.Log.WithName("hostdefinition_controller")

type hostDefinitionReconciler func(instance *hostdefinition.HostDefinition) error

type HostDefinitionReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *HostDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (reconcile.Result, error) {
	reqLogger := hostDefinitionLog.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling HostDefinition")

	instance := hostdefinition.New(&csiv1.HostDefinition{})
	err := r.Get(context.TODO(), req.NamespacedName, instance.Unwrap())
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	r.Scheme.Default(instance.Unwrap())
	changed := instance.SetDefaults()
	if err := instance.Validate(); err != nil {
		err = fmt.Errorf("wrong HostDefinition options: %v", err)
		return reconcile.Result{RequeueAfter: ReconcileTime}, err
	}
	if changed {
		err = r.Update(context.TODO(), instance.Unwrap())
		if err != nil {
			err = fmt.Errorf("failed to update HostDefinition CR: %v", err)
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}
	if err := r.addFinalizerIfNotPresent(instance); err != nil {
		return reconcile.Result{}, err
	}

	if !instance.GetDeletionTimestamp().IsZero() {
		isFinalizerExists, err := r.hasFinalizer(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		if !isFinalizerExists {
			return reconcile.Result{}, nil
		}

		if err := r.deleteClusterRolesAndBindings(instance); err != nil {
			return reconcile.Result{}, err
		}

		if err := r.removeFinalizer(instance); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}
	originalStatus := *instance.Status.DeepCopy()

	for _, rec := range []hostDefinitionReconciler{
		r.reconcileServiceAccount,
		r.reconcileClusterRole,
		r.reconcileClusterRoleBinding,
	} {
		if err = rec(instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	csiControllerSyncer := clustersyncer.NewCSIHostDefinitionSyncer(r.Client, r.Scheme, instance)
	if err := syncer.Sync(context.TODO(), csiControllerSyncer, r.Recorder); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.updateStatus(instance, originalStatus); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *HostDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&csiv1.HostDefinition{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Complete(r)
}

func (r *HostDefinitionReconciler) addFinalizerIfNotPresent(instance *hostdefinition.HostDefinition) error {
	logger := hostDefinitionLog.WithName("addFinalizerIfNotPresent")

	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		return err
	}

	if !util.Contains(accessor.GetFinalizers(), finalizerName) {
		logger.Info("adding", "finalizer", finalizerName, "on", accessor.GetName())
		accessor.SetFinalizers(append(accessor.GetFinalizers(), finalizerName))

		if err := r.Update(context.TODO(), instance.Unwrap()); err != nil {
			logger.Error(err, "failed to add", "finalizer", finalizerName, "on", accessor.GetName())
			return err
		}
	}
	return nil
}

func (r *HostDefinitionReconciler) hasFinalizer(instance *hostdefinition.HostDefinition) (bool, error) {
	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		return false, err
	}

	return util.Contains(accessor.GetFinalizers(), finalizerName), nil
}

func (r *HostDefinitionReconciler) removeFinalizer(instance *hostdefinition.HostDefinition) error {
	logger := hostDefinitionLog.WithName("removeFinalizer")

	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		return err
	}

	accessor.SetFinalizers(util.Remove(accessor.GetFinalizers(), finalizerName))
	if err := r.Update(context.TODO(), instance.Unwrap()); err != nil {
		logger.Error(err, "failed to remove", "finalizer", finalizerName, "from", accessor.GetName())
		return err
	}
	return nil
}

func (r *HostDefinitionReconciler) getAccessorAndFinalizerName(instance *hostdefinition.HostDefinition) (metav1.Object, string, error) {
	logger := hostDefinitionLog.WithName("getAccessorAndFinalizerName")
	lowercaseKind := strings.ToLower(instance.GetObjectKind().GroupVersionKind().Kind)
	finalizerName := fmt.Sprintf("%s.%s", lowercaseKind, oconfig.APIGroup)

	accessor, err := meta.Accessor(instance)
	if err != nil {
		logger.Error(err, "failed to get meta information of instance")
		return nil, "", err
	}
	return accessor, finalizerName, nil
}

func (r *HostDefinitionReconciler) deleteClusterRolesAndBindings(instance *hostdefinition.HostDefinition) error {
	if err := r.deleteClusterRoleBindings(instance); err != nil {
		return err
	}

	if err := r.deleteClusterRoles(instance); err != nil {
		return err
	}
	return nil
}

func (r *HostDefinitionReconciler) reconcileClusterRoleBinding(instance *hostdefinition.HostDefinition) error {
	logger := hostDefinitionLog.WithValues("Resource Type", "ClusterRoleBinding")

	clusterRoleBindings := r.getClusterRoleBindings(instance)

	for _, crb := range clusterRoleBindings {
		found := &rbacv1.ClusterRoleBinding{}
		err := r.Get(context.TODO(), types.NamespacedName{
			Name:      crb.Name,
			Namespace: crb.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating a new ClusterRoleBinding", "Name", crb.GetName())
			err = r.Create(context.TODO(), crb)
			if err != nil {
				return err
			}
		} else if err != nil {
			logger.Error(err, "Failed to get ClusterRole", "Name", crb.GetName())
			return err
		} else {
			// Resource already exists - don't requeue
			//logger.Info("Skip reconcile: ClusterRoleBinding already exists", "Name", crb.GetName())
		}
	}
	return nil
}

func (r *HostDefinitionReconciler) deleteClusterRoleBindings(instance *hostdefinition.HostDefinition) error {
	logger := hostDefinitionLog.WithName("deleteClusterRoleBindings")

	clusterRoleBindings := r.getClusterRoleBindings(instance)

	for _, crb := range clusterRoleBindings {
		found := &rbacv1.ClusterRoleBinding{}
		err := r.Get(context.TODO(), types.NamespacedName{
			Name:      crb.Name,
			Namespace: crb.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			continue
		} else if err != nil {
			logger.Error(err, "failed to get ClusterRoleBinding", "Name", crb.GetName())
			return err
		} else {
			logger.Info("deleting ClusterRoleBinding", "Name", crb.GetName())
			if err := r.Delete(context.TODO(), found); err != nil {
				logger.Error(err, "failed to delete ClusterRoleBinding", "Name", crb.GetName())
				return err
			}
		}
	}
	return nil
}

func (r *HostDefinitionReconciler) getClusterRoleBindings(instance *hostdefinition.HostDefinition) []*rbacv1.ClusterRoleBinding {
	hostdefinition := instance.GenerateHostDefinitionClusterRoleBinding()

	return []*rbacv1.ClusterRoleBinding{
		hostdefinition,
	}
}

func (r *HostDefinitionReconciler) reconcileClusterRole(instance *hostdefinition.HostDefinition) error {
	logger := hostDefinitionLog.WithValues("Resource Type", "ClusterRole")

	clusterRoles := r.getClusterRoles(instance)

	for _, cr := range clusterRoles {
		found := &rbacv1.ClusterRole{}
		err := r.Get(context.TODO(), types.NamespacedName{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating a new ClusterRole", "Name", cr.GetName())
			err = r.Create(context.TODO(), cr)
			if err != nil {
				return err
			}
		} else if err != nil {
			logger.Error(err, "Failed to get ClusterRole", "Name", cr.GetName())
			return err
		} else {
			err = r.Update(context.TODO(), cr)
			if err != nil {
				logger.Error(err, "Failed to update ClusterRole", "Name", cr.GetName())
				return err
			}
		}
	}

	return nil
}

func (r *HostDefinitionReconciler) deleteClusterRoles(instance *hostdefinition.HostDefinition) error {
	logger := hostDefinitionLog.WithName("deleteClusterRoles")

	clusterRoles := r.getClusterRoles(instance)

	for _, cr := range clusterRoles {
		found := &rbacv1.ClusterRole{}
		err := r.Get(context.TODO(), types.NamespacedName{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			continue
		} else if err != nil {
			logger.Error(err, "failed to get ClusterRole", "Name", cr.GetName())
			return err
		} else {
			logger.Info("deleting ClusterRole", "Name", cr.GetName())
			if err := r.Delete(context.TODO(), found); err != nil {
				logger.Error(err, "failed to delete ClusterRole", "Name", cr.GetName())
				return err
			}
		}
	}
	return nil
}

func (r *HostDefinitionReconciler) getClusterRoles(instance *hostdefinition.HostDefinition) []*rbacv1.ClusterRole {
	hostdefinition := instance.GenerateHostDefinitionClusterRole()

	return []*rbacv1.ClusterRole{
		hostdefinition,
	}
}

func (r *HostDefinitionReconciler) reconcileServiceAccount(instance *hostdefinition.HostDefinition) error {
	logger := hostDefinitionLog.WithValues("Resource Type", "ServiceAccount")

	hostDefinition := instance.GenerateServiceAccount()

	for _, sa := range []*corev1.ServiceAccount{
		hostDefinition,
	} {
		if err := controllerutil.SetControllerReference(instance.Unwrap(), sa, r.Scheme); err != nil {
			return err
		}
		found := &corev1.ServiceAccount{}
		err := r.Get(context.TODO(), types.NamespacedName{
			Name:      sa.Name,
			Namespace: sa.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating a new ServiceAccount", "Namespace", sa.GetNamespace(), "Name", sa.GetName())
			err = r.Create(context.TODO(), sa)
			if err != nil {
				return err
			}

			rErr := r.restartDeployment(logger, instance)
			if rErr != nil {
				return rErr
			}

		} else if err != nil {
			logger.Error(err, "Failed to get ServiceAccount", "Name", sa.GetName())
			return err
		} else {
			// Resource already exists - don't requeue
			//logger.Info("Skip reconcile: ServiceAccount already exists", "Namespace", sa.GetNamespace(), "Name", sa.GetName())
		}
	}

	return nil
}

func (r *HostDefinitionReconciler) restartDeployment(logger logr.Logger, instance *hostdefinition.HostDefinition) error {
	deployment, err := r.getDeployment(instance)
	if err != nil {
		return err
	}

	logger.Info("hostDefinition requires restart",
		"ReadyReplicas", deployment.Status.ReadyReplicas,
		"Replicas", deployment.Status.Replicas)
	logger.Info("restarting csi hostDefinition")

	err = r.rolloutRestartDeployment(deployment)
	if err != nil {
		return err
	}
	return nil
}

func (r *HostDefinitionReconciler) getDeployment(instance *hostdefinition.HostDefinition) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      oconfig.GetNameForResource(oconfig.CSIHostDefinition, instance.Name),
		Namespace: instance.Namespace,
	}, deployment)

	return deployment, err
}

func (r *HostDefinitionReconciler) rolloutRestartDeployment(deployment *appsv1.Deployment) error {
	restartedAt := fmt.Sprintf("%s/restartedAt", oconfig.APIGroup)
	timestamp := time.Now().String()
	deployment.Spec.Template.ObjectMeta.Annotations[restartedAt] = timestamp
	return r.Update(context.TODO(), deployment)
}

func (r *HostDefinitionReconciler) updateStatus(instance *hostdefinition.HostDefinition, originalStatus csiv1.HostDefinitionStatus) error {
	logger := log.WithName("updateStatus")
	deployment, err := r.getDeployment(instance)
	if err != nil {
		return err
	}

	r.updateStatusFields(instance, deployment)

	if !reflect.DeepEqual(originalStatus, instance.Status) {
		logger.Info("updating IBMBlockCSI status", "name", instance.Name, "from", originalStatus, "to", instance.Status)
		sErr := r.Status().Update(context.TODO(), instance.Unwrap())
		if sErr != nil {
			return sErr
		}
	}

	return nil
}

func (r *HostDefinitionReconciler) updateStatusFields(instance *hostdefinition.HostDefinition, deployment *appsv1.Deployment) {
	instance.Status.HostDefinitionReady = r.isReady(deployment)
	phase := csiv1.DriverPhaseCreating
	if instance.Status.HostDefinitionReady {
		phase = csiv1.DriverPhaseRunning
	}
	instance.Status.Phase = phase
	instance.Status.Version = oversion.DriverVersion
}

func (r *HostDefinitionReconciler) isReady(deployment *appsv1.Deployment) bool {
	return deployment.Status.ReadyReplicas == deployment.Status.Replicas
}
