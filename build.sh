#!/bin/bash

# Cross-platform build script for process-watch
# Builds binaries for Linux, macOS (Intel & ARM), and Windows

set -e

PROJECT_NAME="process-watch"
OUTPUT_DIR="dist"
VERSION=${1:-dev}

# Create dist directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Define build targets: GOOS/GOARCH combinations
BUILD_TARGETS=(
    "linux/amd64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

echo "Building $PROJECT_NAME v$VERSION..."
echo ""

for target in "${BUILD_TARGETS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$target"
    
    # Determine output filename
    if [ "$GOOS" = "windows" ]; then
        OUTPUT="${OUTPUT_DIR}/${PROJECT_NAME}-${GOOS}-${GOARCH}.exe"
    else
        OUTPUT="${OUTPUT_DIR}/${PROJECT_NAME}-${GOOS}-${GOARCH}"
    fi
    
    echo "Building for $GOOS/$GOARCH -> $OUTPUT"
    GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT" .
    
    # Make binary executable (for Unix-like systems)
    if [ "$GOOS" != "windows" ]; then
        chmod +x "$OUTPUT"
    fi
done

echo ""
echo "✓ Build complete! Binaries in $OUTPUT_DIR/:"
ls -lh "$OUTPUT_DIR"/$PROJECT_NAME-*
