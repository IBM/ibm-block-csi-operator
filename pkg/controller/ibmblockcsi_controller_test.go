package controller_test

import (
	"context"
	"time"

	csiv1 "github.com/IBM/ibm-block-csi-driver-operator/pkg/apis/csi/v1"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
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
				Eventually(func() (*csiv1.IBMBlockCSI, error) {
					err := k8sClient.Get(context.Background(), key, found)
					return found, err
				}, timeout, interval).ShouldNot(BeNil())

				controller := &appsv1.StatefulSet{}
				controllerKey := types.NamespacedName{
					Name:      config.GetNameForResource(config.CSIController, found.Name),
					Namespace: found.Namespace,
				}
				Eventually(func() (*appsv1.StatefulSet, error) {
					err := k8sClient.Get(context.Background(), controllerKey, controller)
					return controller, err
				}, timeout, interval).ShouldNot(BeNil())

				close(done)
			}, timeout.Seconds())
		})
	})
})
