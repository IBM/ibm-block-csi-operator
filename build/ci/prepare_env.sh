#!/bin/bash -xe
set +o pipefail

export csv_version=`yq .channels[0].currentCSV deploy/olm-catalog/ibm-block-csi-operator-community/ibm-block-csi-operator.package.yaml`
export csv_version=`echo ${csv_version//ibm-block-csi-operator.v}`
export csv_version=`echo ${csv_version//\"}`
export repository_path=~/work/$CI_REPOSITORY_NAME/$CI_REPOSITORY_NAME
export csv_file=$repository_path/deploy/olm-catalog/ibm-block-csi-operator-community/$csv_version/ibm-block-csi-operator.v$csv_version.clusterserviceversion.yaml
export operator_image_in_branch=`yq .metadata.annotations.containerImage $csv_file`
export operator_image_in_branch=`echo ${operator_image_in_branch//\"}`
export upstream_community_operators_path=upstream-community-operators/ibm-block-csi-operator-community/$csv_version
export community_operators_path=community-operators/ibm-block-csi-operator-community/$csv_version

echo "::set-output name=operator_image_in_branch::${operator_image_in_branch}"
echo "::set-output name=upstream_community_operators_path::${upstream_community_operators_path}"
echo "::set-output name=community_operators_path::${community_operators_path}"
echo "::set-output name=repository_path::${repository_path}"
echo "::set-output name=repository_path::${csv_file}"
