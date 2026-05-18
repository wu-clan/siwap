#!/usr/bin/env bash
set -euo pipefail
VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
  echo "usage: scripts/release.sh <version> [arch]" >&2
  exit 2
fi
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
WAILS3="${WAILS3:-$(command -v wails3 2>/dev/null || command -v wails3.exe 2>/dev/null || printf "%s/bin/wails3" "$(go env GOPATH)")}"
ARCH="${2:-${ARCH:-$(go env GOARCH)}}"

echo "==> building frontend assets"
"$WAILS3" task common:build:frontend

echo "==> testing Siwap $VERSION"
if [[ "$(go env GOOS)" == "linux" ]] && command -v xvfb-run >/dev/null 2>&1; then
  xvfb-run -a go test ./...
else
  go test ./...
fi

GOOS_VALUE="$(go env GOOS)"
echo "==> packaging Siwap $VERSION for $GOOS_VALUE/$ARCH"
"$WAILS3" task package VERSION="$VERSION" ARCH="$ARCH"

APP="bin/siwap.app"
if [[ "$GOOS_VALUE" == "darwin" && -d "$APP" && -n "${SIWAP_CODESIGN_IDENTITY:-}" ]]; then
  echo "==> signing with $SIWAP_CODESIGN_IDENTITY"
  codesign --force --deep --options runtime --sign "$SIWAP_CODESIGN_IDENTITY" "$APP"
  codesign --verify --strict --deep "$APP"
else
  echo "==> skipping codesign; set SIWAP_CODESIGN_IDENTITY to sign macOS app bundles"
fi

echo "Release build completed. Create a GitHub release and attach artifacts from bin."
