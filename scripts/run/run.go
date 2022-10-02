package main

import (
    "os"
    "os/exec"
    "fmt"

    "github.com/joho/godotenv"
)

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

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("error: could not load .env file")
		os.Exit(1)
	}

	command := args[1]

	if command == "test" {
		test()
	} else if command == "test-sdk4" || command == "test4" {
		test_sdk4()
	} else if command == "test-sdk5" || command == "test5" {
		test_sdk5()
	} else if command == "magic-backend" {
		magic_backend()
	}
}

func help() {
	fmt.Printf("todo: help\n")
}

func test() {
	bash("./scripts/test-backend.sh")
}

func test_sdk4() {
	bash("make test4 && cd ./dist && ./test4")
}

func test_sdk5() {
	bash("make test5 && cd ./dist && ./test5")
}

func magic_backend() {
	bash("make ./dist/magic_backend && ./dist/magic_backend")
}
