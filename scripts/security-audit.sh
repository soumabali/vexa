#!/bin/bash
# Automated Security Audit Script
# SSH Manager Security Audit

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPORTS_DIR="$PROJECT_DIR/reports/security"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${BLUE}[AUDIT]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }
success() { echo -e "${GREEN}[PASS]${NC} $1"; }

# Create reports directory
mkdir -p "$REPORTS_DIR"

# Audit date
AUDIT_DATE=$(date -Iseconds)
REPORT_FILE="$REPORTS_DIR/audit-$(date +%Y%m%d-%H%M%S).md"

echo "# Security Audit Report" > "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "**Date:** $AUDIT_DATE" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

log "Starting security audit..."

# Check 1: File permissions
log "Checking file permissions..."
BAD_PERMS=$(find "$PROJECT_DIR" -type f \( -perm -002 -o -perm -020 \) | grep -v node_modules | grep -v .git || true)
if [ -n "$BAD_PERMS" ]; then
    warn "Found files with bad permissions:"
    echo "$BAD_PERMS"
    echo "## File Permissions" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "```" >> "$REPORT_FILE"
    echo "$BAD_PERMS" >> "$REPORT_FILE"
    echo "```" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
else
    success "File permissions OK"
fi

# Check 2: Secret scanning
log "Scanning for secrets..."
if command -v gitleaks &> /dev/null; then
    gitleaks detect --source "$PROJECT_DIR" --report-format json --report-path "$REPORTS_DIR/gitleaks.json" || true
    if [ -f "$REPORTS_DIR/gitleaks.json" ] && [ -s "$REPORTS_DIR/gitleaks.json" ]; then
        warn "Gitleaks found potential secrets"
        echo "## Secret Scanning" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "Potential secrets found by gitleaks" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    else
        success "No secrets found"
    fi
else
    warn "gitleaks not installed, skipping secret scan"
fi

# Check 3: Dependency vulnerabilities
log "Checking dependencies..."

# Go dependencies
if [ -f "$PROJECT_DIR/go.mod" ]; then
    if command -v govulncheck &> /dev/null; then
        cd "$PROJECT_DIR"
        govulncheck ./... > "$REPORTS_DIR/govulncheck.txt" || true
        if grep -q "Vulnerability" "$REPORTS_DIR/govulncheck.txt"; then
            warn "Go vulnerabilities found"
            echo "## Go Dependencies" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
            echo "```" >> "$REPORT_FILE"
            cat "$REPORTS_DIR/govulncheck.txt" >> "$REPORT_FILE"
            echo "```" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
        else
            success "Go dependencies OK"
        fi
    fi
fi

# Node dependencies
if [ -f "$PROJECT_DIR/package.json" ]; then
    if command -v npm &> /dev/null; then
        cd "$PROJECT_DIR"
        npm audit --json > "$REPORTS_DIR/npm-audit.json" || true
        if [ -f "$REPORTS_DIR/npm-audit.json" ]; then
            VULN_COUNT=$(jq '.metadata.vulnerabilities.total // 0' "$REPORTS_DIR/npm-audit.json")
            if [ "$VULN_COUNT" -gt 0 ]; then
                warn "$VULN_COUNT npm vulnerabilities found"
                echo "## Node Dependencies" >> "$REPORT_FILE"
                echo "" >> "$REPORT_FILE"
                echo "$VULN_COUNT vulnerabilities found" >> "$REPORT_FILE"
                echo "" >> "$REPORT_FILE"
            else
                success "Node dependencies OK"
            fi
        fi
    fi
fi

# Check 4: Static analysis
log "Running static analysis..."

# Go vet
if [ -f "$PROJECT_DIR/go.mod" ]; then
    cd "$PROJECT_DIR"
    go vet ./... > "$REPORTS_DIR/govet.txt" 2>> "$REPORTS_DIR/govet.txt" || true
    if [ -s "$REPORTS_DIR/govet.txt" ]; then
        warn "Go vet found issues"
        echo "## Static Analysis" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "```" >> "$REPORT_FILE"
        cat "$REPORTS_DIR/govet.txt" >> "$REPORT_FILE"
        echo "```" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    else
        success "Go vet passed"
    fi
fi

# Check 5: Docker security
if [ -f "$PROJECT_DIR/Dockerfile" ]; then
    log "Checking Docker security..."
    if command -v hadolint &> /dev/null; then
        hadolint "$PROJECT_DIR/Dockerfile" > "$REPORTS_DIR/hadolint.txt" || true
        if [ -s "$REPORTS_DIR/hadolint.txt" ]; then
            warn "Dockerfile issues found"
            echo "## Docker Security" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
            echo "```" >> "$REPORT_FILE"
            cat "$REPORTS_DIR/hadolint.txt" >> "$REPORT_FILE"
            echo "```" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
        else
            success "Dockerfile OK"
        fi
    fi
fi

# Check 6: SSL/TLS configuration
log "Checking TLS configuration..."
if [ -f "$PROJECT_DIR/apps/api/internal/security/hardening.go" ]; then
    if grep -q "VersionTLS13" "$PROJECT_DIR/apps/api/internal/security/hardening.go"; then
        success "TLS 1.3 configured"
    else
        warn "TLS 1.3 not found"
    fi
fi

# Check 7: Security headers
log "Checking security headers..."
if [ -f "$PROJECT_DIR/apps/api/internal/security/headers.go" ]; then
    HEADERS=("Strict-Transport-Security" "X-Frame-Options" "X-Content-Type-Options" "Content-Security-Policy")
    for header in "${HEADERS[@]}"; do
        if grep -q "$header" "$PROJECT_DIR/apps/api/internal/security/headers.go"; then
            success "$header header configured"
        else
            warn "$header header missing"
        fi
    done
fi

# Check 8: Rate limiting
log "Checking rate limiting..."
if [ -f "$PROJECT_DIR/apps/api/internal/security/rate_limit.go" ]; then
    success "Rate limiting implemented"
else
    warn "Rate limiting not found"
fi

# Check 9: WAF
log "Checking WAF..."
if [ -f "$PROJECT_DIR/apps/api/internal/security/waf.go" ]; then
    success "WAF implemented"
else
    warn "WAF not found"
fi

# Check 10: Audit logging
log "Checking audit logging..."
if [ -f "$PROJECT_DIR/apps/api/internal/security/audit.go" ]; then
    success "Audit logging implemented"
else
    warn "Audit logging not found"
fi

# Check 11: Encryption
log "Checking encryption..."
ENCRYPTED_FILES=$(find "$PROJECT_DIR" -type f \( -name "*.enc" -o -name "*.gpg" -o -name "*.age" \) | wc -l)
if [ "$ENCRYPTED_FILES" -gt 0 ]; then
    success "Encrypted files found"
else
    warn "No encrypted files found"
fi

# Check 12: Test coverage
log "Checking test coverage..."
if [ -f "$PROJECT_DIR/go.mod" ]; then
    cd "$PROJECT_DIR"
    COVERAGE=$(go test -cover ./... 2>/dev/null | grep -oP '\d+(\.\d+)?%' | tail -1 || echo "0%")
    log "Test coverage: $COVERAGE"
    echo "## Test Coverage" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "Coverage: $COVERAGE" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
fi

# Summary
echo "## Summary" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "Security audit completed at $AUDIT_DATE" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "Reports saved to: $REPORTS_DIR" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "## Next Steps" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "1. Review findings" >> "$REPORT_FILE"
echo "2. Prioritize remediation" >> "$REPORT_FILE"
echo "3. Assign owners" >> "$REPORT_FILE"
echo "4. Set deadlines" >> "$REPORT_FILE"
echo "5. Re-run audit" >> "$REPORT_FILE"

success "Security audit complete!"
log "Report saved to: $REPORT_FILE"
