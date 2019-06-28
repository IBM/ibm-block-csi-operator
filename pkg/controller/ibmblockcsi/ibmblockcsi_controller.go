package ibmblockcsi

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	csiv1 "github.com/IBM/ibm-block-csi-driver-operator/pkg/apis/csi/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/util/decoder"
	yamlutil "github.com/IBM/ibm-block-csi-driver-operator/pkg/util/yaml"
	"github.com/presslabs/controller-util/syncer"
)

// ReconcileTime is the delay between reconciliations
const ReconcileTime = 30 * time.Second

var log = logf.Log.WithName("controller_ibmblockcsi")

// Add creates a new IBMBlockCSI Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIBMBlockCSI{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetRecorder("controller_ibmblockcsi"),
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
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
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
	instance := ibmblockcsi.New(&csiv1.IBMBlockCSI{})
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
			sErr := r.client.Status().Update(context.TODO(), instance.Unwrap())
			if sErr != nil {
				reqLogger.Error(sErr, "failed to update IBMBlockCSI status", "name", instance.Name)
			}
		}
	}()

	// Define a new Pod object
	resources, err := generateAllResourcesForCR(instance.Unwrap())
	if err != nil {
		// something bad happened, the controller can not recover.
		// you should check if anything is wrong with the yamls.
		panic(err)
	}

	for _, resource := range resources {
		// Set IBMBlockCSI instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance.Unwrap(), resource, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this resource already exists
		var found runtime.Object
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: resource.GetName(), Namespace: resource.GetNamespace()}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new resource", "Namespace", resource.GetNamespace(), "Name", resource.GetName())
			err = r.client.Create(context.TODO(), resource)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else if err != nil {
			return reconcile.Result{}, err
		}

		// Resource already exists - don't requeue
		reqLogger.Info("Skip reconcile: resource already exists", "Namespace", resource.GetNamespace(), "Name", resource.GetName())
	}

	csiControllerSyncer := clustersyncer.NewCSIControllerSyncer(r.client, r.scheme, instance)
	if err = syncer.Sync(context.TODO(), csiControllerSyncer, r.recorder); err != nil {
		return reconcile.Result{}, err
	}

	// Resource created successfully - don't requeue
	return reconcile.Result{}, nil
}

// Generate all the resources required by CSI driver from source /deploy/ibm-block-csi-driver.yaml
func generateAllResourcesForCR(cr *csiv1.IBMBlockCSI) ([]*unstructured.Unstructured, error) {
	//ns := cr.Namespace
	ns := config.DefaultNamespace
	rootDir := config.DeployPath
	files, err := ioutil.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	objList := []*unstructured.Unstructured{}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		fileName := f.Name()
		if strings.HasSuffix(fileName, ".yaml") {
			fullPath := filepath.Join(rootDir, fileName)
			data, err := ioutil.ReadFile(fullPath)
			if err != nil {
				return nil, err
			}
			manifest, err := yamlutil.Split(data)
			if err != nil {
				return nil, err
			}

			for _, resource := range manifest {
				obj, err := decoder.FromYamlToUnstructured(resource)
				if err != nil {
					return nil, err
				}
				obj.SetNamespace(ns)
				objList = append(objList, obj)
			}
		}
	}

	return objList, nil
}
