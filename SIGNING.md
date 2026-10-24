# Code signing

The release workflow (`.github/workflows/release.yml`) signs builds **only when the matching GitHub secrets are present**. A missing secret never fails the build. That platform just drops to its fallback, and the build log shows a warning saying so.

| Platform | Full mode (secrets set) | Fallback (no secrets, fork PRs, all PRs) |
| --- | --- | --- |
| macOS | Developer ID signed (hardened runtime, secure timestamp), notarized, stapled `.app` + `.dmg` | **Ad-hoc** signature only. Not trusted, not notarized. Users must confirm the first launch. |
| Windows | Authenticode-signed, timestamped `Dlyven.exe` **and** installer (PFX, Azure Artifact Signing, or SignPath) | Unsigned. SmartScreen warns. |
| Linux | `SHA256SUMS.txt` + GPG signature `SHA256SUMS.txt.asc` | `SHA256SUMS.txt` only |

Pull requests always build in fallback mode, even from this repository, so PR code never runs with signing credentials.

Add secrets under **GitHub → Settings → Secrets and variables → Actions → New repository secret**.

---

## macOS

### What you need

A paid [Apple Developer Program](https://developer.apple.com/programs/) membership (USD 99/year). You can't get a Developer ID certificate without it.

### 1. Create and export the Developer ID Application certificate

1. On a Mac, open **Xcode → Settings → Accounts**, select your team, and click **Manage Certificates… → + → Developer ID Application**.
   (Or create one at [developer.apple.com/account/resources/certificates](https://developer.apple.com/account/resources/certificates) from a CSR made in Keychain Access.)
2. Open **Keychain Access → login → My Certificates**, find **Developer ID Application: Your Name (TEAMID)**, and expand it to make sure the private key is underneath.
3. Right-click the certificate → **Export…** → format **Personal Information Exchange (.p12)**. Choose a strong password.
4. Base64 it and copy it to the clipboard:
   ```bash
   base64 -i DeveloperID.p12 | pbcopy
   ```
5. Find the exact identity string:
   ```bash
   security find-identity -v -p codesigning
   # 1) ABCD… "Developer ID Application: Your Name (TEAMID1234)"
   ```

### 2. Notarization credentials (pick one)

**Option A: Apple ID + app-specific password**

1. Go to [account.apple.com](https://account.apple.com) → **Sign-In and Security → App-Specific Passwords → +**. Name it, e.g. `dlyven-notary`.
2. Copy the generated `xxxx-xxxx-xxxx-xxxx` password. Apple shows it only once.
3. Your Team ID is shown at [developer.apple.com/account](https://developer.apple.com/account) → Membership details (10 characters).

**Option B: App Store Connect API key** (used first if both are set)

1. [App Store Connect → Users and Access → Integrations → App Store Connect API](https://appstoreconnect.apple.com/access/integrations/api) → **Team Keys → +**, role **Developer**.
2. Download `AuthKey_XXXXXXXXXX.p8` (one download only). Note the **Key ID** and the **Issuer ID** shown above the table.

### 3. GitHub secrets

| Secret | Required | Value |
| --- | --- | --- |
| `APPLE_CERTIFICATE` | yes | Output of `base64 -i DeveloperID.p12` |
| `APPLE_CERTIFICATE_PASSWORD` | yes | The .p12 export password |
| `APPLE_SIGNING_IDENTITY` | recommended | `Developer ID Application: Your Name (TEAMID1234)`. If empty, the first Developer ID identity in the imported .p12 is used. |
| `APPLE_ID` | option A | Apple ID email |
| `APPLE_APP_SPECIFIC_PASSWORD` | option A | `xxxx-xxxx-xxxx-xxxx` |
| `APPLE_TEAM_ID` | option A | `TEAMID1234` |
| `APPLE_API_KEY` | option B | **Full text** of `AuthKey_XXXXXXXXXX.p8`, including the `-----BEGIN PRIVATE KEY-----` lines (not base64) |
| `APPLE_API_KEY_ID` | option B | Key ID |
| `APPLE_API_ISSUER` | option B | Issuer ID (UUID) |

Legacy names from earlier versions of this repo still work as fallbacks: `MACOS_CERTIFICATE_P12` (= `APPLE_CERTIFICATE`), `MACOS_CERTIFICATE_PASSWORD` (= `APPLE_CERTIFICATE_PASSWORD`), `APPLE_SIGN_IDENTITY` (= `APPLE_SIGNING_IDENTITY`).

With a certificate but no notarization credentials, the build is Developer ID signed but **not notarized**, and Gatekeeper still warns. The log says so.

### What CI does

1. `.github/scripts/macos-keychain.sh import`: creates a temporary keychain in `$RUNNER_TEMP` with a random password, imports the .p12, and sets the key partition list so `codesign` doesn't stop at a prompt. The .p12 file is deleted right after the import, and the keychain is deleted at the end of the job (`if: always()`).
2. `scripts/build.sh darwin` → `scripts/sign-macos.sh`: signs nested Mach-O code and bundles inside-out, then the `.app`, with `--options runtime --timestamp --entitlements build/darwin/entitlements.plist`. This runs after the universal `lipo` and the icon copies, since any later change would break the signature.
3. `scripts/package-macos.sh`: builds `Dlyven.dmg` and signs it with the same identity.
4. `scripts/notarize-macos.sh`: `xcrun notarytool submit Dlyven.dmg --wait`. On failure it prints `xcrun notarytool log <id>` and fails the job. On success it staples the `.dmg` and the `.app`. The `.zip` is made afterwards, so it contains the stapled app.
5. Verifies with `codesign --verify --deep --strict`, `spctl -a -t exec` (app), and `spctl -a -t open --context context:primary-signature` (dmg).

Fallback: `codesign --force --deep --sign -` (ad-hoc). Apple silicon refuses to run code with no valid signature at all, so this stops the “damaged” error you'd get from that. It is **not** a trusted identity, and Gatekeeper still treats the app as from an unidentified developer.

### Entitlements

`build/darwin/entitlements.plist` is applied only in full mode, together with the hardened runtime. It was left unchanged by the signing work. Notes for a future review:

- `cs.allow-jit`, `cs.allow-unsigned-executable-memory`: commonly set for WebView apps. WKWebView runs JavaScript in a separate system process, so the Dlyven process itself may not need them.
- `cs.disable-library-validation`: only needed to load third-party dylibs into the process. yt-dlp, ffmpeg and Deno run as **child processes**, which this key doesn't affect.
- `network.client`, `files.user-selected.read-write`, `files.downloads.read-write`: these only do anything inside the App Sandbox, which Dlyven doesn't use, so they are no-ops today.

Trimming any of these changes runtime behavior and should be tested on a real Mac with a signed build first.

### Verify a signed build locally

```bash
codesign --verify --deep --strict --verbose=2 /Applications/Dlyven.app
codesign -dv --verbose=4 /Applications/Dlyven.app       # Authority=Developer ID…, flags=…runtime, Timestamp=…
codesign -d --entitlements - /Applications/Dlyven.app
spctl -a -vvv -t exec /Applications/Dlyven.app          # "source=Notarized Developer ID"
spctl -a -vvv -t open --context context:primary-signature Dlyven-macos-universal.dmg
xcrun stapler validate Dlyven-macos-universal.dmg
```

An ad-hoc build shows `Signature=adhoc`, and `spctl` reports `rejected`. That's expected.

---

## Windows

The workflow uses the **first complete set** it finds, in this order: PFX → Azure Artifact Signing → SignPath. Both `Dlyven.exe` (before it is packed into the installer) and `Dlyven-windows-amd64-installer.exe` are signed and timestamped. The installer is rebuilt with `makensis` after the exe is signed (`scripts/package-nsis.ps1`), so it ships the signed copy.

### Option 1: PFX certificate (OV/EV code-signing cert you hold as a file)

```powershell
[Convert]::ToBase64String([IO.File]::ReadAllBytes("codesign.pfx")) | Set-Clipboard
```
or on macOS/Linux: `base64 -i codesign.pfx | pbcopy` / `base64 -w0 codesign.pfx`.

| Secret | Value |
| --- | --- |
| `WINDOWS_CERTIFICATE` | base64 of the .pfx (legacy name `WIN_CERTIFICATE_PFX` still works) |
| `WINDOWS_CERTIFICATE_PASSWORD` | .pfx password (legacy name `WIN_CERTIFICATE_PASSWORD` still works) |

Signs with `signtool sign /fd SHA256 /tr http://timestamp.digicert.com /td SHA256`. The .pfx is written to `$RUNNER_TEMP` and deleted in a `finally` block. Most CAs now issue new certificates only on hardware tokens or cloud HSMs, so a plain PFX may not be available to you. If so, use option 2 or 3.

### Option 2: Azure Artifact Signing (formerly Trusted Signing)

1. Create an Artifact Signing account and a **Public Trust** certificate profile in Azure, and complete identity validation.
2. Create an app registration (service principal) with a client secret, and give it the *Certificate Profile Signer* role on the account. The portal calls it “Artifact Signing Certificate Profile Signer”, or “Trusted Signing Certificate Profile Signer” in older tenants.

| Secret | Value |
| --- | --- |
| `AZURE_TENANT_ID` | Directory (tenant) ID |
| `AZURE_CLIENT_ID` | App registration client ID |
| `AZURE_CLIENT_SECRET` | Client secret value |
| `AZURE_SIGNING_ENDPOINT` | Region endpoint, e.g. `https://eus.codesigning.azure.net/` |
| `AZURE_SIGNING_ACCOUNT_NAME` | Artifact Signing account name |
| `AZURE_SIGNING_CERTIFICATE_PROFILE` | Certificate profile name |

Uses `azure/artifact-signing-action` (pinned, v1.2.0) with SHA256 and the `http://timestamp.acs.microsoft.com` timestamp server.

### Option 3: SignPath (free for qualifying open-source projects)

1. Apply at [signpath.org](https://signpath.org) (OSS) or set up a SignPath.io organization. Add the GitHub repository as a **trusted build system** for the project.
2. In the project, create an **artifact configuration** that signs the `.exe` inside a zip archive. GitHub artifacts are zips, and each request contains one exe. One configuration covers both requests:
   ```xml
   <artifact-configuration xmlns="http://signpath.io/artifact-configuration/v1">
     <zip-file>
       <pe-file path="*.exe">
         <authenticode-sign />
       </pe-file>
     </zip-file>
   </artifact-configuration>
   ```
3. Create a signing policy and an API token for a CI user that is allowed to submit to it.

| Secret | Required | Value |
| --- | --- | --- |
| `SIGNPATH_API_TOKEN` | yes | CI user API token |
| `SIGNPATH_ORGANIZATION_ID` | yes | Organization ID (UUID) |
| `SIGNPATH_PROJECT_SLUG` | yes | Project slug |
| `SIGNPATH_SIGNING_POLICY_SLUG` | yes | e.g. `release-signing` |
| `SIGNPATH_ARTIFACT_CONFIGURATION_SLUG` | no | Leave unset to use the project's default configuration |

The build job uploads each unsigned exe as a short-lived artifact (`signpath-unsigned-*`, kept for 1 day). It then waits for SignPath and copies the signed file back. These staging artifacts are excluded from the GitHub Release. If the policy requires manual approval, the job waits until someone approves it.

### Verify locally

```powershell
Get-AuthenticodeSignature .\Dlyven-windows-amd64.exe | Format-List Status, SignerCertificate, TimeStamperCertificate
signtool verify /pa /v .\Dlyven-windows-amd64-installer.exe
```

A valid signature removes the “Unknown publisher” label, but SmartScreen reputation builds up over time and downloads. Since 2024 even EV certificates don't skip that.

---

## Linux (checksums + optional GPG)

Linux has nothing like Gatekeeper or SmartScreen to satisfy, so the binaries are not "signed". Every release ships `SHA256SUMS.txt`. To also publish `SHA256SUMS.txt.asc`:

```bash
gpg --quick-gen-key "Dlyven Releases <you@example.com>" ed25519 sign 2y
gpg --armor --export-secret-keys <FINGERPRINT> | pbcopy   # → GPG_PRIVATE_KEY
gpg --armor --export <FINGERPRINT> > dlyven-release.asc   # publish this public key
```

| Secret | Value |
| --- | --- |
| `GPG_PRIVATE_KEY` | ASCII-armored secret key |
| `GPG_PASSPHRASE` | Its passphrase |

`.github/scripts/gpg-sign.sh` imports the key into a throwaway `GNUPGHOME` in `$RUNNER_TEMP`, which is deleted on exit. Publish the public key somewhere users can trust (README, keyserver, your website) so `gpg --verify` actually means something.

---

## What can only be verified in real CI

The scripts pass `bash -n`, `shellcheck` and `actionlint`, but these can't be tested off a real runner with real credentials:

- Keychain import, `codesign` with a Developer ID, `notarytool`, `stapler`, `spctl` (macOS runner + Apple account)
- `signtool` with a PFX, Azure Artifact Signing, SignPath round-trips (Windows runner + certificates)
- `choco install nsis` fallback and `makensis` repack on `windows-latest`
- `hdiutil create` on `macos-latest`

Fallback mode runs on every pull request, so the no-secrets path is exercised regularly.
