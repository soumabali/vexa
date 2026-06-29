#!/usr/bin/env bash
# Script: scripts/vexa-cleanup.sh
# Purpose: Remove embedded .git directories (except the repository root)
#          and produce a clean git history snapshot.

set -euo pipefail

# Find all .git directories under the repo, excluding the root one
echo "🔍 Searching for embedded .git directories..."
FOUND_GIT_DIRS=()
while IFS= read -r dir; do
    # Exclude the repository root .git
    if [[ "$dir" != "/home/ubuntu/projects/vexa/.git" ]]; then
        FOUND_GIT_DIRS+=("$dir")
        echo "  - Removing $dir"
        rm -rf "$dir"
    fi
done < <(find /home/ubuntu/projects/vexa -type d -name .git -print)

if [[ ${#FOUND_GIT_DIRS[@]} -eq 0 ]]; then
    echo "✅ No embedded .git directories found."
else
    echo "🗑️  Removed ${#FOUND_GIT_DIRS[@]} .git directories."
fi

# Verify git status is clean
echo "🔎 Checking git status..."
git status --porcelain

# Optional: Create a commit with the cleanup changes
# Uncomment the following lines if you want to commit automatically
# git add .
# git commit -m "chore: remove embedded .git directories for clean history"
# echo "📦 Commit created for cleanup."

echo "✨ Cleanup completed."