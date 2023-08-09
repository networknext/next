/*
   Network Next Accelerate.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"time"
)

var hostname = "http://127.0.0.1:50000"
var apiKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZGF0YWJhc2UiOnRydWUsInBvcnRhbCI6dHJ1ZX0.QFPdb-RcP8wyoaOIBYeB_X6uA7jefGPVxm2VevJvpwU"
var apiPrivateKey = "this is the private key that generates API keys. make sure you change this value in production"

// ----------------------------------------------------------------------------------------

func bash(command string) {

	cmd := exec.Command("bash", "-c", command)
	if cmd == nil {
		fmt.Printf("error: could not run bash!\n")
		os.Exit(1)
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("error: failed to run command: %v\n", err)
		os.Exit(1)
	}

	cmd.Wait()
}

func clearDatabase() {
	bash("psql -U developer -h localhost postgres -f ../schemas/sql/destroy.sql")
	bash("psql -U developer -h localhost postgres -f ../schemas/sql/create.sql")
}

func api() (*exec.Cmd, *bytes.Buffer) {

	cmd := exec.Command("./api")
	if cmd == nil {
		panic("could not create api!\n")
		return nil, nil
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ENABLE_PORTAL=false")
	cmd.Env = append(cmd.Env, "ENABLE_DATABASE=false")
	cmd.Env = append(cmd.Env, "HTTP_PORT=50000")
	cmd.Env = append(cmd.Env, fmt.Sprintf("API_PRIVATE_KEY=%s", apiPrivateKey))

	var output bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Start()

	return cmd, &output
}

// ----------------------------------------------------------------------------------------

func test_terraform() {

	fmt.Printf("\ntest_terraform\n\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// todo
}

// ----------------------------------------------------------------------------------------

type test_function func()

func main() {

	allTests := []test_function{
		test_terraform,
	}

	var tests []test_function

	if len(os.Args) > 1 {
		funcName := os.Args[1]
		for _, test := range allTests {
			name := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
			name = name[len("main."):]
			if funcName == name {
				tests = append(tests, test)
				break
			}
		}
		if len(tests) == 0 {
			panic(fmt.Sprintf("could not find any test: '%s'", funcName))
		}
	} else {
		tests = allTests // No command line args, run all tests
	}

	go func() {
		time.Sleep(time.Duration(len(tests)*120) * time.Second)
		panic("tests took too long!")
	}()

	for i := range tests {
		tests[i]()
	}

	fmt.Printf("\n")
}
