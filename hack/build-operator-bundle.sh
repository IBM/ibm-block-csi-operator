#!/bin/bash
#
# Copyright contributors to the ibm-storage-odf-operator project
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
source hack/ensure-opm.sh

build_push_bundle_image(){
  bundle_image_name=${1}
  directory_path=${2}
  package_operator_name=${3}
  operator_channels=${4}
  echo "Building Operator bundle image ${bundle_image_name}..."
  ${OPM_BIN} alpha bundle build --directory "${directory_path}" --tag "${bundle_image_name}" --output-dir . --package "${package_operator_name}" --channels "${operator_channels}"
  echo "Pushing Operator bundle image to image registry..."
  docker push "${bundle_image_name}"
  echo
}

clone_csi_repo(){
  echo "CSI tag doesn't exist yet, cloning CSI GitHub repository"
  oldPWD=$(pwd)
  if [ ! -d "${CSI_OPERATOR_IMAGE_NAME}" ]
  then
    git clone "${CSI_GIT_PATH}"
    cd "${CSI_OPERATOR_IMAGE_NAME}"
  else
    cd "${CSI_OPERATOR_IMAGE_NAME}"
    git pull
  fi
}

override_csi_csv_file(){
  echo "Overriding CSV file to develop registry"
  sed -i "s/registry.connect.redhat.com\/ibm\/ibm-block-csi-operator:${CSI_RELEASE_NUMBER}/${CSI_DEVELOP_REGISTRY}\/ibm-block-csi-operator-amd64:${CSI_LATEST_TAG}/g" "${CSI_CSV_PATH}/${CSI_CSV_FILE}"
  sed -i "s/quay.io\/ibmcsiblock\/ibm-block-csi-driver-controller/${CSI_DEVELOP_REGISTRY}\/ibm-block-csi-driver-controller-amd64/g" "${CSI_CSV_PATH}/${CSI_CSV_FILE}"
  sed -i "s/quay.io\/ibmcsiblock\/ibm-block-csi-driver-node/${CSI_DEVELOP_REGISTRY}\/ibm-block-csi-driver-node-amd64/g" "${CSI_CSV_PATH}/${CSI_CSV_FILE}"
  sed -i "s/quay.io\/ibmcsiblock\/ibm-block-csi-host-definer/${CSI_DEVELOP_REGISTRY}\/ibm-block-csi-host-definer-amd64/g" "${CSI_CSV_PATH}/${CSI_CSV_FILE}"
  sed -i "s/\"tag\": \"${CSI_RELEASE_NUMBER}\"/\"tag\": \"${CSI_LATEST_TAG}\"/g" "${CSI_CSV_PATH}/${CSI_CSV_FILE}"
  sed -i "s/\"repository\": \"quay.io\/ibmcsiblock\/csi-volume-group-operator\"/\"repository\": \"${CSI_VOLUME_GROUP_OPERATOR_DEVELOP_PATH}\"/g" "${CSI_CSV_PATH}/${CSI_CSV_FILE}"
  sed -i "s/\"tag\": \"${CSI_VOLUME_GROUP_OPERATOR_TAG}\"/\"tag\": \"${CSI_LATEST_TAG}\"/g" "${CSI_CSV_PATH}/${CSI_CSV_FILE}"
}

check_and_build_csi_bundle_image(){
  if curl --head --silent --fail "${CSI_GA_CR_URL}" &> /dev/null; then
    echo "CSI release is GAed. Using official images"
  else
    clone_csi_repo
    override_csi_csv_file
    build_push_bundle_image "${IMAGE_REGISTRY}/${CSI_DEVELOP_BUNDLE_FULL_IMAGE_NAME}:${IMAGE_TAG}" "${CSI_CSV_PATH}" "${CSI_OPERATOR_IMAGE_NAME}" "${CSI_CHANNEL}"
    echo

    echo "Deleting CSI repository clone"
    cd "${oldPWD}"
    rm -rf "${CSI_OPERATOR_IMAGE_NAME}"
  fi
}


build_push_bundle_image "${BUNDLE_FULL_IMAGE_NAME}" "bundle/metadata/" "${OPERATOR_IMAGE_NAME}" "${CHANNELS}"
check_and_build_csi_bundle_image
if [ "${ENABLE_UPGRADE}" == "True" ]; then
  echo "Upgrade is enabled, pulling previous bundle image"
  docker pull "${PREVIOUS_BUNDLE_IMAGE_PATH}"
fi
