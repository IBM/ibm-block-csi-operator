#!/bin/bash
#
# Copyright contributors to the ibm-block-csi-operator project
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


set -e

source hack/common.sh

if [ "$LOCAL_OS_TYPE" == "Darwin" ]; then
        OPERATOR_SDK_PLATFORM=darwin_amd64
fi

OPERATOR_SDK_URL="https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk_${OPERATOR_SDK_PLATFORM}"

if [ ! -d "${OUTDIR_BIN}" ]; then
        mkdir -p "${OUTDIR_BIN}"
fi

if [ ! -x "${OPERATOR_SDK_BIN}" ] || [[ -x "${OPERATOR_SDK_BIN}" && "$(${OPERATOR_SDK_BIN} version | awk -F '"' '{print $2}')" != "${OPERATOR_SDK_VERSION}" ]]; then
        echo "Downloading operator-sdk ${OPERATOR_SDK_VERSION} CLI tool for ${LOCAL_OS_TYPE}..."
        curl -JL "${OPERATOR_SDK_URL}" -o "${OPERATOR_SDK_BIN}"
        chmod +x "${OPERATOR_SDK_BIN}"
else
        echo "Using operator-sdk cached at ${OPERATOR_SDK_BIN}"
fi
