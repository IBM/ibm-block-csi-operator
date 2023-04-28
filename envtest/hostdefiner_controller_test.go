/**
 * Copyright 2022 IBM Corp.
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

package envtest

import (
	"context"
	"fmt"
	"time"

	csiv1 "github.com/IBM/ibm-block-csi-operator/api/v1"
	testsutil "github.com/IBM/ibm-block-csi-operator/controllers/util/tests"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Controller", func() {

	err := config.LoadDefaultsOfHostDefiner()
	Expect(err).To(BeNil(), fmt.Sprint("can't load defaults of Host Definition"))

	const timeout = time.Second * 30
	const interval = time.Second * 1
	var hostDefiner *csiv1.HostDefiner
	var namespace = config.DefaultHostDefinerCr.ObjectMeta.Namespace
	var containersImages = testsutil.GetHostDefinerImagesByName(config.DefaultHostDefinerCr)
	var hdName = config.DefaultHostDefinerCr.ObjectMeta.Name
	var clusterRoles = []config.ResourceName{config.HostDefinerClusterRole}
	var clusterRoleBindings = []config.ResourceName{config.HostDefinerClusterRoleBinding}

	BeforeEach(func() {
		hostDefiner = &config.DefaultHostDefinerCr
	})

	Describe("test host definition controller", func() {

		Context("create an host definition instance", func() {

			It("should create all the relevant objects", func(done Done) {
				err := k8sClient.Create(context.Background(), hostDefiner)
				Expect(err).NotTo(HaveOccurred())

				found := &csiv1.HostDefiner{}
				key := types.NamespacedName{
					Name:      hdName,
					Namespace: namespace,
				}

				By("Getting HostDefiner object after creation")
				Eventually(func() (*csiv1.HostDefiner, error) {
					err := k8sClient.Get(context.Background(), key, found)
					return found, err
				}, timeout, interval).ShouldNot(BeNil())

				By("Getting HostDefiner ServiceAccount")
				sa := &corev1.ServiceAccount{}
				Eventually(func() (*corev1.ServiceAccount, error) {
					err := k8sClient.Get(context.Background(),
						testsutil.GetResourceKey(config.HostDefinerServiceAccount, found.Name, found.Namespace), sa)
					return sa, err
				}, timeout, interval).ShouldNot(BeNil())

				By("Getting HostDefiner ClusterRole")
				cr := &rbacv1.ClusterRole{}
				for _, clusterRole := range clusterRoles {
					Eventually(func() (*rbacv1.ClusterRole, error) {
						err := k8sClient.Get(context.Background(),
							testsutil.GetResourceKey(clusterRole, found.Name, ""), cr)
						return cr, err
					}, timeout, interval).ShouldNot(BeNil())
				}

				By("Getting HostDefiner ClusterRoleBinding")
				crb := &rbacv1.ClusterRoleBinding{}
				for _, clusterRoleBinding := range clusterRoleBindings {
					Eventually(func() (*rbacv1.ClusterRoleBinding, error) {
						err := k8sClient.Get(context.Background(),
							testsutil.GetResourceKey(clusterRoleBinding, found.Name, ""), crb)
						return crb, err
					}, timeout, interval).ShouldNot(BeNil())
				}

				By("Getting HostDefiner deployment")
				deployment := &appsv1.Deployment{}
				Eventually(func() (*appsv1.Deployment, error) {
					err := k8sClient.Get(context.Background(),
						testsutil.GetResourceKey(config.HostDefiner, found.Name, found.Namespace), deployment)
					return deployment, err
				}, timeout, interval).ShouldNot(BeNil())
				assertDeployedContainersAreInCR(deployment.Spec.Template.Spec, containersImages)

				By("Checking if all HostDefiner containers were deployed")
				var containersNameInControllerAndNode []string
				containersNameInControllerAndNode = addContainersNameInPod(deployment.Spec.Template.Spec, containersNameInControllerAndNode)
				assertContainersInCRAreDeployed(containersNameInControllerAndNode, containersImages)

				close(done)
			}, timeout.Seconds())
		})
	})
})
