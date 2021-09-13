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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/IBM/ibm-block-csi-operator/controllers/syncer"
	kubeutil "github.com/IBM/ibm-block-csi-operator/pkg/util/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorConfig "github.com/IBM/ibm-block-csi-operator/pkg/config"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	csiv1 "github.com/IBM/ibm-block-csi-operator/api/v1"
	"github.com/IBM/ibm-block-csi-operator/controllers"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	//+kubebuilder:scaffold:imports
)

var (
	scheme               = runtime.NewScheme()
	setupLog             = ctrl.Log.WithName("setup")
	watchNamespaceEnvVar = "WATCH_NAMESPACE"
	topologyPrefixes     = [...]string{"topology.kubernetes.io", "topology.block.csi.ibm.com"}
)

var log = logf.Log.WithName("cmd")

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(csiv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	err := operatorConfig.LoadDefaultsOfIBMBlockCSI()
	if err != nil {
		log.Error(err, "Failed to load default custom resource config")
		os.Exit(1)
	}

	namespace, err := getWatchNamespace()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Port:                   9443,
		Namespace:              namespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.IBMBlockCSIReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Namespace: namespace,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IBMBlockCSI")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	ctx := context.TODO()
	topologyEnabled, err := IsTopologyInUse(ctx)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	syncer.TopologyEnabled = topologyEnabled

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func getWatchNamespace() (string, error) {
	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}

func IsTopologyInUse(ctx context.Context) (bool, error) {
	kubeClient := kubeutil.KubeClient
	nodes, err := kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, node := range nodes.Items {
		for key := range node.Labels {
			for _, prefix := range topologyPrefixes {
				if strings.HasPrefix(key, prefix) {
					return true, nil
				}
			}
		}
	}
	return false, nil
}
