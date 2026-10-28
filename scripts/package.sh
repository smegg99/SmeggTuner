#!/usr/bin/env bash
# scripts/package.sh
#
# Builds OUTDIR/smeggtuner-VERSION-linux-ARCH.tar.zst and .sha256 for a per-user install.
# Usage: package.sh BINARY OUTDIR [ARCH]   (ARCH is the Go architecture of BINARY; default the host's)
set -euo pipefail

[ "$#" -eq 2 ] || [ "$#" -eq 3 ] || { echo "Usage: package.sh BINARY OUTDIR [ARCH]" >&2; exit 2; }
binary=$1
out_dir=$2
arch=${3:-$(go env GOARCH)}

repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
linux_dir="$repo_dir/build/linux"
version=$(tr -d '[:space:]' <"$repo_dir/VERSION")
[ -n "$version" ] || { echo "ERR: VERSION is empty" >&2; exit 1; }
[ -f "$binary" ] && [ -x "$binary" ] || { echo "ERR: no program at $binary, run just build first" >&2; exit 1; }
for command_name in tar zstd sha256sum; do
    command -v "$command_name" >/dev/null 2>&1 || { echo "ERR: missing command: $command_name" >&2; exit 127; }
done

name="smeggtuner-$version-linux-$arch"
stage=$(mktemp -d)
trap 'rm -rf -- "$stage"' EXIT
root="$stage/$name"
mkdir -m 0755 -- "$root"
install -m 0755 -- "$binary" "$root/smeggtuner"
install -m 0755 -- "$linux_dir/install.sh" "$linux_dir/uninstall.sh" "$root/"
install -m 0644 -- "$linux_dir/smeggtuner.desktop" "$root/"
install -m 0644 -- "$repo_dir/assets/smeggtuner-app-icon.svg" "$root/smeggtuner.svg"
printf '%s\n' "$version" >"$root/VERSION"
chmod 0644 -- "$root/VERSION"

# Reproducible archive (sorted, no owner, binary mtime), written to .part and renamed.
mkdir -p -- "$out_dir"
archive="$out_dir/$name.tar.zst"
trap 'rm -rf -- "$stage"; rm -f -- "$archive.part" "$archive.sha256.part"' EXIT
tar --sort=name --owner=0 --group=0 --numeric-owner --mtime="@$(stat -c %Y -- "$binary")" \
    -C "$stage" -cf - "$name" | zstd -q -19 -T0 -f -o "$archive.part"
mv -f -- "$archive.part" "$archive"
(cd -- "$out_dir" && sha256sum -- "$name.tar.zst" >"$name.tar.zst.sha256.part")
mv -f -- "$archive.sha256.part" "$archive.sha256"
echo "OK: package=$archive"
echo "OK: sha256=$(cut -d ' ' -f 1 <"$archive.sha256")"
