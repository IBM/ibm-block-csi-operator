#!/bin/bash -xe
set +o pipefail

sleep 20

print_checks_and_delete_pr(){
  community_operators_branch=$1
  cluster_kind=$2
  export repo_pr=`gh pr list --repo $github_csiblock_community_operators_repository | grep $community_operators_branch`
  if [[ "$repo_pr" == *"$community_operators_branch"* ]]; then
    echo "The $cluster_kind checks:"
    gh pr checks $community_operators_branch --repo $github_csiblock_community_operators_repository
    gh pr close $community_operators_branch --delete-branch --repo $github_csiblock_community_operators_repository
  fi
}

print_checks_and_delete_pr $community_operators_kubernetes_branch 'kubernetes'
print_checks_and_delete_pr $community_operators_openshift_branch 'openshift'
