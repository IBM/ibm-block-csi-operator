#!/bin/bash -xe
set +o pipefail

# CSI-3172 - run rad hat bot checks

did_all_checks_pass(){
  community_operators_branch=$1
  forked_repository=$2
  all_checks_passed=false
  if [ "$(gh pr checks $community_operators_branch --repo $forked_repository | grep -iv pass | wc -l)" -eq 0 ]
  then
    all_checks_passed=true
  fi
  echo "$all_checks_passed"
}

passed_k8s_checks=$(did_all_checks_pass $community_operators_kubernetes_branch $forked_community_operators_repository)
passed_openshift_checks=$(did_all_checks_pass $community_operators_openshift_branch $forked_community_operators_repository_prod)

if [ $passed_k8s_checks == "false" ] || [ $passed_openshift_checks == "false" ]
then
  echo "some test failed :("
  exit 1
fi
