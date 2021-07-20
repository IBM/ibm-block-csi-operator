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

package controller_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/IBM/ibm-block-csi-operator/api"
	"github.com/IBM/ibm-block-csi-operator/controllers/controller"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/util/boolptr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var cfg *rest.Config
var k8sClient client.Client
var k8sManager ctrl.Manager
var testEnv *envtest.Environment
var clientset *kubernetes.Clientset
var kubeVersion = "1.14"
var nodeAgentPort = "10086"
var storageAgentPort = "10010"
var storageAgentAddress = "localhost:" + storageAgentPort

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter))

	os.Setenv(config.ENVKubeVersion, kubeVersion)
	os.Setenv(config.ENVIscsiAgentPort, nodeAgentPort)
	os.Setenv(config.ENVEndpoint, storageAgentAddress)

	By("bootstrapping test environment")
	if os.Getenv("TEST_USE_EXISTING_CLUSTER") == "true" {
		testEnv = &envtest.Environment{
			UseExistingCluster: boolptr.True(),
		}
	} else {
		testEnv = &envtest.Environment{
			CRDDirectoryPaths:  []string{filepath.Join("..", "..", "deploy", "crds")},
			KubeAPIServerFlags: append([]string{"--allow-privileged=true"}, envtest.DefaultKubeAPIServerFlags...),
		}
	}

	var err error

	cfg, err = testEnv.Start()
	Ω(err).ShouldNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	clientset, err = kubernetes.NewForConfig(cfg)
	Ω(err).ShouldNot(HaveOccurred())

	// Create a new Cmd to provide shared dependencies and start components
	k8sManager, err = manager.New(cfg, manager.Options{
		Namespace:      "",
		MapperProvider: apiutil.NewDiscoveryRESTMapper,
	})
	Ω(err).ShouldNot(HaveOccurred())

	// Setup Scheme for all resources
	err = api.AddToScheme(k8sManager.GetScheme())
	Ω(err).ShouldNot(HaveOccurred())

	// Setup all Controllers
	err = controller.AddToManager(k8sManager)
	Ω(err).ShouldNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err := k8sManager.Start(signals.SetupSignalHandler())
		Ω(err).ShouldNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	os.Setenv(config.ENVKubeVersion, "")
	os.Setenv(config.ENVIscsiAgentPort, "")
	os.Setenv(config.ENVEndpoint, "")

	gexec.KillAndWait(5 * time.Second)
	err := testEnv.Stop()
	Ω(err).ShouldNot(HaveOccurred())
})
