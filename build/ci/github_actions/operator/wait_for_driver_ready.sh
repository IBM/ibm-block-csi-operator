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

wait_for_driver_deployment_to_start
wait_for_driver_deployment_to_finish
echo Driver is running
get_csi_pods
