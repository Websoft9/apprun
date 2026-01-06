#!/bin/bash
# Test script to simulate CI build process locally
# Usage: ./scripts/test-ci-build.sh

set -e

REGISTRY="ghcr.io"
REPO_OWNER="websoft9"
BASE_IMAGE="${REGISTRY}/${REPO_OWNER}/apprun-base:latest"
APP_IMAGE="apprun:test-ci"

echo "🧪 Testing CI Build Process Locally"
echo "===================================="
echo ""

# Step 1: Check if base image exists locally
echo "📦 Step 1: Checking base image..."
if docker image inspect ${BASE_IMAGE} >/dev/null 2>&1; then
    echo "✅ Base image found: ${BASE_IMAGE}"
else
    echo "⚠️  Base image not found locally"
    echo "💡 Tagging local base image as GHCR path for testing..."
    docker tag apprun-base:latest ${BASE_IMAGE}
    echo "✅ Tagged: apprun-base:latest → ${BASE_IMAGE}"
fi
echo ""

# Step 2: Build with explicit build args (simulating CI)
echo "🔨 Step 2: Building with CI-like parameters..."
docker build \
    --build-arg BASE_IMAGE=${BASE_IMAGE} \
    --build-arg VERSION=test-$(date +%Y%m%d) \
    --build-arg COMMIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
    --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
    --file docker/Dockerfile \
    --tag ${APP_IMAGE} \
    --platform linux/amd64 \
    .

echo ""
echo "✅ Build completed successfully!"
echo ""

# Step 3: Verify the image
echo "🔍 Step 3: Verifying built image..."
docker images ${APP_IMAGE}
echo ""

# Step 4: Quick smoke test
echo "🚀 Step 4: Running smoke test..."
if docker run --rm ${APP_IMAGE} atlas version >/dev/null 2>&1; then
    echo "✅ Atlas CLI working in image"
else
    echo "❌ Atlas CLI test failed"
    exit 1
fi

echo ""
echo "🎉 All tests passed! CI build process is working correctly."
echo ""
echo "Next steps:"
echo "1. Push changes: git push origin main"
echo "2. Trigger 'Build Base Image' workflow (if not exists in GHCR)"
echo "3. Check 'Docker Build and Publish' workflow runs successfully"
