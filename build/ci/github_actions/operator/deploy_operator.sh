#!/bin/bash -xel
set +o pipefail

edit_operator_yaml_image (){
  operator_image_in_branch=`yq eval '(. | select(.kind == "Deployment") | .spec.template.spec.containers[0].image)' $operator_yaml`
  sed -i "s+$operator_image_in_branch+$operator_image_for_test+g" $operator_yaml ## TODO: CSI-3223 avoid using sed
}

edit_operator_yaml_image
cat $operator_yaml | grep image:
source build/ci/github_actions/deployment.sh
assert_operator_image_in_pod $operator_image_for_test
wait_for_operator_deployment_to_finish
