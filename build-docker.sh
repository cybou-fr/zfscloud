#!/bin/bash
# Build script for Docker deployment on Linux/macOS

set -e

echo "Building ZFSCloud Server for Linux (Docker)..."

# Set environment for Linux build
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=1

# Build server
echo "Building server binary..."
go build -o bin/zfscloud-server ./server/cmd

if [ $? -eq 0 ]; then
    echo "✓ Server built successfully: bin/zfscloud-server"
    
    # Show file info
    if [ -f "bin/zfscloud-server" ]; then
        ls -lh bin/zfscloud-server
    fi
else
    echo "✗ Server build failed!"
    exit 1
fi

echo ""
echo "Now you can run:"
echo "  cd docker"
echo "  docker-compose up -d"
