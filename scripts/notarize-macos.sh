#!/usr/bin/env bash
# Notarize a Developer ID-signed .dmg (or .zip / .pkg), then staple the ticket.
#
#   ./scripts/notarize-macos.sh build/bin/Dlyven.dmg [build/bin/Dlyven.app ...]
#
# The first file is submitted to Apple. It and every extra path are stapled
# afterwards. Notarizing the .dmg covers the .app inside it, so the .app can be
# stapled too (needed for the .zip, since a zip itself cannot be stapled).
#
# Credentials (App Store Connect API key is preferred when both are set):
#   APPLE_API_KEY (contents of AuthKey_XXXX.p8), APPLE_API_KEY_ID, APPLE_API_ISSUER
#   or APPLE_ID, APPLE_APP_SPECIFIC_PASSWORD, APPLE_TEAM_ID
set -euo pipefail
cd "$(dirname "$0")/.."

SUBMIT="${1:?usage: $0 <file-to-submit> [extra-paths-to-staple...]}"
shift

TMP="$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/notary.XXXXXX")"
trap 'rm -rf "$TMP"' EXIT

if [[ -n "${APPLE_API_KEY:-}" && -n "${APPLE_API_KEY_ID:-}" && -n "${APPLE_API_ISSUER:-}" ]]; then
  KEYFILE="$TMP/AuthKey.p8"
  (umask 077; printf '%s' "$APPLE_API_KEY" > "$KEYFILE")
  AUTH=(--key "$KEYFILE" --key-id "$APPLE_API_KEY_ID" --issuer "$APPLE_API_ISSUER")
elif [[ -n "${APPLE_ID:-}" && -n "${APPLE_APP_SPECIFIC_PASSWORD:-}" && -n "${APPLE_TEAM_ID:-}" ]]; then
  AUTH=(--apple-id "$APPLE_ID" --team-id "$APPLE_TEAM_ID" --password "$APPLE_APP_SPECIFIC_PASSWORD")
else
  echo "No notarization credentials — $SUBMIT is signed but NOT notarized. Gatekeeper will still warn."
  exit 0
fi

echo "→ notarytool submit $SUBMIT (waits for Apple, usually a few minutes)"
set +e
xcrun notarytool submit "$SUBMIT" "${AUTH[@]}" --wait --output-format json > "$TMP/submit.json"
rc=$?
set -e
cat "$TMP/submit.json"; echo

ID="$(plutil -extract id raw -o - "$TMP/submit.json" 2>/dev/null || true)"
STATUS="$(plutil -extract status raw -o - "$TMP/submit.json" 2>/dev/null || true)"
if [[ $rc -ne 0 || "$STATUS" != "Accepted" ]]; then
  echo "::error::Notarization failed (status: ${STATUS:-unknown}, exit $rc)"
  if [[ -n "$ID" ]]; then
    xcrun notarytool log "$ID" "${AUTH[@]}" || true
  fi
  exit 1
fi

for path in "$SUBMIT" "$@"; do
  echo "→ staple $path"
  xcrun stapler staple "$path"
  xcrun stapler validate "$path"
done
echo "✓ notarized and stapled: $SUBMIT $*"
