#!/usr/bin/env bash
# build/android-support/android.sh
set -euo pipefail
android_support="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$android_support/../.."
source "$android_support/env.sh"

config() {
    python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))[sys.argv[2]])' "$android_support/config.json" "$1"
}
fail() { echo "$*" >&2; exit 1; }
preflight() { python3 -B "$android_support/preflight.py" "$@"; }
lock_project() {
    local state key
    state="${XDG_STATE_HOME:-$HOME/.local/state}/smeggtuner/android"
    key="$(printf '%s' "$PWD" | sha256sum | cut -d' ' -f1)"
    mkdir -p -- "$state"
    exec 9>"$state/$key.lock"
    flock 9
}
prepare() { python3 -B "$android_support/generate.py"; }
select_artifact() {
    arch="${1:-arm64}"
    mode="${2:-debug}"
    case "$arch" in arm64) abi=arm64-v8a ;; amd64) abi=x86_64 ;; *) fail 'Architecture must be arm64 or amd64' ;; esac
    case "$mode" in debug|release) ;; *) fail 'Mode must be debug or release' ;; esac
    version="$(cat VERSION)"
    output="bin/android/$mode/$arch"
    name="smeggtuner-$version-$mode-$arch"
}
select_device() {
    local devices
    devices="$(timeout 15 adb devices)"
    if [[ -n "${ANDROID_SERIAL:-}" ]]; then
        serial="$ANDROID_SERIAL"
        awk -v serial="$serial" '$1 == serial && $2 == "device" {found=1} END {exit !found}' <<< "$devices" || fail 'ANDROID_SERIAL must name an online adb device'
    else
        mapfile -t online < <(awk 'NR > 1 && $2 == "device" {print $1}' <<< "$devices")
        [[ "${#online[@]}" == 1 ]] || fail 'Set ANDROID_SERIAL or connect exactly one online adb device'
        serial="${online[0]}"
    fi
}
app_pids() {
    timeout 10 adb -s "$serial" shell ps | tr -d '\r' | awk -v app="$app_id" '$NF == app {print $2}'
}
install_app() {
    select_artifact "${1:-arm64}" "${2:-debug}"
    lock_project
    artifact="$PWD/$output/$name.apk"
    [[ -r "$artifact" ]] || fail "Missing APK: $artifact; build the requested architecture and mode first"
    preflight verify "$artifact" "$arch" "$mode"
    select_device
    app_id="$(config applicationId)"
    timeout 120 adb -s "$serial" install -r "$artifact"
    if [[ "${3:-}" == check || "${3:-}" == check-webview ]]; then
        device_sdk="$(timeout 10 adb -s "$serial" shell getprop ro.build.version.sdk | tr -d '\r')"
        if (( device_sdk >= 23 )); then
            permission=grant
            [[ "$3" != check-webview ]] || permission=revoke
            timeout 15 adb -s "$serial" shell pm "$permission" "$app_id" android.permission.RECORD_AUDIO
        fi
        timeout 15 adb -s "$serial" shell am force-stop "$app_id"
    elif [[ "${3:-}" == check-phone ]]; then
        echo 'Unlock the phone and allow microphone access when Android asks.' >&2
        timeout 15 adb -s "$serial" shell am force-stop "$app_id"
    fi
    launch="$(timeout 30 adb -s "$serial" shell am start -W -n "$app_id/com.wails.app.LauncherActivity")"
    printf '%s\n' "$launch"
    [[ "$launch" == *'Status: ok'* && "$launch" != *'Error:'* ]] || fail 'LauncherActivity did not start successfully'
    attempts=10
    [[ "${3:-}" != check-phone ]] || attempts=120
    for ((attempt=0; attempt<attempts; attempt++)); do
        pids="$(app_pids || true)"
        activities="$(timeout 10 adb -s "$serial" shell dumpsys activity activities)"
        if [[ -n "$pids" ]] && awk -v app="$app_id/" '/mResumedActivity|topResumedActivity|mFocusedActivity/ && index($0, app) {found=1} END {exit !found}' <<< "$activities"; then
            sleep 2
            [[ -n "$(app_pids || true)" ]] || fail 'App process exited immediately after launch; inspect Android logs'
            echo "Started $app_id on $serial (PID $pids)"
            return
        fi
        sleep 1
    done
    fail 'App did not reach a running process and resumed activity; inspect Android logs'
}

case "${1:-doctor}" in
    help)
        cat <<'HELP'
just emulator [AVD] [emulator options]  Start a virtual device
just android                          Build, install and launch on the selected device
just android check                    Test the existing debug APK on an emulator
just android logs                     Follow app logs
just adb [arguments]                  Run adb (default: list devices)

Set ANDROID_SERIAL when more than one device is connected.
Other Android actions: doctor, setup, avds, prepare, build [arm64|amd64],
install [arm64|amd64] [debug|release], package, bundle, phone-info,
check-phone, check-webview, webview-fixture, studio.
HELP
        ;;
    adb) shift; exec adb "$@" ;;
    run)
        select_device
        export ANDROID_SERIAL="$serial"
        device_abi="$(timeout 10 adb -s "$serial" shell getprop ro.product.cpu.abi | tr -d '\r')"
        case "$device_abi" in
            arm64-v8a) arch=arm64 ;;
            x86_64) arch=amd64 ;;
            *) fail "Unsupported device ABI: $device_abi" ;;
        esac
        bash "$android_support/android.sh" build "$arch"
        exec bash "$android_support/android.sh" install "$arch" debug
        ;;
    setup)
        sdkmanager --sdk_root="$ANDROID_HOME" --licenses
        sdkmanager --sdk_root="$ANDROID_HOME" platform-tools emulator \
            "platforms;android-$(config platform)" "build-tools;$(config buildTools)" "ndk;$(config ndk)"
        ;;
    prepare) lock_project; prepare ;;
    doctor)
        export GOFLAGS="${GOFLAGS:-} -mod=readonly"
        preflight doctor
        java -version
        emulator -accel-check
        emulator -list-avds
        ;;
    avds)
        mapfile -t entries < <(python3 -c 'import json,sys; [print(k+" "+v) for k,v in json.load(open(sys.argv[1]))["devices"].items()]' "$android_support/config.json")
        for entry in "${entries[@]}"; do
            read -r name image <<< "$entry"
            sdkmanager --sdk_root="$ANDROID_HOME" "$image"
            if ! emulator -list-avds | tr -d '\r' | grep -Fxq "$name"; then
                avdmanager create avd --name "$name" --package "$image" --device pixel_2 <<< no
            fi
        done
        ;;
    emulator)
        # The caller chooses when to show a window; automated verification uses -no-window.
        name="${2:-smeggtuner-api37}"
        if [[ $# -ge 2 ]]; then shift 2; else shift; fi
        if [[ "$name" == smeggtuner-api21 ]]; then
            # This image uses a real SD card, unlike newer emulated-storage images.
            avd_root="${ANDROID_AVD_HOME:-${ANDROID_USER_HOME:-$HOME/.android}/avd}"
            sdcard="$avd_root/$name.avd/sdcard.img"
            [[ -d "$(dirname -- "$sdcard")" ]] || fail 'Create the API 21 AVD with just android avds first'
            [[ -e "$sdcard" ]] || mksdcard 512M "$sdcard"
            set -- -sdcard "$sdcard" "$@"
        fi
        exec emulator -avd "$name" -no-snapshot-save "$@"
        ;;
    build|package|bundle)
        action="$1"
        mode=release
        [[ "$action" != build ]] || mode=debug
        select_artifact "${2:-arm64}" "$mode"
        if [[ "$mode" == release ]]; then
            ANDROID_KEYSTORE_FILE="$(preflight signing)"
            export ANDROID_KEYSTORE_FILE
        fi
        lock_project
        export GOFLAGS="${GOFLAGS:-} -mod=readonly"
        preflight doctor
        extension=apk
        [[ "$action" != bundle ]] || extension=aab
        artifact="$PWD/$output/$name.$extension"
        rm -f -- "$artifact"
        prepare
        export PRODUCTION=false
        [[ "$mode" != release ]] || export PRODUCTION=true
        (cd frontend && pnpm install --frozen-lockfile)
        android_node_bin="$(python3 -B "$android_support/../../scripts/node_env.py" --print-bin)"
        [[ -z "$android_node_bin" ]] || export PATH="$android_node_bin:$PATH"
        preflight doctor
        mkdir -p backend/frontend/public
        touch backend/frontend/public/.keep
        go tool frontendgen --out frontend/app/composables/s99wails --s99ab
        go tool wails3 generate bindings -f '-tags gtk3 -mod=readonly' -clean=true -ts -d frontend/bindings ./backend
        (cd frontend && pnpm run generate)
        rm -rf -- backend/frontend/public
        mkdir -p backend/frontend/public
        cp -a frontend/.output/public/. backend/frontend/public/
        python3 -B "$android_support/generate.py" --overlay
        source_library="$PWD/build/android/app/src/main/jniLibs/$abi/libwails.so"
        # A failed or skipped compilation must never stage a previous build.
        rm -f -- "$source_library"
        go tool wails3 task android:compile:go:shared "ARCH=$arch" "PRODUCTION=$PRODUCTION" "BIN_DIR=$output" "APP_NAME=$name"
        preflight native "$source_library" "$arch" "$mode"
        staging="$PWD/$output/jniLibs"
        rm -rf -- "$staging"
        mkdir -p -- "$staging/$abi"
        cp -- "$source_library" "$staging/$abi/libwails.so"
        task=android:assemble:apk
        [[ "$mode" != release ]] || task=android:assemble:apk:release
        if [[ "$action" == bundle ]]; then task=android:assemble:aab:release; fi
        env ORG_GRADLE_PROJECT_wailsJniDir="$staging" go tool wails3 task "$task" "ARCH=$arch" "PRODUCTION=$PRODUCTION" "BIN_DIR=$output" "APP_NAME=$name"
        if ! preflight verify "$artifact" "$arch" "$mode" "$staging/$abi/libwails.so"; then
            rm -f -- "$artifact"
            fail 'Android package verification failed; rejected artifact removed'
        fi
        ;;
    install) install_app "${2:-arm64}" "${3:-debug}" ;;
    phone-info) python3 -B "$android_support/phone.py" ;;
    check-phone)
        phone_info="$(python3 -B "$android_support/phone.py")"
        read -r serial arch <<< "$(python3 -c 'import json,sys; p=json.load(sys.stdin); print(p["serial"], p["arch"])' <<< "$phone_info")"
        export ANDROID_SERIAL="$serial"
        printf '%s\n' "$phone_info"
        install_app "$arch" debug check-phone
        node "$android_support/check.mjs" phone
        ;;
    check)
        select_device
        [[ "$serial" =~ ^emulator-[0-9]+$ ]] || fail 'Select a dedicated test emulator; use just android check-phone for a phone'
        export ANDROID_SERIAL="$serial"
        install_app "${2:-amd64}" debug check
        node "$android_support/check.mjs"
        ;;
    check-webview)
        select_device
        [[ "$serial" =~ ^emulator-[0-9]+$ ]] || fail 'Select a dedicated test emulator'
        export ANDROID_SERIAL="$serial"
        install_app "${2:-amd64}" debug check-webview
        python3 -B "$android_support/compatibility-check.py"
        ;;
    webview-fixture)
        lock_project
        python3 -B "$android_support/webview-fixture.py"
        ;;
    logs)
        select_device
        app_id="$(config applicationId)"
        pids="$(app_pids || true)"
        [[ -n "$pids" ]] || fail "No running $app_id process on $serial"
        echo "Streaming $app_id (PID $pids) on $serial" >&2
        # Include all tags from the process plus bridge and crash tags from other processes.
        adb -s "$serial" logcat -v threadtime | awk -v pids="$pids" '
            BEGIN {split(pids, values, /[[:space:]]+/); for (i in values) own[values[i]]=1}
            $3 in own || /WailsActivity|WailsBridge|WailsJSBridge|AndroidRuntime|GoLog/ {print; fflush()}'
        ;;
    studio)
        lock_project
        prepare
        if command -v studio >/dev/null; then
            exec studio "$PWD/build/android"
        fi
        exec "$HOME/.local/opt/android-studio/bin/studio" "$PWD/build/android"
        ;;
    *) fail "Unknown Android action: $1" ;;
esac
