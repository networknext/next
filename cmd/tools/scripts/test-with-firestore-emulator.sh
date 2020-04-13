#!/bin/bash

printf "Starting firestore emulator...\n\n"
setsid gcloud beta emulators firestore start --host-port $FIRESTORE_EMULATOR_HOST &
sessionID=$! # Get the session ID of this process so we can close it later

# Trap kill the process group so all firestore emulator processes are closed properly
trap "kill -- -$sessionID" EXIT
sleep 3

printf "Running go tests:\n\n"
go test  ./... -coverprofile ./cover.out
testResult=$?
if [ ! $testResult -eq 0 ]; then
    exit $testResult
fi

printf "\n\nCoverage results of go tests:\n\n"
go tool cover -func ./cover.out
printf "\n"
