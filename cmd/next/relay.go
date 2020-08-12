package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
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
	LatestRelayVersion   = "1.0.4"
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
			updateKey:   r.UpdateKey,
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

type updateOptions struct {
	coreCount uint64
	force     bool // force an update regardless of relay version
	hard      bool // hard update the relay, don't clean shutdown
}

func updateRelays(env Environment, rpcClient jsonrpc.RPCClient, regexes []string, opts updateOptions) {
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

	doAllEnabled := false

	if regexes == nil {
		doAllEnabled = true
		regexes = []string{".*"}
	}

	updatedRelays := 0
	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, env, regex)

		if len(relays) == 0 {
			log.Printf("no relays matched the regex '%s'\n", regex)
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

			if !disableRelays(env, rpcClient, []string{relay.name}, opts.hard) {
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
			log.Printf("no relays matched the regex '%s'\n", regex)
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
			log.Printf("no relays matched the regex '%s'\n", regex)
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

func disableRelays(env Environment, rpcClient jsonrpc.RPCClient, regexes []string, hard bool) bool {
	success := true
	relaysDisabled := 0
	testForSSHKey(env)
	script := DisableRelayScript
	if hard {
		script = DisableRelayScriptHard
	}
	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, env, regex)
		if len(relays) == 0 {
			log.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			fmt.Printf("Disabling relay '%s' (id = %016x)\n", relay.name, relay.id)
			con := NewSSHConn(relay.user, relay.sshAddr, relay.sshPort, env.SSHKeyFilePath)
			if !con.ConnectAndIssueCmd(script) || !updateRelayState(rpcClient, relay, routing.RelayStateDisabled) {
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
		fmt.Printf("%s disabled\n", str)
	}

	return success
}

func setRelayNIC(rpcClient jsonrpc.RPCClient, env Environment, relayName string, nicSpeed uint64) {
	relays := getRelayInfo(rpcClient, env, relayName)

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

func setRelayState(rpcClient jsonrpc.RPCClient, env Environment, stateString string, regexes []string) {
	state, err := routing.ParseRelayState(stateString)
	if err != nil {
		log.Fatal(err)
	}

	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, env, regex)

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
	formatFunc := func(bytes uint64) string {
		const (
			kilo = 1000
			mega = 1000000
			giga = 1000000000
		)

		if bytes > giga {
			return fmt.Sprintf("%.02fGB", float64(bytes)/float64(giga))
		}

		if bytes > mega {
			return fmt.Sprintf("%.02fMB", float64(bytes)/float64(mega))
		}

		if bytes > kilo {
			return fmt.Sprintf("%.02fKB", float64(bytes)/float64(kilo))
		}

		return fmt.Sprintf("%d", bytes)
	}

	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	type trafficStats struct {
		Name       string `table:"Name"`
		InternalTx string `table:"Pings Rx"`
		InternalRx string `table:"Pings Tx"`
		GameTx     string `table:"Game Rx"`
		GameRx     string `table:"Game Tx"`
		UnknownRx  string `table:"Unknown Rx"`
	}

	statsList := []trafficStats{}

	for i := range reply.Relays {
		relay := &reply.Relays[i]

		statsList = append(statsList, trafficStats{
			Name:       relay.Name,
			InternalRx: formatFunc(relay.TrafficStats.InternalStatsRx()),
			InternalTx: formatFunc(relay.TrafficStats.InternalStatsTx()),
			GameRx:     formatFunc(relay.TrafficStats.GameStatsRx()),
			GameTx:     formatFunc(relay.TrafficStats.GameStatsTx()),
			UnknownRx:  formatFunc(relay.TrafficStats.UnknownRx),
		})
	}

	table.Output(statsList)
}
