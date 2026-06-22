#!/usr/bin/env bash
# rollback-last.sh — Interactive rollback helper for vexa root or application subtree
# Usage: rollback-last.sh [--root | --app]

set -euo pipefail

ROOT_DIR="/home/ubuntu/projects/vexa"
APP_DIR="${ROOT_DIR}/02-application"

TARGET="${1:-}"

if [ -z "$TARGET" ] || { [ "$TARGET" != "--root" ] && [ "$TARGET" != "--app" ]; }; then
  echo "Usage: $0 [--root | --app]"
  echo ""
  echo "  --root  Rollback the root orchestrator repo (vexa-root)"
  echo "  --app   Rollback the application subtree (vexa)"
  exit 1
fi

cd "$ROOT_DIR"

if [ "$TARGET" == "--root" ]; then
  echo "Root orchestrator last commit:"
  git log --oneline -1
  echo ""
  echo "Choose action:"
  echo "  1) git revert --no-edit HEAD (safe, keeps history)"
  echo "  2) git reset --soft HEAD~1 (remove last commit, keep changes staged)"
  echo "  3) Cancel"
  read -rp "Enter choice (1/2/3): " CHOICE
  case "$CHOICE" in
    1)
      git revert --no-edit HEAD
      echo "✅ Root reverted. Now push manually if needed."
      ;;
    2)
      git reset --soft HEAD~1
      echo "✅ Root last commit removed (changes staged). Review before recommit."
      ;;
    *)
      echo "Cancelled."
      exit 0
      ;;
  esac
else
  echo "Application subtree last pushed commit:"
  git log app-vexa/main --oneline -1 2>/dev/null || echo "Could not read app-vexa/main"
  echo ""
  echo "WARNING: Rolling back a pushed subtree commit is complex."
  echo "Recommended: create a new commit in 02-application/ that undoes the change,"
  echo "then subtree push again."
  echo ""
  echo "Choose action:"
  echo "  1) Show instructions for manual subtree revert"
  echo "  2) Cancel"
  read -rp "Enter choice (1/2): " CHOICE
  case "$CHOICE" in
    1)
      echo "Manual subtree rollback steps:"
      echo "  1. cd ${APP_DIR}"
      echo "  2. git revert <bad-commit-hash> (this creates a fix commit)"
      echo "  3. cd ${ROOT_DIR}"
      echo "  4. git subtree push --prefix=02-application https://github.com/soumabali/vexa.git main"
      ;;
    *)
      echo "Cancelled."
      exit 0
      ;;
  esac
fi
