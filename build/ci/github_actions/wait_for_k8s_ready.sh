#!/bin/bash -xe
set +o pipefail

are_pods_ready (){
  pods=$@
  for pod in $pods; do
    running_containers_count=`echo $pod | awk -F / '{print$1}'`
    total_containers_count=`echo $pod | awk -F / '{print$2}'`
    if [ $running_containers_count != $total_containers_count ]; then
      echo false
      break
    fi
  done
  echo true
}

is_kubernetes_cluster_ready (){
  pods=`kubectl get pods -A | awk '{print$3}' | grep -iv ready`
  are_all_pods_ready=$(are_pods_ready $pods)
  echo $are_all_pods_ready
}

while [[ `is_kubernetes_cluster_ready` == "false" ]]; do
  kubectl get pods -A
done
