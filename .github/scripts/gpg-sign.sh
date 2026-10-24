#!/usr/bin/env bash
# Detached, ASCII-armored GPG signature for a file (SHA256SUMS.txt → SHA256SUMS.txt.asc).
#
#   gpg-sign.sh dist/SHA256SUMS.txt
#
# Env: GPG_PRIVATE_KEY (armored secret key), GPG_PASSPHRASE.
# Uses a throwaway GNUPGHOME in $RUNNER_TEMP that is deleted on exit.
set -euo pipefail

FILE="${1:?usage: $0 <file>}"
: "${GPG_PRIVATE_KEY:?}" "${GPG_PASSPHRASE:?}"

export GNUPGHOME="${RUNNER_TEMP:?}/dlyven-gnupg"
rm -rf "$GNUPGHOME"
mkdir -m 700 "$GNUPGHOME"
trap 'gpgconf --kill gpg-agent 2>/dev/null || true; rm -rf "$GNUPGHOME"' EXIT

printf '%s\n' "$GPG_PRIVATE_KEY" | gpg --batch --quiet --import
printf '%s' "$GPG_PASSPHRASE" | gpg --batch --yes --pinentry-mode loopback --passphrase-fd 0 \
  --detach-sign --armor --output "${FILE}.asc" "$FILE"
gpg --batch --verify "${FILE}.asc" "$FILE"
echo "Signing key fingerprint:"
gpg --batch --list-secret-keys --with-colons | awk -F: '/^fpr:/{print $10; exit}'
