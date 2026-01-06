#!/bin/bash
# Verify Docker image manifest for unknown/unknown platforms
# Usage: ./scripts/verify-manifest.sh IMAGE_NAME

set -e

IMAGE="${1:-ghcr.io/websoft9/apprun-base:latest}"

echo "🔍 Verifying Docker Image Manifest"
echo "===================================="
echo "Image: $IMAGE"
echo ""

# Check if image exists
if ! docker pull "$IMAGE" >/dev/null 2>&1; then
    echo "⚠️  Cannot pull image: $IMAGE"
    echo "💡 Image may not exist yet or credentials needed"
    exit 1
fi

echo "✅ Image pulled successfully"
echo ""

# Inspect manifest
echo "📋 Manifest platforms:"
docker manifest inspect "$IMAGE" 2>/dev/null | jq -r '.manifests[] | "  - \(.platform.os)/\(.platform.architecture)"' || {
    echo "  Single-arch image (no manifest list)"
    docker inspect "$IMAGE" | jq -r '.[0] | "  - \(.Os)/\(.Architecture)"'
}

echo ""

# Check for unknown platforms
UNKNOWN_COUNT=$(docker manifest inspect "$IMAGE" 2>/dev/null | jq '[.manifests[] | select(.platform.os == "unknown" or .platform.architecture == "unknown")] | length' || echo 0)

if [ "$UNKNOWN_COUNT" -gt 0 ]; then
    echo "❌ Found $UNKNOWN_COUNT unknown/unknown platform entries"
    echo ""
    echo "🔧 Action needed:"
    echo "1. Go to GitHub Actions > Build Base Image"
    echo "2. Run workflow with 'Force rebuild' option"
    echo "3. This will overwrite the manifest with clean platforms"
    echo ""
    exit 1
else
    echo "✅ No unknown/unknown platforms found"
    echo "🎉 Manifest is clean!"
fi

echo ""
echo "📊 Image details:"
docker inspect "$IMAGE" | jq '.[0] | {
    Created: .Created,
    Size: (.Size / 1024 / 1024 | floor | tostring + " MB"),
    Architecture: .Architecture,
    Os: .Os,
    Layers: (.RootFS.Layers | length)
}'
