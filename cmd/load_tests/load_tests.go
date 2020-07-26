/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"os"
	"runtime"
	"reflect"
)

func test_session_map() {

	fmt.Printf("test_session_map\n")

	// ...
}

type test_function func()

func main() {
	allTests := []test_function{
		test_session_map,
	}

	// If there are command line arguments, use reflection to see what tests to run
	var tests []test_function
	prefix := "main."
	if len(os.Args) > 1 {
		for _, funcName := range os.Args[1:] {
			for _, test := range allTests {
				name := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
				name = name[len(prefix):]
				if funcName == name {
					tests = append(tests, test)
				}
			}
		}
	} else {
		tests = allTests // No command line args, run all tests
	}

	for i := range tests {
		tests[i]()
	}
}
