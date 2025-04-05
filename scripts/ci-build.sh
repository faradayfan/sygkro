#!/bin/bash

set -e

echo "Building"

go build -o ./bin/sygkro_linux_amd64 ./main.go
