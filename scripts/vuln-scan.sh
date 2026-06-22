#!/bin/bash
# Vulnerability Scan Script
# SSH Manager Vulnerability Scanning

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPORTS_DIR="$PROJECT_DIR/reports/vuln"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${BLUE}[SCAN]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }
success() { echo -e "${GREEN}[PASS]${NC} $1"; }

# Create reports directory
mkdir -p "$REPORTS_DIR"

SCAN_DATE=$(date -Iseconds)
REPORT_FILE="$REPORTS_DIR/vuln-scan-$(date +%Y%m%d-%H%M%S).md"

echo "# Vulnerability Scan Report" > "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "**Date:** $SCAN_DATE" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

log "Starting vulnerability scan..."

# Check 1: Snyk (if available)
if command -v snyk &> /dev/null; then
    log "Running Snyk scan..."
    cd "$PROJECT_DIR"
    snyk test --json > "$REPORTS_DIR/snyk.json" 2>/dev/null || true
    if [ -f "$REPORTS_DIR/snyk.json" ] && [ -s "$REPORTS_DIR/snyk.json" ]; then
        VULN_COUNT=$(jq 'length' "$REPORTS_DIR/snyk.json" 2>/dev/null || echo "0")
        if [ "$VULN_COUNT" -gt 0 ]; then
            warn "Snyk found $VULN_COUNT vulnerabilities"
            echo "## Snyk Scan" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
            echo "$VULN_COUNT vulnerabilities found" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
        else
            success "Snyk scan passed"
        fi
    fi
else
    warn "Snyk not installed, skipping"
fi

# Check 2: Trivy (if available)
if command -v trivy &> /dev/null; then
    log "Running Trivy scan..."
    
    # Filesystem scan
    trivy fs --format json --output "$REPORTS_DIR/trivy-fs.json" "$PROJECT_DIR" 2>/dev/null || true
    if [ -f "$REPORTS_DIR/trivy-fs.json" ] && [ -s "$REPORTS_DIR/trivy-fs.json" ]; then
        VULN_COUNT=$(jq '.Results | length' "$REPORTS_DIR/trivy-fs.json" 2>/dev/null || echo "0")
        if [ "$VULN_COUNT" -gt 0 ]; then
            warn "Trivy found $VULN_COUNT results"
            echo "## Trivy Filesystem Scan" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
            echo "$VULN_COUNT results" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
        else
            success "Trivy filesystem scan passed"
        fi
    fi
    
    # Docker image scan
    if [ -f "$PROJECT_DIR/Dockerfile" ]; then
        log "Scanning Docker image..."
        trivy image --format json --output "$REPORTS_DIR/trivy-image.json" ssh-manager:latest 2>/dev/null || true
        if [ -f "$REPORTS_DIR/trivy-image.json" ] && [ -s "$REPORTS_DIR/trivy-image.json" ]; then
            VULN_COUNT=$(jq '.Results | length' "$REPORTS_DIR/trivy-image.json" 2>/dev/null || echo "0")
            if [ "$VULN_COUNT" -gt 0 ]; then
                warn "Trivy image scan found $VULN_COUNT results"
                echo "## Trivy Image Scan" >> "$REPORT_FILE"
                echo "" >> "$REPORT_FILE"
                echo "$VULN_COUNT results" >> "$REPORT_FILE"
                echo "" >> "$REPORT_FILE"
            else
                success "Trivy image scan passed"
            fi
        fi
    fi
else
    warn "Trivy not installed, skipping"
fi

# Check 3: OWASP Dependency Check (if available)
if command -v dependency-check &> /dev/null; then
    log "Running OWASP Dependency Check..."
    dependency-check --project "SSH Manager" --scan "$PROJECT_DIR" --format JSON --out "$REPORTS_DIR/owasp.json" 2>/dev/null || true
    if [ -f "$REPORTS_DIR/owasp.json" ] && [ -s "$REPORTS_DIR/owasp.json" ]; then
        VULN_COUNT=$(jq '.dependencies | map(.vulnerabilities | length) | add' "$REPORTS_DIR/owasp.json" 2>/dev/null || echo "0")
        if [ "$VULN_COUNT" -gt 0 ]; then
            warn "OWASP found $VULN_COUNT vulnerabilities"
            echo "## OWASP Dependency Check" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
            echo "$VULN_COUNT vulnerabilities found" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
        else
            success "OWASP scan passed"
        fi
    fi
else
    warn "OWASP Dependency Check not installed, skipping"
fi

# Check 4: npm audit (Node.js)
if [ -f "$PROJECT_DIR/package.json" ] && command -v npm &> /dev/null; then
    log "Running npm audit..."
    cd "$PROJECT_DIR"
    npm audit --json > "$REPORTS_DIR/npm-audit.json" 2>/dev/null || true
    if [ -f "$REPORTS_DIR/npm-audit.json" ] && [ -s "$REPORTS_DIR/npm-audit.json" ]; then
        VULN_COUNT=$(jq '.metadata.vulnerabilities.total // 0' "$REPORTS_DIR/npm-audit.json")
        if [ "$VULN_COUNT" -gt 0 ]; then
            warn "npm audit found $VULN_COUNT vulnerabilities"
            echo "## npm Audit" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
            echo "$VULN_COUNT vulnerabilities found" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
        else
            success "npm audit passed"
        fi
    fi
fi

# Check 5: Go vulnerability check
if [ -f "$PROJECT_DIR/go.mod" ] && command -v govulncheck &> /dev/null; then
    log "Running govulncheck..."
    cd "$PROJECT_DIR"
    govulncheck ./... > "$REPORTS_DIR/govulncheck.txt" 2>/dev/null || true
    if [ -f "$REPORTS_DIR/govulncheck.txt" ] && [ -s "$REPORTS_DIR/govulncheck.txt" ]; then
        if grep -q "Vulnerability" "$REPORTS_DIR/govulncheck.txt"; then
            warn "govulncheck found vulnerabilities"
            echo "## Go Vulnerability Check" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
            echo "```" >> "$REPORT_FILE"
            cat "$REPORTS_DIR/govulncheck.txt" >> "$REPORT_FILE"
            echo "```" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
        else
            success "govulncheck passed"
        fi
    fi
fi

# Check 6: Container scan with Grype (if available)
if command -v grype &> /dev/null; then
    log "Running Grype scan..."
    grype dir:"$PROJECT_DIR" --output json > "$REPORTS_DIR/grype.json" 2>/dev/null || true
    if [ -f "$REPORTS_DIR/grype.json" ] && [ -s "$REPORTS_DIR/grype.json" ]; then
        VULN_COUNT=$(jq '.matches | length' "$REPORTS_DIR/grype.json" 2>/dev/null || echo "0")
        if [ "$VULN_COUNT" -gt 0 ]; then
            warn "Grype found $VULN_COUNT vulnerabilities"
            echo "## Grype Scan" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
            echo "$VULN_COUNT vulnerabilities found" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
        else
            success "Grype scan passed"
        fi
    fi
else
    warn "Grype not installed, skipping"
fi

# Check 7: Checkov (IaC scanning)
if command -v checkov &> /dev/null; then
    log "Running Checkov..."
    checkov --directory "$PROJECT_DIR" --output json --soft-fail > "$REPORTS_DIR/checkov.json" 2>/dev/null || true
    if [ -f "$REPORTS_DIR/checkov.json" ] && [ -s "$REPORTS_DIR/checkov.json" ]; then
        VULN_COUNT=$(jq '.summary.failed // 0' "$REPORTS_DIR/checkov.json" 2>/dev/null || echo "0")
        if [ "$VULN_COUNT" -gt 0 ]; then
            warn "Checkov found $VULN_COUNT issues"
            echo "## Checkov (IaC)" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
            echo "$VULN_COUNT issues found" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
        else
            success "Checkov scan passed"
        fi
    fi
else
    warn "Checkov not installed, skipping"
fi

# Summary
echo "## Summary" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "Vulnerability scan completed at $SCAN_DATE" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "Reports saved to: $REPORTS_DIR" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "## Next Steps" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "1. Review findings" >> "$REPORT_FILE"
echo "2. Prioritize remediation" >> "$REPORT_FILE"
echo "3. Apply patches" >> "$REPORT_FILE"
echo "4. Re-run scan" >> "$REPORT_FILE"
echo "5. Monitor for new vulnerabilities" >> "$REPORT_FILE"

success "Vulnerability scan complete!"
log "Report saved to: $REPORT_FILE"
