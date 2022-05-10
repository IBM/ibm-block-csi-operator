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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Controller", func() {

	err := config.LoadDefaultsOfIBMBlockCSI()
	Expect(err).To(BeNil(), fmt.Sprint("can't load defaults of IBM Block CSI"))

	const timeout = time.Second * 30
	const interval = time.Second * 1
	var ibc *csiv1.IBMBlockCSI
	var namespace = config.DefaultCr.ObjectMeta.Namespace
	var containersImages = testsutil.GetImagesByName(config.DefaultCr, config.DefaultSidecarsByName)
	var ibcName = config.DefaultCr.ObjectMeta.Name
	var clusterRoles = []config.ResourceName{config.ExternalProvisionerClusterRole, config.ExternalAttacherClusterRole,
		config.ExternalSnapshotterClusterRole, config.ExternalResizerClusterRole, config.CSIAddonsReplicatorClusterRole,
		config.CSIControllerSCCClusterRole, config.CSINodeSCCClusterRole}
	var clusterRoleBindings = []config.ResourceName{config.ExternalProvisionerClusterRoleBinding,
		config.ExternalAttacherClusterRoleBinding, config.ExternalSnapshotterClusterRoleBinding,
		config.ExternalResizerClusterRoleBinding, config.CSIAddonsReplicatorClusterRoleBinding,
		config.CSIControllerSCCClusterRoleBinding, config.CSINodeSCCClusterRoleBinding}

	BeforeEach(func() {
		ibc = &config.DefaultCr
	})

	Describe("test ibc controller", func() {

		Context("create an ibc instance", func() {

			It("should create all the relevant objects", func(done Done) {
				err := k8sClient.Create(context.Background(), ibc)
				Expect(err).NotTo(HaveOccurred())

				found := &csiv1.IBMBlockCSI{}
				key := types.NamespacedName{
					Name:      ibcName,
					Namespace: namespace,
				}

				By("Getting IBMBlockCSI object after creation")
				Eventually(func() (*csiv1.IBMBlockCSI, error) {
					err := k8sClient.Get(context.Background(), key, found)
					return found, err
				}, timeout, interval).ShouldNot(BeNil())

				By("Getting CSIDriver")
				cd := &storagev1.CSIDriver{}
				Eventually(func() (*storagev1.CSIDriver, error) {
					err := k8sClient.Get(context.Background(),
						testsutil.GetResourceKey(config.DriverName, "", ""), cd)
					return cd, err
				}, timeout, interval).ShouldNot(BeNil())

				By("Getting IBMBlockCSI ServiceAccount")
				sa := &corev1.ServiceAccount{}
				Eventually(func() (*corev1.ServiceAccount, error) {
					err := k8sClient.Get(context.Background(),
						testsutil.GetResourceKey(config.CSIControllerServiceAccount, found.Name, found.Namespace), sa)
					return sa, err
				}, timeout, interval).ShouldNot(BeNil())

				By("Getting IBMBlockCSI ClusterRole")
				cr := &rbacv1.ClusterRole{}
				for _, clusterRole := range clusterRoles {
					Eventually(func() (*rbacv1.ClusterRole, error) {
						err := k8sClient.Get(context.Background(),
							testsutil.GetResourceKey(clusterRole, found.Name, ""), cr)
						return cr, err
					}, timeout, interval).ShouldNot(BeNil())
				}

				By("Getting IBMBlockCSI ClusterRoleBinding")
				crb := &rbacv1.ClusterRoleBinding{}
				for _, clusterRoleBinding := range clusterRoleBindings {
					Eventually(func() (*rbacv1.ClusterRoleBinding, error) {
						err := k8sClient.Get(context.Background(),
							testsutil.GetResourceKey(clusterRoleBinding, found.Name, ""), crb)
						return crb, err
					}, timeout, interval).ShouldNot(BeNil())
				}

				By("Getting controller StatefulSet")
				controller := &appsv1.StatefulSet{}
				Eventually(func() (*appsv1.StatefulSet, error) {
					err := k8sClient.Get(context.Background(),
						testsutil.GetResourceKey(config.CSIController, found.Name, found.Namespace), controller)
					return controller, err
				}, timeout, interval).ShouldNot(BeNil())
				assertDeployedContainersAreInCR(controller.Spec.Template.Spec, containersImages)

				By("Getting node DaemonSet")
				node := &appsv1.DaemonSet{}
				Eventually(func() (*appsv1.DaemonSet, error) {
					err := k8sClient.Get(context.Background(),
						testsutil.GetResourceKey(config.CSINode, found.Name, found.Namespace), node)
					return node, err
				}, timeout, interval).ShouldNot(BeNil())
				assertDeployedContainersAreInCR(node.Spec.Template.Spec, containersImages)

				By("Checking if all IBMBlockCSI containers were deployed")
				var containersNameInControllerAndNode []string
				containersNameInControllerAndNode = addContainersNameInPod(node.Spec.Template.Spec, containersNameInControllerAndNode)
				containersNameInControllerAndNode = addContainersNameInPod(controller.Spec.Template.Spec, containersNameInControllerAndNode)
				assertContainersInCRAreDeployed(containersNameInControllerAndNode, containersImages)

				close(done)
			}, timeout.Seconds())
		})
	})
})

func assertDeployedContainersAreInCR(deployedPodSpec corev1.PodSpec, containersImagesInCR map[string]string) {
	Expect(deployedPodSpec.Containers).To(Not(BeEmpty()))
	for _, deployedContainer := range deployedPodSpec.Containers {
		image, ok := containersImagesInCR[deployedContainer.Name]
		Expect(ok).To(BeTrue(), fmt.Sprintf("container %s not found in %s", deployedContainer.Name, containersImagesInCR))
		Expect(image).To(Equal(deployedContainer.Image))
	}
}

func addContainersNameInPod(deployedPodSpec corev1.PodSpec, deployedContainersNames []string) []string {
	for _, deployedContainer := range deployedPodSpec.Containers {
		deployedContainersNames = append(deployedContainersNames, deployedContainer.Name)
	}
	return deployedContainersNames
}

func assertContainersInCRAreDeployed(deployedContainersNames []string, containersImagesInCR map[string]string) {
	for deployedContainerName, _ := range containersImagesInCR {
		Expect(isContainerDeployed(deployedContainersNames, deployedContainerName)).To(BeTrue(),
			fmt.Sprintf("container %s not found in CSI deployment", deployedContainerName))
	}
}

func isContainerDeployed(deployedContainersNames []string, wantedContainerName string) bool {
	for _, containerName := range deployedContainersNames {
		if containerName == wantedContainerName {
			return true
		}
	}
	return false
}
