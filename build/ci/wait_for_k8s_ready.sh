#!/bin/bash -xe
set +o pipefail

is_kubernetes_cluster_ready (){
  pods=`kubectl get pods -A | awk '{print$3}' | grep -iv ready`
  all_containers_are_running=false
  has_not_ready_pod=false
  for pod in $pods; do
    running_containers_count=`echo $pod | awk -F / '{print$1}'`
    total_containers_count=`echo $pod | awk -F / '{print$2}'`
    if [ $running_containers_count != $total_containers_count ]; then
      has_not_ready_pod=true
      break
    fi
  done
  if [[ "$has_not_ready_pod" == "false" ]]: then
    all_containers_are_running=true
  fi
  
  echo $all_containers_are_running
}

while [[ `is_kubernetes_cluster_ready` == "false" ]]; do
        kubectl get pods -A
done
