package main

import (
	"fmt"
	"os"

	"github.com/ybbus/jsonrpc"
)

func testForSSHKey(env Environment) {
	if env.SSHKeyFilePath == "" {
		handleRunTimeError(fmt.Sprintln("The ssh key file name is not set, set it with 'next ssh key <path>'"), 0)
	}

	if _, err := os.Stat(env.SSHKeyFilePath); err != nil {
		handleRunTimeError(fmt.Sprintf("The ssh key file '%s' does not exist, set it with 'next ssh key <path>'\n", env.SSHKeyFilePath), 0)
	}
}

func SSHInto(env Environment, rpcClient jsonrpc.RPCClient, relayName string) {
	relays := getRelayInfo(rpcClient, env, relayName)
	if len(relays) == 0 {
		handleRunTimeError(fmt.Sprintf("no relays matches the regex '%s'\n", relayName), 0)
	}
	info := relays[0]
	testForSSHKey(env)
	con := NewSSHConn(info.user, info.sshAddr, info.sshPort, env.SSHKeyFilePath)
	fmt.Printf("Connecting to %s\n", relayName)
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
	args := make([]string, 6)
	args[0] = "-i"
	args[1] = con.keyfile
	args[2] = "-p"
	args[3] = con.port
	args[4] = "-o"
	args[5] = "StrictHostKeyChecking=no"
	return args
}

func (con SSHConn) Connect() {
	args := con.commonSSHCommands()
	args = append(args, "-tt", con.user+"@"+con.address)
	if !runCommandEnv("ssh", args, nil) {
		handleRunTimeError(fmt.Sprintln("could not start ssh session"), 1)
	}
}

func (con SSHConn) ConnectAndIssueCmd(cmd string) bool {
	args := con.commonSSHCommands()
	args = append(args, "-tt", con.user+"@"+con.address, "--", cmd)
	if !runCommandEnv("ssh", args, nil) {
		handleRunTimeError(fmt.Sprintln("could not start ssh session"), 0)
	}

	return true
}

func (con SSHConn) IssueCmdAndGetOutput(cmd string) (string, error) {
	args := con.commonSSHCommands()
	args = append(args, "-tt", con.user+"@"+con.address, "--", cmd)
	return runCommandGetOutput("ssh", args, nil)
}

// TestConnect tries to connect for 60 seconds using ssh's connection timeout functionality
func (con SSHConn) TestConnect() bool {
	args := con.commonSSHCommands()
	args = append(args, "-o", "ConnectTimeout=60", "-tt", con.user+"@"+con.address, "--", ":")
	_, err := runCommandGetOutput("ssh", args, nil)
	return err == nil
}
