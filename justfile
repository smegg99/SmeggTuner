web_dir := "frontend"
app_name := "smeggtuner"
bin_dir := "bin"
config_file := "./build/config.yml"
vite_port := env("WAILS_VITE_PORT", "9245")

# GTK 3 (WebKitGTK 4.1) is the DEFAULT, and it is not a fallback.
#
# Wails builds against GTK 4 / WebKitGTK 6.0 unless told otherwise. That port is
# the newer one and it is the slower one here: its compositing path reallocates
# and re-copies the whole surface on a window drag, which on this app - four
# canvases, and a renderer that copies through the CPU whenever DMABUF is off -
# is the difference between a resize that is smooth and one that seizes.
#
# So the plain recipes build GTK 3, and the -gtk4 recipes build the other one,
# for the day that stops being true. Test both before believing either.
tags := "gtk3"

# The release version: the exact git tag with a leading `v` stripped, else 0.0.0-dev. Tag the
# commit (`git tag v1.0.0`) before a release so the app's own version badge matches. Override on
# a recipe: `just release 1.0.0`, `just package 1.0.0`.
version := `git describe --tags --exact-match 2>/dev/null | sed 's/^v//' | grep . || echo 0.0.0-dev`

# What `just release` builds. Cross targets (anything but this host's os/arch) need the wails
# cross-compile image once (`wails3 task setup:docker`); Windows and macOS are untested and off
# by default. Pass your own list to change it, e.g.
#   just release 1.0.0 "linux/amd64 linux/arm64 windows/amd64"
release_targets := "linux/amd64 linux/arm64"

# List available commands
default:
    @just --list

# Run the full Wails app in development mode (hot-reload, GTK 3)
dev:
    WAILS_BUILD_TAGS="{{tags}}" wails3 dev -config {{config_file}} -port {{vite_port}}

# Dev mode against the GTK 4 / WebKitGTK 6.0 stack
dev-gtk4:
    wails3 dev -config {{config_file}} -port {{vite_port}}

# Start only the Nuxt frontend dev server
dev-frontend:
    cd {{web_dir}} && pnpm dev

# Production build (current platform, GTK 3)
build:
    wails3 build -tags {{tags}}

# Production build against the GTK 4 / WebKitGTK 6.0 stack
build-gtk4:
    wails3 build

# Development build (fast, unoptimised, GTK 3)
build-dev:
    wails3 build -tags {{tags}} DEV=true

# Development build against the GTK 4 / WebKitGTK 6.0 stack
build-dev-gtk4:
    wails3 build DEV=true

# `wails3 package` takes no -tags flag, so the tag goes in through the variable
# the build tasks actually read (see build/linux/Taskfile.yml: EXTRA_TAGS).

# Package the app for the current platform (GTK 3), stamped with the version
package ver=version:
    VERSION="{{ver}}" EXTRA_TAGS="{{tags}}" wails3 package

# Package against the GTK 4 / WebKitGTK 6.0 stack
package-gtk4 ver=version:
    VERSION="{{ver}}" wails3 package

# Build and package release artifacts for every target in release_targets, named the way a
# release is: bin/release/<app>-<version>-<os>-<arch>.<ext>, plus the bare binary. A target that
# fails (usually a missing cross toolchain) is reported and skipped; the others still build.
#   just release                            # version from the exact git tag
#   just release 1.0.0                      # explicit version
#   just release 1.0.0 "linux/amd64 windows/amd64"
release ver=version targets=release_targets:
    #!/usr/bin/env bash
    set -uo pipefail
    shopt -s nullglob
    out="{{bin_dir}}/release"
    mkdir -p "$out"
    built=(); skipped=()
    for target in {{targets}}; do
        os="${target%%/*}"; arch="${target##*/}"
        echo "==> {{app_name}} {{ver}} :: ${os}/${arch}"
        rm -f "{{bin_dir}}/{{app_name}}" "{{bin_dir}}/{{app_name}}".deb "{{bin_dir}}/{{app_name}}".rpm \
              "{{bin_dir}}/{{app_name}}".pkg.tar.zst "{{bin_dir}}/{{app_name}}"-*.AppImage
        if VERSION="{{ver}}" EXTRA_TAGS="{{tags}}" wails3 task "${os}:package" ARCH="${arch}"; then
            for f in "{{bin_dir}}/{{app_name}}".deb "{{bin_dir}}/{{app_name}}".rpm \
                     "{{bin_dir}}/{{app_name}}".pkg.tar.zst "{{bin_dir}}/{{app_name}}"-*.AppImage; do
                case "$f" in
                    *.pkg.tar.zst) ext="pkg.tar.zst" ;;
                    *)             ext="${f##*.}" ;;
                esac
                mv -f "$f" "$out/{{app_name}}-{{ver}}-${os}-${arch}.${ext}"
            done
            [ -e "{{bin_dir}}/{{app_name}}" ] && cp -f "{{bin_dir}}/{{app_name}}" "$out/{{app_name}}-{{ver}}-${os}-${arch}"
            built+=("${os}/${arch}")
        else
            skipped+=("${os}/${arch}")
        fi
    done
    echo
    echo "release {{ver}} -> $out"
    ls -1sh "$out" 2>/dev/null || true
    [ ${#built[@]}   -gt 0 ] && echo "built:   ${built[*]}"
    [ ${#skipped[@]} -gt 0 ] && echo "skipped: ${skipped[*]}  (cross builds need once: wails3 task setup:docker)"
    [ ${#built[@]}   -gt 0 ]

# Run the built application
run:
    ./{{bin_dir}}/{{app_name}}

# Generate frontend bindings from Go services
# Delete every build artefact and cache. Nothing here is source; all of it is regenerated.
#
# The Nuxt cache is the one that matters, and it is the one people miss: it does NOT live
# in frontend/.nuxt, it lives in frontend/node_modules/.cache/nuxt. A stale one serves an
# old bundle out of a fresh build and there is nothing on screen to say so.
#
# node_modules itself is NOT touched: reinstalling it takes minutes and is almost never
# the problem. Use `just clean-deps` when it actually is.
clean:
    rm -rf {{bin_dir}}
    rm -rf {{web_dir}}/dist {{web_dir}}/.output {{web_dir}}/.nuxt
    rm -rf {{web_dir}}/node_modules/.cache {{web_dir}}/node_modules/.vite
    rm -rf {{web_dir}}/bindings
    go clean -cache -testcache
    @echo "cleaned: binaries, frontend build, nuxt/vite caches, bindings, go caches"

# Everything `clean` does, and the installed packages as well.
clean-deps: clean
    rm -rf {{web_dir}}/node_modules
    @echo "cleaned: node_modules. run `just build` to reinstall."

generate-bindings:
    wails3 generate bindings -ts -clean=true

# Generate Go code from CUE schemas
generate-cue:
    go generate ./...

# Generate app icons from build/appicon.png
#
# -iconcomposerinput is not standalone: without -macassetdir wails3 refuses the whole
# run with "mac asset directory is required" and nothing is written, not even the .ico.
# The flags mirror build/Taskfile.yml, which runs from build/ and so uses relative paths.
#
# On Linux this regenerates icon.ico and icons.icns but NOT darwin/Assets.car - that
# needs macOS actool, and wails3 skips it and still exits 0. The committed Assets.car
# is the only copy, so refresh it on a Mac if appicon.icon ever changes.
generate-icons:
    wails3 generate icons -input build/appicon.png -windowsfilename build/windows/icon.ico -iconcomposerinput build/appicon.icon -macfilename build/darwin/icons.icns -macassetdir build/darwin

# Run the Go test suite with the race detector.
#
# The timeout is not decoration. core/dsp sweeps the merged-pair fit across notes,
# beats, reed ratios, bellows strokes and drift, and under -race that alone runs
# past Go's 10 minute default. The suite then dies with a timeout panic that reads
# exactly like a test failure, and you go hunting for a bug that is not there.
test:
    go test ./... -race -timeout 25m

# Check a folder of recordings against the notes in their file names.
# Defaults to sounds/wav and the reference pitch the instrument was tuned to.
#   just samples
#   just samples sounds/wav 440
#
# sounds/ is NOT in the repository - it is ~85 MB of local reference recordings, and
# nothing needs it to build or test. A fresh clone has no sounds/wav, so this recipe
# needs a folder of your own. The committed fixtures (tests/fixtures/, with .expect.json
# goldens) are what `just test` gates on.
samples dir="sounds/wav" a4="442":
    # go test runs in the package directory, so the folder has to be absolute.
    SMEGGTUNER_SAMPLES="$(realpath {{dir}})" SMEGGTUNER_SAMPLES_A4="{{a4}}" \
        go test ./tests/ -run TestSamples -v -count=1

# Lint the frontend
lint:
    cd {{web_dir}} && pnpm lint

# Run the frontend test suite
test-frontend:
    cd {{web_dir}} && pnpm test

# Headless measurement harness, e.g. just analyze -wav tests/fixtures/a-8.wav
analyze *args:
    go run ./scripts/analyze {{args}}

# ---- test lab -------------------------------------------------------------
# Build and run the app on other distros and on Windows. See testlab/README.md.

# List the lab targets
lab-list:
    ./testlab/lab.sh list

# Build the whole lab from nothing (containers + Windows VMs). Unattended.
lab-init:
    ./testlab/lab.sh init

# Push the working tree into a target
lab-sync target:
    ./testlab/lab.sh sync {{target}}

# Build the app inside a target
lab-build target:
    ./testlab/lab.sh build {{target}}

# Launch the app in a target, optionally inside a nested desktop (--de gnome)
lab-run target *args:
    ./testlab/lab.sh run {{target}} {{args}}

# sync + build + run
lab-test target *args:
    ./testlab/lab.sh sync {{target}}
    ./testlab/lab.sh build {{target}}
    ./testlab/lab.sh run {{target}} {{args}}

# Interactive shell inside a target
lab-shell target:
    ./testlab/lab.sh shell {{target}}

lab-down target:
    ./testlab/lab.sh down {{target}}

# Feed a WAV fixture into every target's microphone
mic file *args:
    ./testlab/mic/vmic.sh play {{file}} {{args}}

mic-up:
    ./testlab/mic/vmic.sh up

mic-down:
    ./testlab/mic/vmic.sh down
