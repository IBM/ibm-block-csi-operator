#!/bin/bash -xe
set +o pipefail

driver_images_tag_from_branch=latest
triggering_branch=$CI_ACTION_REF_NAME
branch_image_tag=`build/ci/get_image_tags_from_branch.sh ${triggering_branch}`
target_specific_tag=`echo $branch_image_tag | awk '{print$2}'`

is_private_branch_component_image_exists(){
  driver_component=$1
  image_to_check=$csiblock_docker_registry_username/ibm-block-csi-$driver_component:$target_specific_tag
  is_image_tag_exists=false
  export driver_image_inspect=`docker manifest inspect $image_to_check &> /dev/null; echo $?`
}

is_controller_image_tag_exists=$(is_private_branch_component_image_exists controller)
is_node_image_tag_exists=$(is_private_branch_component_image_exists node)

if [ $is_controller_image_tag_exists == "0" ] && [ $is_node_image_tag_exists == "0" ]; then
  driver_images_tag_from_branch=$target_specific_tag
fi

echo "::set-output name=image_branch_tag::${driver_images_tag_from_branch}"
