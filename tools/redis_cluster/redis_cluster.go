package main

import (
	"os"
	"os/exec"
	"time"
	"fmt"
)

func runCommand(command string, args []string) bool {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	return err != nil
}

func bash(command string) bool {
	return runCommand("bash", []string{"-c", command})
}

func main() {

	fmt.Printf("\nstarting redis cluster\n\n")

	for i := 0; i < 1; i++ {
		bash(fmt.Sprintf("cd redis/%03d && redis-server ./redis.conf", 7000+i))
	}

	// bash("redis-cli --cluster create 127.0.0.1:7000 127.0.0.1:7001 127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 --cluster-replicas 1 --cluster-yes")

	for {
		time.Sleep(time.Hour)
	}
}
