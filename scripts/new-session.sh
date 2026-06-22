#!/usr/bin/env bash
# new-session.sh — Create a new session note and update indexes
# Usage: new-session.sh "Topic of the session"

set -euo pipefail

ROOT_DIR="/home/ubuntu/projects/vexa"
SESSIONS_DIR="${ROOT_DIR}/03-history/sessions"
OBSIDIAN_DIR="/home/ubuntu/Documents/Obsidian Vault/infra"

if [ $# -lt 1 ]; then
  echo "Usage: $0 \"Topic of the session\""
  exit 1
fi

TOPIC="$1"
DATE=$(date +%Y-%m-%d)
DATETIME=$(date +%Y%m%d-%H%M%S)
SLUG=$(echo "$TOPIC" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -cd '[:alnum:]-')
FILENAME="${DATE}-${SLUG}.md"
FILEPATH="${SESSIONS_DIR}/${FILENAME}"

if [ -f "$FILEPATH" ]; then
  echo "ERROR: session note already exists: $FILEPATH"
  exit 1
fi

cat > "$FILEPATH" <<EOF
# ${DATE} — ${TOPIC}

## Goal

<!-- Deskripsikan goal session ini -->

## Changes

<!-- List perubahan penting -->

## Verification

<!-- Gate results -->

## Issues / Blockers

<!-- Jika ada -->

## Links

- Plan: \`06-temp/plans/YYYY-MM-DD-<slug>.md\`
- Obsidian: [[vexa Progress]]
- Obsidian: [[vexa Workflow Runbook]]
EOF

# Update session index
README="${SESSIONS_DIR}/README.md"
if [ -f "$README" ]; then
  # Add new entry to the 2026 table
  sed -i "/^## 2026/,/^## Template/{ /^| 2026-06-/!{ /^| /a\\
| ${DATE} | \`${FILENAME}\` | ${TOPIC} | 🚧 In Progress |
}
}" "$README"
  echo "Updated: $README"
fi

# Update Obsidian Progress.md: add in-progress item
OBS_PROGRESS="${OBSIDIAN_DIR}/vexa Progress.md"
if [ -f "$OBS_PROGRESS" ]; then
  sed -i "/## Yang Sedang Berjalan \/ Next/a\\
- 🚧 ${TOPIC} (\`03-history/sessions/${FILENAME}\`)" "$OBS_PROGRESS"
  echo "Updated: $OBS_PROGRESS"
fi

echo "Created session note: $FILEPATH"
echo "Next steps:"
echo "  1. Fill in the session note."
echo "  2. Update plan index if needed."
echo "  3. Commit with: git add -A && git commit -m \"docs: ${TOPIC}\""
