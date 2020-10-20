# Validations
MANDATORY_ENVS="IMAGE_VERSION BUILD_NUMBER DOCKER_REGISTRY NODE_IMAGE CONTROLLER_IMAGE GIT_BRANCH"
for envi in $MANDATORY_ENVS; do 
    [ -z "${!envi}" ] && { echo "Error - Env $envi is mandatory for the script."; exit 1; } || :
done