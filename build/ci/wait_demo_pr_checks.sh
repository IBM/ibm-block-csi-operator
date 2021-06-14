#!/bin/bash -xe
set +o pipefail

wait_for_checks_to_complete(){
  community_operators_branch=$1
  all_tests_passed=false
  gh_pr_checks_command=(gh pr checks $community_operators_branch --repo $github_csiblock_community_operators_repository)
  export repo_pr=`gh pr list --repo $github_csiblock_community_operators_repository | grep $community_operators_branch`
  if [[ "$repo_pr" == *"$community_operators_branch"* ]]; then
    sleep 5
    while [ `"${gh_pr_checks_command[@]}" | grep -i pending | wc -l` -eq 0 ]; do
      sleep 15
    done
    summary_counter=0
    seconds_to_wait_for_summary_test=20
    while [ "$test_summary" != *"pending"* ] && [ $summary_counter -lt $seconds_to_wait_for_summary_test ] && [ $all_tests_passed == "false" ]; do
      sleep 1
      export test_summary=`"${gh_pr_checks_command[@]}" | grep -i summary`
      ((summary_counter=summary_counter+1))
      if [[ "$test_summary" == *"pass"* ]]; then
        all_tests_passed=true
      fi
    done
    if [[ $all_tests_passed == "false" ]]; then
      while [ `"${gh_pr_checks_command[@]}" | grep -i pending | wc -l` -eq 0 ]; do
        sleep 5
      done
    fi
  fi
}

wait_for_checks_to_complete $community_operators_kubernetes_branch
wait_for_checks_to_complete $community_operators_openshift_branch
