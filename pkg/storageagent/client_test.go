package storageagent

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/IBM/ibm-block-csi-driver-operator/pkg/config"
)

var _ = Describe("Client", func() {

	BeforeEach(func() {
		os.Setenv(config.ENVEndpoint, "test")

	})

	AfterEach(func() {
		os.Setenv(config.ENVEndpoint, "")
	})

	Context("test beautify", func() {

		// rules:
		//     1. The name can contain letters, numbers, spaces, periods, dashes, and underscores.
		//     2. The name must begin with a letter or an underscore.
		//     3. The name must not begin or end with a space.

		It("should not change anything if name is valid", func() {
			before := "a valid name 1 -._"
			after := before
			Expect(beautify(before)).To(Equal(after))
		})

		It("should add _ if name is an ip", func() {
			before := "1.2.3.4"
			after := "_1.2.3.4"
			Expect(beautify(before)).To(Equal(after))
		})

		It("should add _ if name starts with - or .", func() {
			before := "-"
			after := "_-"
			Expect(beautify(before)).To(Equal(after))

			before = "."
			after = "_."
			Expect(beautify(before)).To(Equal(after))
		})

		It("should trim the spaces", func() {
			before := "  name  "
			after := "name"
			Expect(beautify(before)).To(Equal(after))
		})

		It("should replace invalid letter with _", func() {
			before := "a!a@a#a$a%a^a&a*a(a)))))a"
			after := "a_a_a_a_a_a_a_a_a_a_____a"
			Expect(beautify(before)).To(Equal(after))
		})
	})
})
