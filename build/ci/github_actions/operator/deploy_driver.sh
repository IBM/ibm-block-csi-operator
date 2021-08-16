#!/bin/bash -xel
set +o pipefail

install_worker_prerequisites() {
  kind_node_name=`docker ps --format "{{.Names}}"`
  docker exec -i $kind_node_name apt-get update
  docker exec -i $kind_node_name apt -y install open-iscsi
}

edit_cr_images (){
  cd $(dirname $cr_file)
  chmod 547 $(basename $cr_file)
  declare -A cr_image_fields=(
      [".spec.controller.repository"]="$controller_repository_for_test"
      [".spec.controller.tag"]="$driver_images_tag"
      [".spec.node.repository"]="$node_repository_for_test"
      [".spec.node.tag"]="$driver_images_tag"
  )
  for image_field in ${!cr_image_fields[@]}; do
      cr_image_value=${cr_image_fields[${image_field}]}
      yq eval "${image_field} |= \"${cr_image_value}\"" $(basename $cr_file) -i
  done
  cd -
}

edit_operator_yaml_image (){
  cd $(dirname $operator_yaml)
  operator_image_in_branch=`yq eval '(. | select(.kind == "Deployment") | .spec.template.spec.containers[0].image)' $(basename $operator_yaml)`
  sed -i "s+$operator_image_in_branch+$operator_image_for_test+g" $(basename $operator_yaml) ## TODO: CSI-3223 need to edit the operator image only in a specifics places
cd -
}

install_worker_prerequisites
edit_cr_images
edit_operator_yaml_image

cat $operator_yaml | grep image:
cat $cr_file | grep repository:
cat $cr_file | grep tag:

kubectl apply -f $operator_yaml
kubectl apply -f $cr_file
