#!/bin/bash -xe
set +o pipefail

gh_pr_checks_command (){
  community_operators_branch=$1
  forked_repository=$2
  gh pr checks $community_operators_branch --repo $forked_repository
}

wait_for_checks_to_start(){
  community_operators_branch=$1
  forked_repository=$2
  while [ `gh_pr_checks_command $community_operators_branch $forked_repository | wc -l` -eq 0 ]; do
    sleep 1
  done
}
wait_for_checks_to_complete(){
  community_operators_branch=$1
  forked_repository=$2
  all_tests_passed=false
  repo_pr=`gh pr list --repo $forked_repository | grep $community_operators_branch`
  if [[ "$repo_pr" == *"$community_operators_branch"* ]]; then
    wait_for_checks_to_start $community_operators_branch $forked_repository
    test_summary="gh_pr_checks_command $community_operators_branch  $forked_repository | grep -i summary"
    while [[ ! "`eval $test_summary`" =~ "pass" ]] && [[ ! "`eval $test_summary`" =~ "fail" ]]; do
      sleep 1
    done
  fi
}

wait_for_checks_to_complete $community_operators_kubernetes_branch $forked_community_operators_repository
wait_for_checks_to_complete $community_operators_openshift_branch $forked_community_operators_repository_prod
