#!/bin/sh

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

printf "Building backend tool... "
go build -o ./dist/backend ./cmd/tools/functional/backend/*.go
printf "done\n"

printf "Building tests tool... "
go build -o ./dist/tests ./cmd/tools/functional/tests/functional_tests.go
printf "done\n"

cd cmd/tools/keygen && ./build.sh