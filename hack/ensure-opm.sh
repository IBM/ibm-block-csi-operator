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
        OPM_PLATFORM=darwin-amd64-opm
fi

OPM_URL="https://github.com/operator-framework/operator-registry/releases/download/${OPM_VERSION}/${OPM_PLATFORM}"

if [ ! -d "${OUTDIR_BIN}" ]; then
        mkdir -p "${OUTDIR_BIN}"
fi

if [ ! -x "${OPM_BIN}" ] || [[ -x "${OPM_BIN}" && "$(${OPM_BIN} version | awk -F '"' '{print $2}')" != "${OPM_VERSION}" ]]; then
        echo "Downloading opm ${OPM_VERSION} CLI tool for ${LOCAL_OS_TYPE}..."
        curl -JL "${OPM_URL}" -o "${OPM_BIN}"
        chmod +x "${OPM_BIN}"
else
        echo "Using opm cached at ${OPM_BIN}"
fi

FULLPATH_OPM_BIN=$(readlink -f "${OPM_BIN}")
export OPM_BIN="${FULLPATH_OPM_BIN}"
echo "Using opm at full path ${OPM_BIN}"