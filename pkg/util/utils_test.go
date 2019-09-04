package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/IBM/ibm-block-csi-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

type fakeStruct struct{}

func (s *fakeStruct) FakeFunc(ok bool) bool {
	return !ok
}

var _ = Describe("Utils", func() {
	Context("test Invoke", func() {

		It("should return true if false is passed", func() {
			f := &fakeStruct{}
			v, err := Invoke(f, "FakeFunc", false)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(v).To(HaveLen(1))
			Expect(v[0].Interface().(bool)).To(BeTrue())
		})

		It("should return true if false is passed", func() {
			f := &fakeStruct{}
			v, err := Invoke(f, "FakeFunc", true)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(v).To(HaveLen(1))
			Expect(v[0].Interface().(bool)).To(BeFalse())
		})

		It("should return error if method is not found", func() {
			f := &fakeStruct{}
			_, err := Invoke(f, "fakeFunc", false)
			Expect(err.Error()).To(Equal("reflect: call of reflect.Value.Type on zero Value"))
		})
	})

	Context("test GetNodeAddresses", func() {
		var node *corev1.Node

		BeforeEach(func() {
			node = &corev1.Node{Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{
						Type:    corev1.NodeHostName,
						Address: "hostname",
					},
					{
						Type:    corev1.NodeExternalIP,
						Address: "external",
					},
					{
						Type:    corev1.NodeInternalIP,
						Address: "internal",
					},
				},
			}}
		})

		It("should return addresses in right order", func() {
			addr := GetNodeAddresses(node)
			Expect(addr).To(Equal([]string{"internal", "external", "hostname"}))
		})

		It("should return empty if no addresses", func() {
			addr := GetNodeAddresses(&corev1.Node{})
			Expect(addr).To(HaveLen(0))
		})

	})
})
