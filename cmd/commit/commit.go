package main

import (
	"fmt"
	"os"
	"os/exec"
)

func Bash(command string) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func main() {

	args := os.Args

	if len(args) == 2 && args[1] == "help" {
		help()
		return
	}

	message := "commit"
	if len(args) == 2 {
		message = args[1]
		fmt.Printf("\nCommitting '%s'\n\n", message)
	} else {
		fmt.Printf("\nCommitting\n\n")
	}

	Bash(fmt.Sprintf("git pull && git commit -am \"%s\" && git push origin", message))

	fmt.Printf("\n")
}

func help() {
	fmt.Printf("\nsyntax:\n\n    commit (message)\n\n")
}
