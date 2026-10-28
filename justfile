# Daily commands; Android and lab operations share their existing runners.
set shell := ["bash", "-euo", "pipefail", "-c"]
set positional-arguments

app_name := "smeggtuner"
with_node := "python3 scripts/node_env.py"
version := `tr -d '\r\n' < VERSION`
# Every platform 1.4 ships; the Android APK needs the signing keystore from build/Android.md.
release_targets := "linux/amd64 linux/arm64 windows/amd64 android/arm64 web"

[private]
default:
    @just --list --unsorted

# Generate everything and build bin/smeggtuner (production or debug).
build mode="production":
    bash scripts/build.sh "$1"

# Build and open the app on your desktop.
run mode="production": (build mode)
    ./bin/smeggtuner

# Watch sources, rebuild and restart the debug app on your desktop.
dev:
    go tool wails3 tool watcher -config build/dev.yml

# Regenerate version, config types, locales, tray and app icons, bindings and Wails helpers.
generate:
    bash scripts/generate.sh

# Check generated drift, then build and run all repository checks.
check:
    bash scripts/check.sh

# Build and run Go, frontend and build-tooling tests.
test: build
    # The recordings suite runs without the race detector, as in scripts/check.sh.
    CGO_ENABLED=1 go test -race -tags gtk3 -parallel 4 -timeout 25m $(go list ./... | grep -vx smegg.me/smeggtuner/tests)
    CGO_ENABLED=1 go test -tags gtk3 -parallel 4 -timeout 25m ./tests/
    cd deploy/app && go test -race ./...
    {{ with_node }} pnpm --dir frontend test
    python3 -B -m unittest discover -s scripts -p 'test_*.py'

# Build the web edition (static site for app.smeggtuner.com) into dist/web.
web:
    bash scripts/web.sh dist/web

# Build the per-user Linux archive and checksum in dist/.
package: build
    scripts/package.sh bin/smeggtuner dist

# Build the archive and install it for this user; KDE and GNOME take the taskbar icon from its menu entry.
install: package
    #!/usr/bin/env bash
    set -euo pipefail
    stage=$(mktemp -d)
    trap 'rm -rf -- "$stage"' EXIT
    name="{{ app_name }}-{{ version }}-linux-$(go env GOARCH)"
    tar --zstd -xf "dist/$name.tar.zst" -C "$stage"
    "$stage/$name/install.sh"

# Set the version everywhere it is stamped (VERSION, Wails, macOS, Windows), e.g. just bump 1.4.1.
bump new:
    {{ with_node }} node scripts/sync-version.mjs "$1"

# Run every check, then build every shipped platform's packages and SHA256SUMS into dist/release/VERSION.
release targets=release_targets: check
    bash scripts/release.sh {{ targets }}

# Start an emulator; optional arguments are passed to the emulator.
emulator name="smeggtuner-api37" *args:
    bash build/android-support/android.sh emulator "$@"

# Build/install on ADB by default; also supports check, logs, build, setup and phone actions.
android action="run" *args:
    bash build/android-support/android.sh "$@"

# Run SDK adb (default: devices); ANDROID_SERIAL selects the device.
adb *args="devices":
    bash build/android-support/android.sh adb "$@"

# Run a lab operation; test syncs, builds and runs a target.
lab action="list" *args:
    #!/usr/bin/env bash
    set -euo pipefail
    if [[ "$1" == test ]]; then
        shift
        [[ $# -ge 1 ]] || { echo 'Usage: just lab test TARGET [options]' >&2; exit 2; }
        ./testlab/lab.sh sync "$1"
        ./testlab/lab.sh build "$1"
        exec ./testlab/lab.sh run "$@"
    fi
    exec ./testlab/lab.sh "$@"
