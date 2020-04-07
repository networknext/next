package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func SSHInto(env Environment, rpcClient jsonrpc.RPCClient, relayName string) {
	var ok bool
	var keyfile string

	if keyfile, ok = os.LookupEnv("RELAY_SERVER_KEY"); !ok {
		log.Fatal("RELAY_SERVER_KEY env var not set")
	}

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

	relay := &reply.Relays[0]

	con := NewSSHConn(relay.SSHUser, relay.ManagementAddr, relay.SSHPort, keyfile)

	con.Connect()
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

func (con SSHConn) Connect() {
	ssh, err := exec.LookPath("ssh")
	if err != nil {
		log.Fatalf("error: could not find ssh")
	}
	args := make([]string, 6)
	args[0] = "ssh"
	args[1] = "-i"
	args[2] = con.keyfile
	args[3] = "-p"
	args[4] = fmt.Sprintf("%d", con.port)
	args[5] = fmt.Sprintf("%s@%s", con.user, con.address)
	env := os.Environ()
	err = syscall.Exec(ssh, args, env)
	if err != nil {
		log.Fatalf("error: failed to exec ssh")
	}
}
