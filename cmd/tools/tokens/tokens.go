package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
)

type relayaddrFlags []string

func (i *relayaddrFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *relayaddrFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type relaypubkeyFlags []string

func (i *relaypubkeyFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *relaypubkeyFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	privkey := flag.String("privatekey", "", "the private key to encrypt each node")

	token := flag.String("token", "new", "the type of token (new, continue)")

	expires := flag.Duration("expires", 10*time.Second, "the duration it expires after token is built")
	sessionID := flag.Int64("sessionid", 1, "the SessionID to set in the token")
	sessionversion := flag.Int("sessionversion", 1, "the Session Version to set in the token")

	clientaddr := flag.String("clientaddr", "127.0.0.1:10001", "the client's IP")
	clientkey := flag.String("clientpublickey", "", "the client's public key")

	var relayaddrs relayaddrFlags
	flag.Var(&relayaddrs, "relayaddr", "the relay's IP")

	var relaykeys relaypubkeyFlags
	flag.Var(&relaykeys, "relaypublickey", "the relay's public key")

	serveraddr := flag.String("serveraddr", "127.0.0.1:10002", "the servers's IP")
	serverkey := flag.String("serverpublickey", "", "the servers's public key")
	flag.Parse()

	if *privkey == "" {
		log.Fatal("privatekey is required")
	}
	privatekey, err := base64.StdEncoding.DecodeString(*privkey)
	if err != nil {
		log.Fatal(err)
	}

	client, err := net.ResolveUDPAddr("udp", *clientaddr)
	if err != nil {
		log.Fatal(err)
	}

	if *clientkey == "" {
		log.Println("clientpublickey empty, removing client segment at the end of the process")
	}
	clientpubkey, err := base64.StdEncoding.DecodeString(*clientkey)
	if err != nil {
		log.Fatal(err)
	}

	server, err := net.ResolveUDPAddr("udp", *serveraddr)
	if err != nil {
		log.Fatal(err)
	}

	if *serverkey == "" {
		log.Fatal("serverpublickey is required")
	}
	serverpubkey, err := base64.StdEncoding.DecodeString(*serverkey)
	if err != nil {
		log.Fatal(err)
	}

	if len(relayaddrs) != len(relaykeys) {
		log.Fatal("relayaddr and relaypublickey need to have the same number of occurrences")
	}

	var routeToken routing.Token
	tokenTypeIdentifyingByte := make([]byte, 1)
	switch *token {
	case "new":
		tokenTypeIdentifyingByte[0] = routing.TokenTypeRouteRequest
		nextRouteToken := routing.NextRouteToken{
			Expires: uint64(time.Now().Add(*expires).Unix()),

			SessionId:      uint64(*sessionID),
			SessionVersion: uint8(*sessionversion),

			Client: routing.Client{Addr: *client, PublicKey: clientpubkey},

			Server: routing.Server{Addr: *server, PublicKey: serverpubkey},
		}

		for idx := range relayaddrs {
			addr, err := net.ResolveUDPAddr("udp", relayaddrs[idx])
			if err != nil {
				log.Fatal(err)
			}

			relaypubkey, err := base64.StdEncoding.DecodeString(relaykeys[idx])
			if err != nil {
				log.Fatal(err)
			}

			nextRouteToken.Relays = append(nextRouteToken.Relays, routing.Relay{Addr: *addr, PublicKey: relaypubkey})
		}

		routeToken = &nextRouteToken
	case "continue":
		tokenTypeIdentifyingByte[0] = routing.TokenTypeContinueRequest
		continueRouteToken := routing.ContinueRouteToken{
			Expires: uint64(time.Now().Add(*expires).Unix()),

			SessionId:      uint64(*sessionID),
			SessionVersion: uint8(*sessionversion),

			Client: routing.Client{Addr: *client, PublicKey: clientpubkey},

			Server: routing.Server{Addr: *server, PublicKey: serverpubkey},
		}

		for idx := range relayaddrs {
			addr, err := net.ResolveUDPAddr("udp", relayaddrs[idx])
			if err != nil {
				log.Fatal(err)
			}

			relaypubkey, err := base64.StdEncoding.DecodeString(relaykeys[idx])
			if err != nil {
				log.Fatal(err)
			}

			continueRouteToken.Relays = append(continueRouteToken.Relays, routing.Relay{Addr: *addr, PublicKey: relaypubkey})
		}

		routeToken = &continueRouteToken
	default:
		log.Fatalf("%s is not a valid token type", *token)
	}

	enc, _, err := routeToken.Encrypt(privatekey)
	if err != nil {
		log.Fatalf("failed to encrypt token: %v", err)
	}

	if *clientkey == "" {
		enc = append(tokenTypeIdentifyingByte, enc[transport.EncryptedTokenRouteSize:]...)
	}

	_, err = os.Stdout.Write(enc)
	if err != nil {
		log.Fatalf("failed to write token to stdout: %v", err)
	}
}
