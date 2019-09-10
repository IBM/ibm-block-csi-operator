package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	. "github.com/IBM/ibm-block-csi-operator/pkg/internal/config"
)

var _ = Describe("DefaultSetter", func() {
	var c *csiv1.Config = &csiv1.Config{}
	var cWrapper = New(c)
	var changed bool

	Context("test SetDefaults", func() {

		JustBeforeEach(func() {
			changed = cWrapper.SetDefaults()
		})

		Context("nothing is set", func() {

			It("should set right defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(c.Spec.NodeAgent.Repository).To(Equal(config.NodeAgentRepository))
				Expect(c.Spec.NodeAgent.Tag).To(Equal(config.NodeAgentTag))
				Expect(c.Spec.NodeAgent.Port).To(Equal(config.NodeAgentPort))
				Expect(c.Spec.DefineHost).To(BeFalse())
			})
		})

		Context("only node agent repository is set", func() {

			var repo string

			BeforeEach(func() {
				repo = "test_repo"
				c = &csiv1.Config{
					Spec: csiv1.ConfigSpec{
						NodeAgent: csiv1.NodeAgentSpec{
							Repository: repo,
						},
					}}
				cWrapper.Config = c
			})

			It("should set right defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(c.Spec.NodeAgent.Repository).To(Equal(repo))
				Expect(c.Spec.NodeAgent.Tag).To(Equal(""))
				Expect(c.Spec.NodeAgent.Port).To(Equal(config.NodeAgentPort))
			})
		})

		Context("only node agent tag is set", func() {

			var tag string

			BeforeEach(func() {
				tag = "test_tag"
				c = &csiv1.Config{
					Spec: csiv1.ConfigSpec{
						NodeAgent: csiv1.NodeAgentSpec{
							Tag: tag,
						},
					}}
				cWrapper.Config = c
			})

			It("should set right defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(c.Spec.NodeAgent.Repository).To(Equal(config.NodeAgentRepository))
				Expect(c.Spec.NodeAgent.Tag).To(Equal(config.NodeAgentTag))
				Expect(c.Spec.NodeAgent.Port).To(Equal(config.NodeAgentPort))
			})
		})

		Context("only port is set", func() {
			var port string

			BeforeEach(func() {
				port = "test_port"
				c = &csiv1.Config{
					Spec: csiv1.ConfigSpec{
						NodeAgent: csiv1.NodeAgentSpec{
							Port: port,
						},
					}}
				cWrapper.Config = c
			})

			It("should set right defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(c.Spec.NodeAgent.Repository).To(Equal(config.NodeAgentRepository))
				Expect(c.Spec.NodeAgent.Tag).To(Equal(config.NodeAgentTag))
				Expect(c.Spec.NodeAgent.Port).To(Equal(port))
			})
		})

		Context("everything is set", func() {

			BeforeEach(func() {
				cWrapper.SetDefaults()
			})

			It("should do nothing", func() {
				Expect(changed).To(BeFalse())
			})
		})

	})
})
