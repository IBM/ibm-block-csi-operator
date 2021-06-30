#!/bin/bash -xe
set +o pipefail

operator_branch=develop

token=`curl -X POST -H "Content-Type: application/json" -d '{"username": "csiblock1", "password": "'$csiblock_dockerhub_password'"}' https://hub.docker.com/v2/users/login | jq .token`
token=`echo ${token//\"}`
controller_image_tags=`curl -s -H "Authorization: JWT ${token}" https://hub.docker.com/v2/namespaces/csiblock1/repositories/ibm-block-csi-controller/images | jq .results[0] | jq .tags | jq -c '.[]' | jq .tag`
node_image_tags=`curl -s -H "Authorization: JWT ${token}" https://hub.docker.com/v2/namespaces/csiblock1/repositories/ibm-block-csi-node/images | jq .results[0] | jq .tags | jq -c '.[]' | jq .tag`
does_the_docker_image_has_tag(){
  image_tags=$@
  does_docker_image_has_tag=false
  for tag in $image_tags
  do
    tag=`echo ${tag//\"}`
    if [[ "$tag" == `echo $CI_ACTION_REF_NAME | sed 's|/|.|g'` ]]; then
      does_docker_image_has_tag=true
    fi
  done
  echo $does_docker_image_has_tag
}

does_controller_docker_image_has_tag=$(does_the_docker_image_has_tag $controller_image_tags)
does_node_docker_image_has_tag=$(does_the_docker_image_has_tag $node_image_tags)

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
echo "::set-output name=operator_branch::${operator_branch}"
