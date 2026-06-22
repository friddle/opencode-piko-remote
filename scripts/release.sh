#!/usr/bin/env bash
# Delete and recreate a GitHub release for opencode-piko-remote.
#
# Usage:
#   ./scripts/release.sh [version]          # default: 0.1.0
#
# Prereqs:
#   - client/dist/ must contain the 4 built binaries (run `make build-all` in client/)
#   - gh CLI authenticated
#
set -euo pipefail

cd "$(dirname "$0")/.."

VERSION="${1:-0.1.0}"
TAG="v${VERSION}"
REPO="friddle/opencode-piko-remote"
DIST_DIR="client/dist"

BINARIES=(
    "${DIST_DIR}/opencode-piko-darwin-amd64"
    "${DIST_DIR}/opencode-piko-darwin-arm64"
    "${DIST_DIR}/opencode-piko-linux-amd64"
    "${DIST_DIR}/opencode-piko-linux-arm64"
)

RED='\033[0;31m'
GREEN='\033[0;32m'
MUTED='\033[0;2m'
NC='\033[0m'

echo -e "${MUTED}=== Releasing ${TAG} ===${NC}"

# Verify binaries exist
for bin in "${BINARIES[@]}" install.sh; do
    if [ ! -f "$bin" ]; then
        echo -e "${RED}Error: Missing ${bin}${NC}"
        echo -e "${MUTED}Build first: cd client && make build-all${NC}"
        exit 1
    fi
done

# Delete existing release + tag (ignore errors if not found)
if gh release view "$TAG" >/dev/null 2>&1; then
    echo -e "${MUTED}Deleting existing release ${TAG}...${NC}"
    gh release delete "$TAG" --yes --cleanup-tag
else
    echo -e "${MUTED}No existing release ${TAG}, skipping delete.${NC}"
fi

# Create new release with all assets
echo -e "${MUTED}Creating release ${TAG}...${NC}"
gh release create "$TAG" \
    --title "$TAG" \
    --notes-file /tmp/v0.1.0_notes.md \
    "${BINARIES[@]}" \
    install.sh

echo -e "${GREEN}Done. ${TAG} released.${NC}"
echo -e "${MUTED}https://github.com/${REPO}/releases/tag/${TAG}${NC}"
