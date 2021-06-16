#!/bin/bash -xe
set +o pipefail

export passed_k8s_checks=true
export passed_openshift_checks=true

did_all_checks_pass(){
  community_operators_branch=$1
  all_checks_passed=false
  if [ "$(gh pr checks $community_operators_branch --repo $github_csiblock_community_operators_repository | grep -iv pass | wc -l)" -eq 0 ]
  then
    export all_checks_passed=true
  fi
  echo "$all_checks_passed"
}

passed_k8s_checks=$(did_all_checks_pass $community_operators_kubernetes_branch)
passed_openshift_checks=$(did_all_checks_pass $community_operators_openshift_branch)

if [ $passed_k8s_checks == "false" ] || [ $passed_openshift_checks == "false" ]
then
  echo "some test failed :("
  exit 1
fi
