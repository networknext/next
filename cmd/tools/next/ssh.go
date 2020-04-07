package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"syscall"
)

type relaySSHInfo struct {
	User    string `json:"user"`
	Address string `json:"address"`
	Port    int64  `json:"port"`
}

func getSSHInfo(ctx *context.Context, env Environment, relayName string) relaySSHInfo {
	resp, err := http.Get("http://" + env.Hostname + "/ssh_info?relay_name=" + relayName)
	if err != nil {
		log.Fatalf("error querying ssh info: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Fatalf("invalid query")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %v", err)
	}

	var info relaySSHInfo
	err = json.Unmarshal(body, &info)
	if err != nil {
		log.Fatalf("error parsing response: %v\nBody: %s", err, string(body))
	}

	return info
}

func SSHInto(ctx *context.Context, env Environment, relayName string) {
	var ok bool
	var keyfile string

	if keyfile, ok = os.LookupEnv("RELAY_SERVER_KEY"); !ok {
		log.Fatal("RELAY_SERVER_KEY env var not set")
	}

	info := getSSHInfo(ctx, env, relayName)

	con := NewSSHConn(ctx, info.User, info.Address, info.Port, keyfile)

	con.Connect()
}

type SSHConn struct {
	ctx     *context.Context
	user    string
	address string
	port    int64
	keyfile string
}

func NewSSHConn(ctx *context.Context, user, address string, port int64, authKeyFilename string) SSHConn {
	return SSHConn{
		ctx:     ctx,
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
