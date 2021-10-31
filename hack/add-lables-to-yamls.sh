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

declare -A labels_yamls=(
    ["config/rbac/role.yaml"]="config/rbac/patches/role_labels_patch.yaml"
    ["config/crd/bases/csi.ibm.com_ibmblockcsis.yaml"]="config/crd/patches/labels_patch.yaml"
)
merge_yamls (){
  yaml_file=$1
  required_lables=$2
  yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' ${yaml_file} ${required_lables} -i
}

for yaml_file in ${!labels_yamls[@]}; do
    required_lables=${labels_yamls[${yaml_file}]}
    merge_yamls $yaml_file $required_lables
done
