#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$DIR"

BUILD_DIR="$DIR/build"
mkdir -p "$BUILD_DIR"

LDFLAGS="-s -w"

echo "Building racored..."

GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/racored-linux-amd64" ./cmd/racored/
GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/racored-linux-arm64" ./cmd/racored/
GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/racored-darwin-amd64" ./cmd/racored/
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/racored-darwin-arm64" ./cmd/racored/
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/racored-windows-amd64.exe" ./cmd/racored/

echo "Building racore (CLI)..."

GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/racore-linux-amd64" ./cmd/racore/
GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/racore-linux-arm64" ./cmd/racore/
GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/racore-darwin-amd64" ./cmd/racore/
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/racore-darwin-arm64" ./cmd/racore/
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/racore-windows-amd64.exe" ./cmd/racore/

echo ""
echo "Binaries:"
ls -lh "$BUILD_DIR"/
