#!/bin/bash -xe
set +o pipefail

wait_for_checks_to_complete(){
  community_operators_branch=$1
  all_tests_passed=false
  gh_pr_checks_command="gh pr checks $community_operators_branch --repo $github_csiblock_community_operators_repository"
  export repo_pr=`gh pr list --repo $github_csiblock_community_operators_repository | grep $community_operators_branch`
  if [[ "$repo_pr" == *"$community_operators_branch"* ]]; then
    sleep 5
    while [ `eval gh_pr_checks_command | grep -i pending | wc -l` -gt 0 ]; do
      sleep 15
    done
    seconds_waited_for_summary_test=0
    seconds_to_wait_for_summary_test=20
    while [ "$test_summary" != *"pending"* ] && [ $seconds_waited_for_summary_test -lt $summary_test_timeout_seconds] && [ $all_tests_passed == "false" ]; do
      sleep 1
      test_summary=`eval gh_pr_checks_command | grep -i summary`
      ((seconds_waited_for_summary_test=seconds_waited_for_summary_test+1))
      if [[ "$test_summary" == *"pass"* ]]; then
        all_tests_passed=true
      fi
    done
    if [[ $all_tests_passed == "false" ]]; then
      while [ `eval gh_pr_checks_command | grep -i pending | wc -l` -gt 0 ]; do
        sleep 5
      done
    fi
  fi
}

wait_for_checks_to_complete $community_operators_kubernetes_branch
wait_for_checks_to_complete $community_operators_openshift_branch
