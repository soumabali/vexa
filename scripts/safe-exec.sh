#!/usr/bin/env bash
# safe-exec.sh — Gate dangerous commands before execution
# Usage: safe-exec.sh "command string"
# If the command matches a dangerous pattern, this script prompts for confirmation.

set -euo pipefail

COMMAND="${1:-}"

if [ -z "$COMMAND" ]; then
  echo "Usage: $0 \"command to run\""
  exit 1
fi

# Patterns that require explicit human confirmation
DANGEROUS_PATTERNS=(
  "rm -rf"
  "rm -rf /"
  "dd if"
  "mkfs"
  "git push --force"
  "git push -f"
  "git reset --hard"
  "git revert --abort"
  "docker system prune"
  "docker volume rm"
  "docker compose.*down -v"
  "DROP TABLE"
  "DELETE FROM"
  "TRUNCATE"
  "sudo"
  "chmod 777 /"
  "chown -R root"
  "iptables -F"
)

MATCHED=0
for pattern in "${DANGEROUS_PATTERNS[@]}"; do
  if echo "$COMMAND" | grep -qiE "$pattern"; then
    MATCHED=1
    echo "⚠️  DANGEROUS PATTERN MATCHED: $pattern"
  fi
done

if [ "$MATCHED" -eq 1 ]; then
  echo ""
  echo "Command: $COMMAND"
  echo ""
  echo "This command is potentially destructive. Type 'YES' to proceed."
  read -r CONFIRM
  if [ "$CONFIRM" != "YES" ]; then
    echo "❌ Aborted by user."
    exit 1
  fi
fi

echo "[safe-exec] Running: $COMMAND"
eval "$COMMAND"
