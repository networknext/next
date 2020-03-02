#!/bin/sh

printf "Running matrix analysis pipeline...\n"
go run ../cost/cost.go | go run ../optimize/optimize.go | go run ../analyze/analyze.go
printf "Complete!"
