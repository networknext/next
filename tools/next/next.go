/*
   Network Next. Accelerate and Protect. Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
	"io/ioutil"
	"math"
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

	"github.com/networknext/next/modules/admin"
	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
	db "github.com/networknext/next/modules/database"

	jwt "github.com/dgrijalva/jwt-go"
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		sig := <-c
		fmt.Printf("\n\n")
		if cmd.Process != nil {
			cmd.Process.Signal(sig)
			cmd.Wait()
		}
		os.Exit(1)
	}()

	err := cmd.Run()

	if err != nil {
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

func bash(command string) bool {
	return runCommand("bash", []string{"-c", command})
}

func bashQuiet(command string) string {
	_, output := runCommandQuiet("bash", []string{"-c", command}, false)
	return output
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

// level 0: user error
// level 1: program error
func handleRunTimeError(msg string, level int) {
	fmt.Printf(msg)
	fmt.Println()
	os.Exit(level)
}

var env Environment

func getKeyValue(envFile string, keyName string) string {
	value := bashQuiet(fmt.Sprintf("cat %s | awk -v key=%s -F= '$1 == key { sub(/^[^=]+=/, \"\"); print }'", envFile, keyName))
	if len(value) < 1 {
		return ""
	}
	value = value[:len(value)-1]
	if value[0] == '"' || value[0] == '\'' {
		value = value[1 : len(value)-1]
	}
	return value
}

func main() {

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
		ShortHelp:  "Select environment to use (local|dev|staging|prod)",
		Exec: func(_ context.Context, args []string) error {
			if len(args) == 0 {
				handleRunTimeError(fmt.Sprintln("Provide an environment to switch to (local|dev|staging|prod)"), 0)
			}

			if args[0] == "local" {
				bashQuiet("rm -f database.bin && cp envs/local.bin database.bin")
				bashQuiet("psql -U developer postgres -f ../schemas/sql/destroy.sql")
				bashQuiet("psql -U developer postgres -f ../schemas/sql/create.sql")
				bashQuiet("psql -U developer postgres -f ../schemas/sql/local.sql")
			}

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

			env.Name = args[0]
			env.AdminURL = getKeyValue(envFilePath, "ADMIN_REST_API_URL")
			env.PortalURL = getKeyValue(envFilePath, "PORTAL_REST_API_URL")
			env.DatabaseURL = getKeyValue(envFilePath, "DATABASE_REST_API_URL")
			env.SSHKeyFile = getKeyValue(envFilePath, "SSH_KEY_FILE")
			env.APIPrivateKey = getKeyValue(envFilePath, "API_PRIVATE_KEY")
			env.APIKey = getKeyValue(envFilePath, "API_KEY")
			env.VPNAddress = getKeyValue(envFilePath, "VPN_ADDRESS")
			env.RelayBackendHostname = getKeyValue(envFilePath, "RELAY_BACKEND_HOSTNAME")
			env.RelayBackendPublicKey = getKeyValue(envFilePath, "RELAY_BACKEND_PUBLIC_KEY")
			env.RelayArtifactsBucketName = getKeyValue(envFilePath, "RELAY_ARTIFACTS_BUCKET_NAME")
			env.Write()

			cachedDatabase = nil
			if env.Name != "local" {
				bash("rm -f database.bin")
				getDatabase()
			}

			fmt.Printf("Selected %s environment\n\n", env.Name)

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
			fmt.Printf("\n")
			return nil
		},
	}

	var pingCommand = &ffcli.Command{
		Name:       "ping",
		ShortUsage: "next ping",
		ShortHelp:  "Ping the REST API in the current environment",
		Exec: func(_ context.Context, args []string) error {
			ping()
			return nil
		},
	}

	var databaseCommand = &ffcli.Command{
		Name:       "database",
		ShortUsage: "next database",
		ShortHelp:  "Update local database.bin from the current environment Postgres DB and print it",
		Exec: func(_ context.Context, args []string) error {
			printDatabase()
			return nil
		},
	}

	var commitCommand = &ffcli.Command{
		Name:       "commit",
		ShortUsage: "next commit",
		ShortHelp:  "Commit the local database.bin to the current environment runtime (server and relay backends)",
		Exec: func(_ context.Context, args []string) error {
			commitDatabase()
			return nil
		},
	}

	var apiKeyCommand = &ffcli.Command{
		Name:       "api-key",
		ShortUsage: "next api-key",
		ShortHelp:  "Generates an API key for use with the REST API in the current env",
		Exec: func(_ context.Context, args []string) error {
			token := jwt.New(jwt.SigningMethodHS256)
			claims := token.Claims.(jwt.MapClaims)
			claims["database"] = true
			claims["admin"] = true
			claims["portal"] = true
			privateKey := env.APIPrivateKey
			tokenString, err := token.SignedString([]byte(privateKey))
			if err != nil {
				fmt.Printf("error: could not generate API_KEY: %s", err.Error())
				os.Exit(1)
			}
			fmt.Printf("API_KEY = %s\n\n", tokenString)
			return nil
		},
	}

	var hashCommand = &ffcli.Command{
		Name:       "hash",
		ShortUsage: "next hash (string)",
		ShortHelp:  "Provide the 64-bit FNV-1a hash for the provided string",
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				handleRunTimeError(fmt.Sprintf("Please provide a string"), 0)
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

	var sshCommand = &ffcli.Command{
		Name:       "ssh",
		ShortUsage: "next ssh [regex...]",
		ShortHelp:  "SSH into the specified relay(s)",
		Exec: func(_ context.Context, args []string) error {
			regexes := []string{".*"}
			if len(args) > 0 {
				regexes = args
			}

			ssh(env, regexes)

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

	var setupCommand = &ffcli.Command{
		Name:       "setup",
		ShortUsage: "next setup [regex...]",
		ShortHelp:  "Setup the specified relay(s)",
		Exec: func(_ context.Context, args []string) error {
			regexes := []string{".*"}
			if len(args) > 0 {
				regexes = args
			}

			setupRelays(env, regexes)

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
		ShortHelp:  "Generate keys",
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
			fmt.Printf("Cost matrix from %s saved to %s\n\n", env.Name, output)
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

			fmt.Printf("Generated route matrix %s from %s\n\n", output, input)

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
		pingCommand,
		databaseCommand,
		commitCommand,
		relaysCommand,
		sshCommand,
		logCommand,
		setupCommand,
		startCommand,
		stopCommand,
		loadCommand,
		upgradeCommand,
		rebootCommand,
		keygenCommand,
		keysCommand,
		configCommand,
		costCommand,
		optimizeCommand,
		analyzeCommand,
		apiKeyCommand,
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

func GetJSON(url string, object interface{}) {

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
		req.Header.Set("Authorization", "Bearer "+env.APIKey)
		client := &http.Client{}
		response, err = client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %v", url, err))
	}

	if response.Body == nil {
		panic("nil response body")
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

func GetText(url string) string {

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
		req.Header.Set("Authorization", "Bearer "+env.APIKey)
		client := &http.Client{}
		response, err = client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %v", url, err))
	}

	if response == nil {
		core.Error("no response from %s", url)
		os.Exit(1)
	}

	if response.StatusCode != 200 {
		panic(fmt.Sprintf("got %d response for %s", response.StatusCode, url))
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		panic(fmt.Sprintf("could not read response body for %s: %v", url, err))
	}

	response.Body.Close()

	return string(body)
}

func GetBinary(url string) []byte {

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
		req.Header.Set("Authorization", "Bearer "+env.APIKey)
		client := &http.Client{}
		response, err = client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %v", url, err))
	}

	if response == nil {
		core.Error("no response from %s", url)
		os.Exit(1)
	}

	if response.StatusCode != 200 {
		panic(fmt.Sprintf("got %d response for %s", response.StatusCode, url))
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		panic(fmt.Sprintf("could not read response body for %s: %v", url, err))
	}

	response.Body.Close()

	return body
}

func PostJSON(url string, requestData interface{}, responseData interface{}) error {

	buffer := new(bytes.Buffer)

	json.NewEncoder(buffer).Encode(requestData)

	request, _ := http.NewRequest("PUT", url, buffer)

	request.Header.Set("Authorization", "Bearer "+env.APIKey)

	httpClient := &http.Client{}

	var err error
	var response *http.Response
	for i := 0; i < 5; i++ {
		response, err = httpClient.Do(request)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to read %s: %v", url, err)
	}

	if response == nil {
		return fmt.Errorf("no response from %s", url)
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		return fmt.Errorf("could not read response body for %s: %v", url, err)
	}

	response.Body.Close()

	err = json.Unmarshal([]byte(body), &responseData)
	if err != nil {
		return fmt.Errorf("could not parse json response for %s: %v", url, err)
	}

	return nil
}

// -------------------------------------------------------------------------------------------------------

var cachedDatabase *db.Database

type AdminDatabaseResponse struct {
	Database string `json:"database_base64"`
	Error    string `json:"error"`
}

func getDatabase() *db.Database {

	if cachedDatabase != nil {
		return cachedDatabase
	}

	if env.Name != "local" {
		response := AdminDatabaseResponse{}
		GetJSON(fmt.Sprintf("%s/admin/database", env.AdminURL), &response)
		if response.Error != "" {
			fmt.Printf("%s\n", response.Error)
			os.Exit(1)
		}
		database_binary, err := base64.StdEncoding.DecodeString(response.Database)
		if err != nil {
			fmt.Printf("error: could not decode base64 database string\n")
			os.Exit(1)
		}
		os.WriteFile("database.bin", database_binary, 0644)
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
	fmt.Printf("updating database.bin from Postgres SQL instance\n\n")
	database := getDatabase()
	fmt.Println(database.String())
	fmt.Printf("\n")
}

type AdminCommitRequest struct {
	User     string `json:"user"`
	Database string `json:"database_base64"`
}

type AdminCommitResponse struct {
	Error string `json:"error"`
}

func commitDatabase() {

	bash("rm database.bin")

	getDatabase()

	database, err := db.LoadDatabase("database.bin")
	if err != nil {
		fmt.Printf("error: could not load database.bin")
		os.Exit(1)
	}

	gitUser := bashQuiet("git config user.name")
	gitEmail := bashQuiet("git config user.email")
	gitUser = strings.ReplaceAll(gitUser, "\n", "")
	gitEmail = strings.ReplaceAll(gitEmail, "\n", "")

	database_binary := database.GetBinary()

	database_base64 := base64.StdEncoding.EncodeToString(database_binary)

	var request AdminCommitRequest
	var response AdminCommitResponse

	request.User = fmt.Sprintf("%s <%s>", gitUser, gitEmail)
	request.Database = database_base64

	err = PostJSON(fmt.Sprintf("%s/admin/commit", env.AdminURL), &request, &response)
	if err != nil {
		fmt.Printf("error: could not post JSON to commit database endpoint: %v", err)
		os.Exit(1)
	}

	if response.Error != "" {
		fmt.Printf("error: failed to commit database: %s\n\n", response.Error)
		os.Exit(1)
	}

	fmt.Printf("successfully committed database to %s\n\n", env.Name)
}

type PortalRelayData struct {
	RelayName    string `json:"relay_name"`
	RelayId      string `json:"relay_id"`
	RelayAddress string `json:"relay_address"`
	NumSessions  uint32 `json:"num_sessions"`
	MaxSessions  uint32 `json:"max_sessions"`
	StartTime    string `json:"start_time"`
	RelayFlags   string `json:"relay_flags"`
	RelayVersion string `json:"relay_version"`
	Uptime       string `json:"uptime"`
}

type PortalRelaysResponse struct {
	Relays []PortalRelayData `json:"relays"`
}

type AdminRelaysResponse struct {
	Relays []admin.RelayData `json:"relays"`
	Error  string            `json:"error"`
}

func niceUptime(uptimeString string) string {
	value, _ := strconv.ParseInt(uptimeString, 10, 64)
	if value > 86400 {
		return fmt.Sprintf("%dd", int(math.Floor(float64(value/86400))))
	}
	if value > 3600 {
		return fmt.Sprintf("%dh", int(math.Floor(float64(value/3600))))
	}
	if value > 60 {
		return fmt.Sprintf("%dm", int(math.Floor(float64(value/60))))
	}
	return fmt.Sprintf("%ds", value)
}

func printRelays(env Environment, relayCount int64, alphaSort bool, regexName string) {

	adminRelaysResponse := AdminRelaysResponse{}
	portalRelaysResponse := PortalRelaysResponse{}

	GetJSON(fmt.Sprintf("%s/admin/relays", env.AdminURL), &adminRelaysResponse)
	GetJSON(fmt.Sprintf("%s/portal/relays/0/%d", env.PortalURL, constants.MaxRelays), &portalRelaysResponse)

	type RelayRow struct {
		Name            string
		PublicAddress   string
		InternalAddress string
		InternalGroup   string
		Id              string
		Status          string
		Sessions        int
		Uptime          string
		Version         string
	}

	relayMap := make(map[string]*RelayRow)

	for i := range adminRelaysResponse.Relays {
		relayAddress := fmt.Sprintf("%s:%d", adminRelaysResponse.Relays[i].PublicIP, adminRelaysResponse.Relays[i].PublicPort)
		relay := relayMap[relayAddress]
		if relay == nil {
			relay = &RelayRow{}
			relayMap[relayAddress] = relay
		}
		relay.Name = adminRelaysResponse.Relays[i].RelayName
		relay.Id = fmt.Sprintf("%x", common.HashString(relayAddress))
		relay.PublicAddress = relayAddress
		if adminRelaysResponse.Relays[i].InternalIP != "0.0.0.0" {
			relay.InternalAddress = fmt.Sprintf("%s:%d", adminRelaysResponse.Relays[i].InternalIP, adminRelaysResponse.Relays[i].InternalPort)
		}
		relay.InternalGroup = adminRelaysResponse.Relays[i].InternalGroup
		relay.Status = "offline"
		relay.Sessions = 0
		relay.Version = adminRelaysResponse.Relays[i].Version
	}

	for i := range portalRelaysResponse.Relays {
		relayAddress := portalRelaysResponse.Relays[i].RelayAddress
		relay := relayMap[relayAddress]
		if relay == nil {
			continue
		}
		relayFlags, _ := strconv.ParseUint(portalRelaysResponse.Relays[i].RelayFlags, 16, 64)
		if (relayFlags & constants.RelayFlags_ShuttingDown) != 0 {
			relay.Status = "shutting down"
		} else {
			relay.Status = "online"
		}
		relay.Sessions = int(portalRelaysResponse.Relays[i].NumSessions)
		if portalRelaysResponse.Relays[i].RelayVersion != "" {
			relay.Version = portalRelaysResponse.Relays[i].RelayVersion
		}
		relay.Uptime = niceUptime(portalRelaysResponse.Relays[i].Uptime)
	}

	relays := make([]RelayRow, len(relayMap))
	index := 0
	for _, v := range relayMap {
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

	if relayCount != 0 && int(relayCount) < len(outputRelays) {
		outputRelays = outputRelays[0:relayCount]
	}

	if len(outputRelays) > 0 {
		table.Output(outputRelays)
		fmt.Printf("\n")
	} else {
		fmt.Printf("no relays found\n\n")
	}
}

// ----------------------------------------------------------------

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

func (con SSHConn) Connect() {
	args := con.commonSSHCommands()
	args = append(args, "-tt", con.user+"@"+con.address)
	runCommandEnv("ssh", args, nil)
}

func (con SSHConn) ConnectAndIssueCmd(cmd string) bool {
	args := con.commonSSHCommands()
	args = append(args, "-tt", con.user+"@"+con.address, "--", cmd)
	runCommandEnv("ssh", args, nil)
	return true
}

// ------------------------------------------------------------------------------

const (
	SetupRelayScript = `

# run once only

if [[ -f /etc/relay_setup_completed ]]; then echo "already setup" && exit 0; fi

# make the relay prompt cool

echo making the relay prompt cool

sudo echo "export PS1=\"\[\033[36m\]$RELAY_NAME [$ENVIRONMENT] \[\033[00m\]\w # \"" >> ~/.bashrc
sudo echo "source ~/.bashrc" >> ~/.profile.sh

# download the relay binary and rename it to 'relay'

echo downloading relay binary

rm -f $RELAY_VERSION

wget https://storage.googleapis.com/$RELAY_ARTIFACTS_BUCKET_NAME/$RELAY_VERSION --no-cache

if [ ! $? -eq 0 ]; then
    echo "download relay binary failed"
    exit 1
fi

sudo mv $RELAY_VERSION relay

sudo chmod +x relay

# setup the relay environment file

echo setting up relay environment

sudo cat > relay.env <<- EOM
RELAY_NAME=$RELAY_NAME
RELAY_PUBLIC_ADDRESS=$RELAY_PUBLIC_ADDRESS
RELAY_INTERNAL_ADDRESS=$RELAY_INTERNAL_ADDRESS
RELAY_PUBLIC_KEY=$RELAY_PUBLIC_KEY
RELAY_PRIVATE_KEY=$RELAY_PRIVATE_KEY
RELAY_BACKEND_HOSTNAME=$RELAY_BACKEND_HOSTNAME
RELAY_BACKEND_PUBLIC_KEY=$RELAY_BACKEND_PUBLIC_KEY
EOM

# setup the relay service file

echo setting up relay service file

sudo cat > relay.service <<- EOM
[Unit]
Description=Network Next Relay
ConditionPathExists=/app/relay
After=network.target

[Service]
Type=simple
LimitNOFILE=1024
WorkingDirectory=/app
ExecStart=/app/relay
EnvironmentFile=/app/relay.env
Restart=on-failure
RestartSec=12

[Install]
WantedBy=multi-user.target
EOM

# move everything into the /app dir

echo moving everything into /app

sudo rm -rf /app
sudo mkdir /app
sudo mv relay /app/relay
sudo mv relay.env /app/relay.env
sudo mv relay.service /app/relay.service

# limit maximum journalctl logs to 200MB so we don't run out of disk space

echo limiting max journalctl logs to 200MB

sudo sed -i "s/\(.*SystemMaxUse= *\).*/\SystemMaxUse=200M/" /etc/systemd/journald.conf
sudo systemctl restart systemd-journald

# install the relay service, then start it and watch the logs

echo installing relay service

sudo systemctl enable /app/relay.service

echo starting relay service

sudo systemctl start relay

sudo touch /etc/relay_setup_completed

echo setup completed
`

	StartRelayScript = `sudo systemctl enable /app/relay.service && sudo systemctl start relay`

	StopRelayScript = `sudo systemctl stop relay && sudo systemctl disable relay`

	LoadRelayScript = `sudo systemctl stop relay && sudo journalctl --vacuum-size 10M && rm -rf relay && wget https://storage.googleapis.com/%s/%s -O relay --no-cache && chmod +x relay && ./relay version && sudo mv relay /app/relay && sudo systemctl start relay && exit`

	UpgradeRelayScript = `sudo journalctl --vacuum-size 10M && sudo systemctl stop relay; sudo apt update -y && sudo apt upgrade -y && sudo apt dist-upgrade -y && sudo apt autoremove -y && sudo reboot`

	RebootRelayScript = `sudo reboot`

	ConfigRelayScript = `sudo vi /app/relay.env && exit`
)

func getRelayInfo(env Environment, regex string) []admin.RelayData {

	relaysResponse := AdminRelaysResponse{}

	GetJSON(fmt.Sprintf("%s/admin/relays", env.AdminURL), &relaysResponse)

	relays := make([]admin.RelayData, 0)

	for i := range relaysResponse.Relays {
		matched, _ := regexp.Match(regex, []byte(relaysResponse.Relays[i].RelayName))
		if !matched {
			continue
		}
		relays = append(relays, relaysResponse.Relays[i])
	}

	return relays
}

func ssh(env Environment, regexes []string) {

	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for i := range relays {
			if relays[i].SSH_IP == "0.0.0.0" {
				fmt.Printf("%s does not have an SSH address\n", relays[i].RelayName)
				continue
			}
			fmt.Printf("connecting to %s\n", relays[i].RelayName)
			con := NewSSHConn(relays[i].SSH_User, relays[i].SSH_IP, fmt.Sprintf("%d", relays[i].SSH_Port), env.SSHKeyFile)
			con.Connect()
			break
		}
	}

	fmt.Printf("\n")
}

func config(env Environment, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for i := range relays {
			if relays[i].SSH_IP == "0.0.0.0" {
				fmt.Printf("relay %s does not have an SSH address :(\n", relays[i].RelayName)
				continue
			}
			fmt.Printf("connecting to %s\n", relays[i].RelayName)
			con := NewSSHConn(relays[i].SSH_User, relays[i].SSH_IP, fmt.Sprintf("%d", relays[i].SSH_Port), env.SSHKeyFile)
			if !con.ConnectAndIssueCmd(ConfigRelayScript) {
				continue
			}
			break
		}
	}
}

func setupRelays(env Environment, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for i := range relays {

			if relays[i].SSH_IP == "0.0.0.0" {
				fmt.Printf("relay %s does not have an SSH address :(\n", relays[i].RelayName)
				continue
			}

			fmt.Printf("setting up relay %s\n", relays[i].RelayName)

			con := NewSSHConn(relays[i].SSH_User, relays[i].SSH_IP, fmt.Sprintf("%d", relays[i].SSH_Port), env.SSHKeyFile)

			script := SetupRelayScript

			relayName := relays[i].RelayName
			relayVersion := relays[i].Version
			relayPublicAddress := fmt.Sprintf("%s:%d", relays[i].PublicIP, relays[i].PublicPort)
			relayInternalAddress := fmt.Sprintf("%s:%d", relays[i].InternalIP, relays[i].InternalPort)
			relayPublicKeyBase64 := relays[i].PublicKeyBase64
			relayPrivateKeyBase64 := relays[i].PrivateKeyBase64
			relayBackendHostname := env.RelayBackendHostname
			relayBackendPublicKeyBase64 := env.RelayBackendPublicKey
			vpnAddress := env.VPNAddress
			relayArtifactsBucketName := env.RelayArtifactsBucketName

			environment := env.Name

			script = strings.ReplaceAll(script, "$RELAY_NAME", relayName)
			script = strings.ReplaceAll(script, "$RELAY_VERSION", relayVersion)
			script = strings.ReplaceAll(script, "$RELAY_PUBLIC_ADDRESS", relayPublicAddress)
			script = strings.ReplaceAll(script, "$RELAY_INTERNAL_ADDRESS", relayInternalAddress)
			script = strings.ReplaceAll(script, "$RELAY_PUBLIC_KEY", relayPublicKeyBase64)
			script = strings.ReplaceAll(script, "$RELAY_PRIVATE_KEY", relayPrivateKeyBase64)
			script = strings.ReplaceAll(script, "$RELAY_BACKEND_HOSTNAME", relayBackendHostname)
			script = strings.ReplaceAll(script, "$RELAY_BACKEND_PUBLIC_KEY", relayBackendPublicKeyBase64)
			script = strings.ReplaceAll(script, "$VPN_ADDRESS", vpnAddress)
			script = strings.ReplaceAll(script, "$ENVIRONMENT", environment)
			script = strings.ReplaceAll(script, "$RELAY_ARTIFACTS_BUCKET_NAME", relayArtifactsBucketName)

			con.ConnectAndIssueCmd(script)

			if len(relays) > 1 {
				fmt.Printf("\n----------------------------------------------\n\n")
			}
		}
	}
	fmt.Printf("\n")
}

func startRelays(env Environment, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for i := range relays {
			if relays[i].SSH_IP == "0.0.0.0" {
				fmt.Printf("relay %s does not have an SSH address :(\n", relays[i].RelayName)
				continue
			}
			fmt.Printf("starting relay %s\n", relays[i].RelayName)
			con := NewSSHConn(relays[i].SSH_User, relays[i].SSH_IP, fmt.Sprintf("%d", relays[i].SSH_Port), env.SSHKeyFile)
			con.ConnectAndIssueCmd(StartRelayScript)
		}
	}
}

func stopRelays(env Environment, regexes []string) {
	script := StopRelayScript
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for i := range relays {
			if relays[i].SSH_IP == "0.0.0.0" {
				fmt.Printf("relay %s does not have an SSH address :(\n", relays[i].RelayName)
				continue
			}
			fmt.Printf("stopping relay %s\n", relays[i].RelayName)
			con := NewSSHConn(relays[i].SSH_User, relays[i].SSH_IP, fmt.Sprintf("%d", relays[i].SSH_Port), env.SSHKeyFile)
			con.ConnectAndIssueCmd(script)
		}
	}
}

func upgradeRelays(env Environment, regexes []string) {
	script := UpgradeRelayScript
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for i := range relays {
			if relays[i].SSH_IP == "0.0.0.0" {
				fmt.Printf("relay %s does not have an SSH address :(\n", relays[i].RelayName)
				continue
			}
			fmt.Printf("upgrading relay %s\n", relays[i].RelayName)
			con := NewSSHConn(relays[i].SSH_User, relays[i].SSH_IP, fmt.Sprintf("%d", relays[i].SSH_Port), env.SSHKeyFile)
			con.ConnectAndIssueCmd(script)
		}
	}
}

func rebootRelays(env Environment, regexes []string) {
	script := RebootRelayScript
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for i := range relays {
			if relays[i].SSH_IP == "0.0.0.0" {
				fmt.Printf("relay %s does not have an SSH address :(\n", relays[i].RelayName)
				continue
			}
			fmt.Printf("rebooting relay %s\n", relays[i].RelayName)
			con := NewSSHConn(relays[i].SSH_User, relays[i].SSH_IP, fmt.Sprintf("%d", relays[i].SSH_Port), env.SSHKeyFile)
			con.ConnectAndIssueCmd(script)
		}
	}
}

func loadRelays(env Environment, regexes []string, version string) {
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for i := range relays {
			if relays[i].SSH_IP == "0.0.0.0" {
				fmt.Printf("relay %s does not have an SSH address :(\n", relays[i].RelayName)
				continue
			}
			fmt.Printf("loading %s onto %s\n", version, relays[i].RelayName)
			con := NewSSHConn(relays[i].SSH_User, relays[i].SSH_IP, fmt.Sprintf("%d", relays[i].SSH_Port), env.SSHKeyFile)
			con.ConnectAndIssueCmd(fmt.Sprintf(LoadRelayScript, env.RelayArtifactsBucketName, version))
		}
	}
}

func relayLog(env Environment, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		for i := range relays {
			if relays[i].SSH_IP == "0.0.0.0" {
				fmt.Printf("Relay %s does not have an SSH address :(\n", relays[i].RelayName)
				continue
			}
			fmt.Printf("connecting to %s\n", relays[i].RelayName)
			con := NewSSHConn(relays[i].SSH_User, relays[i].SSH_IP, fmt.Sprintf("%d", relays[i].SSH_Port), env.SSHKeyFile)
			con.ConnectAndIssueCmd("journalctl -fu relay -n 1000")
			break
		}
	}
}

func keys(env Environment, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		for i := range relays {
			if relays[i].SSH_IP == "0.0.0.0" {
				fmt.Printf("Relay %s does not have an SSH address :(\n", relays[i].RelayName)
				continue
			}
			fmt.Printf("connecting to %s\n", relays[i].RelayName)
			con := NewSSHConn(relays[i].SSH_User, relays[i].SSH_IP, fmt.Sprintf("%d", relays[i].SSH_Port), env.SSHKeyFile)
			con.ConnectAndIssueCmd("sudo cat /app/relay.env | grep _KEY")
			break
		}
	}
}

// --------------------------------------------------------------------------------------------

func keygen() {

	pingKey := crypto.Auth_Key()

	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("error: could not generate relay keypair\n")
		os.Exit(1)
	}

	pingKeyBase64 := base64.StdEncoding.EncodeToString(pingKey[:])
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey[:])
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKey[:])

	fmt.Printf("export PING_KEY=%s\n", pingKeyBase64)
	fmt.Printf("export RELAY_PUBLIC_KEY=%s\n", publicKeyBase64)
	fmt.Printf("export RELAY_PRIVATE_KEY=%s\n", privateKeyBase64)

	fmt.Printf("\n")
}

// --------------------------------------------------------------------------------------------

type Environment struct {
	Name                     string `json:"name"`
	AdminURL                 string `json:"admin_url"`
	PortalURL                string `json:"portal_url"`
	DatabaseURL              string `json:"database_url"`
	SSHKeyFile               string `json:"ssh_key_filepath"`
	APIPrivateKey            string `json:"api_private_key"`
	APIKey                   string `json:"api_key"`
	VPNAddress               string `json:"vpn_address"`
	RelayBackendHostname     string `json:"relay_backend_hostname"`
	RelayBackendPublicKey    string `json:"relay_backend_public_key"`
	RelayArtifactsBucketName string `json:"relay_artifacts_bucket_name"`
}

func (e *Environment) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s]\n\n", e.Name))
	sb.WriteString(fmt.Sprintf(" + Admin URL = %s\n", e.AdminURL))
	sb.WriteString(fmt.Sprintf(" + Portal URL = %s\n", e.PortalURL))
	sb.WriteString(fmt.Sprintf(" + Database URL = %s\n", e.DatabaseURL))
	sb.WriteString(fmt.Sprintf(" + SSH Key File = %s\n", e.SSHKeyFile))
	sb.WriteString(fmt.Sprintf(" + API Private Key = %s\n", e.APIPrivateKey))
	sb.WriteString(fmt.Sprintf(" + API Key = %s\n", e.APIKey))
	sb.WriteString(fmt.Sprintf(" + VPN Address = %s\n", e.VPNAddress))
	sb.WriteString(fmt.Sprintf(" + Relay Backend Hostname = %s\n", e.RelayBackendHostname))
	sb.WriteString(fmt.Sprintf(" + Relay Backend Public Key = %s\n", e.RelayBackendPublicKey))
	sb.WriteString(fmt.Sprintf(" + Relay Artifacts Bucket Name = %s\n", e.RelayArtifactsBucketName))
	return sb.String()
}

func (e *Environment) Exists() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}

	envFilePath := path.Join(homeDir, ".next")

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

	envFilePath := path.Join(homeDir, ".next")

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

	envFilePath := path.Join(homeDir, ".next")

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

	envFilePath := path.Join(homeDir, ".next")

	err = os.RemoveAll(envFilePath)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to clean environment %v\n", err), 1)

	}
}

// -------------------------------------------------------------------------------------------

func getCostMatrix(env Environment, fileName string) {
	cost_matrix_binary := GetBinary(fmt.Sprintf("%s/portal/cost_matrix", env.PortalURL))
	os.WriteFile("cost.bin", cost_matrix_binary, 0644)
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

	fmt.Printf("\n")
}

// -------------------------------------------------------------------------------------------

func ping() {
	url := fmt.Sprintf("%s/ping", env.AdminURL)
	text := GetText(url)
	fmt.Printf("%s\n\n", text)
}

// -------------------------------------------------------------------------------------------
