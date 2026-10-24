#!/usr/bin/env bash
# The "Download" block at the top of every GitHub Release: one row per platform,
# linking straight to the file, so nobody has to decode the asset list.
#
#   release-notes.sh v1.2.3 dist > release-notes.md
#
# Rows whose file is missing from <dist> (a platform build failed) are left out.
# Env: GITHUB_REPOSITORY (owner/repo, set by Actions).
set -euo pipefail

TAG="${1:?usage: $0 <tag> <dist dir>}"
DIST="${2:?usage: $0 <tag> <dist dir>}"
REPO="${GITHUB_REPOSITORY:?}"
BASE="https://github.com/${REPO}/releases/download/${TAG}"
README="https://github.com/${REPO}#installation-and-security-warnings"

row() { # <platform> <file> <what it is>
  if [[ -f "$DIST/$2" ]]; then
    printf '| %s | [%s](%s/%s) | %s |\n' "$1" "$2" "$BASE" "$2" "$3"
  fi
}

cat <<EOF
## Download

Pick **one** file for your computer.

| Platform | File | Notes |
| --- | --- | --- |
EOF
row "**Windows** 10 / 11 (64-bit)" Dlyven-Windows-x64-Setup.exe "Installer. Recommended."
row "**Windows** 10 / 11 (64-bit)" Dlyven-Windows-x64-Portable.exe "Portable. No install, run it from any folder."
row "**macOS** 11+ (Apple Silicon and Intel)" Dlyven-macOS-Universal.dmg "Open it and drag Dlyven to Applications."
row "**macOS** 11+ (Apple Silicon and Intel)" Dlyven-macOS-Universal.zip "Same app, zipped."
row "**Linux** x64 (most PCs)" Dlyven-Linux-x64 "Ubuntu 22.04+, Debian 12+, Fedora 36+. Run \`chmod +x\` first."
row "**Linux** ARM64" Dlyven-Linux-ARM64 "Raspberry Pi 4/5 (64-bit OS), ARM laptops and servers."
cat <<EOF

Already have a Dlyven with **Settings → Updates**? Click **Update now** there instead. It downloads and installs this release for you.

First launch shows a security warning? See [Installation and security warnings](${README}).
Checksums: [SHA256SUMS.txt](${BASE}/SHA256SUMS.txt)

EOF
