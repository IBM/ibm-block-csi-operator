#!/bin/bash -xe
set +o pipefail

export failed_k8s_checks=false
export failed_openshift_checks=false

export repo_pr=`gh pr list --repo $github_csiblock_community_operators_repository | grep $community_operators_kubernetes_branch`
if [[ "$repo_pr" == *"$community_operators_kubernetes_branch"* ]]; then
  sleep 5
  while [ "$(gh pr checks $community_operators_kubernetes_branch --repo $github_csiblock_community_operators_repository | grep -i pending | wc -l)" -gt 0 ]; do
    sleep 5
    echo "there are still checks that are running..."
  done
  if [ "$(gh pr checks $community_operators_kubernetes_branch --repo $github_csiblock_community_operators_repository | grep -i fail | wc -l)" -gt 0 ]
  then
    export failed_k8s_checks=true
  fi
  echo "The k8s checks:"
  gh pr checks $community_operators_kubernetes_branch --repo $github_csiblock_community_operators_repository
  gh pr close $community_operators_kubernetes_branch --delete-branch --repo $github_csiblock_community_operators_repository
fi

export repo_pr=`gh pr list --repo $github_csiblock_community_operators_repository | grep $community_operators_openshift_branch`
if [[ "$repo_pr" == *"$community_operators_openshift_branch"* ]]; then
  sleep 5
  while [ "$(gh pr checks $community_operators_openshift_branch --repo $github_csiblock_community_operators_repository | grep -i pending | wc -l)" -gt 0 ]; do
    sleep 5
    echo "there are still checks that are running..."
  done
  if [ "$(gh pr checks $community_operators_openshift_branch --repo $github_csiblock_community_operators_repository | grep -i fail | wc -l)" -gt 0 ]
  then
    export failed_openshift_checks=true
  fi
  echo "The Openshift checks:"
  gh pr checks $community_operators_openshift_branch --repo $github_csiblock_community_operators_repository
  gh pr close $community_operators_openshift_branch --delete-branch --repo $github_csiblock_community_operators_repository
fi

if [ $failed_k8s_checks == "true" ] || [ $failed_openshift_checks == "true" ]
then
  exit 1
fi
