#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════
# Suspicious Pattern Checker — Custom Security Hook
# ═══════════════════════════════════════════════════════════════════
set -euo pipefail

RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

ERRORS=0
WARNINGS=0

# Patterns that are outright dangerous
DANGEROUS_PATTERNS=(
    "eval\s*\(.*\$.*\)"
    "exec\s*\(.*\$.*\)"
    "system\s*\(.*\$.*\)"
    "os\.system\s*\(.*\$.*\)"
    "subprocess\.call\s*\(.*shell\s*=\s*True.*\)"
    "subprocess\.Popen\s*\(.*shell\s*=\s*True.*\)"
    "Runtime\.getRuntime\(\)\.exec\s*\(.*\$.*\)"
    "dangerouslySetInnerHTML"
    "innerHTML\s*=.*\$"
    "document\.write\s*\(.*\$.*\)"
    "\.html\s*\(.*\$.*\)"
    "SQL Injection"
    "mysql_query\s*\(.*\$.*\)"
    "mysqli_query\s*\(.*\$.*\)"
    "PDO::query\s*\(.*\$.*\)"
    "sqlite3_exec\s*\(.*\$.*\)"
    "exec\s*\(\s*[\"']rm\s+-rf\s+/[\"']"
    "eval\s*\(.*base64_decode.*\)"
    "create_function\s*\(.*\$.*\)"
    "assert\s*\(.*\$.*\)"
    "preg_replace\s*\(.*\/e.*\$.*\)"
    "passthru\s*\(.*\$.*\)"
    "shell_exec\s*\(.*\$.*\)"
    "popen\s*\(.*\$.*\)"
    "proc_open\s*\(.*\$.*\)"
    "pcntl_exec\s*\(.*\$.*\)"
    "file_get_contents\s*\(.*\$_GET.*\)"
    "file_put_contents\s*\(.*\$_GET.*\)"
    "include\s*\(.*\$_GET.*\)"
    "require\s*\(.*\$_GET.*\)"
    "require_once\s*\(.*\$_GET.*\)"
    "import\s*\(.*\$_GET.*\)"
    "os\.StartProcess"
    "syscall\.Exec"
    "unsafe\."
)

# Patterns that are suspicious and need review
SUSPICIOUS_PATTERNS=(
    "TODO.*security"
    "FIXME.*security"
    "HACK.*"
    "XXX.*"
    "password\s*=\s*[\"'][^\"']+[\"']"
    "passwd\s*=\s*[\"'][^\"']+[\"']"
    "api_key\s*=\s*[\"'][^\"']+[\"']"
    "apikey\s*=\s*[\"'][^\"']+[\"']"
    "secret\s*=\s*[\"'][^\"']+[\"']"
    "token\s*=\s*[\"'][^\"']+[\"']"
    "BEGIN\s+(RSA|DSA|EC|OPENSSH)\s+PRIVATE\s+KEY"
    "AKIA[0-9A-Z]{16}"
    "ghp_[a-zA-Z0-9]{36}"
    "gho_[a-zA-Z0-9]{36}"
    "glpat-[a-zA-Z0-9\-]{20}"
    "sk_live_[0-9a-zA-Z]{24}"
    "sk_test_[0-9a-zA-Z]{24}"
    "[a-f0-9]{32}-us[0-9]{2}"
    "xox[baprs]-[0-9]{10,13}-[0-9]{10,13}[a-zA-Z0-9]*"
    "sk-[a-zA-Z0-9]{20,}T3BlbkFJ[a-zA-Z0-9]{20,}"
)

scan_file() {
    local file="$1"
    local line_num=0
    
    while IFS= read -r line || [[ -n "$line" ]]; do
        ((line_num++))
        
        # Skip comments and strings (basic heuristic)
        local stripped="$line"
        stripped=$(echo "$stripped" | sed -E 's/^[[:space:]]*[#//\*]+.*//' | sed -E 's/"[^"]*"//g' | sed -E "s/'[^']*'//g")
        
        for pattern in "${DANGEROUS_PATTERNS[@]}"; do
            if echo "$stripped" | grep -qiP "$pattern" 2>/dev/null; then
                echo -e "${RED}[BLOCKED]${NC} $file:$line_num — dangerous pattern: $pattern"
                echo "  $line"
                ((ERRORS++))
            fi
        done
        
        for pattern in "${SUSPICIOUS_PATTERNS[@]}"; do
            if echo "$stripped" | grep -qiP "$pattern" 2>/dev/null; then
                echo -e "${YELLOW}[WARNING]${NC} $file:$line_num — suspicious pattern: $pattern"
                echo "  $line"
                ((WARNINGS++))
            fi
        done
    done < "$file"
}

# ─── Main ───────────────────────────────────────────────────────────
info() {
    echo -e "${GREEN}ℹ${NC} $*"
}

echo "Scanning for suspicious patterns..."

# Find relevant files, excluding common non-source directories
while IFS= read -r -d '' file; do
    # Skip binary files
    if file "$file" | grep -q "text"; then
        scan_file "$file"
    fi
done < <( \
    git ls-files -z 2>/dev/null || \
    find . -type f \( \
        -name "*.go" -o -name "*.js" -o -name "*.jsx" -o \
        -name "*.ts" -o -name "*.tsx" -o -name "*.rs" -o \
        -name "*.dart" -o -name "*.py" -o -name "*.sh" -o \
        -name "*.php" -o -name "*.java" -o -name "*.kt" \
    \) ! -path "*/vendor/*" ! -path "*/node_modules/*" \
       ! -path "*/target/*" ! -path "*/build/*" \
       ! -path "*/dist/*" ! -path "*/.dart_tool/*" \
       -print0 \
)

echo ""
if [[ $ERRORS -gt 0 ]]; then
    echo -e "${RED}Found $ERRORS dangerous pattern(s)${NC}"
    echo -e "${YELLOW}Found $WARNINGS suspicious pattern(s)${NC}"
    exit 1
elif [[ $WARNINGS -gt 0 ]]; then
    echo -e "${GREEN}No dangerous patterns found${NC}"
    echo -e "${YELLOW}Found $WARNINGS suspicious pattern(s) — please review${NC}"
    exit 0  # Don't block on warnings alone
else
    echo -e "${GREEN}✓ No suspicious patterns found${NC}"
    exit 0
fi
