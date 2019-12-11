/*
   Network Next. You control the network.
   Copyright © 2017 - 2019 Network Next, Inc. All rights reserved.
*/

package tools

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/text/encoding/unicode"
)

const DockerComposeNetName = "next"

const KeyBytes = 32
const AuthBytes = 16
const MacBytes = 16
const NonceBytes = 24
const AddressBytes = 19
const HeaderBytes = 1 + 8 + 8 + 2 + AuthBytes
const TokenBytes = 8 + 8 + 2 + 1 + (AddressBytes * 2) + KeyBytes
const EncryptedTokenBytes = NonceBytes + TokenBytes + MacBytes
const PublicKeyBytes = 32

type Checksums struct {
	Relays  []byte `json:"relays"`
	Clients []byte `json:"clients"`
	Matcher []byte `json:"matcher"`
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsMac() bool {
	return runtime.GOOS == "darwin"
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

func LocateMsBuildPath() (string, error) {
	if !IsWindows() {
		return "", fmt.Errorf("can't locate MSBuild on non-Windows platform")
	}

	cmd := exec.Command("powershell", "-EncodedCommand", EncodePowershellCommand(".\\tools\\next\\vswhere -latest -requires Microsoft.Component.MSBuild -find MSBuild\\**\\Bin\\MSBuild.exe"))
	filename, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error: unable to locate MSBuild: %s", err)
	}
	return strings.TrimSpace(string(filename)), nil
}

func ParseChecksums(filename string, checksums *Checksums) error {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(jsonData, checksums); err != nil {
		return err
	}
	return nil
}

func Checksum(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

func UpdateConfig(masterConfig *MasterConfigJSON, name string, checksum []byte) ([]byte, error) {
	data, err := GetHttpAuth(masterConfig.Address + "/admin/api/configs/" + name)
	if err != nil {
		return nil, fmt.Errorf("failed to update %s config: %v", name, err)
	}
	return data, nil
}

func SaveFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("couldn't open '%s' for writing: %v", path, err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("couldn't write to '%s': %v", path, err)
	}

	err = file.Sync()
	if err != nil {
		return fmt.Errorf("couldn't commit file '%s' to disk: %v", path, err)
	}
	return nil
}

func CacheConfigs(masterConfig *MasterConfigJSON, env string) error {
	err := os.MkdirAll("configs/.cache", os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	var checksums Checksums
	_ = ParseChecksums("configs/.cache/checksums.json", &checksums)

	checksumData, err := GetHttpAuth(masterConfig.Address + "/admin/api/configs")
	if err != nil {
		return fmt.Errorf("failed to get checksums: %v", err)
	}

	var newChecksums Checksums
	if err = json.Unmarshal(checksumData, &newChecksums); err != nil {
		return fmt.Errorf("failed to parse checksums: %v", err)
	}

	var wait sync.WaitGroup

	var relays []byte
	var relays_err error
	if bytes.Compare(checksums.Relays, newChecksums.Relays) != 0 {
		wait.Add(1)
		go func() {
			relays, relays_err = UpdateConfig(masterConfig, "relays", newChecksums.Relays)
			wait.Done()
		}()
	}

	wait.Wait()

	if relays_err != nil {
		return relays_err
	}

	if relays != nil {
		if err = SaveFile("configs/.cache/relays.json", relays); err != nil {
			return err
		}
	}

	if err = SaveFile("configs/.cache/checksums.json", checksumData); err != nil {
		return err
	}

	return nil
}

func ParseConfigs(path string, env string) *RelayConfigJSON {

	RelayConfig, err := ParseRelayConfig(env, fmt.Sprintf("%s/relays.json", path))
	if err != nil {
		log.Fatalf("error: could not parse relay config: %s", err)
	}

	return RelayConfig
}

type MasterConfigJSON struct {
	Address             string
	ClientMasterAddress string
	RelayAddress        string
	UdpAddress          string
	SSHAddress          string
	SSHUser             string
	SSHPort             int
	Kubernetes          bool
	Production          bool
	Lab                 bool
	UdpV3Address        string
	UdpV3Port           string
}

func ParseMasterConfig(filename string) (*MasterConfigJSON, error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var masterConfig MasterConfigJSON
	if err := json.Unmarshal(jsonData, &masterConfig); err != nil {
		return nil, err
	}
	return &masterConfig, nil
}

const (
	RELAY_ACTIVE      = 0
	RELAY_INACTIVE    = 1
	RELAY_MAINTENANCE = 2
)

type RelayJSON struct {
	Id                string
	State             int
	Name              string
	PublicAddress     string
	ManagementAddress string
	PrivateAddress    string
	SshUser           string
	SshPort           int
	Role              string
	Group             string
	Type              string
	Latitude          float64
	Longitude         float64
	IsEdgeRelay       bool
}

func FindRelay(relayConfig *RelayConfigJSON, name string) *RelayJSON {
	for i := 0; i < len(relayConfig.Relays); i++ {
		relay := &relayConfig.Relays[i]
		if relay.Name == name {
			return relay
		}
	}
	return nil
}

func FindRelayById(relayConfig *RelayConfigJSON, id uint32) *RelayJSON {
	for i := 0; i < len(relayConfig.Relays); i++ {
		relay := &relayConfig.Relays[i]
		if RelayId(relay.Id) == id {
			return relay
		}
	}
	return nil
}

type RelayConfigJSON struct {
	Relays []RelayJSON
}

func ParseRelayConfig(env string, filename string) (*RelayConfigJSON, error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()
	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var relayConfig RelayConfigJSON
	if err := json.Unmarshal(jsonData, &relayConfig); err != nil {
		return nil, err
	}
	for i, relay := range relayConfig.Relays {
		if relay.SshUser == "" {
			relay.SshUser = "root"
		}
		if relay.SshPort == 0 {
			relay.SshPort = 22
		}
		if relay.Name == "" {
			return nil, fmt.Errorf("relay entry missing name")
		}
		if relay.PublicAddress == "" {
			return nil, fmt.Errorf("relay entry missing address")
		}
		if ParseAddress(relay.PublicAddress) == nil {
			return nil, fmt.Errorf("could not parse relay address: %s", relay.PublicAddress)
		}
		if relay.ManagementAddress == "" {
			relay.ManagementAddress = relay.PublicAddress
		}
		if ParseAddress(relay.ManagementAddress) == nil {
			return nil, fmt.Errorf("could not parse relay management address: %s", relay.ManagementAddress)
		}
		if relay.PrivateAddress != "" && ParseAddress(relay.PrivateAddress) == nil {
			return nil, fmt.Errorf("could not parse relay private address: %s", relay.PrivateAddress)
		}
		if relay.Type == "" {
			relay.Type = "cpp"
		}
		relayConfig.Relays[i] = relay
	}
	return &relayConfig, nil
}

func RelayIdToIndex(relayId uint32, relays []uint32) int {
	for i := range relays {
		if relays[i] == relayId {
			return i
		}
	}
	panic(fmt.Sprintf("relay '%x' not found", relayId))
}

var Env string

var MasterConfig *MasterConfigJSON
var RelayConfig *RelayConfigJSON

func ParseAddress(input string) *net.UDPAddr {
	address := &net.UDPAddr{}
	ip_string, port_string, err := net.SplitHostPort(input)
	if err == nil {
		address.IP = net.ParseIP(ip_string)
		address.Port, _ = strconv.Atoi(port_string)
	} else {
		address.IP = net.ParseIP(input)
		address.Port = 0
	}
	if address.IP == nil {
		return nil
	} else {
		return address
	}
}

func ParseAddressFromBase64(input_base64 string) *net.UDPAddr {
	input, err := base64.StdEncoding.DecodeString(input_base64)
	if err != nil {
		return nil
	}
	return ParseAddress(string(input))
}

func RelayId(id string) uint32 {
	hash := fnv.New32a()
	hash.Write([]byte(id))
	return hash.Sum32()
}

func RandomBytes(bytes int) []byte {
	buffer := make([]byte, bytes)
	rand.Read(buffer)
	return buffer
}

func Commit(message string) bool {
	BashQuiet("git add *")
	BashQuiet(fmt.Sprintf("git commit -am \"%s\"", message))
	ok, _ := BashQuiet("git push")
	return ok
}

func BuildGolang(program string, files ...string) (bool, string) {
	// todo: windoows version
	cmd := fmt.Sprintf("go build -o bin/%s %s", program, strings.Join(files, " "))
	return BashQuiet(cmd)
}

func BuildAndRunGolang(program string, args []string, files ...string) bool {
	cmd := fmt.Sprintf("go build -o bin/%s %s", program, strings.Join(files, " "))
	if ok, output := BashQuiet(cmd); ok {
		return Bash(fmt.Sprintf("./bin/%s %s", program, strings.Join(args, " ")))
	} else {
		fmt.Printf("\n%s\n", output)
		return false
	}
}

func BuildAndRunCPP(program string, args []string) bool {
	if IsWindows() {
		msbuild, err := LocateMsBuildPath()
		if err != nil {
			log.Fatalf("%v", err)
		}
		if ok, output := PowershellQuiet(fmt.Sprintf("cd sdk; & \"%s\" /m /p:Configuration=\"Debug Static32\" /p:Platform=Win32 %s.vcxproj", msbuild, program)); ok {
			return Powershell(fmt.Sprintf("./sdk/bin/Static32/Debug/%s %s; exit $LastExitCode", program, strings.Join(args, " ")))
		} else {
			fmt.Printf("\n%s\n", output)
			return false
		}
	} else {
		if ok, output := BashQuiet(fmt.Sprintf("cd sdk && make -j32 %s", program)); ok {
			return Bash(fmt.Sprintf("./sdk/bin/%s %s", program, strings.Join(args, " ")))
		} else {
			fmt.Printf("\n%s\n", output)
			return false
		}
	}
}

func BuildCPP(program string) bool {
	if IsWindows() {
		msbuild, err := LocateMsBuildPath()
		if err != nil {
			log.Fatalf("%v", err)
		}
		ok, output := PowershellQuiet(fmt.Sprintf("cd sdk; & \"%s\" /m /p:Configuration=\"Debug Static32\" /p:Platform=Win32 %s.vcxproj", msbuild, program))
		if !ok {
			fmt.Printf("\n%s\n", output)
		}
		return ok
	} else {
		ok, output := BashQuiet(fmt.Sprintf("cd sdk && make -j32 %s", program))
		if !ok {
			fmt.Printf("\n%s\n", output)
		}
		return ok
	}
}

func RunCommand(command string, args []string) bool {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("RunCommand error: %v\n", err)
		return false
	}

	return true
}

func RunCommandEnv(command string, args []string, env map[string]string) bool {
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
		fmt.Printf("RunCommand error: %v\n", err)
		return false
	}

	return true
}

func RunCommandInteractive(command string, args []string) bool {
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

func RunCommandQuiet(command string, args []string, stdoutOnly bool) (bool, string) {
	cmd := exec.Command(command, args...)

	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return false, ""
	}

	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return false, ""
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

	stderrScanner := bufio.NewScanner(stderrReader)
	wait.Add(1)
	go func() {
		for stderrScanner.Scan() {
			if !stdoutOnly {
				mutex.Lock()
				output += stderrScanner.Text() + "\n"
				mutex.Unlock()
			}
		}
		wait.Done()
	}()

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

func Bash(command string) bool {
	return RunCommand("bash", []string{"-c", command})
}

func BashQuiet(command string) (bool, string) {
	return RunCommandQuiet("bash", []string{"-c", command}, false)
}

func BashQuietStdoutOnly(command string) (bool, string) {
	return RunCommandQuiet("bash", []string{"-c", command}, true)
}

func EncodePowershellCommand(command string) string {
	uni := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	encoded, err := uni.NewEncoder().String(command)
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
		return ""
	}

	return base64.StdEncoding.EncodeToString([]byte(encoded))
}

func Powershell(command string) bool {
	return RunCommand("powershell", []string{"-EncodedCommand", EncodePowershellCommand(command)})
}

func PowershellQuiet(command string) (bool, string) {
	return RunCommandQuiet("powershell", []string{"-EncodedCommand", EncodePowershellCommand(command)}, false)
}

func GetHttp(url string) ([]byte, error) {
	response, err := http.Get(url)
	if response != nil {
		defer response.Body.Close()
		defer io.Copy(ioutil.Discard, response.Body)
	}
	if err != nil {
		return nil, fmt.Errorf("http request failed: %v", err)
	}

	if response.StatusCode == http.StatusOK || response.StatusCode == http.StatusNoContent {
		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read http response: %v", err)
		}
		return responseData, nil
	} else {
		return nil, fmt.Errorf("http status code: %d", response.StatusCode)
	}
}

func GetHttpAuth(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("could not create http request: %v", err)
	}
	request.Header.Set("Authorization", auth.tokens.Access)
	response, err := http.DefaultClient.Do(request)
	if response != nil {
		defer response.Body.Close()
		defer io.Copy(ioutil.Discard, response.Body)
	}
	if err != nil {
		return nil, fmt.Errorf("http request failed: %v", err)
	}

	if response.StatusCode == http.StatusOK || response.StatusCode == http.StatusNoContent {
		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read http response: %v", err)
		}
		return responseData, nil
	} else {
		return nil, fmt.Errorf("http status code: %d", response.StatusCode)
	}
}

func GetHttpRetry(url string, retries int) ([]byte, error) {
	var err error
	for i := 0; i < retries; i += 1 {
		var response []byte
		response, err = GetHttp(url)
		if err == nil {
			return response, nil
		}
		if i < retries-1 {
			time.Sleep(time.Second)
		}
	}
	return nil, fmt.Errorf("http request failed after %d retries: %v", retries, err)
}

func GenerateProtocolBuffers() bool {
	fmt.Printf("generating protocol buffers...\n")
	ok := Bash("./protocol/generate.sh")
	if !ok {
		fmt.Printf("unable to generate protocol buffers\n")
		return false
	}
	return true
}

func CentsToNibblins(cents int64) int64 {
	return cents * 1e9
}

func NibblinsToCents(nibblins int64) int64 {
	return nibblins / 1e9
}

func NibblinsToDollars(nibblins int64) float64 {
	return float64(nibblins) / 1e11
}

func DollarStringToNibblins(str string) (int64, error) {
	if len(str) == 0 {
		return 0, nil
	}
	decimal := strings.Index(str, ".")
	if decimal == -1 {
		decimal = len(str)
	}

	start := 0
	if str[0] == '-' {
		start = 1
	}

	dollars := int64(0)
	if decimal > start {
		var err error
		dollars, err = strconv.ParseInt(str[start:decimal], 10, 64)
		if err != nil {
			return 0, err
		}
	}
	nibblins := int64(0)
	if decimal+1 < len(str) {
		length := len(str) - (decimal + 1)
		if length < 11 {
			length = 11
		}
		for i := 0; i < length; i += 1 {
			if i < 11 {
				nibblins *= 10
			}
			index := decimal + 1 + i
			if index < len(str) {
				char := str[index]
				if char < byte('0') || char > byte('9') {
					return 0, fmt.Errorf("invalid dollar string: %s", str)
				}
				if i < 11 {
					nibblins += int64(char - byte('0'))
				}
			}
		}
	}
	if str[0] == '-' {
		dollars = -dollars
		nibblins = -nibblins
	}
	return (dollars * 1e11) + nibblins, nil
}

func PremakeExecutable() string {
	if IsMac() {
		return "premake5_mac"
	} else {
		return "premake5_linux"
	}
}

func CleanPremake() {
	if IsWindows() {
		if ok, output := PowershellQuiet("cd sdk; .\\premake5 clean"); !ok {
			log.Fatalf("failed to run premake5 clean: %s", output)
		}
	} else {
		if ok, output := BashQuiet(fmt.Sprintf("cd sdk && ./%s clean", PremakeExecutable())); !ok {
			log.Fatalf("failed to run premake clean: %s", output)
		}
	}
}

func RunPremake(options ...string) {

	if IsWindows() {

		// On Windows, we only re-run Premake if the premake5.lua file has changed since
		// the last run.
		stat, err := os.Stat("sdk/premake5.lua")
		if err != nil {
			log.Fatalf("failed to read sdk/premake5.lua modification time\n")
		}
		modifiedTime := stat.ModTime()
		modifiedTimeUnix := modifiedTime.Unix()

		lastRunTime, err := ioutil.ReadFile("sdk/premake5.lua.timestamp")
		var lastRunTimeUnix int64
		if os.IsNotExist(err) {
			// do nothing
		} else if err != nil {
			log.Fatalf("error while reading timestamp: %v\n", err)
		} else {
			re, err := strconv.Atoi(string(lastRunTime))
			if err == nil {
				lastRunTimeUnix = int64(re)
				if lastRunTimeUnix > modifiedTimeUnix {
					log.Printf("premake generated files are up-to-date, skipping premake execution...\n")
					return
				}
			}
		}

		log.Printf("generating solution files with premake (this will take a minute on Windows, and the subsequent build will be slower)...\n")
		if lastRunTimeUnix == 0 {
			log.Printf("this is because premake has not run previously\n")
		} else {
			log.Printf("this is because modified timestamp %d > last run timestamp %d\n", modifiedTimeUnix, lastRunTimeUnix)
		}

		if ok, output := PowershellQuiet("cd sdk; .\\premake5 clean"); !ok {
			log.Fatalf("failed to run premake5 clean: %s", output)
		}

		err = ioutil.WriteFile("sdk/premake5.lua.timestamp", []byte(strconv.Itoa(int(time.Now().Unix()))), 0644)
		if err != nil {
			log.Printf("error while saving timestamp: %v\n", err)
		}
		if ok, output := PowershellQuiet("cd sdk; ./premake5.exe vs2019"); !ok {
			log.Fatalf("failed to run premake5 vs2019: %s\n", output)
		}

		if ok, output := PowershellQuiet(fmt.Sprintf(
			"cd sdk; ./premake5.exe vs2019 %s",
			strings.Join(options, " "),
		)); !ok {
			log.Fatalf("failed to run premake5 vs2019: %s\n", output)
		}

	} else {
		if ok, _ := BashQuiet(fmt.Sprintf(
			"cd sdk && ./%s gmake %s",
			PremakeExecutable(),
			strings.Join(options, " "),
		)); !ok {
			log.Fatalf("failed to run premake5 gmake\n")
		}
	}
}

const HttpTimeoutMs = 2500
const HttpRetryAttempts = 10
const HttpRetrySleepMs = 500

type HttpClientWithTimeoutAndRetry struct {
	client   *http.Client
	attempts int
	sleep    time.Duration
}

func NewHttpClientWithTimeoutAndRetry(timeout time.Duration, retryAttempts int, retrySleep time.Duration) *HttpClientWithTimeoutAndRetry {
	httpClient := &http.Client{
		Timeout: time.Millisecond * timeout,
	}
	h := &HttpClientWithTimeoutAndRetry{
		client:   httpClient,
		attempts: retryAttempts,
		sleep:    retrySleep,
	}
	return h
}

func (h *HttpClientWithTimeoutAndRetry) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	for i := 0; ; i++ {
		resp, err = h.client.Post(url, contentType, body)
		if err == nil {
			return resp, err
		}
		if i >= (h.attempts - 1) {
			break
		}
		time.Sleep(h.sleep)
	}
	return nil, fmt.Errorf("failed after %d attempts, last error: %s", h.attempts, err)
}

func (h *HttpClientWithTimeoutAndRetry) Get(url string) (resp *http.Response, err error) {
	for i := 0; ; i++ {
		resp, err = h.client.Get(url)
		if err == nil {
			return resp, err
		}
		if i >= (h.attempts - 1) {
			break
		}
		time.Sleep(h.sleep)
	}
	return nil, fmt.Errorf("failed after %d attempts, last error: %s", h.attempts, err)
}

func (h *HttpClientWithTimeoutAndRetry) Do(req *http.Request) (resp *http.Response, err error) {
	for i := 0; ; i++ {
		resp, err = h.client.Do(req)
		if err == nil {
			return resp, err
		}
		if i >= (h.attempts - 1) {
			break
		}
		time.Sleep(h.sleep)
	}
	return nil, fmt.Errorf("failed after %d attempts, last error: %s", h.attempts, err)
}

func GenerateAnsibleHost(address string, user string, port int, args ...string) string {
	host := address + " "
	if user != "" {
		host += fmt.Sprintf("ansible_ssh_user=%s ", user)
	}
	if port != 0 {
		host += fmt.Sprintf("ansible_ssh_port=%d ", port)
	}
	host += "ansible_become=yes ansible_become_method=sudo ansible_ssh_extra_args='-o StrictHostKeyChecking=no'"
	for _, arg := range args {
		host += " " + arg
	}
	return host
}

func FindNextDirectory() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	var lastCwd string
	for cwd != lastCwd {
		_, err := os.Stat(filepath.Join(cwd, "next"))
		if err == nil {
			return cwd, nil
		} else if os.IsNotExist(err) {
			lastCwd = cwd
			cwd = filepath.Dir(cwd)
		} else {
			return "", err
		}
	}

	return "", fmt.Errorf("unable to find next directory")
}

func CopyFileFromDockerImage(image string, srcPath string, destPath string) error {
	ok, container := RunCommandQuiet(
		"docker",
		[]string{
			"create",
			image,
		},
		false,
	)
	if !ok {
		return fmt.Errorf("failed to run docker create")
	}

	defer RunCommand(
		"docker",
		[]string{
			"rm",
			strings.TrimSpace(container),
		},
	)

	ok = RunCommand(
		"docker",
		[]string{
			"cp",
			fmt.Sprintf("%s:%s", strings.TrimSpace(container), srcPath),
			destPath,
		},
	)
	if !ok {
		return fmt.Errorf("failed to run docker cp")
	}

	return nil
}

func ExecuteAnsiblePlaybook(inventory string, script string) bool {
	err := ExecuteAnsiblePlaybookWithError(inventory, script, nil)
	return err != nil
}

func ExecuteAnsiblePlaybookWithError(inventory string, script string, extraVars map[string]string) error {
	if inventory == "" || script == "" {
		return fmt.Errorf("inventory or script was empty")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to get current working directory")
	}

	p := filepath.Join(cwd, ".ansible-hosts")

	if err := ioutil.WriteFile(p, []byte(inventory), 0644); err != nil {
		return fmt.Errorf("unable to write to .ansible-hosts")
	}

	extraVarsArgs := ""
	if extraVars != nil {
		for k, v := range extraVars {
			extraVarsArgs = fmt.Sprintf("%s --extra-vars %s=\"%s\"", extraVarsArgs, k, v)
		}
	}

	if IsWindows() || IsMac() {
		var base string
		if IsWindows() {
			base = os.Getenv("USERPROFILE")
		} else {
			base = os.Getenv("HOME")
		}
		sshPath := filepath.Join(base, ".ssh", "gaffer_rsa")

		nextDir, err := FindNextDirectory()
		if err != nil {
			return fmt.Errorf("unable to find next directory")
		}

		relDir, err := filepath.Rel(nextDir, cwd)
		if err != nil {
			return fmt.Errorf("unable to compute relative path to next directory")
		}

		ok := RunCommand(
			"docker",
			[]string{
				"run",
				"--rm",
				"-v",
				fmt.Sprintf("%s:%s", p, "/.ansible-hosts"),
				"-v",
				fmt.Sprintf("%s:%s", sshPath, "/root/.ssh/id_rsa_tmp"),
				"-v",
				fmt.Sprintf("%s:%s", nextDir, "/next"),
				"-e",
				"ANSIBLE_RETRY_FILES_ENABLED=0",
				"-e",
				"ANSIBLE_HOST_KEY_CHECKING=false",
				"-e",
				"ANSIBLE_SSH_RETRIES=3",
				"-e",
				"ANSIBLE_PIPELINING=true",
				"--entrypoint",
				"/bin/bash",
				"gcr.io/network-next-v3-images/devenv:6e977a40-b1b4-9857-19f8-53e8af0c9a5e",
				"-c",
				fmt.Sprintf("cp /root/.ssh/id_rsa_tmp /root/.ssh/id_rsa && cp /root/.ssh/id_rsa_tmp /root/.ssh/gaffer_rsa && chmod 0600 /root/.ssh/*_rsa && cd /next/%s && /usr/bin/ansible-playbook -i /.ansible-hosts -f 50 %s %s", relDir, script, extraVarsArgs),
			},
		)
		if !ok {
			return fmt.Errorf("docker exited with non-zero exit code")
		}
	} else {

		ok := Bash(fmt.Sprintf("bash -c 'ANSIBLE_RETRY_FILES_ENABLED=0 ANSIBLE_SSH_RETRIES=3 ansible-playbook -i %s %s %s'", p, script, extraVarsArgs))
		if !ok {
			return fmt.Errorf("ansible-playbook exited with non-zero exit code")
		}
	}
	return nil
}

func CaptureAnsiblePlaybook(inventory string, script string, extraVars map[string]string) (string, error) {
	if inventory == "" || script == "" {
		return "", fmt.Errorf("inventory or script was empty")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get current working directory")
	}

	p := filepath.Join(cwd, ".ansible-hosts")

	if err := ioutil.WriteFile(p, []byte(inventory), 0644); err != nil {
		return "", fmt.Errorf("unable to write to .ansible-hosts")
	}

	extraVarsArgs := ""
	if extraVars != nil {
		for k, v := range extraVars {
			extraVarsArgs = fmt.Sprintf("%s --extra-vars %s=\"%s\"", extraVarsArgs, k, v)
		}
	}

	if IsWindows() || IsMac() {
		var base string
		if IsWindows() {
			base = os.Getenv("USERPROFILE")
		} else {
			base = os.Getenv("HOME")
		}
		sshPath := filepath.Join(base, ".ssh", "gaffer_rsa")

		nextDir, err := FindNextDirectory()
		if err != nil {
			return "", fmt.Errorf("unable to find next directory")
		}

		relDir, err := filepath.Rel(nextDir, cwd)
		if err != nil {
			return "", fmt.Errorf("unable to compute relative path to next directory")
		}

		ok, output := RunCommandQuiet(
			"docker",
			[]string{
				"run",
				"--rm",
				"-v",
				fmt.Sprintf("%s:%s", p, "/.ansible-hosts"),
				"-v",
				fmt.Sprintf("%s:%s", sshPath, "/root/.ssh/id_rsa_tmp"),
				"-v",
				fmt.Sprintf("%s:%s", nextDir, "/next"),
				"-e",
				"ANSIBLE_RETRY_FILES_ENABLED=0",
				"-e",
				"ANSIBLE_HOST_KEY_CHECKING=false",
				"-e",
				"ANSIBLE_STDOUT_CALLBACK=json",
				"--entrypoint",
				"/bin/bash",
				"gcr.io/network-next-v3-images/devenv:6b7c7541-8f0c-1558-877a-99b32e376bae",
				"-c",
				fmt.Sprintf("cp /root/.ssh/id_rsa_tmp /root/.ssh/id_rsa && cp /root/.ssh/id_rsa_tmp /root/.ssh/gaffer_rsa && chmod 0600 /root/.ssh/*_rsa && cd /next/%s && /usr/bin/ansible-playbook -i /.ansible-hosts %s %s", relDir, script, extraVarsArgs),
			},
			true,
		)
		if !ok {
			return "", fmt.Errorf("docker exited with non-zero exit code")
		}
		return output, nil
	} else {
		ok, output := BashQuiet(fmt.Sprintf("bash -c 'ANSIBLE_RETRY_FILES_ENABLED=0 ANSIBLE_STDOUT_CALLBACK=json ansible-playbook -i %s %s %s'", p, script, extraVarsArgs))
		if !ok {
			return "", fmt.Errorf("ansible-playbook exited with non-zero exit code")
		}
		return output, nil
	}
}

func ExecuteAndCaptureAnsiblePlaybook(inventory string, script string) (bool, string) {
	if inventory == "" || script == "" {
		return false, ""
	}
	if err := ioutil.WriteFile("/tmp/hosts", []byte(inventory), 0644); err != nil {
		return false, ""
	}

	if IsWindows() {
		panic("ExecuteAndCaptureAnsiblePlaybook not supported on Windows yet")
	}

	return BashQuietStdoutOnly(fmt.Sprintf("bash -c 'ANSIBLE_STDOUT_CALLBACK=json ansible-playbook -i /tmp/hosts %s'", script))
}

func BuildAndDeployOnServer(serverBin string, premakeArgs string, address string, user string, port int) {
	if address == "" || user == "" || port == 0 {
		log.Fatalf("error: incorrect args!")
	}
	if _, err := os.Stat(fmt.Sprintf("./docker-compose/%s_build.yaml", serverBin)); os.IsNotExist(err) {
		log.Fatalf("error: no dockerfile for server binary type %s!", serverBin)
	}
	next_home := os.Getenv("NEXT_HOME")
	if next_home == "" {
		next_home, _ = os.Getwd()
	}
	ansible_home := next_home
	if _, err := os.Stat("/.dockerenv"); err == nil {
		ansible_home, _ = os.Getwd()
	}
	BashQuiet(fmt.Sprintf("docker rm $(docker ps -aq --filter name=/%s_build)", serverBin))
	Bash(fmt.Sprintf("rm -rf ./bin/%s && NEXT_HOME=%s NEXT_MASTER=%s PREMAKE_ARGS=\"%s\" docker-compose -f ./docker-compose/%s_build.yaml up --build", serverBin, next_home, MasterConfig.Address, premakeArgs, serverBin))
	host := GenerateAnsibleHost(address, user, port)
	if Bash(fmt.Sprintf("test -e ./bin/%s", serverBin)) {
		ExecuteAnsiblePlaybook(host, fmt.Sprintf("./relay/ansible/deploy_server.yml --extra-vars \"home=%s binary=%s\"", ansible_home, serverBin))
	}
}

func BuildAndDeployOnClient(clientBin string, premakeArgs string, address string, user string, port int) {
	if address == "" || user == "" || port == 0 {
		log.Fatalf("error: incorrect args!")
	}
	if _, err := os.Stat(fmt.Sprintf("./docker-compose/%s_build.yaml", clientBin)); os.IsNotExist(err) {
		log.Fatalf("error: no docker-compose yaml file for client binary type %s!", clientBin)
	}
	next_home := os.Getenv("NEXT_HOME")
	if next_home == "" {
		next_home, _ = os.Getwd()
	}
	ansible_home := next_home
	if _, err := os.Stat("/.dockerenv"); err == nil {
		ansible_home, _ = os.Getwd()
	}
	BashQuiet(fmt.Sprintf("docker rm $(docker ps -aq --filter name=/%s_build)", clientBin))
	Bash(fmt.Sprintf("rm -rf ./bin/%s && NEXT_HOME=%s NEXT_MASTER=%s PREMAKE_ARGS=\"%s\" docker-compose -f ./docker-compose/%s_build.yaml up --build", clientBin, next_home, MasterConfig.Address, premakeArgs, clientBin))
	host := GenerateAnsibleHost(address, user, port)
	if Bash(fmt.Sprintf("test -e ./bin/%s", clientBin)) {
		ExecuteAnsiblePlaybook(host, fmt.Sprintf("./relay/ansible/deploy_client.yml --extra-vars \"home=%s binary=%s\"", ansible_home, clientBin))
	}
}

func OpenBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func SelectEnvironment(e string) {
	Env = e
	fmt.Printf("\nSelected environment: '%s'\n\n", Env)
	ioutil.WriteFile(".env", []byte(Env+"\n"), 0644)
	AuthClear()
}

func UpdateEnvironment() {
	// used to prevent "next local" and "next select local" from asking for permission EVER
	if AuthenticationNotPermitted {
		return
	}

	err := AuthRead()
	if err != nil {
		if AuthTokens().Refresh != "" {
			err = AuthGetToken(context.Background(), Env, "refresh_token", AuthTokens().Refresh)
		}
		if err != nil {
			Auth_main()
		}
	}
	err = CacheConfigs(MasterConfig, Env)
	if err != nil {
		fmt.Printf("error pulling configs: %v\n", err)
		os.Exit(1)
	}
	RelayConfig = ParseConfigs("configs/.cache", Env)
}

func InitEnvironment() {
	envData, err := ioutil.ReadFile(".env")
	if err == nil {
		Env = strings.TrimSpace(string(envData))
	} else {
		SelectEnvironment("local")
	}

	MasterConfig, err = ParseMasterConfig(fmt.Sprintf("configs/%s/master.json", Env))
	if err != nil {
		fmt.Printf("failed to parse master config: %v\n", err)
		os.Exit(1)
	}
}
