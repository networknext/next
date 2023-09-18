package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func bash(command string) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func main() {

	if len(os.Args) != 3 {
		os.Exit(1)
	}

	tag := os.Args[1]
	branch := os.Args[2]

	envs := []string{"dev", "staging", "prod"}

	found := false

	for i := range envs {
		if strings.Contains(tag, envs[i]) || strings.Contains(branch, envs[i]) {
			bash(fmt.Sprintf("cd portal && yarn build-%s", envs[i]))
			found = true
			break
		}
	}

	if !found {
		bash(fmt.Sprintf("cd portal && yarn build-local"))
	}
}
