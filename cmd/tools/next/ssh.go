package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func ConnectToRelay(ctx *context.Context, relayName string) {
	var ok bool
	var gcpProjectID string
	var keyfile string

	if _, ok = os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); !ok {
		log.Fatal("GOOGLE_APPLICATION_CREDENTIALS env var not set")
	}

	if gcpProjectID, ok = os.LookupEnv("GOOGLE_PROJECT_ID"); !ok {
		log.Fatal("GOOGLE_PROJECT_ID env var not set")
	}

	if keyfile, ok = os.LookupEnv("RELAY_SERVER_KEY"); !ok {
		log.Fatal("RELAY_SERVER_KEY env var not set")
	}

	firestoreClient, err := firestore.NewClient(*ctx, gcpProjectID)
	if err != nil {
		log.Fatalf("could not initialize firestore client: %v", err)
	}

	rdoc, err := firestoreClient.Collection("Relay").Where("displayName", "==", relayName).Documents(*ctx).Next()

	if err == iterator.Done {
		log.Fatalf("No relays found matching the query '%s'", relayName)
	} else if err != nil {
		log.Fatalf("error when finding relay in firestore: %v", err)
	}

	type relaySSHInfo struct {
		Address string `firestore:"managementAddress"`
		Port    int64  `firestore:"sshPort"`
		User    string `firestore:"sshUser"`
	}

	var info relaySSHInfo
	rdoc.DataTo(&info)

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

func NewSSHConn(ctx *context.Context, user, address string, port int64, keyfile string) SSHConn {
	return SSHConn{
		ctx:     ctx,
		user:    user,
		address: address,
		port:    port,
		keyfile: keyfile,
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
