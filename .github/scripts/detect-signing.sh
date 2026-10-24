#!/usr/bin/env bash
# Decide which signing path this job can take, from which secrets are non-empty.
# `secrets.*` can't be used in `if:`, so the workflow passes them in as env and
# gates later steps on these outputs instead. Only presence is checked; values
# are never printed.
#
# Outputs:
#   apple_sign     true | false               Developer ID .p12 available
#   apple_notary   api-key | apple-id | none  notarization credentials
#   windows_sign   pfx | azure | signpath | none
set -euo pipefail

has() { [ -n "${!1:-}" ]; }

# all VAR... → 0 when every VAR is set. Warns when only some are, which is
# almost always a typo in a secret name.
all() {
  local label="$1" set=0 missing=()
  shift
  for v in "$@"; do
    if has "$v"; then set=$((set + 1)); else missing+=("$v"); fi
  done
  if [ "$set" -eq "$#" ]; then return 0; fi
  if [ "$set" -gt 0 ]; then
    echo "::warning::${label}: some secrets are set but these are missing: ${missing[*]}. Ignoring ${label}."
  fi
  return 1
}

apple_sign=false
apple_notary=none
windows_sign=none

if [ "${GITHUB_EVENT_NAME:-}" = "pull_request" ]; then
  echo "Pull request build: signing is disabled, building in fallback mode."
else
  if all "Apple Developer ID" APPLE_CERTIFICATE APPLE_CERTIFICATE_PASSWORD; then
    apple_sign=true
  fi
  if all "App Store Connect API key" APPLE_API_KEY APPLE_API_KEY_ID APPLE_API_ISSUER; then
    apple_notary=api-key
  elif all "Apple ID notarization" APPLE_ID APPLE_APP_SPECIFIC_PASSWORD APPLE_TEAM_ID; then
    apple_notary=apple-id
  fi
  # Notarization only accepts Developer ID-signed code.
  if [ "$apple_sign" = false ] && [ "$apple_notary" != none ]; then
    echo "::warning::Apple notarization credentials are set but APPLE_CERTIFICATE is not. Skipping notarization."
    apple_notary=none
  fi

  # First complete set wins: PFX, then Azure Artifact Signing, then SignPath.
  if all "Windows PFX" WINDOWS_CERTIFICATE WINDOWS_CERTIFICATE_PASSWORD; then
    windows_sign=pfx
  elif all "Azure Artifact Signing" AZURE_TENANT_ID AZURE_CLIENT_ID AZURE_CLIENT_SECRET \
      AZURE_SIGNING_ENDPOINT AZURE_SIGNING_ACCOUNT_NAME AZURE_SIGNING_CERTIFICATE_PROFILE; then
    windows_sign=azure
  elif all "SignPath" SIGNPATH_API_TOKEN SIGNPATH_ORGANIZATION_ID SIGNPATH_PROJECT_SLUG \
      SIGNPATH_SIGNING_POLICY_SLUG; then
    windows_sign=signpath
  fi
fi

case "${RUNNER_OS:-}" in
  macOS) echo "macOS signing: developer-id=${apple_sign} notarization=${apple_notary}" ;;
  Windows) echo "Windows signing: ${windows_sign}" ;;
esac

{
  echo "apple_sign=${apple_sign}"
  echo "apple_notary=${apple_notary}"
  echo "windows_sign=${windows_sign}"
} >> "${GITHUB_OUTPUT:-/dev/null}"
