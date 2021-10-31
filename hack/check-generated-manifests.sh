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

check_generation (){
  project_dirname=ibm-block-csi-operator
  cd ..
  cp -r $project_dirname ./$project_dirname-expected
  cd $project_dirname-expected/
  make update
  cd ..
  diff -qr --exclude=bin $project_dirname $project_dirname-expected/
  rm -rf $project_dirname-expected/
  cd $project_dirname
}

verify_no_roles_diff (){
  source hack/update-roles-in-csv.sh
  current_csi_version=$(get_current_csi_version)
  csv_files=$(ls deploy/olm-catalog/*/$current_csi_version/manifests/ibm-block-csi-operator.v$current_csi_version.clusterserviceversion.yaml)
  for csv_file in $csv_files; do
    diff <(yq e .rules config/rbac/role.yaml) <(yq e .spec.install.spec.clusterPermissions[0].rules $csv_file)
  done
}

check_generation
verify_no_roles_diff
