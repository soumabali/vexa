#!/bin/bash
set -euo pipefail

VERSION="$(git describe --tags --always)-beta"
echo "Building beta $VERSION..."

./scripts/release.sh "$VERSION"
./scripts/sign-release.sh --beta

echo "Beta build ready: dist/"
