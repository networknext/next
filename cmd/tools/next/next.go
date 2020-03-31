/*
   Network Next. Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"io"
	"sync"
	"bufio"
	"log"
	"syscall"
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

func cleanEnvironment() {
	bashQuiet("rm env.txt")
	bashQuiet("rm env.json")
}

func selectEnvironment(hostname string) {
	fmt.Printf("selected environment: %s\n", hostname)
	// todo: select environment, eg. "prod.networknext.com". save the current environment string to "env.txt" file
}

func getEnvironment() string {
	// todo: read environment name string from "env.txt". if no env.txt file exists, return "localhost"
	return "localhost"
}

func authEnvironment(hostname string) {
	// todo: auth with environment. error if we can't auth. localhost should not perform any auth. should "just work".
	fmt.Printf("(auth with environment %s)\n", hostname)
}

func syncEnvironment(hostname string) {
	// todo: pull all the data for the environment (eg. set of relays, datacenters whatever...) down into a local json file "env.json"
	fmt.Printf("(sync with environment %s)\n", hostname)
}

type EnvironmentData struct {
	// todo
}

func loadEnvironment() *EnvironmentData {
	// todo: load environment data from "env.json" into a global var. error if we can't load it.
	fmt.Printf("(load environment)\n")
	return &EnvironmentData{}
}

func relays(env *EnvironmentData, filter string) {
	if filter != "" {
		// todo
		fmt.Printf("(print relay names with substring matching filter: %s)\n", filter)
	} else {
		// todo
		fmt.Printf("(print all relay names)\n")
	}
	
}

func datacenters(env *EnvironmentData, filter string) {
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

type SSHInfo struct {
	Address string
	User    string
	Port    int
}

func sshToRelay(env *EnvironmentData, relayName string) {
	// todo
	fmt.Printf("(ssh to relay %s)\n", relayName)
	secureShell("root", "173.255.241.176", 22)
}

func usage() {
	fmt.Printf("\nNetwork Next Operator Tool\n\n")
}

func main() {

	cmdArgs := os.Args

	if len(cmdArgs) < 2 {
		usage()
		os.Exit(0)
	}

	action := cmdArgs[1]

	args := cmdArgs[2:]

	if action == "select" && len(args) == 1 {
		hostname := args[0]
		cleanEnvironment()
		selectEnvironment(hostname)
		authEnvironment(hostname)
		syncEnvironment(hostname)
		return
	}

	hostname := getEnvironment()

	if action == "env" {
		fmt.Printf("environment is %s\n", hostname)
		return
	}

	fmt.Printf("(hostname is \"%s\")\n", hostname)

	authEnvironment(hostname)
	if action == "auth" {
		return
	}

	if action == "sync" {
		syncEnvironment(hostname)
		return
	}

	env := loadEnvironment()

	if action == "relays" {

		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}
		relays(env, filter)

    } else if action == "datacenters" {

		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}
		datacenters(env, filter)

	} else if action == "ssh" && len(args) == 1 {

		relayName := args[0]
		sshToRelay(env, relayName)

	} else if action == "clean" {

		cleanEnvironment()

	} else {

		usage()

	}
	
}
