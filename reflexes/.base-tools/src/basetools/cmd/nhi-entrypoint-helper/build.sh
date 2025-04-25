#!/bin/sh
set -e

# Build the static binary for the helper command package
# Assumes go.sum and potentially vendor/ exist from local generation
CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -o ./cmd/nhi-entrypoint-helper/nhi-entrypoint-helper ./cmd/nhi-entrypoint-helper

echo "Build complete: nhi-entrypoint-helper"