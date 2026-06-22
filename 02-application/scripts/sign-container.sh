#!/bin/bash
set -euo pipefail

IMAGE="$1"

echo "Signing container $IMAGE..."
cosign sign "$IMAGE" --yes
echo "Done."
