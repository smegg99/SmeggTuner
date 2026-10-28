#!/usr/bin/env bash
# scripts/build.sh
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."

mode=${1:-production}
binary=bin/smeggtuner
case "$mode" in
    production) tags=gtk3,production; ldflags="-s -w"; flags=(-trimpath) ;;
    debug) tags=gtk3; ldflags=""; flags=(-gcflags=all=-l) ;;
    *) echo "ERR: build mode must be production or debug" >&2; exit 2 ;;
esac

node=(python3 scripts/node_env.py)
bash scripts/generate.sh
CI=true "${node[@]}" pnpm --dir frontend install --frozen-lockfile
"${node[@]}" pnpm --dir frontend generate
rm -rf -- backend/frontend/public
mkdir -p backend/frontend/public bin
cp -a frontend/.output/public/. backend/frontend/public/
CGO_ENABLED=1 go build -tags "$tags" -buildvcs=false "${flags[@]}" -ldflags "$ldflags" -o "$binary" ./backend
