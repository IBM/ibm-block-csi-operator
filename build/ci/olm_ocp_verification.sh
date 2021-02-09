#!/bin/bash -e

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

if [ "${ARCH}" !=  "ppc64le" ]; then
  for olm_ocp_dict in deploy/olm-catalog/ibm-block-csi-operator/*/ ; do
    echo "Validating ${olm_ocp_dict}"
    operator-sdk bundle --verbose validate "${olm_ocp_dict}"
  done
else
  echo "Skipping OLM OCP validation on ${ARCH} arch"
fi
