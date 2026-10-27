# Android development

The Android host is generated from Wails v3.0.0-beta.34, the version pinned in
`go.mod`. Use `go tool wails3`; a global `wails3` installation may differ.

`build/android-support/config.json` owns Android toolchain versions, package ID,
SDK levels and the emulator matrix. `VERSION` owns the application version.
`just android prepare` regenerates `build/android/`, which is ignored by Git.
Change the generator or its inputs rather than files in that directory.

## Toolchain

This Linux workstation has Android Studio under
`~/.local/opt/android-studio`, SDK under `~/Android/Sdk`, and a `studio` launcher
under `~/.local/bin`. Android Studio also has an application-menu entry.

Android commands select JDK 21 without changing the system Java version. Set
`ANDROID_JAVA_HOME` to use another compatible JDK. Set `ANDROID_HOME` to use a
different SDK location. Android commands select NDK 28.2.13676358 explicitly;
this release still supports API 21 and emits native libraries with 16 KB page
alignment by default.

On another Linux workstation, install Android Studio and Android command-line
tools from [Google](https://developer.android.com/studio), Python 3, JDK 21,
`rsvg-convert` (librsvg), and this app's normal Go/frontend prerequisites. Extract
the command-line tools into `$ANDROID_HOME/cmdline-tools/latest`, with `bin/`
directly below `latest/`. Then run:

```sh
just android setup       # review SDK licenses and install pinned build tools
just android avds        # download images and create virtual devices
just android doctor     # check Java, SDK, NDK, KVM and devices
just android studio     # open build/android as a Gradle project
```

Emulation requires working KVM access on Linux. Images, NDK, IDE, Gradle caches,
and virtual-device disks require tens of GB. `just android avds` preserves existing
AVDs and their data. Starting the API 21 AVD creates and attaches its missing
512 MiB SD-card image; that system image does not provide newer Android
versions' emulated external storage.

## Build and run

```sh
just emulator                 # start the default AVD; keep this terminal open
# In another terminal, after Android finishes booting:
just android                  # detect ABI, build, install and launch
just android check            # check the existing debug APK on the emulator
just android logs             # follow app logs
just adb devices              # list ADB devices
```

Set `ANDROID_SERIAL=emulator-5554` on commands when several devices are
connected. `just android` also supports a selected physical ARM64/x86_64 phone;
`just android check` accepts emulators only. Use `just android check-phone` for
physical-device checks. Installation preserves app data.

For separate build and distribution steps:

```sh
just android build amd64       # build an emulator APK without installing it
just android build             # build an ARM64 debug APK
just android install amd64     # install and launch the existing emulator APK
just android package           # ARM64 release APK
just android bundle            # ARM64 Android App Bundle
just android help              # less common Android operations
```

`ANDROID_SERIAL=emulator-5554 just android install amd64` selects an emulator when more
than one is connected. Installation uses `adb install -r` and preserves app
data. A physical device needs USB debugging enabled and adb authorization.

Gradle packages the Go library built by the recipes. After changing Go or the
frontend, run `just android build` again before using Android Studio's Run button.
Each recipe stages exactly one freshly compiled ABI. Outputs are separated by
mode and architecture, for example
`bin/android/debug/amd64/smeggtuner-1.4.0-debug-amd64.apk`. Release APKs and AABs
use `bin/android/release/<arch>/`. Verification rejects unexpected ABIs, stale
native libraries, wrong build modes, package/version/SDK mismatches and invalid
signatures or alignment. Release verification also checks the signing certificate
against `ANDROID_KEY_ALIAS` in the supplied keystore.

Android builds use the frontend's pinned pnpm, a frozen lockfile and read-only
Go module resolution. They rebuild the frontend, including linked layer sources,
before compiling Go. Android commands select a compatible installed Node version
without changing the global selection.

The launcher checks for WebView 95 or newer before requesting microphone permission
or starting Go. A missing or older provider gets a native update/retry screen
in English or Polish instead of a blank frontend. `webViewMinMajor` also sets the
frontend build target. Android capture requests signed 16-bit PCM and normalizes
it to float32 for analysis; this avoids OpenSL ES falling back from unsupported
float capture to 16 kHz on old devices. Configuration,
SQLite and relative log paths use Android's private files directory; temporary
exports use its private cache. Entering the background stops microphone capture,
tone playback and calibration, then saves the session. Resuming leaves capture
stopped. Android does not use desktop tray/window-state management or the
PulseAudio endpoint-volume backend.

Android Back follows application state: it dismisses the top overlay, returns
from a session tab or calibration step, saves and closes an open session, then
returns to the tuner. At the tuner root, Android owns Back navigation. The
host uses AndroidX's supported Back dispatcher, including on recent Android
versions that no longer call an Activity's `onBackPressed` override.

Navigation waits for pending accordion edits and their durable write. Invalid
drafts or save failures keep the editor open with an error in the status bar.
Failed backend Close calls retain the active session for retry. Accordion
writes carry their session ID so a delayed editor cannot change another session.
Leaving or stopping calibration disarms recording, stops reference tones and
unpins the note.
The generator also guards two network APIs that Wails beta.34 calls without
checking whether API 21/22 provide them.

## Android range

The installation floor is Android 5.0/API 21. Compile and target SDK are API 37.
These are separate settings: targeting a recent SDK does not raise `minSdk`.
The framework builds only ARM64 and x86_64. Older ARM32-only phones need an
additional Wails/native build path and are not supported by these recipes.

| AVD | Image | Purpose |
| --- | --- | --- |
| `smeggtuner-api21` | API 21 | Installation floor, pre-runtime permissions |
| `smeggtuner-api23` | API 23 | Runtime microphone permissions |
| `smeggtuner-api28` | API 28 | Older devices and WebView |
| `smeggtuner-api28-aosp` | API 28, AOSP | Updatable AOSP WebView baseline |
| `smeggtuner-api30` | API 30 | Scoped storage generation |
| `smeggtuner-api35` | API 35 | Android 15 behavior |
| `smeggtuner-api37` | API 37.0 | Current base API |
| `smeggtuner-api37-16k` | API 37.2, 16 KB | Recent release and native page alignment |

Choose one with `just emulator smeggtuner-api23`. For automated checks,
`bash build/android-support/android.sh emulator smeggtuner-api37 -no-window
-no-audio -gpu swiftshader` starts it without using the desktop.

An installation floor is not a tested usability guarantee. Old Android versions
can be limited to old WebViews; this frontend uses Nuxt, Vue and Vuetify, whose
browser requirements must also be satisfied. Test rendering, generated service
calls, capture, denied permissions, rotation, suspend/resume, storage, import
and export on each intended Android/WebView combination before declaring
support. Audio latency and microphone routing also need real hardware tests.
Android exports use the system document picker and content URIs. PDF reports use
Android's print UI, including its Save as PDF destination, instead of desktop
Chrome. A print handoff closes the app's report dialog; it does not claim the
user saved a file. The native host owns system-bar, cutout and keyboard insets
around the WebView.

## Test a connected phone

Enable USB debugging, connect a data-capable USB cable, unlock the phone and
accept the computer's debugging authorization. Then run:

```sh
just android phone-info     # inspect the physical device without installing anything
just android build         # build the ARM64 debug APK for most phones
just android check-phone   # install the matching existing APK and run app checks
just android logs          # follow this app's logs
```

With multiple phones, set `ANDROID_SERIAL` for the command:

```sh
ANDROID_SERIAL=PHONE_SERIAL just android check-phone
```

The phone selector ignores emulators, requires an authorized physical device,
checks Android/ABI/WebView compatibility and chooses ARM64 or x86_64. For an
x86_64 phone, build `just android build amd64` first. Installation uses the
existing debug APK and `adb install -r`, preserving the app's data. It restarts
SmeggTuner. Allow microphone access on the phone when Android asks; the command
waits up to two minutes for authorization and launch.

The phone profile checks rendering, a generated binding, microphone events,
background/resume and a temporary session's save/reopen cycle. It removes its
test session and restores the previously open session. Evidence includes JSON,
screenshots and logs under `bin/android/checks/`. It leaves microphone capture
stopped after testing. It does not grant/revoke permissions, change global
rotation settings, update the phone's WebView or invoke the English-only native
picker automation used on the emulator matrix.

Check native Back, rotation, keyboard behavior, file import/export and PDF
printing interactively on the phone. Pitch accuracy requires a known reference
tone or instrument note; microphone events alone do not establish accuracy.

## Runtime acceptance

Start a dedicated headless emulator, build its APK, then run:

```sh
just android build amd64
ANDROID_SERIAL=emulator-5580 just android check
```

Use the serial assigned to that emulator when several devices are connected.
`just android check` selects a sole online emulator, installs the debug APK
without deleting app data, grants its
microphone permission and restarts this app. It checks the rendered frontend and
theme, a generated backend binding, measurement events, background/resume,
session persistence, document-picker cancellation, a real session archive export,
native Back, pending edits, calibration cleanup, archive import, Save as PDF,
landscape navigation and keyboard visibility. It records the actual WebView user
agent, JSON results, screenshots, exported files and logs under `bin/android/checks/`.

The check creates and deletes its own session and restores the previously open
session. Its exported files remain in the emulator's Downloads directory. Run
it on a test emulator, not a device holding a working session. Permission denial
and physical audio devices require additional checks.

The frontend targets Chrome 95 syntax and CSS, loads the array APIs needed by
its dependencies, and flattens static and runtime theme CSS layers. The Android
installation floor and the browser floor are independent. Updating the app does
not update the installed WebView.

### Old provider checks

Run the native recovery check before updating a stock emulator provider:

```sh
ANDROID_SERIAL=emulator-5580 just android check-webview
```

It checks the detected provider, retry, the update/settings action, return and
Close. Its result is `blocked_by_webview`, not a successful application test.
Providers 37 (API 21), 44 (API 23), 66/69 (API 28 AOSP/Google) and 83 (API 30)
have passed this recovery check. Such stock configurations cannot run the frontend.

The API 21/23 images and the API 28 AOSP image can exercise the browser baseline
with a pinned x86_64 AOSP WebView fixture:

```sh
ANDROID_SERIAL=emulator-5580 just android webview-fixture
ANDROID_SERIAL=emulator-5580 just android check
```

The fixture downloads WebView 95.0.4638.74 from the official AOSP repository at
[commit `3a8d2089900522ca542c39055d51b08188fb11cb`](https://android.googlesource.com/platform/external/chromium-webview/+/3a8d2089900522ca542c39055d51b08188fb11cb/)
and checks its SHA-256. Like the AOSP image build's `default_dev_cert` rule, it
signs this prebuilt with the public
[AOSP development key](https://android.googlesource.com/platform/build/+/refs/tags/android-12.1.0_r1/target/product/security/).
The installer requires an x86_64 API 21/23/28 SDK emulator whose existing provider
already uses that certificate. Android performs a normal
signed package update. It never changes a provider allowlist or installs on a
physical device. Downloaded keys exist only in a temporary directory. Evidence
records the upstream hash, emulator certificate and installed version under
`bin/android/checks/`. This is a test fixture, not an update channel for phones.

### Verification recorded on 2026-10-25

- Android Studio launched on an isolated display.
- ARM64 and x86_64 debug APKs passed the maintained build and artifact verifier.
- ARM64 release APK and AAB builds passed signature and certificate verification
  with a temporary validation key. Artifacts are separated under
  `bin/android/signing-validation/`; they are not distribution releases.
- API 37 and API 37.2 with WebView 145 passed the full runtime check. The latter
  guest reported a 16384-byte memory page size.
- API 35 with WebView 124 and API 30 updated to WebView 124 from Google's SDK
  image passed the full runtime check.
- API 21, API 23 and API 28 AOSP with the WebView 95 fixture passed the runtime
  check. API 21 requires mounted SD-card storage for its Downloads provider;
  its successful run used `-gpu swangle` after SwiftShader hit an emulator
  OpenGL transport abort. API 21 and both API 37 images also passed native
  import of their exported session archives. Every image in the successful
  matrix now passes native import and landscape/keyboard checks.
- Isolated Chromium 95 rendered the production frontend at 360x800 and 1280x900
  with four canvases, parsed theme rules, reachable navigation and no uncaught
  JavaScript errors. That browser-only check has no Go backend.

The seven-image result index and debug APK hashes are recorded locally in
`bin/android/checks/matrix-2026-10-25.json`. Each image passed 22 runtime checks.

The toolbar wraps below 900 CSS pixels. Short windows scroll the tuner panels
instead of compressing their canvases to a few pixels. API 21 and API 37 passed
landscape and native keyboard acceptance. On API 37 at 731x363 CSS pixels, all
navigation controls remained inside the viewport and canvas heights stayed at
least 39 pixels. Focused session fields stayed above the keyboard on both images.
API 37 also passed launch after denying microphone permission, capture failure
reporting and workshop navigation without that permission. Display-cutout variants
and real microphone accuracy still need separate acceptance. A working emulator
does not certify phone hardware.

## Distribution signing

The maintained release commands require all four signing variables and fail
before building when any is missing. Supply an upload key outside the repository:

```sh
export ANDROID_KEYSTORE_FILE=/absolute/path/to/upload.jks
export ANDROID_KEY_ALIAS=smeggtuner
read -rs ANDROID_KEYSTORE_PASSWORD; export ANDROID_KEYSTORE_PASSWORD
read -rs ANDROID_KEY_PASSWORD; export ANDROID_KEY_PASSWORD
just android bundle
```

Do not commit keystores or passwords. Back up the upload key separately. SDK
target, signing, native alignment and store policy still need release checks.

References: [Wails mobile](https://v3.wails.io/guides/mobile/),
[Android Back navigation](https://developer.android.com/guide/navigation/custom-back/predictive-back-gesture),
[AGP 9.2 compatibility](https://developer.android.com/build/releases/agp-9-2-0-release-notes),
[16 KB page sizes](https://developer.android.com/guide/practices/page-sizes).
