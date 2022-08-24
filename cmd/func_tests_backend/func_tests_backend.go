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
	"runtime"
	"reflect"
	"io/ioutil"
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
	cmd.Env = append(cmd.Env, "HTTP_PORT=40000")
	cmd.Env = append(cmd.Env, "NEXT_DEBUG_LOGS=1")
	cmd.Env = append(cmd.Env, "MAGIC_UPDATE_SECONDS=5")

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

	// test the health check

	response, err := http.Get("http://127.0.0.1:40000/health")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: health check failed\n")
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

	// test the magic values endpoint

	response, err = http.Get("http://127.0.0.1:40000/magic")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: magic endpoint failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	magicData, error := ioutil.ReadAll(response.Body)
   	if error != nil {
      	fmt.Printf("error: failed to read magic response data\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	time.Sleep(time.Second)

	check_output("served magic values", cmd, stdout, stderr)

	// test that the magic values shuffle from upcoming -> current -> previous over time

	var upcomingMagic [8]byte
	var currentMagic [8]byte
	var previousMagic [8]byte

	copy(upcomingMagic[:], magicData[0:8])
	copy(currentMagic[:], magicData[8:16])
	copy(previousMagic[:], magicData[16:24])

	magicUpdates := 0

	for i := 0; i < 30; i++ {

		response, err = http.Get("http://127.0.0.1:40000/magic")
		if err != nil || response.StatusCode != 200 {
			fmt.Printf("error: magic endpoint failed\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		magicData, error := ioutil.ReadAll(response.Body)
	   	if error != nil {
	      	fmt.Printf("error: failed to read magic response data\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		if bytes.Compare(magicData[0:8], upcomingMagic[:]) != 0 {

			magicUpdates++

			if bytes.Compare(magicData[8:16], upcomingMagic[:]) != 0 {
		      	fmt.Printf("error: did not see upcoming magic shuffle to current magic\n")
				cmd.Process.Signal(syscall.SIGTERM)
				os.Exit(1)
			}

			if bytes.Compare(magicData[16:24], currentMagic[:]) != 0 {
		      	fmt.Printf("error: did not see current magic shuffle to previous magic\n")
				cmd.Process.Signal(syscall.SIGTERM)
				os.Exit(1)
			}

			copy(upcomingMagic[:], magicData[0:8])
			copy(currentMagic[:], magicData[8:16])
			copy(previousMagic[:], magicData[16:24])
		}

		time.Sleep(time.Second)

	}

	// we should see 5 or 6 magic updates (30 seconds with updates once every 5 seconds...)

	if magicUpdates != 5 && magicUpdates != 6 {
		fmt.Printf("error: did not see magic values update every 5 seconds")		
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
