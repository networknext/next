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

	envs := []string{"dev", "staging", "prod"}

	for i := range envs {
		for j := 1; j < len(os.Args); j++ {
			if strings.Contains(os.Args[j], envs[i]) {
				bash(fmt.Sprintf("cd portal && yarn build-%s", envs[i]))
				os.Exit(1)
			}

		}
	}

	bash(fmt.Sprintf("cd portal && yarn build-local"))
}
