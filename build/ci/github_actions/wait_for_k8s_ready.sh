#!/bin/bash -xe
set +o pipefail

are_pods_ready (){
  pods=$@
  for pod in $pods; do
    running_containers_count=`echo $pod | awk -F / '{print$1}'`
    total_containers_count=`echo $pod | awk -F / '{print$2}'`
    if [ $running_containers_count != $total_containers_count ]; then
      echo true
      break
    fi
  done
  echo false
}

is_kubernetes_cluster_ready (){
  pods=`kubectl get pods -A | awk '{print$3}' | grep -iv ready`
  are_all_pods_ready=$(are_pods_ready $pods)
  if [ $are_all_pods_ready == "false" ]; then
      echo true
      break
  fi
  
  echo false
}

while [[ `is_kubernetes_cluster_ready` == "false" ]]; do
  kubectl get pods -A
done
