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

	// need to run a magic backend in parallel, the frontend relies on it

	backend_cmd := exec.Command("./magic_backend")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	backend_cmd.Stdout = &stdout
	backend_cmd.Stderr = &stderr

	backend_cmd.Env = make([]string, 0)
	backend_cmd.Env = append(backend_cmd.Env, "ENV=local")
	backend_cmd.Env = append(backend_cmd.Env, "PORT=40000")
	backend_cmd.Env = append(backend_cmd.Env, "NEXT_DEBUG_LOGS=1")
	backend_cmd.Env = append(backend_cmd.Env, "MAGIC_UPDATE_FREQUENCY=5s")

	err := backend_cmd.Start()
	if err != nil {
		fmt.Printf("\nerror: failed to run magic backend!\n\n")
		fmt.Printf("%s", stdout.String())
		fmt.Printf("%s", stderr.String())
		os.Exit(1)
	}

	// run the magic frontend and make sure it runs and does things it's expected to do

	frontend_cmd := exec.Command("./magic_frontend")

	var frontend_stdout bytes.Buffer
	var frontend_stderr bytes.Buffer
	frontend_cmd.Stdout = &frontend_stdout
	frontend_cmd.Stderr = &frontend_stderr

	frontend_cmd.Env = make([]string, 0)
	frontend_cmd.Env = append(frontend_cmd.Env, "ENV=local")
	frontend_cmd.Env = append(frontend_cmd.Env, "PORT=50000")
	frontend_cmd.Env = append(frontend_cmd.Env, "NEXT_DEBUG_LOGS=1")

	err = frontend_cmd.Start()
	if err != nil {
		fmt.Printf("\nerror: failed to run magic frontend!\n\n")
		fmt.Printf("%s", frontend_stdout.String())
		fmt.Printf("%s", frontend_stderr.String())
		os.Exit(1)
	}

	time.Sleep(10*time.Second)

	check_output("magic_frontend", frontend_cmd, frontend_stdout, frontend_stderr)
	check_output("starting http server on port 50000", frontend_cmd, frontend_stdout, frontend_stderr)
	check_output("received new magic values", frontend_cmd, frontend_stdout, frontend_stderr)
	check_output("updated status", frontend_cmd, frontend_stdout, frontend_stderr)

	// test the health check

	response, err := http.Get("http://127.0.0.1:50000/health")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: health check failed\n")
		frontend_cmd.Process.Signal(syscall.SIGTERM)
		backend_cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the status endpoint

	_, err = http.Get("http://127.0.0.1:50000/status")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: status endpoint failed\n")
		frontend_cmd.Process.Signal(syscall.SIGTERM)
		backend_cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the version endpoint

	_, err = http.Get("http://127.0.0.1:50000/version")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: version endpoint failed\n")
		frontend_cmd.Process.Signal(syscall.SIGTERM)
		backend_cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the magic values endpoint

	response, err = http.Get("http://127.0.0.1:50000/magic")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: version endpoint failed\n")
		frontend_cmd.Process.Signal(syscall.SIGTERM)
		backend_cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	data, error := ioutil.ReadAll(response.Body)
   	if error != nil {
      	fmt.Printf("error: failed to read magic response data\n")
		frontend_cmd.Process.Signal(syscall.SIGTERM)
		backend_cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	response.Body.Close()

	if len(data) != 3*8 {
      	fmt.Printf("error: magic response should return 24 bytes (3 uint64 values)\n")
		frontend_cmd.Process.Signal(syscall.SIGTERM)
		backend_cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test that the service shuts down cleanly

	frontend_cmd.Process.Signal(os.Interrupt)

	frontend_cmd.Wait()

	check_output("received shutdown signal", frontend_cmd, frontend_stdout, frontend_stderr)
	check_output("successfully shutdown", frontend_cmd, frontend_stdout, frontend_stderr)

	// cleanly shut down the backend

	backend_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
}

// todo: test to check for upcoming -> current -> previous magic value over time

// todo: test to demonstrate multiple magic backends and migration from one to the other

// todo: what if we simplified and just made the upcoming, current and previous magic values a function of time?

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
