#!/bin/bash
set -euo pipefail

VERSION="${1:-$(git describe --tags --always)}"
PREV="${2:-$(git describe --tags --always --abbrev=0 HEAD^)}"

echo "# Changelog $VERSION"
echo ""
echo "## Changes since $PREV"
echo ""
git log "$PREV..HEAD" --pretty=format:"- %s (%h)" --reverse

echo ""
echo "## Contributors"
git log "$PREV..HEAD" --pretty=format:"%an" | sort -u | sed 's/^/- /'
