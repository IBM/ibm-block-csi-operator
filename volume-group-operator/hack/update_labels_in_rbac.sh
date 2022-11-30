#!/bin/bash -e

#
# Copyright 2022 IBM Corp.
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

base_path=config/rbac/
generated_rbac_file_name_prefix=/rbac.authorization.k8s.io_v1_
generated_file_prefix=_volume-group-operator.yaml
v1_file_prefix=v1_
service_account_kind=serviceaccount
service_account_file_suffix=service_account
declare -A rbac_kinds=( ["clusterrole"]="role" ["clusterrolebinding"]="role_binding")
kustomize build ${base_path} -o ${base_path}

for rbac_kind in "${!rbac_kinds[@]}"; do
    mv ${base_path%%/}/${generated_rbac_file_name_prefix}${rbac_kind}${generated_file_prefix} \
        ${base_path%%/}/${rbac_kinds[$rbac_kind]}.yaml
done
mv ${base_path%%/}/${v1_file_prefix}${service_account_kind}${generated_file_prefix} \
    ${base_path%%/}/${service_account_file_suffix}.yaml
