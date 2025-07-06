#!/bin/bash

# Build script for sql-migration

set -e

echo "Starting build process for sql-migration..."

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

# Tidy up modules
echo "Tidying up modules..."
go mod tidy

# Run tests
echo "Running tests..."
go test -v ./...

# Build the application
echo "Building application..."
mkdir -p build
go build -ldflags "-w -s" -o build/sql-migration ./cmd/sql-migration

# Check if build was successful
if [ -f "build/sql-migration" ]; then
    echo "Build successful! Binary created at: build/sql-migration"
    echo "File size: $(du -h build/sql-migration | cut -f1)"
else
    echo "Build failed!"
    exit 1
fi

echo "Build process completed successfully!"
