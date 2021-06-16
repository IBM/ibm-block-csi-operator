#!/bin/bash -xe
set +o pipefail

yq -y --arg operator_image $operator_image_for_test '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image |= $operator_image' $csv_file > temp.yaml
cat temp.yaml > $csv_file
yq -y --arg operator_image $operator_image_for_test '.metadata.annotations.containerImage |= $operator_image' $csv_file > temp.yaml
cat temp.yaml > $csv_file
yq -y --arg operator_image $operator_image_for_test '.spec.relatedImages[0].image |= $operator_image' $csv_file > temp.yaml
cat temp.yaml > $csv_file

echo $github_token > github_token.txt
gh auth login --with-token < github_token.txt
gh repo fork operator-framework/community-operators --clone community-operators-fork
cd community-operators-fork
git remote set-url origin https://csiblock:$github_token@github.com/csiblock/community-operators.git
git fetch upstream
git rebase upstream/master
git push origin master --force

create_demo_pr(){
  community_operators_branch=$1
  dest_path=$2
  cluster_kind=$3  
  export repo_pr=`gh pr list --repo $github_csiblock_community_operators_repository | grep $community_operators_branch`
  if [[ "$repo_pr" == *"$community_operators_branch"* ]]; then
    gh pr close $community_operators_branch --delete-branch --repo $github_csiblock_community_operators_repository
  fi
  git checkout master
  git checkout -b $community_operators_branch
  yes | cp -r $repository_path/deploy/olm-catalog/ibm-block-csi-operator-community/ $dest_path
  git add .
  git commit --signoff -m "build number $github_build_number $cluster_kind"
  git push origin $community_operators_branch
  gh pr create --title "IBM Block CSI update $cluster_kind" --repo $github_csiblock_community_operators_repository --base master --head $community_operators_branch --body "pr check"
}

create_demo_pr $community_operators_kubernetes_branch "upstream-community-operators/" "kubernetes"
create_demo_pr $community_operators_openshift_branch "community-operators/" "openshift"
