#!/bin/bash -xe

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

# Run operator-sdk generate k8s and operator-sdk generate openapi to update code after crd changes.

if ! [ -x "$(command -v operator-sdk)" ]; then
  echo 'Error: operator-sdk is not installed.' >&2
  exit 1
fi

echo "run operator-sdk generate k8s"
operator-sdk generate k8s

echo "run operator-sdk generate openapi"
operator-sdk generate openapi
