#!/bin/bash -e

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

get_current_csi_version (){
  current_csi_version=$(cat version/version.go | grep -i driverversion | awk -F = '{print $2}')
  echo ${current_csi_version//\"}
}

are_csv_files_exsists_in_current_csi_version (){
  current_csi_version=$(get_current_csi_version)
  if ! compgen -G "${PWD}/deploy/olm-catalog/*/$current_csi_version" > /dev/null; then
    exit 0
  fi
}

get_csv_files (){
  current_csi_version=$(get_current_csi_version)
  ls deploy/olm-catalog/*/$current_csi_version/manifests/ibm-block-csi-operator.v$current_csi_version.clusterserviceversion.yaml
}

get_bundle_crds (){
  current_csi_version=$(get_current_csi_version)
  ls deploy/olm-catalog/*/$current_csi_version/manifests/csi.ibm.com_ibmblockcsis.yaml
}

align_roles (){
  are_csv_files_exsists_in_current_csi_version
  csv_files=$(get_csv_files)
  for csv_file in $csv_files
  do
    yq eval-all 'select(fileIndex==0).spec.install.spec.clusterPermissions[0].rules = select(fileIndex==1).rules | select(fi==0)'  $csv_file config/rbac/role.yaml -i
  done
}

align_crds (){
  are_csv_files_exsists_in_current_csi_version
  bundle_crds=$(get_bundle_crds)
  for bundle_crd in $bundle_crds
  do
    yes | cp config/crd/bases/csi.ibm.com_ibmblockcsis.yaml $bundle_crd
  done
}

main() {
  align_roles
  align_crds
}

if [[ "${0##*/}" == "update-yamls-with-the-same-content.sh" ]]; then
    main
fi

