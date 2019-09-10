package controller_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/IBM/ibm-block-csi-operator/pkg/config"

	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	fakenode "github.com/IBM/ibm-block-csi-operator/pkg/node/fakeserver"
	pb "github.com/IBM/ibm-block-csi-operator/pkg/node/nodeagent"
	fakestorage "github.com/IBM/ibm-block-csi-operator/pkg/storageagent/fakeserver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const serverSideTimeoutSeconds = 10

var _ = Describe("Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1
	var node *corev1.Node
	var oc *csiv1.Config
	var pv *corev1.PersistentVolume
	var secret *corev1.Secret
	var va *storagev1.VolumeAttachment
	var count uint64 = 0
	var iqn = "iqn.xxx"
	var ns *corev1.Namespace
	var nsName = "test-node"
	var pvName = "test-pv"
	var secretName = "test-secret"
	var configName = "test-config"
	var arrayIP = "8.8.8.8"

	BeforeEach(func() {
		atomic.AddUint64(&count, 1)
		node = &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("node-%v", count)},
			Spec:       corev1.NodeSpec{},
		}

		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: nsName},
			Spec:       corev1.NamespaceSpec{},
		}
		err := k8sClient.Create(context.Background(), ns)
		Ω(err).ShouldNot(HaveOccurred())

		// start a fake node server.
		go func() {
			defer GinkgoRecover()
			err := fakenode.Serve(fmt.Sprintf("localhost:%s", nodeAgentPort), node.GetName())
			Ω(err).ShouldNot(HaveOccurred())
		}()

		// start a fake storage server.
		go func() {
			defer GinkgoRecover()
			err := fakestorage.Serve(storageAgentAddress)
			Ω(err).ShouldNot(HaveOccurred())
		}()

	})

	AfterEach(func(done Done) {
		// Cleanup
		var zero int64 = 0
		policy := metav1.DeletePropagationForeground
		delOptions := &metav1.DeleteOptions{
			GracePeriodSeconds: &zero,
			PropagationPolicy:  &policy,
		}
		_, err := clientset.CoreV1().Nodes().Get(node.Name, metav1.GetOptions{})
		if err == nil {
			err := clientset.CoreV1().Nodes().Delete(node.Name, delOptions)
			Ω(err).ShouldNot(HaveOccurred())
		}

		_, err = clientset.CoreV1().Namespaces().Get(ns.Name, metav1.GetOptions{})
		if err == nil {
			err := clientset.CoreV1().Namespaces().Delete(ns.Name, delOptions)
			Ω(err).ShouldNot(HaveOccurred())
		}

		fakenode.Stop()
		fakenode.ClearAll()

		fakestorage.Stop()

		close(done)
	}, serverSideTimeoutSeconds)

	Describe("test node and volumeattachment controller", func() {

		BeforeEach(func() {
			req := &pb.GetNodeInfoRequest{Name: node.GetName()}
			res := &pb.GetNodeInfoReply{Node: &pb.Node{Name: node.GetName(), Iqns: []string{iqn}}}
			fakenode.StoreResponse("GetNodeInfo", req, res)

			oc = &csiv1.Config{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configName,
					Namespace: ns.Name,
				},
				Spec: csiv1.ConfigSpec{
					DefineHost: true,
				}}

			va = &storagev1.VolumeAttachment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("va-%v", count),
					Namespace: ns.Name,
				},
				Spec: storagev1.VolumeAttachmentSpec{
					Attacher: config.DriverName,
					NodeName: node.GetName(),
					Source: storagev1.VolumeAttachmentSource{
						PersistentVolumeName: &pvName,
					},
				},
			}

			pv = &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: pvName,
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							ControllerPublishSecretRef: &corev1.SecretReference{
								Name:      secretName,
								Namespace: ns.Name,
							},
							Driver:       config.DriverName,
							VolumeHandle: "vh",
						},
					},
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Capacity:    corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("7Gi")},
				},
			}

			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: ns.Name,
				},
				Data: map[string][]byte{
					"management_address": []byte(arrayIP),
					"username":           []byte("username"),
					"password":           []byte("password"),
				},
			}
		})

		// put all the steps in one case so that the servers(API derver, fake servers)
		// start and stop only once.
		Context("create a nodeInfo and update the status", func() {

			It("should create a new nodeInfo and set iqn in status", func(done Done) {

				By("Create a Config")
				err := k8sClient.Create(context.Background(), oc)
				Ω(err).ShouldNot(HaveOccurred())

				foundC := &csiv1.Config{}
				keyC := types.NamespacedName{
					Name:      configName,
					Namespace: ns.Name,
				}
				Eventually(func() (*csiv1.Config, error) {
					err := k8sClient.Get(context.Background(), keyC, foundC)
					return foundC, err
				}, timeout, interval).ShouldNot(BeNil())

				// By("Check the node agent DaemonSet")
				// securityContext.privileged: Forbidden: disallowed by cluster policy
				// enable this check after the test cluster support running privileged contianers.
				//				nodeAgent := &appsv1.DaemonSet{}
				//				nodeAgentKey := types.NamespacedName{
				//					Name:      config.GetNameForResource(config.NodeAgent, found.Name),
				//					Namespace: found.Namespace,
				//				}
				//				Eventually(func() (*appsv1.DaemonSet, error) {
				//					err := k8sClient.Get(context.Background(), nodeAgentKey, nodeAgent)
				//					return nodeAgent, err
				//				}, timeout, interval).ShouldNot(BeNil())

				By("Update the Config status")
				// set the status of nodeAgent.phase to Running.
				foundC.Status.NodeAgent.Phase = csiv1.NodeAgentPhaseRunning
				err = k8sClient.Status().Update(context.Background(), foundC)
				Ω(err).ShouldNot(HaveOccurred())

				newC := &csiv1.Config{}
				Eventually(func() (bool, error) {
					err := k8sClient.Get(context.Background(), keyC, newC)
					return newC.Status.NodeAgent.Phase == csiv1.NodeAgentPhaseRunning, err
				}, timeout, interval).Should(BeTrue())

				By("Create a Node")
				err = k8sClient.Create(context.Background(), node)
				Ω(err).ShouldNot(HaveOccurred())

				found := &corev1.Node{}
				key := types.NamespacedName{
					Name:      node.GetName(),
					Namespace: "",
				}
				Eventually(func() (*corev1.Node, error) {
					err := k8sClient.Get(context.Background(), key, found)
					return found, err
				}, timeout, interval).ShouldNot(BeNil())

				By("Update the Node's addresses")
				found.Status.Addresses = []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "localhost"}}
				err = k8sClient.Status().Update(context.Background(), found)
				Ω(err).ShouldNot(HaveOccurred())

				Eventually(func() (string, error) {
					newFound := &corev1.Node{}
					err := k8sClient.Get(context.Background(), key, newFound)
					if err != nil {
						return "", err
					}
					addresses := newFound.Status.Addresses
					if len(addresses) == 0 {
						return "", fmt.Errorf("address is empty")
					}
					return addresses[0].Address, err
				}, timeout, interval).Should(Equal("localhost"))

				By("Check the status of NodeInfo")
				Eventually(func() ([]string, error) {
					nodeInfo := &csiv1.NodeInfo{}
					err := k8sClient.Get(context.Background(), key, nodeInfo)
					if err != nil {
						return nil, err
					}
					return nodeInfo.Status.Iqns, err
				}, timeout, interval).Should(Equal([]string{iqn}))

				By("Create a PV")
				err = k8sClient.Create(context.Background(), pv)
				Ω(err).ShouldNot(HaveOccurred())

				foundPv := &corev1.PersistentVolume{}
				keyPv := types.NamespacedName{
					Name:      pvName,
					Namespace: "",
				}
				Eventually(func() (*corev1.PersistentVolume, error) {
					err := k8sClient.Get(context.Background(), keyPv, foundPv)
					return foundPv, err
				}, timeout, interval).ShouldNot(BeNil())

				By("Create a Secret")
				err = k8sClient.Create(context.Background(), secret)
				Ω(err).ShouldNot(HaveOccurred())

				foundSecret := &corev1.Secret{}
				keySecret := types.NamespacedName{
					Name:      secretName,
					Namespace: ns.Name,
				}
				Eventually(func() (*corev1.Secret, error) {
					err := k8sClient.Get(context.Background(), keySecret, foundSecret)
					return foundSecret, err
				}, timeout, interval).ShouldNot(BeNil())

				By("Create a VolumeAttachment")
				err = k8sClient.Create(context.Background(), va)
				Ω(err).ShouldNot(HaveOccurred())

				By("Check the status of NodeInfo again")
				Eventually(func() ([]string, error) {
					nodeInfo := &csiv1.NodeInfo{}
					err := k8sClient.Get(context.Background(), key, nodeInfo)
					if err != nil {
						return nil, err
					}
					return nodeInfo.Status.DefinedOnStorages, err
				}, timeout, interval).Should(Equal([]string{arrayIP}))

				close(done)
			}, timeout.Seconds())
		})
	})
})
