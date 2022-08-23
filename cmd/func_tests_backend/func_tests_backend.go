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
)

func test_magic_backend() {

	fmt.Printf("test_magic_backend\n")

	cmd := exec.Command("./magic_backend")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// todo: clear the env and set it up specifically for this test

	err := cmd.Start()
	if err != nil {
		fmt.Printf("\nerror: failed to run magic backend!\n\n")
		fmt.Printf("%s", stdout.String())
		fmt.Printf("%s", stderr.String())
		os.Exit(1)
	}

	time.Sleep(10*time.Second)

	if !strings.Contains(stdout.String(), "magic_backend") {
		fmt.Printf("error: missing service name\n")
		os.Exit(1)
	}

	if !strings.Contains(stdout.String(), "updated status") {
		fmt.Printf("error: missing updated status\n")
		os.Exit(1)
	}

	if !strings.Contains(stdout.String(), "inserted instance metadata") {
		fmt.Printf("error: missing metadata insert\n")
		os.Exit(1)
	}

	cmd.Process.Signal(os.Interrupt)

	cmd.Wait()

	if !strings.Contains(stdout.String(), "received shutdown signal") ||
	   !strings.Contains(stdout.String(), "successfully shutdown") {
		fmt.Printf("error: missing clean shutdown\n")
		os.Exit(1)
	}

	/*
	// todo
	fmt.Printf("%s", stdout.String())
	fmt.Printf("%s", stderr.String())
	*/
}

func main() {
	test_magic_backend()
}
