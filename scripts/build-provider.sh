#!/usr/bin/env bash
# Build terraform-provider-readthedocs and install it into the local
# Terraform / OpenTofu plugin dir. No HashiCorp Registry listing required.
#
# Mirror layout (filesystem_mirror):
#   ~/.local/share/terraform/plugins/registry.opentofu.org/the-robot-lives/readthedocs/<ver>/<os>_<arch>/
#   ~/.local/share/terraform/plugins/registry.terraform.io/the-robot-lives/readthedocs/<ver>/<os>_<arch>/
#
# Dev-overrides copy (same pattern as signoz/signoz):
#   ~/.local/share/terraform/plugins/terraform-provider-readthedocs
#
# ~/.terraformrc:
#   provider_installation {
#     dev_overrides {
#       "the-robot-lives/readthedocs" = "~/.local/share/terraform/plugins"
#     }
#     filesystem_mirror {
#       path    = "~/.local/share/terraform/plugins"
#       include = ["the-robot-lives/readthedocs"]
#     }
#     direct {}
#   }
#
# Usage: scripts/build-provider.sh
set -euo pipefail

VERSION="0.1.0"
NAMESPACE="the-robot-lives"
TYPE="readthedocs"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROVIDER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

GO_BIN="$(command -v go || true)"
if [[ -z "${GO_BIN}" ]]; then
  echo "error: could not locate a 'go' binary." >&2
  exit 1
fi

OS="$("${GO_BIN}" env GOOS)"
ARCH="$("${GO_BIN}" env GOARCH)"
MIRROR="${HOME}/.local/share/terraform/plugins"
BIN_NAME="terraform-provider-${TYPE}"

echo "Building ${BIN_NAME} v${VERSION} (${OS}/${ARCH}) with ${GO_BIN}"
STAGING="$(mktemp -d)"
trap 'rm -rf "${STAGING}"' EXIT
( cd "${PROVIDER_DIR}" && "${GO_BIN}" build -ldflags "-X main.version=${VERSION}" -o "${STAGING}/${BIN_NAME}" . )

install_one() {
  local dest_dir="$1"
  mkdir -p "${dest_dir}"
  cp "${STAGING}/${BIN_NAME}" "${dest_dir}/${BIN_NAME}"
  chmod +x "${dest_dir}/${BIN_NAME}"
  echo "Installed → ${dest_dir}/${BIN_NAME}"
}

# filesystem_mirror: OpenTofu default host + Terraform default host
for registry in registry.opentofu.org registry.terraform.io; do
  install_one "${MIRROR}/${registry}/${NAMESPACE}/${TYPE}/${VERSION}/${OS}_${ARCH}"
done

# dev_overrides: binary sits in the plugins root (see signoz/signoz)
install_one "${MIRROR}"
