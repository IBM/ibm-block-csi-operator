package iscsi

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var fakeOut string
var fakeRecord = "1.2.3.4:3260,1 iqn.xxx"

type IscsiCmdFunc func(args ...string) (string, error)

var _ = Describe("Iscsiadm", func() {
	var iscsi = NewIscsiAdmin()
	var realIscsiCmd IscsiCmdFunc

	var fakeIscsiCmd = func(args ...string) (string, error) {
		return fakeOut, nil
	}

	Context("test Discover", func() {

		BeforeEach(func() {
			realIscsiCmd = iscsiCmd
			iscsiCmd = fakeIscsiCmd
			fakeOut = fakeRecord
		})

		AfterEach(func() {
			fakeOut = ""
			iscsiCmd = realIscsiCmd
		})

		It("should return the right targets", func() {
			targets, err := iscsi.Discover("test")
			Ω(err).ShouldNot(HaveOccurred())
			Expect(targets).To(HaveLen(1))
			Expect(targets[0].Iqn).To(Equal("iqn.xxx"))
			Expect(targets[0].Portal).To(Equal("1.2.3.4"))
			Expect(targets[0].Port).To(Equal("3260"))
		})
	})
})
