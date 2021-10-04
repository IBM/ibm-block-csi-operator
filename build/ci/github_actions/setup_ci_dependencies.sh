#!/bin/bash -xe
set +o pipefail

install_ci_dependencies (){
  build/ci/github_actions/setup_yq.sh
  python -m pip install --upgrade pip==21.2.4
  echo docker-hub==2.2.0 > dev-requirements.txt
  pip install -r dev-requirements.txt
}

install_ci_dependencies
# CSI-3173 - move image_version value into a common config file
image_version=`cat version/version.go | grep -i driverversion | awk -F = '{print $2}'`
image_version=`echo ${image_version//\"}`
operator_image_tags_for_test=`build/ci/get_image_tags_from_branch.sh ${CI_ACTION_REF_NAME} ${image_version} ${build_number} ${GITHUB_SHA}`
image_branch_tag=`echo $operator_image_tags_for_test | awk '{print$2}'`
operator_specific_tag_for_test=`echo $operator_image_tags_for_test | awk '{print$1}'`

if [ "$image_branch_tag" == "develop" ]; then
  image_branch_tag=latest
fi

echo "::set-output name=image_branch_tag::${image_branch_tag}"
echo "::set-output name=operator_specific_tag_for_test::${operator_specific_tag_for_test}"
