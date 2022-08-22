/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"os/exec"
	"bytes"
)

func run(command string, args ...string) bool {

	cmd := exec.Command(command, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()

    fmt.Printf("%s", stdout.String())
    fmt.Printf("%s", stderr.String())

    if err != nil {
    	fmt.Printf("error: run failed: %v\n", err)
    	return false
    }

    return true
}

/*
func visual_studio_build(compiler string, config string, platform string, solution string) {
	devenv := ""
	if compiler == "vs2017" {
		devenv = "/mnt/c/build/vs2017.sh"
	} else if compiler == "vs2019" {
		devenv = "/mnt/c/build/vs2019.sh"
	} else if compiler == "vs2022" {
		devenv = "/mnt/c/build/vs2022.sh"
	} else {
		fmt.Printf("error: unknown compiler %s\n", compiler)
		os.Exit(1)
	}
	if platform != "" {
		if !run("bash", devenv, "/Build", fmt.Sprintf("%s|%s", config, platform), solution) {
			fmt.Printf("error: failed to build solution\n")
			os.Exit(1)
		}
	} else {
		if !run("bash", devenv, "/Build", config, solution) {
			fmt.Printf("error: failed to build solution\n")
			os.Exit(1)
		}
	}
}
*/

func main() {
	fmt.Printf("hello world\n")
}
