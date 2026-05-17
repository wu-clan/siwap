#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
WAILS3="${WAILS3:-$(go env GOPATH)/bin/wails3}"

echo "==> frontend build"
(cd frontend && pnpm install && pnpm build)

echo "==> go tests"
go test ./...

echo "==> Wails v3 build smoke"
"$WAILS3" task build VERSION=dev

if [[ "$(uname -s)" == "Darwin" ]]; then
  IDENTITY="${SIWAP_DEV_CODESIGN_IDENTITY:-Siwap Local Development}"
  APP=""
  [[ -f "bin/siwap" ]] && APP="bin/siwap"
  if [[ -n "$APP" ]] && security find-identity -v -p codesigning | grep -Fq "$IDENTITY"; then
    echo "==> codesigning dev binary with fixed identity: $IDENTITY"
    codesign --force --sign "$IDENTITY" "$APP"
  else
    echo "==> skipping fixed dev codesign; create '$IDENTITY' or set SIWAP_DEV_CODESIGN_IDENTITY"
  fi
fi

echo "Dev build ready at bin"
