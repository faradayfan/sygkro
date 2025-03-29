#!/bin/bash

set -e

echo "Building"

go build -o ./bin/sygkro ./main.go
