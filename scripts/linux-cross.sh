#!/usr/bin/env bash
# scripts/linux-cross.sh
#
# Builds the desktop binary for another Linux architecture against that architecture's own GTK and
# WebKitGTK, from Debian's multiarch packages and its cross compiler; no emulation is needed. The
# host's Go and module cache are mounted read-only, go.mod's local replacements beside the tree.
# Usage: linux-cross.sh ARCH OUTPUT   (ARCH: arm64)
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."

arch=${1:?ARCH}
output=${2:?OUTPUT}
case "$arch" in
    arm64) debian=arm64 triplet=aarch64-linux-gnu ;;
    *) echo "ERR: no cross toolchain for linux/$arch" >&2; exit 2 ;;
esac
[[ -s backend/frontend/public/index.html ]] || { echo "ERR: build the frontend first (just build)" >&2; exit 1; }

root=$(pwd)
mounts=(--volume "$root:/src/$(basename "$root"):ro")
while read -r sibling; do
    mounts+=(--volume "$(cd -- "$root/../$sibling" && pwd):/src/$sibling:ro")
done < <(go mod edit -json | python3 -c 'import json,sys; [print(r["New"]["Path"].split("/")[1]) for r in json.load(sys.stdin).get("Replace") or [] if not r["New"].get("Version")]' | sort -u)

out=$(mktemp -d)
trap 'rm -rf -- "$out"' EXIT
docker run --rm "${mounts[@]}" \
    --volume "$(go env GOROOT):/usr/local/go:ro" --volume "$(go env GOMODCACHE):/gomod:ro" --volume "$out:/out" \
    --env DEBIAN=$debian --env TRIPLET=$triplet --env GOARCH="$arch" --env APP_DIR="/src/$(basename "$root")" \
    --env OWNER="$(id -u):$(id -g)" debian:13 bash -c '
set -euo pipefail
dpkg --add-architecture "$DEBIAN"
apt-get update -qq
DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends ca-certificates pkg-config \
    "gcc-$TRIPLET" "libc6-dev-$DEBIAN-cross" "libgtk-3-dev:$DEBIAN" "libwebkit2gtk-4.1-dev:$DEBIAN" >/dev/null
export PATH=/usr/local/go/bin:$PATH GOMODCACHE=/gomod GOCACHE=/tmp/gocache GOTOOLCHAIN=local GOPROXY=off
export CGO_ENABLED=1 GOOS=linux CC="$TRIPLET-gcc" PKG_CONFIG_LIBDIR="/usr/lib/$TRIPLET/pkgconfig:/usr/share/pkgconfig"
cd "$APP_DIR"
go build -tags gtk3,production -buildvcs=false -trimpath -ldflags "-s -w" -o /out/binary ./backend
chown "$OWNER" /out/binary'
mkdir -p -- "$(dirname -- "$output")"
mv -f -- "$out/binary" "$output"
echo "linux/$arch -> $output"
