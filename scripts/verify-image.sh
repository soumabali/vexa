#!/usr/bin/env bash
#
# SSH Manager - Container Image Verification Script
# Verifies Cosign signatures and SBOM integrity
#
# Usage:
#   ./scripts/verify-image.sh <image:tag> [identity] [issuer]
#
# Examples:
#   ./scripts/verify-image.sh ghcr.io/user/ssh-manager:latest
#   ./scripts/verify-image.sh ghcr.io/user/ssh-manager:v1.0.0
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
REKOR_URL="${REKOR_URL:-https://rekor.sigstore.dev}"

# Parse arguments
IMAGE="${1:-}"
CERT_IDENTITY="${2:-}"
CERT_ISSUER="${3:-https://token.actions.githubusercontent.com}"
VERBOSE="${VERBOSE:-false}"

if [ -z "$IMAGE" ]; then
    echo "Usage: $0 <image:tag|image@digest> [certificate-identity] [certificate-oidc-issuer]"
    echo ""
    echo "Examples:"
    echo "  $0 ghcr.io/user/ssh-manager:latest"
    echo "  $0 ghcr.io/user/ssh-manager:v1.0.0"
    echo "  $0 ghcr.io/user/ssh-manager@sha256:abc123..."
    echo ""
    echo "Options:"
    echo "  Set VERBOSE=true for detailed output"
    echo ""
    echo "Environment variables:"
    echo "  COSIGN_VERSION  - Cosign version to install (default: v2.2.4)"
    echo "  REKOR_URL       - Rekor transparency log URL (default: https://rekor.sigstore.dev)"
    exit 1
fi

# Check dependencies
check_deps() {
    log_info "Checking dependencies..."
    
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
    
    if [ -w /usr/local/bin ]; then
        mv /tmp/cosign /usr/local/bin/cosign
    else
        PATH="/tmp:$PATH"
    fi
    
    log_success "Cosign installed"
}

# Verify image signature
verify_image_signature() {
    local image="$1"
    
    log_info "Verifying signature for: $image"
    log_info "Rekor URL: $REKOR_URL"
    
    local verify_args=(
        verify
        --rekor-url "$REKOR_URL"
    )
    
    if [ -n "$CERT_IDENTITY" ]; then
        verify_args+=(--certificate-identity "$CERT_IDENTITY")
        log_info "Expected identity: $CERT_IDENTITY"
    fi
    
    if [ -n "$CERT_ISSUER" ]; then
        verify_args+=(--certificate-oidc-issuer "$CERT_ISSUER")
        log_info "Expected issuer: $CERT_ISSUER"
    fi
    
    if [ "$VERBOSE" = "true" ]; then
        verify_args+=(--verbose)
    fi
    
    # Run verification
    if cosign "${verify_args[@]}" "$image"; then
        log_success "Image signature VERIFIED ✓"
        return 0
    else
        log_error "Image signature verification FAILED ✗"
        return 1
    fi
}

# Verify SBOM signature
verify_sbom() {
    local sbom_file="$1"
    
    if [ ! -f "$sbom_file" ]; then
        log_error "SBOM file not found: $sbom_file"
        return 1
    fi
    
    local sig_file="${sbom_file}.sig"
    local cert_file="${sbom_file}.cert"
    
    if [ ! -f "$sig_file" ]; then
        log_error "SBOM signature not found: $sig_file"
        return 1
    fi
    
    if [ ! -f "$cert_file" ]; then
        log_error "SBOM certificate not found: $cert_file"
        return 1
    fi
    
    log_info "Verifying SBOM: $sbom_file"
    
    local verify_args=(
        verify-blob
        --signature "$sig_file"
        --certificate "$cert_file"
        --rekor-url "$REKOR_URL"
    )
    
    if [ -n "$CERT_IDENTITY" ]; then
        verify_args+=(--certificate-identity "$CERT_IDENTITY")
    fi
    
    if [ -n "$CERT_ISSUER" ]; then
        verify_args+=(--certificate-oidc-issuer "$CERT_ISSUER")
    fi
    
    if cosign "${verify_args[@]}" "$sbom_file"; then
        log_success "SBOM signature VERIFIED ✓"
        return 0
    else
        log_error "SBOM signature verification FAILED ✗"
        return 1
    fi
}

# Get signature information
get_signature_info() {
    local image="$1"
    
    log_info "Retrieving signature information..."
    
    local info
    if info=$(cosign verify \
        --rekor-url "$REKOR_URL" \
        --certificate-identity-regexp ".*" \
        --certificate-oidc-issuer-regexp ".*" \
        --output json \
        "$image" 2>&1); then
        
        echo "$info" | jq -r '.[0] | {
            "Certificate Subject": .optional.Subject,
            "Issuer": .optional.Issuer,
            "Repository": .optional.annotations.repo,
            "Commit": .optional.annotations.commit,
            "Branch": .optional.annotations.branch,
            "Signed At": .optional.annotations."signed-at"
        }' 2>/dev/null || echo "$info"
    else
        log_warn "Could not retrieve detailed signature information"
    fi
}

# Check attestation
verify_attestation() {
    local image="$1"
    local predicate_type="${2:-https://slsa.dev/provenance/v1}"
    
    log_info "Checking attestation (type: $predicate_type)..."
    
    local attestation_args=(
        verify-attestation
        --type "$predicate_type"
        --rekor-url "$REKOR_URL"
    )
    
    if [ -n "$CERT_IDENTITY" ]; then
        attestation_args+=(--certificate-identity "$CERT_IDENTITY")
    fi
    
    if [ -n "$CERT_ISSUER" ]; then
        attestation_args+=(--certificate-oidc-issuer "$CERT_ISSUER")
    fi
    
    if cosign "${attestation_args[@]}" "$image" 2>/dev/null; then
        log_success "Attestation VERIFIED ✓"
        return 0
    else
        log_warn "No attestation found or verification failed"
        return 1
    fi
}

# Main execution
main() {
    echo "=========================================="
    echo "  SSH Manager - Image Verification"
    echo "=========================================="
    echo ""
    
    check_deps
    
    log_info "Target image: $IMAGE"
    
    # Verify image signature
    local image_verified=false
    if verify_image_signature "$IMAGE"; then
        image_verified=true
    fi
    
    # Get signature details
    if [ "$image_verified" = "true" ]; then
        echo ""
        log_info "Signature details:"
        get_signature_info "$IMAGE" || true
    fi
    
    # Check for attestation
    echo ""
    verify_attestation "$IMAGE" || true
    
    # Summary
    echo ""
    echo "=========================================="
    echo "  Verification Summary"
    echo "=========================================="
    echo ""
    
    if [ "$image_verified" = "true" ]; then
        log_success "Image signature: VERIFIED"
        echo ""
        log_success "✅ Image is safe to deploy"
        exit 0
    else
        log_error "Image signature: FAILED"
        echo ""
        log_error "❌ DO NOT DEPLOY - Signature verification failed"
        exit 1
    fi
}

main "$@"
