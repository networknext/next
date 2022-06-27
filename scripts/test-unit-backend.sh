#!/bin/bash

printf "\nRunning unit tests:\n\n"
go test ./... -coverprofile ./cover.out -timeout 30s
testResult=$?
if [ ! $testResult -eq 0 ]; then
    exit $testResult
fi

printf "\n"
