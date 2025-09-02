#!/bin/bash

# Build script for the version CLI with unique build information
# This script injects build time, commit hash, and build ID into the binary

echo "Building version CLI with unique build information..."

# Generate build variables
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_ID="build-$(date +%s)"

# Build the application
go build -ldflags "-X main.buildTime=$BUILD_TIME -X main.commitHash=$COMMIT_HASH -X main.buildID=$BUILD_ID" -o version .

if [ $? -eq 0 ]; then
    echo "Build completed successfully!"
    echo "Build Time: $BUILD_TIME"
    echo "Commit Hash: $COMMIT_HASH"
    echo "Build ID: $BUILD_ID"
    echo ""
    echo "To see the build info, run: ./version"
else
    echo "Build failed!"
    exit 1
fi