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

package ibmblockcsi_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	. "github.com/IBM/ibm-block-csi-operator/pkg/internal/ibmblockcsi"
)

var _ = Describe("DefaultSetter", func() {
	var ibc *csiv1.IBMBlockCSI = &csiv1.IBMBlockCSI{}
	var ibcWrapper = New(ibc, "1.13")
	var changed bool

	Context("test SetDefaults", func() {

		JustBeforeEach(func() {
			changed = ibcWrapper.SetDefaults()
		})

		Context("nothing is set", func() {

			It("should set right defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(ibc.Spec.Controller.Repository).To(Equal(config.ControllerRepository))
				Expect(ibc.Spec.Controller.Tag).To(Equal(config.ControllerTag))
				Expect(ibc.Spec.Node.Repository).To(Equal(config.NodeRepository))
				Expect(ibc.Spec.Node.Tag).To(Equal(config.NodeTag))
			})
		})

		Context("only controller repository is set", func() {

			BeforeEach(func() {
				ibc = &csiv1.IBMBlockCSI{
					Spec: csiv1.IBMBlockCSISpec{
						Controller: csiv1.IBMBlockCSIControllerSpec{
							Repository: "test",
						},
					}}
				ibcWrapper.IBMBlockCSI = ibc
			})

			It("should set right defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(ibc.Spec.Controller.Repository).To(Equal("test"))
				Expect(ibc.Spec.Controller.Tag).To(Equal(""))
				Expect(ibc.Spec.Node.Repository).To(Equal(config.NodeRepository))
				Expect(ibc.Spec.Node.Tag).To(Equal(config.NodeTag))
			})
		})

		Context("only controller tag is set", func() {

			BeforeEach(func() {
				ibc = &csiv1.IBMBlockCSI{
					Spec: csiv1.IBMBlockCSISpec{
						Controller: csiv1.IBMBlockCSIControllerSpec{
							Tag: "test",
						},
					}}
				ibcWrapper.IBMBlockCSI = ibc
			})

			It("should set right defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(ibc.Spec.Controller.Repository).To(Equal(config.ControllerRepository))
				Expect(ibc.Spec.Controller.Tag).NotTo(Equal("test"))
				Expect(ibc.Spec.Node.Repository).To(Equal(config.NodeRepository))
				Expect(ibc.Spec.Node.Tag).To(Equal(config.NodeTag))
			})
		})

		Context("only node repository is set", func() {

			BeforeEach(func() {
				ibc = &csiv1.IBMBlockCSI{
					Spec: csiv1.IBMBlockCSISpec{
						Node: csiv1.IBMBlockCSINodeSpec{
							Repository: "test",
						},
					}}
				ibcWrapper.IBMBlockCSI = ibc
			})

			It("should set right defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(ibc.Spec.Controller.Repository).To(Equal(config.ControllerRepository))
				Expect(ibc.Spec.Controller.Tag).To(Equal(config.ControllerTag))
				Expect(ibc.Spec.Node.Repository).To(Equal("test"))
				Expect(ibc.Spec.Node.Tag).To(Equal(""))
			})
		})

		Context("only node tag is set", func() {

			BeforeEach(func() {
				ibc = &csiv1.IBMBlockCSI{
					Spec: csiv1.IBMBlockCSISpec{
						Node: csiv1.IBMBlockCSINodeSpec{
							Tag: "test",
						},
					}}
				ibcWrapper.IBMBlockCSI = ibc
			})

			It("should set right defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(ibc.Spec.Controller.Repository).To(Equal(config.ControllerRepository))
				Expect(ibc.Spec.Controller.Tag).To(Equal(config.ControllerTag))
				Expect(ibc.Spec.Node.Repository).To(Equal(config.NodeRepository))
				Expect(ibc.Spec.Node.Tag).NotTo(Equal("test"))
			})
		})

		Context("everything is set", func() {

			BeforeEach(func() {
				ibcWrapper.SetDefaults()
			})

			It("should do nothing", func() {
				Expect(changed).To(BeFalse())
				Expect(ibc.Spec.Controller.Repository).To(Equal(config.ControllerRepository))
				Expect(ibc.Spec.Controller.Tag).To(Equal(config.ControllerTag))
				Expect(ibc.Spec.Node.Repository).To(Equal(config.NodeRepository))
				Expect(ibc.Spec.Node.Tag).NotTo(Equal("test"))
			})
		})

	})
})
