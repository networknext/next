/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/networknext/backend/crypto"
	"golang.org/x/crypto/nacl/box"
)

func main() {
	signaturePublicKey, signaturePrivateKey, err := crypto.GenerateCustomerKeyPair()
	if err != nil {
		log.Fatalln(err)
	}

	encryptionPublicKey, encryptionPrivateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("\nWelcome to Network Next!\n\n")

	fmt.Printf("This is your public key for packet signature verification:\n\n")
	fmt.Printf("    %s\n\n", base64.StdEncoding.EncodeToString(signaturePublicKey))

	fmt.Printf("This is your private key for packet signature verification:\n\n")
	fmt.Printf("    %s\n\n", base64.StdEncoding.EncodeToString(signaturePrivateKey))

	fmt.Printf("This is your public key for route token encryption:\n\n")
	fmt.Printf("    %s\n\n", base64.StdEncoding.EncodeToString(encryptionPublicKey[:]))

	fmt.Printf("This is your private key for route token encryption:\n\n")
	fmt.Printf("    %s\n\n", base64.StdEncoding.EncodeToString(encryptionPrivateKey[:]))

	fmt.Printf("IMPORTANT: Save your private keys in a secure place and don't share them with anybody, not even us!\n\n")
}
