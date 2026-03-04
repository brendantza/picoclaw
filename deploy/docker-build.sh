#!/bin/bash
#
# PicoClaw Docker Build Script
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
IMAGE_NAME="${IMAGE_NAME:-picoclaw/picoclaw}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
PLATFORM="${PLATFORM:-linux/amd64}"

echo -e "${BLUE}🦞 PicoClaw Docker Builder${NC}"
echo ""
echo "Image: ${IMAGE_NAME}:${IMAGE_TAG}"
echo "Platform: ${PLATFORM}"
echo ""

cd "$PROJECT_ROOT"

# Build the Docker image
echo -e "${YELLOW}Building Docker image...${NC}"
docker build \
  --platform "$PLATFORM" \
  --tag "${IMAGE_NAME}:${IMAGE_TAG}" \
  --tag "${IMAGE_NAME}:$(date +%Y%m%d)" \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --progress=plain \
  -f Dockerfile \
  .

echo ""
echo -e "${GREEN}✅ Build complete!${NC}"
echo ""
echo "To test locally:"
echo "  docker run -it --rm -p 18790:18790 ${IMAGE_NAME}:${IMAGE_TAG} version"
echo ""
echo "To push to registry:"
echo "  docker push ${IMAGE_NAME}:${IMAGE_TAG}"
echo ""
