#!/bin/bash -x

get_all_pods_by_type (){
    pod_type=$1
    kubectl get pod -l csi | grep $pod_type | awk '{print$1}'
}

save_action_output (){
    pod_type=$1
    action=$2
    action_for_artifacte_files=`echo $action | awk '{print$1}'`
    extra_args=$3
    pod_names=$(get_all_pods_by_type $pod_type)
    kubectl $action $pod_names $extra_args > "/tmp/${pod_names}_${action_for_artifacte_files}.txt"
}

declare -a pod_types=(
    "node"
    "controller"
    "operator"
)

for pod_type in "${pod_types[@]}"
do
    save_action_output $pod_type logs "-c ibm-block-csi-$pod_type"
    save_action_output $pod_type "describe pod" ""
done
