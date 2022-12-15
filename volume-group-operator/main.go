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
	"github.com/IBM/volume-group-operator/controllers/utils"
	grpcClient "github.com/IBM/volume-group-operator/pkg/client"
	"github.com/IBM/volume-group-operator/pkg/config"
	"github.com/go-logr/logr"
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

	cfg := config.NewDriverConfig()

	defineFlags(cfg)

	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	err := cfg.Validate()
	exitWithError(err, "error in driver configuration")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Port:   9443,
	})
	exitWithError(err, "unable to start manager")

	log := ctrl.Log.WithName("controllers").WithName("VolumeGroup")
	grpcClientInstance, err := getControllerGrpcClient(cfg, log)
	exitWithError(err, "failed to get controller GRPC client")

	controllerUtils := utils.ControllerUtils{
		Client: mgr.GetClient(),
	}

	err = (&controllers.VolumeGroupReconciler{
		Client:       mgr.GetClient(),
		Utils:        controllerUtils,
		Log:          log,
		Scheme:       mgr.GetScheme(),
		DriverConfig: cfg,
		GRPCClient:   grpcClientInstance,
	}).SetupWithManager(mgr, cfg)
	exitWithError(err, "unable to create controller  with controller VolumeGroup")

	//+kubebuilder:scaffold:builder

	err = mgr.AddHealthzCheck("healthz", healthz.Ping)
	exitWithError(err, "unable to set up health check")

	err = mgr.AddReadyzCheck("readyz", healthz.Ping)
	exitWithError(err, "unable to set up ready check")

	setupLog.Info("starting manager")
	err = mgr.Start(ctrl.SetupSignalHandler())
	exitWithError(err, "problem running manager")

}

func defineFlags(cfg *config.DriverConfig) {
	flag.StringVar(&cfg.DriverName, "driver-name", "", "The CSI driver name.")
	flag.StringVar(&cfg.DriverEndpoint, "csi-address", "/run/csi/socket", "Address of the CSI driver socket.")
	flag.DurationVar(&cfg.RPCTimeout, "rpc-timeout", defaultTimeout, "The timeout for RPCs to the CSI driver.")
}

func getControllerGrpcClient(cfg *config.DriverConfig, log logr.Logger) (*grpcClient.Client, error) {
	grpcClientInstance, err := grpcClient.New(cfg.DriverEndpoint, cfg.RPCTimeout)
	if err != nil {
		log.Error(err, "failed to create GRPC Client", "Endpoint", cfg.DriverEndpoint, "GRPC Timeout", cfg.RPCTimeout)

		return nil, err
	}
	err = grpcClientInstance.Probe()
	if err != nil {
		log.Error(err, "failed to connect to driver", "Endpoint", cfg.DriverEndpoint, "GRPC Timeout", cfg.RPCTimeout)

		return nil, err
	}
	return grpcClientInstance, err
}

func exitWithError(err error, msg string) {
	if err != nil {
		setupLog.Error(err, msg)
		os.Exit(1)
	}
}
