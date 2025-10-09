#!/usr/bin/env bash
set -euo pipefail

VERSION="5.5"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIBDIR="$ROOT/libs"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

mkdir -p "$LIBDIR/macos" "$LIBDIR/linux_amd64" "$LIBDIR/windows_amd64"

curl -sSL "https://github.com/raysan5/raylib/releases/download/${VERSION}/raylib-${VERSION}_macos.tar.gz" \
  -o "$TMPDIR/raylib-macos.tar.gz"
tar -xzf "$TMPDIR/raylib-macos.tar.gz" -C "$TMPDIR"
cp "$TMPDIR"/raylib-${VERSION}_macos/lib/libraylib.${VERSION}.0.dylib "$LIBDIR/macos/"

curl -sSL "https://github.com/raysan5/raylib/releases/download/${VERSION}/raylib-${VERSION}_linux_amd64.tar.gz" \
  -o "$TMPDIR/raylib-linux-amd64.tar.gz"
tar -xzf "$TMPDIR/raylib-linux-amd64.tar.gz" -C "$TMPDIR"
cp "$TMPDIR"/raylib-${VERSION}_linux_amd64/lib/libraylib.so.${VERSION}.0 "$LIBDIR/linux_amd64/"

curl -sSL "https://github.com/raysan5/raylib/releases/download/${VERSION}/raylib-${VERSION}_win64_mingw-w64.zip" \
  -o "$TMPDIR/raylib-win64.zip"
unzip -q "$TMPDIR/raylib-win64.zip" -d "$TMPDIR"
cp "$TMPDIR"/raylib-${VERSION}_win64_mingw-w64/lib/raylib.dll "$LIBDIR/windows_amd64/"

echo "Raylib ${VERSION} libraries downloaded into $LIBDIR"
