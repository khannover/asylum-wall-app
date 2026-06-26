#!/usr/bin/env bash
set -euo pipefail

IMAGE="ghcr.io/khannover/asylum-wall-app"
TAG="${1:-latest}"
SHA="$(git rev-parse --short HEAD 2>/dev/null || echo local)"

echo "Logging in to ghcr.io..."
gh auth token | docker login ghcr.io -u "$(gh api user -q .login)" --password-stdin

echo "Building ${IMAGE}:${TAG}..."
docker build -t "${IMAGE}:${TAG}" -t "${IMAGE}:${SHA}" .

echo "Pushing..."
docker push "${IMAGE}:${TAG}"
docker push "${IMAGE}:${SHA}"

echo "Done: ${IMAGE}:${TAG} and ${IMAGE}:${SHA}"
echo "If the package is private, make it public:"
echo "  https://github.com/users/$(gh api user -q .login)/packages/container/asylum-wall-app/settings"