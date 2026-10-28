#!/usr/bin/env bash
# deploy/app/fetch.sh
#
# Puts the web edition of this checkout's VERSION into deploy/app/site, from that version's GitHub release,
# after checking it against the release's SHA256SUMS. depl runs it as the first build step.
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.."

version=$(tr -d '[:space:]' <VERSION)
name=smeggtuner-$version-web.tar.zst
release=https://github.com/smegg99/Smegg-s-professional-Accordion-Tuner/releases/download/v$version
work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT

curl -fsSL --retry 3 -o "$work/$name" "$release/$name"
curl -fsSL --retry 3 -o "$work/SHA256SUMS" "$release/SHA256SUMS"
grep -E "^[0-9a-f]{64}  $name\$" "$work/SHA256SUMS" >"$work/check" ||
    { echo "ERR: $name is not in the release's SHA256SUMS" >&2; exit 1; }
(cd -- "$work" && sha256sum -c --quiet check)

mkdir -p -- "$work/site"
tar --zstd -xf "$work/$name" -C "$work/site" --strip-components=1
[[ -s "$work/site/index.html" ]] || { echo "ERR: $name has no index.html" >&2; exit 1; }
rm -rf -- deploy/app/site.old
[[ ! -e deploy/app/site ]] || mv -- deploy/app/site deploy/app/site.old
mv -- "$work/site" deploy/app/site
rm -rf -- deploy/app/site.old
echo "web edition $version -> deploy/app/site"
