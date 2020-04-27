package main

import (
	"log"
	"os"

	"github.com/ybbus/jsonrpc"
)

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
	con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
	con.Connect()
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
