#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════
# SSH Manager — Pre-commit Hooks Setup Script
# ═══════════════════════════════════════════════════════════════════
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# ─── Colors ─────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# ─── Helpers ────────────────────────────────────────────────────────
info()    { echo -e "${BLUE}ℹ${NC}  $*"; }
success() { echo -e "${GREEN}✓${NC}  $*"; }
warn()    { echo -e "${YELLOW}⚠${NC}  $*"; }
error()   { echo -e "${RED}✗${NC}  $*" >&2; }
header()  { echo -e "\n${BOLD}${CYAN}$*${NC}\n"; }

# ─── Check Dependency ─────────────────────────────────────────────
check_cmd() {
    local cmd="$1"
    local install_url="${2:-}"
    
    if command -v "$cmd" &>/dev/null; then
        local version
        version=$($cmd --version 2>/dev/null | head -n1 || echo "unknown")
        success "$cmd found: $version"
        return 0
    else
        error "$cmd not found"
        if [[ -n "$install_url" ]]; then
            echo "   Install: ${install_url}"
        fi
        return 1
    fi
}

# ─── Install via pipx or pip ──────────────────────────────────────
install_python_tool() {
    local pkg="$1"
    local cmd="${2:-$pkg}"
    
    if command -v pipx &>/dev/null; then
        pipx install "$pkg" --include-deps 2>/dev/null || pipx upgrade "$pkg" 2>/dev/null || true
    elif command -v pip3 &>/dev/null; then
        pip3 install --user "$pkg" 2>/dev/null || true
    else
        return 1
    fi
}

# ─── Main ───────────────────────────────────────────────────────────
cd "$PROJECT_ROOT"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║   SSH Manager — Pre-commit Hooks Setup                           ║"
echo "╚══════════════════════════════════════════════════════════════════╝"

# ═══════════════════════════════════════════════════════════════════
# 1. PRE-COMMIT FRAMEWORK
# ═══════════════════════════════════════════════════════════════════
header "1. Pre-commit Framework"

if ! check_cmd "pre-commit" "https://pre-commit.com/#install"; then
    info "Installing pre-commit..."
    
    if command -v pipx &>/dev/null; then
        pipx install pre-commit
    elif command -v pip3 &>/dev/null; then
        pip3 install --user pre-commit
    elif command -v brew &>/dev/null; then
        brew install pre-commit
    elif command -v apt-get &>/dev/null; then
        sudo apt-get update && sudo apt-get install -y pre-commit
    elif command -v pacman &>/dev/null; then
        sudo pacman -S --noconfirm python-pre-commit
    else
        error "Cannot install pre-commit automatically."
        echo "   Please install manually: https://pre-commit.com/#install"
        exit 1
    fi
    
    success "pre-commit installed"
else
    pre-commit --version
fi

# ═══════════════════════════════════════════════════════════════════
# 2. SECURITY TOOLS
# ═══════════════════════════════════════════════════════════════════
header "2. Security Scanning Tools"

# gitleaks
if ! check_cmd "gitleaks" "https://github.com/gitleaks/gitleaks#installing"; then
    info "Installing gitleaks..."
    
    if command -v brew &>/dev/null; then
        brew install gitleaks
    elif command -v go &>/dev/null; then
        go install github.com/gitleaks/gitleaks/v8@latest
    else
        # Fallback: download latest release
        local latest
        latest=$(curl -sL "https://api.github.com/repos/gitleaks/gitleaks/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/' || echo "8.18.2")
        
        local os arch
        os=$(uname -s | tr '[:upper:]' '[:lower:]')
        arch=$(uname -m)
        case "$arch" in
            x86_64)  arch="x64" ;;
            aarch64) arch="arm64" ;;
        esac
        
        curl -sL "https://github.com/gitleaks/gitleaks/releases/download/v${latest}/gitleaks_${latest}_${os}_${arch}.tar.gz" | \
            tar -xz -C /tmp/ gitleaks 2>/dev/null && \
            sudo mv /tmp/gitleaks /usr/local/bin/ 2>/dev/null || \
            mv /tmp/gitleaks "$HOME/.local/bin/" 2>/dev/null || true
    fi
    
    command -v gitleaks &>/dev/null && success "gitleaks installed" || warn "gitleaks install may have failed"
fi

# trufflehog
if ! check_cmd "trufflehog" "https://github.com/trufflesecurity/trufflehog#install"; then
    info "Installing trufflehog..."
    
    if command -v brew &>/dev/null; then
        brew install trufflesecurity/trufflehog/trufflehog
    elif command -v go &>/dev/null; then
        go install github.com/trufflesecurity/trufflehog/v3@latest
    else
        warn "Could not auto-install trufflehog. Install manually: https://github.com/trufflesecurity/trufflehog#install"
    fi
    
    command -v trufflehog &>/dev/null && success "trufflehog installed" || warn "trufflehog install may have failed"
fi

# semgrep
if ! check_cmd "semgrep" "https://semgrep.dev/docs/getting-started/"; then
    info "Installing semgrep..."
    
    if ! install_python_tool "semgrep" "semgrep"; then
        if command -v brew &>/dev/null; then
            brew install semgrep
        else
            python3 -m pip install --user semgrep || true
        fi
    fi
    
    command -v semgrep &>/dev/null && success "semgrep installed" || warn "semgrep install may have failed"
fi

# gosec
if command -v go &>/dev/null; then
    if ! check_cmd "gosec" "https://github.com/securego/gosec#install"; then
        info "Installing gosec..."
        go install github.com/securego/gosec/v2/cmd/gosec@latest
        
        command -v gosec &>/dev/null && success "gosec installed" || warn "gosec install may have failed"
    fi
else
    warn "Go not found — gosec skipped (Go hooks disabled)"
fi

# ═══════════════════════════════════════════════════════════════════
# 3. LANGUAGE-SPECIFIC TOOLS
# ═══════════════════════════════════════════════════════════════════
header "3. Language-Specific Tools"

# Go tools
if command -v go &>/dev/null; then
    info "Installing Go tools..."
    
    if ! command -v goimports &>/dev/null; then
        go install golang.org/x/tools/cmd/goimports@latest
    fi
    
    success "Go tools ready"
else
    warn "Go not found — Go hooks disabled"
fi

# Node.js / TypeScript tools
if command -v npm &>/dev/null || command -v node &>/dev/null; then
    info "Installing Node.js tools..."
    
    if [[ -f "package.json" ]]; then
        npm install 2>/dev/null || yarn install 2>/dev/null || true
    fi
    
    # Ensure eslint-plugin-security is available
    if [[ -f ".eslintrc.security.js" ]] && [[ ! -d "node_modules/eslint-plugin-security" ]]; then
        npm install --save-dev eslint-plugin-security @typescript-eslint/parser @typescript-eslint/eslint-plugin 2>/dev/null || true
    fi
    
    success "Node.js tools ready"
else
    warn "Node.js not found — JS/TS hooks disabled"
fi

# Rust tools
if command -v cargo &>/dev/null; then
    if ! command -v rustfmt &>/dev/null; then
        rustup component add rustfmt 2>/dev/null || true
    fi
    success "Rust tools ready"
else
    warn "Rust not found — Rust hooks disabled"
fi

# Dart tools
if command -v dart &>/dev/null || command -v flutter &>/dev/null; then
    success "Dart tools ready"
else
    warn "Dart not found — Dart hooks disabled"
fi

# ═══════════════════════════════════════════════════════════════════
# 4. INSTALL PRE-COMMIT HOOKS
# ═══════════════════════════════════════════════════════════════════
header "4. Installing Pre-commit Hooks"

if [[ ! -f ".pre-commit-config.yaml" ]]; then
    error ".pre-commit-config.yaml not found in project root"
    exit 1
fi

info "Installing hooks into .git/hooks/..."
pre-commit install --install-hooks
pre-commit install --hook-type pre-commit
pre-commit install --hook-type commit-msg 2>/dev/null || true

success "Hooks installed successfully"

# ═══════════════════════════════════════════════════════════════════
# 5. OPTIONAL: RUN ONCE OVER ALL FILES
# ═══════════════════════════════════════════════════════════════════
header "5. Initial Scan"

read -r -p "Run pre-commit on all files now? (may take a while) [y/N]: " response 2>/dev/null || response="n"

if [[ "$response" =~ ^[Yy]$ ]]; then
    info "Running pre-commit on all files..."
    pre-commit run --all-files || true
else
    info "Skipped. Run 'pre-commit run --all-files' anytime."
fi

# ═══════════════════════════════════════════════════════════════════
# 6. STATUS FILE
# ═══════════════════════════════════════════════════════════════════
header "6. Saving Status"

mkdir -p ".status"
cat > ".status/precommit.status" << 'EOF'
# Pre-commit hooks status — generated by scripts/setup-hooks.sh
# Last updated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
status: installed
tools:
  pre-commit: $(pre-commit --version 2>/dev/null || echo "unknown")
  gitleaks: $(gitleaks version 2>/dev/null || echo "not found")
  trufflehog: $(trufflehog --version 2>/dev/null || echo "not found")
  semgrep: $(semgrep --version 2>/dev/null || echo "not found")
  gosec: $(gosec --version 2>/dev/null | head -1 || echo "not found")
hooks: 
  - general
  - secret-scanning
  - sast
  - go-security
  - typescript-security
  - rust-security
  - dart-security
languages:
  - go
  - typescript
  - rust
  - dart
EOF

success "Status saved to .status/precommit.status"

# ═══════════════════════════════════════════════════════════════════
echo ""
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║   ✅  Pre-commit hooks setup complete!                            ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""
echo "  Next steps:"
echo "    • Hooks run automatically on git commit"
echo "    • Manual run: pre-commit run --all-files"
echo "    • Skip hooks: git commit --no-verify  (emergency only)"
echo "    • Update hooks: pre-commit autoupdate"
echo ""
