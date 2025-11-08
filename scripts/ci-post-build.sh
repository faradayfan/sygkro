#!/bin/bash

set -e

echo "Post-Build"

echo "Testing sygkro..."

## new template and project create
cp ./bin/sygkro_linux_amd64 ../sygkro_linux_amd64
cd ..
./sygkro_linux_amd64 version

./sygkro_linux_amd64 template new test-template

./sygkro_linux_amd64 project create --template https://github.com/faradayfan/sygkro-test-template.git --git-ref v1.0.0 --quiet

cd my-project

git config --global user.email "sygkro-test-user@test.com"
git config --global user.name "sygkro-test-user"

git init
git add .
git commit -m "Initial commit"

## Diff and Sync
../sygkro_linux_amd64 project diff --git-ref v1.1.1

../sygkro_linux_amd64 project sync --git-ref v1.1.1

git diff

## Linking
cd ..
mkdir linked-project
cd linked-project
git init
echo "# Linked Project" > README.md
git add README.md
git commit -m "Initial commit for linked project"

../sygkro_linux_amd64 project link --template https://github.com/faradayfan/sygkro-test-template.git --quiet

../sygkro_linux_amd64 project diff

../sygkro_linux_amd64 project sync

git diff

echo "All tests passed!"
