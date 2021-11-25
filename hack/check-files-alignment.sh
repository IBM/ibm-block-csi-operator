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

operator_yaml_path=deploy/installer/generated/ibm-block-csi-operator.yaml
roles_yaml_path=config/rbac/role.yaml
origin_crd_yaml_path=config/crd/bases/csi.ibm.com_ibmblockcsis.yaml

check_generation (){
  echo "check generation"
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

verify_full_operator_yaml_is_align(){
  echo "check full operator yaml alignment"
  declare -A orignial_yamls=(
      ["config/rbac/role.yaml"]="ClusterRole"
      ["config/crd/bases/csi.ibm.com_ibmblockcsis.yaml"]="CustomResourceDefinition"
      ["config/rbac/service_account.yaml"]="ServiceAccount"
      ["config/rbac/role_binding.yaml"]="ClusterRoleBinding"
      ["config/manager/manager.yaml"]="Deployment"
  )
  for orignial_yaml in ${!orignial_yamls[@]}; do
    export resource_type=${orignial_yamls[${orignial_yaml}]}
    diff <(yq e '... comments=""' $orignial_yaml) <(yq eval '(. | select(.kind == env(resource_type)))' $operator_yaml_path)
  done
}

verify_no_roles_diff (){
  echo "check roles alignment"
  source hack/get_information_helper.sh
  are_manifest_files_exsists_in_current_csi_version
  csv_files=$(get_csv_files)
  for csv_file in $csv_files; do
    diff <(yq e .rules $roles_yaml_path) <(yq e .spec.install.spec.clusterPermissions[0].rules $csv_file)
  done
}

verify_no_crds_diff (){
  echo "check crds alignment"
  source hack/get_information_helper.sh
  are_manifest_files_exsists_in_current_csi_version
  crd_files=$(get_bundle_crds)
  for crd_file in $crd_files; do
    diff $origin_crd_yaml_path $crd_file
  done
}

check_generation
verify_full_operator_yaml_is_align
verify_no_roles_diff
verify_no_crds_diff