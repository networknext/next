#!/usr/bin/env bash
find cmd -type f -name '*.go' | sed -r 's|[^/]+$||' | sed -r 's|cmd/||' | sed -r 's|/$||' | sort | uniq | awk '{printf "dist/%s\n", $0}'
