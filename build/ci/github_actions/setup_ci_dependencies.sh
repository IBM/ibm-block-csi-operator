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

# CSI-3173 - move image_version value into a common config file
image_version=`cat version/version.go | grep -i driverversion | awk -F = '{print $2}'`
image_version=`echo ${image_version//\"}`
GITHUB_SHA=${GITHUB_SHA:0:7}_
operator_image_tags_for_test=`build/ci/get_image_tags_from_branch.sh ${CI_ACTION_REF_NAME} ${image_version} ${build_number} ${GITHUB_SHA}`
docker_image_branch_tag=`echo $operator_image_tags_for_test | awk '{print$2}'`
operator_image_tag_for_test=`echo $operator_image_tags_for_test | awk '{print$1}'`

if [ "$docker_image_branch_tag" == "develop" ]; then
  docker_image_branch_tag=latest
fi

echo "::set-output name=docker_image_branch_tag::${docker_image_branch_tag}"
echo "::set-output name=operator_image_tag_for_test::${operator_image_tag_for_test}"
