#!/usr/bin/env bash
# scripts/web.sh
#
# Builds the web edition, a static site, into OUT (default dist/web): the Nuxt app in web mode, the engine
# wasm with the Go loader that matches it, the PWA icons, a service worker stamped with this build, and
# precompressed copies of every text and wasm file.
# Usage: web.sh [OUT]
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."

out=${1:-dist/web}
node=(python3 scripts/node_env.py)
version=$(tr -d '[:space:]' <VERSION)
# Nuxt serves this directory at / in web mode (nuxt.config.ts), so dev and generate both find /engine/ and /icons/.
assets=frontend/.web-assets
icon=assets/smeggtuner-app-icon.svg

bash scripts/generate.sh
rm -rf -- "$assets"
mkdir -p -- "$assets/engine" "$assets/icons"
GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags "-s -w" -o "$assets/engine/smeggtuner.wasm" ./backend/web/wasm
# The loader must come from the Go that built the wasm; their ABI changes between releases.
cp -- "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$assets/engine/wasm_exec.js"

for size in 32 180 192 512; do
    rsvg-convert -w "$size" -h "$size" "$icon" -o "$assets/icons/icon-$size.png"
done
magick "$assets/icons/icon-32.png" "$assets/favicon.ico"
# Maskable icons are cropped to a circle by some launchers: the glyph sits in the middle 80% on a solid tile.
rsvg-convert -w 410 -h 410 --page-width 512 --page-height 512 --left 51 --top 51 -b '#000000' "$icon" \
    -o "$assets/icons/icon-maskable-512.png"

CI=true "${node[@]}" pnpm --dir frontend install --frozen-lockfile
SMEGGTUNER_EDITION=web "${node[@]}" pnpm --dir frontend generate

rm -rf -- "$out"
mkdir -p -- "$out"
cp -a frontend/.output-web/public/. "$out/"

# A new cache per build: the service worker drops every other smeggtuner-web-* cache when it activates.
build=$(cat -- "$out/index.html" "$out/engine/smeggtuner.wasm" | sha256sum | cut -c1-12)
sed -i "s/__BUILD__/$version-$build/" "$out/sw.js"

find "$out" -type f \( -name '*.html' -o -name '*.js' -o -name '*.css' -o -name '*.json' -o -name '*.svg' \
    -o -name '*.webmanifest' -o -name '*.wasm' \) -print0 |
    while IFS= read -r -d '' f; do
        gzip -9 -k -f -- "$f"
        brotli -q 11 -k -f -- "$f"
    done

echo "web edition $version-$build -> $out ($(du -sh -- "$out" | cut -f1))"
