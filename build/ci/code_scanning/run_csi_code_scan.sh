#!/bin/bash -x

CODE_SCANNING_STAGE=$1
OUTPUT_PATH="`pwd`/build/reports"

if [ ${CODE_SCANNING_STAGE} == "operator" ]
then
  docker build -f build/ci/code_scanning/Dockerfile-csi-operator-code-scan -t csi-operator-code-scan . && \
  docker run --rm -t -v ${OUTPUT_PATH}:/results csi-operator-code-scan
else
  docker build -f build/ci/code_scanning/Dockerfile-csi-operator-dep-code-scan -t csi-operator-dep-code-scan . && \
  docker run --rm -t -v ${OUTPUT_PATH}:/results csi-operator-dep-code-scan
fi
