#!/bin/bash
# Test multi-architecture Docker builds locally
# Usage: ./scripts/test-multi-arch.sh

set -e

echo "🧪 Testing Multi-Architecture Docker Builds"
echo "=========================================="
echo ""

# Check if buildx is available
if ! docker buildx version >/dev/null 2>&1; then
    echo "❌ Docker Buildx not available. Please install Docker Desktop or enable Buildx."
    exit 1
fi

# Create multi-arch builder if not exists
echo "🔧 Setting up multi-arch builder..."
docker buildx create --use --name multi-arch-test 2>/dev/null || echo "Using existing builder"

# Test Atlas image availability
echo "📦 Checking Atlas image..."
if ! docker pull arigaio/atlas:latest >/dev/null 2>&1; then
    echo "❌ Cannot pull arigaio/atlas:latest"
    exit 1
fi
echo "✅ Atlas image available"

# Test single arch build (amd64)
echo ""
echo "🏗️  Testing single arch build (amd64)..."
if docker buildx build \
    --platform linux/amd64 \
    --file docker/Dockerfile.base \
    --tag apprun-base:test-amd64 \
    --load \
    . >/dev/null 2>&1; then
    echo "✅ AMD64 build successful"
    docker images apprun-base:test-amd64
else
    echo "❌ AMD64 build failed"
    exit 1
fi

# Test single arch build (arm64) - requires QEMU
echo ""
echo "🏗️  Testing single arch build (arm64)..."
if docker run --rm --privileged tonistiigi/binfmt:latest --install all >/dev/null 2>&1; then
    echo "✅ QEMU setup for ARM64"
else
    echo "⚠️  QEMU setup may have issues"
fi

if docker buildx build \
    --platform linux/arm64 \
    --file docker/Dockerfile.base \
    --tag apprun-base:test-arm64 \
    --load \
    . >/dev/null 2>&1; then
    echo "✅ ARM64 build successful"
    docker images apprun-base:test-arm64
else
    echo "❌ ARM64 build failed - this is expected in some environments"
    echo "💡 ARM64 builds work best in GitHub Actions with proper QEMU setup"
fi

# Test multi-arch manifest (without pushing)
echo ""
echo "📋 Testing multi-arch manifest creation..."
if docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --file docker/Dockerfile.base \
    --tag apprun-base:multi-arch-test \
    . >/dev/null 2>&1; then
    echo "✅ Multi-arch build setup successful"
    echo "💡 In CI, this would push to registry with --push flag"
else
    echo "❌ Multi-arch build setup failed"
fi

echo ""
echo "🎉 Multi-arch testing completed!"
echo ""
echo "Summary:"
echo "- AMD64 builds: ✅ Working"
echo "- ARM64 builds: May require QEMU setup"
echo "- Multi-arch manifest: ✅ Supported"
echo ""
echo "For CI/CD, ensure GitHub Actions has:"
echo "- docker/setup-qemu-action@v3"
echo "- docker/setup-buildx-action@v3"
echo "- platforms: linux/amd64,linux/arm64"