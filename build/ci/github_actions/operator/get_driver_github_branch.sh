#!/bin/bash -xe
set +o pipefail

operator_branch=develop
DOCKER_HUB_USERNAME=csiblock1
DOCKER_HUB_PASSWORD=$csiblock_dockerhub_password
triggering_branch=$CI_ACTION_REF_NAME
target_image_tag=`echo $triggering_branch | sed 's|/|.|g'`

is_private_branch_component_image_exist(){
  driver_component=$1
  does_docker_image_has_tag=false
  export image_tags=`docker-hub tags --orgname csiblock1 --reponame ibm-block-csi-$driver_component --all-pages | grep $target_image_tag | awk '{print$2}'`
  for tag in $image_tags
  do
    if [[ "$tag" == "$target_image_tag" ]]; then
      does_docker_image_has_tag=true
      break
    fi
  done
  echo $does_docker_image_has_tag
}

is_controller_docker_image_has_tag=$(is_private_branch_component_image_exist controller)
is_node_docker_image_has_tag=$(is_private_branch_component_image_exist node)

if [ $is_controller_docker_image_has_tag == "true" ] && [ $is_node_docker_image_has_tag == "true" ]; then
  operator_branch=$triggering_branch
fi

docker_image_branch_tag=`echo $operator_branch| sed 's|/|.|g'`
echo "::set-output name=docker_image_branch_tag::${docker_image_branch_tag}"
