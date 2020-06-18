package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
	"golang.org/x/crypto/nacl/box"
)

const (
	MinimumUbuntuVersion = 18

	// DisableRelayScript is the bash script used to disable relays
	DisableRelayScript = `
	service="$(sudo systemctl list-unit-files --state=enabled | grep 'relay.service')"
	if [ -z "$service" ]; then
		echo 'Relay service has already been disabled'
		exit
	fi

	echo "Waiting for the relay service to clean shutdown"

	sudo systemctl stop relay || exit 1

	while systemctl is-active --quiet relay; do
		sleep 1
	done

	sudo systemctl disable relay

	echo 'Relay service shutdown'
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

type relayInfo struct {
	id          uint64
	name        string
	user        string
	sshAddr     string
	sshPort     string
	publicAddr  string
	publicKey   string
	updateKey   string
	nicSpeed    string
	firestoreID string
}

func getRelayInfo(rpcClient jsonrpc.RPCClient, regex string) []relayInfo {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		log.Fatal(err)
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
			updateKey:   r.UpdateKey,
			nicSpeed:    fmt.Sprintf("%d", r.NICSpeedMbps),
			firestoreID: r.FirestoreID,
		}
	}

	return relays
}

func getInfoForAllRelays(rpcClient jsonrpc.RPCClient) []relayInfo {
	args := localjsonrpc.RelaysArgs{}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		log.Fatal(err)
	}

	if len(reply.Relays) == 0 {
		log.Fatal("could not find a single relay")
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

func relays(rpcClient jsonrpc.RPCClient, env Environment, regex string, relaysStateShowFlags [6]bool, relaysStateHideFlags [6]bool, relaysDownFlag bool) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	sort.Slice(reply.Relays, func(i int, j int) bool {
		return reply.Relays[i].SessionCount > reply.Relays[j].SessionCount
	})

	relays := []struct {
		Name        string
		Address     string
		State       string
		Sessions    string
		Tx          string
		Rx          string
		Version     string
		LastUpdated string
	}{}

	for _, relay := range reply.Relays {
		relayState, err := routing.ParseRelayState(relay.State)
		if err != nil {
			log.Fatalf("could not parse invalid relay state %s", relay.State)
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

		tx := fmt.Sprintf("%.02fGB", float64(relay.BytesSent)/float64(1000000000))
		if relay.BytesSent < 1000000000 {
			tx = fmt.Sprintf("%.02fMB", float64(relay.BytesSent)/float64(1000000))
		}
		rx := fmt.Sprintf("%.02fGB", float64(relay.BytesReceived)/float64(1000000000))
		if relay.BytesReceived < 1000000000 {
			rx = fmt.Sprintf("%.02fMB", float64(relay.BytesReceived)/float64(1000000))
		}
		lastUpdateDuration := time.Since(relay.LastUpdateTime).Truncate(time.Second)
		lastUpdated := "n/a"
		if relay.State == "enabled" {
			lastUpdated = lastUpdateDuration.String()
		}

		if relaysDownFlag && lastUpdateDuration < 30*time.Second {
			// Relay is still up and shouldn't be included in the final output
			includeRelay = false
		}

		if !includeRelay {
			continue
		}

		address := relay.Addr

		relays = append(relays, struct {
			Name        string
			Address     string
			State       string
			Sessions    string
			Tx          string
			Rx          string
			Version     string
			LastUpdated string
		}{
			Name:        relay.Name,
			Address:     address,
			State:       relay.State,
			Sessions:    fmt.Sprintf("%d", relay.SessionCount),
			Tx:          tx,
			Rx:          rx,
			Version:     relay.Version,
			LastUpdated: lastUpdated,
		})
	}

	table.Output(relays)
}

func addRelay(rpcClient jsonrpc.RPCClient, env Environment, relay routing.Relay) {
	args := localjsonrpc.AddRelayArgs{
		Relay: relay,
	}

	var reply localjsonrpc.AddRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.AddRelay", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Relay \"%s\" added to storage.\n", relay.Name)
}

func removeRelay(rpcClient jsonrpc.RPCClient, env Environment, name string) {
	relays := getRelayInfo(rpcClient, name)

	if len(relays) == 0 {
		log.Fatalf("no relays matched the name '%s'\n", name)
	}

	info := relays[0]

	args := localjsonrpc.RemoveRelayArgs{
		RelayID: info.id,
	}

	var reply localjsonrpc.RemoveRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.RemoveRelay", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Relay \"%s\" decommissioned.\n", name)
}

func updateRelayState(rpcClient jsonrpc.RPCClient, info relayInfo, state routing.RelayState) bool {
	args := localjsonrpc.RelayStateUpdateArgs{
		RelayID:    info.id,
		RelayState: state,
	}
	var reply localjsonrpc.RelayStateUpdateReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayStateUpdate", &args); err != nil {
		fmt.Printf("could not update relay state: %v\n", err)
		return false
	}

	return true
}

func updateRelays(env Environment, rpcClient jsonrpc.RPCClient, regexes []string, coreCount uint64) {
	// Fetch and save the latest binary
	url, err := env.RelayArtifactURL()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	r, err := http.Get(url)
	if err != nil {
		log.Fatalf("could not acquire relay tar: %v\n", err)
	}

	defer r.Body.Close()

	file, err := os.Create("dist/relay.tar.gz")
	if err != nil {
		log.Fatalf("could not open 'dist/relay.tar.gz' for writing: %v\n", err)
	}

	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		log.Fatalf("failed to copy http response to file: %v\n", err)
	}

	if !runCommand("tar", []string{"-C", "./dist", "-xzf", "dist/relay.tar.gz"}) {
		log.Fatalln("failed to untar relay")
	}

	updatedRelays := 0
	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, regex)

		if len(relays) == 0 {
			log.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}

		for _, relay := range relays {
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

			if !disableRelays(env, rpcClient, []string{relay.name}) {
				continue
			}

			// Update the relay's state to offline in storage
			if !updateRelayState(rpcClient, relay, routing.RelayStateOffline) {
				continue
			}

			// Create the public and private keys for the relay
			publicKey, privateKey, err := box.GenerateKey(rand.Reader)

			if err != nil {
				fmt.Println("could not generate public private keypair")
				continue
			}

			publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey[:])
			privateKeyB64 := base64.StdEncoding.EncodeToString(privateKey[:])

			// Create the environment
			{
				routerPublicKey, err := env.RouterPublicKey()

				if err != nil {
					log.Fatalf("could not get router public key: %v\n", err)
				}

				backendURL, err := env.RelayBackendURL()

				if err != nil {
					log.Fatalf("could not get backend url: %v\n", err)
				}

				oldBackendHostname, err := env.OldRelayBackendHostname()

				if err != nil {
					log.Fatalf("could not get old backend hostname: %v\n", err)
				}

				envvars := make(map[string]string)
				envvars["RELAY_ADDRESS"] = relay.publicAddr
				envvars["RELAY_PUBLIC_KEY"] = publicKeyB64
				envvars["RELAY_PRIVATE_KEY"] = privateKeyB64
				envvars["RELAY_ROUTER_PUBLIC_KEY"] = routerPublicKey
				envvars["RELAY_BACKEND_HOSTNAME"] = backendURL
				envvars["RELAY_V3_ENABLED"] = "0"
				envvars["RELAY_V3_BACKEND_HOSTNAME"] = oldBackendHostname
				envvars["RELAY_V3_BACKEND_PORT"] = "40000"
				envvars["RELAY_V3_UPDATE_KEY"] = relay.updateKey
				envvars["RELAY_V3_SPEED"] = relay.nicSpeed
				envvars["RELAY_V3_NAME"] = relay.firestoreID

				if coreCount > 0 {
					envvars["RELAY_MAX_CORES"] = strconv.FormatUint(coreCount, 10)
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

			updatedRelays++
		}

		fmt.Printf("finished updating relays matching '%s'\n", regex)
	}

	// Give the portal enough time to pull down the new state so that
	// the relay state doesn't appear incorrectly
	fmt.Println("Waiting for portal to sync changes...")
	time.Sleep(11 * time.Second)

	str := "Updates"
	if updatedRelays == 1 {
		str = "Update"
	}
	fmt.Printf("%s complete\n", str)
}

func revertRelays(env Environment, rpcClient jsonrpc.RPCClient, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, regex)
		if len(relays) == 0 {
			log.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			fmt.Printf("Reverting relay '%s' (id = %d)\n", relay.name, relay.id)
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
		relays := getRelayInfo(rpcClient, regex)
		if len(relays) == 0 {
			log.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			fmt.Printf("Enabling relay '%s' (id = %d)\n", relay.name, relay.id)
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

	str := "Reverts"
	if enabledRelays == 1 {
		str = "Revert"
	}
	fmt.Printf("%s complete\n", str)
}

func disableRelays(env Environment, rpcClient jsonrpc.RPCClient, regexes []string) bool {
	success := true
	relaysDisabled := 0
	testForSSHKey(env)
	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, regex)
		if len(relays) == 0 {
			log.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			fmt.Printf("Disabling relay '%s' (id = %d)\n", relay.name, relay.id)
			con := NewSSHConn(relay.user, relay.sshAddr, relay.sshPort, env.SSHKeyFilePath)
			if !con.ConnectAndIssueCmd(DisableRelayScript) || !updateRelayState(rpcClient, relay, routing.RelayStateDisabled) {
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
		fmt.Printf("%s enabled\n", str)
	}

	return success
}

func setRelayNIC(rpcClient jsonrpc.RPCClient, relayName string, nicSpeed uint64) {
	relays := getRelayInfo(rpcClient, relayName)

	if len(relays) == 0 {
		log.Fatalf("no relays matched the name '%s'\n", relayName)
	}

	info := relays[0]

	args := localjsonrpc.RelayNICSpeedUpdateArgs{
		RelayID:       info.id,
		RelayNICSpeed: nicSpeed,
	}

	var reply localjsonrpc.RelayNICSpeedUpdateReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayNICSpeedUpdate", args); err != nil {
		log.Fatal(err)
	}

	// Give the portal enough time to pull down the new state so that
	// the relay state doesn't appear incorrectly
	fmt.Println("Waiting for portal to sync changes...")
	time.Sleep(11 * time.Second)

	fmt.Printf("NIC speed set for %s\n", info.name)
}

func setRelayState(rpcClient jsonrpc.RPCClient, stateString string, regexes []string) {
	state, err := routing.ParseRelayState(stateString)
	if err != nil {
		log.Fatal(err)
	}

	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, regex)

		if len(relays) == 0 {
			log.Printf("no relay matched the regex '%s'\n", regex)
			continue
		}

		for _, relay := range relays {
			if !updateRelayState(rpcClient, relay, state) {
				continue
			}
			fmt.Printf("Relay state updated for %s to %v\n", relay.name, state)
		}
	}
}

func viewRelay(rpcClient jsonrpc.RPCClient, env Environment, relayName string) {
	args := localjsonrpc.RelaysArgs{
		Regex: relayName,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	if len(reply.Relays) == 0 {
		log.Fatalf("Could not find relay with pattern %s", relayName)
	}

	if len(reply.Relays) > 1 {
		log.Fatalf("Found more than one relay matching %s", relayName)
	}

	relay := relay{
		Name:                reply.Relays[0].Name,
		Addr:                reply.Relays[0].Addr,
		PublicKey:           reply.Relays[0].PublicKey,
		SellerID:            reply.Relays[0].SellerID,
		DatacenterName:      reply.Relays[0].DatacenterName,
		NicSpeedMbps:        reply.Relays[0].NICSpeedMbps,
		IncludedBandwidthGB: reply.Relays[0].IncludedBandwidthGB,
		ManagementAddr:      reply.Relays[0].ManagementAddr,
		SSHUser:             reply.Relays[0].SSHUser,
		SSHPort:             reply.Relays[0].SSHPort,
		MaxSessions:         reply.Relays[0].MaxSessionCount,
	}

	jsonData, err := json.MarshalIndent(relay, "", "\t")
	if err != nil {
		log.Fatalf("Could not marshal json data for relay: %v", err)
	}

	fmt.Println(string(jsonData))
}

func editRelay(rpcClient jsonrpc.RPCClient, env Environment, relayName string, editRelayData map[string]interface{}) {
	relayInfo := getRelayInfo(rpcClient, relayName)

	if len(relayInfo) == 0 {
		log.Fatalf("Could not find relay with pattern %s", relayName)
	}

	if len(relayInfo) > 1 {
		log.Fatalf("Found more than one relay matching %s", relayName)
	}

	args := localjsonrpc.RelayEditArgs{
		RelayID:   relayInfo[0].id,
		RelayData: editRelayData,
	}

	var reply localjsonrpc.RelayEditReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayEdit", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Relay %s edited\n", relayInfo[0].name)
}

func checkRelays(rpcClient jsonrpc.RPCClient, env Environment, regex string, relaysStateShowFlags [6]bool, relaysStateHideFlags [6]bool, relaysDownFlag bool) {
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
					log.Printf("error when acquiring ubuntu version for relay %s: %v\n", r.name, err)
					infoIndx.UbuntuVersion = "Error"
				}
			}

			// get logical core count
			{
				if out, err := con.IssueCmdAndGetOutput(CoreCheckScript); err == nil {
					infoIndx.CPUCores = out
				} else {
					log.Printf("error when acquiring number of logical cpu cores for relay %s: %v\n", r.name, err)
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
						log.Printf("error when checking relay %s can ping the backend: %v\n", r.name, err)
						infoIndx.CanPingBackend = "Error"
					}
				} else {
					log.Printf("%v\n", err)
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
					log.Printf("error when checking if relay %s has the service running: %v\n", r.name, err)
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
					log.Printf("error when checking if relay %s has the right port bound: %v\n", r.name, err)
					infoIndx.PortBound = "Error"
				}
			}

			log.Printf("gathered info for relay %s\n", r.name)
		}(i, &wg)
	}

	wg.Wait()

	if len(info) > 0 {
		fmt.Printf("\n")
	}

	table.Output(info)
}
