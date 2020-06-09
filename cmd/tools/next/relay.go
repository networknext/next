package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

	EnableRelayScript = `
	if systemctl is-active --quiet relay; then
		echo 'Relay service is already running'
		exit
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

func updateRelays(env Environment, rpcClient jsonrpc.RPCClient, relayNames []string) {
	// Fetch and save the latest binary
	const relayTarFile = "dist/relay.tar.gz"

	if url, err := env.RelayArtifactURL(); err == nil {
		if r, err := http.Get(url); err == nil {
			defer r.Body.Close()
			if file, err := os.Create(relayTarFile); err == nil {
				defer file.Close()
				if _, err := io.Copy(file, r.Body); err != nil {
					log.Fatalf("failed to copy http response to file: %v\n", err)
				}
			} else {
				log.Fatalf("could not open "+relayTarFile+" for writing: %v\n", err)
			}
		} else {
			log.Fatalf("could not acquire relay tar: %v\n", err)
		}
	} else {
		log.Fatalf("%v\n", err)
	}

	runCommand("tar", []string{"-C", "./dist", "-xzf", relayTarFile})

	for _, relayName := range relayNames {
		fmt.Printf("Updating %s\n", relayName)
		info := getRelayInfo(rpcClient, relayName)

		// Use the update key that exists on the relay already so that the relay
		// can configure with the old backend correctly.
		// Once we drop the old backend we can remove this block of code.
		{
			// Retrieve the update key that exists on the relay
			success, output := runCommandQuiet("deploy/relay/retrieve-update-key.sh", []string{env.SSHKeyFilePath, info.user + "@" + info.sshAddr}, true)
			if !success {
				// Relay does not have the update key on it yet, so just make it blank for now
				// This won't work if RELAY_V3_ENABLED=1 but it will make spinning up relays in dev
				// faster for now

				info.updateKey = ""

				// log.Fatalf("could not execute the retrieve-update-key.sh script: %s", output)
			}

			// Make sure the update key env var on the relay wasn't empty
			if len(output) == 0 {
				info.updateKey = ""

				// log.Fatalln("no update key found on relay")
			}

			// Remove extra newline and assign to relay info
			info.updateKey = output[:len(output)-1]
		}

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

			const envFile = "dist/relay.env"
			f, err := os.Create(envFile)
			if err != nil {
				log.Fatalf("could not create "+envFile+" file: %v", err)
			}
			defer f.Close()

			for k, v := range envvars {
				if _, err := f.WriteString(fmt.Sprintf("%s=%s\n", k, v)); err != nil {
					log.Fatalf("could not write %s=%s to "+envFile+": %v", k, v, err)
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
		time.Sleep(10 * time.Second)

		// Run the relay update script
		if !runCommandEnv("deploy/relay-update.sh", []string{env.SSHKeyFilePath, info.user + "@" + info.sshAddr}, nil) {
			log.Fatal("could not execute the relay-update.sh script")
		}
	}
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
}
