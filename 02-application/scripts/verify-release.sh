#!/bin/bash
set -euo pipefail

FILE="$1"
SIG="$2"

echo "Verifying $FILE..."
cosign verify-blob "$FILE" --signature "$SIG"
echo "OK."
