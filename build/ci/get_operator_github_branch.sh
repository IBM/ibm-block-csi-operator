#!/bin/bash -xe
set +o pipefail

operator_branch=develop
DOCKER_HUB_USERNAME=csiblock1
DOCKER_HUB_PASSWORD=$csiblock_dockerhub_password
wanted_image_tag=`echo $CI_ACTION_REF_NAME | sed 's|/|.|g'`

does_the_docker_image_has_tag(){
  driver_component=$1
  does_docker_image_has_tag=false
  export image_tags=`docker-hub tags --orgname csiblock1 --reponame ibm-block-csi-$driver_component --all-pages | grep $wanted_image_tag | awk '{print$2}'`
  for tag in $image_tags
  do
    if [[ "$tag" == "$wanted_image_tag" ]]; then
      does_docker_image_has_tag=true
      break
    fi
  done
  echo $does_docker_image_has_tag
}

does_controller_docker_image_has_tag=$(does_the_docker_image_has_tag controller)
does_node_docker_image_has_tag=$(does_the_docker_image_has_tag node)

if [ $does_controller_docker_image_has_tag == "true" ] && [ $does_node_docker_image_has_tag == "true" ]; then
  operator_branches=`curl -H "Authorization: token $github_token" https://api.github.com/repos/IBM/ibm-block-csi-operator/branches | jq -c '.[]' | jq -r .name`
  for branch_name in $operator_branches
  do
    if [ "$branch_name" == "$CI_ACTION_REF_NAME" ]; then
      operator_branch=$CI_ACTION_REF_NAME
    fi
  
  done
fi

docker_image_branch_tag=`echo $operator_branch| sed 's|/|.|g'`
echo "::set-output name=docker_image_branch_tag::${docker_image_branch_tag}"
