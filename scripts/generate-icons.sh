#!/usr/bin/env bash
# Build macOS icons from build/appicon.png + build/darwin/AppIcon.icon.
#
# Wails writes a thin .icns and no Assets.car. On macOS 26 Finder/Dock then
# show a generic glass glyph. This writes a full iconutil .icns and compiles
# the Icon Composer document so the purple logo actually appears.
set -euo pipefail
cd "$(dirname "$0")/.."

SRC="build/appicon.png"
ICON_DOC="build/darwin/AppIcon.icon"
if [[ ! -f "$SRC" ]]; then
  echo "missing $SRC" >&2
  exit 1
fi

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "skip macOS icons (iconutil/actool are Mac-only)"
  exit 0
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p build/darwin

ICONSET="$tmp/Dlyven.iconset"
mkdir -p "$ICONSET"
sips -z 16 16     "$SRC" --out "$ICONSET/icon_16x16.png" >/dev/null
sips -z 32 32     "$SRC" --out "$ICONSET/icon_16x16@2x.png" >/dev/null
sips -z 32 32     "$SRC" --out "$ICONSET/icon_32x32.png" >/dev/null
sips -z 64 64     "$SRC" --out "$ICONSET/icon_32x32@2x.png" >/dev/null
sips -z 128 128   "$SRC" --out "$ICONSET/icon_128x128.png" >/dev/null
sips -z 256 256   "$SRC" --out "$ICONSET/icon_128x128@2x.png" >/dev/null
sips -z 256 256   "$SRC" --out "$ICONSET/icon_256x256.png" >/dev/null
sips -z 512 512   "$SRC" --out "$ICONSET/icon_256x256@2x.png" >/dev/null
sips -z 512 512   "$SRC" --out "$ICONSET/icon_512x512.png" >/dev/null
sips -z 1024 1024 "$SRC" --out "$ICONSET/icon_512x512@2x.png" >/dev/null
iconutil -c icns "$ICONSET" -o build/darwin/iconfile.icns
cp build/darwin/iconfile.icns build/darwin/AppIcon.icns
echo "✓ build/darwin/iconfile.icns"

ACTOOL="$(xcrun --find actool 2>/dev/null || true)"
if [[ -z "$ACTOOL" ]]; then
  echo "actool not found — using committed Assets.car if present"
  exit 0
fi

if [[ -d "$ICON_DOC" ]]; then
  mkdir -p "$tmp/car"
  if "$ACTOOL" \
    --compile "$tmp/car" \
    --platform macosx \
    --minimum-deployment-target 26.0 \
    --app-icon AppIcon \
    --include-all-app-icons \
    --output-partial-info-plist "$tmp/partial.plist" \
    --output-format human-readable-text \
    --notices --warnings --errors \
    "$ICON_DOC"
  then
    if [[ -f "$tmp/car/Assets.car" ]]; then
      cp "$tmp/car/Assets.car" build/darwin/Assets.car
      echo "✓ build/darwin/Assets.car (Liquid Glass)"
      exit 0
    fi
  fi
  echo "actool could not compile $ICON_DOC — keeping existing Assets.car"
fi
