#!/bin/bash

set -e

echo "Pre-Build"

go mod tidy
# go mod vendor
# go vet ./...
go fmt ./...
go test ./...
