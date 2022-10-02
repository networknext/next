package main

import (
    "os"
    "os/exec"
    "fmt"
)

var processes []*os.Process

func bash(command string) {

    cmd := exec.Command("bash", "-c", command)
    if cmd == nil {
        fmt.Printf("error: could not run bash!\n")
        os.Exit(1)
    }

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
	    fmt.Printf("error: failed to run command: %v", err)
	    os.Exit(1)
	}
}

func main() {
	
	args := os.Args
	
	if len(args) < 2 || (len(args) == 2 && args[1]=="help") {
		help()
		return
	}

	command := args[1]
	
	if command == "test" {
		test()
	}
}

func help() {
	fmt.Printf("todo: help\n")
}

func test() {
	fmt.Printf("test\n")
	bash("./scripts/test-backend.sh")
}
