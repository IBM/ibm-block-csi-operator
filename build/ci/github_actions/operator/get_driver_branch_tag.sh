#!/bin/bash -xe
set +o pipefail

operator_branch=develop
DOCKER_HUB_USERNAME=csiblock1
DOCKER_HUB_PASSWORD=$csiblock_dockerhub_password
triggering_branch=$CI_ACTION_REF_NAME
target_image_tags=`build/ci/get_image_tags_from_branch.sh ${triggering_branch}`
target_image_tag=`echo $target_image_tags | awk '{print$2}'`

is_private_branch_controller_image_exists(){
  driver_component=$1
  is_docker_image_has_tag=false
  export image_tags=`docker-hub tags --orgname csiblock1 --reponame ibm-block-csi-$driver_component --all-pages | grep $target_image_tag | awk '{print$2}'`
  for tag in $image_tags
  do
    if [[ "$tag" == "$target_image_tag" ]]; then
      is_docker_image_has_tag=true
      break
    fi
  done
  echo $is_docker_image_has_tag
}

is_controller_docker_image_has_tag=$(is_private_branch_controller_image_exists controller)
is_node_docker_image_has_tag=$(is_private_branch_controller_image_exists node)

if [ $is_controller_docker_image_has_tag == "true" ] && [ $is_node_docker_image_has_tag == "true" ]; then
  operator_branch=$target_image_tag
fi

echo "::set-output name=docker_image_branch_tag::${operator_branch}"
