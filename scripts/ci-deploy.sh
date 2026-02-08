#!/bin/bash

set -e

echo "Publishing"

# Upload all build artifacts to the GitHub release created by release-please
gh release upload "$TAG_NAME" \
    --clobber \
    ./bin/sygkro_linux_amd64 \
    ./bin/sygkro_linux_arm64 \
    ./bin/sygkro_darwin_amd64 \
    ./bin/sygkro_darwin_arm64 \
    ./bin/sygkro_windows_amd64.exe

echo "Uploaded binaries to release $TAG_NAME"
