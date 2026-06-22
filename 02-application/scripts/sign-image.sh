#!/usr/bin/env bash
#
# SSH Manager - Container Image Signing Script
# Uses Cosign keyless signing with Fulcio + Rekor
#
# Usage:
#   ./scripts/sign-image.sh <image:tag> [identity] [issuer]
#
# Examples:
#   ./scripts/sign-image.sh ghcr.io/user/ssh-manager:latest
#   ./scripts/sign-image.sh ghcr.io/user/ssh-manager:v1.0.0
#

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_success() { echo -e "${GREEN}[OK]${NC} $1"; }

# Configuration
COSIGN_VERSION="${COSIGN_VERSION:-v2.2.4}"
COSIGN_EXPERIMENTAL="${COSIGN_EXPERIMENTAL:-1}"
export COSIGN_EXPERIMENTAL

# Parse arguments
IMAGE="${1:-}"
CERT_IDENTITY="${2:-}"
CERT_ISSUER="${3:-https://token.actions.githubusercontent.com}"

if [ -z "$IMAGE" ]; then
    echo "Usage: $0 <image:tag|image@digest> [certificate-identity] [certificate-oidc-issuer]"
    echo ""
    echo "Examples:"
    echo "  $0 ghcr.io/user/ssh-manager:latest"
    echo "  $0 ghcr.io/user/ssh-manager:v1.0.0"
    echo "  $0 ghcr.io/user/ssh-manager@sha256:abc123..."
    echo ""
    echo "Environment variables:"
    echo "  COSIGN_VERSION    - Cosign version to install (default: v2.2.4)"
    echo "  COSIGN_EXPERIMENTAL - Enable experimental features (default: 1)"
    exit 1
fi

# Resolve image to digest if needed
resolve_image() {
    local img="$1"
    
    # If already a digest reference, use as-is
    if echo "$img" | grep -q '@sha256:'; then
        echo "$img"
        return 0
    fi
    
    # Try to get digest from registry
    log_info "Resolving image digest for: $img"
    
    if command -v docker &> /dev/null; then
        docker pull "$img" &> /dev/null || true
        local digest
        digest=$(docker inspect --format='{{index .RepoDigests 0}}' "$img" 2>/dev/null || true)
        if [ -n "$digest" ]; then
            echo "$digest"
            return 0
        fi
    fi
    
    if command -v crane &> /dev/null; then
        local digest
        digest=$(crane digest "$img" 2>/dev/null || true)
        if [ -n "$digest" ]; then
            echo "${img%%:*}@${digest}"
            return 0
        fi
    fi
    
    log_warn "Could not resolve digest, using tag reference"
    echo "$img"
}

# Check dependencies
check_deps() {
    log_info "Checking dependencies..."
    
    # Check for cosign
    if ! command -v cosign &> /dev/null; then
        log_warn "Cosign not found. Installing..."
        install_cosign
    fi
    
    log_info "Cosign version: $(cosign version --json 2>/dev/null | grep -o '"gitVersion":"[^"]*"' | cut -d'"' -f4 || cosign version | head -1)"
}

# Install cosign if needed
install_cosign() {
    local os
    local arch
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)
    
    case "$arch" in
        x86_64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
    esac
    
    local url="https://github.com/sigstore/cosign/releases/download/${COSIGN_VERSION}/cosign-${os}-${arch}"
    
    log_info "Downloading Cosign from: $url"
    curl -sL "$url" -o /tmp/cosign
    chmod +x /tmp/cosign
    
    # Verify checksum
    local checksum_url="https://github.com/sigstore/cosign/releases/download/${COSIGN_VERSION}/cosign-${os}-${arch}.sig"
    
    if [ -w /usr/local/bin ]; then
        mv /tmp/cosign /usr/local/bin/cosign
    else
        log_warn "Cannot write to /usr/local/bin, using /tmp/cosign"
        PATH="/tmp:$PATH"
    fi
    
    log_success "Cosign installed"
}

# Sign image with Cosign
sign_image() {
    local image="$1"
    
    log_info "Signing image: $image"
    log_info "Using keyless signing (Fulcio + Rekor)"
    
    local cosign_args=(
        sign
        --yes
        --recursive
    )
    
    # Add annotations
    cosign_args+=(
        --annotations "repo=${GITHUB_REPOSITORY:-$(git remote get-url origin 2>/dev/null | sed 's/.*github.com\///' | sed 's/\.git$//' || echo 'unknown')}"
        --annotations "commit=${GITHUB_SHA:-$(git rev-parse HEAD 2>/dev/null || echo 'unknown')}"
        --annotations "branch=${GITHUB_REF_NAME:-$(git branch --show-current 2>/dev/null || echo 'unknown')}"
        --annotations "workflow=${GITHUB_WORKFLOW:-manual}"
        --annotations "signed-by=${USER:-$(whoami)}"
        --annotations "signed-at=$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    )
    
    # Sign the image
    cosign "${cosign_args[@]}" "$image"
    
    log_success "Image signed successfully!"
}

# Sign SBOM file
sign_sbom() {
    local sbom_file="$1"
    
    if [ ! -f "$sbom_file" ]; then
        log_error "SBOM file not found: $sbom_file"
        return 1
    fi
    
    log_info "Signing SBOM: $sbom_file"
    
    cosign sign-blob \
        --yes \
        --output-signature "${sbom_file}.sig" \
        --output-certificate "${sbom_file}.cert" \
        "$sbom_file"
    
    log_success "SBOM signed!"
    log_info "Signature: ${sbom_file}.sig"
    log_info "Certificate: ${sbom_file}.cert"
}

# Main execution
main() {
    echo "=========================================="
    echo "  SSH Manager - Container Image Signing"
    echo "=========================================="
    echo ""
    
    check_deps
    
    # Resolve image to digest
    local resolved_image
    resolved_image=$(resolve_image "$IMAGE")
    
    if [ "$resolved_image" != "$IMAGE" ]; then
        log_info "Resolved to digest: $resolved_image"
    fi
    
    # Sign the image
    sign_image "$resolved_image"
    
    # Display verification command
    echo ""
    echo "=========================================="
    echo "  Verification Command"
    echo "=========================================="
    echo ""
    echo "To verify this image signature:"
    echo ""
    echo "  cosign verify \\"
    
    if [ -n "$CERT_IDENTITY" ]; then
        echo "    --certificate-identity=\"$CERT_IDENTITY\" \\"
    fi
    if [ -n "$CERT_ISSUER" ]; then
        echo "    --certificate-oidc-issuer=\"$CERT_ISSUER\" \\"
    fi
    echo "    \"$resolved_image\""
    echo ""
    
    log_success "Signing complete!"
    
    # Export for downstream use
    echo ""
    echo "Export variables:"
    echo "  SIGNED_IMAGE=$resolved_image"
}

main "$@"
