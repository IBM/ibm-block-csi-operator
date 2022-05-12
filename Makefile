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

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:crdVersions=v1"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen kustomize## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=ibm-block-csi-operator webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	hack/update_labels_in_crd.sh

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

CONTROLLER_GEN = controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0)

KUSTOMIZE = kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# custom
run_unit_tests_image=docker run --rm -v $(CURDIR):/go/src/github.com/IBM/ibm-block-csi-operator -t operator-unittests

.PHONY: olm-validation
olm-validation:
	build/ci/olm_validation.sh

.PHONY: build-unit-tests-image
build-unit-tests-image:
	docker build -f build/ci/Dockerfile.unittest -t operator-unittests .

.PHONY: run-unit-tests
run-unit-tests:
	$(run_unit_tests_image) make test

KUBERNETES_VERSION=1.23.1
.PHONY: test
test: check-generated-manifests update
ifeq (s390x, $(shell hack/get-arch.sh))
	ginkgo -r -v -skipPackage envtest
else
	export KUBEBUILDER_ASSETS=$(shell setup-envtest use -p path ${KUBERNETES_VERSION});\
	ginkgo -r -v
endif

.PHONY: update
update: kustomize
	hack/update-all.sh

.PHONY: check-generated-manifests
check-generated-manifests:
	hack/check-generated-manifests.sh

.PHONY: update-generated-yamls
update-generated-yamls:
	$(run_unit_tests_image) hack/update-config-yamls.sh
	$(run_unit_tests_image) hack/update-installer.sh

.PHONY: list
list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'
