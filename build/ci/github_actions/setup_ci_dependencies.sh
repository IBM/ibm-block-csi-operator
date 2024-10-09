#!/bin/bash -xe
set +o pipefail

triggering_branch=$(echo ${GITHUB_HEAD_REF:-${GITHUB_REF#refs/heads/}})
image_version=${IMAGE_VERSION}
build_number=${BUILD_NUMBER}
commit_hash=${GITHUB_SHA:0:7}
specific_tag="${image_version}_b${build_number}_${commit_hash}_${triggering_branch}"


if [ "$triggering_branch" == "develop" ]; then
  global_tag=latest
else
  global_tag=${triggering_branch}
fi

if [ "$PRODUCTION" = true ]; then
  repository=${PROD_REPOSITORY}
  global_tag=${image_version}
else
  repository=${STAGING_REPOSITORY}
fi

echo "repository=${repository}" >> $GITHUB_OUTPUT
echo "specific_tag=${specific_tag}" >> $GITHUB_OUTPUT
echo "global_tag=${global_tag}" >> $GITHUB_OUTPUT
