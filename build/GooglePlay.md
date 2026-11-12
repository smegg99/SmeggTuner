# Google Play release

Application ID: `me.smegg.smeggtuner`. The version and version code come from
`VERSION` through `build/android-support/generate.py`.

## Privacy policy

Use `https://smeggtuner.com/privacy` in Play Console. The Polish page is
`https://smeggtuner.com/pl/privacy`. Publish the website before submitting the
Android release: a local page or an unpublished source file is not sufficient.
The app includes the complete policy in About and reads it without an internet
connection. Its optional online-version button opens the public website.

The canonical policy files are bundled from the SmeggTuner repository root:

- `PRIVACY.md`
- `PRIVACY.pl.md`

Run `just privacy-site` after editing them to update the generated copies in
SmeggTunerSite's `web/app/content/privacy/`. `just privacy-site --check` verifies
that both copies match before deployment. The Help dialog likewise bundles
`USAGE.md`, `USAGE.pl.md` and their screenshots, so the guide works offline.

Developer name: `smegg99`. Public privacy contact: `support@smegg.me`.
Keep the app's behavior, policy and Play Console declarations consistent.
The current policy covers microphone processing, selected files, local storage
and logs, Android backups, support requests, deletion and website processing.
Website analytics and security providers are separate from Android app behavior.

The page must return readable HTML over HTTPS worldwide, without signing in,
an editor, a download or a bot challenge. Check the app link, both languages,
mobile layout and the published content after deployment. Google requires both
an in-app policy and a public URL. See the
[User Data policy](https://support.google.com/googleplay/android-developer/answer/10144311?hl=en).

## Data safety declarations

These recommendations describe the current Android code, not future Workshop
features or a later SDK integration. Recheck the exact release bundle.

| Play Console item | Current Android behavior |
| --- | --- |
| Data collected or shared by the app | No app-managed off-device collection or sharing found. Microphone and file analysis run locally. Review the final bundle before answering No. |
| Microphone / audio files | Used on device for tuning and playback, not uploaded. Local-only processing is not collection under the Data safety definition. |
| Location, contacts, camera, advertising identifiers | Not used by the tuner. The package check rejects permissions outside the allowlist. |
| App analytics / crash upload | No analytics or automatic crash-upload SDK. Diagnostic output stays in local files or Android logcat. |
| Accounts | No sign-in or account creation. Account deletion is not applicable; local data can be removed in Android settings. |
| Ads | No ads. |
| External links | Open the system browser. The destination has its own data practices. |
| Android backup / transfer | Enabled through the operating system. Disclosed in the policy; the developer does not receive those backups. |

Google defines collection as sending user data off the device and distinguishes
local processing from collection. User-initiated transfers and service providers
have specific sharing exceptions; do not assume all transfers are exempt. Use
the actual questions shown in the Console and the
[Data safety guidance](https://support.google.com/googleplay/android-developer/answer/10787469?hl=en).
Do not claim independent security certification or app-level encryption of the
local database. Requests to external websites use HTTPS; that is a different
claim from encrypting stored app data.

Android backup behavior depends on the device, OS and user settings. See
[Android Auto Backup](https://developer.android.com/identity/data/autobackup).

## Build and package checks

`build/android-support/config.json` owns the target SDK and permission allowlist.
The tuner needs microphone access; the Wails host also uses internet and network
state permissions. AndroidX adds an app-scoped signature permission for its
private broadcast receiver. The generator removes unused Wails capabilities,
and preflight checks the actual APK or AAB permission list after manifest merging.
The local S99Wails browser service also contains the Android URL-handler fix
needed by the About links. Include that change when publishing the shared
dependency and removing the development `replace` directive.

Run the existing checks and release command from a terminal:

```sh
just check
just android-release arm64
```

The release command needs the real upload keystore and its password. Keep them
out of source and shell history; use the existing interactive signing prompt.
It verifies the signing certificate, package/version metadata, build mode,
native ABI and 16 KB ELF alignment. APK verification also checks ZIP alignment.
Upload the newly generated AAB from `bin/android/release/arm64/`, not a debug
APK or an older bundle in Gradle's intermediate output directory. Review Play
App Signing and retain the upload key securely.

The installation minimum is API 24 (Android 7.0), which meets the minimum SDK
requirement for [Google Play automatic protection](https://support.google.com/googleplay/android-developer/answer/10183279?hl=en).
This does not enable protection; its settings are managed in Play Console.

The configured target is API 37. The current submission minimum is API 36;
verify the [target API requirements](https://support.google.com/googleplay/android-developer/answer/11926878?hl=en)
when submitting. Native libraries must also work on 16 KB devices. Run an
emulator check on the `smeggtuner-api37-16k` profile and review
[Android's page-size requirements](https://developer.android.com/guide/practices/page-sizes).
Compile and alignment checks alone do not prove runtime compatibility.

## Console and device acceptance

1. Confirm the developer account, public identity, privacy contact, target
   countries and any applicable trader declarations.
2. Fill in Privacy policy, Data safety, Ads, App access, Target audience and
   Content rating. App access needs no credentials. Choose the intended age
   groups yourself; the code cannot determine whether you intend to target
   children. Complete the IARC questionnaire from the actual features.
3. Supply the store description, icon, feature graphic and screenshots from the
   release build. Describe live tuning and WAV analysis. Do not advertise the
   hidden Workshop, session recording or report tools as available features.
   Review the About dialog's Ko-fi destination: Google's tipping exception
   requires the contribution to go entirely to the creator and grant no digital
   content or services. Do not use that link to sell app features. See the
   [Payments policy guidance](https://support.google.com/googleplay/android-developer/answer/10281818?hl=en).
4. Test an internal-track build: microphone grant and denial, revocation, live
   tuning, background stop/resume, WAV import, rotations, both themes and
   languages, and the About privacy link. Check a physical instrument and
   microphone; emulator rendering does not establish tuning accuracy.
5. Check Android vitals and the Play pre-launch report, then use the appropriate
   test track before production. New personal accounts may require a closed
   test with 12 continuously opted-in testers for 14 days before applying for
   production access. Check the
   [personal-account testing requirements](https://support.google.com/googleplay/android-developer/answer/14151465).

The website must be published and the final signed AAB tested before this work
can be treated as a completed store submission. See Google's
[review preparation guide](https://support.google.com/googleplay/android-developer/answer/9859455?hl=en).
