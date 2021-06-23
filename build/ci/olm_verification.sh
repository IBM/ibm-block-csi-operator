#!/bin/bash -xe

#
# Copyright 2020 IBM Corp.
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

ARCH=$(uname -m)

NON_BUNDLE_FORMAT_VERSIONS_FOR_CERTIFIED=()
NON_BUNDLE_FORMAT_VERSIONS_FOR_COMMUNITY=("1.0.0" "1.1.0" "1.2.0" "1.3.0" "1.4.0" "1.5.0")

verify(){
  olm_bundles_dir="$1*/"
  non_bundle_format_versions=$2[@]
  non_bundle_format_versions=("${!non_bundle_format_versions}")
  for olm_bundle_dir in ${olm_bundles_dir} ; do
    version=$(basename "${olm_bundle_dir}")
    exclude_version=$(echo "${non_bundle_format_versions[@]}" | grep -o "${version}" | wc -w)
    if [ "${exclude_version}" -eq 0 ]; then
      echo "Validating ${olm_bundle_dir}"
      operator-sdk bundle --verbose validate "${olm_bundle_dir}"
    else
      echo "Not validating non bundle format version ${olm_bundle_dir}"
    fi
  done
}

if [ "${ARCH}" != "ppc64le" ]; then
  verify "deploy/olm-catalog/ibm-block-csi-operator/" "NON_BUNDLE_FORMAT_VERSIONS_FOR_CERTIFIED"
  verify "deploy/olm-catalog/ibm-block-csi-operator-community/" "NON_BUNDLE_FORMAT_VERSIONS_FOR_COMMUNITY"
else
  echo "The tool does not support ${ARCH} arch. Skipping OLM validation"
fi
