#!/bin/bash

printf "\nRunning unit tests:\n\n"
time go test ./modules/... -coverprofile ./cover.out -timeout 30s
testResult=$?
if [ ! $testResult -eq 0 ]; then
    exit $testResult
fi
