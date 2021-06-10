#!/bin/bash -xe
set +o pipefail

# install gh command
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt-get update
sudo apt-get install gh

echo $github_token > github_token.txt
gh auth login --with-token < github_token.txt
gh repo fork operator-framework/community-operators --clone community-operators-fork
cd community-operators-fork
git remote set-url origin https://csiblock:$github_token@github.com/csiblock/community-operators.git
git fetch upstream
git rebase upstream/master
git push origin master --force
git config --global user.email csi.block1@il.ibm.com
git config --global user.name csiblock

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
