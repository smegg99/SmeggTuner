#!/usr/bin/env bash
# scripts/check.sh
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."

node=(python3 scripts/node_env.py)
python3 -B scripts/check-generated.py
bash scripts/build.sh

unformatted=$(gofmt -l backend tests scripts deploy build/packaging_test.go)
if [ -n "$unformatted" ]; then
    printf 'ERR: gofmt would rewrite:\n%s\n' "$unformatted" >&2
    exit 1
fi
CGO_ENABLED=1 go vet -tags gtk3 ./...
"${node[@]}" node scripts/sync-version.mjs --check
go tool loc check -c i18n/config.cue
"${node[@]}" node scripts/check-locales.mjs
# The recordings suite replays its audio through the engine, which the race detector slows about
# eightfold; the engine's goroutines are raced by core/dsp's own tests, so the suite runs without it.
CGO_ENABLED=1 go test -race -tags gtk3 -parallel 4 -timeout 25m $(go list ./... | grep -vx smegg.me/smeggtuner/tests)
CGO_ENABLED=1 go test -tags gtk3 -parallel 4 -timeout 25m ./tests/
# The web host is its own module, so ./... above leaves it out.
(cd deploy/app && go vet ./... && go test -race ./...)
"${node[@]}" pnpm --dir frontend test
python3 -B -m unittest discover -s scripts -p 'test_*.py'
"${node[@]}" pnpm --dir frontend typecheck
"${node[@]}" pnpm --dir frontend lint
