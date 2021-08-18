#!/bin/bash -x

get_all_pods_by_type (){
    pod_type=$1
    kubectl get pod -l csi | grep $pod_type | awk '{print$1}'
}

run_action_and_save_output (){
    pod_type=$1
    action=$2
    action_name=$3
    extra_args=$4
    container_name=$5
    pod_names=$(get_all_pods_by_type $pod_type)
    kubectl $action $pod_names $extra_args > "/tmp/${pod_names}_${container_name}_${action_name}.txt"
}

save_logs_of_all_containers_in_pod (){
    pod_type=$1
    pod_names=$(get_all_pods_by_type $pod_type)
    containers=`kubectl get pods $pod_names -o jsonpath='{.spec.containers[*].name}'`
    for container in $containers
    do
        run_action_and_save_output $pod_type logs "log" "-c $container" $container
    done
}

declare -a pod_types=(
    "node"
    "controller"
    "operator"
)

for pod_type in "${pod_types[@]}"
do
    save_logs_of_all_containers_in_pod $pod_type
    run_action_and_save_output $pod_type "describe pod" "describe" "" $pod_type
done
