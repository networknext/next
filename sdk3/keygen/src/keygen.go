/*
   Network Next Key Generator
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

func main() {

	customerId := make([]byte, 8)
	rand.Read(customerId)

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalln(err)
	}

	customerPublicKey := make([]byte, 0)
	customerPublicKey = append(customerPublicKey, customerId...)
	customerPublicKey = append(customerPublicKey, publicKey...)

	customerPrivateKey := make([]byte, 0)
	customerPrivateKey = append(customerPrivateKey, customerId...)
	customerPrivateKey = append(customerPrivateKey, privateKey...)

	fmt.Printf("\nWelcome to Network Next!\n\n")
	fmt.Printf("This is your public key:\n\n    %s\n\n", base64.StdEncoding.EncodeToString(customerPublicKey[:]))
	fmt.Printf("This is your private key:\n\n    %s\n\n", base64.StdEncoding.EncodeToString(customerPrivateKey[:]))
	fmt.Printf("IMPORTANT: Save your private key in a secure place and don't share it with anybody, not even us!\n\n")
}
