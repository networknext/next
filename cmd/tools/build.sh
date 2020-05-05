#!/bin/sh

printf "Building next tool... "
go build -o ./dist/next ./cmd/tools/next/*.go
printf "done\n"

printf "Building analyze tool... "
go build -o ./dist/analyze ./cmd/tools/analyze/analyze.go
printf "done\n"

printf "Building cost tool... "
go build -o ./dist/cost ./cmd/tools/cost/cost.go
printf "done\n"

printf "Building debug tool... "
go build -o ./dist/debug ./cmd/tools/debug/debug.go
printf "done\n"

printf "Building optimize tool... "
go build -o ./dist/optimize ./cmd/tools/optimize/optimize.go
printf "done\n"

printf "Building route tool... "
go build -o ./dist/route ./cmd/tools/route/route.go
printf "done\n"

printf "Building tokens tool... "
go build -o ./dist/tokens ./cmd/tools/tokens/tokens.go
printf "done\n"

printf "Building functional backend... "
go build -o ./dist/func_backend ./cmd/tools/functional/backend/*.go
printf "done\n"

printf "Building old functional backend... "
go build -o ./dist/func_backend_old ./cmd/tools/functional/backend_old/*.go
printf "done\n"

printf "Building functional tests... "
go build -o ./dist/func_tests ./cmd/tools/functional/tests/func_tests.go
printf "done\n"

cd cmd/tools/keygen && ./build.sh