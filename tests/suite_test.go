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

package controllers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/IBM/ibm-block-csi-operator/controllers"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	csiv1 "github.com/IBM/ibm-block-csi-operator/api/v1"
	//+kubebuilder:scaffold:imports
)

//const controllerName = "ibmblockcsi-controller"

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

var kubeVersion = "1.18"
var nodeAgentPort = "10086"
var storageAgentPort = "10010"
var storageAgentAddress = "localhost:" + storageAgentPort

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	Expect(os.Setenv(config.ENVKubeVersion, kubeVersion)).To(Succeed())
	Expect(os.Setenv(config.ENVIscsiAgentPort, nodeAgentPort)).To(Succeed())
	Expect(os.Setenv(config.ENVEndpoint, storageAgentAddress)).To(Succeed())

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:        []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing:    true,
		AttachControlPlaneOutput: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = csiv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.IBMBlockCSIReconciler{
		Client:    k8sClient,
		Scheme:    scheme.Scheme,
		Namespace: "default",
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = mgr.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
	Expect(os.Unsetenv(config.ENVKubeVersion)).To(Succeed())
	Expect(os.Unsetenv(config.ENVIscsiAgentPort)).To(Succeed())
	Expect(os.Unsetenv(config.ENVEndpoint)).To(Succeed())
})