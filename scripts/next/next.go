/*
   Network Next. Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/networknext/backend/modules/common"
	db "github.com/networknext/backend/modules/database"

	// todo: we don't want to use old modules here
	localjsonrpc "github.com/networknext/backend/modules-old/transport/jsonrpc"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tidwall/gjson"
	"github.com/ybbus/jsonrpc"
	"github.com/modood/table"
)

const (
	// Prod
	PROD_AUTH0_AUDIENCE      = "https://next-prod.networknext.com"
	PROD_AUTH0_CLIENT_ID     = "6W6PCgPc6yj6tzO9PtW6IopmZAWmltgb"
	PROD_AUTH0_CLIENT_SECRET = "EPZEHccNbjqh_Zwlc5cSFxvxFQHXZ990yjo6RlADjYWBz47XZMf-_JjVxcMW-XDj"
	PROD_AUTH0_DOMAIN        = "auth.networknext.com"
	// Dev
	DEV_AUTH0_AUDIENCE      = "https://next-dev.networknext.com"
	DEV_AUTH0_CLIENT_ID     = "qUcgJkTEztKAbJirBexzAkau4mXm6n9Q"
	DEV_AUTH0_CLIENT_SECRET = "XQEeSI3CZLeSEbboMpgla-EmLyOzPqIc1zYKB2qTWQGmvrHvWrLzd5iOXXxkzDdY"
	DEV_AUTH0_DOMAIN        = "auth-dev.networknext.com"
	// Local Development
	LOCAL_AUTH0_AUDIENCE      = "https://next-local.networknext.com"
	LOCAL_AUTH0_CLIENT_ID     = "3lxkAg0s0tiaCAeVoe2p61QSGDYJ6MsV"
	LOCAL_AUTH0_CLIENT_SECRET = "kTXtSGiH9oDBZqR4G-unfw5Bytjb8fcRoJGCuY3TEiJrdGmVEP8JO74tpNZChBzA"
	LOCAL_AUTH0_DOMAIN        = "auth-dev.networknext.com"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

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
	cmd.Stdin = os.Stdin
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
	cmd.Stdin = os.Stdin
	finalEnv := os.Environ()
	for k, v := range env {
		finalEnv = append(finalEnv, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = finalEnv

	// Make a os.Signal channel and attach any incoming os signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Start a goroutine and block waiting for any os.Signal
	go func() {
		sig := <-c

		// If the command still exists send the os.Signal captured by next tool
		// to the underlying process.
		// If the signal is interrupt, then try to directly kill the process,
		// otherwise forward the signal.
		if cmd.Process != nil {
			if sig == syscall.SIGINT {
				if err := cmd.Process.Kill(); err != nil {
					fmt.Printf("Error trying to kill a process: %v\n", err)
				}
			} else if err := cmd.Process.Signal(sig); err != nil {
				fmt.Printf("Error trying to interrupt a process: %v\n", err)
			}
			os.Exit(1)
		}
	}()

	err := cmd.Run()
	if err != nil {
		fmt.Printf("runCommand error: %v\n", err)
		return false
	}

	return true
}

// stdout is the string return value
// stderr is contained in the error return value or nil if the command exited successfully
func runCommandGetOutput(command string, args []string, env map[string]string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	finalEnv := os.Environ()
	for k, v := range env {
		finalEnv = append(finalEnv, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = finalEnv

	err := cmd.Run()

	stdoutStr := strings.Trim(stdout.String(), "\r\n")
	if err != nil {
		stderrStr := strings.Trim(stderr.String(), "\r\n")
		return stdoutStr, fmt.Errorf("%v, %s", err, stderrStr)
	}

	return stdoutStr, nil
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
		handleRunTimeError(fmt.Sprintln("error: could not find ssh"), 1)
	}
	args := make([]string, 4)
	args[0] = "ssh"
	args[1] = "-p"
	args[2] = fmt.Sprintf("%d", port)
	args[3] = fmt.Sprintf("%s@%s", user, address)
	env := os.Environ()
	err = syscall.Exec(ssh, args, env)
	if err != nil {
		handleRunTimeError(fmt.Sprintln("error: failed to exec ssh"), 1)
	}
}

func readJSONData(entity string, args []string) []byte {
	// Check if the input is piped or a filepath
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		handleRunTimeError(fmt.Sprintf("Error checking stdin stat: %v\n", err), 1)
	}
	isPipedInput := fileInfo.Mode()&os.ModeCharDevice == 0

	var data []byte
	if isPipedInput {
		// Read the piped input from stdin
		data, err = ioutil.ReadAll(bufio.NewReader(os.Stdin))
		if err != nil {
			handleRunTimeError(fmt.Sprintf("Error reading from stdin: %v\n", err), 1)
		}
	} else {
		// Read the file at the given filepath
		if len(args) == 0 {
			handleRunTimeError(fmt.Sprintf("Supply a file path to read the %s JSON or pipe it through stdin\n", entity), 0)
		}

		data, err = ioutil.ReadFile(args[0])
		if err != nil {
			// Can't read the file, assume it is raw json data
			data = []byte(args[0])
			if !json.Valid(data) {
				// It's not valid json, so error out
				handleRunTimeError("invalid input, not a valid filepath or valid JSON", 1)
			}
		}
	}

	return data
}

// level 0: user error
// level 1: program error
func handleRunTimeError(msg string, level int) {
	fmt.Println()
	fmt.Printf(msg)
	fmt.Println()
	os.Exit(level)
}

func handleJSONRPCError(env Environment, err error) {
	handleJSONRPCErrorCustom(env, err, fmt.Sprint(err))
}

func handleJSONRPCErrorCustom(env Environment, err error, msg string) {
	switch e := err.(type) {
	case *jsonrpc.HTTPError:
		switch e.Code {
		case http.StatusUnauthorized:
			handleRunTimeError(fmt.Sprintf("%d: %s - use `next auth` to authorize the operator tool\n", e.Code, http.StatusText(e.Code)), 0)
		default:
			handleRunTimeError(fmt.Sprintf("%d: %s\n", e.Code, http.StatusText(e.Code)), 0)
		}
	default:
		if env.Name != "local" && env.Name != "dev" && env.Name != "prod" {
			handleRunTimeError(fmt.Sprintf("%v - make sure the env name is set to either 'prod', 'dev', or 'local' with\nnext select <env>\n", err), 0)
		} else {
			handleRunTimeError(fmt.Sprintf("%s\n\n", msg), 1)
		}
	}
	os.Exit(1)

}

func refreshAuth(env Environment) error {
	audience := ""
	clientID := ""
	clientSecret := ""
	domain := ""

	switch env.Name {
	case "prod":
		audience = PROD_AUTH0_AUDIENCE
		clientID = PROD_AUTH0_CLIENT_ID
		clientSecret = PROD_AUTH0_CLIENT_SECRET
		domain = PROD_AUTH0_DOMAIN
	case "dev":
		audience = DEV_AUTH0_AUDIENCE
		clientID = DEV_AUTH0_CLIENT_ID
		clientSecret = DEV_AUTH0_CLIENT_SECRET
		domain = DEV_AUTH0_DOMAIN
	default:
		audience = LOCAL_AUTH0_AUDIENCE
		clientID = LOCAL_AUTH0_CLIENT_ID
		clientSecret = LOCAL_AUTH0_CLIENT_SECRET
		domain = LOCAL_AUTH0_DOMAIN
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://%s/oauth/token", domain),
		strings.NewReader(fmt.Sprintf(`{
                "client_id":"%s",
                "client_secret":"%s",
                "audience":"%s",
                "grant_type":"client_credentials"
            }`, clientID, clientSecret, audience)),
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
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	env.AuthToken = gjson.ParseBytes(body).Get("access_token").String()
	env.Write()

	fmt.Print("Successfully authorized\n")
	return nil
}

func makeRPCCall(env Environment, reply interface{}, method string, params interface{}) error {
	protocol := "https"
	if env.PortalHostname() == PortalHostnameLocal {
		protocol = "http"
	}

	rpcClient := jsonrpc.NewClientWithOpts(protocol+"://"+env.PortalHostname()+"/rpc", &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", env.AuthToken),
		},
	})

	if err := rpcClient.CallFor(&reply, method, params); err != nil {
		switch e := err.(type) {
		case *jsonrpc.HTTPError:
			switch e.Code {
			case http.StatusUnauthorized:
				// Refresh token and try again
				if err := refreshAuth(env); err != nil {
					handleRunTimeError(err.Error(), 1)
				}
				env.Read()

				rpcClient := jsonrpc.NewClientWithOpts(protocol+"://"+env.PortalHostname()+"/rpc", &jsonrpc.RPCClientOpts{
					CustomHeaders: map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", env.AuthToken),
					},
				})

				if err := rpcClient.CallFor(&reply, method, params); err != nil {
					return err
				}
			default:
				return err
			}
		default:
			return err
		}
	}
	return nil
}

// used to decode dcMap hex strings from json
type dcMapStrings struct {
	BuyerID    string `json:"buyer_id"`
	Datacenter string `json:"datacenter"`
}

func (dcm dcMapStrings) String() string {
	return fmt.Sprintf("{\n\tBuyer ID     : %s\n\tDatacenter ID: %s\n}", dcm.BuyerID, dcm.Datacenter)
}

func main() {
	var env Environment

	if !env.Exists() {
		env.Write()
	}
	env.Read()

	if env.AuthToken == "" {
		if err := refreshAuth(env); err != nil {
			handleRunTimeError(err.Error(), 1)
		}
		env.Read()
	}

	buyersfs := flag.NewFlagSet("buyers", flag.ExitOnError)
	var buyersIdSigned bool
	buyersfs.BoolVar(&buyersIdSigned, "signed", false, "Display buyer IDs as signed ints")

	buyerfs := flag.NewFlagSet("buyers", flag.ExitOnError)
	var csvOutput bool
	var signedIDs bool
	buyerfs.BoolVar(&csvOutput, "csv", false, "Send output to CSV file")
	buyerfs.BoolVar(&signedIDs, "signed", false, "Display buyer and datacenter IDs as signed ints")

	datacentersfs := flag.NewFlagSet("datacenters", flag.ExitOnError)
	var datacenterIdSigned bool
	datacentersfs.BoolVar(&datacenterIdSigned, "signed", false, "Display datacenter IDs as signed ints")

	var datacentersCSV bool
	datacentersfs.BoolVar(&datacentersCSV, "csv", false, "Send output to CSV instead of the command line")

	sessionsfs := flag.NewFlagSet("sessions", flag.ExitOnError)
	var sessionCount int64
	sessionsfs.Int64Var(&sessionCount, "n", 0, "number of top sessions to display (default: all)")
	var buyerName string
	sessionsfs.StringVar(&buyerName, "buyer", "", "specify a buyer to filter sessions on")

	relaysfs := flag.NewFlagSet("relays state", flag.ExitOnError)

	// Limit the number of relays displayed, in descending order of sessions carried
	var relaysCount int64
	relaysfs.Int64Var(&relaysCount, "n", 0, "number of relays to display (default: all)")

	var relaysAlphaSort bool
	relaysfs.BoolVar(&relaysAlphaSort, "alpha", false, "Sort relays by name, not by sessions carried")

	relaysDbFs := flag.NewFlagSet("relays state", flag.ExitOnError)

	// -list and -csv should work with all other flags
	// Show only a list or relay names
	var relaysListFlag bool
	relaysDbFs.BoolVar(&relaysListFlag, "list", false, "show list of names")

	// Return a CSV file instead of a table
	var csvOutputFlag bool
	relaysDbFs.BoolVar(&csvOutputFlag, "csv", false, "output in csv format")

	// Return all relays at this version
	var relayVersionFilter string
	relaysDbFs.StringVar(&relayVersionFilter, "version", "all", "show only relays at this version level")

	// Display relay IDs as signed ints instead of the default hex
	var relayIDSigned bool
	relaysDbFs.BoolVar(&relayIDSigned, "signed", false, "display relay IDs as signed integers")

	var authCommand = &ffcli.Command{
		Name:       "auth",
		ShortUsage: "next auth",
		ShortHelp:  "Authorize the operator tool to interact with the Portal API",
		Exec: func(_ context.Context, args []string) error {
			refreshAuth(env)
			return nil
		},
	}

	var selectCommand = &ffcli.Command{
		Name:       "select",
		ShortUsage: "next select <local|dev|prod>",
		ShortHelp:  "Select environment to use (local|dev|prod)",
		Exec: func(_ context.Context, args []string) error {
			if len(args) == 0 {
				handleRunTimeError(fmt.Sprintln("Provide an environment to switch to (local|dev|prod)"), 0)
			}

			env.Name = args[0]
			env.Write()

			if args[0] == "local" {

				bashQuiet("rm -f database.bin && cp envs/local.bin database.bin")

				// Start redis server if it isn't already
				runnable := exec.Command("ps", "aux")
				buffer, err := runnable.CombinedOutput()

				if err != nil {
					fmt.Printf("Failed to run ps aux: %v\n", err)
				}

				psAuxOutput := string(buffer)

				if !strings.Contains(psAuxOutput, "redis-server") {
					runnable := exec.Command("redis-server")
					if err := runnable.Start(); err != nil {
						fmt.Printf("Failed to start redis: %v\n", err)
					}
				}

			}

			// If we can find a matching file, "envs/<env>.env", copy it to .envs. This is loaded by the makefile to get envs!
			envFilePath := fmt.Sprintf("envs/%s.env", args[0])

			if _, err := os.Stat(envFilePath); err != nil {
				return err
			}

			rawFile, err := os.Open(envFilePath)
			if err != nil {
				return err
			}

			defer rawFile.Close()

			rootEnvFile, err := os.Create(".env")
			if err != nil {
				return err
			}

			defer rootEnvFile.Close()

			if _, err = io.Copy(rootEnvFile, rawFile); err != nil {
				return err
			}

			fmt.Printf("Selected %s environment\n", env.Name)

			return nil
		},
	}

	var envCommand = &ffcli.Command{
		Name:       "env",
		ShortUsage: "next env",
		ShortHelp:  "Display environment",
		Exec: func(_ context.Context, args []string) error {
			if len(args) > 0 {
				env.Name = args[0]
				env.Write()
				fmt.Printf("Selected %s environment\n", env.Name)
			}

			env.RemoteRelease = "Unknown"
			env.RemoteBuildTime = "Unknown"
			var reply localjsonrpc.CurrentReleaseReply = localjsonrpc.CurrentReleaseReply{}
			if err := makeRPCCall(env, &reply, "OpsService.CurrentRelease", localjsonrpc.CurrentReleaseArgs{}); err == nil {
				env.RemoteRelease = reply.Release
				env.RemoteBuildTime = reply.BuildTime
			}

			fmt.Print(env.String())
			return nil
		},
	}

	var databaseCommand = &ffcli.Command{
		Name:       "database",
		ShortUsage: "next database",
		ShortHelp:  "Print the database for the current environment",
		Exec: func(_ context.Context, args []string) error {
			printDatabase()
			return nil
		},
	}

	var hashCommand = &ffcli.Command{
		Name:       "hash",
		ShortUsage: "next hash (string)",
		ShortHelp:  "Provide the 64-bit FNV-1a hash for the provided string",
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				handleRunTimeError(fmt.Sprintf("Please provided a string"), 0)
			}

			hashValue := common.HashString(args[0])
			hexStr := fmt.Sprintf("%016x\n", hashValue)

			fmt.Printf("unsigned: %d\n", hashValue)
			fmt.Printf("signed  : %d\n", int64(hashValue))
			fmt.Printf("hex     : 0x%s\n", strings.ToUpper(hexStr))

			return nil
		},
	}

	var relaysCommand = &ffcli.Command{
		Name:       "relays",
		ShortUsage: "next relays <regex>",
		ShortHelp:  "List relays in the current environment",
		FlagSet:    relaysfs,
		Exec: func(_ context.Context, args []string) error {

			var regexName string
			if len(args) > 0 {
				regexName = args[0]
			}

			getFleetRelays(env, relaysCount, relaysAlphaSort, regexName)

			return nil
		},
	}

	var logCommand = &ffcli.Command{
		Name:       "log",
		ShortUsage: "next log <regex> [regex]",
		ShortHelp:  "View the journalctl log for a relay",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				handleRunTimeError(fmt.Sprintln("you must supply at least one argument"), 0)
			}

			relayLog(env, args)

			return nil
		},
	}

	var startCommand = &ffcli.Command{
		Name:       "start",
		ShortUsage: "next start [regex...]",
		ShortHelp:  "Start the specified relay(s)",
		Exec: func(_ context.Context, args []string) error {
			regexes := []string{".*"}
			if len(args) > 0 {
				regexes = args
			}

			startRelays(env, regexes)

			return nil
		},
	}

	var stopCommand = &ffcli.Command{
		Name:       "stop",
		ShortUsage: "next stop [regex...]",
		ShortHelp:  "Stop the specified relay(s)",
		Exec: func(_ context.Context, args []string) error {
			regexes := []string{".*"}
			if len(args) > 0 {
				regexes = args
			}

			stopRelays(env, regexes)

			return nil
		},
	}

	var keygenCommand = &ffcli.Command{
		Name:       "keygen",
		ShortUsage: "next keygen",
		ShortHelp:  "Generate a relay keypair",
		Exec: func(_ context.Context, args []string) error {
			keygen()
			return nil
		},
	}

	var keysCommand = &ffcli.Command{
		Name:       "keys",
		ShortUsage: "next keys <relay name>",
		ShortHelp:  "Print out relay keys",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				handleRunTimeError(fmt.Sprintln("You need to supply a relay name"), 0)
			}

			keys(env, args)

			return nil
		},
	}

	var sshCommand = &ffcli.Command{
		Name:       "ssh",
		ShortUsage: "next ssh <relay name>",
		ShortHelp:  "SSH into a relay",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				handleRunTimeError(fmt.Sprintln("You need to supply a relay name"), 0)
			}

			SSHInto(env, args[0])

			return nil
		},
		Subcommands: []*ffcli.Command{
			{
				Name:       "key",
				ShortUsage: "next ssh key <path to ssh key>",
				ShortHelp:  "Set the key you'd like to use for ssh",
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
	}

	var costCommand = &ffcli.Command{
		Name:       "cost",
		ShortUsage: "next cost [output_file]",
		ShortHelp:  "Get cost matrix from current environment",
		Exec: func(ctx context.Context, args []string) error {
			output := "cost.bin"
			if len(args) > 0 {
				output = args[0]
			}
			getCostMatrix(env, output)
			fmt.Printf("Cost matrix from %s saved to %s\n", env.Name, output)
			return nil
		},
	}

	var optimizeCommand = &ffcli.Command{
		Name:       "optimize",
		ShortUsage: "next optimize [rtt] [input_file] [output_file]",
		ShortHelp:  "Optimize cost matrix into a route matrix",
		Exec: func(ctx context.Context, args []string) error {
			input := "cost.bin"
			output := "optimize.bin"
			rtt := int32(1)

			if len(args) > 0 {
				if res, err := strconv.ParseInt(args[0], 10, 32); err == nil {
					rtt = int32(res)
				} else {
					handleRunTimeError(fmt.Sprintf("could not parse 1st argument to number: %v\n", err), 1)
				}
			}

			if len(args) > 1 {
				input = args[1]
			}

			if len(args) > 2 {
				output = args[2]
			}

			optimizeCostMatrix(input, output, rtt)

			fmt.Printf("Generated route matrix %s from %s\n", output, input)

			return nil
		},
	}

	var analyzeCommand = &ffcli.Command{
		Name:       "analyze",
		ShortUsage: "next analyze <input_file>",
		ShortHelp:  "Analyze route matrix from optimize",
		Exec: func(ctx context.Context, args []string) error {
			input := "optimize.bin"

			if len(args) > 0 {
				input = args[0]
			}

			analyzeRouteMatrix(input)

			return nil
		},
	}

	var signedCommand = &ffcli.Command{
		Name:       "signed",
		ShortUsage: "next signed (uint64 in hex)",
		ShortHelp:  "Provide the signed int64 representation of the provided hex uint64 value",
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				handleRunTimeError(fmt.Sprintf("Please provided an unsigned uint64 in hexadecimal format"), 0)
			}

			hexString := args[0]

			unsigned, err := strconv.ParseUint(hexString, 16, 64)
			if err != nil {
				handleRunTimeError(fmt.Sprintf("Error: %v\n", err), 1)
			}
			signed := int64(unsigned)

			fmt.Printf("Hex   : %s\nuint64: %d\nint64 : %d\n", hexString, unsigned, signed)

			return nil
		},
	}

	var unsignedCommand = &ffcli.Command{
		Name:       "unsigned",
		ShortUsage: "next unsigned (uint64)",
		ShortHelp:  "Provide the signed int64 representation of the provided hex uint64 value",
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				handleRunTimeError(fmt.Sprintf("Please provided a signed int64 (omit negative sign)"), 0)
			}

			signedString := os.Args[2]

			signed, err := strconv.ParseInt(signedString, 10, 64)
			if err != nil {
				handleRunTimeError(fmt.Sprintf("Error: %v\n", err), 1)
			}
			unsigned := uint64(signed)

			fmt.Println("Positive value:")
			fmt.Printf("\tint64 : %d\n\tHex   : %016x\n\tuint64: %d\n\n", signed, unsigned, unsigned)

			signed *= -1
			unsigned = uint64(signed)

			fmt.Println("Negative value:")
			fmt.Printf("\tint64 : %d\n\tHex   : %016x\n\tuint64: %d\n\n", signed, unsigned, unsigned)

			return nil
		},
	}

	var commands = []*ffcli.Command{
		authCommand,
		selectCommand,
		envCommand,
		databaseCommand,
		relaysCommand,
		logCommand,
		startCommand,
		stopCommand,
		keygenCommand,
		keysCommand,
		sshCommand,
		costCommand,
		optimizeCommand,
		analyzeCommand,
		hashCommand,
		signedCommand,
		unsignedCommand,
	}

	root := &ffcli.Command{
		ShortUsage:  "next <subcommand>",
		Subcommands: commands,
		Exec: func(context.Context, []string) error {
			fmt.Printf("Network Next Operator Tool\n\n")
			return nil
		},
	}

	fmt.Printf("\n")

	args := os.Args[1:]
	if len(args) == 0 || args[0] == "-h" || args[0] == "-help" || args[0] == "--h" || args[0] == "--help" {
		args = []string{}
	}

	if err := root.ParseAndRun(context.Background(), args); err != nil {
		fmt.Printf("\n")
		handleRunTimeError(fmt.Sprintf("%v\n", err), 1)
	}

	if len(args) == 0 {
		root.FlagSet.Usage()
	}

	fmt.Printf("\n")
}

func printDatabase() {

	// todo: if the env is not "local", pull down the database.bin from the admin service

	// load the database

	database, err := db.LoadDatabase("database.bin")
	if err != nil {
		fmt.Printf("error: could not load database.bin: %v\n", err)
		os.Exit(1)
	}

	// header

	fmt.Printf("Header:\n\n")

	type HeaderRow struct {
		Creator      string
		CreationTime string
	}

	header := [1]HeaderRow{}

	header[0] = HeaderRow{Creator: database.Creator, CreationTime: database.CreationTime}

	table.Output(header[:])

	// buyers

	fmt.Printf("\nBuyers:\n\n")

	type BuyerRow struct {
		Name     string
		Id       string
		Live     string
		Debug    string
	}

	buyers := []BuyerRow{}

	for _, v := range database.BuyerMap {
		
		row := BuyerRow{
			Id:    fmt.Sprintf("%0x", v.Id),
			Name:  v.Name,
			Live:  fmt.Sprintf("%v", v.Live),
			Debug: fmt.Sprintf("%v", v.Debug),
		}

		buyers = append(buyers, row)
	}

	table.Output(buyers)

	// sellers

	fmt.Printf("\nSellers:\n\n")

	type SellerRow struct {
		Name     string
		Id       string
	}

	sellers := []SellerRow{}

	for _, v := range database.SellerMap {
		
		row := SellerRow{
			Id:    fmt.Sprintf("%0x", v.Id),
			Name:  v.Name,
		}

		sellers = append(sellers, row)
	}

	table.Output(sellers)

	// datacenters

	fmt.Printf("\nDatacenters:\n\n")

	type DatacenterRow struct {
		Name      string
		Id        string
		Latitude  string
		Longitude string
	}

	datacenters := []DatacenterRow{}

	for _, v := range database.DatacenterMap {
		
		row := DatacenterRow{
			Id:    fmt.Sprintf("%0x", v.Id),
			Name:  v.Name,
			Latitude: fmt.Sprintf("%+3.2f", v.Latitude),
			Longitude: fmt.Sprintf("%+3.2f", v.Longitude),
		}

		datacenters = append(datacenters, row)
	}

	table.Output(datacenters)
}
