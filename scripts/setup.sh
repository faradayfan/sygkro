#!/bin/bash
set -e

echo "Setting up project"

while read -r plugin _; do
    asdf plugin add "$plugin"
done < .tool-versions

asdf install
