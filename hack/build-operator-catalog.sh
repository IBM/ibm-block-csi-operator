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
source hack/ensure-opm.sh

add_bundle_image_to_existing_catalog() {
  operator_package_name=${1}
  channel=${2}
  bundle_quay_io_image=${3}
  operator_package_name_version=${4}
  initialize_operator_catalog "${operator_package_name}" "${channel}"
  add_bundle_to_catalog "${operator_package_name}" "${bundle_quay_io_image}"
  add_channel_entry_for_bundle "${operator_package_name}" "${channel}" "${operator_package_name_version}"

  ${OPM_BIN} validate "${operator_package_name}"
}

initialize_operator_catalog() {
  operator_package_name=${1}
  channel=${2}
  if [ ! -d "${operator_package_name}" ]; then
    echo "Initializing operator ${operator_package_name} catalog"
    mkdir "${operator_package_name}" || exit
    ${OPM_BIN} init "${operator_package_name}" --default-channel="${channel}" --output yaml >> "${operator_package_name}"/index.yaml
  fi
}

add_bundle_to_catalog() {
  operator_package_name=${1}
  bundle_image_name=${2}
  echo "Add bundle image ${bundle_image_name} into catalog"
  ${OPM_BIN} render "${bundle_image_name}" --output=yaml >> "${operator_package_name}"/index.yaml
}

add_channel_entry_for_bundle() {
  operator_package_name=${1}
  channel=${2}
  operator_package_name_version=${3}
  echo "Adding operator ${operator_package_name} channel entry into catalog"
  cat << EOF >> "${operator_package_name}"/index.yaml
---
schema: olm.channel
package: ${operator_package_name}
name: ${channel}
entries:
- name: ${operator_package_name_version}
EOF
}

build_push_catalog_image() {
  catalog_package_name=${1}
  catalog_quay_io_image=${2}
  echo "Building and pushing catalog ${catalog_package_name}"
  docker build -f "${catalog_package_name}".Dockerfile -t "${catalog_quay_io_image}" .
  docker push "${catalog_quay_io_image}"
  echo
}

init_parent_catalog() {
  catalog_name=${1}
  mkdir "${catalog_name}" || exit
  echo "Generating parent catalog Dockerfile"
  ${OPM_BIN} alpha generate dockerfile "${catalog_name}"
  cd "${catalog_name}"
  echo
}

check_and_add_csi_bundle_to_catalog() {
  if curl --head --silent --fail "${CSI_GA_CR_URL}" &> /dev/null; then
    echo "CSI release is GAed. Using official images"
  else
    echo "CSI tag doesn't exist yet, adding CSI bundle into CSI catalog."
    add_bundle_image_to_existing_catalog "${CSI_OPERATOR_IMAGE_NAME}" "${CSI_CHANNEL}" "${IMAGE_REGISTRY}/${CSI_DEVELOP_BUNDLE_FULL_IMAGE_NAME}:${IMAGE_TAG}" "${CSI_OPERATOR_IMAGE_NAME}.${CSI_RELEASE}"
    echo
  fi
}

#check_and_add_previous_odf_bundle_to_catalog(){
#  if [ "${ENABLE_UPGRADE}" == "True" ]; then
#    echo "Adding previous ODF release bundle to catalog"
#    add_bundle_image_to_existing_catalog "${OPERATOR_IMAGE_NAME}" "${PREVIOUS_CHANNELS}" "${PREVIOUS_BUNDLE_IMAGE_PATH}" "${PREVIOUS_OPERATOR_IMAGE_NAME_VERSION}"
#  fi
#}


init_parent_catalog "${CATALOG_IMAGE_NAME}"
#check_and_add_previous_odf_bundle_to_catalog
echo
echo "Adding current CSI bundle image into catalog"
add_bundle_image_to_existing_catalog "${OPERATOR_IMAGE_NAME}" "${CHANNELS}" "${BUNDLE_FULL_IMAGE_NAME}" "${OPERATOR_IMAGE_NAME_VERSION}"
echo
#check_and_add_csi_bundle_to_catalog
#echo

#cd -
#build_push_catalog_image "${CATALOG_IMAGE_NAME}" "${CATALOG_FULL_IMAGE_NAME}"
#
#echo "Cleaning leftovers"
#rm -rf "${CATALOG_IMAGE_NAME}".Dockerfile "${CATALOG_IMAGE_NAME}"
