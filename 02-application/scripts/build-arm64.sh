#!/usr/bin/env bash
set -euo pipefail

# SSH Manager ARM64 Native Build Script
# Usage: ./scripts/build-arm64.sh [dev|staging|production]

ENV=${1:-production}
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "🔧 SSH Manager ARM64 Build"
echo "   Version: $VERSION"
echo "   Environment: $ENV"
echo "   Commit: $GIT_COMMIT"
echo "   Build Time: $BUILD_TIME"
echo ""

# Check if running on ARM64
ARCH=$(uname -m)
if [[ "$ARCH" != "aarch64" && "$ARCH" != "arm64" ]]; then
    echo "⚠️  Not on ARM64 host ($ARCH). Using cross-compilation..."
    CROSS_COMPILE=1
else
    echo "✅ Running on ARM64 host"
    CROSS_COMPILE=0
fi

# Install cross-compilation tools if needed
if [[ "$CROSS_COMPILE" == "1" ]]; then
    echo "📦 Installing cross-compilation tools..."
    if command -v apt-get &>/dev/null; then
        sudo apt-get update -qq
        sudo apt-get install -y -qq gcc-aarch64-linux-gnu g++-aarch64-linux-gnu
    elif command -v brew &>/dev/null; then
        brew install aarch64-elf-gcc 2>/dev/null || true
    fi
fi

# Build Rust library
echo "📦 Building Rust library (ARM64)..."
cd crates/core
if [[ "$CROSS_COMPILE" == "1" ]]; then
    export CC=aarch64-linux-gnu-gcc
    export CXX=aarch64-linux-gnu-g++
    cargo build --release --target aarch64-unknown-linux-gnu
else
    cargo build --release --target aarch64-unknown-linux-gnu
fi
cd "$PROJECT_ROOT"

# Build Go API server
echo "📦 Building Go API server (ARM64)..."
cd apps/api
if [[ "$CROSS_COMPILE" == "1" ]]; then
    export CC=aarch64-linux-gnu-gcc
    export CXX=aarch64-linux-gnu-g++
    export GOOS=linux
    export GOARCH=arm64
    export CGO_ENABLED=1
fi

go build -ldflags "-s -w \
    -X main.Version=$VERSION \
    -X main.BuildTime=$BUILD_TIME \
    -X main.GitCommit=$GIT_COMMIT \
    -X main.Environment=$ENV" \
    -o ../../bin/ssh-manager-api-arm64 \
    ./cmd/server
cd "$PROJECT_ROOT"

# Build Web client
echo "📦 Building Web client..."
cd apps/web
if ! command -v pnpm &>/dev/null; then
    npm install -g pnpm
fi
pnpm install --frozen-lockfile
pnpm run build:production
cd "$PROJECT_ROOT"

# Create distribution
echo "📦 Creating distribution..."
mkdir -p dist/arm64/{bin,web,config}
cp bin/ssh-manager-api-arm64 dist/arm64/bin/
cp -r apps/web/dist dist/arm64/web/
cp apps/api/config.yaml dist/arm64/config/
cp README.md LICENSE dist/arm64/

# Create tarball
echo "📦 Creating tarball..."
tar -czf "dist/ssh-manager-${VERSION}-linux-arm64.tar.gz" -C dist/arm64 .

# Create Debian package
echo "📦 Creating Debian package..."
if command -v dpkg-deb &>/dev/null; then
    mkdir -p "dist/deb-arm64/DEBIAN"
    cat > "dist/deb-arm64/DEBIAN/control" << EOF
Package: ssh-manager
Version: ${VERSION}
Section: net
Priority: optional
Architecture: arm64
Depends: libc6 (>= 2.31)
Maintainer: SSH Manager Team <team@sshmanager.dev>
Description: SSH Manager - Multi-Protocol Remote Access Platform
 Web-based SSH, RDP, VNC client with team collaboration.
EOF
    
    mkdir -p "dist/deb-arm64/usr/bin"
    mkdir -p "dist/deb-arm64/usr/share/ssh-manager/web"
    mkdir -p "dist/deb-arm64/etc/ssh-manager"
    mkdir -p "dist/deb-arm64/lib/systemd/system"
    
    cp dist/arm64/bin/ssh-manager-api-arm64 dist/deb-arm64/usr/bin/ssh-manager
    cp -r dist/arm64/web/* dist/deb-arm64/usr/share/ssh-manager/web/
    cp dist/arm64/config/config.yaml dist/deb-arm64/etc/ssh-manager/
    
    # Systemd service
    cat > "dist/deb-arm64/lib/systemd/system/ssh-manager.service" << 'EOF'
[Unit]
Description=SSH Manager
After=network.target

[Service]
Type=simple
User=ssh-manager
Group=ssh-manager
ExecStart=/usr/bin/ssh-manager
Restart=always
RestartSec=5
Environment=CONFIG_PATH=/etc/ssh-manager/config.yaml

[Install]
WantedBy=multi-user.target
EOF
    
    dpkg-deb --build "dist/deb-arm64" "dist/ssh-manager-${VERSION}-arm64.deb"
fi

# Create RPM package
echo "📦 Creating RPM package..."
if command -v rpmbuild &>/dev/null; then
    mkdir -p ~/rpmbuild/{SPECS,SOURCES,BUILD,RPMS,SRPMS}
    
    cat > ~/rpmbuild/SPECS/ssh-manager.spec << EOF
Name:           ssh-manager
Version:        ${VERSION}
Release:        1%{?dist}
Summary:        SSH Manager - Multi-Protocol Remote Access Platform
License:        MIT
BuildArch:      aarch64

%description
Web-based SSH, RDP, VNC client with team collaboration.

%install
mkdir -p %{buildroot}/usr/bin
mkdir -p %{buildroot}/usr/share/ssh-manager/web
mkdir -p %{buildroot}/etc/ssh-manager
mkdir -p %{buildroot}/lib/systemd/system
cp ${PROJECT_ROOT}/dist/arm64/bin/ssh-manager-api-arm64 %{buildroot}/usr/bin/ssh-manager
cp -r ${PROJECT_ROOT}/dist/arm64/web/* %{buildroot}/usr/share/ssh-manager/web/
cp ${PROJECT_ROOT}/dist/arm64/config/config.yaml %{buildroot}/etc/ssh-manager/
cp ${PROJECT_ROOT}/dist/deb-arm64/lib/systemd/system/ssh-manager.service %{buildroot}/lib/systemd/system/

%files
/usr/bin/ssh-manager
/usr/share/ssh-manager/web/*
/etc/ssh-manager/config.yaml
/lib/systemd/system/ssh-manager.service
EOF
    
    rpmbuild -bb ~/rpmbuild/SPECS/ssh-manager.spec
    cp ~/rpmbuild/RPMS/aarch64/*.rpm "dist/"
fi

echo ""
echo "✅ ARM64 build complete!"
echo ""
echo "Artifacts:"
ls -lh dist/*arm64* 2>/dev/null || true
ls -lh dist/*.deb 2>/dev/null || true
ls -lh dist/*.rpm 2>/dev/null || true
