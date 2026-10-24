#!/usr/bin/env bash
# Temporary CI keychain for the Developer ID Application certificate.
#
#   macos-keychain.sh import    # needs APPLE_CERTIFICATE (base64 .p12) + APPLE_CERTIFICATE_PASSWORD
#   macos-keychain.sh cleanup   # always run this at the end of the job
#
# The keychain gets its own random password, so the .p12 password is only used
# for the import. Files live in $RUNNER_TEMP and the .p12 is removed right away.
set -euo pipefail

KEYCHAIN="${RUNNER_TEMP:?}/dlyven-signing.keychain-db"
CERT="${RUNNER_TEMP}/dlyven-signing.p12"

case "${1:-}" in
  import)
    : "${APPLE_CERTIFICATE:?}" "${APPLE_CERTIFICATE_PASSWORD:?}"
    trap 'rm -f "$CERT"' EXIT
    printf '%s' "$APPLE_CERTIFICATE" | base64 --decode > "$CERT"

    KEYCHAIN_PASSWORD="$(openssl rand -base64 32)"
    echo "::add-mask::${KEYCHAIN_PASSWORD}"
    security create-keychain -p "$KEYCHAIN_PASSWORD" "$KEYCHAIN"
    security set-keychain-settings -lut 21600 "$KEYCHAIN"
    security unlock-keychain -p "$KEYCHAIN_PASSWORD" "$KEYCHAIN"
    security import "$CERT" -P "$APPLE_CERTIFICATE_PASSWORD" -A -t cert -f pkcs12 -k "$KEYCHAIN"
    # Without this codesign stops at a GUI "allow access" prompt and the job hangs.
    security set-key-partition-list -S apple-tool:,apple: -s -k "$KEYCHAIN_PASSWORD" "$KEYCHAIN" >/dev/null
    # Put ours first, keep the login/system keychains for the Apple intermediates.
    # shellcheck disable=SC2046
    security list-keychains -d user -s "$KEYCHAIN" $(security list-keychains -d user | tr -d '"')
    security default-keychain -d user -s "$KEYCHAIN"

    if ! security find-identity -v -p codesigning "$KEYCHAIN" | grep -q "Developer ID Application"; then
      echo "::error::APPLE_CERTIFICATE imported but holds no 'Developer ID Application' identity."
      exit 1
    fi
    echo "Imported Developer ID certificate into a temporary keychain."
    ;;
  cleanup)
    rm -f "$CERT"
    if [ -f "$KEYCHAIN" ]; then
      security delete-keychain "$KEYCHAIN" || true
      echo "Deleted temporary keychain."
    fi
    ;;
  *)
    echo "usage: $0 import | cleanup" >&2
    exit 2
    ;;
esac
