#!/bin/bash
# Build script with BuildKit support for faster builds

# Enable BuildKit for cache mounts
export DOCKER_BUILDKIT=1

# Build with cache
docker build -t 212.64.215.155:30835/simple-vault-api:latest .

echo "Build complete!"
