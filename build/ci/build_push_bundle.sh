#!/bin/bash -xe

#
# Copyright 2026 IBM Corp.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Script to build and push CSI operator bundle image

# Environment variables with defaults
BUNDLE_VERSION="${BUNDLE_VERSION:-1.14.0}"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-quay.io/ibmcsiblock}"
BUNDLE_IMAGE_NAME="${BUNDLE_IMAGE_NAME:-ibm-block-csi-operator-bundle}"
OPERATOR_PACKAGE_NAME="${OPERATOR_PACKAGE_NAME:-ibm-block-csi-operator}"
OPERATOR_CHANNELS="${OPERATOR_CHANNELS:-stable-v1.14.0}"
BUILD_NUMBER="${BUILD_NUMBER:-1}"
GIT_BRANCH="${GIT_BRANCH:-develop}"
OVERRIDE_CSV="${OVERRIDE_CSV:-false}"
PUSH_IMAGE="${PUSH_IMAGE:-false}"
BUNDLE_DIR="${BUNDLE_DIR:-deploy/olm-catalog/ibm-block-csi-operator/${BUNDLE_VERSION}}"
IMAGE_REGISTRY="${IMAGE_REGISTRY:-${DOCKER_REGISTRY}}"

# Bundle image configuration - use only version as tag
bundle_registry="${DOCKER_REGISTRY}/${BUNDLE_IMAGE_NAME}"
bundle_tag_version="${bundle_registry}:${BUNDLE_VERSION}"

# Function to override CSV file for development registry
override_csi_csv_file() {
    if [ "$OVERRIDE_CSV" = "true" ]; then
        echo "Overriding CSV file for development registry..."
        
        CSV_FILE=$(find "${BUNDLE_DIR}/manifests" -name "*.clusterserviceversion.yaml" | head -n 1)
        
        if [ -z "$CSV_FILE" ]; then
            echo "Warning: CSV file not found in ${BUNDLE_DIR}/manifests"
            return 1
        fi
        
        echo "Found CSV file: $CSV_FILE"
        
        # Override operator image registry
        if [ -n "$OPERATOR_IMAGE_TAG" ]; then
            sed -i.bak "s|registry.connect.redhat.com/ibm/ibm-block-csi-operator:.*|${IMAGE_REGISTRY}/ibm-block-csi-operator:${OPERATOR_IMAGE_TAG}|g" "$CSV_FILE"
        fi
        
        # Override CSI driver images
        sed -i.bak "s|quay.io/ibmcsiblock/ibm-block-csi-driver-controller|${IMAGE_REGISTRY}/ibm-block-csi-driver-controller|g" "$CSV_FILE"
        sed -i.bak "s|quay.io/ibmcsiblock/ibm-block-csi-driver-node|${IMAGE_REGISTRY}/ibm-block-csi-driver-node|g" "$CSV_FILE"
        sed -i.bak "s|quay.io/ibmcsiblock/ibm-block-csi-host-definer|${IMAGE_REGISTRY}/ibm-block-csi-host-definer|g" "$CSV_FILE"
        
        # Remove backup file
        rm -f "${CSV_FILE}.bak"
        
        echo "CSV file override completed"
    else
        echo "Skipping CSV file override (OVERRIDE_CSV=${OVERRIDE_CSV})"
    fi
}

# Function to build and push bundle image
build_push_bundle_image() {
    local bundle_image=$1
    local bundle_directory=$2
    
    echo "=========================================="
    echo "Building Operator bundle image: ${bundle_image}"
    echo "Bundle directory: ${bundle_directory}"
    echo "Package name: ${OPERATOR_PACKAGE_NAME}"
    echo "Channels: ${OPERATOR_CHANNELS}"
    echo "=========================================="
    
    # Verify bundle directory exists
    if [ ! -d "$bundle_directory" ]; then
        echo "Error: Bundle directory does not exist: $bundle_directory"
        exit 1
    fi
    
    # Check if manifests and metadata directories exist
    if [ ! -d "${bundle_directory}/manifests" ] || [ ! -d "${bundle_directory}/metadata" ]; then
        echo "Error: Bundle directory must contain 'manifests' and 'metadata' subdirectories"
        exit 1
    fi
    
    # Build the bundle image using docker
    echo "Building bundle image with docker..."
    docker build -t "${bundle_image}" -f "${bundle_directory}/bundle-${BUNDLE_VERSION}.Dockerfile" "${bundle_directory}"
    
    if [ "$PUSH_IMAGE" = "true" ]; then
        echo "Pushing Operator bundle image to registry: ${bundle_image}"
        docker push "${bundle_image}"
        echo "Bundle image built and pushed successfully!"
    else
        echo "Bundle image built successfully (not pushed, PUSH_IMAGE=${PUSH_IMAGE})"
    fi
}

# Main execution
echo "=========================================="
echo "CSI Operator Bundle Build Script"
echo "=========================================="
echo "Bundle Version: ${BUNDLE_VERSION}"
echo "Docker Registry: ${DOCKER_REGISTRY}"
echo "Bundle Directory: ${BUNDLE_DIR}"
echo "Push Image: ${PUSH_IMAGE}"
echo "=========================================="

# Override CSV file if requested
override_csi_csv_file

# Build and optionally push bundle image
build_push_bundle_image "${bundle_tag_version}" "${BUNDLE_DIR}"

set +x
echo ""
echo "=========================================="
echo "Bundle image ready:"
echo "   ${bundle_tag_version}"
if [ "$PUSH_IMAGE" = "true" ]; then
    echo "   (pushed to registry)"
else
    echo "   (local only - not pushed)"
fi
echo "=========================================="

# If param $1 is given, write the tag to that file
[ -n "$1" ] && printf "${bundle_tag_version}" > "$1" || :

exit 0

# Made with Bob
