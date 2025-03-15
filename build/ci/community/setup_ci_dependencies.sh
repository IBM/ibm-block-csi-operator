#!/bin/bash -xe
set +o pipefail

echo $'yq() {\n  docker run --rm -e operator_image_for_test=$operator_image_for_test -i -v "${PWD}":/workdir mikefarah/yq:4 "$@"\n}' >> /home/runner/.bash_profile

curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt-get update
sudo apt-get install gh

git config --global user.email csi.block1@il.ibm.com
git config --global user.name csiblock

image_version=`cat version/version.go | grep -i driverversion | awk -F = '{print $2}'`
image_version=`echo ${image_version//\"}`

echo "::set-output name=image_branch_tag::${image_version}"