# build/android-support/signing.sh
# Source from Bash; passwords remain in the build process environment only.
android_signing() {
    export ANDROID_KEYSTORE_FILE="${ANDROID_KEYSTORE_FILE:-$HOME/.android-keys/smeggtuner-upload.jks}"
    export ANDROID_KEY_ALIAS="${ANDROID_KEY_ALIAS:-smeggtuner}"
    if [[ ! -r "$ANDROID_KEYSTORE_FILE" ]]; then
        echo "Android upload keystore is not readable: $ANDROID_KEYSTORE_FILE" >&2
        return 1
    fi
    if [[ -z "${ANDROID_KEYSTORE_PASSWORD:-}" ]]; then
        if [[ ! -t 0 ]]; then
            echo 'Run just android-release in a terminal to enter the signing password, or set ANDROID_KEYSTORE_PASSWORD for automation.' >&2
            return 1
        fi
        IFS= read -r -s -p 'Android upload keystore password: ' ANDROID_KEYSTORE_PASSWORD || return 1
        printf '\n' >&2
        if [[ -z "$ANDROID_KEYSTORE_PASSWORD" ]]; then
            echo 'Android signing password must not be empty.' >&2
            return 1
        fi
    fi
    export ANDROID_KEYSTORE_PASSWORD
    export ANDROID_KEY_PASSWORD="${ANDROID_KEY_PASSWORD:-$ANDROID_KEYSTORE_PASSWORD}"
}
