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

declare -A yamls_with_required_lables=(
    ["config/rbac/role.yaml"]="config/rbac/patches/role_labels_patch.yaml"
    ["config/crd/bases/csi.ibm.com_ibmblockcsis.yaml"]="config/crd/patches/labels_patch.yaml"
)
for yaml_file in ${!yamls_with_required_lables[@]}; do
    required_lables=${yamls_with_required_lables[${yaml_file}]}
    yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' ${yaml_file} ${required_lables} -i
done
