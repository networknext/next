package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

const (
	DisableRelayScript = `if [ ! $(id -u) = 0 ]; then sudo systemctl stop relay; else systemctl stop relay; fi`
)

func run(cmd string, args []string, env []string) {
	if err := syscall.Exec(cmd, args, env); err != nil {
		log.Fatalf("failed to exec %s: %v", cmd, err)
	}
}

func findRelay(rpcClient jsonrpc.RPCClient, relayName string) *localjsonrpc.Relay {
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
	return &relay
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
	relay := findRelay(rpcClient, relayName)
	testForSSHKey(env)
	con := NewSSHConn(relay.SSHUser, relay.ManagementAddr, relay.SSHPort, env.SSHKeyFilePath)
	con.Connect()
}

func Disable(env Environment, rpcClient jsonrpc.RPCClient, relayName string) {
	relay := findRelay(rpcClient, relayName)
	fmt.Printf("Disabling relay '%s' (id = %d)\n", relay.Name, relay.ID)
	testForSSHKey(env)
	relay.State = routing.RelayStateDisabled
	args := localjsonrpc.RelayStateUpdateArgs{
		Relay: *relay,
	}
	var reply localjsonrpc.RelayStateUpdateReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayStateUpdate", &args); err != nil {
		log.Fatalf("could not update relay state: %v", err)
	}
	con := NewSSHConn(relay.SSHUser, relay.ManagementAddr, relay.SSHPort, env.SSHKeyFilePath)
	con.ConnectAndIssueCmd(DisableRelayScript)
}

type SSHConn struct {
	user    string
	address string
	port    int64
	keyfile string
}

func NewSSHConn(user, address string, port int64, authKeyFilename string) SSHConn {
	return SSHConn{
		user:    user,
		address: address,
		port:    port,
		keyfile: authKeyFilename,
	}
}

func (con SSHConn) commonSSHCommands(prgm string) []string {
	args := make([]string, 5)
	args[0] = prgm
	args[1] = "-i"
	args[2] = con.keyfile
	args[3] = "-p"
	args[4] = fmt.Sprintf("%d", con.port)
	return args
}

func (con SSHConn) Connect() {
	ssh, err := exec.LookPath("ssh")
	if err != nil {
		log.Fatalf("error: could not find ssh")
	}
	args := con.commonSSHCommands(ssh)
	args = append(args, fmt.Sprintf("%s@%s", con.user, con.address))
	env := os.Environ()
	run(ssh, args, env)
}

func (con SSHConn) ConnectAndIssueCmd(cmd string) {
	ssh, err := exec.LookPath("ssh")
	if err != nil {
		log.Fatalf("could not find ssh: %v", err)
	}
	args := con.commonSSHCommands(ssh)
	args = append(args, "-t")
	args = append(args, fmt.Sprintf("%s@%s", con.user, con.address))
	args = append(args, cmd)
	env := os.Environ()
	run(ssh, args, env)
}
