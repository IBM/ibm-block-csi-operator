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
        KUSTOMIZE_PLATFORM=darwin_amd64
fi

KUSTOMIZE_URL="https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_${KUSTOMIZE_PLATFORM}.tar.gz"

if [ ! -d "${OUTDIR_BIN}" ]; then
        mkdir -p "${OUTDIR_BIN}"
fi

if [ ! -x "${KUSTOMIZE_BIN}" ] || [[ -x "${KUSTOMIZE_BIN}" && "$(${KUSTOMIZE_BIN} version | awk '{print $1}' | awk -F '/' '{print $2}')" != "${KUSTOMIZE_VERSION}" ]]; then
        echo "Downloading kustomize ${KUSTOMIZE_VERSION} CLI tool for ${LOCAL_OS_TYPE}..."
        curl -JL "${KUSTOMIZE_URL}" | tar -zxvf - -C "${OUTDIR_BIN}"
        chmod +x "${KUSTOMIZE_BIN}"
else
        echo "Using kustomize cached at ${KUSTOMIZE_BIN}"
fi
