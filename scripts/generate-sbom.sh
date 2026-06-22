#!/usr/bin/env bash
#
# SSH Manager - SBOM Generation Script
# Uses Syft to generate SPDX-formatted SBOMs
#
# Usage:
#   ./scripts/generate-sbom.sh [image:tag|directory] [output-format] [output-file]
#
# Examples:
#   ./scripts/generate-sbom.sh ghcr.io/user/ssh-manager:latest
#   ./scripts/generate-sbom.sh . spdx-json sbom.spdx.json
#   ./scripts/generate-sbom.sh ghcr.io/user/ssh-manager:v1.0.0 spdx-json sbom.spdx.json
#

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_success() { echo -e "${GREEN}[OK]${NC} $1"; }
log_step() { echo -e "${CYAN}[STEP]${NC} $1"; }

# Configuration
SYFT_VERSION="${SYFT_VERSION:-1.4.1}"
SOURCE="${1:-.}"
FORMAT="${2:-spdx-json}"
OUTPUT="${3:-}"

if [ -z "$OUTPUT" ]; then
    # Generate default output filename
    if [ "$SOURCE" = "." ]; then
        OUTPUT="sbom-source.spdx.json"
    else
        local safe_name
        safe_name=$(echo "$SOURCE" | tr '/:' '_')
        OUTPUT="sbom-${safe_name}.spdx.json"
    fi
fi

# Help
show_help() {
    echo "SSH Manager - SBOM Generation Script"
    echo ""
    echo "Usage: $0 [source] [format] [output-file]"
    echo ""
    echo "Arguments:"
    echo "  source       Image tag, directory, or archive (default: current directory)"
    echo "  format       Output format (default: spdx-json)"
    echo "  output-file  Output file path (default: auto-generated)"
    echo ""
    echo "Supported formats:"
    echo "  spdx-json      SPDX JSON format (default)"
    echo "  spdx-tag-value SPDX tag-value format"
    echo "  cyclonedx-json CycloneDX JSON format"
    echo "  cyclonedx-xml  CycloneDX XML format"
    echo "  syft-json      Syft native JSON format"
    echo "  table          Human-readable table"
    echo "  text           Plain text"
    echo ""
    echo "Examples:"
    echo "  $0                                            # Generate SBOM for current directory"
    echo "  $0 ghcr.io/user/ssh-manager:latest            # Generate SBOM for container image"
    echo "  $0 . spdx-json sbom.spdx.json               # Specify format and output"
    echo "  $0 /path/to/source cyclonedx-json bom.json    # Generate CycloneDX SBOM"
    echo ""
    echo "Environment variables:"
    echo "  SYFT_VERSION  - Syft version to install (default: 1.4.1)"
    exit 0
}

if [ "$SOURCE" = "-h" ] || [ "$SOURCE" = "--help" ]; then
    show_help
fi

# Check dependencies
check_deps() {
    log_step "Checking dependencies..."
    
    if ! command -v syft &> /dev/null; then
        log_warn "Syft not found. Installing..."
        install_syft
    fi
    
    log_success "Syft version: $(syft version | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1 || syft version | head -1)"
    
    if command -v jq &> /dev/null; then
        log_success "jq: available"
    else
        log_warn "jq not found (optional, for JSON processing)"
    fi
}

# Install Syft
install_syft() {
    local os
    local arch
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)
    
    case "$arch" in
        x86_64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
    esac
    
    local install_url="https://raw.githubusercontent.com/anchore/syft/main/install.sh"
    
    log_info "Installing Syft ${SYFT_VERSION}..."
    
    if curl -sSf "$install_url" | sh -s -- -b /tmp "v${SYFT_VERSION#v}"; then
        if [ -w /usr/local/bin ]; then
            mv /tmp/syft /usr/local/bin/syft
            log_success "Syft installed to /usr/local/bin/syft"
        else
            PATH="/tmp:$PATH"
            log_warn "Cannot write to /usr/local/bin, using /tmp/syft"
        fi
    else
        log_error "Failed to install Syft"
        exit 1
    fi
}

# Detect source type
detect_source_type() {
    local source="$1"
    
    if [ -d "$source" ]; then
        echo "directory"
    elif echo "$source" | grep -qE '^(docker://|oci://|registry://|podman://)?[a-zA-Z0-9._/-]+(:[a-zA-Z0-9._-]+|@[a-zA-Z0-9:]+)$'; then
        echo "image"
    elif echo "$source" | grep -qE '\.(tar|tar\.gz|tgz)$'; then
        echo "archive"
    else
        echo "unknown"
    fi
}

# Generate SBOM
generate_sbom() {
    local source="$1"
    local format="$2"
    local output="$3"
    
    local source_type
    source_type=$(detect_source_type "$source")
    
    log_step "Generating SBOM..."
    log_info "Source: $source"
    log_info "Source type: $source_type"
    log_info "Format: $format"
    log_info "Output: $output"
    
    # Build syft arguments
    local syft_args=()
    
    # Add source with scheme if needed
    if [ "$source_type" = "image" ] && ! echo "$source" | grep -qE '^(docker|oci|registry|podman)://'; then
        source="docker:$source"
    fi
    
    syft_args+=("$source")
    syft_args+=(-o "$format="$output"")
    
    # Add metadata
    syft_args+=(--source-name "ssh-manager")
    syft_args+=(--source-version "${GITHUB_REF_NAME:-$(git describe --tags --always 2>/dev/null || echo 'dev')}")
    
    # Run syft
    log_info "Running: syft ${syft_args[*]}"
    syft "${syft_args[@]}"
    
    if [ -f "$output" ]; then
        log_success "SBOM generated: $output"
    else
        log_error "SBOM generation failed - output file not found"
        exit 1
    fi
}

# Validate SBOM
validate_sbom() {
    local output="$1"
    
    log_step "Validating SBOM..."
    
    if [ ! -f "$output" ]; then
        log_error "SBOM file not found: $output"
        return 1
    fi
    
    # Check file size
    local size
    size=$(stat -f%z "$output" 2>/dev/null || stat -c%s "$output" 2>/dev/null || echo "0")
    log_info "File size: $(numfmt --to=iec "$size" 2>/dev/null || echo "${size} bytes")"
    
    if [ "$size" -lt 100 ]; then
        log_error "SBOM file is too small, likely invalid"
        return 1
    fi
    
    # Validate JSON structure for SPDX
    if echo "$output" | grep -q "spdx"; then
        if command -v jq >/dev/null 2>&1; then
            if jq empty "$output" 2>/dev/null; then
                log_success "Valid JSON structure"
            else
                log_error "Invalid JSON structure"
                return 1
            fi
            
            # Check required SPDX fields
            local spdx_version
            spdx_version=$(jq -r '.spdxVersion // empty' "$output" 2>/dev/null)
            local spdx_id
            spdx_id=$(jq -r '.SPDXID // empty' "$output" 2>/dev/null)
            local name
            name=$(jq -r '.name // empty' "$output" 2>/dev/null)
            
            if [ -n "$spdx_version" ] && [ -n "$spdx_id" ] && [ -n "$name" ]; then
                log_success "Required SPDX fields present"
                log_info "SPDX Version: $spdx_version"
                log_info "SPDX ID: $spdx_id"
                log_info "Name: $name"
            else
                log_warn "Some SPDX fields are missing"
            fi
            
            # Count packages
            local packages
            packages=$(jq '.packages | length' "$output" 2>/dev/null || echo "0")
            log_info "Packages found: $packages"
            
            # Count files
            local files
            files=$(jq '.files | length' "$output" 2>/dev/null || echo "0")
            log_info "Files found: $files"
        else
            log_warn "jq not available, skipping JSON validation"
        fi
    fi
    
    log_success "SBOM validation passed"
}

# Sign SBOM
sign_sbom() {
    local output="$1"
    
    log_step "Signing SBOM..."
    
    if ! command -v cosign >/dev/null 2>&1; then
        log_warn "Cosign not found, skipping SBOM signing"
        return 0
    fi
    
    if cosign sign-blob \
        --yes \
        --output-signature "${output}.sig" \
        --output-certificate "${output}.cert" \
        "$output" 2>/dev/null; then
        log_success "SBOM signed!"
        log_info "Signature: ${output}.sig"
        log_info "Certificate: ${output}.cert"
    else
        log_warn "Failed to sign SBOM (this is optional)"
    fi
}

# Display summary
display_summary() {
    local output="$1"
    local format="$2"
    
    echo ""
    echo "=========================================="
    echo "  SBOM Generation Complete"
    echo "=========================================="
    echo ""
    log_info "Output file: $(realpath "$output" 2>/dev/null || echo "$output")"
    log_info "Format: $format"
    
    if [ -f "${output}.sig" ]; then
        log_info "Signature: ${output}.sig"
    fi
    
    echo ""
    echo "Usage:"
    echo "  View:     cat $output | jq .packages[].name"
    echo "  Verify:   ./scripts/verify-image.sh ${IMAGE:-image:tag}"
    echo "  Sign:     cosign sign-blob --yes $output"
    echo ""
}

# Main execution
main() {
    echo "=========================================="
    echo "  SSH Manager - SBOM Generation"
    echo "=========================================="
    echo ""
    
    check_deps
    
    # Generate SBOM
    generate_sbom "$SOURCE" "$FORMAT" "$OUTPUT"
    
    # Validate
    validate_sbom "$OUTPUT"
    
    # Sign SBOM (optional)
    sign_sbom "$OUTPUT"
    
    # Display summary
    display_summary "$OUTPUT" "$FORMAT"
    
    log_success "Done!"
}

main "$@"
