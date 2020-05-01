package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"

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

	sudo systemctl stop relay || exit 1

	echo "Waiting for the relay service to clean shutdown"

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
	id         uint64
	name       string
	user       string
	sshAddr    string
	sshPort    string
	publicAddr string
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
		id:         relay.ID,
		name:       relay.Name,
		user:       relay.SSHUser,
		sshAddr:    relay.ManagementAddr,
		sshPort:    fmt.Sprintf("%d", relay.SSHPort),
		publicAddr: relay.Addr,
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
	makeEnv := func(info relayInfo) {
		publicKey, privateKey, err := box.GenerateKey(rand.Reader)

		publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey[:])
		privateKeyB64 := base64.StdEncoding.EncodeToString(privateKey[:])

		if err != nil {
			log.Fatal("could not generate public private keypair")
		}

		routerPublicKey, err := env.RouterPublicKey()

		if err != nil {
			log.Fatalf("could not get router public key: %v", err)
		}

		backendHostname, err := env.RelayBackendHostname()

		if err != nil {
			log.Fatalf("could not get backend hostname: %v", err)
		}

		envvars := make(map[string]string)
		envvars["RELAY_ADDRESS"] = info.publicAddr
		envvars["RELAY_PUBLIC_KEY"] = publicKeyB64
		envvars["RELAY_PRIVATE_KEY"] = privateKeyB64
		envvars["RELAY_ROUTER_PUBLIC_KEY"] = routerPublicKey
		envvars["RELAY_BACKEND_HOSTNAME"] = backendHostname

		f, err := os.Create("deploy/relay/relay.env")
		defer f.Close()

		for k, v := range envvars {
			f.WriteString(fmt.Sprintf("%s=%s\n", k, v))
		}

		args := localjsonrpc.RelayPublicKeyUpdateArgs{
			RelayID:        info.id,
			RelayPublicKey: publicKeyB64,
		}

		var reply localjsonrpc.RelayStateUpdateReply

		if err := rpcClient.CallFor(&reply, "OpsService.RelayPublicKeyUpdate", &args); err != nil {
			log.Fatalf("could not update relay state: %v", err)
		}
	}

	if !runCommandEnv("make", []string{"build-relay"}, nil) {
		log.Fatal("Failed to build relay")
	}

	for _, relayName := range relayNames {
		fmt.Printf("Updating %s\n", relayName)
		info := getRelayInfo(rpcClient, relayName)
		updateRelayState(rpcClient, info, routing.RelayStateEnabled)
		makeEnv(info)
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
			updateRelayState(rpcClient, info, routing.RelayStateEnabled)
			con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
			con.ConnectAndIssueCmd("./install.sh -r")
		}
	} else {
		for _, relayName := range relayNames {
			info := getRelayInfo(rpcClient, relayName)
			fmt.Printf("Reverting relay '%s' (id = %d)\n", info.name, info.id)
			testForSSHKey(env)
			updateRelayState(rpcClient, info, routing.RelayStateEnabled)
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
		updateRelayState(rpcClient, info, routing.RelayStateEnabled)
		con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
		con.ConnectAndIssueCmd(EnableRelayScript)
	}
}

func disableRelays(env Environment, rpcClient jsonrpc.RPCClient, relayNames []string) {
	for _, relayName := range relayNames {
		info := getRelayInfo(rpcClient, relayName)
		fmt.Printf("Disabling relay '%s' (id = %d)\n", relayName, info.id)
		testForSSHKey(env)
		updateRelayState(rpcClient, info, routing.RelayStateDisabled)
		con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
		con.ConnectAndIssueCmd(DisableRelayScript)
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
