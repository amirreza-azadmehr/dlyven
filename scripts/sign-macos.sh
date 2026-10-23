#!/usr/bin/env bash
# Sign + notarize + staple Dlyven.app so Gatekeeper lets it open.
#
# Needs a paid Apple Developer ID Application certificate. Secrets (CI) or env:
#   APPLE_SIGN_IDENTITY   e.g. "Developer ID Application: Name (TEAMID)"
#   APPLE_API_KEY         contents of AuthKey_XXXXXX.p8
#   APPLE_API_KEY_ID
#   APPLE_API_ISSUER      App Store Connect issuer UUID
#   APPLE_TEAM_ID
set -euo pipefail
cd "$(dirname "$0")/.."

APP="${1:-build/bin/Dlyven.app}"
ENTITLEMENTS="build/darwin/entitlements.plist"
IDENTITY="${APPLE_SIGN_IDENTITY:-}"

if [[ ! -d "$APP" ]]; then
  echo "no app bundle at $APP" >&2
  exit 1
fi

if [[ -z "$IDENTITY" ]]; then
  IDENTITY="$(security find-identity -v -p codesigning 2>/dev/null | awk -F'"' '/Developer ID Application/{print $2; exit}')"
fi

if [[ -z "$IDENTITY" ]]; then
  echo "No Developer ID certificate — Gatekeeper will still block downloaded copies."
  echo "Add MACOS_CERTIFICATE_P12 + Apple notary secrets on GitHub, then cut a new release."
  exit 0
fi

echo "→ codesign as $IDENTITY"
codesign --force --deep --strict --timestamp --options runtime \
  --entitlements "$ENTITLEMENTS" \
  --sign "$IDENTITY" \
  "$APP"
codesign --verify --deep --strict --verbose=2 "$APP"

if [[ -z "${APPLE_API_KEY:-}" || -z "${APPLE_API_KEY_ID:-}" || -z "${APPLE_API_ISSUER:-}" ]]; then
  echo "Signed, but not notarized (missing APPLE_API_KEY / KEY_ID / ISSUER)."
  echo "Apple will still warn until notarization lands."
  exit 0
fi

KEYFILE="$(mktemp)"
trap 'rm -f "$KEYFILE" "${APP}.notarize.zip"' EXIT
printf '%s' "$APPLE_API_KEY" > "$KEYFILE"

echo "→ notarize"
ditto -c -k --keepParent "$APP" "${APP}.notarize.zip"
xcrun notarytool submit "${APP}.notarize.zip" \
  --key "$KEYFILE" \
  --key-id "$APPLE_API_KEY_ID" \
  --issuer "$APPLE_API_ISSUER" \
  ${APPLE_TEAM_ID:+--team-id "$APPLE_TEAM_ID"} \
  --wait
xcrun stapler staple "$APP"
spctl --assess --type execute --verbose "$APP" || true
echo "✓ signed, notarized, stapled: $APP"
