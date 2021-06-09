#!/bin/bash -xe
set +o pipefail

sed -i "s+$current_operator_image+$wanted_operator_image+g" $csv_file

# install gh command
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install gh

echo $github_token > mytoken.txt
gh auth login --with-token < mytoken.txt
gh repo fork operator-framework/community-operators --clone community-operators-fork
cd community-operators-fork
git remote set-url origin https://csiblock:$github_token@github.com/csiblock/community-operators.git
git fetch upstream
git rebase upstream/master
git push origin master --force
git config --global user.email csi.block1@il.ibm.com
git config --global user.name csiblock

export repo_pr=`gh pr list --repo $github_csiblock_community_operators_repository | grep $community_operators_kubernetes_branch`
if [[ "$repo_pr" == *"$community_operators_kubernetes_branch"* ]]; then
  gh pr close $community_operators_kubernetes_branch --delete-branch --repo $github_csiblock_community_operators_repository
fi
git checkout -b $community_operators_kubernetes_branch
yes | cp -r $repository_path/deploy/olm-catalog/ibm-block-csi-operator-community/ $repository_path/community-operators-fork/upstream-community-operators/
git add .
git commit --signoff -m "build number $github_build_number kubernetes"
git push origin $community_operators_kubernetes_branch
gh pr create --title "IBM Block CSI update to v1.6.0 kubernetes" --repo $github_csiblock_community_operators_repository --base master --head $community_operators_kubernetes_branch --body "pr check"

export repo_pr=`gh pr list --repo $github_csiblock_community_operators_repository | grep $community_operators_openshift_branch`
if [[ "$repo_pr" == *"$community_operators_openshift_branch"* ]]; then
  gh pr close $community_operators_openshift_branch --delete-branch --repo $github_csiblock_community_operators_repository
fi
git checkout master
git checkout -b $community_operators_openshift_branch
yes | cp -r $repository_path/deploy/olm-catalog/ibm-block-csi-operator-community/ $repository_path/community-operators-fork/community-operators/
git add .
git commit --signoff -m "build number $github_build_number openshift"
git push origin $community_operators_openshift_branch
gh pr create --title "IBM Block CSI update to v1.6.0 operator" --repo $github_csiblock_community_operators_repository --base master --head $community_operators_openshift_branch --body "pr check"
