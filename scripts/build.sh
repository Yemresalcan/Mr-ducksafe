#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-dev}"
DIST_DIR="dist"

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

LDFLAGS="-s -w -X github.com/yemresalcan/mr-ducksafe/internal/cmd.Version=${VERSION}"

build() {
  local goos="$1" goarch="$2" ext="$3"
  local out="$DIST_DIR/mr-ducksafe-${goos}-${goarch}${ext}"
  echo "Building $out"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build \
    -ldflags "$LDFLAGS" -o "$out" ./cmd/mr-ducksafe
}

build windows amd64 .exe
build linux   amd64 ""
build darwin  amd64 ""
build darwin  arm64 ""

echo
echo "Done. Artifacts in $DIST_DIR/"
ls -lh "$DIST_DIR"
