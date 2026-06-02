/**
 * Copyright 2026 IBM Corp.
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

	"github.com/IBM/ibm-block-csi-operator/controllers/internal/hostdefiner"
	clustersyncer "github.com/IBM/ibm-block-csi-operator/controllers/syncer"
	"github.com/IBM/ibm-block-csi-operator/controllers/util"
	oconfig "github.com/IBM/ibm-block-csi-operator/pkg/config"
	oversion "github.com/IBM/ibm-block-csi-operator/version"
	"github.com/go-logr/logr"
	"github.com/presslabs/controller-util/pkg/syncer"
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

var hostDefinerLog = logf.Log.WithName("hostdefiner_controller")

type hostDefinerReconciler func(instance *hostdefiner.HostDefiner) error

type HostDefinerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *HostDefinerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (reconcile.Result, error) {
	reqLogger := hostDefinerLog.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling HostDefiner")

	instance := hostdefiner.New(&csiv1.HostDefiner{})
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
	if changed {
		patchJson := []byte(fmt.Sprintf(`{"spec":{"hostDefiner":{"repository": "%s","tag": "%s"}}}`,
			instance.Spec.HostDefiner.Repository, instance.Spec.HostDefiner.Tag))
		err := r.patchCr(patchJson, instance)
		if err != nil {
			err = fmt.Errorf("failed to update HostDefiner CR: %v", err)
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

	// Check if this is an upgrade scenario (deployment already exists)
	// Only enforce upgrade ordering during upgrades, not initial deployment
	isUpgrade, err := r.isUpgradeScenario(instance)
	if err != nil {
		reqLogger.Error(err, "Failed to check if this is an upgrade scenario")
		return reconcile.Result{RequeueAfter: 30 * time.Second}, err
	}

	if isUpgrade {
		// Check if IBMBlockCSI controller and node are ready before proceeding with HostDefiner upgrade
		// This ensures proper upgrade ordering: controller -> node -> hostdefiner
		ready, requeueAfter, err := r.checkIBMBlockCSIReadiness(instance)
		if err != nil {
			reqLogger.Error(err, "Failed to check IBMBlockCSI readiness")
			return reconcile.Result{RequeueAfter: 30 * time.Second}, err
		}
		if !ready {
			reqLogger.Info("Waiting for IBMBlockCSI controller and node to be ready before upgrading HostDefiner",
				"RequeueAfter", requeueAfter)
			return reconcile.Result{RequeueAfter: requeueAfter}, nil
		}
	} else {
		reqLogger.Info("Initial HostDefiner deployment detected, proceeding without waiting for IBMBlockCSI")
	}

	originalStatus := *instance.Status.DeepCopy()

	for _, rec := range []hostDefinerReconciler{
		r.reconcileServiceAccount,
		r.reconcileClusterRole,
		r.reconcileClusterRoleBinding,
	} {
		if err = rec(instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	hostDefinerSyncer := clustersyncer.NewHostDefinerSyncer(r.Client, r.Scheme, instance)
	if err := syncer.Sync(context.TODO(), hostDefinerSyncer, r.Recorder); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.updateStatus(instance, originalStatus); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *HostDefinerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&csiv1.HostDefiner{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Complete(r)
}

func (r *HostDefinerReconciler) addFinalizerIfNotPresent(instance *hostdefiner.HostDefiner) error {
	logger := hostDefinerLog.WithName("addFinalizerIfNotPresent")

	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		return err
	}

	if !util.Contains(accessor.GetFinalizers(), finalizerName) {
		logger.Info("adding", "finalizer", finalizerName, "on", accessor.GetName())
		patchJson := []byte(fmt.Sprintf(`{"metadata":{"finalizers":["%s"]}}`, finalizerName))
		err := r.patchCr(patchJson, instance)
		if err != nil {
			logger.Error(err, "failed to add", "finalizer", finalizerName, "on", accessor.GetName())
			return err
		}
	}
	return nil
}

func (r *HostDefinerReconciler) patchCr(patchJson []byte, instance *hostdefiner.HostDefiner) error {
	patch := client.RawPatch(types.MergePatchType, patchJson)
	if err := r.Patch(context.TODO(), instance.Unwrap(), patch); err != nil {
		return err
	}
	return nil
}

func (r *HostDefinerReconciler) hasFinalizer(instance *hostdefiner.HostDefiner) (bool, error) {
	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		return false, err
	}

	return util.Contains(accessor.GetFinalizers(), finalizerName), nil
}

func (r *HostDefinerReconciler) removeFinalizer(instance *hostdefiner.HostDefiner) error {
	logger := hostDefinerLog.WithName("removeFinalizer")

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

func (r *HostDefinerReconciler) getAccessorAndFinalizerName(instance *hostdefiner.HostDefiner) (metav1.Object, string, error) {
	logger := hostDefinerLog.WithName("getAccessorAndFinalizerName")
	lowercaseKind := strings.ToLower(instance.GetObjectKind().GroupVersionKind().Kind)
	finalizerName := fmt.Sprintf("%s.%s", lowercaseKind, oconfig.APIGroup)

	accessor, err := meta.Accessor(instance)
	if err != nil {
		logger.Error(err, "failed to get meta information of instance")
		return nil, "", err
	}
	return accessor, finalizerName, nil
}

func (r *HostDefinerReconciler) deleteClusterRolesAndBindings(instance *hostdefiner.HostDefiner) error {
	if err := r.deleteClusterRoleBindings(instance); err != nil {
		return err
	}

	if err := r.deleteClusterRoles(instance); err != nil {
		return err
	}
	return nil
}

func (r *HostDefinerReconciler) reconcileClusterRoleBinding(instance *hostdefiner.HostDefiner) error {
	logger := hostDefinerLog.WithValues("Resource Type", "ClusterRoleBinding")

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

func (r *HostDefinerReconciler) deleteClusterRoleBindings(instance *hostdefiner.HostDefiner) error {
	logger := hostDefinerLog.WithName("deleteClusterRoleBindings")

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

func (r *HostDefinerReconciler) getClusterRoleBindings(instance *hostdefiner.HostDefiner) []*rbacv1.ClusterRoleBinding {
	hostdefiner := instance.GenerateHostDefinerClusterRoleBinding()

	return []*rbacv1.ClusterRoleBinding{
		hostdefiner,
	}
}

func (r *HostDefinerReconciler) reconcileClusterRole(instance *hostdefiner.HostDefiner) error {
	logger := hostDefinerLog.WithValues("Resource Type", "ClusterRole")

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

func (r *HostDefinerReconciler) deleteClusterRoles(instance *hostdefiner.HostDefiner) error {
	logger := hostDefinerLog.WithName("deleteClusterRoles")

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

func (r *HostDefinerReconciler) getClusterRoles(instance *hostdefiner.HostDefiner) []*rbacv1.ClusterRole {
	hostdefiner := instance.GenerateHostDefinerClusterRole()

	return []*rbacv1.ClusterRole{
		hostdefiner,
	}
}

func (r *HostDefinerReconciler) reconcileServiceAccount(instance *hostdefiner.HostDefiner) error {
	logger := hostDefinerLog.WithValues("Resource Type", "ServiceAccount")

	hostDefiner := instance.GenerateServiceAccount()

	for _, sa := range []*corev1.ServiceAccount{
		hostDefiner,
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
		}
	}

	return nil
}

func (r *HostDefinerReconciler) restartDeployment(logger logr.Logger, instance *hostdefiner.HostDefiner) error {
	deployment, err := r.getDeployment(instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	logger.Info("hostDefiner requires restart",
		"ReadyReplicas", deployment.Status.ReadyReplicas,
		"Replicas", deployment.Status.Replicas)
	logger.Info("restarting hostDefiner")

	err = r.rolloutRestartDeployment(deployment)
	if err != nil {
		return err
	}
	return nil
}

func (r *HostDefinerReconciler) getDeployment(instance *hostdefiner.HostDefiner) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      oconfig.GetNameForResource(oconfig.HostDefiner, instance.Name),
		Namespace: instance.Namespace,
	}, deployment)

	return deployment, err
}

func (r *HostDefinerReconciler) rolloutRestartDeployment(deployment *appsv1.Deployment) error {
	restartedAt := fmt.Sprintf("%s/restartedAt", oconfig.APIGroup)
	timestamp := time.Now().String()
	deployment.Spec.Template.ObjectMeta.Annotations[restartedAt] = timestamp
	return r.Update(context.TODO(), deployment)
}

func (r *HostDefinerReconciler) updateStatus(instance *hostdefiner.HostDefiner, originalStatus csiv1.HostDefinerStatus) error {
	logger := log.WithName("updateStatus")
	deployment, err := r.getDeployment(instance)
	if err != nil {
		return err
	}

	r.updateStatusFields(instance, deployment)

	if !reflect.DeepEqual(originalStatus, instance.Status) {
		logger.Info("updating HostDefiner status", "name", instance.Name, "from", originalStatus, "to", instance.Status)
		sErr := r.Status().Update(context.TODO(), instance.Unwrap())
		if sErr != nil {
			return sErr
		}
	}

	return nil
}

// isUpgradeScenario checks if the HostDefiner deployment already exists
// Returns true if deployment exists with replicas (upgrade scenario), false if not (initial deployment)
func (r *HostDefinerReconciler) isUpgradeScenario(instance *hostdefiner.HostDefiner) (bool, error) {
	deployment, err := r.getDeployment(instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Deployment doesn't exist yet - this is an initial deployment
			return false, nil
		}
		// Some other error occurred
		return false, err
	}

	// Deployment exists with replicas - this is an upgrade scenario
	return deployment.Status.ReadyReplicas > 0 || deployment.Status.Replicas > 0, nil
}

// checkIBMBlockCSIReadiness checks if the IBMBlockCSI controller and node are ready
// Returns: (ready bool, requeueAfter time.Duration, error)
// - ready: true if both controller and node are ready, false otherwise
// - requeueAfter: duration to wait before requeuing if not ready
// - error: any error encountered while checking readiness
func (r *HostDefinerReconciler) checkIBMBlockCSIReadiness(instance *hostdefiner.HostDefiner) (bool, time.Duration, error) {
	logger := hostDefinerLog.WithName("checkIBMBlockCSIReadiness")

	// Try to find IBMBlockCSI CR in the same namespace
	ibmBlockCSIList := &csiv1.IBMBlockCSIList{}
	err := r.List(context.TODO(), ibmBlockCSIList, client.InNamespace(instance.Namespace))
	if err != nil {
		logger.Error(err, "Failed to list IBMBlockCSI resources", "Namespace", instance.Namespace)
		return false, 30 * time.Second, err
	}

	// If no IBMBlockCSI CR found, allow HostDefiner to proceed
	// This handles cases where HostDefiner is deployed independently
	if len(ibmBlockCSIList.Items) == 0 {
		logger.Info("No IBMBlockCSI CR found in namespace, proceeding with HostDefiner reconciliation",
			"Namespace", instance.Namespace)
		return true, 0, nil
	}

	// Check the first IBMBlockCSI CR (typically there should be only one per namespace)
	ibmBlockCSI := &ibmBlockCSIList.Items[0]

	// First, check if an upgrade is in progress by comparing CR spec images with actual deployments
	// This catches upgrades before the status is updated
	upgradeInProgress, err := r.isIBMBlockCSIUpgradeInProgress(ibmBlockCSI)
	if err != nil {
		logger.Error(err, "Failed to check if IBMBlockCSI upgrade is in progress")
		return false, 30 * time.Second, err
	}

	if upgradeInProgress {
		logger.Info("IBMBlockCSI upgrade detected (image mismatch), waiting before upgrading HostDefiner",
			"IBMBlockCSI", ibmBlockCSI.Name)
		return false, 30 * time.Second, nil
	}

	// Check if both controller and node are ready
	if !ibmBlockCSI.Status.ControllerReady {
		logger.Info("IBMBlockCSI controller is not ready yet, waiting before upgrading HostDefiner",
			"IBMBlockCSI", ibmBlockCSI.Name,
			"ControllerReady", ibmBlockCSI.Status.ControllerReady,
			"Phase", ibmBlockCSI.Status.Phase)
		return false, 30 * time.Second, nil
	}

	if !ibmBlockCSI.Status.NodeReady {
		logger.Info("IBMBlockCSI node is not ready yet, waiting before upgrading HostDefiner",
			"IBMBlockCSI", ibmBlockCSI.Name,
			"NodeReady", ibmBlockCSI.Status.NodeReady,
			"Phase", ibmBlockCSI.Status.Phase)
		return false, 30 * time.Second, nil
	}

	// Both controller and node are ready
	logger.Info("IBMBlockCSI controller and node are ready, proceeding with HostDefiner reconciliation",
		"IBMBlockCSI", ibmBlockCSI.Name,
		"ControllerReady", ibmBlockCSI.Status.ControllerReady,
		"NodeReady", ibmBlockCSI.Status.NodeReady,
		"Phase", ibmBlockCSI.Status.Phase)

	return true, 0, nil
}

// isIBMBlockCSIUpgradeInProgress checks if IBMBlockCSI controller or node are being upgraded
// by comparing actual running pod images with StatefulSet/DaemonSet specs
func (r *HostDefinerReconciler) isIBMBlockCSIUpgradeInProgress(ibmBlockCSI *csiv1.IBMBlockCSI) (bool, error) {
	logger := hostDefinerLog.WithName("isIBMBlockCSIUpgradeInProgress")

	// Check if controller pod images match StatefulSet spec
	// This catches the case where StatefulSet is updated but pod hasn't been recreated yet
	controllerSS := &appsv1.StatefulSet{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      oconfig.GetNameForResource(oconfig.CSIController, ibmBlockCSI.Name),
		Namespace: ibmBlockCSI.Namespace,
	}, controllerSS)

	if err == nil {
		controllerPod := &corev1.Pod{}
		controllerPodName := fmt.Sprintf("%s-0", controllerSS.Name)
		err = r.Get(context.TODO(), types.NamespacedName{
			Name:      controllerPodName,
			Namespace: ibmBlockCSI.Namespace,
		}, controllerPod)

		if err == nil {
			// Compare pod images with StatefulSet spec images
			for _, ssContainer := range controllerSS.Spec.Template.Spec.Containers {
				for _, podContainer := range controllerPod.Spec.Containers {
					if ssContainer.Name == podContainer.Name && ssContainer.Image != podContainer.Image {
						logger.Info("Controller pod image not synced with StatefulSet",
							"container", ssContainer.Name,
							"statefulSetImage", ssContainer.Image,
							"podImage", podContainer.Image)
						return true, nil
					}
				}
			}
		}
	}

	// Check if node DaemonSet update is in progress
	nodeDaemonSet := &appsv1.DaemonSet{}
	err = r.Get(context.TODO(), types.NamespacedName{
		Name:      oconfig.GetNameForResource(oconfig.CSINode, ibmBlockCSI.Name),
		Namespace: ibmBlockCSI.Namespace,
	}, nodeDaemonSet)

	if err == nil {
		if nodeDaemonSet.Status.UpdatedNumberScheduled < nodeDaemonSet.Status.DesiredNumberScheduled {
			logger.Info("Node DaemonSet update in progress",
				"UpdatedNumberScheduled", nodeDaemonSet.Status.UpdatedNumberScheduled,
				"DesiredNumberScheduled", nodeDaemonSet.Status.DesiredNumberScheduled)
			return true, nil
		}
	}

	return false, nil
}

func (r *HostDefinerReconciler) updateStatusFields(instance *hostdefiner.HostDefiner, deployment *appsv1.Deployment) {
	instance.Status.HostDefinerReady = r.isReady(deployment)
	phase := csiv1.DriverPhaseCreating
	if instance.Status.HostDefinerReady {
		phase = csiv1.DriverPhaseRunning
	}
	instance.Status.Phase = phase
	instance.Status.Version = oversion.DriverVersion
}

func (r *HostDefinerReconciler) isReady(deployment *appsv1.Deployment) bool {
	return deployment.Status.ReadyReplicas == deployment.Status.Replicas
}
