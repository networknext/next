#!/bin/bash

printf "\nRunning unit tests:\n\n"
time go test ./modules/... ./modules-old/... -coverprofile ./cover.out -timeout 10s
testResult=$?
if [ ! $testResult -eq 0 ]; then
    exit $testResult
fi
