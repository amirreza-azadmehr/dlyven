#!/usr/bin/env bash
# Code-sign Dlyven.app. Must run after every step that touches the bundle
# (lipo, icon copies, Info.plist edits) since any change breaks the signature.
#
#   ./scripts/sign-macos.sh [build/bin/Dlyven.app]
#
# With a Developer ID Application identity (APPLE_SIGNING_IDENTITY, or the first
# one found in the keychain): nested code first, then the app, with hardened
# runtime + secure timestamp + build/darwin/entitlements.plist. That is what
# notarization requires; run scripts/notarize-macos.sh afterwards.
#
# Without one: ad-hoc signature only. That stops Apple silicon calling the app
# "damaged", but it is NOT a trusted or notarized signature — Gatekeeper will
# still ask the user to confirm the first launch.
set -euo pipefail
cd "$(dirname "$0")/.."

APP="${1:-build/bin/Dlyven.app}"
ENTITLEMENTS="build/darwin/entitlements.plist"
IDENTITY="${APPLE_SIGNING_IDENTITY:-${APPLE_SIGN_IDENTITY:-}}"

if [[ ! -d "$APP" ]]; then
  echo "no app bundle at $APP" >&2
  exit 1
fi

if [[ -z "$IDENTITY" ]]; then
  IDENTITY="$(security find-identity -v -p codesigning 2>/dev/null | awk -F'"' '/Developer ID Application/{print $2; exit}')"
fi

if [[ -z "$IDENTITY" ]]; then
  echo "→ no Developer ID identity: ad-hoc signing $APP"
  codesign --force --deep --sign - "$APP"
  codesign --verify --deep --strict --verbose=2 "$APP"
  msg="macOS build is only AD-HOC signed (not Developer ID, not notarized). Users must confirm the first launch; see README → Installation and security warnings."
  if [[ -n "${GITHUB_ACTIONS:-}" ]]; then echo "::warning::$msg"; else echo "⚠ $msg"; fi
  exit 0
fi

echo "→ codesign as $IDENTITY"
MAIN="$APP/Contents/MacOS/$(/usr/libexec/PlistBuddy -c 'Print :CFBundleExecutable' "$APP/Contents/Info.plist")"
SIGN=(codesign --force --options runtime --timestamp --sign "$IDENTITY")

# Inside-out: find -depth lists children before their parent bundle.
while IFS= read -r -d '' item; do
  if [[ -d "$item" ]]; then
    "${SIGN[@]}" "$item"
  elif [[ "$item" != "$MAIN" ]] && file -b "$item" | grep -q 'Mach-O'; then
    "${SIGN[@]}" "$item"
  fi
done < <(find "$APP/Contents" -depth -mindepth 1 -not -path '*/_CodeSignature/*' \
  \( -type f -o -type d \( -name '*.framework' -o -name '*.app' -o -name '*.xpc' -o -name '*.appex' -o -name '*.bundle' \) \) -print0)

"${SIGN[@]}" --entitlements "$ENTITLEMENTS" "$APP"
codesign --verify --deep --strict --verbose=2 "$APP"
echo "✓ Developer ID signed: $APP (not notarized yet)"
