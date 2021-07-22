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

package controller

import (
	"os"

	"github.com/IBM/ibm-block-csi-operator/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme   			 = runtime.NewScheme()
	setupLog 			 = ctrl.Log.WithName("setup")
	watchNamespaceEnvVar = "WATCH_NAMESPACE"
	topologyPrefixes	 = [...]string{"topology.kubernetes.io", "topology.block.csi.ibm.com"}
	metricsAddr          = ":8080"
	probeAddr            = ":8081"
	enableLeaderElection = false
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "csi.ibm.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	AddToManagerFuncs = append(AddToManagerFuncs, controllers.SetupWithManager)
}
