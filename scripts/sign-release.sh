#!/bin/bash
set -euo pipefail

cd dist

echo "Signing with Cosign..."
for f in *.tar.gz *.zip *.msi *.dmg *.AppImage; do
  [[ -f "$f" ]] || continue
  cosign sign-blob "$f" --output-signature "$f.sig" --yes
done

echo "Signing checksums..."
cosign sign-blob checksums.txt --output-signature checksums.txt.sig --yes

echo "Signing containers..."
for img in $(docker images --format '{{.Repository}}:{{.Tag}}' | grep ssh-manager); do
  cosign sign "$img" --yes
done

echo "Done."
