#!/bin/bash -xe
set +o pipefail

triggering_branch=$(echo ${GITHUB_HEAD_REF:-${GITHUB_REF#refs/heads/}})
image_version=${IMAGE_VERSION}
build_number=${BUILD_NUMBER}
commit_hash=${GITHUB_SHA:0:7}
specific_tag="${image_version}_b${build_number}_${commit_hash}_${triggering_branch}"
global_tag=${triggering_branch}
#
#if [ "$triggering_branch" == "develop" ]; then
#  global_tag=latest
#else
#  global_tag=${triggering_branch}
#fi
#
#if [ "$PRODUCTION" = true ]; then
#  operator_repository=${OPERATOR_PROD_REPOSITORY}
#  bundle_repository=${BUNDLE_OPERATOR_PROD_REPOSITORY}
#  catalog_repository=${CATALOG_OPERATOR_PROD_REPOSITORY}
#  global_tag=${image_version}
#else
#  operator_repository=${OPERATOR_STAGING_REPOSITORY}
#  bundle_repository=${BUNDLE_OPERATOR_STAGING_REPOSITORY}
#  catalog_repository=${CATALOG_OPERATOR_STAGING_REPOSITORY}
#fi

operator_repository=${OPERATOR_STAGING_REPOSITORY}
bundle_repository=${BUNDLE_OPERATOR_STAGING_REPOSITORY}
catalog_repository=${CATALOG_OPERATOR_STAGING_REPOSITORY}

echo "operator_repository=${operator_repository}" >> $GITHUB_OUTPUT
echo "bundle_repository=${bundle_repository}" >> $GITHUB_OUTPUT
echo "catalog_repository=${catalog_repository}" >> $GITHUB_OUTPUT
echo "specific_tag=${specific_tag}" >> $GITHUB_OUTPUT
echo "global_tag=${global_tag}" >> $GITHUB_OUTPUT
