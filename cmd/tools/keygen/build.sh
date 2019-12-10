#!/bin/sh

CGO_ENABLED=0

printf "Building keygen Linux 64-bit... "
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./dist/keygen_linux ./keygen.go
printf "done\n"

printf "Building keygen FreeBSD 64-bit... "
GOOS=freebsd GOARCH=amd64 go build -ldflags="-s -w" -o ./dist/keygen_freebsd ./keygen.go
printf "done\n"

printf "Building keygen OpenBSD 64-bit... "
GOOS=openbsd GOARCH=amd64 go build -ldflags="-s -w" -o ./dist/keygen_openbsd ./keygen.go
printf "done\n"

printf "Building keygen ARM 64-bit... "
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ./dist/keygen_arm ./keygen.go
printf "done\n"

printf "Building keygen MacOS 64-bit... "
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ./dist/keygen_mac ./keygen.go
printf "done\n"

printf "Building keygen Windows 64-bit... "
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./dist/keygen.exe ./keygen.go
printf "done\n"