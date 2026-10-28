#!/usr/bin/env bash
# scripts/check.sh
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."

node=(python3 scripts/node_env.py)
python3 -B scripts/check-generated.py
bash scripts/build.sh

unformatted=$(gofmt -l backend tests scripts build/packaging_test.go)
if [ -n "$unformatted" ]; then
    printf 'ERR: gofmt would rewrite:\n%s\n' "$unformatted" >&2
    exit 1
fi
CGO_ENABLED=1 go vet -tags gtk3 ./...
"${node[@]}" node scripts/sync-version.mjs --check
go tool loc check -c i18n/config.cue
"${node[@]}" node scripts/check-locales.mjs
CGO_ENABLED=1 go test -race -tags gtk3 -parallel 4 -timeout 25m ./...
"${node[@]}" pnpm --dir frontend test
python3 -B -m unittest discover -s scripts -p 'test_*.py'
"${node[@]}" pnpm --dir frontend typecheck
"${node[@]}" pnpm --dir frontend lint
