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

package hostdefinition_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	csiv1 "github.com/IBM/ibm-block-csi-operator/api/v1"
	. "github.com/IBM/ibm-block-csi-operator/controllers/internal/hostdefinition"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
)

var _ = Describe("DefaultSetter", func() {
	var hd = &csiv1.HostDefinition{}
	var hdWrapper = New(hd)
	var changed bool

	It("should have a host definition cr yaml configured", func() {
		err := config.LoadDefaultsOfHostDefinition()
		Expect(err).To(BeNil())
	})

	Context("test SetDefaults", func() {

		JustBeforeEach(func() {
			changed = hdWrapper.SetDefaults()
		})

		Context("nothing is set", func() {

			It("should set right host definition defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(hd.Spec.HostDefinition.Repository).To(Equal(config.DefaultHostDefinitionCr.Spec.HostDefinition.Repository))
				Expect(hd.Spec.HostDefinition.Tag).To(Equal(config.DefaultHostDefinitionCr.Spec.HostDefinition.Tag))
			})
		})

		Context("only host definition repository is unofficial", func() {

			BeforeEach(func() {
				hd = &csiv1.HostDefinition{
					Spec: csiv1.HostDefinitionSpec{
						HostDefinition: csiv1.IBMBlockCSIHostDefinitionSpec{
							Repository: "test",
						},
					},
				}
				hdWrapper.HostDefinition = hd
			})

			It("should not set any defaults", func() {
				Expect(changed).To(BeFalse())
				Expect(hd.Spec.HostDefinition.Repository).To(Equal("test"))
				Expect(hd.Spec.HostDefinition.Tag).To(Equal(""))
			})
		})

		Context("only host definition tag is set", func() {

			BeforeEach(func() {
				hd = &csiv1.HostDefinition{
					Spec: csiv1.HostDefinitionSpec{
						HostDefinition: csiv1.IBMBlockCSIHostDefinitionSpec{
							Tag: "test",
						},
					},
				}
				hdWrapper.HostDefinition = hd
			})

			It("should set right defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(hd.Spec.HostDefinition.Repository).To(Equal(config.DefaultHostDefinitionCr.Spec.HostDefinition.Repository))
				Expect(hd.Spec.HostDefinition.Tag).NotTo(Equal("test"))
			})
		})

		Context("everything is set", func() {

			BeforeEach(func() {
				hdWrapper.SetDefaults()
			})

			It("should do nothing", func() {
				Expect(changed).To(BeFalse())
				Expect(hd.Spec.HostDefinition.Repository).To(Equal(config.DefaultHostDefinitionCr.Spec.HostDefinition.Repository))
				Expect(hd.Spec.HostDefinition.Tag).To(Equal(config.DefaultHostDefinitionCr.Spec.HostDefinition.Tag))
			})
		})

	})
})
