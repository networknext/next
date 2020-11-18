package main

import (
	"fmt"
	"os"
	"encoding/base64"
	"github.com/networknext/backend/modules/core"	
)

func main() {

    fmt.Printf( "\nNetwork Next Mock Relay\n");

    fmt.Printf( "\nEnvironment:\n\n" );

	relayAddressEnv := os.Getenv("RELAY_ADDRESS")
	if relayAddressEnv == "" {
		fmt.Printf("error: RELAY_ADDRESS is not set\n\n")
		os.Exit(1)
	}

	fmt.Printf("    relay address is \"%s\"\n", relayAddressEnv)

	relayAddress := core.ParseAddress(relayAddressEnv)
	if relayAddress == nil {
		fmt.Printf("error: failed to parse RELAY_ADDRESS\n\n")
		os.Exit(1)
	}

	relayPort := relayAddress.Port

	fmt.Printf("    relay bind port is %d\n", relayPort)

	relayPrivateKeyEnv := os.Getenv("RELAY_PRIVATE_KEY")
	if relayPrivateKeyEnv == "" {
		fmt.Printf("error: RELAY_PRIVATE_KEY is not set\n\n")
		os.Exit(1)
	}

	relayPrivateKey, err := base64.StdEncoding.DecodeString(relayPrivateKeyEnv)
	if err != nil {
		fmt.Printf("error: could not parse RELAY_PRIVATE_KEY as base64\n\n")
		os.Exit(1)
	}

	_ = relayPrivateKey

    fmt.Printf( "    relay private key is \"%s\"\n", relayPrivateKeyEnv );

	relayPublicKeyEnv := os.Getenv("RELAY_PUBLIC_KEY")
	if relayPublicKeyEnv == "" {
		fmt.Printf("error: RELAY_PUBLIC_KEY is not set\n\n")
		os.Exit(1)
	}

	relayPublicKey, err := base64.StdEncoding.DecodeString(relayPublicKeyEnv)
	if err != nil {
		fmt.Printf("error: could not parse RELAY_PUBLIC_KEY as base64\n\n")
		os.Exit(1)
	}

	_ = relayPublicKey

    fmt.Printf( "    relay public key is \"%s\"\n", relayPublicKeyEnv );

	relayRouterPublicKeyEnv := os.Getenv("RELAY_ROUTER_PUBLIC_KEY")
	if relayRouterPublicKeyEnv == "" {
		fmt.Printf("error: RELAY_ROUTER_PUBLIC_KEY is not set\n\n")
		os.Exit(1)
	}

	relayRouterPublicKey, err := base64.StdEncoding.DecodeString(relayRouterPublicKeyEnv)
	if err != nil {
		fmt.Printf("error: could not parse RELAY_ROUTER_PUBLIC_KEY as base64\n\n")
		os.Exit(1)
	}

	_ = relayRouterPublicKey

    fmt.Printf( "    relay router public key is \"%s\"\n", relayRouterPublicKeyEnv );

	relayBackendHostnameEnv := os.Getenv("RELAY_BACKEND_HOSTNAME")
	if relayBackendHostnameEnv == "" {
		fmt.Printf("error: RELAY_BACKEND_HOSTNAME is not set\n\n")
		os.Exit(1)
	}

	fmt.Printf("    relay backend hostname is \"%s\"\n", relayBackendHostnameEnv)

	// init the relay

	// loop and update the relay

	fmt.Printf("\n")
}
