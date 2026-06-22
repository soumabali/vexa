#!/usr/bin/env bash
set -euo pipefail

# SSH Manager - Desktop Clippy Check Script
# Security-focused Rust linting with auto-fix capability
# CI integration ready

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

SSH Manager Desktop - Rust Clippy Lint Runner

OPTIONS:
    -h, --help          Show this help message
    -f, --fix           Auto-fix issues where possible
    -c, --ci            CI mode (strict, no interactive)
    -t, --tests         Include test files in linting
    -d, --deny-warnings  Treat warnings as errors
    -a, --all-targets   Check all targets (bin, lib, tests, examples)
    -j, --json          Output in JSON format (CI friendly)
    -q, --quiet         Suppress non-error output
    -v, --verbose       Verbose output

EXAMPLES:
    $(basename "$0")              # Run standard clippy check
    $(basename "$0") --fix        # Run with auto-fix
    $(basename "$0") --ci         # CI mode (strict)
    $(basename "$0") --all-targets --deny-warnings

EOF
}

# Default options
FIX=false
CI_MODE=false
INCLUDE_TESTS=true
DENY_WARNINGS=false
ALL_TARGETS=false
JSON_OUTPUT=false
QUIET=false
VERBOSE=false
CLIPPY_ARGS=()

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -f|--fix)
            FIX=true
            shift
            ;;
        -c|--ci)
            CI_MODE=true
            DENY_WARNINGS=true
            shift
            ;;
        -t|--tests)
            INCLUDE_TESTS=true
            shift
            ;;
        -d|--deny-warnings)
            DENY_WARNINGS=true
            shift
            ;;
        -a|--all-targets)
            ALL_TARGETS=true
            shift
            ;;
        -j|--json)
            JSON_OUTPUT=true
            shift
            ;;
        -q|--quiet)
            QUIET=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Change to project directory
cd "${PROJECT_DIR}"

# Check if cargo is available
if ! command -v cargo &> /dev/null; then
    log_error "cargo not found. Please install Rust toolchain."
    exit 1
fi

# Check if clippy is available
if ! cargo clippy --version &> /dev/null; then
    log_error "cargo clippy not available. Installing..."
    rustup component add clippy
fi

# Build clippy arguments
if [[ "$FIX" == true ]]; then
    CLIPPY_ARGS+=("--fix")
    CLIPPY_ARGS+=("--allow-dirty")
    CLIPPY_ARGS+=("--allow-staged")
    log_info "Running clippy with auto-fix enabled..."
else
    log_info "Running clippy check..."
fi

if [[ "$ALL_TARGETS" == true ]]; then
    CLIPPY_ARGS+=("--all-targets")
    log_info "Including all targets (bin, lib, tests, examples)..."
fi

if [[ "$DENY_WARNINGS" == true ]]; then
    CLIPPY_ARGS+=("--" "-Dwarnings")
    log_info "Warnings will be treated as errors."
fi

if [[ "$INCLUDE_TESTS" == true ]]; then
    CLIPPY_ARGS+=("--tests")
fi

# Always include these
CLIPPY_ARGS+=("--all-features")

if [[ "$VERBOSE" == true ]]; then
    CLIPPY_ARGS+=("-v")
fi

if [[ "$QUIET" == true ]]; then
    CLIPPY_ARGS+=("-q")
fi

if [[ "$JSON_OUTPUT" == true ]]; then
    export RUSTFLAGS="-Z unstable-options --error-format=json"
fi

# Print configuration
if [[ "$VERBOSE" == true ]]; then
    log_info "Project directory: ${PROJECT_DIR}"
    log_info "Clippy config: ${PROJECT_DIR}/clippy.toml"
    log_info "Arguments: ${CLIPPY_ARGS[*]}"
fi

# Run cargo check first (fast feedback)
if [[ "$FIX" == false ]]; then
    log_info "Running cargo check for quick validation..."
    if ! cargo check --all-features 2>&1 | tee /tmp/cargo-check.log; then
        log_error "cargo check failed! Fix compilation errors before running clippy."
        exit 1
    fi
    log_success "cargo check passed."
    echo
fi

# Run clippy
log_info "Running cargo clippy..."
if cargo clippy "${CLIPPY_ARGS[@]}" 2>&1; then
    log_success "Clippy check passed!"
    
    # Also run cargo fmt check in CI mode
    if [[ "$CI_MODE" == true ]] && [[ "$FIX" == false ]]; then
        log_info "Running cargo fmt check..."
        if cargo fmt -- --check; then
            log_success "Formatting check passed!"
        else
            log_error "Formatting issues found. Run 'cargo fmt' to fix."
            exit 1
        fi
    fi
    
    # Run cargo doc check in CI mode
    if [[ "$CI_MODE" == true ]]; then
        log_info "Running cargo doc check..."
        if RUSTDOCFLAGS="-D warnings" cargo doc --no-deps --all-features 2>&1 | grep -E "^error" || true; then
            log_warn "Documentation warnings found."
        else
            log_success "Documentation check passed!"
        fi
    fi
    
    exit 0
else
    log_error "Clippy found issues. Please review and fix."
    if [[ "$FIX" == false ]]; then
        log_info "Tip: Run '$(basename "$0") --fix' to auto-fix some issues."
    fi
    exit 1
fi
