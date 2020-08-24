#!/bin/sh

GREEN="\033[32m"
BLUE="\033[34m"
NORMAL="\033[0;39m"

CGO_ENABLED=0

printf "Building Linux 64-bit... "
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./keygen_linux ./src/keygen.go
printf "${GREEN}done${NORMAL}\n"

printf "Building FreeBSD 64-bit... "
GOOS=freebsd GOARCH=amd64 go build -ldflags="-s -w" -o ./keygen_freebsd ./src/keygen.go
printf "${GREEN}done${NORMAL}\n"

printf "Building OpenBSD 64-bit... "
GOOS=openbsd GOARCH=amd64 go build -ldflags="-s -w" -o ./keygen_openbsd ./src/keygen.go
printf "${GREEN}done${NORMAL}\n"

printf "Building ARM 64-bit... "
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ./keygen_arm ./src/keygen.go
printf "${GREEN}done${NORMAL}\n"

printf "Building MacOS 64-bit... "
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ./keygen_mac ./src/keygen.go
printf "${GREEN}done${NORMAL}\n"

printf "Building Windows 64-bit... "
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./keygen.exe ./src/keygen.go
printf "${GREEN}done${NORMAL}\n"