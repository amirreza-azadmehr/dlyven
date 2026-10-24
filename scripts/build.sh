#!/usr/bin/env bash
# Build Dlyven for the current OS, or cross-compile Windows from Linux.
#
#   ./scripts/build.sh            # current OS
#   ./scripts/build.sh darwin     # macOS universal (Intel + Apple Silicon) — must run on macOS
#   ./scripts/build.sh linux           # linux/<host arch>
#   ./scripts/build.sh linux arm64     # linux/arm64 (must run on arm64)
#   ./scripts/build.sh windows    # windows/amd64 (MinGW / llvm-mingw; works from Linux)
#
# Env: VERSION=1.2.3 to stamp the version.
set -euo pipefail
cd "$(dirname "$0")/.."

VERSION="${VERSION:-$(grep -o '"productVersion": *"[^"]*"' wails.json | cut -d'"' -f4)}"
LDFLAGS="-s -w -X github.com/amirreza-azadmehr/dlyven/internal/appinfo.Version=${VERSION}"
TARGET="${1:-$(uname -s | tr '[:upper:]' '[:lower:]')}"

command -v wails >/dev/null || { echo "wails CLI not found: go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0"; exit 1; }

ensure_mingw() {
  if command -v x86_64-w64-mingw32-gcc >/dev/null; then
    return 0
  fi
  local dir
  for dir in "$HOME/.local/opt/llvm-mingw/bin" /usr/bin; do
    if [ -x "$dir/x86_64-w64-mingw32-gcc" ]; then
      PATH="$dir:$PATH"
      return 0
    fi
  done
  echo "Windows cross-compile needs MinGW."
  echo "  sudo apt install gcc-mingw-w64-x86-64"
  echo "or unpack llvm-mingw to ~/.local/opt/llvm-mingw"
  exit 1
}

echo "→ Running Go tests"
go test ./internal/...

case "$TARGET" in
  darwin)
    if [ "$(uname -s)" != "Darwin" ]; then
      echo "macOS builds must run on a Mac (or GitHub Actions macos-latest)."
      echo "Wails cannot cross-compile to darwin from Linux."
      exit 1
    fi
    ./scripts/generate-icons.sh
    wails build -clean -trimpath -platform darwin/universal -ldflags "$LDFLAGS"
    # Wails re-encodes a thin .icns and never writes Assets.car. Put both in
    # so Finder, Dock, and Spotlight show the purple logo instead of a generic glyph.
    RES=build/bin/Dlyven.app/Contents/Resources
    mkdir -p "$RES"
    [[ -f build/darwin/iconfile.icns ]] && cp build/darwin/iconfile.icns "$RES/iconfile.icns"
    [[ -f build/darwin/AppIcon.icns ]] && cp build/darwin/AppIcon.icns "$RES/AppIcon.icns"
    [[ -f build/darwin/Assets.car ]] && cp build/darwin/Assets.car "$RES/Assets.car"
    touch build/bin/Dlyven.app
    ./scripts/sign-macos.sh build/bin/Dlyven.app
    echo "✓ build/bin/Dlyven.app"
    ;;
  linux)
    ARCH="${2:-}"
    if [ -z "$ARCH" ]; then
      case "$(uname -m)" in
        x86_64|amd64) ARCH=amd64 ;;
        aarch64|arm64) ARCH=arm64 ;;
        *) echo "unsupported linux arch: $(uname -m)"; exit 1 ;;
      esac
    fi
    TAGS=()
    # Ubuntu 24.04+/Debian 13+ ship only WebKitGTK 4.1.
    if pkg-config --exists webkit2gtk-4.1 2>/dev/null && ! pkg-config --exists webkit2gtk-4.0 2>/dev/null; then
      TAGS=(-tags webkit2_41)
    fi
    wails build -clean -trimpath -platform "linux/${ARCH}" "${TAGS[@]}" -ldflags "$LDFLAGS"
    echo "✓ build/bin/Dlyven (linux/${ARCH})"
    ;;
  windows)
    ensure_mingw
    # Skip -clean so a Linux binary already in build/bin/ is kept.
    CGO_ENABLED=1 wails build -trimpath -platform windows/amd64 -webview2 embed -ldflags "$LDFLAGS"
    echo "✓ build/bin/Dlyven.exe"
    ;;
  *)
    echo "unknown target: $TARGET (use darwin | linux | windows)"
    exit 1
    ;;
esac
