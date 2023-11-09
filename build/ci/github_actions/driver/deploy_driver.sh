#!/bin/bash -xel
set +o pipefail

expected_node_image=$node_repository_for_test:$driver_images_tag
expected_controller_image=$controller_repository_for_test:$driver_images_tag

install_worker_prerequisites() {
  kind_node_name=$(docker ps --format "{{.Names}}")
  docker exec -i $kind_node_name apt-get update
  docker exec -i $kind_node_name apt -y install open-iscsi
}

edit_cr_images (){
  chmod 547 $cr_yaml
  declare -A cr_image_fields=(
      [".spec.controller.repository"]="$controller_repository_for_test"
      [".spec.controller.tag"]="$driver_images_tag"
      [".spec.node.repository"]="$node_repository_for_test"
      [".spec.node.tag"]="$driver_images_tag"
  )
  for image_field in ${!cr_image_fields[@]}; do
      cr_image_value=${cr_image_fields[${image_field}]}
      yq eval "${image_field} |= \"${cr_image_value}\"" $cr_yaml -i
  done
}

install_worker_prerequisites
edit_cr_images
cat $cr_yaml | grep repository:
cat $cr_yaml | grep tag:
kubectl apply -f $cr_yaml
source build/ci/github_actions/deployment.sh
wait_for_driver_deployment_to_start
assert_driver_images_in_pods $expected_node_image $expected_controller_image
wait_for_driver_deployment_to_finish
