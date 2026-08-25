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
#     export UPGRADE_CSI_VOLUME_GROUP_IMAGE="quay.io/csiblock/csi-volume-group-operator:1.13.2_b1825_origin.release-1.13.2"
#     export UPGRADE_CSI_VOLUMEREPLICATION_IMAGE="quay.io/csiblock/csi-block-volumereplication-operator:1.13.2_b1825_origin.release-1.13.2"
#     ./build/ci/upgrade-tests/modify_upgrade_images.sh
#
# Environment Variables:
#   UPGRADE_CSI_CONTROLLER_IMAGE      - Full image path for controller (repository:tag)
#   UPGRADE_CSI_NODE_IMAGE            - Full image path for node (repository:tag)
#   UPGRADE_CSI_HOST_DEFINITION_IMAGE - Full image path for host definer (repository:tag)
#   UPGRADE_CSI_VOLUME_GROUP_IMAGE       - Full image path for csi-volume-group sidecar (repository:tag)
#   UPGRADE_CSI_VOLUMEREPLICATION_IMAGE  - Full image path for csi-block-volumereplication sidecar (repository:tag)
#
# Modified Files:
#   - config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml
#   - config/samples/csi_v1_hostdefiner_cr.yaml
#   - pkg/config/settings.go
#
################################################################################

set -e

# Show usage (as defined in lines 3-34) if --help is provided
if [[ "$1" == "--help" ]] || [[ "$1" == "-h" ]]; then
    sed -n '3,34p' "$0" | sed 's/^# //' | sed 's/^#//' | sed 's/^#\+$//'
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
if [ -z "${UPGRADE_CSI_CONTROLLER_IMAGE}" ] && [ -z "${UPGRADE_CSI_NODE_IMAGE}" ] && [ -z "${UPGRADE_CSI_HOST_DEFINITION_IMAGE}" ] && [ -z "${UPGRADE_CSI_VOLUME_GROUP_IMAGE}" ] && [ -z "${UPGRADE_CSI_VOLUMEREPLICATION_IMAGE}" ]; then
    echo "ERROR: No upgrade images provided!"
    echo "Please set at least one of the following environment variables:"
    echo "  - UPGRADE_CSI_CONTROLLER_IMAGE"
    echo "  - UPGRADE_CSI_NODE_IMAGE"
    echo "  - UPGRADE_CSI_HOST_DEFINITION_IMAGE"
    echo "  - UPGRADE_CSI_VOLUME_GROUP_IMAGE"
    echo "  - UPGRADE_CSI_VOLUMEREPLICATION_IMAGE"
    echo ""
    echo "Run with --help for usage information"
    exit 1
fi

echo "Input Images:"
echo "  Upgrade Controller Image: ${UPGRADE_CSI_CONTROLLER_IMAGE:-<not set>}"
echo "  Upgrade Node Image: ${UPGRADE_CSI_NODE_IMAGE:-<not set>}"
echo "  Upgrade Host Definer Image: ${UPGRADE_CSI_HOST_DEFINITION_IMAGE:-<not set>}"
echo "  Upgrade CSI Volume Group Image: ${UPGRADE_CSI_VOLUME_GROUP_IMAGE:-<not set>}"
echo "  Upgrade CSI Volumereplication Image: ${UPGRADE_CSI_VOLUMEREPLICATION_IMAGE:-<not set>}"

# Extract repository and tag for each image
UPGRADE_CSI_CONTROLLER_REPO=$(extract_repository "${UPGRADE_CSI_CONTROLLER_IMAGE}")
UPGRADE_CSI_CONTROLLER_TAG=$(extract_tag "${UPGRADE_CSI_CONTROLLER_IMAGE}")

UPGRADE_CSI_NODE_REPO=$(extract_repository "${UPGRADE_CSI_NODE_IMAGE}")
UPGRADE_CSI_NODE_TAG=$(extract_tag "${UPGRADE_CSI_NODE_IMAGE}")

UPGRADE_CSI_HOST_DEFINER_REPO=$(extract_repository "${UPGRADE_CSI_HOST_DEFINITION_IMAGE}")
UPGRADE_CSI_HOST_DEFINER_TAG=$(extract_tag "${UPGRADE_CSI_HOST_DEFINITION_IMAGE}")

UPGRADE_CSI_VOLUME_GROUP_REPO=$(extract_repository "${UPGRADE_CSI_VOLUME_GROUP_IMAGE}")
UPGRADE_CSI_VOLUME_GROUP_TAG=$(extract_tag "${UPGRADE_CSI_VOLUME_GROUP_IMAGE}")

UPGRADE_CSI_VOLUMEREPLICATION_REPO=$(extract_repository "${UPGRADE_CSI_VOLUMEREPLICATION_IMAGE}")
UPGRADE_CSI_VOLUMEREPLICATION_TAG=$(extract_tag "${UPGRADE_CSI_VOLUMEREPLICATION_IMAGE}")

# Collect all unique custom registry usernames from all provided images,
# skipping registries already present in pkg/config/settings.go
UPGRADE_CUSTOM_REGISTRY_USERNAMES=()
for _repo in "${UPGRADE_CSI_CONTROLLER_REPO}" "${UPGRADE_CSI_NODE_REPO}" "${UPGRADE_CSI_HOST_DEFINER_REPO}" "${UPGRADE_CSI_VOLUME_GROUP_REPO}" "${UPGRADE_CSI_VOLUMEREPLICATION_REPO}"; do
    if [ -z "${_repo}" ]; then
        continue
    fi
    _reg=$(extract_registry_username "${_repo}")
    # Skip if already present in settings.go or already collected
    if grep -q "\"${_reg}\"" pkg/config/settings.go; then
        continue
    fi
    _already=false
    for _existing in "${UPGRADE_CUSTOM_REGISTRY_USERNAMES[@]}"; do
        if [ "${_existing}" = "${_reg}" ]; then
            _already=true
            break
        fi
    done
    if [ "${_already}" = false ]; then
        UPGRADE_CUSTOM_REGISTRY_USERNAMES+=("${_reg}")
    fi
done

echo ""
echo "Parsed values:"
echo "Controller Repository: ${UPGRADE_CSI_CONTROLLER_REPO}"
echo "Controller Tag: ${UPGRADE_CSI_CONTROLLER_TAG}"
echo "Node Repository: ${UPGRADE_CSI_NODE_REPO}"
echo "Node Tag: ${UPGRADE_CSI_NODE_TAG}"
echo "Host Definer Repository: ${UPGRADE_CSI_HOST_DEFINER_REPO}"
echo "Host Definer Tag: ${UPGRADE_CSI_HOST_DEFINER_TAG}"
echo "CSI Volume Group Repository: ${UPGRADE_CSI_VOLUME_GROUP_REPO}"
echo "CSI Volume Group Tag: ${UPGRADE_CSI_VOLUME_GROUP_TAG}"
echo "CSI Volumereplication Repository: ${UPGRADE_CSI_VOLUMEREPLICATION_REPO}"
echo "CSI Volumereplication Tag: ${UPGRADE_CSI_VOLUMEREPLICATION_TAG}"
echo "Custom Registry Usernames: ${UPGRADE_CUSTOM_REGISTRY_USERNAMES[*]:-<none>}"
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

if [ -n "${UPGRADE_CSI_VOLUMEREPLICATION_REPO}" ] && [ -n "${UPGRADE_CSI_VOLUMEREPLICATION_TAG}" ]; then
    echo "Modifying config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml for csi-addons-replicator sidecar image..."
    sed -i "s|repository: quay.io/.*csi-block-volumereplication-operator.*|repository: ${UPGRADE_CSI_VOLUMEREPLICATION_REPO}|g" config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml
    sed -i "/csi-addons-replicator/,/tag:/ s|tag: \".*\"|tag: \"${UPGRADE_CSI_VOLUMEREPLICATION_TAG}\"|" config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml
fi

if [ -n "${UPGRADE_CSI_VOLUME_GROUP_REPO}" ] && [ -n "${UPGRADE_CSI_VOLUME_GROUP_TAG}" ]; then
    echo "Modifying config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml for csi-volume-group sidecar image..."
    sed -i "s|repository: quay.io/.*csi-volume-group-operator.*|repository: ${UPGRADE_CSI_VOLUME_GROUP_REPO}|g" config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml
    sed -i "/csi-volume-group/,/tag:/ s|tag: \".*\"|tag: \"${UPGRADE_CSI_VOLUME_GROUP_TAG}\"|" config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml
fi

# Modify csi_v1_hostdefiner_cr.yaml
if [ -n "${UPGRADE_CSI_HOST_DEFINER_REPO}" ] && [ -n "${UPGRADE_CSI_HOST_DEFINER_TAG}" ]; then
    echo "Modifying config/samples/csi_v1_hostdefiner_cr.yaml for host definer image..."
    sed -i "s|repository: quay.io/.*ibm-block-csi-host-definer.*|repository: ${UPGRADE_CSI_HOST_DEFINER_REPO}|g" config/samples/csi_v1_hostdefiner_cr.yaml
    sed -i "/hostDefiner:/,/tag:/ s|tag: \".*\"|tag: \"${UPGRADE_CSI_HOST_DEFINER_TAG}\"|" config/samples/csi_v1_hostdefiner_cr.yaml
fi

# Modify pkg/config/settings.go to add all custom registry usernames
if [ "${#UPGRADE_CUSTOM_REGISTRY_USERNAMES[@]}" -gt 0 ]; then
    echo "Modifying pkg/config/settings.go to add custom registry usernames: ${UPGRADE_CUSTOM_REGISTRY_USERNAMES[*]}..."

    for _reg in "${UPGRADE_CUSTOM_REGISTRY_USERNAMES[@]}"; do
        if ! grep -q "\"${_reg}\"" pkg/config/settings.go; then
            # Escape dots for sed (slashes are safe inside | delimiters)
            _reg_escaped=$(echo "${_reg}" | sed 's/\./\\./g')
            # Append before the closing paren on the last line of sets.NewString(...)
            # Match the line that contains RedHatRegistryUsername and replace its trailing )
            sed -i "/RedHatRegistryUsername/ s|)$|, \"${_reg_escaped}\")|" pkg/config/settings.go
        fi
    done
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
if [ -n "${UPGRADE_CSI_CONTROLLER_IMAGE}" ] || [ -n "${UPGRADE_CSI_NODE_IMAGE}" ] || [ -n "${UPGRADE_CSI_VOLUME_GROUP_IMAGE}" ] || [ -n "${UPGRADE_CSI_VOLUMEREPLICATION_IMAGE}" ]; then
    show_diff "config/samples/csi.ibm.com_v1_ibmblockcsi_cr.yaml" "${TEMP_DIR}/csi.ibm.com_v1_ibmblockcsi_cr.yaml.orig"
fi

if [ -n "${UPGRADE_CSI_HOST_DEFINITION_IMAGE}" ]; then
    show_diff "config/samples/csi_v1_hostdefiner_cr.yaml" "${TEMP_DIR}/csi_v1_hostdefiner_cr.yaml.orig"
fi

if [ "${#UPGRADE_CUSTOM_REGISTRY_USERNAMES[@]}" -gt 0 ]; then
    show_diff "pkg/config/settings.go" "${TEMP_DIR}/settings.go.orig"
fi

# Cleanup temp directory
rm -rf "${TEMP_DIR}"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ All modifications applied successfully!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Made with Bob
