/*
   Network Next. Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/admin"
	"github.com/networknext/backend/modules/portal"
	db "github.com/networknext/backend/modules/database"

	"github.com/modood/table"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
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

func main() {
	var env Environment

	if !env.Exists() {
		env.Write()
	}
	env.Read()

	relaysfs := flag.NewFlagSet("relays state", flag.ExitOnError)
	var relaysCount int64
	relaysfs.Int64Var(&relaysCount, "n", 0, "Number of relays to display (default: all)")

	var relaysAlphaSort bool
	relaysfs.BoolVar(&relaysAlphaSort, "alpha", false, "Sort relays by name, not by sessions carried")

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
				bash("rm -f database.bin && cp envs/dev.bin database.bin")
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

			fmt.Printf("Selected %s environment\n\n", env.Name)

			// todo: if this is local we should definitely set up postgres with local env data here

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
		Name:       "logs",
		ShortUsage: "next logs <regex> [regex]",
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
		ShortUsage: "next ssh [regex...]",
		ShortHelp:  "SSH into a relay",
		Exec: func(ctx context.Context, args []string) error {
			ssh(env, args)
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

	var configCommand = &ffcli.Command{
		Name:       "config",
		ShortUsage: "next config [regex...]",
		ShortHelp:  "Edit the configuration of a relay",
		Exec: func(ctx context.Context, args []string) error {
			config(env, args)
			return nil
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

	var commands = []*ffcli.Command{
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
		configCommand,
		costCommand,
		optimizeCommand,
		analyzeCommand,
		hashCommand,
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
}

// -------------------------------------------------------------------------------------------------------

var cachedDatabase *db.Database

func getDatabase() *db.Database {

	// todo: get the database from the API database/binary endpoint

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
	database := getDatabase()
	fmt.Println(database.String())
}

func GetJSON(url string, object interface{}) {

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		response, err = http.Get(url)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %v", url, err))
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		panic(fmt.Sprintf("could not read response body for %s: %v", url, err))
	}

	response.Body.Close()

	err = json.Unmarshal([]byte(body), &object)
	if err != nil {
		panic(fmt.Sprintf("could not parse json response for %s: %v", url, err))
	}
}

func GetBinary(url string) []byte {

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		response, err = http.Get(url)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %v", url, err))
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		panic(fmt.Sprintf("could not read response body for %s: %v", url, err))
	}

	response.Body.Close()

	return body
}

type PortalRelaysResponse struct {
	Relays []portal.RelayData `json:"relays"`
}

type AdminRelaysResponse struct {
	Relays []admin.RelayData `json:"relays"`
	Error  string            `json:"error"`
}

// todo: make these URLs configurable per-env
var adminURL = "http://127.0.0.1:50000"
var portalURL = "http://127.0.0.1:50000"
var databaseURL = "http://127.0.0.1:50000"

func printRelays(env Environment, relayCount int64, alphaSort bool, regexName string) {

	adminRelaysResponse := AdminRelaysResponse{}
	portalRelaysResponse := PortalRelaysResponse{}

	GetJSON(fmt.Sprintf("%s/admin/relays", adminURL), &adminRelaysResponse)
	GetJSON(fmt.Sprintf("%s/portal/relays/0/%d", portalURL, constants.MaxRelays), &portalRelaysResponse)

	type RelayRow struct {
		Name     string
		PublicAddress  string
		InternalAddress  string
		Id       string
		Status   string
		Sessions int
		Version  string
	}

	relayMap := make(map[uint64]*RelayRow)

	for i := range adminRelaysResponse.Relays {
		relayId := adminRelaysResponse.Relays[i].RelayId
		relay := relayMap[relayId]
		if relay == nil {
			relay = &RelayRow{}
			relayMap[relayId] = relay
		}
		relay.Name = adminRelaysResponse.Relays[i].RelayName
		relay.Id = fmt.Sprintf("%016x", relayId)
		relay.PublicAddress = fmt.Sprintf("%s:%d", adminRelaysResponse.Relays[i].PublicIP, adminRelaysResponse.Relays[i].PublicPort)
		if adminRelaysResponse.Relays[i].InternalIP != "0.0.0.0" {
			relay.InternalAddress = fmt.Sprintf("%s:%d", adminRelaysResponse.Relays[i].InternalIP, adminRelaysResponse.Relays[i].InternalPort)
		}
		relay.Status = "offline"
		relay.Sessions = 0
		relay.Version = adminRelaysResponse.Relays[i].Version
	}

	for i := range portalRelaysResponse.Relays {
		relayId := portalRelaysResponse.Relays[i].RelayId
		relay := relayMap[relayId]
		if relay == nil {
			continue
		}
		relay.Status = "online"
		// todo: handle "shutting down" state here
		relay.Sessions = int(portalRelaysResponse.Relays[i].NumSessions)
		relay.Version = portalRelaysResponse.Relays[i].Version
	}

	relays := make([]RelayRow, len(relayMap))
	index := 0
	for _,v := range relayMap {
		relays[index] = *v
		index++
	}

	filtered := []RelayRow{}

	for _, relay := range relays {
		if match, err := regexp.Match(regexName, []byte(relay.Name)); match && err == nil {
			filtered = append(filtered, relay)
			continue
		}
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})

	if !alphaSort {
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].Sessions > filtered[j].Sessions
		})
	}

	outputRelays := filtered

	for i := range outputRelays {
		if outputRelays[i].Sessions < 0 {
			outputRelays[i].Sessions = 0
		}
	}

	if relayCount != 0 {
		table.Output(outputRelays[0:relayCount])
	} else {
		table.Output(outputRelays)
	}

	fmt.Printf("\n")
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
	StartRelayScript   = `sudo systemctl enable /app/relay.service && sudo systemctl start relay`
	StopRelayScript    = `sudo systemctl stop relay && sudo systemctl disable relay`
	LoadRelayScript    = `sudo systemctl stop relay && sudo journalctl --vacuum-size 10M && rm -rf relay && wget https://storage.googleapis.com/relay_artifacts/relay-%s -O relay --no-cache && chmod +x relay && ./relay version && sudo mv relay /app/relay && sudo systemctl start relay && exit`
	UpgradeRelayScript = `sudo journalctl --vacuum-size 10M && sudo systemctl stop relay; sudo apt update -y && sudo apt upgrade -y && sudo apt dist-upgrade -y && sudo apt autoremove -y && sudo reboot`
	RebootRelayScript  = `sudo reboot`
	ConfigRelayScript  = `sudo vi /app/relay.env && exit`
)

type RelayInfo struct {
	Id         uint64 `json:"id"`
	Name       string `json:"name"`
	SSHAddress string `json:"management_addr"`
	SSHUser    string `json:"ssh_user"`
	SSHPort    int64  `json:"ssh_port"`
	State      string `json:"state"`
}

type RelaysArgs struct {
	Regex string `json:"name"`
}

type RelaysReply struct {
	Relays []RelayInfo `json:"relays"`
}

func getRelayInfo(env Environment, regex string) []RelayInfo {

	// todo: bring back
	return nil

	/*
	args := RelaysArgs{
		Regex: regex,
	}
	var reply RelaysReply
	if err := makeRPCCall(env, &reply, "OpsService.Relays", args); err != nil {
		fmt.Printf("error: could not get relay info\n")
		os.Exit(1)
	}
	return reply.Relays
	*/
}

func ssh(env Environment, regexes []string) {

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
			testForSSHKey(env)
			riot := false
			if strings.Split(relay.Name, ".")[0] == "riot" {
				riot = true
			}
			con := NewSSHConn(relay.SSHUser, relay.SSHAddress, fmt.Sprintf("%d", relay.SSHPort), env.SSHKeyFilePath)
			fmt.Printf("Connecting to %s\n", relay.Name)
			con.Connect(riot)
			break
		}
	}
}

func config(env Environment, regexes []string) {

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
			testForSSHKey(env)
			con := NewSSHConn(relay.SSHUser, relay.SSHAddress, fmt.Sprintf("%d", relay.SSHPort), env.SSHKeyFilePath)
			fmt.Printf("Connecting to %s\n", relay.Name)
			if !con.ConnectAndIssueCmd(ConfigRelayScript) {
				continue
			}
			break
		}
	}
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
			con.ConnectAndIssueCmd(StartRelayScript)
		}
	}
}

func stopRelays(env Environment, regexes []string) {
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
			con.ConnectAndIssueCmd(script)
		}
	}
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
	Name           string `json:"name"`
	SSHKeyFilePath string `json:"ssh_key_filepath"`
}

func (e *Environment) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Environment: %s\n", e.Name))
	sb.WriteString("\n")
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

func getCostMatrix(env Environment, fileName string) {

	// todo: call in to portal interface directly to get cost matrix in binary format (grab it from redis directly...)

	/*
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
	*/
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
		RouteEntries:       core.Optimize2(numRelays, numSegments, costMatrix.Costs, costMatrix.RelayDatacenterIds, costMatrix.DestRelays),
	}

	routeMatrixData, err := routeMatrix.Write()
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
