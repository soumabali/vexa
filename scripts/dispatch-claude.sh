#!/usr/bin/env bash
# dispatch-claude.sh — Safe Claude Code dispatcher for vexa
# Usage: dispatch-claude.sh /path/to/plan.md [slug]
# Output: writes structured log and returns exit code from compliance check.

set -euo pipefail

ROOT_DIR="/home/ubuntu/projects/vexa"
PLAN_FILE="${1:-}"
SLUG="${2:-$(basename "$PLAN_FILE" .md)}"
DATE_PREFIX=$(date +%Y%m%d-%H%M%S)
LOG_DIR="${ROOT_DIR}/03-history/sessions/run-logs"
LOG_FILE="${LOG_DIR}/${DATE_PREFIX}-${SLUG}.log"
RESPONSE_FILE="${LOG_DIR}/${DATE_PREFIX}-${SLUG}.response.json"
SCHEMA="${ROOT_DIR}/00-meta/claude-response-schema.json"

if [ -z "$PLAN_FILE" ] || [ ! -f "$PLAN_FILE" ]; then
  echo "Usage: $0 /path/to/plan.md [slug]" >&2
  exit 1
fi

mkdir -p "$LOG_DIR"

echo "[dispatch] $(date) Starting dispatch for plan: $PLAN_FILE" | tee -a "$LOG_FILE"
echo "[dispatch] Log: $LOG_FILE" | tee -a "$LOG_FILE"

# Read plan
PLAN_CONTENT=$(cat "$PLAN_FILE")

# Build prompt enforcing JSON response
PROMPT=$(cat <<EOF
AGENT COMPLIANCE CHECK — vexa project.

Before doing anything, you MUST confirm in your response that you have read:
1. AGENTS.md
2. Root CLAUDE.md
3. 00-meta/skills.md
4. 00-meta/git-structure.md
5. 02-application/CLAUDE.md
6. The active plan below.

Then execute the plan. You may ONLY edit files inside /home/ubuntu/projects/vexa/02-application/. You must NOT commit, push, or edit files outside 02-application/. You must use skills superpowers, caveman, and graphify. If any E2E/UI test changes are needed, also use the playwright MCP.

IMPORTANT: After execution, respond ONLY with a single valid JSON object conforming exactly to /home/ubuntu/projects/vexa/00-meta/claude-response-schema.json. Do not include markdown fences, explanations, or any text outside the JSON object.

Plan:
$PLAN_CONTENT
EOF
)

# Run Claude Code with timeout, capturing all output
# First attempt
echo "[dispatch] $(date) Attempt 1/2..." | tee -a "$LOG_FILE"
if ! timeout 1200 ollama launch claude --model kimi-k2.7-code:cloud -- \
  -p "$PROMPT" \
  --print \
  --allowedTools "Read,Write,Edit,Bash(go test),Bash(go build),Bash(npm run build),Bash(make),Bash(cp),Bash(rm -f),Bash(mkdir),Bash(ls),Bash(grep),Bash(cd)" \
  > "$RESPONSE_FILE" 2>> "$LOG_FILE"; then
  echo "[dispatch] $(date) Attempt 1 failed or timed out." | tee -a "$LOG_FILE"
  echo "[dispatch] $(date) Attempt 2/2..." | tee -a "$LOG_FILE"
  if ! timeout 1200 ollama launch claude --model kimi-k2.7-code:cloud -- \
    -p "$PROMPT" \
    --print \
    --allowedTools "Read,Write,Edit,Bash(go test),Bash(go build),Bash(npm run build),Bash(make),Bash(cp),Bash(rm -f),Bash(mkdir),Bash(ls),Bash(grep),Bash(cd)" \
    > "$RESPONSE_FILE" 2>> "$LOG_FILE"; then
    echo "[dispatch] $(date) Attempt 2 also failed. Giving up." | tee -a "$LOG_FILE"
    exit 1
  fi
fi

echo "[dispatch] $(date) Claude Code finished. Response saved to $RESPONSE_FILE" | tee -a "$LOG_FILE"
echo "[dispatch] $(date) Response size: $(wc -c < "$RESPONSE_FILE") bytes" | tee -a "$LOG_FILE"

# Validate response
if [ -s "$RESPONSE_FILE" ]; then
  echo "[dispatch] $(date) Validating structured response..." | tee -a "$LOG_FILE"
  if "${ROOT_DIR}/scripts/claude-compliance-check.sh" "$RESPONSE_FILE" >> "$LOG_FILE" 2>&1; then
    echo "[dispatch] $(date) ✅ Compliance validation PASSED" | tee -a "$LOG_FILE"
    exit 0
  else
    echo "[dispatch] $(date) ❌ Compliance validation FAILED" | tee -a "$LOG_FILE"
    exit 2
  fi
else
  echo "[dispatch] $(date) ❌ Response file empty" | tee -a "$LOG_FILE"
  exit 3
fi
