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

package main

import (
	"flag"
	"github.com/IBM/volume-group-operator/pkg/config"
	"os"
	"time"

	uberzap "go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	csiv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/IBM/volume-group-operator/controllers"
	//+kubebuilder:scaffold:imports
)

const (
	// defaultTimeout is default timeout for RPC call.
	defaultTimeout = time.Minute
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(csiv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	opts := zap.Options{
		ZapOpts: []uberzap.Option{
			uberzap.AddCaller(),
		},
	}

	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	cfg := config.NewDriverConfig()

	flag.StringVar(&cfg.DriverName, "driver-name", "", "The CSI driver name.")
	flag.StringVar(&cfg.DriverEndpoint, "csi-address", "/run/csi/socket", "Address of the CSI driver socket.")
	flag.DurationVar(&cfg.RPCTimeout, "rpc-timeout", defaultTimeout, "The timeout for RPCs to the CSI driver.")

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	err := cfg.Validate()
	if err != nil {
		setupLog.Error(err, "error in driver configuration")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Port:   9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.VolumeGroupReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("VolumeGroup"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr, cfg); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "VolumeGroup")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
