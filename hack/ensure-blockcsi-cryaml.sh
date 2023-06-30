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

CSI_CR_PATH="config/manager/${CSI_CR_FILE}"
CSI_SAMPLES_CR_PATH="config/samples/${CSI_CR_FILE}"

if curl --head --silent --fail "${CSI_GA_CR_URL}" &> /dev/null; then
        echo "Downloading the IBM Block CSI CR file on version ${CSI_RELEASE} ..."
        curl -JL "${CSI_GA_CR_URL}" -o "${CSI_CR_PATH}"
else
        echo "CSI tag doesn't exist yet, downloading develop version"
        CSI_DEVELOP_CR_URL="https://raw.githubusercontent.com/IBM/ibm-block-csi-operator/develop/config/samples/${CSI_CR_FILE}"
        curl -JL "${CSI_DEVELOP_CR_URL}" -o "${CSI_CR_PATH}"
        echo "Overriding CSI CR file to develop registry"
        sed -i "s/quay.io\/ibmcsiblock\/ibm-block-csi-driver-controller/${CSI_DEVELOP_REGISTRY}\/ibm-block-csi-driver-controller-amd64/g" "${CSI_CR_PATH}"
        sed -i "s/quay.io\/ibmcsiblock\/ibm-block-csi-driver-node/${CSI_DEVELOP_REGISTRY}\/ibm-block-csi-driver-node-amd64/g" "${CSI_CR_PATH}"
        sed -i "s/tag: \"${CSI_RELEASE_NUMBER}\"/tag: \"${CSI_LATEST_TAG}\"/g" "${CSI_CR_PATH}"
fi

echo "Coping CR file to all directories"
cp "${CSI_CR_PATH}" "${CSI_SAMPLES_CR_PATH}"
