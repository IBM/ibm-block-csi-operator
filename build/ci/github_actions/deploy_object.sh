#!/bin/bash -xe
set +o pipefail

is_driver_ready=false
actual_driver_running_time_in_seconds=0
minimum_driver_running_time_in_seconds=10
containers_prefix=ibm-block-csi
declare -a driver_pods_types=(
  "controller"
  "node"
)

get_csi_pods (){
  kubectl get pod -A -l csi
}

get_operator_pod (){
  kubectl get pod -A -l app.kubernetes.io/name=ibm-block-csi-operator
}

get_image_pod_by_type (){
  pod_type=$1
  component_to_check=$2
  containers_images=`kubectl get pods $(get_csi_pods | grep $pod_type | awk '{print$2}') -o jsonpath='{range .spec.containers[*]}{.name},{.image} {end}'`
  for containers_image in $containers_images
  do
    if [[  "$containers_image" =~ "$component_to_check," ]]; then
      echo $containers_image | awk -F , '{print$2}'
      break
    fi
  done
}

wait_for_pod_to_start (){
  driver_pod_type=$1
  while [ "$(get_csi_pods | grep $driver_pod_type | wc -l)" -eq 0 ]; do
    echo "The $driver_pod_type is not deployed"
    sleep 1
  done
}

wait_for_driver_deployment_to_start (){
  for driver_pods_type in "${driver_pods_types[@]}"
  do
    wait_for_pod_to_start $driver_pods_type
  done
}

wait_for_driver_deployment_to_finish (){
  wait_for_pods_to_finish get_csi_pods
}

wait_for_operator_deployment_to_finish (){
  wait_for_pods_to_finish get_operator_pod
}


wait_for_pods_to_finish (){
  get_pods_command=$1
  while [ $is_driver_ready == "false" ]; do
    if [ "$($get_pods_command | grep -iv running | grep -iv name | wc -l)" -eq 0 ]; then
      ((++actual_driver_running_time_in_seconds))
      if [ $actual_driver_running_time_in_seconds -eq $minimum_driver_running_time_in_seconds ]; then
        is_driver_ready=true
      fi
    else
      actual_driver_running_time_in_seconds=0
    fi
    get_csi_pods
    sleep 1
  done
}

assert_expected_image_in_pod (){
  pod_type=$1
  expected_pod_image=$2
  component_to_check=$containers_prefix-$pod_type
  image_in_pod=`get_image_pod_by_type $pod_type $component_to_check`
  if [[ $image_in_pod != $expected_pod_image ]]; then
    echo "$pod_type's image ($image_in_pod) is not the expected image ($expected_pod_image)"
    exit 1
  fi
}

assert_pods_images (){
  expected_node_image=$1
  expected_controller_image=$2
  declare -A drivers_components_in_k8s=(
      ["controller"]="$expected_controller_image"
      ["node"]="$expected_node_image"
  )
  for driver_component in ${!drivers_components_in_k8s[@]}; do
      driver_component_expected_image=${drivers_components_in_k8s[${driver_component}]}
      assert_expected_image_in_pod $driver_component $driver_component_expected_image
  done
}
