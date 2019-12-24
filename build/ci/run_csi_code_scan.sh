#!/bin/bash -x

CODE_SCANNING_STAGE=$1
OUTPUT_PATH="`pwd`/build/reports"

if [ -z "$GOSEC_EXCLUDE" ]
then
    GOSEC_EXCLUDE=""
fi

if [ -z "$OWASPDC_EXCLUDE" ]
then
    OWASPDC_EXCLUDE=""
fi


if [ ${CODE_SCANNING_STAGE} == "operator" ]
then
  OPERATOR_OUTPUT_PATH="${OUTPUT_PATH}/GOSEC_OPERATOR"
  mkdir -p ${OPERATOR_OUTPUT_PATH} && chmod 777 ${OPERATOR_OUTPUT_PATH}
  docker build -f Dockerfile-csi-operator-code-scan -t csi-operator-code-scan . && \
  docker run --rm -t -v ${OPERATOR_OUTPUT_PATH}:/results -e EXCLUDE=${GOSEC_EXCLUDE} csi-operator-code-scan
else
  OPERATOR_DEP_OUTPUT_PATH="${OUTPUT_PATH}/OWASPDC_OPERATOR_DEP"
  ./build/ci/csi_operator_dep_code_scan.sh ${OPERATOR_DEP_OUTPUT_PATH} ${OWASPDC_EXCLUDE}
fi

