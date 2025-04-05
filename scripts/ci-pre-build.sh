#!/bin/bash

set -e

echo "Pre-Build"

go mod tidy
go fmt ./...
go test ./...
