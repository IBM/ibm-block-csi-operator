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

base_path=config/crd/bases/
generated_crd_file_name_prefix=/apiextensions.k8s.io_v1_customresourcedefinition_
api_group=csi.ibm.com
kinds=(volumegroups volumegroupclasses volumegroupcontents)
kustomize build config/crd/ -o ${base_path}
for kind in ${kinds[@]}; do
    mv ${base_path%%/}${generated_crd_file_name_prefix}${kind}.${api_group}.yaml ${base_path%%/}/${api_group}_${kind}.yaml
done
