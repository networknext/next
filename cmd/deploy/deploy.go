package main

import (
	"fmt"
	"log"
	"bytes"
	"os"
	"os/exec"
	"strings"
	"strconv"
)

func Bash(command string) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func BashQuiet(command string) string {
	var output bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	if err != nil {
		return ""
	}
	return output.String()
}

func main() {

	args := os.Args

	if len(args) < 2 || (len(args) == 2 && args[1] == "help") {
		help()
		return
	}

	env := args[1]

	if env != "dev" && env != "staging" && env != "prod" && env != "test" && env != "relay" && env != "config" {
		log.Fatalf("unknown env '%s'", env)
	}

	tag := BashQuiet(fmt.Sprintf("git tag --list --sort=-version:refname \"%s-*\" | head -n 1", env))
	if tag == "" {
		tag = fmt.Sprintf("%s-001", env)
	}

	tag = strings.TrimSpace(tag)

	values := strings.Split(tag, "-")
	if len(values) != 2 {
		log.Fatalf("invalid tag '%s'", tag)
	}

	number, err := strconv.ParseInt(values[1], 10, 0)
	if err != nil || number == 0 {
		log.Fatalf("invalid tag '%s'", tag)
	}

	number++

	tag = fmt.Sprintf("%s-%03d", env, number)

	fmt.Printf("\nDeploying %s\n\n", tag)

	Bash(fmt.Sprintf("git tag %s", tag))

	Bash(fmt.Sprintf("git push origin %s", tag))

	fmt.Printf("\n")
}

func help() {
	fmt.Printf("\nsyntax:\n\n    deploy [dev|staging|prod|relay|test|config]\n\n")
}
