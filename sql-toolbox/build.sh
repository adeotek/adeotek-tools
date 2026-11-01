#!/bin/bash

# Build script for sql-toolbox

set -e

echo "Starting build process for sql-toolbox..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

# Print Go version
echo "Go version: $(go version)"

# Download dependencies
echo "Downloading dependencies..."
go mod download

# Run tests and static checks
echo "Running tests..."
go test ./...

echo "Running go vet..."
go vet ./...

# Build the application
echo "Building application..."
VERSION=$(git describe --tags --always --dirty 2>/dev/null | sed 's|.*/||' || echo dev)
mkdir -p build
go build -ldflags "-w -s -X github.com/adeotek/adeotek-tools/sql-toolbox/internal/app.Version=$VERSION" \
    -o build/sql-toolbox ./cmd/sql-toolbox

# Check if build was successful
if [ -f "build/sql-toolbox" ]; then
    echo "Build successful! Binary created at: build/sql-toolbox"
    echo "File size: $(du -h build/sql-toolbox | cut -f1)"
else
    echo "Build failed!"
    exit 1
fi

echo "Build process completed successfully!"
