#!/bin/bash

# Docker build script for multi-platform support

set -e

VERSION=${1:-latest}
IMAGE_NAME="jnieto1908/payments-ms-go"

echo "🐳 Building Docker image..."
echo "  Image: $IMAGE_NAME:$VERSION"
echo "  Platform: linux/amd64"

# Enable BuildKit
export DOCKER_BUILDKIT=1

# Build for linux/amd64 (compatible with most cloud platforms)
docker buildx build \
    --platform linux/amd64 \
    -t "$IMAGE_NAME:$VERSION" \
    -t "$IMAGE_NAME:latest" \
    .

echo ""
echo "✅ Docker image built successfully!"
echo ""
echo "To push to Docker Hub:"
echo "  docker login"
echo "  docker push $IMAGE_NAME:$VERSION"
echo "  docker push $IMAGE_NAME:latest"
echo ""
echo "To run locally:"
echo "  docker run -p 8080:8080 --name payments-ms-go $IMAGE_NAME:$VERSION"
