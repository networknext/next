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
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
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

func readJSONData(entity string, args []string) []byte {
	// Check if the input is piped or a filepath
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("Error checking stdin stat: %v", err)
	}
	isPipedInput := fileInfo.Mode()&os.ModeCharDevice == 0

	var data []byte
	if isPipedInput {
		// Read the piped input from stdin
		data, err = ioutil.ReadAll(bufio.NewReader(os.Stdin))
		if err != nil {
			log.Fatalf("Error reading from stdin: %v", err)
		}
	} else {
		// Read the file at the given filepath
		if len(args) == 0 {
			log.Fatalf("Supply a file path to read the %s JSON or pipe it through stdin\nnext %s add [filepath]\nor\ncat <filepath> | next %s add\n\nFor an example JSON schema:\nnext %s add example", entity, entity, entity, entity)
		}

		data, err = ioutil.ReadFile(args[0])
		if err != nil {
			log.Fatalf("Error reading %s JSON file: %v", entity, err)
		}
	}

	return data
}

func handleJSONRPCError(env Environment, err error) {
	switch e := err.(type) {
	case *jsonrpc.HTTPError:
		switch e.Code {
		case http.StatusUnauthorized:
			log.Fatalf("%d: %s - use `next auth` to authorize the operator tool", e.Code, http.StatusText(e.Code))
		default:
			log.Fatalf("%d: %s", e.Code, http.StatusText(e.Code))
		}
	default:
		if env.Name != "local" && env.Name != "dev" && env.Name != "prod" {
			log.Fatalf("%v - make sure the env name is set to either 'prod', 'dev', or 'local' with\nnext select <env>", err)
		} else {
			log.Fatal(err)
		}
	}
}

type buyer struct {
	Name      string
	Domain    string
	Active    bool
	Live      bool
	PublicKey string
}

type seller struct {
	Name              string
	IngressPriceCents uint64
	EgressPriceCents  uint64
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

	var relayCoreCount uint64
	relayupdatefs := flag.NewFlagSet("relay update", flag.ExitOnError)
	relayupdatefs.Uint64Var(&relayCoreCount, "cores", 0, "number of cores for the relay to utilize")

	root := &ffcli.Command{
		ShortUsage: "next <subcommand>",
		Subcommands: []*ffcli.Command{
			{
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

					env.AuthToken = gjson.ParseBytes(body).Get("access_token").String()
					env.Write()

					fmt.Print("Successfully authorized\n")

					return nil
				},
			},
			{
				Name:       "select",
				ShortUsage: "next select <local|dev|prod>",
				ShortHelp:  "Select environment to use (local|dev|prod)",
				Exec: func(_ context.Context, args []string) error {
					if len(args) == 0 {
						log.Fatal("Provide an environment to switch to (local|dev|prod)")
					}

					if args[0] != "local" && args[0] != "dev" && args[0] != "prod" {
						log.Fatalf("Invalid environment %s: use (local|dev|prod)", args[0])
					}

					env.Name = args[0]
					env.Write()

					fmt.Printf("Selected %s environment\n", env.Name)
					return nil
				},
			},
			{
				Name:       "env",
				ShortUsage: "next env",
				ShortHelp:  "Display environment",
				Exec: func(_ context.Context, args []string) error {
					if len(args) > 0 {
						if args[0] != "local" && args[0] != "dev" && args[0] != "prod" {
							log.Fatalf("Invalid environment %s: use (local|dev|prod)", args[0])
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
			},

			// commands to print out helpful info in this section

			{
				Name:       "sessions",
				ShortUsage: "next sessions",
				ShortHelp:  "List sessions",
				Exec: func(_ context.Context, args []string) error {
					if len(args) > 0 {
						sessions(rpcClient, env, args[0])
						return nil
					}
					sessions(rpcClient, env, "")
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
			},

			{
				Name:       "relays",
				ShortUsage: "next relays <name>",
				ShortHelp:  "List relays",
				Exec: func(_ context.Context, args []string) error {
					if len(args) > 0 {
						relays(rpcClient, env, args[0])
						return nil
					}
					relays(rpcClient, env, "")
					return nil
				},
			},
			{
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
			},
			{
				Name:       "datacenters",
				ShortUsage: "next datacenters <name>",
				ShortHelp:  "List datacenters",
				Exec: func(_ context.Context, args []string) error {
					if len(args) > 0 {
						datacenters(rpcClient, env, args[0])
						return nil
					}
					datacenters(rpcClient, env, "")
					return nil
				},
			},
			{
				Name:       "customers",
				ShortUsage: "next customers",
				ShortHelp:  "List customers",
				Exec: func(_ context.Context, args []string) error {
					customers(rpcClient, env)
					return nil
				},
			},
			{
				Name:       "buyers",
				ShortUsage: "next buyers",
				ShortHelp:  "List buyers",
				Exec: func(_ context.Context, args []string) error {
					buyers(rpcClient, env)
					return nil
				},
			},
			{
				Name:       "sellers",
				ShortUsage: "next sellers",
				ShortHelp:  "List sellers",
				Exec: func(_ context.Context, args []string) error {
					sellers(rpcClient, env)
					return nil
				},
			},

			// more complex commands to modify things below here

			{
				Name:       "relay",
				ShortUsage: "next relay <subcommand>",
				ShortHelp:  "Manage relays",
				Subcommands: []*ffcli.Command{
					{
						Name:       "check",
						ShortUsage: "next relay check [filter]",
						ShortHelp:  "List all or a subset of relays and see diagnostic information. Refer to the README for more information",
						Exec: func(ctx context.Context, args []string) error {
							filter := ""

							if len(args) > 0 {
								filter = args[0]
							}

							checkRelays(rpcClient, env, filter)
							return nil
						},
					},
					{
						Name:       "keys",
						ShortUsage: "next relay keys <relay name>",
						ShortHelp:  "Show the public keys for the relay",
						Exec: func(ctx context.Context, args []string) error {
							relay := getRelayInfo(rpcClient, args[0])

							fmt.Printf("Public Key: %s\n", relay.publicKey)
							fmt.Printf("Update Key: %s\n", relay.updateKey)

							return nil
						},
					},
					{
						Name:       "update",
						ShortUsage: "next relay update <relay name...>",
						ShortHelp:  "Update the specified relay(s)",
						FlagSet:    relayupdatefs,
						Exec: func(ctx context.Context, args []string) error {
							if len(args) == 0 {
								log.Fatal("You need to supply at least one relay name")
							}

							updateRelays(env, rpcClient, args, relayCoreCount)

							return nil
						},
					},
					{
						Name:       "revert",
						ShortUsage: "next relay revert <ALL|relay name...>",
						ShortHelp:  "revert all or some relays to the last binary placed on the server",
						Exec: func(ctx context.Context, args []string) error {
							if len(args) == 0 {
								log.Fatal("You need to supply at least one relay name or 'ALL'")
							}

							revertRelays(env, rpcClient, args)

							return nil
						},
					},
					{
						Name:       "enable",
						ShortUsage: "next relay enable <relay name...>",
						ShortHelp:  "Enable the specified relay(s)",
						Exec: func(_ context.Context, args []string) error {
							if len(args) == 0 {
								log.Fatal("You need to supply at least one relay name")
							}

							enableRelays(env, rpcClient, args)

							return nil
						},
					},
					{
						Name:       "disable",
						ShortUsage: "next relay disable <relay name...>",
						ShortHelp:  "Disable the specified relay(s)",
						Exec: func(_ context.Context, args []string) error {
							if len(args) == 0 {
								log.Fatal("You need to supply at least one relay name")
							}

							disableRelays(env, rpcClient, args)

							return nil
						},
					},
					{
						Name:       "speed",
						ShortUsage: "next relay speed <relay name> <value (Mbps)>",
						ShortHelp:  "sets the speed value of a relay",
						Exec: func(_ context.Context, args []string) error {
							if len(args) == 0 {
								log.Fatal("You need to supply a relay name")
							}

							if len(args) == 1 {
								log.Fatal("You need to supply a relay NIC speed in Mbps")
							}

							nicSpeed, err := strconv.ParseUint(args[1], 10, 64)
							if err != nil {
								log.Fatalf("Unable to parse %s as uint64", args[1])
							}

							setRelayNIC(rpcClient, args[0], nicSpeed)

							return nil
						},
					},
					{
						Name:       "state",
						ShortUsage: "next relay state <state> <relay name> [relay names...]",
						ShortHelp:  "Sets the relay state directly",
						LongHelp:   "This command should be avoided unless something goes wrong and the operator knows what he or she is doing.\nState values:\nenabled\noffline\nmaintenance\ndisabled\nquarantine\ndecommissioned",
						Exec: func(ctx context.Context, args []string) error {
							if len(args) == 0 {
								log.Fatal("You need to supply a relay state")
							}

							if len(args) == 1 {
								log.Fatal("You need to supply at least one relay name")
							}

							setRelayState(rpcClient, args[0], args[1:])
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
								log.Fatalf("Could not unmarshal relay: %v", err)
							}

							addr, err := net.ResolveUDPAddr("udp", relay.Addr)
							if err != nil {
								log.Fatalf("Could not resolve udp address %s: %v", relay.Addr, err)
							}

							publicKey, err := base64.StdEncoding.DecodeString(relay.PublicKey)
							if err != nil {
								log.Fatalf("Could not decode bas64 public key %s: %v", relay.PublicKey, err)
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
								NICSpeedMbps:        relay.NicSpeedMbps,
								IncludedBandwidthGB: relay.IncludedBandwidthGB,
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
										log.Fatal("Failed to marshal relay struct")
									}

									fmt.Println("Exmaple JSON schema to add a new relay:")
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
								log.Fatal("Provide the relay name of the relay you wish to remove\nFor a list of relay, use next relay")
							}

							removeRelay(rpcClient, env, args[0])
							return nil
						},
					},
				},
			},
			{
				Name:       "datacenter",
				ShortUsage: "next datacenter <name>",
				ShortHelp:  "Manage datacenters",
				Exec: func(_ context.Context, args []string) error {
					if len(args) > 0 {
						datacenters(rpcClient, env, args[0])
						return nil
					}
					datacenters(rpcClient, env, "")
					return nil
				},
				Subcommands: []*ffcli.Command{
					{
						Name:       "add",
						ShortUsage: "next datacenter add <filepath>",
						ShortHelp:  "Add a datacenter to storage from a JSON file or piped from stdin",
						Exec: func(_ context.Context, args []string) error {
							jsonData := readJSONData("datacenters", args)

							// Unmarshal the JSON and create the datacenter struct
							var datacenter datacenter
							if err := json.Unmarshal(jsonData, &datacenter); err != nil {
								log.Fatalf("Could not unmarshal datacenter: %v", err)
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
										log.Fatal("Failed to marshal datacenter struct")
									}

									fmt.Println("Exmaple JSON schema to add a new datacenter:")
									fmt.Println(string(jsonBytes))
									return nil
								},
							},
						},
					},
					{
						Name:       "remove",
						ShortUsage: "next datacenter remove <name>",
						ShortHelp:  "Remove a datacenter from storage",
						Exec: func(_ context.Context, args []string) error {
							if len(args) == 0 {
								log.Fatal("Provide the datacenter name of the datacenter you wish to remove\nFor a list of datacenters, use next datacenters")
							}

							removeDatacenter(rpcClient, env, args[0])
							return nil
						},
					},
				},
			},
			{
				Name:       "buyer",
				ShortUsage: "next buyer",
				ShortHelp:  "Manage buyers",
				Exec: func(_ context.Context, args []string) error {
					buyers(rpcClient, env)
					return nil
				},
				Subcommands: []*ffcli.Command{
					{
						Name:       "add",
						ShortUsage: "next buyer add [filepath]",
						ShortHelp:  "Add a buyer from a JSON file or piped from stdin",
						Exec: func(_ context.Context, args []string) error {
							jsonData := readJSONData("buyers", args)

							// Unmarshal the JSON and create the Buyer struct
							var b buyer
							if err := json.Unmarshal(jsonData, &b); err != nil {
								log.Fatalf("Could not unmarshal buyer: %v", err)
							}

							// Get the ID from the first 8 bytes of the public key
							publicKey, err := base64.StdEncoding.DecodeString(b.PublicKey)
							if err != nil {
								log.Fatalf("Could not get buyer ID from public key: %v", err)
							}

							if len(publicKey) != crypto.KeySize+8 {
								log.Fatalf("Invalid public key length %d", len(publicKey))
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
										log.Fatal("Failed to marshal buyer struct")
									}

									fmt.Println("Exmaple JSON schema to add a new buyer:")
									fmt.Println(string(jsonBytes))
									return nil
								},
							},
						},
					},
					{
						Name:       "remove",
						ShortUsage: "next buyer remove <id>",
						ShortHelp:  "Remove a buyer from storage",
						Exec: func(_ context.Context, args []string) error {
							if len(args) == 0 {
								log.Fatal("Provide the buyer ID of the buyer you wish to remove\nFor a list of buyers, use next buyers")
							}

							removeBuyer(rpcClient, env, args[0])
							return nil
						},
					},
				},
			},
			{
				Name:       "seller",
				ShortUsage: "next sellers",
				ShortHelp:  "Manage sellers",
				Subcommands: []*ffcli.Command{
					{
						Name:       "add",
						ShortUsage: "next seller add [filepath]",
						ShortHelp:  "Add a seller to storage from a JSON file or piped from stdin",
						Exec: func(_ context.Context, args []string) error {
							jsonData := readJSONData("sellers", args)

							// Unmarshal the JSON and create the Seller struct
							var s seller
							if err := json.Unmarshal(jsonData, &s); err != nil {
								log.Fatalf("Could not unmarshal seller: %v", err)
							}

							// Add the Seller to storage
							addSeller(rpcClient, env, routing.Seller{
								ID:                s.Name,
								Name:              s.Name,
								IngressPriceCents: s.IngressPriceCents,
								EgressPriceCents:  s.EgressPriceCents,
							})
							return nil
						},
						Subcommands: []*ffcli.Command{
							{
								Name:       "example",
								ShortUsage: "next seller add example",
								ShortHelp:  "Displays an example seller for the correct JSON schema",
								Exec: func(_ context.Context, args []string) error {
									example := seller{
										Name: "amazon",
									}

									jsonBytes, err := json.MarshalIndent(example, "", "\t")
									if err != nil {
										log.Fatal("Failed to marshal seller struct")
									}

									fmt.Println("Examaple JSON schema to add a new seller:")
									fmt.Println(string(jsonBytes))
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
								log.Fatal("Provide the seller ID of the seller you wish to remove\nFor a list of sellers, use next sellers")
							}

							removeSeller(rpcClient, env, args[0])
							return nil
						},
					},
				},
			},
			{
				Name:       "shader",
				ShortUsage: "next shader <buyer ID>",
				ShortHelp:  "Manage shaders",
				Exec: func(_ context.Context, args []string) error {
					if len(args) == 0 {
						log.Fatal("No buyer ID provided.\nUsage:\nnext shader <buyer ID>\nbuyer ID: the buyer's ID\nFor a list of buyers, use next buyers")
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
								log.Fatal("No buyer ID provided.\nUsage:\nnext shader set <buyer ID> [filepath]\nbuyer ID: the buyer's ID\n(Optional) filepath: the filepath to a JSON file with the new route shader data. If this data is piped through stdin, this parameter is optional.\nFor a list of buyers, use next buyers")
							}

							jsonData := readJSONData("buyers", args[1:])

							// Unmarshal the JSON and create the RoutingRuleSettings struct
							var rrs routing.RoutingRulesSettings
							if err := json.Unmarshal(jsonData, &rrs); err != nil {
								log.Fatalf("Could not unmarshal route shader: %v", err)
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
										log.Fatal("Failed to marshal route shader struct")
									}

									fmt.Println("Exmaple JSON schema to set a new route shader:")
									fmt.Println(string(jsonBytes))
									return nil
								},
							},
						},
					},
				},
			},
			{
				Name:       "ssh",
				ShortUsage: "next ssh <relay name>",
				ShortHelp:  "SSH into a relay by name",
				Exec: func(ctx context.Context, args []string) error {
					if len(args) == 0 {
						log.Fatal("You need to supply a device identifer")
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
			{
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
			},
			{
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
							log.Fatalln(fmt.Errorf("could not parse 1st argument to number: %w", err))
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
			},
			{
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
			},
			{
				Name:       "debug",
				ShortUsage: "next debug <relay_name> [input_file]",
				ShortHelp:  "Debug tool",
				Exec: func(ctx context.Context, args []string) error {
					if len(args) == 0 {
						log.Fatal("You need to supply a relay name")
					}
					relayName := args[0]
					inputFile := "optimize.bin"
					if len(args) > 1 {
						inputFile = args[1]
					}
					debug(relayName, inputFile)
					return nil
				},
			},
		},
		Exec: func(context.Context, []string) error {
			fmt.Printf("Network Next Operator Tool\n\n")
			return flag.ErrHelp
		},
	}

	fmt.Printf("\n")

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Printf("\n")
		log.Fatal(err)
	}

	fmt.Printf("\n")
}
