package main

import (
	"os/exec"
	"fmt"
)

func runCommand(command string, args []string) bool {
	cmd := exec.Command(command, args...)
	err := cmd.Start()
	return err != nil
}

func bash(command string) bool {
	return runCommand("bash", []string{"-c", command})
}

func main() {

	fmt.Printf("\nstarting redis cluster\n\n")
}
