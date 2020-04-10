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

func disableRelays(env Environment, rpcClient jsonrpc.RPCClient, relayNames []string) {
	for _, relayName := range relayNames {
		info := getRelayInfo(rpcClient, relayName)
		fmt.Printf("Disabling relay '%s' (id = %d)\n", relayName, info.id)
		testForSSHKey(env)
		args := localjsonrpc.RelayStateUpdateArgs{
			RelayID:    info.id,
			RelayState: routing.RelayStateDisabled,
		}
		var reply localjsonrpc.RelayStateUpdateReply
		if err := rpcClient.CallFor(&reply, "OpsService.RelayStateUpdate", &args); err != nil {
			log.Fatalf("could not update relay state: %v", err)
		}
		con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
		con.ConnectAndIssueCmd(DisableRelayScript)
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

	for _, relayName := range relayNames {
		fmt.Printf("Updating %s\n", relayName)
		info := getRelayInfo(rpcClient, relayName)
		makeEnv(info)
		if !runCommandEnv("deploy/relay-update.sh", []string{env.SSHKeyFilePath, info.user + "@" + info.sshAddr}, nil) {
			log.Fatal("could not execute the relay-update.sh script")
		}
	}
}
