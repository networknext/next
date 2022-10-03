#!/bin/bash

printf "\nRunning unit tests:\n\n"
time go test ./... -coverprofile ./cover.out -timeout 10s
testResult=$?
if [ ! $testResult -eq 0 ]; then
    exit $testResult
fi
