#!/usr/bin/env bash
set -euo pipefail

# SSH Manager Multi-Architecture Build Script
# Builds for: amd64, arm64, riscv64

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "🔧 SSH Manager Multi-Architecture Build"
echo "   Version: $VERSION"
echo "   Architectures: amd64, arm64, riscv64"
echo ""

# Setup Docker buildx
echo "📦 Setting up Docker buildx..."
docker buildx create --use --name ssh-manager-builder 2>/dev/null || docker buildx use ssh-manager-builder

# Build for each architecture
for ARCH in amd64 arm64 riscv64; do
    echo ""
    echo "🔨 Building for $ARCH..."
    
    # Determine Rust target
    case $ARCH in
        amd64)
            RUST_TARGET="x86_64-unknown-linux-gnu"
            GOARCH="amd64"
            ;;
        arm64)
            RUST_TARGET="aarch64-unknown-linux-gnu"
            GOARCH="arm64"
            ;;
        riscv64)
            RUST_TARGET="riscv64gc-unknown-linux-gnu"
            GOARCH="riscv64"
            ;;
    esac
    
    # Build Docker image
    docker buildx build \
        --platform "linux/$ARCH" \
        --build-arg "RUST_TARGET=$RUST_TARGET" \
        --build-arg "GOARCH=$GOARCH" \
        --build-arg "VERSION=$VERSION" \
        --build-arg "BUILD_TIME=$BUILD_TIME" \
        --build-arg "GIT_COMMIT=$GIT_COMMIT" \
        -t "ssh-manager:$VERSION-$ARCH" \
        -f Dockerfile.multiarch \
        --load \
        .
    
    # Export binary from image
    echo "📦 Extracting binary for $ARCH..."
    CONTAINER_ID=$(docker create "ssh-manager:$VERSION-$ARCH")
    mkdir -p "dist/$ARCH/bin"
    docker cp "$CONTAINER_ID:/app/ssh-manager-api" "dist/$ARCH/bin/ssh-manager-api-$ARCH"
    docker rm "$CONTAINER_ID"
    
    # Create tarball
    tar -czf "dist/ssh-manager-${VERSION}-linux-${ARCH}.tar.gz" -C "dist/$ARCH" .
    
    echo "✅ $ARCH build complete"
done

# Create multi-arch manifest
echo ""
echo "📦 Creating multi-arch manifest..."
docker manifest create "ssh-manager:$VERSION" \
    "ssh-manager:$VERSION-amd64" \
    "ssh-manager:$VERSION-arm64" \
    "ssh-manager:$VERSION-riscv64"

docker manifest annotate "ssh-manager:$VERSION" "ssh-manager:$VERSION-amd64" --arch amd64
docker manifest annotate "ssh-manager:$VERSION" "ssh-manager:$VERSION-arm64" --arch arm64
docker manifest annotate "ssh-manager:$VERSION" "ssh-manager:$VERSION-riscv64" --arch riscv64

docker manifest push "ssh-manager:$VERSION"

echo ""
echo "✅ Multi-arch build complete!"
echo ""
echo "Artifacts:"
ls -lh dist/*.tar.gz
