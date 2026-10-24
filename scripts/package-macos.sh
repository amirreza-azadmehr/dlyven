#!/usr/bin/env bash
# Wrap an already signed Dlyven.app in a drag-to-Applications .dmg.
#
#   ./scripts/package-macos.sh [build/bin/Dlyven.app] [build/bin/Dlyven.dmg]
#
# With dmgbuild installed (python3 -m pip install dmgbuild) the image opens as the classic
# installer window: Dlyven, an arrow, Applications (build/darwin/dmg-settings.py). Without
# it you get a plain window with the same two items, so a missing tool never blocks a build.
#
# If the app carries a Developer ID signature, the .dmg is signed with the same
# identity. An ad-hoc app gets an unsigned .dmg (ad-hoc signing a disk image adds nothing).
set -euo pipefail
cd "$(dirname "$0")/.."

APP="${1:-build/bin/Dlyven.app}"
OUT="${2:-build/bin/Dlyven.dmg}"

if [[ ! -d "$APP" ]]; then
  echo "no app bundle at $APP" >&2
  exit 1
fi

rm -f "$OUT"
# hdiutil (dmgbuild runs it too) on hosted runners sometimes fails with "Resource busy";
# retry a few times.
if command -v dmgbuild >/dev/null 2>&1; then
  for attempt in 1 2 3 4 5; do
    if dmgbuild -s build/darwin/dmg-settings.py -D app="$APP" Dlyven "$OUT"; then
      break
    fi
    [[ $attempt -eq 5 ]] && { echo "dmgbuild failed" >&2; exit 1; }
    rm -f "$OUT"
    sleep $((attempt * 5))
  done
else
  echo "→ dmgbuild not found: plain .dmg without the arrow layout (python3 -m pip install dmgbuild)"
  STAGE="$(mktemp -d)"
  trap 'rm -rf "$STAGE"' EXIT
  ditto "$APP" "$STAGE/$(basename "$APP")"
  ln -s /Applications "$STAGE/Applications"
  for attempt in 1 2 3 4 5; do
    if hdiutil create -volname Dlyven -srcfolder "$STAGE" -fs HFS+ -format UDZO -ov "$OUT"; then
      break
    fi
    [[ $attempt -eq 5 ]] && { echo "hdiutil create failed" >&2; exit 1; }
    sleep $((attempt * 5))
  done
fi

IDENTITY="$( (codesign -dvv "$APP" 2>&1 || true) | sed -n 's/^Authority=//p' | head -1)"
if [[ "$IDENTITY" == "Developer ID Application:"* ]]; then
  echo "→ codesign dmg as $IDENTITY"
  codesign --force --timestamp --sign "$IDENTITY" "$OUT"
  codesign --verify --verbose=2 "$OUT"
else
  echo "→ app is not Developer ID signed; leaving $OUT unsigned"
fi
echo "✓ $OUT"
