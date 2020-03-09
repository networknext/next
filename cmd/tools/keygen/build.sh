#!/bin/bash

CGO_ENABLED=0

build() {
    printf "Building keygen $1... "
    GOOS=$2 GOARCH=$3 go build -ldflags="-s -w" -o ./dist/keygen_$2 ./keygen.go
    printf "done\n"
}

case "$OSTYPE" in
  darwin*)  build "MacOS 64-bit" darwin amd64 ;; 
  linux*)   build "Linux 64-bit" linux amd64 ;; 
  msys*)    build "Windows 64-bit" windows amd64 ;;
  *)        echo "unknown or unsupported OS type: $OSTYPE" ;;
esac