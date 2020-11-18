package main

import (
	"fmt"
	"os"
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

	// todo: parse address and get port
	relayPort := 0
	fmt.Printf("    relay bind port is %d\n", relayPort)

	relayPrivateKeyEnv := os.Getenv("RELAY_PRIVATE_KEY")
	if relayPrivateKeyEnv == "" {
		fmt.Printf("error: RELAY_PRIVATE_KEY is not set\n\n")
		os.Exit(1)
	}

	// todo: parse base64 for relay private key

    fmt.Printf( "    relay private key is \"%s\"\n", relayPrivateKeyEnv );

	relayPublicKeyEnv := os.Getenv("RELAY_PUBLIC_KEY")
	if relayPublicKeyEnv == "" {
		fmt.Printf("error: RELAY_PUBLIC_KEY is not set\n\n")
		os.Exit(1)
	}

	// todo: parse base64 for relay public key

    fmt.Printf( "    relay public key is \"%s\"\n", relayPublicKeyEnv );

	relayRouterPublicKeyEnv := os.Getenv("RELAY_ROUTER_PUBLIC_KEY")
	if relayRouterPublicKeyEnv == "" {
		fmt.Printf("error: RELAY_ROUTER_PUBLIC_KEY is not set\n\n")
		os.Exit(1)
	}

	// todo: parse base64 for relay public key

    fmt.Printf( "    relay router public key is \"%s\"\n", relayRouterPublicKeyEnv );

	relayBackendHostnameEnv := os.Getenv("RELAY_BACKEND_HOSTNAME")
	if relayBackendHostnameEnv == "" {
		fmt.Printf("error: RELAY_BACKEND_HOSTNAME is not set\n\n")
		os.Exit(1)
	}

	fmt.Printf("    relay backend hostname is \"%s\"\n", relayBackendHostnameEnv)

    /*
    const char * backend_hostname = relay_platform_getenv( "RELAY_BACKEND_HOSTNAME" );
    if ( !backend_hostname )
    {
        printf( "\nerror: RELAY_BACKEND_HOSTNAME not set\n\n" );
        return 1;
    }

    printf( "    backend hostname is '%s'\n", backend_hostname );

    if ( relay_initialize() != RELAY_OK )
    {
        printf( "\nerror: failed to initialize relay\n\n" );
        return 1;
    }
    */

	// init the relay

	// loop and update the relay

	fmt.Printf("\n")
}
