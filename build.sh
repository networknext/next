#!/usr/bin/env bash

mkdir -p dist

CFLAGS="-fPIC"

LDFLAGS="-lsodium -lcurl -lpthread -lm"

if [[ $OSTYPE == 'darwin'* ]]; then
  LDFLAGS="${LDFLAGS} -framework CoreFoundation -framework SystemConfiguration"
fi

parallel ::: \
"cd ./dist && g++ ${CFLAGS} -I../sdk4/include -shared -o sdk4.so ../sdk4/source/*.cpp ${LDFLAGS}" \
"cd ./dist && g++ ${CFLAGS} -I../sdk5/include -shared -o sdk5.so ../sdk5/source/*.cpp ${LDFLAGS}" \
"cd ./dist && g++ ${CFLAGS} -o reference_relay ../reference/relay/*.cpp ${LDFLAGS}" \
"go build -o ./dist/func_tests_backend ./cmd/func_tests_backend/*.go" \
"go build -o ./dist/magic_backend ./cmd/magic_backend/*.go" \
"go build -o ./dist/magic_frontend ./cmd/magic_frontend/*.go" \
"go build -o ./dist/relay_gateway ./cmd/relay_gateway/*.go" \
"go build -o ./dist/relay_backend ./cmd/relay_backend/*.go" \
"go build -o ./dist/relay_frontend ./cmd/relay_frontend/*.go" \
