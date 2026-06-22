#!/bin/bash
set -euo pipefail

FILE="$1"
IDENTITY="${APPLE_DEVELOPER_ID:-Developer ID Application}"

echo "Signing $FILE with $IDENTITY..."
codesign --force --options runtime --sign "$IDENTITY" --timestamp "$FILE"
echo "Done."
