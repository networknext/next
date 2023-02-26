/*
   Network Next. You control the network.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"reflect"
	"time"
	"net/http"
	"io/ioutil"
)

func api() (*exec.Cmd, *bytes.Buffer) {

	cmd := exec.Command("./api")
	if cmd == nil {
		panic("could not create api!\n")
		return nil, nil
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "HTTP_PORT=50000")

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Start()

	return cmd, &output
}

func test_customers() {

	fmt.Printf("test_customers\n")

	api_cmd, _ := api()

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		response, err = http.Get("http://127.0.0.1:50000/admin/customers")
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to get customers: %v", err))
	}

   body, error := ioutil.ReadAll(response.Body)
   if error != nil {
      panic(fmt.Sprintf("could not read response: %v", err))
   }

   response.Body.Close()

	fmt.Printf("--------------------------------------------------\n%s--------------------------------------------------\n", body)

	api_cmd.Process.Signal(os.Interrupt)
	api_cmd.Wait()
}

type test_function func()

func main() {

	allTests := []test_function{
		test_customers,
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

	fmt.Printf("\n")

	for i := range tests {
		tests[i]()
	}
}
