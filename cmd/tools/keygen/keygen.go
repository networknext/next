/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"

	"golang.org/x/crypto/nacl/box"
)

func main() {
	signaturePublicKey, signaturePrivateKey, err := generateCustomerKeyPair()
	if err != nil {
		log.Fatalln(err)
	}

	encryptionPublicKey, encryptionPrivateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("\nWelcome to Network Next!\n\n")

	fmt.Printf("This is your public key for packet signature verification:\n\n")
	fmt.Printf("    base64: %s\n", base64.StdEncoding.EncodeToString(signaturePublicKey))
	fmt.Printf("    byte data: %v\n\n", signaturePublicKey)

	fmt.Printf("This is your private key for packet signature verification:\n\n")
	fmt.Printf("    base64: %s\n", base64.StdEncoding.EncodeToString(signaturePrivateKey))
	fmt.Printf("    byte data: %v\n\n", signaturePrivateKey)

	fmt.Printf("This is your public key for token encryption:\n\n")
	fmt.Printf("    base64: %s\n", base64.StdEncoding.EncodeToString(encryptionPublicKey[:]))
	fmt.Printf("    byte data: %v\n\n", encryptionPublicKey[:])

	fmt.Printf("This is your private key for token encryption:\n\n")
	fmt.Printf("    base64: %s\n", base64.StdEncoding.EncodeToString(encryptionPrivateKey[:]))
	fmt.Printf("    byte data: %v\n\n", encryptionPrivateKey[:])

	fmt.Printf("IMPORTANT: Save your private keys in a secure place and don't share them with anybody, not even us!\n\n")
}

// generateRelayKeyPair creates a public and private keypair using crypto/ed25519 and prepends a random 8 byte customer ID
// This is copied from crypto to avoid libsodium dependency.
func generateCustomerKeyPair() ([]byte, []byte, error) {
	customerID := make([]byte, 8)
	rand.Read(customerID)

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, err
	}

	customerPublicKey := make([]byte, 0)
	customerPublicKey = append(customerPublicKey, customerID...)
	customerPublicKey = append(customerPublicKey, publicKey...)
	customerPrivateKey := make([]byte, 0)
	customerPrivateKey = append(customerPrivateKey, customerID...)
	customerPrivateKey = append(customerPrivateKey, privateKey...)

	return customerPublicKey, customerPrivateKey, nil
}
