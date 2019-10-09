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
	"context"
	"time"

	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1
	var ibc *csiv1.IBMBlockCSI
	var ns = "kube-system"
	var ibcName = "test-ibc"

	BeforeEach(func() {
		ibc = &csiv1.IBMBlockCSI{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ibcName,
				Namespace: ns,
			},
			Spec: csiv1.IBMBlockCSISpec{
				Controller: csiv1.IBMBlockCSIControllerSpec{
					Repository: "fake-controller-repo",
					Tag:        "fake-controller-tag",
				},
				Node: csiv1.IBMBlockCSINodeSpec{
					Repository: "fake-node-repo",
					Tag:        "fake-node-tag",
				},
			},
		}

	})

	Describe("test ibc controller", func() {

		Context("create an ibc instance", func() {

			It("should create all the relevant objects", func(done Done) {
				err := k8sClient.Create(context.Background(), ibc)
				Ω(err).ShouldNot(HaveOccurred())

				found := &csiv1.IBMBlockCSI{}
				key := types.NamespacedName{
					Name:      ibcName,
					Namespace: ns,
				}

				By("Getting IBMBlockCSI object after creation")
				Eventually(func() (*csiv1.IBMBlockCSI, error) {
					err := k8sClient.Get(context.Background(), key, found)
					return found, err
				}, timeout, interval).ShouldNot(BeNil())

				By("Getting ServiceAccount")
				sa := &corev1.ServiceAccount{}
				saKey := types.NamespacedName{
					Name:      config.GetNameForResource(config.CSIControllerServiceAccount, found.Name),
					Namespace: found.Namespace,
				}
				Eventually(func() (*corev1.ServiceAccount, error) {
					err := k8sClient.Get(context.Background(), saKey, sa)
					return sa, err
				}, timeout, interval).ShouldNot(BeNil())

				By("Getting controller provisioner ClusterRole")
				cr := &rbacv1.ClusterRole{}
				crKey := types.NamespacedName{
					Name:      config.GetNameForResource(config.ExternalProvisionerClusterRole, found.Name),
					Namespace: "",
				}
				Eventually(func() (*rbacv1.ClusterRole, error) {
					err := k8sClient.Get(context.Background(), crKey, cr)
					return cr, err
				}, timeout, interval).ShouldNot(BeNil())

				By("Getting controller provisioner ClusterRoleBinding")
				crb := &rbacv1.ClusterRoleBinding{}
				crbKey := types.NamespacedName{
					Name:      config.GetNameForResource(config.ExternalProvisionerClusterRoleBinding, found.Name),
					Namespace: "",
				}
				Eventually(func() (*rbacv1.ClusterRoleBinding, error) {
					err := k8sClient.Get(context.Background(), crbKey, crb)
					return crb, err
				}, timeout, interval).ShouldNot(BeNil())

				By("Getting controller StatefulSet")
				controller := &appsv1.StatefulSet{}
				controllerKey := types.NamespacedName{
					Name:      config.GetNameForResource(config.CSIController, found.Name),
					Namespace: found.Namespace,
				}
				Eventually(func() (*appsv1.StatefulSet, error) {
					err := k8sClient.Get(context.Background(), controllerKey, controller)
					return controller, err
				}, timeout, interval).ShouldNot(BeNil())

				// securityContext.privileged: Forbidden: disallowed by cluster policy
				// enable this check after the test cluster support running privileged contianers.
				//				By("Getting node DaemonSet")
				//				node := &appsv1.DaemonSet{}
				//				nodeKey := types.NamespacedName{
				//					Name:      config.GetNameForResource(config.CSINode, found.Name),
				//					Namespace: found.Namespace,
				//				}
				//				Eventually(func() (*appsv1.DaemonSet, error) {
				//					err := k8sClient.Get(context.Background(), nodeKey, node)
				//					return node, err
				//				}, timeout, interval).ShouldNot(BeNil())

				close(done)
			}, timeout.Seconds())
		})
	})
})
