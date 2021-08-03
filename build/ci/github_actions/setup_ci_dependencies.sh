#!/bin/bash -xe
set +o pipefail

install_ci_dependencies (){
  python -m pip install --upgrade pip
  echo docker-hub==2.2.0 > dev-requirements.txt
  pip install -r dev-requirements.txt
}

install_ci_dependencies
cat >>/home/runner/.bash_profile <<'EOL'
yq() {
  docker run --rm -i -v "${PWD}":/workdir mikefarah/yq "$@"
}
EOL

# CSI-3173 - move image_version value into a common config file
image_version=`cat version/version.go | grep -i driverversion | awk -F = '{print $2}'`
image_version=`echo ${image_version//\"}`
GITHUB_SHA=${GITHUB_SHA:0:7}_
operator_image_tags_for_test=`build/ci/get_image_tags_from_branch.sh ${CI_ACTION_REF_NAME} ${image_version} ${build_number} ${GITHUB_SHA}`
image_branch_tag=`echo $operator_image_tags_for_test | awk '{print$2}'`
operator_image_tag_for_test=`echo $operator_image_tags_for_test | awk '{print$1}'`

if [ "$image_branch_tag" == "develop" ]; then
  image_branch_tag=latest
fi

echo "::set-output name=image_branch_tag::${image_branch_tag}"
echo "::set-output name=operator_image_tag_for_test::${operator_image_tag_for_test}"
