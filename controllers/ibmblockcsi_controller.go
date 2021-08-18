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

package controllers

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
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
	"github.com/IBM/ibm-block-csi-operator/controllers/internal/ibmblockcsi"
	clustersyncer "github.com/IBM/ibm-block-csi-operator/controllers/syncer"
	"github.com/IBM/ibm-block-csi-operator/controllers/util"
	oconfig "github.com/IBM/ibm-block-csi-operator/pkg/config"
	kubeutil "github.com/IBM/ibm-block-csi-operator/pkg/util/kubernetes"
	oversion "github.com/IBM/ibm-block-csi-operator/version"
	"github.com/go-logr/logr"
	"github.com/presslabs/controller-util/syncer"
)

// ReconcileTime is the delay between reconciliations
const ReconcileTime = 30 * time.Second

// ticket to remove those vars - CSI-3071
var daemonSetRestartedKey = ""
var daemonSetRestartedValue = ""

var log = logf.Log.WithName("ibmblockcsi_controller")

type reconciler func(instance *ibmblockcsi.IBMBlockCSI) error

// IBMBlockCSIReconciler reconciles a IBMBlockCSI object
type IBMBlockCSIReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	Namespace     string
	Recorder      record.EventRecorder
	ServerVersion string
}

// the rbac rule requires an empty row at the end to render
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;delete;list;watch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;create;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;update
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims/status,verbs=patch
//+kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;delete;list;watch;update;create;patch
//+kubebuilder:rbac:groups="",resources=events,verbs=*
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;statefulsets,verbs=get;list;watch;update;create;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=create;delete;get;watch;list
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=create;delete;get;watch;list;update
//+kubebuilder:rbac:groups=storage.k8s.io,resources=volumeattachments,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=storage.k8s.io,resources=volumeattachments/status,verbs=patch
//+kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=get;list;watch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create
//+kubebuilder:rbac:groups=apps,resourceNames=ibm-block-csi-operator,resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.k8s.io,resources=csidrivers,verbs=create;delete;get;watch;list
//+kubebuilder:rbac:groups=storage.k8s.io,resources=csinodes,verbs=get;list;watch
//+kubebuilder:rbac:groups=security.openshift.io,resourceNames=anyuid;privileged,resources=securitycontextconstraints,verbs=use
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=create;list;watch;delete
//+kubebuilder:rbac:groups=csi.ibm.com,resources=*,verbs=*
//+kubebuilder:rbac:groups=snapshot.storage.k8s.io,resources=volumesnapshotclasses,verbs=get;watch;list
//+kubebuilder:rbac:groups=snapshot.storage.k8s.io,resources=volumesnapshotcontents,verbs=get;watch;list;create;update;delete
//+kubebuilder:rbac:groups=snapshot.storage.k8s.io,resources=volumesnapshotcontents/status,verbs=update
//+kubebuilder:rbac:groups=snapshot.storage.k8s.io,resources=volumesnapshots,verbs=get;watch;list;update
func (r *IBMBlockCSIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling IBMBlockCSI")

	// Fetch the IBMBlockCSI instance
	instance := ibmblockcsi.New(&csiv1.IBMBlockCSI{}, r.ServerVersion)
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
		err = fmt.Errorf("wrong IBMBlockCSI options: %v", err)
		return reconcile.Result{RequeueAfter: ReconcileTime}, err
	}

	// update CR if there was changes after defaulting
	if changed {
		err = r.Update(context.TODO(), instance.Unwrap())
		if err != nil {
			err = fmt.Errorf("failed to update IBMBlockCSI CR: %v", err)
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

		if err := r.deleteCSIDriver(instance); err != nil {
			return reconcile.Result{}, err
		}

		if err := r.removeFinalizer(instance); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	originalStatus := *instance.Status.DeepCopy()

	// create the resources which never change if not exist
	for _, rec := range []reconciler{
		r.reconcileCSIDriver,
		r.reconcileServiceAccount,
		r.reconcileClusterRole,
		r.reconcileClusterRoleBinding,
	} {
		if err = rec(instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	// sync the resources which change over time
	csiControllerSyncer := clustersyncer.NewCSIControllerSyncer(r.Client, r.Scheme, instance)
	if err := syncer.Sync(context.TODO(), csiControllerSyncer, r.Recorder); err != nil {
		return reconcile.Result{}, err
	}

	csiNodeSyncer := clustersyncer.NewCSINodeSyncer(r.Client, r.Scheme, instance, daemonSetRestartedKey, daemonSetRestartedValue)
	if err := syncer.Sync(context.TODO(), csiNodeSyncer, r.Recorder); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.updateStatus(instance, originalStatus); err != nil {
		return reconcile.Result{}, err
	}

	// Resource created successfully - don't requeue
	return reconcile.Result{}, nil
}

func getServerVersion() (string, error) {
	kubeVersion, found := os.LookupEnv(oconfig.ENVKubeVersion)
	if found {
		return kubeVersion, nil
	}

	kubeClient := kubeutil.KubeClient

	serverVersion, err := kubeutil.ServerVersion(kubeClient.Discovery())
	if err != nil {
		return serverVersion, err
	}
	if strings.HasSuffix(serverVersion, "+") {
		serverVersion = strings.TrimSuffix(serverVersion, "+")
	}
	return serverVersion, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IBMBlockCSIReconciler) SetupWithManager(mgr ctrl.Manager) error {

	serverVersion, err := getServerVersion()
	if err != nil {
		panic(err)
	}

	log.Info(fmt.Sprintf("Kubernetes Version: %s", serverVersion))

	return ctrl.NewControllerManagedBy(mgr).
		For(&csiv1.IBMBlockCSI{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&corev1.ServiceAccount{}).
		Complete(r)
}

func (r *IBMBlockCSIReconciler) addFinalizerIfNotPresent(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithName("addFinalizerIfNotPresent")

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

func (r *IBMBlockCSIReconciler) hasFinalizer(instance *ibmblockcsi.IBMBlockCSI) (bool, error) {
	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		return false, err
	}

	return util.Contains(accessor.GetFinalizers(), finalizerName), nil
}

func (r *IBMBlockCSIReconciler) removeFinalizer(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithName("removeFinalizer")

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

func (r *IBMBlockCSIReconciler) getAccessorAndFinalizerName(instance *ibmblockcsi.IBMBlockCSI) (metav1.Object, string, error) {
	logger := log.WithName("getAccessorAndFinalizerName")
	lowercaseKind := strings.ToLower(instance.GetObjectKind().GroupVersionKind().Kind)
	finalizerName := fmt.Sprintf("%s.%s", lowercaseKind, oconfig.APIGroup)

	accessor, err := meta.Accessor(instance)
	if err != nil {
		logger.Error(err, "failed to get meta information of instance")
		return nil, "", err
	}
	return accessor, finalizerName, nil
}

func (r *IBMBlockCSIReconciler) updateStatus(instance *ibmblockcsi.IBMBlockCSI, originalStatus csiv1.IBMBlockCSIStatus) error {
	logger := log.WithName("updateStatus")
	controllerPod := &corev1.Pod{}
	controllerStatefulset, err := r.getControllerStatefulSet(instance)
	if err != nil {
		return err
	}

	nodeDaemonSet, err := r.getNodeDaemonSet(instance)
	if err != nil {
		return err
	}

	instance.Status.ControllerReady = r.isControllerReady(controllerStatefulset)
	instance.Status.NodeReady = r.isNodeReady(nodeDaemonSet)
	phase := csiv1.DriverPhaseNone
	if instance.Status.ControllerReady && instance.Status.NodeReady {
		phase = csiv1.DriverPhaseRunning
	} else {
		if !instance.Status.ControllerReady {
			err := r.getControllerPod(controllerStatefulset, controllerPod)
			if err != nil {
				logger.Error(err, "failed to get controller pod")
				return err
			}

			if !r.areAllPodImagesSynced(controllerStatefulset, controllerPod) {
				r.restartControllerPodfromStatefulSet(logger, controllerStatefulset, controllerPod)
			}
		}
		phase = csiv1.DriverPhaseCreating
	}
	instance.Status.Phase = phase
	instance.Status.Version = oversion.DriverVersion

	if !reflect.DeepEqual(originalStatus, instance.Status) {
		logger.Info("updating IBMBlockCSI status", "name", instance.Name, "from", originalStatus, "to", instance.Status)
		sErr := r.Status().Update(context.TODO(), instance.Unwrap())
		if sErr != nil {
			return sErr
		}
	}

	return nil
}

func (r *IBMBlockCSIReconciler) areAllPodImagesSynced(controllerStatefulset *appsv1.StatefulSet, controllerPod *corev1.Pod) bool {
	logger := log.WithName("areAllPodImagesSynced")
	statefulSetContainers := controllerStatefulset.Spec.Template.Spec.Containers
	podContainers := controllerPod.Spec.Containers
	if len(statefulSetContainers) != len(podContainers) {
		return false
	}
	for i := 0; i < len(statefulSetContainers); i++ {
		statefulSetImage := statefulSetContainers[i].Image
		podImage := podContainers[i].Image

		if statefulSetImage != podImage {
			logger.Info("csi controller image not in sync",
				"statefulSetImage", statefulSetImage, "podImage", podImage)
			return false
		}
	}
	return true
}

func (r *IBMBlockCSIReconciler) restartControllerPod(logger logr.Logger, instance *ibmblockcsi.IBMBlockCSI) error {
	controllerPod := &corev1.Pod{}
	controllerStatefulset, err := r.getControllerStatefulSet(instance)
	if err != nil {
		return err
	}

	logger.Info("controller requires restart",
		"ReadyReplicas", controllerStatefulset.Status.ReadyReplicas,
		"Replicas", controllerStatefulset.Status.Replicas)
	logger.Info("restarting csi controller")

	err = r.getControllerPod(controllerStatefulset, controllerPod)
	if errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		logger.Error(err, "failed to get controller pod")
		return err
	}

	return r.restartControllerPodfromStatefulSet(logger, controllerStatefulset, controllerPod)
}

func (r *IBMBlockCSIReconciler) restartControllerPodfromStatefulSet(logger logr.Logger,
	controllerStatefulset *appsv1.StatefulSet, controllerPod *corev1.Pod) error {
	logger.Info("controller requires restart",
		"ReadyReplicas", controllerStatefulset.Status.ReadyReplicas,
		"Replicas", controllerStatefulset.Status.Replicas)
	logger.Info("restarting csi controller")

	return r.Delete(context.TODO(), controllerPod)
}

func (r *IBMBlockCSIReconciler) getControllerPod(controllerStatefulset *appsv1.StatefulSet, controllerPod *corev1.Pod) error {
	controllerPodName := fmt.Sprintf("%s-0", controllerStatefulset.Name)
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      controllerPodName,
		Namespace: controllerStatefulset.Namespace,
	}, controllerPod)
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

func (r *IBMBlockCSIReconciler) rolloutRestartNode(node *appsv1.DaemonSet) error {
	restartedAt := fmt.Sprintf("%s/restartedAt", oconfig.APIGroup)
	timestamp := time.Now().String()
	node.Spec.Template.ObjectMeta.Annotations[restartedAt] = timestamp
	return r.Update(context.TODO(), node)
}

func (r *IBMBlockCSIReconciler) reconcileCSIDriver(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithValues("Resource Type", "CSIDriver")

	cd := instance.GenerateCSIDriver()
	found := &storagev1.CSIDriver{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      cd.Name,
		Namespace: "",
	}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new CSIDriver", "Name", cd.GetName())
		err = r.Create(context.TODO(), cd)
		if err != nil {
			return err
		}
	} else if err != nil {
		logger.Error(err, "Failed to get CSIDriver", "Name", cd.GetName())
		return err
	} else {
		// Resource already exists - don't requeue
	}

	return nil
}

func (r *IBMBlockCSIReconciler) reconcileServiceAccount(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithValues("Resource Type", "ServiceAccount")

	controller := instance.GenerateControllerServiceAccount()
	node := instance.GenerateNodeServiceAccount()

	controllerServiceAccountName := oconfig.GetNameForResource(oconfig.CSIControllerServiceAccount, instance.Name)
	nodeServiceAccountName := oconfig.GetNameForResource(oconfig.CSINodeServiceAccount, instance.Name)

	for _, sa := range []*corev1.ServiceAccount{
		controller,
		node,
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

			nodeDaemonSet, err := r.getNodeDaemonSet(instance)
			if err != nil {
				return err
			}

			if controllerServiceAccountName == sa.Name {
				rErr := r.restartControllerPod(logger, instance)

				if rErr != nil {
					return rErr
				}
			}
			if nodeServiceAccountName == sa.Name {
				logger.Info("node rollout requires restart",
					"DesiredNumberScheduled", nodeDaemonSet.Status.DesiredNumberScheduled,
					"NumberAvailable", nodeDaemonSet.Status.NumberAvailable)
				logger.Info("csi node stopped being ready - restarting it")
				rErr := r.rolloutRestartNode(nodeDaemonSet)

				if rErr != nil {
					return rErr
				}

				daemonSetRestartedKey, daemonSetRestartedValue = r.getRestartedAtAnnotation(nodeDaemonSet.Spec.Template.ObjectMeta.Annotations)
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

func (r *IBMBlockCSIReconciler) getRestartedAtAnnotation(Annotations map[string]string) (string, string) {
	restartedAt := fmt.Sprintf("%s/restartedAt", oconfig.APIGroup)
	for key, element := range Annotations {
		if key == restartedAt {
			return key, element
		}
	}
	return "", ""
}

func (r *IBMBlockCSIReconciler) getControllerStatefulSet(instance *ibmblockcsi.IBMBlockCSI) (*appsv1.StatefulSet, error) {
	controllerStatefulset := &appsv1.StatefulSet{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      oconfig.GetNameForResource(oconfig.CSIController, instance.Name),
		Namespace: instance.Namespace,
	}, controllerStatefulset)

	return controllerStatefulset, err
}

func (r *IBMBlockCSIReconciler) getNodeDaemonSet(instance *ibmblockcsi.IBMBlockCSI) (*appsv1.DaemonSet, error) {
	node := &appsv1.DaemonSet{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      oconfig.GetNameForResource(oconfig.CSINode, instance.Name),
		Namespace: instance.Namespace,
	}, node)

	return node, err
}

func (r *IBMBlockCSIReconciler) isControllerReady(controller *appsv1.StatefulSet) bool {
	return controller.Status.ReadyReplicas == controller.Status.Replicas
}

func (r *IBMBlockCSIReconciler) isNodeReady(node *appsv1.DaemonSet) bool {
	return node.Status.DesiredNumberScheduled == node.Status.NumberAvailable
}

func (r *IBMBlockCSIReconciler) reconcileClusterRole(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithValues("Resource Type", "ClusterRole")

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

func (r *IBMBlockCSIReconciler) deleteClusterRolesAndBindings(instance *ibmblockcsi.IBMBlockCSI) error {
	if err := r.deleteClusterRoleBindings(instance); err != nil {
		return err
	}

	if err := r.deleteClusterRoles(instance); err != nil {
		return err
	}
	return nil
}

func (r *IBMBlockCSIReconciler) deleteClusterRoles(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithName("deleteClusterRoles")

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

func (r *IBMBlockCSIReconciler) getClusterRoles(instance *ibmblockcsi.IBMBlockCSI) []*rbacv1.ClusterRole {
	externalProvisioner := instance.GenerateExternalProvisionerClusterRole()
	externalAttacher := instance.GenerateExternalAttacherClusterRole()
	externalSnapshotter := instance.GenerateExternalSnapshotterClusterRole()
	externalResizer := instance.GenerateExternalResizerClusterRole()
	controllerSCC := instance.GenerateSCCForControllerClusterRole()
	nodeSCC := instance.GenerateSCCForNodeClusterRole()

	return []*rbacv1.ClusterRole{
		externalProvisioner,
		externalAttacher,
		externalSnapshotter,
		externalResizer,
		controllerSCC,
		nodeSCC,
	}
}

func (r *IBMBlockCSIReconciler) reconcileClusterRoleBinding(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithValues("Resource Type", "ClusterRoleBinding")

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

func (r *IBMBlockCSIReconciler) deleteClusterRoleBindings(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithName("deleteClusterRoleBindings")

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

func (r *IBMBlockCSIReconciler) getClusterRoleBindings(instance *ibmblockcsi.IBMBlockCSI) []*rbacv1.ClusterRoleBinding {
	externalProvisioner := instance.GenerateExternalProvisionerClusterRoleBinding()
	externalAttacher := instance.GenerateExternalAttacherClusterRoleBinding()
	externalSnapshotter := instance.GenerateExternalSnapshotterClusterRoleBinding()
	externalResizer := instance.GenerateExternalResizerClusterRoleBinding()
	controllerSCC := instance.GenerateSCCForControllerClusterRoleBinding()
	nodeSCC := instance.GenerateSCCForNodeClusterRoleBinding()

	return []*rbacv1.ClusterRoleBinding{
		externalProvisioner,
		externalAttacher,
		externalSnapshotter,
		externalResizer,
		controllerSCC,
		nodeSCC,
	}
}

func (r *IBMBlockCSIReconciler) deleteCSIDriver(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithName("deleteCSIDriver")

	csiDriver := instance.GenerateCSIDriver()
	found := &storagev1.CSIDriver{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      csiDriver.Name,
		Namespace: csiDriver.Namespace,
	}, found)
	if err == nil {
		logger.Info("deleting CSIDriver", "Name", csiDriver.GetName())
		if err := r.Delete(context.TODO(), found); err != nil {
			logger.Error(err, "failed to delete CSIDriver", "Name", csiDriver.GetName())
			return err
		}
	} else if errors.IsNotFound(err) {
		return nil
	} else {
		logger.Error(err, "failed to get CSIDriver", "Name", csiDriver.GetName())
		return err
	}
	return nil
}
