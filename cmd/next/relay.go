package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/modood/table"
	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
	"golang.org/x/crypto/nacl/box"
	"google.golang.org/api/iterator"
)

const (
	LatestRelayVersion   = "1.1.0"
	MinimumUbuntuVersion = 18

	// DisableRelayScript is the bash script used to disable relays
	DisableRelayScript = `
	service="$(sudo systemctl list-unit-files --state=enabled | grep 'relay.service')"
	if [ -z "$service" ]; then
		echo 'Relay service has already been disabled'
		exit
	fi

	echo "Waiting for the relay service to clean shutdown"

	sudo systemctl disable relay || exit 1

	sudo systemctl stop relay || exit 1

	while systemctl is-active --quiet relay; do
		sleep 1
	done

	echo 'Relay service shutdown'
	`

	DisableRelayScriptHard = `
	service="$(sudo systemctl list-unit-files --state=enabled | grep 'relay.service')"
	if [ -z "$service" ]; then
		echo 'Relay service has already been disabled'
		exit
	fi

	sudo systemctl disable relay || exit 1
	sudo systemctl kill -s SIGKILL relay || exit 1
	sudo systemctl stop relay || exit 1

	echo 'Relay service shutdown hard'
	`

	// EnableRelayScript is the bash script used to enable relays
	// If the relay service is already enabled, it will clean shut down before re-enabling.
	EnableRelayScript = `
	service="$(sudo systemctl list-unit-files --state=enabled | grep 'relay.service')"
	if [ ! -z "$service" ]; then
		echo 'Relay service is already running, cleanly shutting down...'

		echo "Waiting for the relay service to clean shutdown"

		sudo systemctl stop relay || exit 1

		while systemctl is-active --quiet relay; do
			sleep 1
		done

		sudo systemctl disable relay

		echo 'Relay service shutdown'
	fi

	sudo systemctl enable relay || exit 1
	sudo systemctl start relay || exit 1

	echo 'Relay service started'
	`

	VersionCheckScript = `lsb_release -r | grep -Po "([0-9]{2}\.[0-9]{2})"`

	CoreCheckScript = `
		source /app/relay.env > /dev/null 2&>1
		cores="$(nproc)"
		if [ -z "$RELAY_MAX_CORES" ]; then
			echo "$cores/$cores"
		else
			echo "$RELAY_MAX_CORES/$cores"
		fi
	`

	PortCheckScript = `echo "$(sudo lsof -i -P -n 2>/dev/null | grep '*:40000' | tr -s ' ' | cut -d ' ' -f 1 | head -1)"`
)

func unitFormat(bits uint64) string {
	const (
		kilo = 1000
		mega = 1000000
		giga = 1000000000
	)

	if bits > giga {
		return fmt.Sprintf("%.02fGb/s", float64(bits)/float64(giga))
	}

	if bits > mega {
		return fmt.Sprintf("%.02fMb/s", float64(bits)/float64(mega))
	}

	if bits > kilo {
		return fmt.Sprintf("%.02fKb/s", float64(bits)/float64(kilo))
	}

	return fmt.Sprintf("%d", bits)
}

type relayInfo struct {
	id          uint64
	name        string
	user        string
	sshAddr     string
	sshPort     string
	publicAddr  string
	publicKey   string
	nicSpeed    string
	firestoreID string
	state       string
	version     string
}

func getRelayInfo(rpcClient jsonrpc.RPCClient, env Environment, regex string) []relayInfo {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return nil
	}

	relays := make([]relayInfo, len(reply.Relays))

	for i, r := range reply.Relays {
		relays[i] = relayInfo{
			id:          r.ID,
			name:        r.Name,
			user:        r.SSHUser,
			sshAddr:     r.ManagementAddr,
			sshPort:     fmt.Sprintf("%d", r.SSHPort),
			publicAddr:  r.Addr,
			publicKey:   r.PublicKey,
			nicSpeed:    fmt.Sprintf("%d", r.NICSpeedMbps),
			firestoreID: r.FirestoreID,
			state:       r.State,
			version:     r.Version,
		}
	}

	return relays
}

func getInfoForAllRelays(rpcClient jsonrpc.RPCClient, env Environment) []relayInfo {
	args := localjsonrpc.RelaysArgs{}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return nil
	}

	if len(reply.Relays) == 0 {
		handleRunTimeError(fmt.Sprintln("could not find a single relay"), 0)
	}

	relays := make([]relayInfo, len(reply.Relays))

	for i, relay := range reply.Relays {
		relays[i] = relayInfo{
			id:         relay.ID,
			name:       relay.Name,
			user:       relay.SSHUser,
			sshAddr:    relay.ManagementAddr,
			sshPort:    fmt.Sprintf("%d", relay.SSHPort),
			publicAddr: relay.Addr,
		}
	}

	return relays
}

func updateRelayState(rpcClient jsonrpc.RPCClient, info relayInfo, state routing.RelayState) bool {
	args := localjsonrpc.RelayStateUpdateArgs{
		RelayID:    info.id,
		RelayState: state,
	}
	var reply localjsonrpc.RelayStateUpdateReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayStateUpdate", &args); err != nil {
		handleRunTimeError(fmt.Sprintf("could not update relay state: %v\n", err), 1)
	}

	return true
}

type updateOptions struct {
	coreCount uint64
	force     bool // force an update regardless of relay version
	hard      bool // hard update the relay, don't clean shutdown
}

func updateRelays(env Environment, rpcClient jsonrpc.RPCClient, regexes []string, opts updateOptions) {
	// Fetch and save the latest binary
	url, err := env.RelayArtifactURL()
	if err != nil {
		handleRunTimeError(fmt.Sprintf("%v\n", err), 0)
	}

	r, err := http.Get(url)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not acquire relay tar: %v\n", err), 0)
	}

	defer r.Body.Close()

	file, err := os.Create("dist/relay.tar.gz")
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open 'dist/relay.tar.gz' for writing: %v\n", err), 0)
	}

	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		handleRunTimeError(fmt.Sprintf("failed to copy http response to file: %v\n", err), 0)
	}

	if !runCommand("tar", []string{"-C", "./dist", "-xzf", "dist/relay.tar.gz"}) {
		handleRunTimeError(fmt.Sprintln("failed to untar relay"), 1)
	}

	doAllEnabled := false

	if regexes == nil {
		doAllEnabled = true
		regexes = []string{".*"}
	}

	updatedRelays := 0
	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, env, regex)

		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}

		updates := 0
		for _, relay := range relays {
			if doAllEnabled && relay.state != "enabled" {
				continue
			}

			if !opts.force && relay.version == LatestRelayVersion {
				continue
			}

			fmt.Printf("Updating %s\n", relay.name)

			// validate ubuntu version
			{
				con := NewSSHConn(relay.user, relay.sshAddr, relay.sshPort, env.SSHKeyFilePath)
				out, err := con.IssueCmdAndGetOutput(VersionCheckScript + ` | awk 'BEGIN {FS="."}{print $1}'`)
				if err != nil {
					fmt.Printf("error when acquiring ubuntu version for relay '%s': %v\n", relay.name, err)
					continue
				}

				if val, err := strconv.ParseUint(out, 10, 32); err == nil {
					if val < MinimumUbuntuVersion {
						fmt.Printf("%s's ubuntu version is too low (%s), please upgrade to 18.04 or greater\n", relay.name, out)
						continue
					}
				} else {
					fmt.Printf("error when parsing ubuntu version for relay '%s': version = '%s', error = %v\n", relay.name, out, err)
					continue
				}
			}

			if !disableRelays(env, rpcClient, []string{relay.name}, opts.hard, false) {
				continue
			}

			// Update the relay's state to offline in storage
			if !updateRelayState(rpcClient, relay, routing.RelayStateOffline) {
				continue
			}

			var publicKeyB64 string
			var privateKeyB64 string

			// Create the public and private keys for the relay
			if env.Name == "local" {
				// if local, just reuse the ones from the environment
				publicKeyB64 = os.Getenv("RELAY_PUBLIC_KEY")
				privateKeyB64 = os.Getenv("RELAY_PRIVATE_KEY")
			} else {
				publicKey, privateKey, err := box.GenerateKey(rand.Reader)
				if err != nil {
					fmt.Println("could not generate public private keypair")
					continue
				}
				publicKeyB64 = base64.StdEncoding.EncodeToString(publicKey[:])
				privateKeyB64 = base64.StdEncoding.EncodeToString(privateKey[:])
			}

			// Create the environment
			{
				routerPublicKey, err := env.RouterPublicKey()

				if err != nil {
					handleRunTimeError(fmt.Sprintf("could not get router public key: %v\n", err), 0)
				}

				backendURL, err := env.RelayBackendURL()

				if err != nil {
					handleRunTimeError(fmt.Sprintf("could not get backend url: %v\n", err), 0)
				}

				if err != nil {
					handleRunTimeError(fmt.Sprintf("could not get old backend hostname: %v\n", err), 0)
				}

				envvars := make(map[string]string)
				envvars["RELAY_ADDRESS"] = relay.publicAddr
				envvars["RELAY_PUBLIC_KEY"] = publicKeyB64
				envvars["RELAY_PRIVATE_KEY"] = privateKeyB64
				envvars["RELAY_ROUTER_PUBLIC_KEY"] = routerPublicKey
				envvars["RELAY_BACKEND_HOSTNAME"] = backendURL

				if opts.coreCount > 0 {
					envvars["RELAY_MAX_CORES"] = strconv.FormatUint(opts.coreCount, 10)
				}

				f, err := os.Create("dist/relay.env")
				if err != nil {
					fmt.Printf("could not create 'dist/relay.env': %v\n", err)
					continue
				}
				defer f.Close()

				for k, v := range envvars {
					if _, err := f.WriteString(fmt.Sprintf("%s=%s\n", k, v)); err != nil {
						fmt.Printf("could not write %s=%s to 'dist/relay.env': %v\n", k, v, err)
					}
				}
			}

			// Set the public key in storage
			{
				args := localjsonrpc.RelayPublicKeyUpdateArgs{
					RelayID:        relay.id,
					RelayPublicKey: publicKeyB64,
				}

				var reply localjsonrpc.RelayStateUpdateReply

				if err := rpcClient.CallFor(&reply, "OpsService.RelayPublicKeyUpdate", &args); err != nil {
					fmt.Printf("could not update relay public key: %v\n", err)
					continue
				}
			}

			// Give the relay backend enough time to pull down the new public key so that
			// we don't get crypto open failed logs when the relay tries to initialize at first
			fmt.Println("Waiting for backend to sync changes...")
			time.Sleep(11 * time.Second)

			// Run the relay update script
			if !runCommandEnv("deploy/relay-update.sh", []string{env.SSHKeyFilePath, relay.sshPort, relay.user + "@" + relay.sshAddr}, nil) {
				fmt.Println("could not execute the relay-update.sh script")
				continue
			}

			updates++
		}

		if updates > 0 {
			updatedRelays += updates
			fmt.Printf("finished updating relays matching '%s'\n", regex)
		}
	}

	if updatedRelays > 0 {
		// Give the portal enough time to pull down the new state so that
		// the relay state doesn't appear incorrectly
		fmt.Println("Waiting for portal to sync changes...")
		time.Sleep(11 * time.Second)

		str := "Updates"
		if updatedRelays == 1 {
			str = "Update"
		}
		fmt.Printf("%s complete\n", str)
	} else {
		fmt.Println("No relays need to be updated")
	}
}

func revertRelays(env Environment, rpcClient jsonrpc.RPCClient, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			fmt.Printf("Reverting relay '%s' (id = %016x)\n", relay.name, relay.id)
			testForSSHKey(env)
			if !updateRelayState(rpcClient, relay, routing.RelayStateOffline) {
				continue
			}
			con := NewSSHConn(relay.user, relay.sshAddr, relay.sshPort, env.SSHKeyFilePath)
			con.ConnectAndIssueCmd("./install.sh -r")
		}
	}
}

func enableRelays(env Environment, rpcClient jsonrpc.RPCClient, regexes []string) {
	enabledRelays := 0
	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			fmt.Printf("Enabling relay '%s' (id = %016x)\n", relay.name, relay.id)
			testForSSHKey(env)
			if !updateRelayState(rpcClient, relay, routing.RelayStateOffline) {
				continue
			}
			con := NewSSHConn(relay.user, relay.sshAddr, relay.sshPort, env.SSHKeyFilePath)
			if !con.ConnectAndIssueCmd(EnableRelayScript) {
				continue
			}
			enabledRelays++
		}
	}

	// Give the portal enough time to pull down the new state so that
	// the relay state doesn't appear incorrectly
	fmt.Println("Waiting for portal to sync changes...")
	time.Sleep(11 * time.Second)

	str := "Enabling"
	if enabledRelays == 1 {
		str = "Enable"
	}
	fmt.Printf("%s complete\n", str)
}

func disableRelays(env Environment, rpcClient jsonrpc.RPCClient, regexes []string, hard bool, maintenance bool) bool {
	success := true
	relaysDisabled := 0
	testForSSHKey(env)
	script := DisableRelayScript
	if hard {
		script = DisableRelayScriptHard
	}

	relayState := routing.RelayStateDisabled
	infoText := "Disabling"
	successText := "disabled."
	if maintenance {
		relayState = routing.RelayStateMaintenance
		infoText = "Setting maintenance mode on"
		successText = "is now in maintenance mode."
	}

	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			fmt.Printf("%s relay '%s' (id = %016x)\n", infoText, relay.name, relay.id)
			con := NewSSHConn(relay.user, relay.sshAddr, relay.sshPort, env.SSHKeyFilePath)
			if !con.ConnectAndIssueCmd(script) || !updateRelayState(rpcClient, relay, relayState) {
				success = false
				continue
			}
			relaysDisabled++
		}
	}

	if relaysDisabled > 0 {
		// Give the portal enough time to pull down the new state so that
		// the relay state doesn't appear incorrectly
		fmt.Println("Waiting for portal to sync changes...")
		time.Sleep(11 * time.Second)

		str := "Relays"
		if relaysDisabled == 1 {
			str = "Relay"
		}
		fmt.Printf("%s %s\n", str, successText)
	}

	return success
}

// TODO modify to use the OpsService.UpdateRelay endpoint
func updateRelayName(rpcClient jsonrpc.RPCClient, env Environment, oldName string, newName string) {

	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, oldName); !ok {
		// error msg printed by called function
		return
	}

	reply := localjsonrpc.RelayNameUpdateReply{}
	args := localjsonrpc.RelayNameUpdateArgs{
		RelayID:   relayID,
		RelayName: newName,
	}

	if err := rpcClient.CallFor(&reply, "OpsService.RelayNameUpdate", args); err != nil {
		fmt.Printf("error renaming relay: %v\n", (err))
	} else {
		fmt.Printf("Relay renamed successfully: %s -> %s\n", oldName, newName)
	}

}

func checkRelays(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	regex string,
	relaysStateShowFlags [6]bool,
	relaysStateHideFlags [6]bool,
	relaysDownFlag bool,
	csvOutputFlag bool,
) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	sort.Slice(reply.Relays, func(i int, j int) bool {
		return reply.Relays[i].Name < reply.Relays[j].Name
	})

	type checkInfo struct {
		Name           string
		Address        string `table:"Address"`
		CanSSH         string `table:"SSH"`
		UbuntuVersion  string `table:"Ubuntu"`
		CPUCores       string `table:"Cores"`
		CanPingBackend string `table:"Ping Backend"`
		ServiceRunning string `table:"Running"`
		PortBound      string `table:"Bound"`
	}

	// filter the relays
	includedRelays := make([]relayInfo, 0)

	for _, relay := range reply.Relays {
		relayState, err := routing.ParseRelayState(relay.State)
		if err != nil {
			fmt.Printf("could not parse relay state %s for relay %s", relay.State, relay.Name)
			continue
		}

		includeRelay := true

		for i, flag := range relaysStateShowFlags {
			if flag {
				if relayState != routing.RelayState(i) {
					// An "only show" flag is set and this relay doesn't match that state, so don't include it in the final output
					includeRelay = false
				} else {
					// One of the flags should include the relay, so set to true and break out, since combining the flags is an OR operation
					includeRelay = true
					break
				}
			}
		}

		if relaysStateHideFlags[relayState] {
			// Relay should be hidden, so don't include in final output
			includeRelay = false
		}

		lastUpdateDuration := time.Since(relay.LastUpdateTime).Truncate(time.Second)
		if relaysDownFlag && lastUpdateDuration < 30*time.Second {
			// Relay is still up and shouldn't be included in the final output
			includeRelay = false
		}

		if !includeRelay {
			continue
		}

		includedRelays = append(includedRelays, relayInfo{
			name:    relay.Name,
			user:    relay.SSHUser,
			sshAddr: relay.ManagementAddr,
			sshPort: strconv.FormatInt(relay.SSHPort, 10),
		})
	}

	var wg sync.WaitGroup
	wg.Add(len(includedRelays))

	fmt.Printf("Checking %d relays\n\n", len(includedRelays))

	info := make([]checkInfo, len(includedRelays))

	for i, relay := range includedRelays {
		r := relay
		go func(indx int, wg *sync.WaitGroup) {
			defer wg.Done()

			infoIndx := &info[indx]
			infoIndx.Name = r.name
			infoIndx.Address = r.sshAddr

			con := NewSSHConn(r.user, r.sshAddr, r.sshPort, env.SSHKeyFilePath)

			// test ssh capability, if not success return
			if con.TestConnect() {
				infoIndx.CanSSH = "yes"
			} else {
				infoIndx.CanSSH = "no"
				return
			}

			// get ubuntu version
			{
				if out, err := con.IssueCmdAndGetOutput(VersionCheckScript); err == nil {
					infoIndx.UbuntuVersion = out
				} else {
					fmt.Printf("error when acquiring ubuntu version for relay %s: %v\n", r.name, err)
					infoIndx.UbuntuVersion = "Error"
				}
			}

			// get logical core count
			{
				if out, err := con.IssueCmdAndGetOutput(CoreCheckScript); err == nil {
					infoIndx.CPUCores = out
				} else {
					fmt.Printf("error when acquiring number of logical cpu cores for relay %s: %v\n", r.name, err)
					infoIndx.CPUCores = "Error"
				}
			}

			// test ping ability
			{
				if backend, err := env.RelayBackendHostname(); err == nil {
					if out, err := con.IssueCmdAndGetOutput("ping -c 20 " + backend + " > /dev/null; echo $?"); err == nil {
						if out == "0" {
							infoIndx.CanPingBackend = "yes"
						} else {
							infoIndx.CanPingBackend = "no"
						}
					} else {
						fmt.Printf("error when checking relay %s can ping the backend: %v\n", r.name, err)
						infoIndx.CanPingBackend = "Error"
					}
				} else {
					fmt.Printf("%v\n", err)
				}
			}

			// check if the service is running
			{
				if out, err := con.IssueCmdAndGetOutput("sudo systemctl status relay > /dev/null 2>&1; echo $?"); err == nil {
					if out == "0" {
						infoIndx.ServiceRunning = "yes"
					} else {
						infoIndx.ServiceRunning = "no"
					}
				} else {
					fmt.Printf("error when checking if relay %s has the service running: %v\n", r.name, err)
					infoIndx.ServiceRunning = "Error"
				}
			}

			// check if the port is bound
			{
				if out, err := con.IssueCmdAndGetOutput(PortCheckScript); err == nil {
					if out == "relay" {
						infoIndx.PortBound = "yes"
					} else {
						infoIndx.PortBound = "no"
					}
				} else {
					fmt.Printf("error when checking if relay %s has the right port bound: %v\n", r.name, err)
					infoIndx.PortBound = "Error"
				}
			}

			fmt.Printf("gathered info for relay %s\n", r.name)
		}(i, &wg)
	}

	wg.Wait()

	if len(info) > 0 {
		fmt.Printf("\n")
	}

	if csvOutputFlag {
		var csvInfo [][]string
		csvInfo = append(csvInfo, []string{
			"Name", "Address", "SSH", "Ubuntu", "Cores", "Ping Backend", "Running", "Bound"})

		for _, relayInfo := range info {
			csvInfo = append(csvInfo, []string{
				relayInfo.Name,
				relayInfo.Address,
				relayInfo.CanSSH,
				relayInfo.UbuntuVersion,
				relayInfo.CPUCores,
				relayInfo.CanPingBackend,
				relayInfo.ServiceRunning,
				relayInfo.PortBound,
			})

			fileName := "./relay-check.csv"
			f, err := os.Create(fileName)
			if err != nil {
				fmt.Printf("Error creating local CSV file %s: %v\n", fileName, err)
				return
			}

			writer := csv.NewWriter(f)
			err = writer.WriteAll(csvInfo)
			if err != nil {
				fmt.Printf("Error writing local CSV file %s: %v\n", fileName, err)
			}
		}
		fmt.Println("CSV file written: relay-check.csv")

	} else {
		table.Output(info)
	}

}

func relayLogs(rpcClient jsonrpc.RPCClient, env Environment, lines uint, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, env, regex)
		for i, relay := range relays {
			con := NewSSHConn(relay.user, relay.sshAddr, relay.sshPort, env.SSHKeyFilePath)
			if out, err := con.IssueCmdAndGetOutput("journalctl -u relay -n " + strconv.FormatUint(uint64(lines), 10) + " | cat"); err == nil {
				fmt.Printf("%s\n%s\n", relay.name, out)
				if i < len(relays)-1 {
					fmt.Printf("\n")
				}
			} else {
				fmt.Printf("error gathering logs for relay %s: %v\n", relay.name, err)
			}
		}
	}
}

func relayTraffic(rpcClient jsonrpc.RPCClient, env Environment, regex string) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	type trafficStats struct {
		Name      string `table:"Name"`
		PingsTx   string `table:"Pings Tx"`
		PingsRx   string `table:"Pings Rx"`
		GameTx    string `table:"Game Tx"`
		GameRx    string `table:"Game Rx"`
		UnknownRx string `table:"Unknown Rx"`
	}

	statsList := []trafficStats{}

	for i := range reply.Relays {
		relay := &reply.Relays[i]

		statsList = append(statsList, trafficStats{
			Name:      relay.Name,
			PingsTx:   unitFormat(relay.TrafficStats.OtherStatsTx() * 8),
			PingsRx:   unitFormat(relay.TrafficStats.OtherStatsRx() * 8),
			GameTx:    unitFormat(relay.TrafficStats.GameStatsTx() * 8),
			GameRx:    unitFormat(relay.TrafficStats.GameStatsRx() * 8),
			UnknownRx: unitFormat(relay.TrafficStats.UnknownRx * 8),
		})
	}

	table.Output(statsList)
}

func relayHeatmap(rpcClient jsonrpc.RPCClient, env Environment, regex string, filepath string) {
	// Get the relay that we want to generate the connectivity image for
	relayInfo := getRelayInfo(rpcClient, env, regex)

	if len(relayInfo) == 0 {
		handleRunTimeError(fmt.Sprintf("no relay found with regex %q", regex), 0)
	}

	if len(relayInfo) > 1 {
		handleRunTimeError(fmt.Sprintf("more than one relay matching regex %q - please be more specific", regex), 0)
	}

	relay := relayInfo[0]

	relayPingData, err := getRelayPingBigQueryRows(relay.id, env)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to get relay ping data from BigQuery for relay %q: %v", relay.name, err), 1)
	}

	if len(relayPingData) == 0 {
		handleRunTimeError(fmt.Sprintf("no relay ping data for relay %q", regex), 1)
	}

	imageData := convertBigQueryRows(relayPingData)
	if err := generateHeatmapImage(imageData, filepath); err != nil {
		handleRunTimeError(fmt.Sprintf("could not write image to file at %q: %v", filepath, err), 1)
	}
}

func getRelayPingBigQueryRows(relayID uint64, env Environment) ([]BigQueryRelayPingsEntry, error) {

	ctx := context.Background()

	var rows []BigQueryRelayPingsEntry

	var dbName string
	var sql bytes.Buffer

	sql.Write([]byte(`select * from `))

	switch env.Name {
	case "prod":
		dbName = "network-next-v3-prod"
		sql.Write([]byte(fmt.Sprintf("%s.prod.relay_pings", dbName)))

	case "staging":
		dbName = "network-next-v3-staging"
		sql.Write([]byte(fmt.Sprintf("%s.staging.relay_pings", dbName)))

	case "dev":
		dbName = "network-next-v3-dev"
		sql.Write([]byte(fmt.Sprintf("%s.dev.relay_pings", dbName)))

	case "local":
		return nil, errors.New("local env not implemented")

	default:
		return nil, errors.New("unknown or unimplemented env")
	}

	sql.Write([]byte(fmt.Sprintf(" where relay_a = %d", relayID)))
	sql.Write([]byte(" and timestamp > timestamp_sub(current_timestamp, INTERVAL 24 HOUR)"))
	sql.Write([]byte(" order by timestamp"))

	bqClient, err := bigquery.NewClient(ctx, dbName)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("getRelayPingData() failed to create BigQuery client: %v", err), 1)
		return nil, err
	}
	defer bqClient.Close()

	q := bqClient.Query(string(sql.String()))

	job, err := q.Run(ctx)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("getRelayPingData() failed to query BigQuery: %v", err), 1)
		return nil, err
	}

	status, err := job.Wait(ctx)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("getRelayPingData() error waiting for job to complete: %v", err), 1)
		return nil, err
	}
	if err := status.Err(); err != nil {
		handleRunTimeError(fmt.Sprintf("getRelayPingData() job returned an error: %v", err), 1)
		return nil, err
	}

	it, err := job.Read(ctx)
	if err != nil {
		return nil, err
	}

	// process result set and load rows
	for {
		var rec BigQueryRelayPingsEntry
		err := it.Next(&rec)

		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		rows = append(rows, rec)
	}

	return rows, nil
}

type pingData struct {
	relayID    uint64
	rtt        float64
	jitter     float64
	packetLoss float64
	routable   bool
}

func convertBigQueryRows(relayPingData []BigQueryRelayPingsEntry) [][]pingData {
	// We know the the relay ping data from BigQuery will be stored based on timestamp,
	// since that's how we query the data. Based on this, we can construct a data structure
	// by iterating through each timestamp of the slice in order, then sorting each subslice
	// by relay ID

	var partitionedPingData [][]pingData

	// We need to keep track of the timestamp for this iteration so that if it changes,
	// we know we are done with that timestamp
	currTimestamp := relayPingData[0].Timestamp

	var bqRowsIndex int
	var stop bool

	// This is the outer loop that will loop through the entire slice of BigQuery results
	for {
		if stop {
			break
		}

		// Add a new entry for this timestamp
		partitionedPingData = append(partitionedPingData, []pingData{})

		// Add an entry to this subslice for each relay ID (BigQuery row with this timestamp)
		for ; true; bqRowsIndex++ {
			if bqRowsIndex >= len(relayPingData) {
				stop = true
				break
			}

			timestamp := relayPingData[bqRowsIndex].Timestamp

			// The timestamp has changed, so we've looped through all relay IDs for this timestamp
			// We need to sort the subslice, and break from the loop
			if timestamp != currTimestamp {
				currTimestamp = timestamp

				sort.SliceStable(partitionedPingData[len(partitionedPingData)-1], func(i, j int) bool {
					return partitionedPingData[len(partitionedPingData)-1][i].relayID < partitionedPingData[len(partitionedPingData)-1][j].relayID
				})

				break
			}

			// We still have more relay IDs to go, so add this one and continue
			partitionedPingData[len(partitionedPingData)-1] = append(partitionedPingData[len(partitionedPingData)-1], pingData{
				relayID:    uint64(relayPingData[bqRowsIndex].RelayB),
				rtt:        relayPingData[bqRowsIndex].RTT,
				jitter:     relayPingData[bqRowsIndex].Jitter,
				packetLoss: relayPingData[bqRowsIndex].PacketLoss,
				routable:   relayPingData[bqRowsIndex].Routable,
			})
		}
	}

	return partitionedPingData
}

func generateHeatmapImage(partitionedPingData [][]pingData, filepath string) error {
	// Now that we have each list of sorted relays in order indexed by timestamp, we can generate the image

	// width is simply the length of timestamps
	width := len(partitionedPingData)

	// height will be the maximum number of relays we see in any timestamp
	var height int
	for _, relays := range partitionedPingData {
		if len(relays) > height {
			height = len(relays)
		}
	}

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		numRelays := len(partitionedPingData[x])
		for y := 0; y < numRelays; y++ {
			var c color.NRGBA

			// Red dots have high packet loss
			if partitionedPingData[x][y].packetLoss > 0.1 {
				c = color.NRGBA{255, 0, 0, 255}
				img.SetNRGBA(x, y, c)
				continue
			}

			// Blue dots have high packet jitter
			if partitionedPingData[x][y].jitter > 10 {
				c = color.NRGBA{0, 0, 255, 255}
				img.SetNRGBA(x, y, c)
				continue
			}

			// Black dots are otherwise unroutable
			if !partitionedPingData[x][y].routable {
				c = color.NRGBA{0, 0, 0, 255}
				img.SetNRGBA(x, y, c)
				continue
			}

			// White dots are good
			c = color.NRGBA{255, 255, 255, 255}
			img.SetNRGBA(x, y, c)
			continue
		}
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return err
	}

	return nil
}
