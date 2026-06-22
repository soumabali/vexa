#!/bin/bash
set -euo pipefail

FILE="$1"

echo "Signing $FILE with GPG..."
gpg --detach-sign --armor "$FILE"
echo "Done."
