#!/usr/bin/env bash
cd cmd && find . -type f -name '*.go' | sed -r 's|/[^/]+$||' | sort | uniq | awk '/a/ {printf "dist/%s\n", $0}'
