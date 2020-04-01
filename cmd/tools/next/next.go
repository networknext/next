/*
   Network Next. Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"

	"github.com/ybbus/jsonrpc"
)

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func isMac() bool {
	return runtime.GOOS == "darwin"
}

func isLinux() bool {
	return runtime.GOOS == "linux"
}

func runCommand(command string, args []string) bool {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("runCommand error: %v\n", err)
		return false
	}
	return true
}

func runCommandEnv(command string, args []string, env map[string]string) bool {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	finalEnv := os.Environ()
	for k, v := range env {
		finalEnv = append(finalEnv, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = finalEnv

	err := cmd.Run()
	if err != nil {
		fmt.Printf("runCommand error: %v\n", err)
		return false
	}

	return true
}

func runCommandQuiet(command string, args []string, stdoutOnly bool) (bool, string) {
	cmd := exec.Command(command, args...)

	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return false, ""
	}

	var stderrReader io.ReadCloser
	if !stdoutOnly {
		stderrReader, err = cmd.StderrPipe()
		if err != nil {
			return false, ""
		}
	}

	var wait sync.WaitGroup
	var mutex sync.Mutex

	output := ""

	stdoutScanner := bufio.NewScanner(stdoutReader)
	wait.Add(1)
	go func() {
		for stdoutScanner.Scan() {
			mutex.Lock()
			output += stdoutScanner.Text() + "\n"
			mutex.Unlock()
		}
		wait.Done()
	}()

	if !stdoutOnly {
		stderrScanner := bufio.NewScanner(stderrReader)
		wait.Add(1)
		go func() {
			for stderrScanner.Scan() {
				mutex.Lock()
				output += stderrScanner.Text() + "\n"
				mutex.Unlock()
			}
			wait.Done()
		}()
	} else {
		cmd.Stderr = os.Stderr
	}

	err = cmd.Start()
	if err != nil {
		return false, output
	}

	wait.Wait()

	err = cmd.Wait()
	if err != nil {
		return false, output
	}

	return true, output
}

func runCommandInteractive(command string, args []string) bool {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func bash(command string) bool {
	return runCommand("bash", []string{"-c", command})
}

func bashQuiet(command string) (bool, string) {
	return runCommandQuiet("bash", []string{"-c", command}, false)
}

func datacenters(env Environment, filter string) {
	if filter != "" {
		// todo
		fmt.Printf("(print datacenter names with substring matching filter: %s)\n", filter)
	} else {
		// todo
		fmt.Printf("(print all datacenter names)\n")
	}
}

func secureShell(user string, address string, port int) {
	ssh, err := exec.LookPath("ssh")
	if err != nil {
		log.Fatalf("error: could not find ssh")
	}
	args := make([]string, 4)
	args[0] = "ssh"
	args[1] = "-p"
	args[2] = fmt.Sprintf("%d", port)
	args[3] = fmt.Sprintf("%s@%s", user, address)
	env := os.Environ()
	err = syscall.Exec(ssh, args, env)
	if err != nil {
		log.Fatalf("error: failed to exec ssh")
	}
}

func sshToRelay(env Environment, relayName string) {
	fmt.Printf("(ssh to relay %s)\n", relayName)
	// todo: look up relay by name, get ssh data from relay entry.
	user := "root"
	address := "173.255.241.176"
	port := 22
	secureShell(user, address, port)
}

func usage() {
	fmt.Printf("\nNetwork Next Operator Tool\n\n")
}

func main() {
	var env Environment
	// var err error

	if !env.Exists() {
		env.Write()
	}
	env.Read()

	cmdArgs := os.Args

	if len(cmdArgs) < 2 {
		usage()
		os.Exit(0)
	}

	action := cmdArgs[1]

	args := cmdArgs[2:]

	switch action {
	case "env":
		if len(args) == 1 {
			switch args[0] {
			case "clean":
				env.Clean()

			default:
				env.Hostname = args[0]
				env.Write()
			}
		}
		fmt.Println(env.String())
		return
	case "auth":
		// hostname := getEnvironment()
		// authEnvironment(hostname)
	}

	rpcClient := jsonrpc.NewClient("http://" + env.Hostname + "/rpc")

	switch action {
	case "relays":
		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}
		relays(rpcClient, filter)
	case "datacenters":
		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}
		datacenters(env, filter)
	case "ssh":
		relayName := args[0]
		sshToRelay(env, relayName)
	default:
		usage()
	}
}
