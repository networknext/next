/*
   Network Next. Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"os"
	"fmt"
	"hash/fnv"
)

func main() {
	
	args := os.Args[1:]
	
	if len(args) != 1 {
		fmt.Printf("\nNetwork Next User Hash Tool\n\nUsage:\n\n    go run userhash.go [userid]\n\n")
		os.Exit(1)
	}

	userId := args[0]

	hash := fnv.New64a()
	hash.Write([]byte(userId))
	userHash := int64(hash.Sum64())

	fmt.Printf("\nuser hash: \"%s\" -> %d\n\n", userId, userHash)

}
					