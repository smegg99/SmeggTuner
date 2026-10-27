# build/android-support/env.sh
# Source from Bash. Scope Java, Node and Android paths to Android commands.
export ANDROID_HOME="${ANDROID_HOME:-$HOME/Android/Sdk}"
export ANDROID_NDK_HOME="$ANDROID_HOME/ndk/$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["ndk"])' "$android_support/config.json")"
if [[ -n "${ANDROID_JAVA_HOME:-}" ]]; then
    export JAVA_HOME="$ANDROID_JAVA_HOME"
elif [[ -d /usr/lib/jvm/java-21-openjdk ]]; then
    export JAVA_HOME=/usr/lib/jvm/java-21-openjdk
elif [[ -d "$HOME/.local/opt/android-studio/jbr" ]]; then
    export JAVA_HOME="$HOME/.local/opt/android-studio/jbr"
elif [[ -z "${JAVA_HOME:-}" ]]; then
    echo "Set ANDROID_JAVA_HOME to a Gradle-compatible JDK (21 recommended)." >&2
    exit 1
fi
export PATH="$JAVA_HOME/bin:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools:$ANDROID_HOME/emulator:$PATH"
if [[ "${1:-doctor}" =~ ^(doctor|build|package|bundle|check|check-phone)$ ]]; then
    android_node_bin="$(python3 -B "$android_support/../../scripts/node_env.py" --print-bin)"
    if [[ -n "$android_node_bin" ]]; then
        export PATH="$android_node_bin:$PATH"
    fi
fi
