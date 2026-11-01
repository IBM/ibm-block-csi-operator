#!/bin/bash
#
# Copyright IBM Corp.
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

# Prepares the current release bundle for submission to the operator catalog repositories.
# For each target the bundle dir is copied, the image tag in the CSV is replaced with an image digest,
# and the previous version is added to the "skips" section of the CSV.
#
#   certified  : deploy/olm-catalog/ibm-block-csi-operator/<version>
#                  -> <certified-operators>/operators/ibm-block-csi-operator/<version>
#                image registry.connect.redhat.com/ibm/ibm-block-csi-operator:<version> -> @sha256:<digest>
#                (digest taken from the Red Hat Connect catalog)
#   community  : deploy/olm-catalog/ibm-block-csi-operator-community/<version>
#                  -> <community-operators-prod>/operators/ibm-block-csi-operator-community/<version>
#                image quay.io/ibmcsiblock/ibm-block-csi-operator:<version> -> @sha256:<digest>
#                (digest taken from quay.io)
#   community-k8s : deploy/olm-catalog/ibm-block-csi-operator-community/<version>
#                  -> <community-operators>/operators/ibm-block-csi-operator-community/<version>
#                the image tag is kept as is; the whole "skips" section in the CSV is replaced by
#                "replaces: ibm-block-csi-operator.v<previous version>"
#
# Usage:
#   hack/prepare-certified-operators.sh [certified|community|community-k8s|all] [path-to-target-repo]
#   (default target is "all"; the repo path argument is only allowed with a single target)
#
# Environment variables (all optional):
#   CERTIFIED_OPERATORS_DIR   Path to the certified-operators repo (default: ../certified-operators)
#   COMMUNITY_OPERATORS_DIR   Path to the community-operators-prod repo (default: ../community-operators-prod)
#   COMMUNITY_K8S_OPERATORS_DIR Path to the community-operators repo (default: ../community-operators)
#   PREVIOUS_VERSION          Version to add to "skips" / to put in "replaces" (default: the highest version dir below the current one)
#   CERTIFIED_IMAGE_DIGEST    Skip the Red Hat Catalog lookup and use this digest (sha256:...)
#   COMMUNITY_IMAGE_DIGEST    Skip the quay.io lookup and use this digest (sha256:...)
#
# Requires: curl, jq (unless the digests are provided).

set -e

OPERATOR_NAME="ibm-block-csi-operator"
RELEASE_VERSION=$(grep "Version =" version/version.go | head -1 | awk -F '"' '{print $2}')

CERTIFIED_BUNDLES_DIR="deploy/olm-catalog/${OPERATOR_NAME}"
CERTIFIED_IMAGE_REPO="registry.connect.redhat.com/ibm/${OPERATOR_NAME}"
CERTIFIED_CATALOG_API="https://catalog.redhat.com/api/containers/v1/repositories/registry/registry.connect.redhat.com/repository/ibm/${OPERATOR_NAME}/images"

COMMUNITY_PACKAGE_NAME="${OPERATOR_NAME}-community"
COMMUNITY_BUNDLES_DIR="deploy/olm-catalog/${COMMUNITY_PACKAGE_NAME}"
COMMUNITY_IMAGE_REPO="quay.io/ibmcsiblock/${OPERATOR_NAME}"
COMMUNITY_QUAY_API="https://quay.io/api/v1/repository/ibmcsiblock/${OPERATOR_NAME}/tag/"

TARGET="all"
TARGET_REPO_ARG=""
if [ $# -ge 1 ]; then
  TARGET="${1}"
  TARGET_REPO_ARG="${2}"
fi

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

get_previous_version() {
  local bundles_dir=${1}
  ls "${bundles_dir}" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | grep -B1 -x "${RELEASE_VERSION}" | head -1
}

get_certified_image_digest() {
  local digests
  digests=$(curl -sf -m 60 -G "${CERTIFIED_CATALOG_API}" \
    --data-urlencode "page_size=100" \
    --data-urlencode "include=data.repositories.manifest_list_digest,data.repositories.tags.name,data.repositories.published" \
    --data-urlencode "filter=repositories.tags.name==${RELEASE_VERSION}" |
    jq -r --arg v "${RELEASE_VERSION}" \
      '[.data[].repositories[] | select(.published == true and any(.tags[]; .name == $v)) | .manifest_list_digest] | unique | .[]')
  if [ -z "${digests}" ]; then
    fail "no published image found for tag ${RELEASE_VERSION} in ${CERTIFIED_IMAGE_REPO}. Make sure the image was published in Red Hat Connect, or set CERTIFIED_IMAGE_DIGEST."
  fi
  if [ "$(echo "${digests}" | wc -l)" -ne 1 ]; then
    fail "found more than one digest for tag ${RELEASE_VERSION}: ${digests}"
  fi
  echo "${digests}"
}

get_community_image_digest() {
  local digests
  digests=$(curl -sf -m 60 -G "${COMMUNITY_QUAY_API}" \
    --data-urlencode "specificTag=${RELEASE_VERSION}" \
    --data-urlencode "onlyActiveTags=true" |
    jq -r --arg v "${RELEASE_VERSION}" '[.tags[] | select(.name == $v) | .manifest_digest] | unique | .[]')
  if [ -z "${digests}" ]; then
    fail "no image found for tag ${RELEASE_VERSION} in ${COMMUNITY_IMAGE_REPO}. Make sure the image was pushed to quay.io, or set COMMUNITY_IMAGE_DIGEST."
  fi
  if [ "$(echo "${digests}" | wc -l)" -ne 1 ]; then
    fail "found more than one digest for tag ${RELEASE_VERSION}: ${digests}"
  fi
  echo "${digests}"
}

# prepare_bundle <name> <source bundles dir> <target repo dir> <package name in target repo>
#                <csv file name> <image repo> <image digest or empty to look up> <digest lookup function>
prepare_bundle() {
  local name=${1} bundles_dir=${2} target_repo=${3} package_name=${4} csv_name=${5}
  local image_repo=${6} image_digest=${7} digest_lookup=${8}
  local source_dir="${bundles_dir}/${RELEASE_VERSION}"
  local target_dir="${target_repo}/operators/${package_name}/${RELEASE_VERSION}"
  local target_csv="${target_dir}/manifests/${csv_name}"

  echo
  echo "=== ${name} ==="
  [ -d "${source_dir}" ] || fail "${source_dir} does not exist."
  [ -d "${target_repo}/operators/${package_name}" ] ||
    fail "${target_repo}/operators/${package_name} does not exist. Pass the ${name} repo path."

  local previous_version="${PREVIOUS_VERSION:-$(get_previous_version "${bundles_dir}")}"
  if [ -z "${previous_version}" ] || [ "${previous_version}" == "${RELEASE_VERSION}" ]; then
    fail "could not determine the previous version. Set PREVIOUS_VERSION."
  fi

  image_digest="${image_digest:-$(${digest_lookup})}"
  echo "${image_digest}" | grep -qE '^sha256:[0-9a-f]{64}$' || fail "invalid digest '${image_digest}'."

  echo "Release version:  ${RELEASE_VERSION}"
  echo "Previous version: ${previous_version}"
  echo "Image digest:     ${image_digest}"
  echo "Target:           ${target_dir}"

  rm -rf "${target_dir}"
  mkdir -p "${target_dir}"
  cp -R "${source_dir}/." "${target_dir}/"
  [ -f "${target_csv}" ] || fail "${target_csv} does not exist."

  # 1. Replace the image tag with the digest.
  sed -i.bak "s#${image_repo}:${RELEASE_VERSION}#${image_repo}@${image_digest}#g" "${target_csv}"

  # 2. Add the previous version as the last entry of "skips" (if not already there).
  local skip_entry="${OPERATOR_NAME}.v${previous_version}"
  if ! grep -qx "    - ${skip_entry}" "${target_csv}"; then
    awk -v entry="    - ${skip_entry}" '
      /^  skips:/ { in_skips = 1; print; next }
      in_skips && !/^    - / { print entry; in_skips = 0; added = 1 }
      { print }
      END { if (!added) exit 1 }
    ' "${target_csv}" > "${target_csv}.tmp" || {
      rm -f "${target_csv}.tmp" "${target_csv}.bak"
      fail "could not find a 'skips' section in ${target_csv}."
    }
    mv "${target_csv}.tmp" "${target_csv}"
  fi
  rm -f "${target_csv}.bak"

  # Sanity check.
  if grep -q "${image_repo}:${RELEASE_VERSION}" "${target_csv}"; then
    fail "tag-based image references are still present in ${target_csv}."
  fi

  echo "Image references in ${target_csv}:"
  grep -n "${image_repo}" "${target_csv}"
  echo "Skips tail:"
  grep -n -A1 -x "    - ${skip_entry}" "${target_csv}"
}

prepare_community_k8s() {
  local target_repo="${1:-${COMMUNITY_K8S_OPERATORS_DIR:-../community-operators}}"
  local source_dir="${COMMUNITY_BUNDLES_DIR}/${RELEASE_VERSION}"
  local target_dir="${target_repo}/operators/${COMMUNITY_PACKAGE_NAME}/${RELEASE_VERSION}"
  local target_csv="${target_dir}/manifests/${OPERATOR_NAME}.v${RELEASE_VERSION}.clusterserviceversion.yaml"

  echo
  echo "=== community-operators ==="
  [ -d "${source_dir}" ] || fail "${source_dir} does not exist."
  [ -d "${target_repo}/operators/${COMMUNITY_PACKAGE_NAME}" ] ||
    fail "${target_repo}/operators/${COMMUNITY_PACKAGE_NAME} does not exist. Pass the community-operators repo path."

  local previous_version="${PREVIOUS_VERSION:-$(get_previous_version "${COMMUNITY_BUNDLES_DIR}")}"
  if [ -z "${previous_version}" ] || [ "${previous_version}" == "${RELEASE_VERSION}" ]; then
    fail "could not determine the previous version. Set PREVIOUS_VERSION."
  fi

  echo "Release version:  ${RELEASE_VERSION}"
  echo "Previous version: ${previous_version}"
  echo "Target:           ${target_dir}"

  rm -rf "${target_dir}"
  mkdir -p "${target_dir}"
  cp -R "${source_dir}/." "${target_dir}/"
  [ -f "${target_csv}" ] || fail "${target_csv} does not exist."

  # Replace the whole "skips" section with a single "replaces" line.
  awk -v replaces="  replaces: ${OPERATOR_NAME}.v${previous_version}" '
    /^  skips:/ { in_skips = 1; print replaces; replaced = 1; next }
    in_skips && /^    - / { next }
    { in_skips = 0; print }
    END { if (!replaced) exit 1 }
  ' "${target_csv}" > "${target_csv}.tmp" || {
    rm -f "${target_csv}.tmp"
    fail "could not find a 'skips' section in ${target_csv}."
  }
  mv "${target_csv}.tmp" "${target_csv}"

  echo "Result in ${target_csv}:"
  grep -n -B1 -A1 "^  replaces:" "${target_csv}"
}

prepare_certified() {
  prepare_bundle "certified-operators" "${CERTIFIED_BUNDLES_DIR}" \
    "${1:-${CERTIFIED_OPERATORS_DIR:-../certified-operators}}" "${OPERATOR_NAME}" \
    "${OPERATOR_NAME}.clusterserviceversion.yaml" "${CERTIFIED_IMAGE_REPO}" \
    "${CERTIFIED_IMAGE_DIGEST}" get_certified_image_digest
}

prepare_community() {
  prepare_bundle "community-operators-prod" "${COMMUNITY_BUNDLES_DIR}" \
    "${1:-${COMMUNITY_OPERATORS_DIR:-../community-operators-prod}}" "${COMMUNITY_PACKAGE_NAME}" \
    "${OPERATOR_NAME}.v${RELEASE_VERSION}.clusterserviceversion.yaml" "${COMMUNITY_IMAGE_REPO}" \
    "${COMMUNITY_IMAGE_DIGEST}" get_community_image_digest
}

case "${TARGET}" in
  certified) prepare_certified "${TARGET_REPO_ARG}" ;;
  community) prepare_community "${TARGET_REPO_ARG}" ;;
  community-k8s) prepare_community_k8s "${TARGET_REPO_ARG}" ;;
  all)
    [ -z "${TARGET_REPO_ARG}" ] || fail "a repo path can't be used with target 'all'; use the *_OPERATORS_DIR variables."
    prepare_certified
    prepare_community
    prepare_community_k8s
    ;;
  *) fail "unknown target '${TARGET}'. Use certified, community, community-k8s or all." ;;
esac
