#!/bin/bash -xe
set +o pipefail

sed -i "s+$current_operator_image+$wanted_operator_image+g" $csv_file
sed -i 's/.*minKubeVersion.*/  minKubeVersion: 1.19.0/' $csv_file

mkdir $repository_path/upstream-community-operators/
mkdir $repository_path/community-operators/

cp -r $repository_path/deploy/olm-catalog/ibm-block-csi-operator-community/ $repository_path/upstream-community-operators/
cp -r $repository_path/deploy/olm-catalog/ibm-block-csi-operator-community/ $repository_path/community-operators/
