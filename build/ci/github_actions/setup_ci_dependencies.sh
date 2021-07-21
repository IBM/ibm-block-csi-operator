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

image_version=`cat version/version.go | grep -i driverversion | awk -F = '{print $2}'`
image_version=`echo ${image_version//\"}`
# CSI-3173 - move image_version value into a common config file
operator_image_tag_for_test=`build/ci/get_image_tag_from_branch.sh ${image_version} ${build_number} ${CI_ACTION_REF_NAME}`
operator_image_tag_for_test=`echo $operator_image_tag_for_test | awk '{print$1}'`
docker_image_branch_tag=`echo $operator_image_tag_for_test | awk '{print$2}'`
echo "::set-output name=docker_image_branch_tag::${docker_image_branch_tag}"
echo "::set-output name=operator_image_tag_for_test::${operator_image_tag_for_test}"
