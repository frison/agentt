#!/bin/sh
set -e

# Build the static binary for the manifest command package
# Assumes go.sum and potentially vendor/ exist from local generation
CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -o ./cmd/manifest/manifest ./cmd/manifest

echo "Build complete: manifest"