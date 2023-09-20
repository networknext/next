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

	for i := 0; i < 6; i++ {
		bash(fmt.Sprintf("cd redis/%03d && redis-server ./redis.conf", 10000+i))
	}

	time.Sleep(time.Second)

	bash("redis-cli --cluster create 127.0.0.1:10000 127.0.0.1:10001 127.0.0.1:10002 127.0.0.1:10003 127.0.0.1:10004 127.0.0.1:10005 --cluster-replicas 1 --cluster-yes")

	for {
		time.Sleep(time.Hour)
	}
}
