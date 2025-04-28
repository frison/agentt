#!/usr/bin/env sh

set -e # Exit immediately if a command exits with a non-zero status.

# Build the Go application
echo "Building discover-reflexes..."
# Build specifying the package path relative to the build context (/build)
# Output the binary directly into its own directory
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./cmd/discover-reflexes/discover-reflexes ./cmd/discover-reflexes
echo "Build complete: discover-reflexes"