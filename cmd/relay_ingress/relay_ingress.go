package main

import "log"

// These vars will hold the commit SHA and build time passed in during compilation
var (
	commitsha = ""
	buildtime = ""
)

func main() {
	log.Fatalf("Router build %s from %s not implemented", commitsha, buildtime)
}
