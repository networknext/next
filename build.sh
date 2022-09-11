#!/usr/bin/env bash

mkdir -p dist

CFLAGS="-fPIC"

LDFLAGS="-lsodium -lcurl -lpthread -lm"

if [[ $OSTYPE == 'darwin'* ]]; then
  LDFLAGS="${LDFLAGS} -framework CoreFoundation -framework SystemConfiguration"
fi

MODULE="github.com/networknext/backend/modules/common"

BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
COMMIT_HASH=$(git rev-parse --short HEAD) 
COMMIT_MESSAGE=$(git log -1 --pretty=%B | tr "\n" " " | tr \' '*')

#echo ${COMMIT_MESSAGE}

parallel ::: \
"cd ./dist && g++ ${CFLAGS} -I../sdk4/include -shared -o libnext4.so ../sdk4/source/*.cpp ${LDFLAGS}" \
"cd ./dist && g++ ${CFLAGS} -I../sdk5/include -shared -o libnext5.so ../sdk5/source/*.cpp ${LDFLAGS}" \
"cd ./dist && g++ ${CFLAGS} -o reference_relay ../reference/relay/*.cpp ${LDFLAGS}" \
"go build -o ./dist/func_tests_backend ./cmd/func_tests_backend/*.go" \
"go build -ldflags \"-s -w -X ${MODULE}.buildTime=${BUILD_TIME} -X ${MODULE}.commitHash=${COMMIT_HASH} -X '${MODULE}.commitMessage=${COMMIT_MESSAGE}' \" -o ./dist/magic_backend ./cmd/magic_backend/*.go" \
"go build -ldflags \"-s -w -X ${MODULE}.buildTime=${BUILD_TIME} -X ${MODULE}.commitHash=${COMMIT_HASH} -X '${MODULE}.commitMessage=${COMMIT_MESSAGE}' \" -o ./dist/relay_gateway ./cmd/relay_gateway/*.go" \
"go build -ldflags \"-s -w -X ${MODULE}.buildTime=${BUILD_TIME} -X ${MODULE}.commitHash=${COMMIT_HASH} -X '${MODULE}.commitMessage=${COMMIT_MESSAGE}' \" -o ./dist/relay_backend ./cmd/relay_backend/*.go" \
"go build -ldflags \"-s -w -X ${MODULE}.buildTime=${BUILD_TIME} -X ${MODULE}.commitHash=${COMMIT_HASH} -X '${MODULE}.commitMessage=${COMMIT_MESSAGE}' \" -o ./dist/analytics ./cmd/analytics/*.go" \
"go build -o ./dist/server_backend4 ./cmd/server_backend4/*.go" \
"go build -o ./dist/server_backend5 ./cmd/server_backend5/*.go" \
