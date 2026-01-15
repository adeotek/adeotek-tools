#!/bin/bash

# Build script for docker-net-tools Docker image

set -e

REGISTRY_URL=""
IMAGE_NAME="docker-net-tools"
TAG="latest"
PUSH=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --push)
      PUSH=true
      shift
      ;;
    *)
      TAG="$1"
      shift
      ;;
  esac
done

echo "Building ${IMAGE_NAME}:${TAG}..."
docker build -t "${REGISTRY_URL}${IMAGE_NAME}:${TAG}" .

echo "Build complete!"

if [ "$PUSH" = true ]; then
  echo "Pushing ${REGISTRY_URL}${IMAGE_NAME}:${TAG}..."
  docker push "${REGISTRY_URL}${IMAGE_NAME}:${TAG}"
  echo "Push complete!"
else
  echo "To push the image, run: ./build.sh ${TAG} --push"
fi

echo "To run the container: docker run -it --rm ${REGISTRY_URL}${IMAGE_NAME}:${TAG}"
