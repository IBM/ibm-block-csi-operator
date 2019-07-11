package ibmblockcsi

import (
	"context"
	"fmt"
	"reflect"
	"time"

	csiv1 "github.com/IBM/ibm-block-csi-driver-operator/pkg/apis/csi/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/IBM/ibm-block-csi-driver-operator/pkg/config"
	clustersyncer "github.com/IBM/ibm-block-csi-driver-operator/pkg/controller/ibmblockcsi/syncer"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/internal/ibmblockcsi"
	kubeutil "github.com/IBM/ibm-block-csi-driver-operator/pkg/util/kubernetes"
	"github.com/presslabs/controller-util/syncer"
)

// ReconcileTime is the delay between reconciliations
const ReconcileTime = 30 * time.Second

var log = logf.Log.WithName("controller_ibmblockcsi")

type reconciler func(instance *ibmblockcsi.IBMBlockCSI) error

// Add creates a new IBMBlockCSI Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func getServerVersion() (string, error) {
	clientConfig, err := kubeutil.KubeConfig()
	if err != nil {
		return "", err
	}

	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return "", err
	}

	return kubeutil.ServerVersion(kubeClient.Discovery())
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
		recorder:      mgr.GetRecorder("controller_ibmblockcsi"),
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
	}

	status := *instance.Status.DeepCopy()
	defer func() {
		if !reflect.DeepEqual(status, instance.Status) {
			reqLogger.Info("updating IBMBlockCSI status", "name", instance.Name)
			sErr := r.client.Status().Update(context.TODO(), instance.Unwrap())
			if sErr != nil {
				reqLogger.Error(sErr, "failed to update IBMBlockCSI status", "name", instance.Name)
			}
		}
	}()

	// create the resources which never change if not exist
	for _, rec := range []reconciler{
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

	csiNodeSyncer := clustersyncer.NewCSINodeSyncer(r.client, r.scheme, instance)
	if err := syncer.Sync(context.TODO(), csiNodeSyncer, r.recorder); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.updateStatus(instance); err != nil {
		return reconcile.Result{}, err
	}

	// Resource created successfully - don't requeue
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMBlockCSI) updateStatus(instance *ibmblockcsi.IBMBlockCSI) error {
	controller := &appsv1.StatefulSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      config.GetNameForResource(config.CSIController, instance.Name),
		Namespace: instance.Namespace,
	}, controller)

	if err != nil {
		return err
	}

	node := &appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      config.GetNameForResource(config.CSINode, instance.Name),
		Namespace: instance.Namespace,
	}, node)

	if err != nil {
		return err
	}

	instance.Status.ControllerReady = controller.Status.ReadyReplicas == controller.Status.Replicas
	instance.Status.NodeReady = node.Status.DesiredNumberScheduled == node.Status.NumberAvailable
	instance.Status.Ready = instance.Status.ControllerReady && instance.Status.NodeReady

	// no need to push to status to API Server here.
	return nil
}

func (r *ReconcileIBMBlockCSI) reconcileServiceAccount(instance *ibmblockcsi.IBMBlockCSI) error {
	recLogger := log.WithValues("Resource Type", "ServiceAccount")

	sa := instance.GenerateControllerServiceAccount()
	if err := controllerutil.SetControllerReference(instance.Unwrap(), sa, r.scheme); err != nil {
		return err
	}
	found := &corev1.ServiceAccount{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sa.Name,
		Namespace: sa.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		recLogger.Info("Creating a new ServiceAccount", "Namespace", sa.GetNamespace(), "Name", sa.GetName())
		err = r.client.Create(context.TODO(), sa)
		if err != nil {
			return err
		}
	} else if err != nil {
		recLogger.Error(err, "Failed to get ServiceAccount", "Name", sa.GetName())
		return err
	} else {
		// Resource already exists - don't requeue
		//recLogger.Info("Skip reconcile: ServiceAccount already exists", "Namespace", sa.GetNamespace(), "Name", sa.GetName())
	}

	return nil
}

func (r *ReconcileIBMBlockCSI) reconcileClusterRole(instance *ibmblockcsi.IBMBlockCSI) error {
	recLogger := log.WithValues("Resource Type", "ClusterRole")

	externalProvisioner := instance.GenerateExternalProvisionerClusterRole()
	externalAttacher := instance.GenerateExternalAttacherClusterRole()
	clusterDriverRegistrar := instance.GenerateClusterDriverRegistrarClusterRole()
	externalSnapshotter := instance.GenerateExternalSnapshotterClusterRole()

	for _, cr := range []*rbacv1.ClusterRole{
		externalProvisioner,
		externalAttacher,
		clusterDriverRegistrar,
		externalSnapshotter,
	} {
		if err := controllerutil.SetControllerReference(instance.Unwrap(), cr, r.scheme); err != nil {
			return err
		}
		found := &rbacv1.ClusterRole{}
		err := r.client.Get(context.TODO(), types.NamespacedName{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			recLogger.Info("Creating a new ClusterRole", "Name", cr.GetName())
			err = r.client.Create(context.TODO(), cr)
			if err != nil {
				return err
			}
		} else if err != nil {
			recLogger.Error(err, "Failed to get ClusterRole", "Name", cr.GetName())
			return err
		} else {
			// Resource already exists - don't requeue
			//recLogger.Info("Skip reconcile: ClusterRole already exists", "Name", cr.GetName())
		}
	}

	return nil
}

func (r *ReconcileIBMBlockCSI) reconcileClusterRoleBinding(instance *ibmblockcsi.IBMBlockCSI) error {
	recLogger := log.WithValues("Resource Type", "ClusterRoleBinding")

	externalProvisioner := instance.GenerateExternalProvisionerClusterRoleBinding()
	externalAttacher := instance.GenerateExternalAttacherClusterRoleBinding()
	clusterDriverRegistrar := instance.GenerateClusterDriverRegistrarClusterRoleBinding()
	externalSnapshotter := instance.GenerateExternalSnapshotterClusterRoleBinding()

	for _, crb := range []*rbacv1.ClusterRoleBinding{
		externalProvisioner,
		externalAttacher,
		clusterDriverRegistrar,
		externalSnapshotter,
	} {
		if err := controllerutil.SetControllerReference(instance.Unwrap(), crb, r.scheme); err != nil {
			return err
		}
		found := &rbacv1.ClusterRoleBinding{}
		err := r.client.Get(context.TODO(), types.NamespacedName{
			Name:      crb.Name,
			Namespace: crb.Namespace,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			recLogger.Info("Creating a new ClusterRoleBinding", "Name", crb.GetName())
			err = r.client.Create(context.TODO(), crb)
			if err != nil {
				return err
			}
		} else if err != nil {
			recLogger.Error(err, "Failed to get ClusterRole", "Name", crb.GetName())
			return err
		} else {
			// Resource already exists - don't requeue
			//recLogger.Info("Skip reconcile: ClusterRoleBinding already exists", "Name", crb.GetName())
		}
	}
	return nil
}
