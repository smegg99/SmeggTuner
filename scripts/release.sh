#!/usr/bin/env bash
# scripts/release.sh
#
# Builds the release artifacts of each TARGET into dist/release/VERSION and writes SHA256SUMS there.
# Usage: release.sh TARGET...   (linux/amd64 linux/arm64 windows/amd64 android/arm64 android/amd64 web)
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."

[ "$#" -gt 0 ] || { echo "Usage: release.sh TARGET..." >&2; exit 2; }
app=smeggtuner
version=$(tr -d '[:space:]' <VERSION)
[ -n "$version" ] || { echo "ERR: VERSION is empty" >&2; exit 1; }
out=dist/release/$version
lab=${WINDOWS_LAB:-win11}
node=(python3 scripts/node_env.py)
shopt -s nullglob
mkdir -p -- "$out"
built=()
failed=()

require() {
    [[ -s "$1" ]] || { echo "ERR: missing or empty release artifact: $1" >&2; exit 1; }
}

# AppImage, deb, rpm and pacman packages from the Wails tasks, plus the per-user archive. Another
# architecture is linked in a multiarch container and has no AppImage: linuxdeploy collects a
# binary's libraries by running them.
linux() {
    local arch=$1 name="$app-$version-linux-$1" f host
    rm -f -- "bin/$app" "bin/$app.deb" "bin/$app.rpm" "bin/$app.pkg.tar.zst" bin/"$app"-*.AppImage
    host=$(uname -m)
    case "$host" in x86_64) host=amd64 ;; aarch64) host=arm64 ;; esac
    if [[ "$arch" == "$host" ]]; then
        VERSION=$version EXTRA_TAGS=gtk3 CI=true "${node[@]}" go tool wails3 task linux:package ARCH="$arch" || return 1
        local images=(bin/"$app"-*.AppImage)
        [[ ${#images[@]} == 1 ]] || { echo "ERR: expected one AppImage for linux/$arch, found ${#images[@]}" >&2; exit 1; }
        require "${images[0]}"
        mv -f -- "${images[0]}" "$out/$name.AppImage"
    else
        bash scripts/linux-cross.sh "$arch" "bin/$app" || return 1
        for f in deb rpm aur; do
            VERSION=$version CI=true "${node[@]}" go tool wails3 task "linux:generate:$f" ARCH="$arch" || return 1
        done
    fi
    for f in "bin/$app" "bin/$app.deb" "bin/$app.rpm" "bin/$app.pkg.tar.zst"; do require "$f"; done
    mv -f -- "bin/$app.deb" "$out/$name.deb"
    mv -f -- "bin/$app.rpm" "$out/$name.rpm"
    mv -f -- "bin/$app.pkg.tar.zst" "$out/$name.pkg.tar.zst"
    scripts/package.sh "bin/$app" "$out" "$arch"
    require "$out/$name.tar.zst"
    rm -f -- "$out/$name.tar.zst.sha256"
}

# The portable program and the NSIS installer, built in the Windows lab VM (amd64 only).
windows() {
    local arch=$1 name="$app-$version-windows-$1" f
    [[ "$arch" == amd64 ]] || { echo "ERR: no windows/$arch lab VM" >&2; return 1; }
    ./testlab/lab.sh sync "$lab" && ./testlab/lab.sh package "$lab" || return 1
    local portable=() installers=()
    for f in "bin/lab-out/$lab"/*.exe; do
        case "$f" in *installer*) installers+=("$f") ;; *) portable+=("$f") ;; esac
    done
    [[ ${#portable[@]} == 1 && ${#installers[@]} == 1 ]] ||
        { echo "ERR: expected one program and one installer for windows/$arch" >&2; exit 1; }
    require "${portable[0]}"
    require "${installers[0]}"
    mv -f -- "${portable[0]}" "$out/$name.exe"
    mv -f -- "${installers[0]}" "$out/$name-installer.exe"
}

# The signed release APK and app bundle; the keystore comes from the Android preflight.
android() {
    local arch=$1 name="$app-$version-android-$1"
    local base="bin/android/release/$arch/$app-$version-release-$arch"
    bash build/android-support/android.sh package "$arch" || return 1
    bash build/android-support/android.sh bundle "$arch" || return 1
    require "$base.apk"
    require "$base.aab"
    cp -f -- "$base.apk" "$out/$name.apk"
    cp -f -- "$base.aab" "$out/$name.aab"
}

# The web edition's static site as one archive, deployed to app.smeggtuner.com as is.
web() {
    local name="$app-$version-web" site
    site=$(mktemp -d)
    bash scripts/web.sh "$site/$name" || { rm -rf -- "$site"; return 1; }
    tar --zstd -cf "$out/$name.tar.zst" -C "$site" "$name"
    rm -rf -- "$site"
    require "$out/$name.tar.zst"
}

for target in "$@"; do
    os=${target%%/*}
    arch=${target##*/}
    echo "==> $app $version :: $target"
    rm -f -- "$out/$app-$version-$os-$arch".* "$out/$app-$version-$os-$arch-installer".*
    [[ "$target" != web ]] || rm -f -- "$out/$app-$version-web".*
    case "$os" in
        linux | windows | android | web)
            # A subshell outside any condition, so set -e still stops the target at its first failure.
            set +e
            (set -e; "$os" "$arch")
            status=$?
            set -e
            if [ "$status" -eq 0 ]; then built+=("$target"); else failed+=("$target"); fi
            ;;
        *) echo "ERR: no release path for $target" >&2; failed+=("$target") ;;
    esac
done

files=()
for f in "$out"/*; do
    [[ "$(basename -- "$f")" == SHA256SUMS ]] || files+=("$(basename -- "$f")")
done
if [ ${#files[@]} -gt 0 ]; then
    (cd -- "$out" && sha256sum -- "${files[@]}" >SHA256SUMS.part && mv -f -- SHA256SUMS.part SHA256SUMS)
fi

echo
echo "release $version -> $out"
ls -1sh -- "$out"
[ ${#built[@]} -eq 0 ] || echo "built:  ${built[*]}"
[ ${#failed[@]} -eq 0 ] || echo "failed: ${failed[*]}" >&2
[[ ${#built[@]} -gt 0 && ${#failed[@]} -eq 0 ]]
