package informer

import (
	"io/ioutil"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Informer", func() {
	var realIscsiFullPath, fakeIscsiPath, fakeFcPath, realFcPath string
	var fakeIqn = "iqn.xxx"
	var fakeWwpn = "wwpn"
	var testInformer = NewInformer()

	Context("test GetNodeIscsiIQNs", func() {

		BeforeEach(func() {
			realIscsiFullPath = iscsiFullPath

			fakeIscsiPath, _ = ioutil.TempDir("", "")
			fakeIscsiFullPath := path.Join(fakeIscsiPath, iscsiFile)
			iqn := "iqn=" + fakeIqn
			ioutil.WriteFile(fakeIscsiFullPath, []byte(iqn), 0644)

			iscsiFullPath = fakeIscsiFullPath

		})

		AfterEach(func() {
			iscsiFullPath = realIscsiFullPath
			os.RemoveAll(fakeIscsiPath)
		})

		It("should return the right iqn", func() {
			iqns, err := testInformer.GetNodeIscsiIQNs()
			Ω(err).ShouldNot(HaveOccurred())
			Expect(iqns).To(HaveLen(1))
			Expect(iqns[0]).To(Equal(fakeIqn))
		})
	})

	Context("test GetNodeFcWWPNs", func() {

		BeforeEach(func() {
			realFcPath = fcPath

			fakeFcPath, _ = ioutil.TempDir("", "")
			hostName := "host1"
			fakeHostPath := path.Join(fakeFcPath, hostName)
			fakePortNamePath := path.Join(fakeHostPath, portName)
			fakePortStatePath := path.Join(fakeHostPath, portState)
			os.Mkdir(fakeHostPath, os.ModePerm)

			err := ioutil.WriteFile(fakePortStatePath, []byte(portOnline), 0644)
			Ω(err).ShouldNot(HaveOccurred())
			err = ioutil.WriteFile(fakePortNamePath, []byte(fakeWwpn), 0644)
			Ω(err).ShouldNot(HaveOccurred())
			fcPath = fakeFcPath

		})

		AfterEach(func() {
			fcPath = realFcPath
			os.RemoveAll(fakeFcPath)
		})

		Context("there is only one online port", func() {

			It("should return the right wwpn", func() {
				wwpns, err := testInformer.GetNodeFcWWPNs()
				Ω(err).ShouldNot(HaveOccurred())
				Expect(wwpns).To(HaveLen(1))
				Expect(wwpns[0]).To(Equal(fakeWwpn))
			})
		})

		Context("there are two online ports", func() {

			BeforeEach(func() {
				hostName := "host2"
				fakeHostPath := path.Join(fakeFcPath, hostName)
				fakePortNamePath := path.Join(fakeHostPath, portName)
				fakePortStatePath := path.Join(fakeHostPath, portState)
				os.Mkdir(fakeHostPath, os.ModePerm)

				err := ioutil.WriteFile(fakePortStatePath, []byte(portOnline), 0644)
				Ω(err).ShouldNot(HaveOccurred())
				err = ioutil.WriteFile(fakePortNamePath, []byte(fakeWwpn), 0644)
				Ω(err).ShouldNot(HaveOccurred())
				fcPath = fakeFcPath

			})

			It("should return two right wwpn", func() {
				wwpns, err := testInformer.GetNodeFcWWPNs()
				Ω(err).ShouldNot(HaveOccurred())
				Expect(wwpns).To(HaveLen(2))
				Expect(wwpns[0]).To(Equal(fakeWwpn))
				Expect(wwpns[1]).To(Equal(fakeWwpn))
			})
		})

		Context("there are one online port and one offline port", func() {

			BeforeEach(func() {
				hostName := "host3"
				fakeHostPath := path.Join(fakeFcPath, hostName)
				fakePortNamePath := path.Join(fakeHostPath, portName)
				fakePortStatePath := path.Join(fakeHostPath, portState)
				os.Mkdir(fakeHostPath, os.ModePerm)

				err := ioutil.WriteFile(fakePortStatePath, []byte("offline"), 0644)
				Ω(err).ShouldNot(HaveOccurred())
				err = ioutil.WriteFile(fakePortNamePath, []byte(fakeWwpn), 0644)
				Ω(err).ShouldNot(HaveOccurred())
				fcPath = fakeFcPath

			})

			It("should return one right wwpn", func() {
				wwpns, err := testInformer.GetNodeFcWWPNs()
				Ω(err).ShouldNot(HaveOccurred())
				Expect(wwpns).To(HaveLen(1))
				Expect(wwpns[0]).To(Equal(fakeWwpn))
			})
		})

	})
})
