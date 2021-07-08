#
# Copyright 2019 IBM Corp.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

#all: test

.PHONY: build
build:
	CGO_ENABLED=1 GOOS=linux go build     -o build/_output/bin/ibm-block-csi-operator     -gcflags all=-trimpath=${GOPATH} -asmflags all=-trimpath=${GOPATH} -mod=vendor cmd/manager/main.go

.PHONY: olm-validation
olm-validation:
	build/ci/olm_validation.sh

.PHONY: test
test: update
	# for go 1.13+, set GOFLAGS to enable vendor mod for ginkgo
	GO111MODULE=on GOFLAGS='-mod=vendor' ginkgo -r -skipPackage pkg/controller

.PHONY: update
update:
	hack/update-all.sh

.PHONY: list
list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'
	