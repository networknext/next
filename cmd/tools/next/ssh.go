package main

import (
	"fmt"
	"log"
	"os"

	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

const (
	DisableRelayScript = `if [ ! $(id -u) = 0 ]; then sudo systemctl stop relay; else systemctl stop relay; fi`
)

type relayInfo struct {
	id         uint64
	user       string
	address    string
	port       string
	publicAddr string
	publicKey  string
	privateKey string
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
		user:       relay.SSHUser,
		address:    relay.ManagementAddr,
		port:       fmt.Sprintf("%d", relay.SSHPort),
		publicAddr: relay.Addr,
	}
}

func testForSSHKey(env Environment) {
	if env.SSHKeyFilePath == "" {
		log.Fatalf("The ssh key file name is not set, set it with 'next ssh key <path>'")
	}

	if _, err := os.Stat(env.SSHKeyFilePath); err != nil {
		log.Fatalf("The ssh key file '%s' does not exist, set it with 'next ssh key <path>'", env.SSHKeyFilePath)
	}
}

func SSHInto(env Environment, rpcClient jsonrpc.RPCClient, relayName string) {
	info := getRelayInfo(rpcClient, relayName)
	testForSSHKey(env)
	con := NewSSHConn(info.user, info.address, info.port, env.SSHKeyFilePath)
	con.Connect()
}

func Disable(env Environment, rpcClient jsonrpc.RPCClient, relayName string) {
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
	con := NewSSHConn(info.user, info.address, info.port, env.SSHKeyFilePath)
	con.ConnectAndIssueCmd(DisableRelayScript)
}

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
	args := make([]string, 4)
	args[0] = "-i"
	args[1] = con.keyfile
	args[2] = "-p"
	args[3] = con.port
	return args
}

func (con SSHConn) Connect() {
	args := con.commonSSHCommands()
	args = append(args, "-tt", con.user+"@"+con.address)
	if !runCommandEnv("ssh", args, nil) {
		log.Fatalf("could not start ssh session")
	}
}

func (con SSHConn) ConnectAndIssueCmd(cmd string) {
	args := con.commonSSHCommands()
	args = append(args, "-tt", con.user+"@"+con.address, "--", cmd)
	if !runCommandEnv("ssh", args, nil) {
		log.Fatalf("could not start ssh session")
	}
}
