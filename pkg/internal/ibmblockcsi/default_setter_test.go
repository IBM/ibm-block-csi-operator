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
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	. "github.com/IBM/ibm-block-csi-operator/pkg/internal/ibmblockcsi"
)

var _ = Describe("DefaultSetter", func() {
	var ibc *csiv1.IBMBlockCSI = &csiv1.IBMBlockCSI{}
	var ibcWrapper = New(ibc, "1.14")
	var changed bool

	Describe("test Kubernetes SetDefaults", func() {

		JustBeforeEach(func() {
			changed = ibcWrapper.SetDefaults(config.Kubernetes)
		})

		Context("nothing is set", func() {

			BeforeEach(func() {
				ibc = &csiv1.IBMBlockCSI{}
				ibcWrapper.IBMBlockCSI = ibc
			})

			It("should set right defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(ibc.Spec.Controller.Repository).To(Equal(config.ControllerRepository))
				Expect(ibc.Spec.Controller.Tag).To(Equal(config.ControllerTag))
				Expect(ibc.Spec.Node.Repository).To(Equal(config.NodeRepository))
				Expect(ibc.Spec.Node.Tag).To(Equal(config.NodeTag))
				Expect(ibc.Spec.Sidecars).To(HaveLen(4))
				Expect(ibc.Spec.ImagePullSecrets).NotTo(BeNil())
				Expect(ibc.Spec.ImagePullSecrets).To(HaveLen(0))
				Expect(ibc.Spec.Controller.Tolerations).NotTo(BeNil())
				Expect(ibc.Spec.Controller.Tolerations).To(HaveLen(0))
				Expect(ibc.Spec.Node.Tolerations).NotTo(BeNil())
				Expect(ibc.Spec.Node.Tolerations).To(HaveLen(0))
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
				ibcWrapper.SetDefaults(config.Kubernetes)
			})

			It("should do nothing", func() {
				Expect(changed).To(BeFalse())
				Expect(ibc.Spec.Controller.Repository).To(Equal(config.ControllerRepository))
				Expect(ibc.Spec.Controller.Tag).To(Equal(config.ControllerTag))
				Expect(ibc.Spec.Node.Repository).To(Equal(config.NodeRepository))
				Expect(ibc.Spec.Node.Tag).NotTo(Equal("test"))
				Expect(ibc.Spec.Sidecars).To(HaveLen(4))
			})
		})

		Context("controller and node version is 1.0.0", func() {

			BeforeEach(func() {
				ibc = &csiv1.IBMBlockCSI{
					Spec: csiv1.IBMBlockCSISpec{
						Controller: csiv1.IBMBlockCSIControllerSpec{
							Repository: config.ControllerRepository,
							Tag:        "1.0.0",
						},
						Node: csiv1.IBMBlockCSINodeSpec{
							Repository: config.NodeRepository,
							Tag:        "1.0.0",
						},
					}}
				ibcWrapper.IBMBlockCSI = ibc
			})

			It("should be updated to new version", func() {
				Expect(changed).To(BeTrue())
				Expect(ibc.Spec.Controller.Repository).To(Equal(config.ControllerRepository))
				Expect(ibc.Spec.Controller.Tag).To(Equal(config.ControllerTag))
				Expect(ibc.Spec.Node.Repository).To(Equal(config.NodeRepository))
				Expect(ibc.Spec.Node.Tag).To(Equal(config.NodeTag))
			})
		})

		Context("csi provisioner version is v1.3.0", func() {

			BeforeEach(func() {
				ibc = &csiv1.IBMBlockCSI{
					Spec: csiv1.IBMBlockCSISpec{
						Sidecars: []csiv1.CSISidecar{
							{
								Name:       config.CSIProvisioner,
								Repository: "quay.io/k8scsi/csi-provisioner",
								Tag:        "v1.3.0",
							},
						},
					}}
				ibcWrapper.IBMBlockCSI = ibc
			})

			It("should be updated to new version", func() {
				Expect(changed).To(BeTrue())
				Expect(ibc.Spec.Sidecars).To(HaveLen(4))

				for _, sidecar := range ibc.Spec.Sidecars {
					if sidecar.Name == config.CSIProvisioner {
						tag := strings.Split(config.CSIProvisionerImage, ":")[1]
						Expect(sidecar.Tag).To(Equal(tag))
					}
				}
			})
		})
	})

	Describe("test OpenShift SetDefaults", func() {

		JustBeforeEach(func() {
			changed = ibcWrapper.SetDefaults(config.OpenShift)
		})

		Context("nothing is set", func() {

			BeforeEach(func() {
				ibc = &csiv1.IBMBlockCSI{}
				ibcWrapper.IBMBlockCSI = ibc
			})

			It("should set right defaults", func() {
				Expect(changed).To(BeTrue())
				Expect(ibc.Spec.Controller.Repository).To(Equal(config.OpenShiftControllerRepository))
				Expect(ibc.Spec.Controller.Tag).To(Equal(config.ControllerTag))
				Expect(ibc.Spec.Node.Repository).To(Equal(config.OpenShiftNodeRepository))
				Expect(ibc.Spec.Node.Tag).To(Equal(config.NodeTag))
				Expect(ibc.Spec.Sidecars).To(HaveLen(4))
			})
		})

		Context("everything is set", func() {

			BeforeEach(func() {
				ibcWrapper.SetDefaults(config.OpenShift)
			})

			It("should do nothing", func() {
				Expect(changed).To(BeFalse())
			})
		})

		Context("controller and node version is 1.0.0", func() {

			BeforeEach(func() {
				ibc = &csiv1.IBMBlockCSI{
					Spec: csiv1.IBMBlockCSISpec{
						Controller: csiv1.IBMBlockCSIControllerSpec{
							Repository: config.ControllerRepository,
							Tag:        "1.0.0",
						},
						Node: csiv1.IBMBlockCSINodeSpec{
							Repository: config.NodeRepository,
							Tag:        "1.0.0",
						},
					}}
				ibcWrapper.IBMBlockCSI = ibc
			})

			It("should be updated to new version", func() {
				Expect(changed).To(BeTrue())
				Expect(ibc.Spec.Controller.Repository).To(Equal(config.OpenShiftControllerRepository))
				Expect(ibc.Spec.Controller.Tag).To(Equal(config.ControllerTag))
				Expect(ibc.Spec.Node.Repository).To(Equal(config.OpenShiftNodeRepository))
				Expect(ibc.Spec.Node.Tag).To(Equal(config.NodeTag))
			})
		})

		Context("csi provisioner version is v1.3.0", func() {

			BeforeEach(func() {
				ibc = &csiv1.IBMBlockCSI{
					Spec: csiv1.IBMBlockCSISpec{
						Sidecars: []csiv1.CSISidecar{
							{
								Name:       config.CSIProvisioner,
								Repository: "quay.io/k8scsi/csi-provisioner",
								Tag:        "v1.3.0",
							},
							{
								Name:       config.CSIAttacher,
								Repository: "quay.io/k8scsi/csi-attacher",
								Tag:        "v1.2.1",
							},
							{
								Name:       config.CSINodeDriverRegistrar,
								Repository: "quay.io/k8scsi/csi-node-driver-registrar",
								Tag:        "v1.2.0",
							},
							{
								Name:       config.LivenessProbe,
								Repository: "quay.io/k8scsi/livenessprobe",
								Tag:        "v1.1.0",
							},
						},
					}}
				ibcWrapper.IBMBlockCSI = ibc
			})

			It("should be updated to new version", func() {
				Expect(changed).To(BeTrue())
				Expect(ibc.Spec.Sidecars).To(HaveLen(4))

				for _, sidecar := range ibc.Spec.Sidecars {
					repAndTag := strings.Split(ibcWrapper.GetOpenShiftDefaultImageByName(sidecar.Name), ":")
					Expect(sidecar.Repository).To(Equal(repAndTag[0]))
					Expect(sidecar.Tag).To(Equal(repAndTag[1]))
				}
			})
		})

	})
})
