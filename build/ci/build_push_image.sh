#!/bin/bash -xe

# Validations
MANDATORY_ENVS="IMAGE_VERSION BUILD_NUMBER DOCKER_REGISTRY OPERATOR_IMAGE GIT_BRANCH"
for envi in $MANDATORY_ENVS; do 
    [ -z "${!envi}" ] && { echo "Error - Env $envi is mandatory for the script."; exit 1; } || :
done

NODE_IMAGE=ibm-node-agent

# Prepare specific tag for the image
tags=`build/ci/get_image_tags_from_branch.sh ${GIT_BRANCH} ${IMAGE_VERSION} ${BUILD_NUMBER} ${GIT_COMMIT:0:7}`
specific_tag=`echo $tags | awk '{print$1}'`

# Set latest tag only if its from develop branch or master and prepare tags
[ "$GIT_BRANCH" = "develop" -o "$GIT_BRANCH" = "origin/develop" -o "$GIT_BRANCH" = "master" ] && tag_latest="true" || tag_latest="false"


# Operator
# --------------
operator_registry="${DOCKER_REGISTRY}/${OPERATOR_IMAGE}"
operator_tag_specific="${operator_registry}:${specific_tag}"
operator_tag_latest=${operator_registry}:latest
[ "$tag_latest" = "true" ] && taglatestflag="-t ${operator_tag_latest}" 

echo "Build and push the Operator image"
docker build -t ${operator_tag_specific} $taglatestflag -f build/Dockerfile.operator --build-arg VERSION="${IMAGE_VERSION}" --build-arg BUILD_NUMBER="${BUILD_NUMBER}" .
docker push ${operator_tag_specific}
[ "$tag_latest" = "true" ] && docker push ${operator_tag_latest} || :

# Node agent
# --------
node_registry="${DOCKER_REGISTRY}/${NODE_IMAGE}"
node_tag_specific="${node_registry}:${specific_tag}"
node_tag_latest=${node_registry}:latest
[ "$tag_latest" = "true" ] && taglatestflag="-t ${node_tag_latest}" 

echo "Build and push the node agent image"
docker build -t ${node_tag_specific} $taglatestflag -f build/Dockerfile.nodeagent .
docker push ${node_tag_specific}
[ "$tag_latest" = "true" ] && docker push ${node_tag_latest} || :


set +x
echo ""
echo "Image ready:"
echo "   ${operator_tag_specific}"
echo "   ${node_tag_specific}"
[ "$tag_latest" = "true" ] && { echo "   ${operator_tag_specific}"; echo "   ${node_tag_latest}"; } || :

# if param $1 given the script echo the specific tag
[ -n "$1" ] && printf "${operator_tag_specific}\n${node_tag_specific}\n" > $1 || :

