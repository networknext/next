/*
   Network Next. Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"path"
	"errors"
	"crypto/rand"
	"golang.org/x/crypto/nacl/box"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"
	db "github.com/networknext/backend/modules/database"

	"github.com/modood/table"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tidwall/gjson"
	"github.com/ybbus/jsonrpc"
)

const (

	// todo: we must not store client secrets in our source code

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

func bash(command string) (bool, string) {
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
	fmt.Printf(msg)
	fmt.Println()
	os.Exit(level)
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

   relaysfs := flag.NewFlagSet("relays state", flag.ExitOnError)
	var relaysCount int64
	relaysfs.Int64Var(&relaysCount, "n", 0, "Number of relays to display (default: all)")

	var relaysAlphaSort bool
	relaysfs.BoolVar(&relaysAlphaSort, "alpha", false, "Sort relays by name, not by sessions carried")

	var authCommand = &ffcli.Command{
		Name:       "auth",
		ShortUsage: "next auth",
		ShortHelp:  "Authorize the operator tool",
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

				bash("rm -f database.bin && cp envs/local.bin database.bin")

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

			// todo: temporary -- copy envs/dev.bin to database.bin when we select dev
			if args[0] == "dev" {
				bash("rm -f database.bin && cp envs/local.bin database.bin")
			}

			// If we can find a matching file, "envs/<env>.env", copy it to .envs. This is loaded by the makefile to get environment vars for the env
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

			printRelays(env, relaysCount, relaysAlphaSort, regexName)

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

	var upgradeCommand = &ffcli.Command{
		Name:       "upgrade",
		ShortUsage: "next upgrade [regex...]",
		ShortHelp:  "Upgrade the specified relay(s)",
		Exec: func(_ context.Context, args []string) error {
			regexes := []string{".*"}
			if len(args) > 0 {
				regexes = args
			}

			upgradeRelays(env, regexes)

			return nil
		},
	}

	var rebootCommand = &ffcli.Command{
		Name:       "reboot",
		ShortUsage: "next reboot [regex...]",
		ShortHelp:  "Reboot the specified relay(s)",
		Exec: func(_ context.Context, args []string) error {
			regexes := []string{".*"}
			if len(args) > 0 {
				regexes = args
			}

			rebootRelays(env, regexes)

			return nil
		},
	}

	var loadCommand = &ffcli.Command{
		Name:       "load",
		ShortUsage: "next load [version] [regex...]",
		ShortHelp:  "Load the specific relay binary version onto one or more relays",
		Exec: func(_ context.Context, args []string) error {
			if len(args) < 1 {
				handleRunTimeError(fmt.Sprintf("Please provide a version"), 0)
			}
			version := args[0]
			regexes := []string{".*"}
			if len(args) > 1 {
				regexes = args[1:]
			}

			loadRelays(env, regexes, version)

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
		loadCommand,
		upgradeCommand,
		rebootCommand,
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

// -------------------------------------------------------------------------------------------------------

var cachedDatabase *db.Database

func getDatabase() *db.Database {
	if cachedDatabase != nil {
		return cachedDatabase
	}
	cachedDatabase, err := db.LoadDatabase("database.bin")
	if err != nil {
		fmt.Printf("error: could not load database.bin: %v\n", err)
		os.Exit(1)
		return nil
	}
	return cachedDatabase
}

func printDatabase() {

	// todo: if the env is not "local", pull down the database.bin from the admin service

	database := getDatabase()

	// header

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
		Name  string
		Id    string
		Live  string
		Debug string
	}

	buyers := []BuyerRow{}

	for _, v := range database.BuyerMap {

		// todo: want to reconstruct customer public key here for checking

		row := BuyerRow{
			Id:    fmt.Sprintf("%0x", v.Id),
			Name:  v.Name,
			Live:  fmt.Sprintf("%v", v.Live),
			Debug: fmt.Sprintf("%v", v.Debug),
		}

		buyers = append(buyers, row)
	}

	sort.SliceStable(buyers, func(i, j int) bool { return buyers[i].Name < buyers[j].Name })

	table.Output(buyers)

	// sellers

	fmt.Printf("\nSellers:\n\n")

	type SellerRow struct {
		Name string
		Id   string
	}

	sellers := []SellerRow{}

	for _, v := range database.SellerMap {

		row := SellerRow{
			Id:   fmt.Sprintf("%0x", v.Id),
			Name: v.Name,
		}

		sellers = append(sellers, row)
	}

	sort.SliceStable(sellers, func(i, j int) bool { return sellers[i].Id < sellers[j].Id })

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
			Id:        fmt.Sprintf("%0x", v.Id),
			Name:      v.Name,
			Latitude:  fmt.Sprintf("%+3.2f", v.Latitude),
			Longitude: fmt.Sprintf("%+3.2f", v.Longitude),
		}

		datacenters = append(datacenters, row)
	}

	sort.SliceStable(datacenters, func(i, j int) bool { return datacenters[i].Name < datacenters[j].Name })

	table.Output(datacenters)

	// relays

	fmt.Printf("\nRelays:\n\n")

	type RelayRow struct {
		Name            string
		Id              string
		PublicAddress   string
		InternalAddress string
		PublicKey       string
		PrivateKey      string
	}

	relays := []RelayRow{}

	for _, v := range database.RelayMap {

		row := RelayRow{
			Id:            fmt.Sprintf("%0x", v.Id),
			Name:          v.Name,
			PublicAddress: v.PublicAddress.String(),
			PublicKey:     base64.StdEncoding.EncodeToString(v.PublicKey),
			PrivateKey:    base64.StdEncoding.EncodeToString(v.PrivateKey),
		}

		if v.HasInternalAddress {
			row.InternalAddress = v.InternalAddress.String()
		}

		relays = append(relays, row)
	}

	sort.SliceStable(relays, func(i, j int) bool { return relays[i].Name < relays[j].Name })

	table.Output(relays)

	// route shaders

	type PropertyRow struct {
		Property string
		Value    string
	}

	for _, v := range database.BuyerMap {

		fmt.Printf("\nRoute Shader for '%s'\n\n", v.Name)

		routeShader := v.RouteShader

		properties := []PropertyRow{}

		properties = append(properties, PropertyRow{"Disable Network Next", fmt.Sprintf("%v", routeShader.DisableNetworkNext)})
		properties = append(properties, PropertyRow{"Analysis Only", fmt.Sprintf("%v", routeShader.AnalysisOnly)})
		properties = append(properties, PropertyRow{"AB Test", fmt.Sprintf("%v", routeShader.ABTest)})
		properties = append(properties, PropertyRow{"Reduce Latency", fmt.Sprintf("%v", routeShader.ReduceLatency)})
		properties = append(properties, PropertyRow{"Reduce Packet Loss", fmt.Sprintf("%v", routeShader.ReducePacketLoss)})
		properties = append(properties, PropertyRow{"Multipath", fmt.Sprintf("%v", routeShader.Multipath)})
		properties = append(properties, PropertyRow{"Force Next", fmt.Sprintf("%v", routeShader.ForceNext)})
		properties = append(properties, PropertyRow{"Selection Percent", fmt.Sprintf("%d%%", routeShader.SelectionPercent)})
		properties = append(properties, PropertyRow{"Acceptable Latency", fmt.Sprintf("%dms", routeShader.AcceptableLatency)})
		properties = append(properties, PropertyRow{"Latency Threshold", fmt.Sprintf("%dms", routeShader.LatencyThreshold)})
		properties = append(properties, PropertyRow{"Acceptable Packet Loss", fmt.Sprintf("%.1f%%", routeShader.AcceptablePacketLoss)})
		properties = append(properties, PropertyRow{"Packet Loss Sustained", fmt.Sprintf("%.1f%%", routeShader.PacketLossSustained)})
		properties = append(properties, PropertyRow{"Bandwidth Envelope Up", fmt.Sprintf("%dkbps", routeShader.BandwidthEnvelopeUpKbps)})
		properties = append(properties, PropertyRow{"Bandwidth Envelope Down", fmt.Sprintf("%dkbps", routeShader.BandwidthEnvelopeDownKbps)})
		properties = append(properties, PropertyRow{"Route Select Threshold", fmt.Sprintf("%dms", routeShader.RouteSelectThreshold)})
		properties = append(properties, PropertyRow{"Route Switch Threshold", fmt.Sprintf("%dms", routeShader.RouteSwitchThreshold)})
		properties = append(properties, PropertyRow{"Max Latency Trade Off", fmt.Sprintf("%dms", routeShader.MaxLatencyTradeOff)})
		properties = append(properties, PropertyRow{"RTT Veto (Default)", fmt.Sprintf("%dms", routeShader.RTTVeto_Default)})
		properties = append(properties, PropertyRow{"RTT Veto (Multipath)", fmt.Sprintf("%dms", routeShader.RTTVeto_Multipath)})
		properties = append(properties, PropertyRow{"RTT Veto (PacketLoss)", fmt.Sprintf("%dms", routeShader.RTTVeto_PacketLoss)})
		properties = append(properties, PropertyRow{"Max Next RTT", fmt.Sprintf("%dms", routeShader.MaxNextRTT)})
		properties = append(properties, PropertyRow{"Route Diversity", fmt.Sprintf("%d", routeShader.RouteDiversity)})

		table.Output(properties)
	}
}

type RelayFleetEntry struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Id       string `json:"hex_id"`
	Status   string `json:"status"`
	Sessions string `json:"sessions"`
	Version  string `json:"version"`
}

type RelayFleetArgs struct{}

type RelayFleetReply struct {
	RelayFleet []RelayFleetEntry `json:"relay_fleet"`
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

func printRelays(env Environment, relayCount int64, alphaSort bool, regexName string) {

	var reply RelayFleetReply = RelayFleetReply{}
	var args = RelayFleetArgs{}
	if err := makeRPCCall(env, &reply, "RelayFleetService.RelayFleet", args); err != nil {
		fmt.Printf("error: could not get relays\n")
	   return
	}

	type RelayRow struct {
		Name     string
		Address  string
		Id       string
		Status   string
		Sessions int
		Version  string
	}

	relays := []RelayRow{}

	filtered := []RelayRow{}

	for _, relay := range reply.RelayFleet {

		maxSessions, err := strconv.Atoi(relay.Sessions)
		if err != nil {
			maxSessions = -1
		}

		relays = append(relays, RelayRow{
			relay.Name,
			strings.Split(relay.Address, ":")[0],
			strings.ToUpper(relay.Id),
			relay.Status,
			maxSessions,
			relay.Version,
		})
	}

	for _, relay := range relays {
		if match, err := regexp.Match(regexName, []byte(relay.Name)); match && err == nil {
			filtered = append(filtered, relay)
			continue
		}
	}

	if alphaSort {
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].Name < filtered[j].Name
		})
	} else {
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].Sessions > filtered[j].Sessions
		})
	}

	outputRelays := filtered

	if relayCount != 0 {
		table.Output(outputRelays[0:relayCount])
	} else {
		table.Output(outputRelays)
	}
}

// ----------------------------------------------------------------

func testForSSHKey(env Environment) {
	if env.SSHKeyFilePath == "" {
		handleRunTimeError(fmt.Sprintln("The ssh key file name is not set, set it with 'next ssh key <path>'"), 0)
	}

	if _, err := os.Stat(env.SSHKeyFilePath); err != nil {
		handleRunTimeError(fmt.Sprintf("The ssh key file '%s' does not exist, set it with 'next ssh key <path>'\n", env.SSHKeyFilePath), 0)
	}
}

func SSHInto(env Environment, relayName string) {

	riot := false
	if strings.Split(relayName, ".")[0] == "riot" {
		riot = true
	}

	relays := getRelayInfo(env, relayName)
	if len(relays) == 0 {
		handleRunTimeError(fmt.Sprintf("no relays matches the regex '%s'\n", relayName), 0)
	}
	info := relays[0]
	testForSSHKey(env)
	con := NewSSHConn(info.SSHUser, info.SSHAddress, fmt.Sprintf("%d", info.SSHPort), env.SSHKeyFilePath)
	fmt.Printf("Connecting to %s\n", relayName)
	con.Connect(riot)
}

type SSHConn struct {
	user    string
	address string
	port    string
	keyfile string
}

func NewSSHConn(user, address string, port string, authKeyFilename string) SSHConn {
	return SSHConn{
		user:    user,
		address: address,
		port:    port,
		keyfile: authKeyFilename,
	}
}

func (con SSHConn) commonSSHCommands() []string {
	args := make([]string, 6)
	args[0] = "-i"
	args[1] = con.keyfile
	args[2] = "-p"
	args[3] = con.port
	args[4] = "-o"
	args[5] = "StrictHostKeyChecking=no"
	return args
}

func (con SSHConn) Connect(isRiotRelay bool) {
	args := con.commonSSHCommands()
	if isRiotRelay {
		args = append(args, "-R 9000")
	}
	args = append(args, "-tt", con.user+"@"+con.address)
	if !runCommandEnv("ssh", args, nil) {
		handleRunTimeError(fmt.Sprintln("could not start ssh session"), 1)
	}
}

func (con SSHConn) ConnectAndIssueCmd(cmd string) bool {
	args := con.commonSSHCommands()
	args = append(args, "-tt", con.user+"@"+con.address, "--", cmd)
	runCommandEnv("ssh", args, nil)
	return true
}

// ------------------------------------------------------------------------------

const (
	StartRelayScript = `sudo systemctl enable /app/relay.service && sudo systemctl start relay`
	StopRelayScript = `sudo systemctl stop relay && sudo systemctl disable relay`
	LoadRelayScript = `sudo systemctl stop relay && rm -rf relay && wget https://storage.googleapis.com/relay_artifacts/relay-%s -O relay --no-cache && chmod +x relay && ./relay version && sudo mv relay /app/relay && sudo systemctl start relay && exit`
	UpgradeRelayScript = `sudo systemctl stop relay; sudo apt update -y && sudo apt upgrade -y && sudo apt dist-upgrade -y && sudo apt autoremove -y && sudo reboot`
	RebootRelayScript = `sudo reboot`
)

type RelayInfo struct {
	Id                            uint64                `json:"id"`
	Name                          string                `json:"name"`
	SSHAddress                    string                `json:"management_addr"`
	SSHUser                       string                `json:"ssh_user"`
	SSHPort                       int64                 `json:"ssh_port"`
	State                         string                `json:"state"`
}

type RelaysArgs struct {
	Regex string `json:"name"`
}

type RelaysReply struct {
	Relays []RelayInfo `json:"relays"`
}

func getRelayInfo(env Environment, regex string) []RelayInfo {
	args := RelaysArgs{
		Regex: regex,
	}
	var reply RelaysReply
	if err := makeRPCCall(env, &reply, "OpsService.Relays", args); err != nil {
		fmt.Printf("error: could not get relay info\n")
		os.Exit(1)
	}
	return reply.Relays
}

func startRelays(env Environment, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			if strings.Contains(relay.Name, "-removed-") || relay.State != "enabled" {
				continue
			}
			fmt.Printf("starting relay %s\n", relay.Name)
			testForSSHKey(env)
			con := NewSSHConn(relay.SSHUser, relay.SSHAddress, fmt.Sprintf("%d", relay.SSHPort), env.SSHKeyFilePath)
			if !con.ConnectAndIssueCmd(StartRelayScript) {
				continue
			}
		}
	}
}

func stopRelays(env Environment, regexes []string) bool {
	success := true
	testForSSHKey(env)
	script := StopRelayScript
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			if strings.Contains(relay.Name, "-removed-") || relay.State != "enabled" {
				continue
			}
			fmt.Printf("stopping relay %s\n", relay.Name)
			con := NewSSHConn(relay.SSHUser, relay.SSHAddress, fmt.Sprintf("%d", relay.SSHPort), env.SSHKeyFilePath)
			if !con.ConnectAndIssueCmd(script) {
				success = false
				continue
			}
		}
	}

	return success
}

func upgradeRelays(env Environment, regexes []string) {
	testForSSHKey(env)
	script := UpgradeRelayScript
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			if strings.Contains(relay.Name, "-removed-") || relay.State != "enabled" {
				continue
			}			
			fmt.Printf("upgrading relay %s\n", relay.Name)
			con := NewSSHConn(relay.SSHUser, relay.SSHAddress, fmt.Sprintf("%d", relay.SSHPort), env.SSHKeyFilePath)
			con.ConnectAndIssueCmd(script)
		}
	}
}

func rebootRelays(env Environment, regexes []string) {
	testForSSHKey(env)
	script := RebootRelayScript
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			if strings.Contains(relay.Name, "-removed-") || relay.State != "enabled" {
				continue
			}			
			fmt.Printf("rebooting relay %s\n", relay.Name)
			con := NewSSHConn(relay.SSHUser, relay.SSHAddress, fmt.Sprintf("%d", relay.SSHPort), env.SSHKeyFilePath)
			con.ConnectAndIssueCmd(script)
		}
	}
}

func loadRelays(env Environment, regexes []string, version string) {
	testForSSHKey(env)
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			if strings.Contains(relay.Name, "-removed-") || relay.State != "enabled" {
				continue
			}
			fmt.Printf("loading relay-%s onto %s\n", version, relay.Name)
			testForSSHKey(env)
			con := NewSSHConn(relay.SSHUser, relay.SSHAddress, fmt.Sprintf("%d", relay.SSHPort), env.SSHKeyFilePath)
			con.ConnectAndIssueCmd(fmt.Sprintf(LoadRelayScript, version))
		}
	}
}

func relayLog(env Environment, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		for _, relay := range relays {
			con := NewSSHConn(relay.SSHUser, relay.SSHAddress, fmt.Sprintf("%d", relay.SSHPort), env.SSHKeyFilePath)
			con.ConnectAndIssueCmd("journalctl -fu relay -n 1000")
			break
		}
	}
}

func keys(env Environment, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		for _, relay := range relays {
			con := NewSSHConn(relay.SSHUser, relay.SSHAddress, fmt.Sprintf("%d", relay.SSHPort), env.SSHKeyFilePath)
			con.ConnectAndIssueCmd("sudo cat /app/relay.env | grep _KEY")
			break
		}
	}
}

// --------------------------------------------------------------------------------------------

func keygen() {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("error: could not generate relay keypair\n")
		os.Exit(1)
	}
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey[:])
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKey[:])
	fmt.Printf("export RELAY_PUBLIC_KEY=%s\n", publicKeyBase64)
	fmt.Printf("export RELAY_PRIVATE_KEY=%s\n", privateKeyBase64)
}

// --------------------------------------------------------------------------------------------

const (

	PortalHostnameLocal   = "localhost:20000"
	PortalHostnameDev     = "portal-dev.networknext.com"
	PortalHostnameStaging = "portal-staging.networknext.com"
	PortalHostnameProd    = "portal.networknext.com"

	RouterPublicKeyLocal   = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyDev     = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyStaging = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyProd    = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="

	RelayArtifactURLDev     = "https://storage.googleapis.com/development_artifacts/relay.dev.tar.gz"
	RelayArtifactURLStaging = "https://storage.googleapis.com/staging_artifacts/relay.staging.tar.gz"
	RelayArtifactURLProd    = "https://storage.googleapis.com/prod_artifacts/relay.prod.tar.gz"

	RelayBackendHostnameLocal   = "localhost"
	RelayBackendHostnameDev     = "34.117.47.154"
	RelayBackendHostnameStaging = "35.190.44.124"
	RelayBackendHostnameProd    = "35.227.196.44"

	RelayBackendURLLocal   = "http://" + RelayBackendHostnameLocal + ":30005"
	RelayBackendURLDev     = "http://" + RelayBackendHostnameDev
	RelayBackendURLStaging = "http://" + RelayBackendHostnameStaging
	RelayBackendURLProd    = "http://" + RelayBackendHostnameProd
)

type Environment struct {
	CLIRelease   string `json:"-"`
	CLIBuildTime string `json:"-"`

	RemoteRelease   string `json:"-"`
	RemoteBuildTime string `json:"-"`

	Name           string `json:"name"`
	Hostname       string `json:"hostname"`
	AuthToken      string `json:"auth_token"`
	SSHKeyFilePath string `json:"ssh_key_filepath"`
}

func (e *Environment) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Environment: %s\n", e.Name))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Hostname: %s\n", e.PortalHostname()))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("AuthToken:\n\n    %s\n\n", e.AuthToken))
	sb.WriteString(fmt.Sprintf("SSHKeyFilePath: %s\n", e.SSHKeyFilePath))

	return sb.String()
}

func (e *Environment) Exists() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	if _, err := os.Stat(envFilePath); err != nil {
		return false
	}

	return true
}

func (e *Environment) Read() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	f, err := os.Open(envFilePath)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(e); err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}
}

func (e *Environment) Write() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	f, err := os.Create(envFilePath)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(e); err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}
}

func (e *Environment) Clean() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to clean environment %v\n", err), 1)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	err = os.RemoveAll(envFilePath)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to clean environment %v\n", err), 1)

	}
}

func (e *Environment) PortalHostname() string {
	if hostname, err := e.switchEnvLocal(PortalHostnameLocal, PortalHostnameDev, PortalHostnameStaging, PortalHostnameProd); err == nil {
		return hostname
	}
	return e.Hostname
}

func (e *Environment) RouterPublicKey() (string, error) {
	return e.switchEnvLocal(RouterPublicKeyLocal, RouterPublicKeyDev, RouterPublicKeyStaging, RouterPublicKeyProd)
}

func (e *Environment) RelayBackendURL() (string, error) {
	return e.switchEnvLocal(RelayBackendURLLocal, RelayBackendURLDev, RelayBackendURLStaging, RelayBackendURLProd)
}

func (e *Environment) RelayArtifactURL() (string, error) {
	return e.switchEnv(RelayArtifactURLDev, RelayArtifactURLStaging, RelayArtifactURLProd)
}

func (e *Environment) RelayBackendHostname() (string, error) {
	return e.switchEnvLocal(RelayBackendHostnameLocal, RelayBackendHostnameDev, RelayBackendHostnameStaging, RelayBackendHostnameProd)
}

// todo: holy shit this is bad?
func (e *Environment) switchEnvLocal(ifIsLocal, ifIsDev, ifIsStaging, ifIsProd string) (string, error) {
	switch e.Name {
	case "local":
		return ifIsLocal, nil
	case "dev":
		return ifIsDev, nil
	case "staging":
		return ifIsStaging, nil
	case "prod":
		return ifIsProd, nil
	default:
		return "", errors.New("Environment does not match 'local', 'dev', 'staging', or 'prod'")
	}
}

// todo: would be nice if we didn't hard code envs, and they were defined by the set of .env files under "envs" directory...
func (e *Environment) switchEnv(ifIsDev, ifIsStaging, ifIsProd string) (string, error) {
	switch e.Name {
	case "dev":
		return ifIsDev, nil
	case "staging":
		return ifIsStaging, nil
	case "prod":
		return ifIsProd, nil
	default:
		return "", errors.New("Environment does not match 'dev', 'staging', or 'prod'")
	}
}

// -------------------------------------------------------------------------------------------

type NextCostMatrixHandlerArgs struct{}

type NextCostMatrixHandlerReply struct {
	CostMatrix []byte `json:"costMatrix"`
}

func getCostMatrix(env Environment, fileName string) {

	args := NextCostMatrixHandlerArgs{}

	var reply NextCostMatrixHandlerReply
	if err := makeRPCCall(env, &reply, "RelayFleetService.NextCostMatrixHandler", args); err != nil {
		fmt.Printf("error: could not get cost matrix\n")
		os.Exit(1)
	}

	err := ioutil.WriteFile(fileName, reply.CostMatrix, 0777)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not write %s to the filesystem: %v\n", fileName, err), 0)
	}
}

func optimizeCostMatrix(costMatrixFilename, routeMatrixFilename string, costThreshold int32) {

	costMatrixData, err := os.ReadFile(costMatrixFilename)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not read the cost matrix file: %v\n", err), 1)
	}

	var costMatrix common.CostMatrix

	err = costMatrix.Read(costMatrixData)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("error reading cost matrix: %v\n", err), 1)
	}

	numRelays := len(costMatrix.RelayIds)

	numDestRelays := 0
	for i := range costMatrix.DestRelays {
		if costMatrix.DestRelays[i] {
			numDestRelays++
		}
	}

	numCPUs := runtime.NumCPU()
	numSegments := numRelays
	if numCPUs < numRelays {
		numSegments = numRelays / 5
		if numSegments == 0 {
			numSegments = 1
		}
	}

	routeMatrix := &common.RouteMatrix{
		Version:            common.RouteMatrixVersion_Write,
		RelayIds:           costMatrix.RelayIds,
		RelayAddresses:     costMatrix.RelayAddresses,
		RelayNames:         costMatrix.RelayNames,
		RelayLatitudes:     costMatrix.RelayLatitudes,
		RelayLongitudes:    costMatrix.RelayLongitudes,
		RelayDatacenterIds: costMatrix.RelayDatacenterIds,
		DestRelays:         costMatrix.DestRelays,
		RouteEntries:       core.Optimize2(numRelays, numSegments, costMatrix.Costs, costThreshold, costMatrix.RelayDatacenterIds, costMatrix.DestRelays),
	}

	routeMatrixData, err := routeMatrix.Write(100 * 1024 * 1024)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not write route matrix: %v", err), 1)
	}

	err = os.WriteFile(routeMatrixFilename, routeMatrixData, 0644)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open the route matrix file for writing: %v\n", err), 1)
	}

	// todo: temporary -- print out route matrix as csv

	fmt.Printf(",")
	for i := range costMatrix.RelayNames {
		fmt.Printf("%s,", costMatrix.RelayNames[i])
	}
	fmt.Printf("\n")
	for i := range costMatrix.RelayNames {
		fmt.Printf("%s,", costMatrix.RelayNames[i])
		for j := range costMatrix.RelayNames {
			if i == j {
				fmt.Printf("-1,")
			} else {
				index := core.TriMatrixIndex(i, j)
				cost := costMatrix.Costs[index]
				fmt.Printf("%d,", cost)
			}
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")

	// todo: temporary -- print out dest relays

	fmt.Printf("dest relays: ")
	for i := range costMatrix.RelayNames {
		if costMatrix.DestRelays[i] {
			fmt.Printf("%s,", costMatrix.RelayNames[i])
		}
	}
	fmt.Printf("\n\n")
}

func analyzeRouteMatrix(inputFile string) {

	routeMatrixFilename := "optimize.bin"

	routeMatrixData, err := ioutil.ReadFile(routeMatrixFilename)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not read the route matrix file: %v\n", err), 1)
	}

	var routeMatrix common.RouteMatrix

	err = routeMatrix.Read(routeMatrixData)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("error reading route matrix: %v\n", err), 1)
	}

	analysis := routeMatrix.Analyze()

	fmt.Printf("RTT Improvement\n\n")

	fmt.Printf("    None: %.1f%%\n", analysis.RTTBucket_NoImprovement)
	fmt.Printf("    0-5ms: %.1f%%\n", analysis.RTTBucket_0_5ms)
	fmt.Printf("    5-10ms: %.1f%%\n", analysis.RTTBucket_5_10ms)
	fmt.Printf("    10-15ms: %.1f%%\n", analysis.RTTBucket_10_15ms)
	fmt.Printf("    15-20ms: %.1f%%\n", analysis.RTTBucket_15_20ms)
	fmt.Printf("    20-25ms: %.1f%%\n", analysis.RTTBucket_20_25ms)
	fmt.Printf("    25-30ms: %.1f%%\n", analysis.RTTBucket_25_30ms)
	fmt.Printf("    30-35ms: %.1f%%\n", analysis.RTTBucket_30_35ms)
	fmt.Printf("    35-40ms: %.1f%%\n", analysis.RTTBucket_35_40ms)
	fmt.Printf("    40-45ms: %.1f%%\n", analysis.RTTBucket_40_45ms)
	fmt.Printf("    45-50ms: %.1f%%\n", analysis.RTTBucket_45_50ms)
	fmt.Printf("    50ms+: %.1f%%\n", analysis.RTTBucket_50ms_Plus)

	fmt.Printf("\nRoute Summary:\n\n")

	fmt.Printf("    %d relays\n", len(routeMatrix.RelayIds))
	fmt.Printf("    %d total routes\n", analysis.TotalRoutes)
	fmt.Printf("    %.1f routes per-relay pair on average\n", analysis.AverageNumRoutes)
	fmt.Printf("    %.1f relays per-route on average\n", analysis.AverageRouteLength)
	fmt.Printf("    %.1f%% of relay pairs have only one route\n", analysis.OneRoutePercent)
	fmt.Printf("    %.1f%% of relay pairs have no direct route\n", analysis.NoDirectRoutePercent)
	fmt.Printf("    %.1f%% of relay pairs have no route\n", analysis.NoRoutePercent)
}

// -------------------------------------------------------------------------------------------
