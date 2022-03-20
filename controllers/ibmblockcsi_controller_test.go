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

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	csiv1 "github.com/IBM/ibm-block-csi-operator/api/v1"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)
  
var _ = Describe("Controller", func() {
  
	const timeout = time.Second * 30
	const interval = time.Second * 1
	var ibc *csiv1.IBMBlockCSI
	var ns = "default"
	var ibcName = "test-ibc"
	var apiVersion = "csi.ibm.com/v1"
	var kind = "IBMBlockCSI"
	containersImages := map[string]string{
		"ibm-block-csi-controller":  "fake-controller-repo:fake-controller-tag",
		"ibm-block-csi-node": 		 "fake-node-repo:fake-node-tag",
		"csi-provisioner": 			 "fake-provisioner-repo:fake-provisioner-tag",
		"csi-attacher": 			 "fake-attacher-repo:fake-attacher-tag",
		"csi-snapshotter": 			 "fake-snapshotter-repo:fake-snapshotter-tag",
		"csi-resizer": 				 "fake-resizer-repo:fake-resizer-tag",
		"csi-addons-replicator": 	 "fake-replicator-repo:fake-replicator-tag",
		"livenessprobe": 			 "fake-livenessprobe-repo:fake-livenessprobe-tag",
		"csi-node-driver-registrar": "fake-registrar-repo:fake-registrar-tag",
	}

	BeforeEach(func() {
		ibc = &csiv1.IBMBlockCSI{
			TypeMeta: metav1.TypeMeta{
				Kind:       kind,
				APIVersion: apiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      ibcName,
				Namespace: ns,
			},
			Spec: csiv1.IBMBlockCSISpec{
				Controller: csiv1.IBMBlockCSIControllerSpec{
					Repository: strings.Split(containersImages["ibm-block-csi-controller"], ":")[0],
					Tag:        strings.Split(containersImages["ibm-block-csi-controller"], ":")[1],
				},
				Node: csiv1.IBMBlockCSINodeSpec{
					Repository: strings.Split(containersImages["ibm-block-csi-node"], ":")[0],
					Tag:        strings.Split(containersImages["ibm-block-csi-node"], ":")[1],
				},
				Sidecars: []csiv1.CSISidecar{
				{
					Name:            "csi-node-driver-registrar",
					Repository:      strings.Split(containersImages["csi-node-driver-registrar"], ":")[0],
					Tag:             strings.Split(containersImages["csi-node-driver-registrar"], ":")[1],
					ImagePullPolicy: "IfNotPresent",
				},
				{
					Name:            "csi-provisioner",
					Repository:      strings.Split(containersImages["csi-provisioner"], ":")[0],
					Tag:             strings.Split(containersImages["csi-provisioner"], ":")[1],
					ImagePullPolicy: "IfNotPresent",
				},
				{
					Name:            "csi-attacher",
					Repository:      strings.Split(containersImages["csi-attacher"], ":")[0],
					Tag:             strings.Split(containersImages["csi-attacher"], ":")[1],
					ImagePullPolicy: "IfNotPresent",
				},
				{
				  	Name:            "csi-snapshotter",
				  	Repository:      strings.Split(containersImages["csi-snapshotter"], ":")[0],
				  	Tag:             strings.Split(containersImages["csi-snapshotter"], ":")[1],
				  	ImagePullPolicy: "IfNotPresent",
				},
				{
				  	Name:            "csi-resizer",
				  	Repository:      strings.Split(containersImages["csi-resizer"], ":")[0],
				  	Tag:             strings.Split(containersImages["csi-resizer"], ":")[1],
				  	ImagePullPolicy: "IfNotPresent",
				},
				{
					Name:            "csi-addons-replicator",
					Repository:      strings.Split(containersImages["csi-addons-replicator"], ":")[0],
					Tag:             strings.Split(containersImages["csi-addons-replicator"], ":")[1],
					ImagePullPolicy: "IfNotPresent",
				},
				{
					Name:            "livenessprobe",
					Repository:      strings.Split(containersImages["livenessprobe"], ":")[0],
					Tag:             strings.Split(containersImages["livenessprobe"], ":")[1],
					ImagePullPolicy: "IfNotPresent",
				},
			    },
			},
		}
	  })
  
	Describe("test ibc controller", func() {

		Context("create an ibc instance", func() {

			It("should create all the relevant objects", func(done Done) {
				err := k8sClient.Create(context.Background(), ibc)
				Expect(err).NotTo(HaveOccurred())

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

				By("Getting CSIDriver")
				cd := &storagev1.CSIDriver{}
				cdKey := types.NamespacedName{
				  Name:      config.DriverName,
				  Namespace: "",
				}
				Eventually(func() (*storagev1.CSIDriver, error) {
				  err := k8sClient.Get(context.Background(), cdKey, cd)
				  return cd, err
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

				By("Getting ClusterRole")
				cr := &rbacv1.ClusterRole{}
				clusterRoles := []config.ResourceName{config.ExternalProvisionerClusterRole, config.ExternalAttacherClusterRole, 
				config.ExternalSnapshotterClusterRole, config.ExternalResizerClusterRole, config.CSIAddonsReplicatorClusterRole,
				config.CSIControllerSCCClusterRole, config.CSINodeSCCClusterRole}

				for _, clusterRole := range clusterRoles {
					crKey := types.NamespacedName{
						Name:      config.GetNameForResource(clusterRole, found.Name),
						Namespace: "",
					}
					Eventually(func() (*rbacv1.ClusterRole, error) {
						err := k8sClient.Get(context.Background(), crKey, cr)
						return cr, err
					}, timeout, interval).ShouldNot(BeNil())
				}

				By("Getting ClusterRoleBinding")
				crb := &rbacv1.ClusterRoleBinding{}
				clusterRoleBindings := []config.ResourceName{config.ExternalProvisionerClusterRoleBinding,
					config.ExternalAttacherClusterRoleBinding, config.ExternalSnapshotterClusterRoleBinding,
					config.ExternalResizerClusterRoleBinding, config.CSIAddonsReplicatorClusterRoleBinding,
					config.CSIControllerSCCClusterRoleBinding, config.CSINodeSCCClusterRoleBinding}

				for _, clusterRoleBinding := range clusterRoleBindings {
					crbKey := types.NamespacedName{
						Name:      config.GetNameForResource(clusterRoleBinding, found.Name),
						Namespace: "",
					}
					Eventually(func() (*rbacv1.ClusterRoleBinding, error) {
						err := k8sClient.Get(context.Background(), crbKey, crb)
						return crb, err
					  }, timeout, interval).ShouldNot(BeNil())
				}

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
				Expect(len(controller.Spec.Template.Spec.Containers)).To(BeNumerically(">", 0))
				for _, container := range controller.Spec.Template.Spec.Containers {
					image, ok := containersImages[container.Name]
					Expect(ok).To(BeTrue(), fmt.Sprintf("container %s not found in %s", container.Name, containersImages))
					Expect(container.Image).To(Equal(image))
				}


				By("Getting node DaemonSet")
				node := &appsv1.DaemonSet{}
				nodeKey := types.NamespacedName{
				  Name:      config.GetNameForResource(config.CSINode, found.Name),
				  Namespace: found.Namespace,
				}
				Eventually(func() (*appsv1.DaemonSet, error) {
				  err := k8sClient.Get(context.Background(), nodeKey, node)
				  return node, err
				}, timeout, interval).ShouldNot(BeNil())
				Expect(len(node.Spec.Template.Spec.Containers)).To(BeNumerically(">", 0))
				for _, container := range node.Spec.Template.Spec.Containers {
					image, ok := containersImages[container.Name]
					Expect(ok).To(BeTrue(), fmt.Sprintf("container %s not found in %s", container.Name, containersImages))
					Expect(container.Image).To(Equal(image))
				}

				close(done)
			  }, timeout.Seconds())
		  })
	  })
  })
