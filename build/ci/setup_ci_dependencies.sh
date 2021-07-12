#!/bin/bash -xe
set +o pipefail

python -m pip install --upgrade pip docker-hub==2.2.0
echo yq > dev-requirements.txt

cat >>/home/runner/.bash_profile <<'EOL'
yq() {
  docker run --rm -e operator_image_for_test=$operator_image_for_test\
                  -e cr_image_value=$cr_image_value\
                  -i -v "${PWD}":/workdir mikefarah/yq "$@"
}
EOL

source /home/runner/.bash_profile
cd deploy/olm-catalog/ibm-block-csi-operator
image_version=`yq eval .channels[0].currentCSV ibm-block-csi-operator.package.yaml`
image_version=`echo ${image_version//ibm-block-csi-operator.v}`
echo "::set-output name=image_version::${image_version}"
