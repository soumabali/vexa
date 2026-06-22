#!/bin/bash
set -euo pipefail

VERSION="${1:-$(git describe --tags --always)}"
OUT="dist"

echo "Building release $VERSION..."

rm -rf "$OUT"
mkdir -p "$OUT"

# API
echo "Building API..."
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o "$OUT/api-linux-amd64" ./cmd/api
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$VERSION" -o "$OUT/api-linux-arm64" ./cmd/api
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o "$OUT/api-darwin-amd64" ./cmd/api
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$VERSION" -o "$OUT/api-darwin-arm64" ./cmd/api

# Web
echo "Building Web..."
cd apps/web && npm run build && cd ../..
cp -r apps/web/dist "$OUT/web"

# Desktop
echo "Building Desktop..."
cd apps/desktop && npm run tauri build && cd ../..
cp apps/desktop/src-tauri/target/release/bundle/msi/*.msi "$OUT/" || true
cp apps/desktop/src-tauri/target/release/bundle/dmg/*.dmg "$OUT/" || true
cp apps/desktop/src-tauri/target/release/bundle/appimage/*.AppImage "$OUT/" || true

# Mobile
echo "Building Mobile..."
cd apps/mobile && flutter build apk --release && cd ../..
cp apps/mobile/build/app/outputs/flutter-apk/app-release.apk "$OUT/mobile-$VERSION.apk"

# Package
echo "Packaging..."
for f in "$OUT"/*; do
  if [[ -f "$f" && ! "$f" =~ \.(zip|tar\.gz|msi|dmg|AppImage|apk)$ ]]; then
    tar czf "$f.tar.gz" -C "$OUT" "$(basename "$f")"
  fi
done

# Checksums
echo "Generating checksums..."
cd "$OUT" && sha256sum *.* > "checksums.txt" && cd ..

echo "Done. Artifacts in $OUT/"
