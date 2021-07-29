#!/bin/bash -xe
set +o pipefail

driver_is_ready=false
actual_driver_running_time_in_seconds=0
minimum_driver_running_time_in_seconds=10
declare -a driver_pods_types=(
  "controller"
  "node"
)

get_csi_pods (){
  kubectl get pod -A -l csi
}

wait_for_driver_pod_to_start (){
  driver_pod_type=$1
  while [ "$(get_csi_pods | grep $driver_pod_type | wc -l)" -eq 0 ]; do
    echo "The $driver_pod_type is not deployed"
    sleep 1
  done
}

wait_for_driver_deployment_to_start (){
  for driver_pods_type in "${driver_pods_types[@]}"
  do
    wait_for_driver_pod_to_start $driver_pods_type
  done
}

wait_for_driver_deployment_to_finish (){
  while [ $driver_is_ready == "false" ]; do
    if [ "$(get_csi_pods | grep -iv running | grep -iv name | wc -l)" -eq 0 ]; then
      ((++actual_driver_running_time_in_seconds))
      if [ $actual_driver_running_time_in_seconds -eq $minimum_driver_running_time_in_seconds ]; then
        driver_is_ready=true
      fi
    else
      actual_driver_running_time_in_seconds=0
    fi
    get_csi_pods
    sleep 1
  done
}

is_expected_image_in_pod (){
  pod_type=$1
  wanted_pod_image=$2
  image_in_pod=`kubectl describe pod $(get_csi_pods | grep $pod_type) | grep -i image: | \
   grep $wanted_pod_image | awk -F Image: '{print$2}' | awk '{print$1}'`
  if [[ $image_in_pod != $wanted_pod_image ]]; then
    echo "$pod_type's image is not the wanted image ($wanted_pod_image)"
    exit 1
  fi
}

chec_pods_images (){
  expected_node_image=$node_repository_for_test:$driver_images_tag
  expected_controller_image=$controller_repository_for_test:$driver_images_tag
  expected_operator_image=$operator_image_repository_for_test:$operator_image_tag_for_test
  is_expected_image_in_pod "operator" $wanted_operator_image
  is_expected_image_in_pod "controller" $wanted_controller_image
  is_expected_image_in_pod "node" $wanted_node_image
}

wait_for_driver_deployment_to_start
wait_for_driver_deployment_to_finish
chec_pods_images
echo Driver is running
get_csi_pods
