#!/bin/bash -xe
set +o pipefail

export passed_k8s_checks=true
export passed_openshift_checks=true

wait_for_checks_to_complete(){
  community_operators_branch=$1
  is_summary_test_passed=false
  export repo_pr=`gh pr list --repo $github_csiblock_community_operators_repository | grep $community_operators_branch`
  if [[ "$repo_pr" == *"$community_operators_branch"* ]]; then
    sleep 5
    while [ "$(gh pr checks $community_operators_branch --repo $github_csiblock_community_operators_repository | grep -i pending | wc -l)" -gt 0 ]; do
      sleep 15
    done
    summary_counter=0
    export search_summary_test=`gh pr checks $community_operators_branch --repo $github_csiblock_community_operators_repository | grep -i summary`
    while [ "$search_summary_test" != *"pending"* ] && [ $summary_counter -lt 20 ] && [ $is_summary_test_passed == "false" ]; do
      sleep 1
      export search_summary_test=`gh pr checks $community_operators_branch --repo $github_csiblock_community_operators_repository | grep -i summary`
      ((summary_counter=summary_counter+1))
      if [[ "$search_summary_test" == *"pass"* ]]; then
        is_summary_test_passed=true
      fi
    done
    if [[ $is_summary_test_passed == "false" ]]; then
      while [ "$(gh pr checks $community_operators_branch --repo $github_csiblock_community_operators_repository | grep -i pending | wc -l)" -gt 0 ]; do
      sleep 5
    done
    fi
  fi
}

are_all_tests_were_passed(){
  community_operators_branch=$1
  all_checks_passed=false
  if [ "$(gh pr checks $community_operators_branch --repo $github_csiblock_community_operators_repository | grep -iv pass | wc -l)" -eq 0 ]
  then
    export all_checks_passed=true
  fi
  echo "$all_checks_passed"
}

wait_for_checks_to_complete $community_operators_kubernetes_branch
wait_for_checks_to_complete $community_operators_openshift_branch

passed_k8s_checks=$(are_all_tests_were_passed $community_operators_kubernetes_branch)
passed_openshift_checks=$(are_all_tests_were_passed $community_operators_openshift_branch)

if [ $passed_k8s_checks == "false" ] || [ $passed_openshift_checks == "false" ]
then
  echo "some test failed :("
  exit 1
fi
