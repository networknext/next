package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"os"
)

func keygen() {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("error: could not generate relay keypair\n")
		os.Exit(1)
	}
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey[:])
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKey[:])
	fmt.Printf("export RELAY_PUBLIC_KEY=%s\n", publicKeyBase64)
	fmt.Printf("export RELAY_PRIVATE_KEY=%s\n", privateKeyBase64)
}
