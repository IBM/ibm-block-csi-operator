#!/bin/bash -xe
set +o pipefail

print_checks_and_delete_pr(){
  community_operators_branch=$1
  cluster_kind=$2
  repo_pr=`gh pr list --repo $forked_community_operators_repository | grep $community_operators_branch`
  if [[ "`eval $repo_pr`" == *"$community_operators_branch"* ]]; then
    echo "The $cluster_kind checks:"
    gh pr checks $community_operators_branch --repo $forked_community_operators_repository || true
    gh pr close $community_operators_branch --delete-branch --repo $forked_community_operators_repository
  fi
}

print_checks_and_delete_pr $community_operators_kubernetes_branch 'kubernetes'
print_checks_and_delete_pr $community_operators_openshift_branch 'openshift'
