package ibmblockcsi

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"

	csiv1 "github.com/IBM/ibm-block-csi-driver-operator/pkg/apis/csi/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/IBM/ibm-block-csi-driver-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/util/decoder"
	yamlutil "github.com/IBM/ibm-block-csi-driver-operator/pkg/util/yaml"
)

var log = logf.Log.WithName("controller_ibmblockcsi")

// Add creates a new IBMBlockCSI Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIBMBlockCSI{client: mgr.GetClient(), scheme: mgr.GetScheme()}
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

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner IBMBlockCSI
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &csiv1.IBMBlockCSI{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileIBMBlockCSI implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileIBMBlockCSI{}

// ReconcileIBMBlockCSI reconciles a IBMBlockCSI object
type ReconcileIBMBlockCSI struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
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
	instance := &csiv1.IBMBlockCSI{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	// Define a new Pod object
	resources, err := generateAllResourcesForCR(instance)
	if err != nil {
		// something bad happened, the controller can not recover.
		// you should check if anything is wrong with the yamls.
		panic(err)
	}

	for _, resource := range resources {
		// Set IBMBlockCSI instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, resource, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this resource already exists
		var found runtime.Object
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: resource.Name, Namespace: resource.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new resource", "Namespace", resource.Namespace, "Name", resource.Name)
			err = r.client.Create(context.TODO(), resource)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else if err != nil {
			return reconcile.Result{}, err
		}

		// Resource already exists - don't requeue
		reqLogger.Info("Skip reconcile: resource already exists", "Namespace", found.Namespace, "Name", found.Name)
	}

	// Resource created successfully - don't requeue
	return reconcile.Result{}, nil
}

// generateAllResourcesForCR returns a busybox pod with the same name/namespace as the cr
func generateAllResourcesForCR(cr *csiv1.IBMBlockCSI) ([]*unstructured.Unstructured, error) {
	ns := cr.Namespace
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
			data, err := ioutil.ReadFile()
			if err != nil {
				return nil, err
			}

			data = updatePlaceholders(cr, data)
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

func updatePlaceholders(cr *csiv1.IBMBlockCSI, data []byte) []byte {
	controllerRep := cr.Spec.Controller.Repository
	controllerTag := cr.Spec.Controller.Tag
	controllerUri := controllerRep + ":" + controllerTag

	nodeRep := cr.Spec.Node.Repository
	nodeTag := cr.Spec.Node.Tag
	nodeUri := nodeRep + ":" + nodeTag

	dataString := strings.Replace(string(data), config.ControllerImage, controllerUri, -1)
	dataString = strings.Replace(dataString, config.NodeImage, nodeUri, -1)

	return []byte(dataString)
}
