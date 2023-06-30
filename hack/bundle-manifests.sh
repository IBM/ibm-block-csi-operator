#!/bin/bash
#
# Copyright contributors to the ibm-block-csi-operator project
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


set -e

source hack/common.sh
source hack/ensure-blockcsi-cryaml.sh

BUNDLE_METADATA_OPTS="${BUNDLE_CHANNELS} ${BUNDLE_DEFAULT_CHANNEL}"

# Always start fresh and remove any previous artifacts that may exist
echo "Cleaning the previous artifacts that may exist..."
rm -rf "$(dirname ${BUNDLE_METADATA_DIR})"
mkdir -p "${BUNDLE_METADATA_DIR}"

#BRKD - no need dependencies for csi
## Generate the file dependencies.yaml, which requires the minimum version of IBM Block CSI Operator.
#echo "Generating the file dependencies.yaml..."
#cat << EOF > ${BUNDLE_METADATA_DIR}/dependencies.yaml
#dependencies:
#  - type: olm.package
#    value:
#      packageName: ibm-block-csi-operator
#      version: ">=${CSI_RELEASE_NUMBER}"
#EOF

echo "Generating bundle manifests and metadata... ${OPERATOR_SDK_BIN} generate kustomize manifests -q "
${OPERATOR_SDK_BIN} generate kustomize manifests -q

pushd config/manager
if [ "$LOCAL_OS_TYPE" == "Darwin" ] && [[ "$(sed --version 2>&1 | head -n 1 | awk -F " " '{print $2}')" == "illegal" ]]; then
        sed -i "" "s#value: .*#value: ${FLASHSYSTEM_DRIVER_FULL_IMAGE_NAME}#" ../default/manager_config_patch.yaml
else
        sed -i "s#value: .*#value: ${FLASHSYSTEM_DRIVER_FULL_IMAGE_NAME}#" ../default/manager_config_patch.yaml
fi

../../${KUSTOMIZE_BIN} edit set image controller="${OPERATOR_FULL_IMAGE_NAME}"
popd


pushd config/console

../../${KUSTOMIZE_BIN} edit set image ibm-console="${CONSOLE_FULL_IMAGE_NAME}"
popd


${KUSTOMIZE_BIN} build config/manifests | ${OPERATOR_SDK_BIN} generate bundle -q --overwrite --version "${RELEASE_VERSION}" ${BUNDLE_METADATA_OPTS}

echo "Validating the generated files..."
${OPERATOR_SDK_BIN} bundle validate ./bundle

#echo "Updating the olm.skipRange for new release version ${RELEASE_VERSION}..."
#${YQ_BIN} eval -i ".metadata.annotations.\"olm.skipRange\" = \">=0.0.1 <${RELEASE_VERSION}\"" "${CSV_PATH}"

#echo "add certification required labels to bundle.Dockerfile"
#echo "LABEL com.redhat.openshift.versions=\"v4.13\"" >> ./bundle.Dockerfile
#echo "LABEL com.redhat.delivery.operator.bundle=true" >> ./bundle.Dockerfile
