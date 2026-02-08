#!/bin/bash

set -e

echo "Building"

# Primary build (used by ci-post-build.sh for integration tests)
go build -o ./bin/sygkro_linux_amd64 ./main.go

# Cross-compile for release (only when RELEASE_VERSION is set)
if [ -n "$RELEASE_VERSION" ]; then
  echo "Building release binaries v${RELEASE_VERSION}..."

  GOOS=linux   GOARCH=arm64 go build -o ./bin/sygkro_linux_arm64   ./main.go
  GOOS=darwin  GOARCH=amd64 go build -o ./bin/sygkro_darwin_amd64  ./main.go
  GOOS=darwin  GOARCH=arm64 go build -o ./bin/sygkro_darwin_arm64  ./main.go
  GOOS=windows GOARCH=amd64 go build -o ./bin/sygkro_windows_amd64.exe ./main.go

  echo "Built binaries:"
  ls -lh ./bin/
fi
