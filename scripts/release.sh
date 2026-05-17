#!/usr/bin/env bash
set -euo pipefail
VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
  echo "usage: scripts/release.sh <version>" >&2
  exit 2
fi
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
WAILS3="${WAILS3:-$(go env GOPATH)/bin/wails3}"

echo "==> building Siwap $VERSION"
"$WAILS3" task build
go test ./...

APP="bin/siwap"
if [[ -f "$APP" && -n "${SIWAP_CODESIGN_IDENTITY:-}" ]]; then
  echo "==> signing with $SIWAP_CODESIGN_IDENTITY"
  codesign --force --options runtime --sign "$SIWAP_CODESIGN_IDENTITY" "$APP"
  codesign --verify --strict "$APP"
else
  echo "==> skipping codesign; set SIWAP_CODESIGN_IDENTITY to sign macOS binaries"
fi

echo "Release build completed. Create a GitHub release and attach artifacts from bin."
