#!/usr/bin/env bash
# The caller prepares the embedded frontend and icon before calling this script.
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."
version=$(tr -d '[:space:]' < VERSION)
cache=${XDG_CACHE_HOME:-$HOME/.cache}/smeggtuner/appimage-go
mkdir -p bin "$cache"
work=$(mktemp -d "$PWD/bin/.appimage.XXXXXXXX")
trap 'rm -rf -- "$work"' EXIT

# Wails only orchestrates packaging. Compile its CLI statically so it can run inside Jammy.
CGO_ENABLED=0 go build -trimpath -o "$work/wails3" github.com/wailsapp/wails/v3/cmd/wails3
go mod download
docker build --platform linux/amd64 -t smeggtuner-appimage:jammy - <build/linux/appimage.Dockerfile
docker run --rm --platform linux/amd64 --network none --user "$(id -u):$(id -g)" \
    -v "$PWD:/src:ro" -v "$work:/work" -v "$cache:/gocache" \
    -v "$(go env GOROOT):/opt/go:ro" -v "$(go env GOMODCACHE):/gomod:ro" \
    -w /src -e HOME=/tmp -e GOCACHE=/gocache -e GOMODCACHE=/gomod -e GOTOOLCHAIN=local -e GOPROXY=off \
    -e CGO_ENABLED=1 -e NO_STRIP=1 -e LDAI_RUNTIME_FILE=/opt/runtime-x86_64 -e LDAI_NO_APPSTREAM=1 \
    smeggtuner-appimage:jammy bash -euc '
        /opt/go/bin/go build -tags production,gtk3 -trimpath -buildvcs=false -ldflags="-s -w" -o /work/smeggtuner ./backend
        mkdir /work/build
        cp /opt/linuxdeploy.AppImage /work/build/linuxdeploy-x86_64.AppImage
        cp build/appicon.png /work/smeggtuner.png
        /work/wails3 generate appimage -binary /work/smeggtuner -icon /work/smeggtuner.png \
            -desktopfile /src/build/linux/smeggtuner.desktop -outputdir /work -builddir /work/build
    '
image=${1:-bin/SmeggTuner-$version-x86_64.AppImage}
test -s "$work/smeggtuner-x86_64.AppImage"
mv -f -- "$work/smeggtuner-x86_64.AppImage" "$image"
echo "AppImage -> $image"
