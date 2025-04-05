#!/bin/bash

set -e

echo "Publishing"

# use the gh CLI to publish the build artifact to the GitHub release
# gh release upload "$GITHUB_REF_NAME" \
#     --clobber \
#     --title "$RELEASE_VERSION" \
#     --notes "Release $RELEASE_VERSION" \
#     --target "$GIT_SHA" \
#     ./bin/sygkro_linux_amd64
