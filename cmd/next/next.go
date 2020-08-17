/*
   Network Next. Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tidwall/gjson"
	"github.com/ybbus/jsonrpc"
)

var (
	release   string
	buildtime string
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
			handleRunTimeError(fmt.Sprintf("Supply a file path to read the %s JSON or pipe it through stdin\nnext %s add [filepath]\nor\ncat <filepath> | next %s add\n\nFor an example JSON schema:\nnext %s add example\n", entity, entity, entity, entity), 0)
		}

		data, err = ioutil.ReadFile(args[0])
		if err != nil {
			handleRunTimeError(fmt.Sprintf("Error reading %s JSON file: %v\n", entity, err), 1)
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

type buyer struct {
	Name      string
	Domain    string
	Active    bool
	Live      bool
	PublicKey string
}

type seller struct {
	Name                 string
	IngressPriceNibblins routing.Nibblin
	EgressPriceNibblins  routing.Nibblin
}

type relay struct {
	Name                string
	Addr                string
	PublicKey           string
	SellerID            string
	DatacenterName      string
	NicSpeedMbps        uint64
	IncludedBandwidthGB uint64
	ManagementAddr      string
	SSHUser             string
	SSHPort             int64
}

type datacenter struct {
	Name     string
	Enabled  bool
	Location routing.Location
}

// used to decode dcMap hex strings from json
type dcMapStrings struct {
	BuyerID    string `json:"buyer_id"`
	Datacenter string `json:"datacenter"`
	Alias      string `json:"alias"`
}

func (dcm dcMapStrings) String() string {
	return fmt.Sprintf("{\n\tBuyer ID     : %s\n\tDatacenter ID: %s\n\tAlias        : %s\n}", dcm.BuyerID, dcm.Datacenter, dcm.Alias)
}

func main() {
	var env Environment

	if !env.Exists() {
		env.Write()
	}
	env.Read()

	protocol := "https"
	if env.PortalHostname() == PortalHostnameLocal {
		protocol = "http"
	}

	rpcClient := jsonrpc.NewClientWithOpts(protocol+"://"+env.PortalHostname()+"/rpc", &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", env.AuthToken),
		},
	})

	var srcRelays arrayFlags
	var destRelays arrayFlags
	var routeRTT float64
	var routeHash uint64
	routesfs := flag.NewFlagSet("routes", flag.ExitOnError)
	routesfs.Var(&srcRelays, "src", "source relay names")
	routesfs.Var(&destRelays, "dest", "destination relay names")
	routesfs.Float64Var(&routeRTT, "rtt", 5, "route RTT required for selection")
	routesfs.Uint64Var(&routeHash, "hash", 0, "a previous hash to use")

	buyersfs := flag.NewFlagSet("buyers", flag.ExitOnError)
	var buyersIdSigned bool
	buyersfs.BoolVar(&buyersIdSigned, "signed", false, "Display buyer IDs as signed ints")

	datacentersfs := flag.NewFlagSet("datacenters", flag.ExitOnError)
	var datacenterIdSigned bool
	datacentersfs.BoolVar(&datacenterIdSigned, "signed", false, "Display datacenter IDs as signed ints")

	sessionsfs := flag.NewFlagSet("sessions", flag.ExitOnError)
	var sessionCount int64
	sessionsfs.Int64Var(&sessionCount, "n", 0, "number of top sessions to display (default: all)")
	var buyerName string
	sessionsfs.StringVar(&buyerName, "buyer", "", "specify a buyer to filter sessions on")

	relaylogfs := flag.NewFlagSet("relay logs", flag.ExitOnError)

	var loglines uint
	relaylogfs.UintVar(&loglines, "n", 10, "the number of log lines to display")

	relaydisablefs := flag.NewFlagSet("relay disable", flag.ExitOnError)

	var hardDisable bool
	relaydisablefs.BoolVar(&hardDisable, "hard", false, "hard disable the relay(s), killing the process immediately")

	relayupdatefs := flag.NewFlagSet("relay update", flag.ExitOnError)

	var updateOpts updateOptions
	relayupdatefs.Uint64Var(&updateOpts.coreCount, "cores", 0, "number of cores for the relay to utilize")
	relayupdatefs.BoolVar(&updateOpts.force, "force", false, "force the relay update regardless of the version")
	relayupdatefs.BoolVar(&updateOpts.hard, "hard", false, "hard update the relay(s), killing the process immediately")

	relaysfs := flag.NewFlagSet("relays state", flag.ExitOnError)

	// Flags to only show relays in certain states
	var relaysStateShowFlags [6]bool
	relaysfs.BoolVar(&relaysStateShowFlags[routing.RelayStateEnabled], "enabled", false, "only show enabled relays")
	relaysfs.BoolVar(&relaysStateShowFlags[routing.RelayStateMaintenance], "maintenance", false, "only show relays in maintenance")
	relaysfs.BoolVar(&relaysStateShowFlags[routing.RelayStateDisabled], "disabled", false, "only show disabled relays")
	relaysfs.BoolVar(&relaysStateShowFlags[routing.RelayStateQuarantine], "quarantined", false, "only show quarantined relays")
	relaysfs.BoolVar(&relaysStateShowFlags[routing.RelayStateDecommissioned], "decommissioned", false, "only show decommissioned relays")
	relaysfs.BoolVar(&relaysStateShowFlags[routing.RelayStateOffline], "offline", false, "only show offline relays")

	// Flags to hide relays in certain states
	var relaysStateHideFlags [6]bool
	relaysfs.BoolVar(&relaysStateHideFlags[routing.RelayStateEnabled], "noenabled", false, "hide enabled relays")
	relaysfs.BoolVar(&relaysStateHideFlags[routing.RelayStateMaintenance], "nomaintenance", false, "hide relays in maintenance")
	relaysfs.BoolVar(&relaysStateHideFlags[routing.RelayStateDisabled], "nodisabled", false, "hide disabled relays")
	relaysfs.BoolVar(&relaysStateHideFlags[routing.RelayStateQuarantine], "noquarantined", false, "hide quarantined relays")
	relaysfs.BoolVar(&relaysStateHideFlags[routing.RelayStateDecommissioned], "nodecommissioned", false, "hide decommissioned relays")
	relaysfs.BoolVar(&relaysStateHideFlags[routing.RelayStateOffline], "nooffline", false, "hide offline relays")

	// Flag to see relays that are down (haven't pinged backend in 30 seconds)
	var relaysDownFlag bool
	relaysfs.BoolVar(&relaysDownFlag, "down", false, "show relays that are down")

	// Show all relays (including decommissioned ones) regardless of other flags
	var relaysAllFlag bool
	relaysfs.BoolVar(&relaysAllFlag, "all", false, "show all relays")

	// -list and -csv should work with all other flags
	// Show only a list or relay names
	var relaysListFlag bool
	relaysfs.BoolVar(&relaysListFlag, "list", false, "show list of names")

	// Return a CSV file instead of a table
	var csvOutputFlag bool
	relaysfs.BoolVar(&csvOutputFlag, "csv", false, "return a CSV file")

	// Return a CSV file instead of a table
	var relayVersionFilter string
	relaysfs.StringVar(&relayVersionFilter, "version", "all", "show only relays at this version level")

	// Limit the number of relays displayed, in descending order of sessions carried
	var relaysCount int64
	relaysfs.Int64Var(&relaysCount, "n", 0, "number of relays to display (default: all)")

	// Display relay IDs as signed ints instead of the default hex
	var relayIDSigned bool
	relaysfs.BoolVar(&relayIDSigned, "signed", false, "display relay IDs as signed integers")

	// display the OPS version of the relay output
	var relayOpsOutput bool
	relaysfs.BoolVar(&relayOpsOutput, "ops", false, "display ops metadata (costs, bandwidth, terms, etc)")

	// Sort -ops output by IncludedBandwidthGB, descending
	var relayBWSort bool
	relaysfs.BoolVar(&relayBWSort, "bw", false, "Sort -ops output by IncludedBandwidthGB, descending (ignored w/o -ops)")

	var authCommand = &ffcli.Command{
		Name:       "auth",
		ShortUsage: "next auth",
		ShortHelp:  "Authorize the operator tool to interact with the Portal API",
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
			defer res.Body.Close()

			env.AuthToken = gjson.ParseBytes(body).Get("access_token").String()
			env.Write()

			fmt.Print("Successfully authorized\n")

			return nil
		},
	}

	var selectCommand = &ffcli.Command{
		Name:       "select",
		ShortUsage: "next select <local|dev|staging|prod>",
		ShortHelp:  "Select environment to use (local|dev|staging|prod)",
		Exec: func(_ context.Context, args []string) error {
			if len(args) == 0 {
				handleRunTimeError(fmt.Sprintln("Provide an environment to switch to (local|dev|prod)"), 0)
			}

			if args[0] != "local" && args[0] != "dev" && args[0] != "staging" && args[0] != "prod" {
				handleRunTimeError(fmt.Sprintf("Invalid environment %s: use (local|dev|prod)\n", args[0]), 0)
			}

			env.Name = args[0]
			env.Write()

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
				if args[0] != "local" && args[0] != "dev" && args[0] != "staging" && args[0] != "prod" {
					handleRunTimeError(fmt.Sprintf("Invalid environment %s: use (local|dev|prod)\n", args[0]), 0)
				}

				env.Name = args[0]
				env.Write()

				fmt.Printf("Selected %s environment\n", env.Name)
			}

			env.RemoteRelease = "Unknown"
			env.RemoteBuildTime = "Unknown"
			var reply localjsonrpc.CurrentReleaseReply
			if err := rpcClient.CallFor(&reply, "OpsService.CurrentRelease", localjsonrpc.CurrentReleaseArgs{}); err == nil {
				env.RemoteRelease = reply.Release
				env.RemoteBuildTime = reply.BuildTime
			}

			env.CLIRelease = release
			env.CLIBuildTime = buildtime
			fmt.Print(env.String())
			return nil
		},
	}

	var sessionsCommand = &ffcli.Command{
		Name:       "sessions",
		ShortUsage: "next sessions",
		ShortHelp:  "List sessions",
		FlagSet:    sessionsfs,
		Exec: func(_ context.Context, args []string) error {
			if len(args) > 0 {
				sessions(rpcClient, env, args[0], sessionCount)
				return nil
			}
			sessionsByBuyer(rpcClient, env, buyerName, sessionCount)
			return nil
		},
		Subcommands: []*ffcli.Command{
			{
				Name:       "flush",
				ShortUsage: "next sessions flush",
				ShortHelp:  "Flush all sessions in Redis in the Portal",
				Exec: func(ctx context.Context, args []string) error {
					flushsessions(rpcClient, env)
					fmt.Println("All sessions flushed.")
					return nil
				},
			},
		},
	}

	var sessionCommand = &ffcli.Command{
		Name:       "session",
		ShortUsage: "next session <session id>",
		ShortHelp:  "Display details for the specified session",
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				fmt.Printf("A session ID must be provided (see ./next sessions).")
			}
			sessions(rpcClient, env, args[0], sessionCount)
			return nil
		},
	}
	var relaysCommand = &ffcli.Command{
		Name:       "relays",
		ShortUsage: "next relays <regex>",
		ShortHelp:  "List relays",
		FlagSet:    relaysfs,
		Exec: func(_ context.Context, args []string) error {
			if relaysfs.NFlag() == 0 {
				// If no flags are given, set the default set of flags
				relaysStateShowFlags[routing.RelayStateEnabled] = true
				relaysStateShowFlags[routing.RelayStateQuarantine] = true
				relaysStateHideFlags[routing.RelayStateDecommissioned] = true
			}

			if relaysAllFlag {
				// Show all relays (except for decommissioned relays) with --all flag
				relaysStateShowFlags[routing.RelayStateEnabled] = true
				relaysStateShowFlags[routing.RelayStateMaintenance] = true
				relaysStateShowFlags[routing.RelayStateDisabled] = true
				relaysStateShowFlags[routing.RelayStateQuarantine] = true
				relaysStateShowFlags[routing.RelayStateOffline] = true
			}

			if relaysStateShowFlags[routing.RelayStateDecommissioned] {
				//  Show decommissioned relays with --decommissioned flag by essentially disabling --nodecommissioned flag
				relaysStateHideFlags[routing.RelayStateDecommissioned] = false
			}

			var arg string
			if len(args) > 0 {
				arg = args[0]
			}

			if relayOpsOutput {
				opsRelays(
					rpcClient,
					env,
					arg,
					relaysStateShowFlags,
					relaysStateHideFlags,
					relaysDownFlag,
					csvOutputFlag,
					relayVersionFilter,
					relaysCount,
					relayIDSigned,
					relayBWSort,
				)
			} else {
				relays(
					rpcClient,
					env,
					arg,
					relaysStateShowFlags,
					relaysStateHideFlags,
					relaysDownFlag,
					relaysListFlag,
					csvOutputFlag,
					relayVersionFilter,
					relaysCount,
					relayIDSigned,
				)
			}

			return nil
		},
		Subcommands: []*ffcli.Command{
			{
				Name:       "count",
				ShortUsage: "next relays count <regex>",
				ShortHelp:  "Return the number of relays in each state",
				Exec: func(ctx context.Context, args []string) error {
					if len(args) > 0 {
						countRelays(rpcClient, env, args[0])
						return nil
					}

					countRelays(rpcClient, env, "")
					return nil
				},
			},
		},
	}

	var relayCommand = &ffcli.Command{
		Name:       "relay",
		ShortUsage: "next relay <subcommand>",
		ShortHelp:  "Manage relays",
		Exec: func(_ context.Context, args []string) error {

			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			{
				Name:       "logs",
				ShortUsage: "next relay logs <regex> [regex]",
				ShortHelp:  "Print the last n journalctl lines for each matching relay, if the n flag is unset it defaults to 10",
				FlagSet:    relaylogfs,
				Exec: func(ctx context.Context, args []string) error {
					if len(args) == 0 {
						handleRunTimeError(fmt.Sprintln("you must supply at least one argument"), 0)
					}

					relayLogs(rpcClient, env, loglines, args)

					return nil
				},
			},
			{
				Name:       "check",
				ShortUsage: "next relay check [regex]",
				ShortHelp:  "List all or a subset of relays and see diagnostic information. Refer to the README for more information",
				FlagSet:    relaysfs,
				Exec: func(ctx context.Context, args []string) error {
					if relaysfs.NFlag() == 0 {
						// If no flags are given, set the default set of flags
						relaysStateHideFlags[routing.RelayStateDecommissioned] = true
					}

					if relaysAllFlag {
						// Show all relays (except for decommissioned relays) with --all flag
						relaysStateShowFlags[routing.RelayStateEnabled] = true
						relaysStateShowFlags[routing.RelayStateMaintenance] = true
						relaysStateShowFlags[routing.RelayStateDisabled] = true
						relaysStateShowFlags[routing.RelayStateQuarantine] = true
						relaysStateShowFlags[routing.RelayStateOffline] = true
					}

					if relaysStateShowFlags[routing.RelayStateDecommissioned] {
						//  Show decommissioned relays with --decommissioned flag by essentially disabling --nodecommissioned flag
						relaysStateHideFlags[routing.RelayStateDecommissioned] = false
					}

					regex := ".*"
					if len(args) > 0 {
						regex = args[0]
					}

					checkRelays(rpcClient, env, regex, relaysStateShowFlags, relaysStateHideFlags, relaysDownFlag, csvOutputFlag)
					return nil
				},
			},
			{
				Name:       "keys",
				ShortUsage: "next relay keys <relay name>",
				ShortHelp:  "Show the public keys for the relay",
				Exec: func(ctx context.Context, args []string) error {
					relays := getRelayInfo(rpcClient, env, args[0])

					if len(relays) == 0 {
						handleRunTimeError(fmt.Sprintf("no relays matched the name '%s'\n", args[0]), 0)
					}

					relay := &relays[0]

					fmt.Printf("Public Key: %s\n", relay.publicKey)
					fmt.Printf("Update Key: %s\n", relay.updateKey)

					return nil
				},
			},
			{
				Name:       "update",
				ShortUsage: "next relay update [regex...]",
				ShortHelp:  "Update the specified relay(s)",
				FlagSet:    relayupdatefs,
				Exec: func(ctx context.Context, args []string) error {
					var regexes []string
					if len(args) > 0 {
						regexes = args
					}

					updateRelays(env, rpcClient, regexes, updateOpts)

					return nil
				},
			},
			{
				Name:       "revert",
				ShortUsage: "next relay revert [regex...]",
				ShortHelp:  "revert all or some relays to the last binary placed on the server",
				Exec: func(ctx context.Context, args []string) error {
					regexes := []string{".*"}
					if len(args) > 0 {
						regexes = args
					}

					revertRelays(env, rpcClient, regexes)

					return nil
				},
			},
			{
				Name:       "enable",
				ShortUsage: "next relay enable [regex...]",
				ShortHelp:  "Enable the specified relay(s)",
				Exec: func(_ context.Context, args []string) error {
					regexes := []string{".*"}
					if len(args) > 0 {
						regexes = args
					}

					enableRelays(env, rpcClient, regexes)

					return nil
				},
			},
			{
				Name:       "disable",
				ShortUsage: "next relay disable [regex...]",
				ShortHelp:  "Disable the specified relay(s)",
				FlagSet:    relaydisablefs,
				Exec: func(_ context.Context, args []string) error {
					regexes := []string{".*"}
					if len(args) > 0 {
						regexes = args
					}

					disableRelays(env, rpcClient, regexes, hardDisable, false)

					return nil
				},
			},
			{
				Name:       "maintenance",
				ShortUsage: "next relay maintenance [regex...]",
				ShortHelp:  "Move the specified relay(s) to maintenance mode",
				FlagSet:    relaydisablefs,
				Exec: func(_ context.Context, args []string) error {
					regexes := []string{".*"}
					if len(args) > 0 {
						regexes = args
					}

					disableRelays(env, rpcClient, regexes, hardDisable, true)

					return nil
				},
			},
			{
				Name:       "speed",
				ShortUsage: "next relay speed <relay name> <value (Mbps)>",
				ShortHelp:  "sets the speed value of a relay",
				Exec: func(_ context.Context, args []string) error {
					if len(args) == 0 {
						handleRunTimeError(fmt.Sprintln("You need to supply a relay name"), 0)
					}

					if len(args) == 1 {
						handleRunTimeError(fmt.Sprintln("You need to supply a relay NIC speed in Mbps"), 0)
					}

					nicSpeed, err := strconv.ParseUint(args[1], 10, 64)
					if err != nil {
						handleRunTimeError(fmt.Sprintf("Unable to parse %s as uint64\n", args[1]), 1)
					}

					setRelayNIC(rpcClient, env, args[0], nicSpeed)

					return nil
				},
			},
			{
				Name:       "state",
				ShortUsage: "next relay state <state> <regex> [regex...]",
				ShortHelp:  "Sets the relay state directly",
				LongHelp:   "This command should be avoided unless something goes wrong and the operator knows what he or she is doing.\nState values:\nenabled\noffline\nmaintenance\ndisabled\nquarantine\ndecommissioned",
				Exec: func(ctx context.Context, args []string) error {
					if len(args) == 0 {
						handleRunTimeError(fmt.Sprintln("You need to supply a relay state"), 0)
					}

					if len(args) == 1 {
						handleRunTimeError(fmt.Sprintln("You need to supply at least one relay name regex"), 0)
					}

					setRelayState(rpcClient, env, args[0], args[1:])
					return nil
				},
			},
			{
				Name:       "rename",
				ShortUsage: "next relay rename <old name> <new name>",
				ShortHelp:  "Rename the specified relay",
				Exec: func(ctx context.Context, args []string) error {
					if len(args) == 0 {
						handleRunTimeError(fmt.Sprintln("You need to supply a current relay name and a new name for it."), 0)
					}

					if len(args) == 1 {
						handleRunTimeError(fmt.Sprintln("You need to supply a new name for the relay as well"), 0)
					}

					updateRelayName(rpcClient, env, args[0], args[1])

					return nil
				},
			},
			{
				Name:       "add",
				ShortUsage: "next relay add <filepath>",
				ShortHelp:  "Add relay(s) from a JSON file or piped from stdin",
				Exec: func(_ context.Context, args []string) error {
					jsonData := readJSONData("relays", args)

					// Unmarshal the JSON and create the relay struct
					var relay relay
					if err := json.Unmarshal(jsonData, &relay); err != nil {
						handleRunTimeError(fmt.Sprintf("Could not unmarshal relay: %v\n", err), 1)
					}

					addr, err := net.ResolveUDPAddr("udp", relay.Addr)
					if err != nil {
						handleRunTimeError(fmt.Sprintf("Could not resolve udp address %s: %v\n", relay.Addr, err), 1)
					}

					publicKey, err := base64.StdEncoding.DecodeString(relay.PublicKey)
					if err != nil {
						handleRunTimeError(fmt.Sprintf("Could not decode bas64 public key %s: %v\n", relay.PublicKey, err), 1)
					}

					// Build the actual Relay struct from the input relay struct
					realRelay := routing.Relay{
						ID:        crypto.HashID(relay.Addr),
						Name:      relay.Name,
						Addr:      *addr,
						PublicKey: publicKey,
						Seller: routing.Seller{
							ID: relay.SellerID,
						},
						Datacenter: routing.Datacenter{
							ID:   crypto.HashID(relay.DatacenterName),
							Name: relay.DatacenterName,
						},
						NICSpeedMbps:        int32(relay.NicSpeedMbps),
						IncludedBandwidthGB: int32(relay.IncludedBandwidthGB),
						State:               routing.RelayStateMaintenance,
						ManagementAddr:      relay.ManagementAddr,
						SSHUser:             relay.SSHUser,
						SSHPort:             relay.SSHPort,
					}

					// Add the Relay to storage
					addRelay(rpcClient, env, realRelay)
					return nil
				},
				Subcommands: []*ffcli.Command{
					{
						Name:       "example",
						ShortUsage: "next relay add example",
						ShortHelp:  "Displays an example relay for the correct JSON schema",
						Exec: func(_ context.Context, args []string) error {
							example := relay{
								Name:                "amazon.ohio.2",
								Addr:                "127.0.0.1:40000",
								PublicKey:           "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
								SellerID:            "5tCm7KjOw3EBYojLe6PC",
								DatacenterName:      "amazon.ohio.2",
								NicSpeedMbps:        1000,
								IncludedBandwidthGB: 1,
								ManagementAddr:      "127.0.0.1",
								SSHUser:             "root",
								SSHPort:             40000,
							}

							jsonBytes, err := json.MarshalIndent(example, "", "\t")
							if err != nil {
								handleRunTimeError(fmt.Sprintln("Failed to marshal relay struct"), 1)
							}

							fmt.Println("Example JSON schema to add a new relay:")
							fmt.Println(string(jsonBytes))
							return nil
						},
					},
				},
			},
			{
				Name:       "remove",
				ShortUsage: "next relay remove <name>",
				ShortHelp:  "Remove a relay from storage",
				Exec: func(_ context.Context, args []string) error {
					if len(args) == 0 {
						handleRunTimeError(fmt.Sprintln("Provide the relay name of the relay you wish to remove\nFor a list of relay, use next relay"), 0)
					}

					removeRelay(rpcClient, env, args[0])
					return nil
				},
			},
			{
				Name:       "ops",
				ShortUsage: "next relay ops <command>",
				ShortHelp:  "Operations-related relay setup commands",
				Exec: func(_ context.Context, args []string) error {
					if len(args) == 0 {
						handleRunTimeError(fmt.Sprintln("Please provide at least one command name"), 0)
					}

					removeRelay(rpcClient, env, args[0])
					return nil
				},
				Subcommands: []*ffcli.Command{
					{
						Name:       "mrc",
						ShortUsage: "next relay ops mrc <relay> <value>",
						ShortHelp:  "Set the mrc value for the given relay (in $USD)",
						Exec: func(_ context.Context, args []string) error {
							if len(args) != 2 {
								handleRunTimeError(fmt.Sprintln("Must provide the relay name and an MRC value in $USD"), 0)
								return nil
							}

							mrc, err := strconv.ParseFloat(args[1], 64)
							if err != nil {
								handleRunTimeError(fmt.Sprintf("Could not parse %s as a decimal number\n", args[1]), 1)
							}
							opsMRC(rpcClient, env, args[0], mrc)

							return nil
						},
					},
					{
						Name:       "overage",
						ShortUsage: "next relay ops overage <relay> <value>",
						ShortHelp:  "Set the overage value for the given relay (in $USD)",
						Exec: func(_ context.Context, args []string) error {
							if len(args) != 2 {
								handleRunTimeError(fmt.Sprintln("Must provide the relay name and an overage value in $USD"), 0)
							}

							overage, err := strconv.ParseFloat(args[1], 64)
							if err != nil {
								handleRunTimeError(fmt.Sprintf("Could not parse %s as a decimal number\n", args[1]), 1)
							}
							opsOverage(rpcClient, env, args[0], overage)

							return nil
						},
					},
					{
						Name:       "bwrule",
						ShortUsage: "next relay ops bwrule <relay> <pool|burst|flat|none>",
						ShortHelp:  "Set the bandwidth rule for the given relay",
						Exec: func(_ context.Context, args []string) error {
							if len(args) != 2 {
								handleRunTimeError(fmt.Sprintln("Must provide the relay name and abandwidth rule"), 0)
							}

							rules := []string{"none", "flat", "pool", "burst"}
							for _, rule := range rules {
								if rule == args[1] {
									opsBWRule(rpcClient, env, args[0], args[1])
									return nil
								}
							}

							handleRunTimeError(fmt.Sprintf("'%s' not a valid bandwidth rule - must be one of none, flat, pool or burst\n", args[1]), 0)
							return nil
						},
					},
					{
						Name:       "type",
						ShortUsage: "next relay ops type <relay> <bare|vm|none>",
						ShortHelp:  "Set the machine/server type for the given relay",
						Exec: func(_ context.Context, args []string) error {
							if len(args) != 2 {
								handleRunTimeError(fmt.Sprintln("Must provide the relay name and a machine type"), 0)
							}

							rules := []string{"none", "vm", "bare"}
							for _, rule := range rules {
								if rule == args[1] {
									opsType(rpcClient, env, args[0], args[1])
									return nil
								}
							}

							handleRunTimeError(fmt.Sprintf("'%s' not a valid machine type - must be one of none, vm or bare\n", args[1]), 0)
							return nil
						},
					},
					{
						Name:       "term",
						ShortUsage: "next relay ops term <relay> <value>",
						ShortHelp:  "Set the contract term for the given relay (in months)",
						Exec: func(_ context.Context, args []string) error {
							if len(args) != 2 {
								handleRunTimeError(fmt.Sprintln("Must provide the relay name and a contract term in months"), 0)
							}

							term, err := strconv.ParseUint(args[1], 10, 32)
							if err != nil {
								handleRunTimeError(fmt.Sprintf("Could not parse %s as an integer\n", args[1]), 0)
							}
							opsTerm(rpcClient, env, args[0], int32(term))

							return nil
						},
					},
					{
						Name:       "startdate",
						ShortUsage: "next relay ops startdate <relay> <value, e.g. 'January 2, 2006'>",
						ShortHelp:  "Set the contract start date for the given relay (e.g. 'January 2, 2006')",
						Exec: func(_ context.Context, args []string) error {
							if len(args) != 2 {
								handleRunTimeError(fmt.Sprintln("Must provide the relay name and a contract start date e.g. 'January 2, 2006'"), 0)
							}

							opsStartDate(rpcClient, env, args[0], args[1])
							return nil
						},
					},
					{
						Name:       "enddate",
						ShortUsage: "next relay ops enddate <relay> <value, e.g. 'January 2, 2006'>",
						ShortHelp:  "Set the contract end date for the given relay (e.g. 'January 2, 2006')",
						Exec: func(_ context.Context, args []string) error {
							if len(args) != 2 {
								handleRunTimeError(fmt.Sprintln("Must provide the relay name and a contract end date e.g. 'January 2, 2006'"), 0)
							}

							opsEndDate(rpcClient, env, args[0], args[1])
							return nil
						},
					},
					{
						Name:       "bandwidth",
						ShortUsage: "next relay ops bandwidth <relay> <value in GB, e.g. 20000>",
						ShortHelp:  "Set the bandwidth allotment for the give relay",
						Exec: func(_ context.Context, args []string) error {
							if len(args) != 2 {
								handleRunTimeError(fmt.Sprintln("Must provide the relay name and a bandwidth value in GB"), 0)
							}

							bw, err := strconv.ParseUint(args[1], 10, 32)
							if err != nil {
								handleRunTimeError(fmt.Sprintf("Could not parse %s as an integer\n", args[1]), 0)
							}

							opsBandwidth(rpcClient, env, args[0], int32(bw))
							return nil
						},
					},
					{
						Name:       "nic",
						ShortUsage: "next relay ops nic <relay> <value in Mbps, e.g. 10000>",
						ShortHelp:  "Set the NIC available to the specified relay",
						Exec: func(_ context.Context, args []string) error {
							if len(args) != 2 {
								handleRunTimeError(fmt.Sprintln("Must provide the relay name and a value in Mbps"), 0)
							}

							nic, err := strconv.ParseUint(args[1], 10, 32)
							if err != nil {
								handleRunTimeError(fmt.Sprintf("Could not parse %s as an integer\n", args[1]), 0)
							}

							opsNic(rpcClient, env, args[0], int32(nic))
							return nil
						},
					},
				},
			},
		},
	}

	var routesCommand = &ffcli.Command{
		Name:       "routes",
		ShortUsage: "next routes <name-1> <name-2>",
		ShortHelp:  "List routes between relays",
		Exec: func(_ context.Context, args []string) error {

			if len(args) == 0 {
				routes(rpcClient, env, []string{}, []string{}, 0, 0)
				return nil
			}

			routes(rpcClient, env, []string{args[0]}, []string{args[1]}, 0, 0)
			return nil
		},
		Subcommands: []*ffcli.Command{
			{
				Name:       "selection",
				ShortUsage: "next routes selection <relay name>",
				ShortHelp:  "Select routes between sets of relays",
				FlagSet:    routesfs,
				Exec: func(ctx context.Context, args []string) error {
					routes(rpcClient, env, srcRelays, destRelays, routeRTT, routeHash)

					return nil
				},
			},
		},
	}

	var datacentersCommand = &ffcli.Command{
		Name:       "datacenters",
		ShortUsage: "next datacenters",
		ShortHelp:  "List datacenters",
		FlagSet:    datacentersfs,
		Exec: func(_ context.Context, args []string) error {
			if len(args) > 0 {
				datacenters(rpcClient, env, args[0], datacenterIdSigned)
				return nil
			}
			datacenters(rpcClient, env, "", datacenterIdSigned)
			return nil
		},
	}

	var customersCommand = &ffcli.Command{
		Name:       "customers",
		ShortUsage: "next customers",
		ShortHelp:  "List customers",
		Exec: func(_ context.Context, args []string) error {
			customers(rpcClient, env)
			return nil
		},
	}

	var sellersCommand = &ffcli.Command{
		Name:       "sellers",
		ShortUsage: "next sellers",
		ShortHelp:  "List sellers",
		Exec: func(_ context.Context, args []string) error {
			sellers(rpcClient, env)
			return nil
		},
	}

	var datacenterCommand = &ffcli.Command{
		Name:       "datacenter",
		ShortUsage: "next datacenter <subcommand>",
		ShortHelp:  "Manage datacenters",
		Exec: func(_ context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			{ // add
				Name:       "add",
				ShortUsage: "next datacenter add <filepath>",
				ShortHelp:  "Add a datacenter to storage from a JSON file or piped from stdin",
				Exec: func(_ context.Context, args []string) error {
					jsonData := readJSONData("datacenters", args)

					// Unmarshal the JSON and create the datacenter struct
					var datacenter datacenter
					if err := json.Unmarshal(jsonData, &datacenter); err != nil {
						handleRunTimeError(fmt.Sprintf("Could not unmarshal datacenter: %v\n", err), 1)
					}

					// Build the actual Datacenter struct from the input datacenter struct
					realDatacenter := routing.Datacenter{
						ID:       crypto.HashID(datacenter.Name),
						Name:     datacenter.Name,
						Enabled:  datacenter.Enabled,
						Location: datacenter.Location,
					}

					// Add the Datacenter to storage
					addDatacenter(rpcClient, env, realDatacenter)
					return nil
				},
				Subcommands: []*ffcli.Command{
					{
						Name:       "example",
						ShortUsage: "next datacenter add example",
						ShortHelp:  "Displays an example datacenter for the correct JSON schema",
						Exec: func(_ context.Context, args []string) error {
							example := datacenter{
								Name:     "amazon.ohio.2",
								Enabled:  false,
								Location: routing.LocationNullIsland,
							}

							jsonBytes, err := json.MarshalIndent(example, "", "\t")
							if err != nil {
								handleRunTimeError(fmt.Sprintln("Failed to marshal datacenter struct"), 1)
							}

							fmt.Println("Example JSON schema to add a new datacenter:")
							fmt.Println(string(jsonBytes))
							return nil
						},
					},
				},
			},
			{ // remove
				Name:       "remove",
				ShortUsage: "next datacenter remove <name>",
				ShortHelp:  "Remove a datacenter from storage",
				Exec: func(_ context.Context, args []string) error {
					if len(args) == 0 {
						err := errors.New("Provide the datacenter name of the datacenter you wish to remove\nFor a list of datacenters, use next datacenters")
						return err
					}

					removeDatacenter(rpcClient, env, args[0])
					return nil
				},
			},
			{ //buyers
				Name:       "buyers",
				ShortUsage: "next datacenter buyers <datacenter ID|name>",
				ShortHelp:  "Returns a list of all buyers and aliases for a given datacenter",
				LongHelp:   "Returns a list of all buyers and aliases for a given datacenter. ",
				Exec: func(_ context.Context, args []string) error {

					if len(args) == 0 {
						handleRunTimeError(fmt.Sprintln("Exactly zero or one datacenter ID or name must be provided."), 0)
					}

					listDatacenterMaps(rpcClient, env, args[0])
					return nil
				},
			},
		},
	}

	var buyersCommand = &ffcli.Command{
		Name:       "buyers",
		ShortUsage: "next buyers",
		ShortHelp:  "Return a list of all current buyers",
		FlagSet:    buyersfs,
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 0 {
				fmt.Println("No arguments necessary, everything after 'buyers' is ignored.\n\nA list of all current buyers:")
			}
			buyers(rpcClient, env, buyersIdSigned)
			return nil
		},
	}

	var buyerCommand = &ffcli.Command{
		Name:       "buyer",
		ShortUsage: "next buyer <subcommand>",
		ShortHelp:  "Manage buyers",
		FlagSet:    buyersfs,
		Exec: func(_ context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			{ // list
				Name:       "list",
				ShortUsage: "next buyer list",
				ShortHelp:  "Return a list of all current buyers",
				Exec: func(_ context.Context, args []string) error {
					buyers(rpcClient, env, buyersIdSigned)
					return nil
				},
			},
			{ // add
				Name:       "add",
				ShortUsage: "next buyer add [filepath]",
				ShortHelp:  "Add a buyer from a JSON file or piped from stdin",
				Exec: func(_ context.Context, args []string) error {
					jsonData := readJSONData("buyers", args)

					// Unmarshal the JSON and create the Buyer struct
					var b buyer
					if err := json.Unmarshal(jsonData, &b); err != nil {
						handleRunTimeError(fmt.Sprintf("Could not unmarshal buyer: %v\n", err), 1)
					}

					// Get the ID from the first 8 bytes of the public key
					publicKey, err := base64.StdEncoding.DecodeString(b.PublicKey)
					if err != nil {
						handleRunTimeError(fmt.Sprintf("Could not get buyer ID from public key: %v\n", err), 1)
					}

					if len(publicKey) != crypto.KeySize+8 {
						handleRunTimeError(fmt.Sprintf("Invalid public key length %d\n", len(publicKey)), 1)
					}

					// Add the Buyer to storage
					addBuyer(rpcClient, env, routing.Buyer{
						ID:        binary.LittleEndian.Uint64(publicKey[:8]),
						Name:      b.Name,
						Domain:    b.Domain,
						Active:    b.Active,
						Live:      b.Live,
						PublicKey: publicKey,
					})
					return nil
				},
				Subcommands: []*ffcli.Command{
					{
						Name:       "example",
						ShortUsage: "next buyer add example",
						ShortHelp:  "Displays an example buyer for the correct JSON schema",
						Exec: func(_ context.Context, args []string) error {
							examplePublicKey := make([]byte, crypto.KeySize+8) // 8 bytes for buyer ID
							examplePublicKeyString := base64.StdEncoding.EncodeToString(examplePublicKey)

							example := buyer{
								Name:      "Psyonix",
								Domain:    "example.com",
								Active:    true,
								Live:      true,
								PublicKey: examplePublicKeyString,
							}

							jsonBytes, err := json.MarshalIndent(example, "", "\t")
							if err != nil {
								handleRunTimeError(fmt.Sprintln("Failed to marshal buyer struct"), 1)
							}

							fmt.Println("Example JSON schema to add a new buyer:")
							fmt.Println(string(jsonBytes))
							return nil
						},
					},
				},
			},
			{ // remove
				Name:       "remove",
				ShortUsage: "next buyer remove <id>",
				ShortHelp:  "Remove a buyer from storage",
				Exec: func(_ context.Context, args []string) error {
					if len(args) == 0 {
						handleRunTimeError(fmt.Sprintln("Provide the buyer ID of the buyer you wish to remove\nFor a list of buyers, use next buyers"), 0)
					}

					removeBuyer(rpcClient, env, args[0])
					return nil
				},
			},
			{
				Name:       "datacenters",
				ShortUsage: "next buyer datacenters <buyer id|name|string, optional>",
				ShortHelp:  "Return a list of datacenters and aliases for the given buyer ID or buyer name",
				Exec: func(_ context.Context, args []string) error {
					if len(args) != 1 {
						datacenterMapsForBuyer(rpcClient, env, "")
						return nil
					}

					datacenterMapsForBuyer(rpcClient, env, args[0])
					return nil
				},
			},
			{
				Name:       "datacenter",
				ShortUsage: "next buyer datacenter <command>",
				ShortHelp:  "Display and manipulate datacenters and aliases",
				Exec: func(_ context.Context, args []string) error {
					return flag.ErrHelp
				},
				Subcommands: []*ffcli.Command{
					{
						Name:       "list",
						ShortUsage: "next buyer datacenter list <buyer id|name|substring>",
						ShortHelp:  "Return a list of datacenters and aliases for the given buyer ID or buyer name",
						LongHelp:   "A buyer ID or name must be supplied. If the name includes spaces it must be enclosed in quotations marks.",
						Exec: func(_ context.Context, args []string) error {
							if len(args) != 1 {
								datacenterMapsForBuyer(rpcClient, env, "")
								return nil
							}

							datacenterMapsForBuyer(rpcClient, env, args[0])
							return nil
						},
					},
					{
						Name:       "add",
						ShortUsage: "next buyer datacenter add <json file>",
						ShortHelp:  "Create a new datacenter/alias entry using info supplied in a json file (see -h for an example)",
						LongHelp: `Reads the specifics for the new datacenter alias entry from
the contents of the specified json file. The json file layout
is as follows:

{
	"alias": "some.server.alias",
	"datacenter": "2fe32c22450fb4c9",
	"buyer_id": "bdbebdbf0f7be395"
}

The buyer and datacenter must exist. Hex IDs are required for now.
						`,
						Exec: func(_ context.Context, args []string) error {
							var err error

							if len(args) == 0 {
								handleRunTimeError(fmt.Sprintf("An input file name must be supplied. For more info run:\n\n./next buyer datacenter add -h\n"), 0)
							}

							jsonData, err := ioutil.ReadFile(args[0])
							if err != nil {
								handleRunTimeError(fmt.Sprintf("Error reading JSON input file: %s\n", args[0]), 1)
							}

							// Unmarshal the JSON and create the Buyer struct
							var dcmStrings dcMapStrings
							if err = json.Unmarshal(jsonData, &dcmStrings); err != nil {
								handleRunTimeError(fmt.Sprintf("Could not unmarshal datacenter map: %v\n", err), 1)
							}

							err = addDatacenterMap(rpcClient, env, dcmStrings)

							if err != nil {
								// error handled in sender
								return nil
							}

							fmt.Println("Datacenter alias created:")
							fmt.Println(dcmStrings)

							return nil
						},
					},
					{
						Name:       "remove",
						ShortUsage: "next buyer datacenter remove <json file>",
						ShortHelp:  "Removes the datacenter alias described in the json file from the system (see -h for an example)",
						LongHelp: `Reads the specifics for the datacenter alias to be removed from
the contents of the specified json file. The json file layout
is as follows:

{
	"alias": "some.server.alias",
	"datacenter": "2fe32c22450fb4c9",
	"buyer_id": "bdbebdbf0f7be395"
}

The alias is uniquely defined by all three entries, so they must be provided. Hex IDs and names are supported."
						`,
						Exec: func(_ context.Context, args []string) error {
							var err error

							if len(args) == 0 {
								handleRunTimeError(fmt.Sprintln("An input file name must be supplied. For more info run:\n\n./next buyer datacenter remove -h"), 0)
							}

							jsonData, err := ioutil.ReadFile(args[0])
							if err != nil {
								handleRunTimeError(fmt.Sprintf("Error reading JSON input file: %s\n", args[0]), 1)
							}

							// Unmarshal the JSON and create the Buyer struct
							var dcmStrings dcMapStrings
							if err = json.Unmarshal(jsonData, &dcmStrings); err != nil {
								handleRunTimeError(fmt.Sprintf("Could not unmarshal datacenter map: %v\n", err), 1)
							}

							err = removeDatacenterMap(rpcClient, env, dcmStrings)

							if err != nil {
								return err
							}

							fmt.Println("Datacenter alias removed.")

							return nil
						},
					},
				},
			},
		},
	}

	var sellerCommand = &ffcli.Command{
		Name:       "seller",
		ShortUsage: "next seller <subcommand>",
		ShortHelp:  "Manage sellers",
		Exec: func(_ context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			{
				Name:       "add",
				ShortUsage: "next seller add [filepath]",
				ShortHelp:  "Add a seller to storage from a JSON file or piped from stdin",
				Exec: func(_ context.Context, args []string) error {
					jsonData := readJSONData("sellers", args)

					// Unmarshal the JSON and create the Seller struct
					var sellerUSD struct {
						Name            string
						IngressPriceUSD string
						EgressPriceUSD  string
					}

					if err := json.Unmarshal(jsonData, &sellerUSD); err != nil {
						handleRunTimeError(fmt.Sprintf("Could not unmarshal seller: %v\n", err), 1)
					}

					ingressUSD, err := strconv.ParseFloat(sellerUSD.IngressPriceUSD, 64)
					if err != nil {
						fmt.Printf("Unable to convert %s to a decimal number.", sellerUSD.IngressPriceUSD)
						os.Exit(0)
					}
					egressUSD, err := strconv.ParseFloat(sellerUSD.EgressPriceUSD, 64)
					if err != nil {
						fmt.Printf("Unable to convert %s to a decimal number.", sellerUSD.EgressPriceUSD)
						os.Exit(0)
					}

					s := seller{
						Name:                 sellerUSD.Name,
						IngressPriceNibblins: routing.DollarsToNibblins(ingressUSD),
						EgressPriceNibblins:  routing.DollarsToNibblins(egressUSD),
					}

					// Add the Seller to storage
					addSeller(rpcClient, env, routing.Seller{
						ID:                        s.Name,
						Name:                      s.Name,
						IngressPriceNibblinsPerGB: s.IngressPriceNibblins,
						EgressPriceNibblinsPerGB:  s.EgressPriceNibblins,
					})
					return nil
				},
				Subcommands: []*ffcli.Command{
					{
						Name:       "example",
						ShortUsage: "next seller add example",
						ShortHelp:  "Displays an example seller for the correct JSON schema",
						Exec: func(_ context.Context, args []string) error {
							example := struct {
								Name            string
								IngressPriceUSD string
								EgressPriceUSD  string
							}{
								Name:            "amazon",
								IngressPriceUSD: "0.01",
								EgressPriceUSD:  "0.1",
							}

							jsonBytes, err := json.MarshalIndent(example, "", "\t")
							if err != nil {
								handleRunTimeError(fmt.Sprintln("Failed to marshal seller struct"), 1)
							}

							fmt.Println("Example JSON schema to add a new seller - note that prices are in $USD:")
							fmt.Println(string(jsonBytes))
							return nil
							return nil
						},
					},
				},
			},
			{
				Name:       "remove",
				ShortUsage: "next seller remove <id>",
				ShortHelp:  "Remove a seller from storage",
				Exec: func(_ context.Context, args []string) error {
					if len(args) == 0 {
						handleRunTimeError(fmt.Sprintln("Provide the seller ID of the seller you wish to remove\nFor a list of sellers, use next sellers"), 0)
					}

					removeSeller(rpcClient, env, args[0])
					return nil
				},
			},
		},
	}

	var shaderCommand = &ffcli.Command{
		Name:       "shader",
		ShortUsage: "next shader <buyer name or substring>",
		ShortHelp:  "Retrieve route shader settings for the specified buyer",
		Exec: func(_ context.Context, args []string) error {
			if len(args) == 0 {
				handleRunTimeError(fmt.Sprintf("No buyer name or substring provided.\nUsage:\nnext shader <buyer name or substring>\n"), 0)
			}

			// Get the buyer's route shader
			routingRulesSettings(rpcClient, env, args[0])
			return nil
		},
		Subcommands: []*ffcli.Command{
			{
				Name:       "set",
				ShortUsage: "next shader set <buyer ID> [filepath]",
				ShortHelp:  "Set the buyer's route shader in storage from a JSON file or piped from stdin",
				Exec: func(_ context.Context, args []string) error {
					if len(args) == 0 {
						handleRunTimeError(fmt.Sprintf("No buyer ID provided.\nUsage:\nnext shader set <buyer ID> [filepath]\nbuyer ID: the buyer's ID\n(Optional) filepath: the filepath to a JSON file with the new route shader data. If this data is piped through stdin, this parameter is optional.\nFor a list of buyers, use next buyers\n"), 0)
					}

					jsonData := readJSONData("buyers", args[1:])

					// Unmarshal the JSON and create the RoutingRuleSettings struct
					var rrs routing.RoutingRulesSettings
					if err := json.Unmarshal(jsonData, &rrs); err != nil {
						handleRunTimeError(fmt.Sprintf("Could not unmarshal route shader: %v\n", err), 1)
					}

					// Set the route shader in storage
					setRoutingRulesSettings(rpcClient, env, args[0], rrs)
					return nil
				},
				Subcommands: []*ffcli.Command{
					{
						Name:       "example",
						ShortUsage: "next shader set example",
						ShortHelp:  "Displays an example route shader for the correct JSON schema",
						Exec: func(_ context.Context, args []string) error {
							jsonBytes, err := json.MarshalIndent(routing.DefaultRoutingRulesSettings, "", "\t")
							if err != nil {
								handleRunTimeError(fmt.Sprintln("Failed to marshal route shader struct"), 0)
							}

							fmt.Println("Example JSON schema to set a new route shader:")
							fmt.Println(string(jsonBytes))
							return nil
						},
					},
				},
			},
			{
				Name:       "id",
				ShortUsage: "next shader id <buyer ID>",
				ShortHelp:  "Retrieve route shader information for the given buyer ID",
				Exec: func(_ context.Context, args []string) error {
					if len(args) == 0 {
						handleRunTimeError(fmt.Sprintf("No buyer ID provided.\nUsage:\nnext shader <buyer ID>\nbuyer ID: the buyer's ID\nFor a list of buyers, use next buyers\n"), 0)
					}

					// Get the buyer's route shader
					routingRulesSettingsByID(rpcClient, env, args[0])
					return nil
				},
			},
		},
	}

	var customerCommand = &ffcli.Command{
		Name:       "customer",
		ShortUsage: "next customer <subcommand>",
		ShortHelp:  "Manage customers",
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			{
				Name:       "link",
				ShortUsage: "next customer link <subcommand>",
				ShortHelp:  "Edit customer links",
				Exec: func(ctx context.Context, args []string) error {
					return flag.ErrHelp
				},
				Subcommands: []*ffcli.Command{
					{
						Name:       "buyer",
						ShortUsage: "next customer link buyer <customer name> <new buyer ID>",
						ShortHelp:  "Edit what buyer this customer is linked to",
						Exec: func(ctx context.Context, args []string) error {
							if len(args) == 0 {
								handleRunTimeError(fmt.Sprintln("You need to provide a customer name"), 0)
							}

							if len(args) == 1 {
								handleRunTimeError(fmt.Sprintln("You need to provide a new buyer ID for the customer to link to"), 0)
							}

							buyerID, err := strconv.ParseUint(args[1], 10, 64)
							if err != nil {
								handleRunTimeError(fmt.Sprintf("Could not parse %s as an unsigned 64-bit integer\n", args[1]), 1)
							}

							customerLink(rpcClient, env, args[0], buyerID, "")
							return nil
						},
					},
					{
						Name:       "seller",
						ShortUsage: "next customer link seller <customer name> <new seller ID>",
						ShortHelp:  "Edit what seller this customer is linked to",
						Exec: func(ctx context.Context, args []string) error {
							if len(args) == 0 {
								handleRunTimeError(fmt.Sprintln("You need to provide a customer name"), 0)
							}

							if len(args) == 1 {
								handleRunTimeError(fmt.Sprintln("You need to provide a new seller ID for the customer to link to"), 0)
							}

							customerLink(rpcClient, env, args[0], 0, args[1])
							return nil
						},
					},
				},
			},
		},
	}

	var sshCommand = &ffcli.Command{
		Name:       "ssh",
		ShortUsage: "next ssh <relay name>",
		ShortHelp:  "SSH into a relay by name",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				handleRunTimeError(fmt.Sprintln("You need to supply a device identifer"), 0)
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
			saveCostMatrix(env, output)
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

	var debugCommand = &ffcli.Command{
		Name:       "debug",
		ShortUsage: "next debug <relay_name> [input_file]",
		ShortHelp:  "Debug tool",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				handleRunTimeError(fmt.Sprintln("You need to supply a relay name"), 0)
			}
			relayName := args[0]
			inputFile := "optimize.bin"
			if len(args) > 1 {
				inputFile = args[1]
			}
			debug(relayName, inputFile)
			return nil
		},
	}

	var viewCommand = &ffcli.Command{
		Name:       "view",
		ShortUsage: "next view <subcommand>",
		ShortHelp:  "View data",
		Exec: func(ctx context.Context, args []string) error {
			return nil
		},
		Subcommands: []*ffcli.Command{
			{
				Name:       "cost",
				ShortUsage: "next view cost",
				ShortHelp:  "View the entries of the cost matrix",
				Exec: func(ctx context.Context, args []string) error {
					input := "cost.bin"
					viewCostMatrix(input)
					return nil
				},
			},
			{
				Name:       "route",
				ShortUsage: "next view route [srcRelay] [destRelay]",
				ShortHelp:  "View the entries of the route matrix with optional relay filtering.",
				Exec: func(ctx context.Context, args []string) error {
					input := "optimize.bin"

					var srcRelay string
					var destRelay string

					if len(args) > 0 {
						srcRelay = args[0]
					}

					if len(args) == 1 {
						handleRunTimeError(fmt.Sprintln("You must provide a destination relay if you provide a source relay. For all entries, omit the relay parameters"), 0)
					}

					if len(args) > 1 {
						destRelay = args[1]
					}

					viewRouteMatrix(input, srcRelay, destRelay)
					return nil
				},
			},
		},
	}

	var commands = []*ffcli.Command{
		authCommand,
		selectCommand,
		envCommand,
		sessionCommand,
		sessionsCommand,
		relaysCommand,
		relayCommand,
		routesCommand,
		datacentersCommand,
		datacenterCommand,
		customersCommand,
		customerCommand,
		sellersCommand,
		sellerCommand,
		buyerCommand,
		buyersCommand,
		shaderCommand,
		sshCommand,
		costCommand,
		optimizeCommand,
		analyzeCommand,
		debugCommand,
		viewCommand,
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
