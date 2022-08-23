/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
	"strings"
	"syscall"
	"net/http"
	"io/ioutil"
	"runtime"
	"reflect"
)

func check_output(substring string, cmd *exec.Cmd, stdout bytes.Buffer, stderr bytes.Buffer) {
	if !strings.Contains(stdout.String(), substring) {
		fmt.Printf("\nerror: missing output '%s'\n\n", substring)
		fmt.Printf("--------------------------------------------------\n")
		fmt.Printf("%s", stdout.String())
		fmt.Printf("--------------------------------------------------\n")
		if len(stderr.String()) > 0 {
			fmt.Printf("%s", stderr.String())
			fmt.Printf("--------------------------------------------------\n")
		}
		fmt.Printf("\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}
}

func test_magic_backend() {

	fmt.Printf("test_magic_backend\n")

	// run the magic backend and make sure it runs and does things it's expected to do

	cmd := exec.Command("./magic_backend")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Env = make([]string, 0)

	cmd.Env = append(cmd.Env, "ENV=local")
	cmd.Env = append(cmd.Env, "PORT=40000")
	cmd.Env = append(cmd.Env, "NEXT_DEBUG_LOGS=1")
	cmd.Env = append(cmd.Env, "MAGIC_UPDATE_FREQUENCY=5s")

	err := cmd.Start()
	if err != nil {
		fmt.Printf("\nerror: failed to run magic backend!\n\n")
		fmt.Printf("%s", stdout.String())
		fmt.Printf("%s", stderr.String())
		os.Exit(1)
	}

	time.Sleep(10*time.Second)

	check_output("magic_backend", cmd, stdout, stderr)
	check_output("starting http server on port 40000", cmd, stdout, stderr)
	check_output("updated status", cmd, stdout, stderr)
	check_output("inserted instance metadata", cmd, stdout, stderr)
	check_output("we are the oldest instance", cmd, stdout, stderr)
	check_output("updated magic values", cmd, stdout, stderr)

	// test the health check

	response, err := http.Get("http://127.0.0.1:40000/health")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: health check failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the status endpoint

	_, err = http.Get("http://127.0.0.1:40000/status")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: status endpoint failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the version endpoint

	_, err = http.Get("http://127.0.0.1:40000/version")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: version endpoint failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test that the service shuts down cleanly

	cmd.Process.Signal(os.Interrupt)

	cmd.Wait()

	check_output("received shutdown signal", cmd, stdout, stderr)
	check_output("successfully shutdown", cmd, stdout, stderr)
}

func test_magic_frontend() {

	fmt.Printf("test_magic_frontend\n")

	// run the magic frontend and make sure it runs and does things it's expected to do

	cmd := exec.Command("./magic_frontend")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Env = make([]string, 0)

	cmd.Env = append(cmd.Env, "ENV=local")
	cmd.Env = append(cmd.Env, "PORT=50000")
	cmd.Env = append(cmd.Env, "NEXT_DEBUG_LOGS=1")

	err := cmd.Start()
	if err != nil {
		fmt.Printf("\nerror: failed to run magic frontend!\n\n")
		fmt.Printf("%s", stdout.String())
		fmt.Printf("%s", stderr.String())
		os.Exit(1)
	}

	time.Sleep(5*time.Second)

	check_output("magic_frontend", cmd, stdout, stderr)
	check_output("starting http server on port 50000", cmd, stdout, stderr)
	check_output("received new magic values", cmd, stdout, stderr)
	check_output("updated status", cmd, stdout, stderr)

	// test the health check

	response, err := http.Get("http://127.0.0.1:50000/health")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: health check failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the status endpoint

	_, err = http.Get("http://127.0.0.1:50000/status")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: status endpoint failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the version endpoint

	_, err = http.Get("http://127.0.0.1:50000/version")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: version endpoint failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the magic values endpoint

	response, err = http.Get("http://127.0.0.1:50000/magic")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: version endpoint failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	data, error := ioutil.ReadAll(response.Body)
   	if error != nil {
      	fmt.Printf("error: failed to read magic response data\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	response.Body.Close()

	if len(data) != 3*8 {
      	fmt.Printf("error: magic response should return 24 bytes (3 uint64 values)\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test that the service shuts down cleanly

	cmd.Process.Signal(os.Interrupt)

	cmd.Wait()

	check_output("received shutdown signal", cmd, stdout, stderr)
	check_output("successfully shutdown", cmd, stdout, stderr)
}

type test_function func()

func main() {
	allTests := []test_function{
		test_magic_backend,
		test_magic_frontend,
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
}
