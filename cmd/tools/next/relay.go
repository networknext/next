package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
	"golang.org/x/crypto/nacl/box"
)

const (
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

		// disable after getting the info, in case the info fails for
		// whatever reason the relays aren't stuck in a weird state
		disableRelays(env, rpcClient, []string{regex})

		for _, info := range relays {
			fmt.Printf("Updating %s\n", info.name)
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

				backendHostname, err := env.RelayBackendHostname()

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
				envvars["RELAY_BACKEND_HOSTNAME"] = backendHostname
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
		for _, info := range relays {
			fmt.Printf("Reverting relay '%s' (id = %d)\n", info.name, info.id)
			testForSSHKey(env)
			updateRelayState(rpcClient, info, routing.RelayStateOffline)
			con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
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
		for _, info := range relays {
			fmt.Printf("Enabling relay '%s' (id = %d)\n", info.name, info.id)
			testForSSHKey(env)
			updateRelayState(rpcClient, info, routing.RelayStateOffline)
			con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
			con.ConnectAndIssueCmd(EnableRelayScript)
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

func disableRelays(env Environment, rpcClient jsonrpc.RPCClient, regexes []string) {
	relaysDisabled := 0
	for _, regex := range regexes {
		relays := getRelayInfo(rpcClient, regex)
		if len(relays) == 0 {
			log.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, info := range relays {
			fmt.Printf("Disabling relay '%s' (id = %d)\n", info.name, info.id)
			testForSSHKey(env)
			con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
			con.ConnectAndIssueCmd(DisableRelayScript)
			updateRelayState(rpcClient, info, routing.RelayStateDisabled)
			relaysDisabled++
		}
	}

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

		for _, info := range relays {
			updateRelayState(rpcClient, info, state)
			fmt.Printf("Relay state updated for %s to %v\n", info.name, state)
		}
	}
}
