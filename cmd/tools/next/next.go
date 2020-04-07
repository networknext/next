/*
   Network Next. Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tidwall/gjson"
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

func handleJSONRPCError(err error) {
	switch e := err.(type) {
	case *jsonrpc.HTTPError:
		switch e.Code {
		case http.StatusUnauthorized:
			log.Fatalf("%d: %s - use `next auth` to authorize the CLI", e.Code, http.StatusText(e.Code))
		default:
			log.Fatalf("%d: %s", e.Code, http.StatusText(e.Code))
		}
	default:
		log.Fatal(err)
	}
}

func main() {
	var env Environment

	if !env.Exists() {
		env.Write()
	}
	env.Read()

	rpcClient := jsonrpc.NewClientWithOpts("http://"+env.Hostname+"/rpc", &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", env.AuthToken),
		},
	})

	root := &ffcli.Command{
		ShortUsage: "next <subcommand>",
		Subcommands: []*ffcli.Command{
			{
				Name:       "auth",
				ShortUsage: "next auth",
				ShortHelp:  "Authorize the CLI to interact with the Portal API",
				Exec: func(_ context.Context, args []string) error {
					req, err := http.NewRequest(
						http.MethodPost,
						"https://networknext.auth0.com/oauth/token",
						strings.NewReader(`{
							"client_id":"6W6PCgPc6yj6tzO9PtW6IopmZAWmltgb",
							"client_secret":"EPZEHccNbjqh_Zwlc5cSFxvxFQHXZ990yjo6RlADjYWBz47XZMf-_JjVxcMW-XDj",
							"audience":"https://portal.networknext.com",
							"grant_type":"client_credentials"
						}`),
					)
					if err != nil {
						return err
					}

					req.Header.Add("Content-Type", "application/json")

					res, err := http.DefaultClient.Do(req)
					if err != nil {
						return err
					}
					defer res.Body.Close()

					if res.StatusCode != http.StatusOK {
						return fmt.Errorf("auth0 returned code %d", res.StatusCode)
					}

					body, err := ioutil.ReadAll(res.Body)
					if err != nil {
						return err
					}

					env.AuthToken = gjson.ParseBytes(body).Get("access_token").String()
					env.Write()

					fmt.Println(env.String())

					return nil
				},
			},

			{
				Name:       "env",
				ShortUsage: "next env <hostname>",
				ShortHelp:  "Manage environment",
				Exec: func(_ context.Context, args []string) error {
					if len(args) > 0 {
						env.Hostname = args[0]
						env.Write()
					}
					fmt.Println(env.String())
					return nil
				},
			},

			{
				Name:       "buyers",
				ShortUsage: "next buyers",
				ShortHelp:  "Manage buyers",
				Exec: func(_ context.Context, args []string) error {
					buyers(rpcClient)
					return nil
				},
			},

			{
				Name:       "datacenters",
				ShortUsage: "next datacenters <name>",
				ShortHelp:  "Manage datacenters",
				Exec: func(_ context.Context, args []string) error {
					if len(args) > 0 {
						datacenters(rpcClient, args[0])
						return nil
					}
					datacenters(rpcClient, "")
					return nil
				},
			},

			{
				Name:       "relays",
				ShortUsage: "next relays <name>",
				ShortHelp:  "Manage relays",
				Exec: func(_ context.Context, args []string) error {
					if len(args) > 0 {
						relays(rpcClient, args[0])
						return nil
					}
					relays(rpcClient, "")
					return nil
				},
			},
			{
				Name:       "ssh",
				ShortUsage: "next ssh <device identifier>",
				ShortHelp:  "SSH into a remote device, for relays the identifier is their name",
				Exec: func(ctx context.Context, args []string) error {
					if len(args) < 1 {
						log.Fatal("need a device identifer")
					}

					SSHInto(env, rpcClient, args[0])

					return nil
				},
				Subcommands: []*ffcli.Command{
					{
						Name:       "key",
						ShortUsage: "next ssh key <path to ssh key>",
						ShortHelp:  "Set the key you'd like to use for ssh-ing",
						Exec: func(ctx context.Context, args []string) error {
							if len(args) > 0 {
								env.SSHKeyFilePath = args[0]
								env.Write()
							}

							fmt.Println(env.String())

							return nil
						},
					},
				},
			},
		},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
