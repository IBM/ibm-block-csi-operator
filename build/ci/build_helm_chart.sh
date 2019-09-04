#!/bin/bash -e

set -x
set -e

# update_chart_version will add the build number to the chart version
# for example: 1.0.0 -> 1.0.0-648
function update_chart_version()
{
  if [ $PRODUCTION_BUILD != "yes" ]; then
    chart_file="$CHART_PATH/Chart.yaml"
    sed -i -r "s/^version: [0-9]+\.[0-9]+\.[0-9]+$/&-$BUILD_NUMBER/" $chart_file
  fi
}

# use right helm bindary according to the arch
function update_helm_executor()
{
    if [ "$(uname)" = "Darwin" ]; then
	    arch="Darwin"
    else
        arch=`uname -i`
    fi
	HELM="$HELM_PATH/helm-$arch"
}

function cleanup_helm()
{
	rm -rf $HELM_HOME
}

PRODUCTION_BUILD=yes
CHART_REPOSITORY="https://stg-artifactory.haifa.ibm.com/artifactory/chart-repo"

if [ -z $CHART_REPOSITORY ]; then
  echo "Warning: Set CHART_REPOSITORY if you want to build and upload Ubiquity helm chart!"
  exit 0
fi

CHART_REPOSITORY_NAME="artifactory"
CHART_FOLDER="artifactory-charts"
INDEX_PATH="$CHART_REPOSITORY/index.yaml"
CHART_NAME="ibm-block-csi-operator"

CURRENT_PATH=$(dirname "$BASH_SOURCE")
PROJECT_ROOT="$CURRENT_PATH/../.."
HELM_PATH=$CURRENT_PATH
BUILD_OUTPUT="$PROJECT_ROOT/build/_output"
export PATH=$PATH:$HELM_PATH
export HELM_HOME=/tmp/helm3
CHART_PATH="$PROJECT_ROOT/deploy/helm/$CHART_NAME"

# load artifactory info, like ci_user and ci_password
if [ -f site_vars ]; then
  . site_vars
fi

update_chart_version

update_helm_executor

cd $PROJECT_ROOT

# init helm
chmod +x $HELM
$HELM init

# add operator helm repo
$HELM repo add $CHART_REPOSITORY_NAME $CHART_REPOSITORY

mkdir -p "$BUILD_OUTPUT/$CHART_FOLDER"
# remove all contents if the path already exists
rm -rf "$BUILD_OUTPUT/$CHART_FOLDER/*"

# download index.yaml
wget $INDEX_PATH
mv index.yaml "$BUILD_OUTPUT/$CHART_FOLDER"

# package ubiquity helm chart
$HELM package "$CHART_PATH/"
CHART_NAME_TGZ=`ls $CHART_NAME*`
mv $CHART_NAME_TGZ "$BUILD_OUTPUT/$CHART_FOLDER"

# merge index.yaml
$HELM repo index --merge "$BUILD_OUTPUT/$CHART_FOLDER/index.yaml" --url $CHART_REPOSITORY "$BUILD_OUTPUT/$CHART_FOLDER"

# upload chart and new index
curl -u $ci_user:$ci_password -T "$BUILD_OUTPUT/$CHART_FOLDER/index.yaml" "$CHART_REPOSITORY/"
curl -u $ci_user:$ci_password -T "$BUILD_OUTPUT/$CHART_FOLDER/$CHART_NAME_TGZ" "$CHART_REPOSITORY/"

cleanup_helm
