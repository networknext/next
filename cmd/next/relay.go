package main

import (
	"crypto/rand"
	"encoding/base64"
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
	if ! systemctl is-active --quiet relay; then
		echo 'Relay service has already been stopped'
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
	// If the relay is already running, it will clean shut down before re-enabling.
	EnableRelayScript = `
	if systemctl is-active --quiet relay; then
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

	PortCheckScript = `echo "$(sudo lsof -i -P -n | grep '*:40000' | tr -s ' ' | cut -d ' ' -f 1 | head -1)"`
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

func getRelayInfo(rpcClient jsonrpc.RPCClient, relayName string) relayInfo {
	args := localjsonrpc.RelaysArgs{
		Name: relayName,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		log.Fatal(err)
	}

	if len(reply.Relays) == 0 {
		log.Fatalf("could not find relay with name '%s'", relayName)
	}

	relay := reply.Relays[0]
	return relayInfo{
		id:          relay.ID,
		name:        relay.Name,
		user:        relay.SSHUser,
		sshAddr:     relay.ManagementAddr,
		sshPort:     fmt.Sprintf("%d", relay.SSHPort),
		publicAddr:  relay.Addr,
		publicKey:   relay.PublicKey,
		updateKey:   relay.UpdateKey,
		nicSpeed:    fmt.Sprintf("%d", relay.NICSpeedMbps),
		firestoreID: relay.FirestoreID,
	}
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

func updateRelayState(rpcClient jsonrpc.RPCClient, info relayInfo, state routing.RelayState) {
	args := localjsonrpc.RelayStateUpdateArgs{
		RelayID:    info.id,
		RelayState: state,
	}
	var reply localjsonrpc.RelayStateUpdateReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayStateUpdate", &args); err != nil {
		log.Fatalf("could not update relay state: %v", err)
	}
}

func updateRelays(env Environment, rpcClient jsonrpc.RPCClient, relayNames []string, coreCount uint64) {
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

	for _, relayName := range relayNames {
		fmt.Printf("Updating %s\n", relayName)
		info := getRelayInfo(rpcClient, relayName)

		con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
		out, err := con.IssueCmdAndGetOutput(VersionCheckScript + ` | awk 'BEGIN {FS="."}{print $1}'`)
		if err != nil {
			fmt.Printf("error when acquiring ubuntu version for relay '%s': %v\n", info.name, err)
			continue
		}

		if val, err := strconv.ParseUint(out, 10, 32); err == nil {
			if val < MinimumUbuntuVersion {
				fmt.Printf("%s's ubuntu version is too low, please upgrade to 18.04 or greater: %d\n", info.name, val)
				continue
			}
		} else {
			fmt.Printf("error when parsing ubuntu version for relay '%s': Version = '%s', error = %v\n", info.name, out, err)
			continue
		}

		disableRelays(env, rpcClient, []string{info.name})

		// Update the relay's state to offline in storage
		updateRelayState(rpcClient, info, routing.RelayStateOffline)

		// Create the public and private keys for the relay
		publicKey, privateKey, err := box.GenerateKey(rand.Reader)

		if err != nil {
			log.Fatal("could not generate public private keypair")
		}

		publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey[:])
		privateKeyB64 := base64.StdEncoding.EncodeToString(privateKey[:])

		// Create the environment
		{
			routerPublicKey, err := env.RouterPublicKey()

			if err != nil {
				log.Fatalf("could not get router public key: %v", err)
			}

			backendURL, err := env.RelayBackendURL()

			if err != nil {
				log.Fatalf("could not get backend hostname: %v", err)
			}

			oldBackendHostname, err := env.OldRelayBackendHostname()

			if err != nil {
				log.Fatalf("could not get old backend hostname: %v", err)
			}

			envvars := make(map[string]string)
			envvars["RELAY_ADDRESS"] = info.publicAddr
			envvars["RELAY_PUBLIC_KEY"] = publicKeyB64
			envvars["RELAY_PRIVATE_KEY"] = privateKeyB64
			envvars["RELAY_ROUTER_PUBLIC_KEY"] = routerPublicKey
			envvars["RELAY_BACKEND_HOSTNAME"] = backendURL
			envvars["RELAY_V3_ENABLED"] = "0"
			envvars["RELAY_V3_BACKEND_HOSTNAME"] = oldBackendHostname
			envvars["RELAY_V3_BACKEND_PORT"] = "40000"
			envvars["RELAY_V3_UPDATE_KEY"] = info.updateKey
			envvars["RELAY_V3_SPEED"] = info.nicSpeed
			envvars["RELAY_V3_NAME"] = info.firestoreID

			if coreCount > 0 {
				envvars["RELAY_MAX_CORES"] = strconv.FormatUint(coreCount, 10)
			}

			f, err := os.Create("dist/relay.env")
			if err != nil {
				log.Fatalf("could not create 'dist/relay.env': %v", err)
			}
			defer f.Close()

			for k, v := range envvars {
				if _, err := f.WriteString(fmt.Sprintf("%s=%s\n", k, v)); err != nil {
					log.Fatalf("could not write %s=%s to 'dist/relay.env': %v", k, v, err)
				}
			}
		}

		// Set the public key in storage
		{
			args := localjsonrpc.RelayPublicKeyUpdateArgs{
				RelayID:        info.id,
				RelayPublicKey: publicKeyB64,
			}

			var reply localjsonrpc.RelayStateUpdateReply

			if err := rpcClient.CallFor(&reply, "OpsService.RelayPublicKeyUpdate", &args); err != nil {
				log.Fatalf("could not update relay public key: %v", err)
			}
		}

		// Give the relay backend enough time to pull down the new public key so that
		// we don't get crypto open failed logs when the relay tries to initialize at first
		fmt.Println("Waiting for backend to sync changes...")
		time.Sleep(11 * time.Second)

		// Run the relay update script
		if !runCommandEnv("deploy/relay-update.sh", []string{env.SSHKeyFilePath, info.user + "@" + info.sshAddr}, nil) {
			log.Fatal("could not execute the relay-update.sh script")
		}

		fmt.Printf("%s finished updating\n", relayName)
	}

	// Give the portal enough time to pull down the new state so that
	// the relay state doesn't appear incorrectly
	fmt.Println("Waiting for portal to sync changes...")
	time.Sleep(11 * time.Second)

	str := "Updates"
	if len(relayNames) == 1 {
		str = "Update"
	}
	fmt.Printf("%s complete\n", str)
}

func revertRelays(env Environment, rpcClient jsonrpc.RPCClient, relayNames []string) {
	if relayNames[0] == "ALL" {
		relays := getInfoForAllRelays(rpcClient)
		for _, info := range relays {
			fmt.Printf("Reverting relay '%s' (id = %d)\n", info.name, info.id)
			testForSSHKey(env)
			updateRelayState(rpcClient, info, routing.RelayStateOffline)
			con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
			con.ConnectAndIssueCmd("./install.sh -r")
		}
	} else {
		for _, relayName := range relayNames {
			info := getRelayInfo(rpcClient, relayName)
			fmt.Printf("Reverting relay '%s' (id = %d)\n", info.name, info.id)
			testForSSHKey(env)
			updateRelayState(rpcClient, info, routing.RelayStateOffline)
			con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
			con.ConnectAndIssueCmd("./install.sh -r")
		}
	}

	// Give the portal enough time to pull down the new state so that
	// the relay state doesn't appear incorrectly
	fmt.Println("Waiting for portal to sync changes...")
	time.Sleep(11 * time.Second)

	str := "Reverts"
	if len(relayNames) == 1 {
		str = "Revert"
	}
	fmt.Printf("%s complete\n", str)
}

func enableRelays(env Environment, rpcClient jsonrpc.RPCClient, relayNames []string) {
	for _, relayName := range relayNames {
		info := getRelayInfo(rpcClient, relayName)
		fmt.Printf("Enabling relay '%s' (id = %d)\n", relayName, info.id)
		testForSSHKey(env)
		updateRelayState(rpcClient, info, routing.RelayStateOffline)
		con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
		con.ConnectAndIssueCmd(EnableRelayScript)
	}

	// Give the portal enough time to pull down the new state so that
	// the relay state doesn't appear incorrectly
	fmt.Println("Waiting for portal to sync changes...")
	time.Sleep(11 * time.Second)

	str := "Relays"
	if len(relayNames) == 1 {
		str = "Relay"
	}
	fmt.Printf("%s enabled\n", str)
}

func disableRelays(env Environment, rpcClient jsonrpc.RPCClient, relayNames []string) {
	for _, relayName := range relayNames {
		info := getRelayInfo(rpcClient, relayName)
		fmt.Printf("Disabling relay '%s' (id = %d)\n", relayName, info.id)
		testForSSHKey(env)
		con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
		con.ConnectAndIssueCmd(DisableRelayScript)
		updateRelayState(rpcClient, info, routing.RelayStateDisabled)
	}

	// Give the portal enough time to pull down the new state so that
	// the relay state doesn't appear incorrectly
	fmt.Println("Waiting for portal to sync changes...")
	time.Sleep(11 * time.Second)

	str := "Relays"
	if len(relayNames) == 1 {
		str = "Relay"
	}
	fmt.Printf("%s disabled\n", str)
}

func setRelayNIC(rpcClient jsonrpc.RPCClient, relayName string, nicSpeed uint64) {
	info := getRelayInfo(rpcClient, relayName)

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

func setRelayState(rpcClient jsonrpc.RPCClient, stateString string, relayNames []string) {
	state, err := routing.ParseRelayState(stateString)
	if err != nil {
		log.Fatal(err)
	}

	for _, relayName := range relayNames {
		info := getRelayInfo(rpcClient, relayName)

		updateRelayState(rpcClient, info, state)

		fmt.Printf("Relay state updated for %s to %v\n", info.name, state)
	}
}

func checkRelays(rpcClient jsonrpc.RPCClient, env Environment, filter string) {
	args := localjsonrpc.RelaysArgs{
		Name: filter,
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

	info := make([]checkInfo, len(reply.Relays))

	var wg sync.WaitGroup
	wg.Add(len(reply.Relays))

	fmt.Printf("Checking %d relays\n\n", len(info))

	for i, relay := range reply.Relays {
		r := relay
		go func(indx int, wg *sync.WaitGroup) {
			defer wg.Done()

			infoIndx := &info[indx]
			infoIndx.Name = r.Name

			con := NewSSHConn(r.SSHUser, r.ManagementAddr, strconv.FormatInt(r.SSHPort, 10), env.SSHKeyFilePath)

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
					log.Printf("error when acquiring ubuntu version for relay %s: %v\n", r.Name, err)
					infoIndx.UbuntuVersion = "SSH Error"
				}
			}

			// get logical core count
			{
				if out, err := con.IssueCmdAndGetOutput(CoreCheckScript); err == nil {
					infoIndx.CPUCores = out
				} else {
					log.Printf("error when acquiring number of logical cpu cores for relay %s: %v\n", r.Name, err)
					infoIndx.CPUCores = "SSH Error"
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
						log.Printf("error when checking relay %s can ping the backend: %v\n", r.Name, err)
						infoIndx.CanPingBackend = "SSH Error"
					}
				} else {
					log.Printf("%v\n", err)
				}
			}

			// check if the service is running
			{
				if out, err := con.IssueCmdAndGetOutput("sudo systemctl status relay > /dev/null; echo $?"); err == nil {
					if out == "0" {
						infoIndx.ServiceRunning = "yes"
					} else {
						infoIndx.ServiceRunning = "no"
					}
				} else {
					log.Printf("error when checking if relay %s has the service running: %v\n", r.Name, err)
					infoIndx.ServiceRunning = "SSH Error"
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
					log.Printf("error when checking if relay %s has the right port bound: %v\n", r.Name, err)
					infoIndx.PortBound = "SSH Error"
				}
			}

			log.Printf("gathered info for relay %s\n", r.Name)
		}(i, &wg)
	}

	wg.Wait()

	table.Output(info)
}
