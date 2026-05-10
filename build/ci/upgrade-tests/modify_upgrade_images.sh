#!/bin/bash

################################################################################
# Script: modify_upgrade_images.sh
#
# Description: Modify operator code with upgrade images for testing
#              Updates CR YAML files and settings.go to point to custom upgrade images
#
# Usage:
#   In Jenkins (automatic):
#     ./build/ci/upgrade-tests/modify_upgrade_images.sh
#     (Uses environment variables set by Jenkins parameters)
#
#   Local execution (manual):
#     export UPGRADE_CSI_CONTROLLER_IMAGE="quay.io/csiblock/ibm-block-csi-driver-controller-amd64:1.13.2_b1825_origin.release-1.13.2"
#     export UPGRADE_CSI_NODE_IMAGE="quay.io/csiblock/ibm-block-csi-driver-node-amd64:1.13.2_b1825_origin.release-1.13.2"
#     export UPGRADE_CSI_HOST_DEFINITION_IMAGE="quay.io/csiblock/ibm-block-csi-host-definer-amd64:1.13.2_b1825_origin.release-1.13.2"
#     ./build/ci/upgrade-tests/modify_upgrade_images.sh
#
# Environment Variables:
#   UPGRADE_CSI_CONTROLLER_IMAGE      - Full image path for controller (repository:tag)
#   UPGRADE_CSI_NODE_IMAGE            - Full image path for node (repository:tag)
#   UPGRADE_CSI_HOST_DEFINITION_IMAGE - Full image path for host definer (repository:tag)
#
# Modified Files:
#   - config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml
#   - config/samples/csi_v1_hostdefiner_cr.yaml
#   - pkg/config/settings.go
#
################################################################################

set -e

# Show usage (as defined in lines 3-29) if --help is provided
if [[ "$1" == "--help" ]] || [[ "$1" == "-h" ]]; then
    sed -n '3,29p' "$0" | sed 's/^# //' | sed 's/^#//' | sed 's/^#\+$//'
    exit 0
fi

echo "=========================================="
echo "Modifying operator code for upgrade images"
echo "=========================================="

extract_repository() {
    local image=$1
    if [ -z "$image" ]; then
        echo ""
        return
    fi
    # Extract repository (everything before the last colon)
    echo "${image%:*}"
}

extract_tag() {
    local image=$1
    if [ -z "$image" ]; then
        echo ""
        return
    fi
    # Extract tag (everything after the last colon)
    echo "${image##*:}"
}

extract_registry_username() {
    local repository=$1
    if [ -z "$repository" ]; then
        echo ""
        return
    fi
    # Extract registry username (e.g., quay.io/csiblock from quay.io/csiblock/image-name)
    echo "${repository%/*}"
}

# Validate that at least one image is provided
if [ -z "${UPGRADE_CSI_CONTROLLER_IMAGE}" ] && [ -z "${UPGRADE_CSI_NODE_IMAGE}" ] && [ -z "${UPGRADE_CSI_HOST_DEFINITION_IMAGE}" ]; then
    echo "ERROR: No upgrade images provided!"
    echo "Please set at least one of the following environment variables:"
    echo "  - UPGRADE_CSI_CONTROLLER_IMAGE"
    echo "  - UPGRADE_CSI_NODE_IMAGE"
    echo "  - UPGRADE_CSI_HOST_DEFINITION_IMAGE"
    echo ""
    echo "Run with --help for usage information"
    exit 1
fi

echo "Input Images:"
echo "  Upgrade Controller Image: ${UPGRADE_CSI_CONTROLLER_IMAGE:-<not set>}"
echo "  Upgrade Node Image: ${UPGRADE_CSI_NODE_IMAGE:-<not set>}"
echo "  Upgrade Host Definer Image: ${UPGRADE_CSI_HOST_DEFINITION_IMAGE:-<not set>}"

# Extract repository and tag for each image
UPGRADE_CSI_CONTROLLER_REPO=$(extract_repository "${UPGRADE_CSI_CONTROLLER_IMAGE}")
UPGRADE_CSI_CONTROLLER_TAG=$(extract_tag "${UPGRADE_CSI_CONTROLLER_IMAGE}")

UPGRADE_CSI_NODE_REPO=$(extract_repository "${UPGRADE_CSI_NODE_IMAGE}")
UPGRADE_CSI_NODE_TAG=$(extract_tag "${UPGRADE_CSI_NODE_IMAGE}")

UPGRADE_CSI_HOST_DEFINER_REPO=$(extract_repository "${UPGRADE_CSI_HOST_DEFINITION_IMAGE}")
UPGRADE_CSI_HOST_DEFINER_TAG=$(extract_tag "${UPGRADE_CSI_HOST_DEFINITION_IMAGE}")

# Extract custom registry username
UPGRADE_CUSTOM_REGISTRY_USERNAME=$(extract_registry_username "${UPGRADE_CSI_CONTROLLER_REPO}")

echo ""
echo "Parsed values:"
echo "Controller Repository: ${UPGRADE_CSI_CONTROLLER_REPO}"
echo "Controller Tag: ${UPGRADE_CSI_CONTROLLER_TAG}"
echo "Node Repository: ${UPGRADE_CSI_NODE_REPO}"
echo "Node Tag: ${UPGRADE_CSI_NODE_TAG}"
echo "Host Definer Repository: ${UPGRADE_CSI_HOST_DEFINER_REPO}"
echo "Host Definer Tag: ${UPGRADE_CSI_HOST_DEFINER_TAG}"
echo "Custom Registry Username: ${UPGRADE_CUSTOM_REGISTRY_USERNAME}"
echo ""

# Store original files for diff comparison
TEMP_DIR=$(mktemp -d)
cp config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml "${TEMP_DIR}/csi.ibm.com_v1_ibmblockcsi_cr.yaml.orig" 2>/dev/null || true
cp config/samples/csi_v1_hostdefiner_cr.yaml "${TEMP_DIR}/csi_v1_hostdefiner_cr.yaml.orig" 2>/dev/null || true
cp pkg/config/settings.go "${TEMP_DIR}/settings.go.orig" 2>/dev/null || true

# Modify csi.ibm.com_v1_ibmblockcsi_cr.yaml
if [ -n "${UPGRADE_CSI_CONTROLLER_REPO}" ] && [ -n "${UPGRADE_CSI_CONTROLLER_TAG}" ]; then
    echo "Modifying config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml for controller image..."
    sed -i "s|repository: quay.io/.*ibm-block-csi-driver-controller.*|repository: ${UPGRADE_CSI_CONTROLLER_REPO}|g" config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml
    sed -i "/controller:/,/tag:/ s|tag: \".*\"|tag: \"${UPGRADE_CSI_CONTROLLER_TAG}\"|" config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml
fi

if [ -n "${UPGRADE_CSI_NODE_REPO}" ] && [ -n "${UPGRADE_CSI_NODE_TAG}" ]; then
    echo "Modifying config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml for node image..."
    sed -i "s|repository: quay.io/.*ibm-block-csi-driver-node.*|repository: ${UPGRADE_CSI_NODE_REPO}|g" config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml
    sed -i "/node:/,/tag:/ s|tag: \".*\"|tag: \"${UPGRADE_CSI_NODE_TAG}\"|" config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml
fi

# Modify csi_v1_hostdefiner_cr.yaml
if [ -n "${UPGRADE_CSI_HOST_DEFINER_REPO}" ] && [ -n "${UPGRADE_CSI_HOST_DEFINER_TAG}" ]; then
    echo "Modifying config/samples/csi_v1_hostdefiner_cr.yaml for host definer image..."
    sed -i "s|repository: quay.io/.*ibm-block-csi-host-definer.*|repository: ${UPGRADE_CSI_HOST_DEFINER_REPO}|g" config/samples/csi_v1_hostdefiner_cr.yaml
    sed -i "/hostDefiner:/,/tag:/ s|tag: \".*\"|tag: \"${UPGRADE_CSI_HOST_DEFINER_TAG}\"|" config/samples/csi_v1_hostdefiner_cr.yaml
fi

# Modify pkg/config/settings.go to add custom registry username
if [ -n "${UPGRADE_CUSTOM_REGISTRY_USERNAME}" ]; then
    echo "Modifying pkg/config/settings.go to add custom registry username..."
    
    # Check if the custom registry username constant already exists
    if ! grep -q "QuayCSIBlockCustomRegistryUsername" pkg/config/settings.go; then
        # Add the constant after QuayCSIBlockRegistryUsername
        sed -i "/QuayCSIBlockRegistryUsername = \"quay.io\/ibmcsiblock\"/a\\
\\tQuayCSIBlockCustomRegistryUsername = \"${UPGRADE_CUSTOM_REGISTRY_USERNAME}\"" pkg/config/settings.go
        
        # Add to OfficialRegistriesUsernames
        sed -i "s|RedHatRegistryUsername)|QuayCSIBlockCustomRegistryUsername, RedHatRegistryUsername)|g" pkg/config/settings.go
    else
        # Update existing constant
        sed -i "s|QuayCSIBlockCustomRegistryUsername = \".*\"|QuayCSIBlockCustomRegistryUsername = \"${UPGRADE_CUSTOM_REGISTRY_USERNAME}\"|g" pkg/config/settings.go
    fi
fi

echo ""
echo "=========================================="
echo "Modifications completed successfully!"
echo "=========================================="
echo ""

# Display git-style diff of changes
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║                           CHANGES SUMMARY (git diff)                       ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

# Function to show colored diff
show_diff() {
    local file=$1
    local orig_file=$2
    
    if [ -f "${orig_file}" ]; then
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "📄 File: ${file}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        
        # Check if git is available and we're in a git repo
        if command -v git &> /dev/null && git rev-parse --git-dir > /dev/null 2>&1; then
            # Use git diff with color if available
            git diff --no-index --color=always "${orig_file}" "${file}" 2>/dev/null | tail -n +5 || \
            diff -u "${orig_file}" "${file}" | tail -n +3 | sed 's/^-/\x1b[31m-/; s/^+/\x1b[32m+/; s/^@/\x1b[36m@/; s/$/\x1b[0m/'
        else
            # Fallback to regular diff with manual coloring
            diff -u "${orig_file}" "${file}" 2>/dev/null | tail -n +3 | sed 's/^-/\x1b[31m-/; s/^+/\x1b[32m+/; s/^@/\x1b[36m@/; s/$/\x1b[0m/' || echo "  (no changes)"
        fi
        echo ""
    fi
}

# Show diffs for each modified file
if [ -n "${UPGRADE_CSI_CONTROLLER_IMAGE}" ] || [ -n "${UPGRADE_CSI_NODE_IMAGE}" ]; then
    show_diff "config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml" "${TEMP_DIR}/csi.ibm.com_v1_ibmblockcsi_cr.yaml.orig"
fi

if [ -n "${UPGRADE_CSI_HOST_DEFINITION_IMAGE}" ]; then
    show_diff "config/samples/csi_v1_hostdefiner_cr.yaml" "${TEMP_DIR}/csi_v1_hostdefiner_cr.yaml.orig"
fi

if [ -n "${UPGRADE_CUSTOM_REGISTRY_USERNAME}" ]; then
    show_diff "pkg/config/settings.go" "${TEMP_DIR}/settings.go.orig"
fi

# Cleanup temp directory
rm -rf "${TEMP_DIR}"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ All modifications applied successfully!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
