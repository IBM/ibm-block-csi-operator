#!/bin/bash -xel
set +o pipefail

edit_operator_image_in_csv_yaml_file (){
  cd $(dirname $csv_file)
  chmod 547 $(basename $csv_file) 
  declare -a operator_image_fields=(
      ".spec.install.spec.deployments[0].spec.template.spec.containers[0].image"
      ".metadata.annotations.containerImage"
      ".spec.relatedImages[0].image"
  )
  for image_field in "${operator_image_fields[@]}"
  do
      yq eval "$image_field |= env(operator_image_for_test)" $(basename $csv_file) -i
  done
cd -
}

create_demo_pr(){
  community_operators_branch=$1
  dest_path=$2
  cluster_kind=$3
  forked_repository=$4
  cd $forked_repository-fork
  repo_pr="gh pr list --repo $forked_repository | grep $community_operators_branch"
  if [[ "`eval $repo_pr`" == *"$community_operators_branch"* ]]; then
    gh pr close $community_operators_branch --delete-branch --repo $forked_repository
  fi
  git checkout master
  git checkout -b $community_operators_branch
  yes | cp -r $repository_path/deploy/olm-catalog/ibm-block-csi-operator-community/ $dest_path
  git add .
  git commit --signoff -m "build number $github_build_number $cluster_kind"
  git push origin $community_operators_branch
  gh pr create --title "IBM Block CSI update $cluster_kind" --repo $forked_repository --base master --head $community_operators_branch --body "pr check"
  cd ..
}

update_community_operators_fork (){
  forked_repository=$1
  original_repository=$2
  echo $github_token > github_token.txt
  gh auth login --with-token < github_token.txt
  gh repo fork $original_repository --clone $forked_repository-fork
  cd $forked_repository-fork
  git remote set-url origin https://csiblock:$github_token@github.com/$forked_repository.git
  git fetch upstream
  git rebase upstream/master
  git push origin master --force
  cd ..
}

edit_operator_image_in_csv_yaml_file
update_community_operators_fork $forked_community_operators_repository $original_community_operators_repository
update_community_operators_fork $forked_community_operators_repository_prod $original_community_operators_repository_prod
create_demo_pr $community_operators_kubernetes_branch "operators/" "kubernetes" $forked_community_operators_repository
create_demo_pr $community_operators_openshift_branch "operators/" "openshift" $forked_community_operators_repository_prod
