#!/usr/bin/env bash
# scripts/generate.sh
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."

# Bindings load the main package before the frontend has been built.
mkdir -p backend/frontend/public
touch backend/frontend/public/.keep
python3 scripts/node_env.py node scripts/sync-version.mjs
go generate ./backend/common/... ./backend/internal/...
go generate ./backend
rsvg-convert -w 1024 -h 1024 assets/smeggtuner-app-icon.svg -o build/appicon.png
# The macOS icon bundle carries its own copy of the glyph; icon.json sets the tile.
cp assets/smeggtuner-cat.svg build/appicon.icon/Assets/smeggtuner-cat.svg
go tool wails3 generate icons -input build/appicon.png -windowsfilename build/windows/icon.ico \
    -iconcomposerinput build/appicon.icon -macfilename build/darwin/icons.icns -macassetdir build/darwin
