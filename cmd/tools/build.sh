#!/bin/sh

printf "Building next tool... "
go build -o ./dist/next ./cmd/tools/next/*.go
printf "done\n"

printf "Building tokens tool... "
go build -o ./dist/tokens ./cmd/tools/tokens/tokens.go
printf "done\n"

printf "Building functional backend... "
go build -o ./dist/func_backend ./cmd/tools/functional/backend/*.go
printf "done\n"

printf "Building functional tests... "
go build -o ./dist/func_tests ./cmd/tools/functional/tests/func_tests.go
printf "done\n"

cd cmd/tools/keygen && ./build.sh