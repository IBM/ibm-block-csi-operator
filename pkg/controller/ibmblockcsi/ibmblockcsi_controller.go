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

package ibmblockcsi

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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	oconfig "github.com/IBM/ibm-block-csi-operator/pkg/config"
	clustersyncer "github.com/IBM/ibm-block-csi-operator/pkg/controller/ibmblockcsi/syncer"
	"github.com/IBM/ibm-block-csi-operator/pkg/controller/util"
	"github.com/IBM/ibm-block-csi-operator/pkg/internal/ibmblockcsi"
	kubeutil "github.com/IBM/ibm-block-csi-operator/pkg/util/kubernetes"
	oversion "github.com/IBM/ibm-block-csi-operator/version"
	"github.com/presslabs/controller-util/syncer"
)

// ReconcileTime is the delay between reconciliations
const (
	ReconcileTime = 30 * time.Second
	Node = "node"
	Controller = "controller"
)

var ds_restarted_key = ""
var ds_restarted_value = ""

var log = logf.Log.WithName("ibmblockcsi_controller")

type reconciler func(instance *ibmblockcsi.IBMBlockCSI) error

// Add creates a new IBMBlockCSI Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func getServerVersion() (string, error) {
	kubeVersion, found := os.LookupEnv(oconfig.ENVKubeVersion)
	if found {
		return kubeVersion, nil
	}
	clientConfig, err := config.GetConfig()
	if err != nil {
		return "", err
	}

	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return "", err
	}

	serverVersion, err := kubeutil.ServerVersion(kubeClient.Discovery())
	if err != nil {
		return serverVersion, err
	}
	if strings.HasSuffix(serverVersion, "+") {
		serverVersion = strings.TrimSuffix(serverVersion, "+")
	}
	return serverVersion, nil
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {

	serverVersion, err := getServerVersion()
	if err != nil {
		panic(err)
	}

	log.Info(fmt.Sprintf("Kubernetes Version: %s", serverVersion))

	return &ReconcileIBMBlockCSI{
		client:        mgr.GetClient(),
		scheme:        mgr.GetScheme(),
		recorder:      mgr.GetEventRecorderFor("controller_ibmblockcsi"),
		serverVersion: serverVersion,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ibmblockcsi-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource IBMBlockCSI
	err = c.Watch(&source.Kind{Type: &csiv1.IBMBlockCSI{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	subresources := []runtime.Object{
		&appsv1.StatefulSet{},
		&appsv1.DaemonSet{},
		&corev1.ServiceAccount{},
	}

	for _, subresource := range subresources {
		err = c.Watch(&source.Kind{Type: subresource}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &csiv1.IBMBlockCSI{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileIBMBlockCSI implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileIBMBlockCSI{}

// ReconcileIBMBlockCSI reconciles a IBMBlockCSI object
type ReconcileIBMBlockCSI struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client        client.Client
	scheme        *runtime.Scheme
	recorder      record.EventRecorder
	serverVersion string
}

// Reconcile reads that state of the cluster for a IBMBlockCSI object and makes changes based on the state read
// and what is in the IBMBlockCSI.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIBMBlockCSI) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling IBMBlockCSI")

	// Fetch the IBMBlockCSI instance
	instance := ibmblockcsi.New(&csiv1.IBMBlockCSI{}, r.serverVersion)
	//instance := &csiv1.IBMBlockCSI{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance.Unwrap())
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

	r.scheme.Default(instance.Unwrap())
	changed := instance.SetDefaults()

	if err := instance.Validate(); err != nil {
		err = fmt.Errorf("wrong IBMBlockCSI options: %v", err)
		return reconcile.Result{RequeueAfter: ReconcileTime}, err
	}

	// update CR if there was changes after defaulting
	if changed {
		err = r.client.Update(context.TODO(), instance.Unwrap())
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
	csiControllerSyncer := clustersyncer.NewCSIControllerSyncer(r.client, r.scheme, instance)
	if err := syncer.Sync(context.TODO(), csiControllerSyncer, r.recorder); err != nil {
		return reconcile.Result{}, err
	}

	csiNodeSyncer := clustersyncer.NewCSINodeSyncer(r.client, r.scheme, instance, ds_restarted_key, ds_restarted_value)
	if err := syncer.Sync(context.TODO(), csiNodeSyncer, r.recorder); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.updateStatus(instance, originalStatus); err != nil {
		return reconcile.Result{}, err
	}

	// Resource created successfully - don't requeue
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMBlockCSI) addFinalizerIfNotPresent(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithName("addFinalizerIfNotPresent")

	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		return err
	}

	if !util.Contains(accessor.GetFinalizers(), finalizerName) {
		logger.Info("adding", "finalizer", finalizerName, "on", accessor.GetName())
		accessor.SetFinalizers(append(accessor.GetFinalizers(), finalizerName))

		if err := r.client.Update(context.TODO(), instance.Unwrap()); err != nil {
			logger.Error(err, "failed to add", "finalizer", finalizerName, "on", accessor.GetName())
			return err
		}
	}
	return nil
}

func (r *ReconcileIBMBlockCSI) hasFinalizer(instance *ibmblockcsi.IBMBlockCSI) (bool, error) {
	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		return false, err
	}

	return util.Contains(accessor.GetFinalizers(), finalizerName), nil
}

func (r *ReconcileIBMBlockCSI) removeFinalizer(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithName("removeFinalizer")

	accessor, finalizerName, err := r.getAccessorAndFinalizerName(instance)
	if err != nil {
		return err
	}

	accessor.SetFinalizers(util.Remove(accessor.GetFinalizers(), finalizerName))
	if err := r.client.Update(context.TODO(), instance.Unwrap()); err != nil {
		logger.Error(err, "failed to remove", "finalizer", finalizerName, "from", accessor.GetName())
		return err
	}
	return nil
}

func (r *ReconcileIBMBlockCSI) getAccessorAndFinalizerName(instance *ibmblockcsi.IBMBlockCSI) (metav1.Object, string, error) {
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

func (r *ReconcileIBMBlockCSI) updateStatus(instance *ibmblockcsi.IBMBlockCSI, originalStatus csiv1.IBMBlockCSIStatus) error {
	logger := log.WithName("updateStatus")
	controllerPod := &corev1.Pod{}
	err, controllerStatefulset := r.getControllerK8sObject(instance)
	if err != nil {
		return err
	}

	err, nodeDaemonSet := r.getNodeK8sObject(instance)
	if err != nil {
		return err
	}

	if !instance.Status.ControllerReady {
		err := r.getControllerPod(controllerStatefulset, controllerPod)
		if err != nil {
			logger.Error(err, "failed to get controller pod")
			return err
		}

		for _, containerstatus := range controllerPod.Status.ContainerStatuses {
			if containerstatus.State.Waiting != nil {
				if containerstatus.State.Waiting.Reason == "ImagePullBackOff" {
					logger.Info("controller requires restart",
											"ReadyReplicas", controllerStatefulset.Status.ReadyReplicas,
											"Replicas", controllerStatefulset.Status.Replicas)
					r.restartControllerPod(controllerPod)
				}
			}
		}
	}

	instance.Status.ControllerReady, instance.Status.NodeReady = r.getDriverPodsStatus(instance, 
		controllerStatefulset, nodeDaemonSet)
	phase := csiv1.DriverPhaseNone
	if instance.Status.ControllerReady && instance.Status.NodeReady {
		phase = csiv1.DriverPhaseRunning
	} else {
		phase = csiv1.DriverPhaseCreating
	}
	instance.Status.Phase = phase
	instance.Status.Version = oversion.DriverVersion
	logger.Info("updating IBMBlockCSI status")

	if !reflect.DeepEqual(originalStatus, instance.Status) {
		logger.Info("updating IBMBlockCSI status", "name", instance.Name, "from", originalStatus, "to", instance.Status)
		sErr := r.client.Status().Update(context.TODO(), instance.Unwrap())
		if sErr != nil {
			return sErr
		}
	}

	return nil
}

func (r *ReconcileIBMBlockCSI) areAllPodImagesSynced(controllerStatefulset *appsv1.StatefulSet, controllerPod *corev1.Pod) bool {
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

func (r *ReconcileIBMBlockCSI) restartControllerPod(controllerPod *corev1.Pod) error {
	return r.client.Delete(context.TODO(), controllerPod)
}

func (r *ReconcileIBMBlockCSI) getControllerPod(controllerStatefulset *appsv1.StatefulSet, controllerPod *corev1.Pod) error {
	controllerPodName := fmt.Sprintf("%s-0", controllerStatefulset.Name)
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      controllerPodName,
		Namespace: controllerStatefulset.Namespace,
	}, controllerPod)
	return err
}

func (r *ReconcileIBMBlockCSI) rolloutRestartNode(node *appsv1.DaemonSet) error {
	restartedAt := fmt.Sprintf("%s/restartedAt", oconfig.APIGroup)
	timestamp := time.Now().String()
	node.Spec.Template.ObjectMeta.Annotations[restartedAt] = timestamp
	return r.client.Update(context.TODO(), node)
}

func (r *ReconcileIBMBlockCSI) reconcileCSIDriver(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithValues("Resource Type", "CSIDriver")

	cd := instance.GenerateCSIDriver()
	found := &storagev1.CSIDriver{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      cd.Name,
		Namespace: "",
	}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new CSIDriver", "Name", cd.GetName())
		err = r.client.Create(context.TODO(), cd)
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

func (r *ReconcileIBMBlockCSI) reconcileServiceAccount(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithValues("Resource Type", "ServiceAccount")

	controller := instance.GenerateControllerServiceAccount()
	node := instance.GenerateNodeServiceAccount()

	for _, sa := range []*corev1.ServiceAccount{
		controller,
		node,
	} {
		if err := controllerutil.SetControllerReference(instance.Unwrap(), sa, r.scheme); err != nil {
			return err
		}
		found := &corev1.ServiceAccount{}
		err := r.client.Get(context.TODO(), types.NamespacedName{
			Name:      sa.Name,
			Namespace: sa.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating a new ServiceAccount", "Namespace", sa.GetNamespace(), "Name", sa.GetName())
			err = r.client.Create(context.TODO(), sa)
			if err != nil {
				return err
			}

			err, controllerStatefulset := r.getControllerK8sObject(instance)
			if err != nil {
				return err
			}

			err, nodeDaemonSet := r.getNodeK8sObject(instance)
			if err != nil {
				return err
			}

			if strings.Contains(sa.Name, Controller) {
				controllerlogger := log.WithValues("Resource Type", "Controller")
				controllerPod := &corev1.Pod{}
				err := r.getControllerPod(controllerStatefulset, controllerPod)
				if err != nil {
					controllerlogger.Error(err, "failed to get controller pod")
					return err
				}

				controllerlogger.Info("controller requires restart",
							"ReadyReplicas", controllerStatefulset.Status.ReadyReplicas,
							"Replicas", controllerStatefulset.Status.Replicas)
				controllerlogger.Info("restarting csi controller")
				rErr := r.restartControllerPod(controllerPod)
				
				if rErr != nil {
					return rErr
				}
			}
			if strings.Contains(sa.Name, Node) {
				nodelogger := log.WithValues("Resource Type", "Node DaemonSet")
				nodelogger.Info("node rollout requires restart",
				"DesiredNumberScheduled", nodeDaemonSet.Status.DesiredNumberScheduled,				
				"NumberAvailable", nodeDaemonSet.Status.NumberAvailable)
				nodelogger.Info("csi node stopped being ready - restarting it")
				rErr := r.rolloutRestartNode(nodeDaemonSet)

				if rErr != nil {
					return rErr
				}

				ds_restarted_key, ds_restarted_value = r.getRestartedAtAnnotation(nodeDaemonSet.Spec.Template.ObjectMeta.Annotations)
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

func (r *ReconcileIBMBlockCSI) getRestartedAtAnnotation(Annotations map[string]string) (string, string){
	restartedAt := fmt.Sprintf("%s/restartedAt", oconfig.APIGroup)
	for key, element := range Annotations {
		if key == restartedAt {
			return key, element
		}
	}
	return "", ""
}

func (r *ReconcileIBMBlockCSI) getControllerK8sObject(instance *ibmblockcsi.IBMBlockCSI) (error, *appsv1.StatefulSet) {
			controllerStatefulset := &appsv1.StatefulSet{}
			err := r.client.Get(context.TODO(), types.NamespacedName{
				Name:      oconfig.GetNameForResource(oconfig.CSIController, instance.Name),
				Namespace: instance.Namespace,
			}, controllerStatefulset)
		
			if err != nil {
				return err, controllerStatefulset
			}
			return err, controllerStatefulset
}

func (r *ReconcileIBMBlockCSI) getNodeK8sObject(instance *ibmblockcsi.IBMBlockCSI) (error, *appsv1.DaemonSet) {
	node := &appsv1.DaemonSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      oconfig.GetNameForResource(oconfig.CSINode, instance.Name),
		Namespace: instance.Namespace,
	}, node)

	if err != nil {
		return err, node
	}
	return err, node
}

func (r *ReconcileIBMBlockCSI) getDriverPodsStatus(instance *ibmblockcsi.IBMBlockCSI, 
	controller *appsv1.StatefulSet, node *appsv1.DaemonSet) (bool, bool) {
	ControllerReady := false
	NodeReady := false

	instance.Status.ControllerReady = controller.Status.ReadyReplicas == controller.Status.Replicas
	instance.Status.NodeReady = node.Status.DesiredNumberScheduled == node.Status.NumberAvailable
	if instance.Status.ControllerReady {
		ControllerReady = true
	}
	if instance.Status.NodeReady {
		NodeReady = true
	}
	return ControllerReady, NodeReady
}

func (r *ReconcileIBMBlockCSI) reconcileClusterRole(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithValues("Resource Type", "ClusterRole")

	clusterRoles := r.getClusterRoles(instance)

	for _, cr := range clusterRoles {
		found := &rbacv1.ClusterRole{}
		err := r.client.Get(context.TODO(), types.NamespacedName{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating a new ClusterRole", "Name", cr.GetName())
			err = r.client.Create(context.TODO(), cr)
			if err != nil {
				return err
			}
		} else if err != nil {
			logger.Error(err, "Failed to get ClusterRole", "Name", cr.GetName())
			return err
		} else {
			err = r.client.Update(context.TODO(), cr)
			if err != nil {
				logger.Error(err, "Failed to update ClusterRole", "Name", cr.GetName())
				return err
			}
		}
	}

	return nil
}

func (r *ReconcileIBMBlockCSI) deleteClusterRolesAndBindings(instance *ibmblockcsi.IBMBlockCSI) error {
	if err := r.deleteClusterRoleBindings(instance); err != nil {
		return err
	}

	if err := r.deleteClusterRoles(instance); err != nil {
		return err
	}
	return nil
}

func (r *ReconcileIBMBlockCSI) deleteClusterRoles(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithName("deleteClusterRoles")

	clusterRoles := r.getClusterRoles(instance)

	for _, cr := range clusterRoles {
		found := &rbacv1.ClusterRole{}
		err := r.client.Get(context.TODO(), types.NamespacedName{
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
			if err := r.client.Delete(context.TODO(), found); err != nil {
				logger.Error(err, "failed to delete ClusterRole", "Name", cr.GetName())
				return err
			}
		}
	}
	return nil
}

func (r *ReconcileIBMBlockCSI) getClusterRoles(instance *ibmblockcsi.IBMBlockCSI) []*rbacv1.ClusterRole {
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

func (r *ReconcileIBMBlockCSI) reconcileClusterRoleBinding(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithValues("Resource Type", "ClusterRoleBinding")

	clusterRoleBindings := r.getClusterRoleBindings(instance)

	for _, crb := range clusterRoleBindings {
		found := &rbacv1.ClusterRoleBinding{}
		err := r.client.Get(context.TODO(), types.NamespacedName{
			Name:      crb.Name,
			Namespace: crb.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating a new ClusterRoleBinding", "Name", crb.GetName())
			err = r.client.Create(context.TODO(), crb)
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

func (r *ReconcileIBMBlockCSI) deleteClusterRoleBindings(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithName("deleteClusterRoleBindings")

	clusterRoleBindings := r.getClusterRoleBindings(instance)

	for _, crb := range clusterRoleBindings {
		found := &rbacv1.ClusterRoleBinding{}
		err := r.client.Get(context.TODO(), types.NamespacedName{
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
			if err := r.client.Delete(context.TODO(), found); err != nil {
				logger.Error(err, "failed to delete ClusterRoleBinding", "Name", crb.GetName())
				return err
			}
		}
	}
	return nil
}

func (r *ReconcileIBMBlockCSI) getClusterRoleBindings(instance *ibmblockcsi.IBMBlockCSI) []*rbacv1.ClusterRoleBinding {
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

func (r *ReconcileIBMBlockCSI) deleteCSIDriver(instance *ibmblockcsi.IBMBlockCSI) error {
	logger := log.WithName("deleteCSIDriver")

	csiDriver := instance.GenerateCSIDriver()
	found := &storagev1.CSIDriver{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      csiDriver.Name,
		Namespace: csiDriver.Namespace,
	}, found)
	if err == nil {
		logger.Info("deleting CSIDriver", "Name", csiDriver.GetName())
		if err := r.client.Delete(context.TODO(), found); err != nil {
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
