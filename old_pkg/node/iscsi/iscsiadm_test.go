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
