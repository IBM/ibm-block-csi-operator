#!/bin/bash -xe
set +o pipefail

wait_fot_checks_to_start(){
  gh_pr_checks_command=$1
  while [ `$gh_pr_checks_command | grep -i pending | wc -l` -eq 0 ]; do
    sleep 1
  done
}
wait_for_checks_to_complete(){
  community_operators_branch=$1
  all_tests_passed=false
  gh_pr_checks_command="gh pr checks $community_operators_branch --repo $forked_community_operators_repository"
  repo_pr=`gh pr list --repo $forked_community_operators_repository | grep $community_operators_branch`
  if [[ "$repo_pr" == *"$community_operators_branch"* ]]; then
    wait_fot_checks_to_start $gh_pr_checks_command
    test_summary="$gh_pr_checks_command | grep -i summary"
    while [[ ! "$test_summary" =~ "pass" ]] && [[ ! "$test_summary" =~ "fail" ]]; do
      sleep 1
    done
  fi
}

wait_for_checks_to_complete $community_operators_kubernetes_branch
wait_for_checks_to_complete $community_operators_openshift_branch
