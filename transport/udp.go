package transport

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

type UDPPacket struct {
	SourceAddr *net.UDPAddr
	Data       []byte
}

// UDPHandlerFunc acts the same way http.HandlerFunc does, but for UDP packets and address
type UDPHandlerFunc func(io.Writer, *UDPPacket)

// ServerIngress is a simple UDP router for specific packets and runs each UDPHandlerFunc based on the incoming packet type
type UDPServerMux struct {
	Conn          *net.UDPConn
	MaxPacketSize int

	ServerUpdateHandlerFunc  UDPHandlerFunc
	SessionUpdateHandlerFunc UDPHandlerFunc
}

// Start begins accepting UDP packets from the UDP connection and will block
func (m *UDPServerMux) Start(ctx context.Context, handlers int) error {
	if m.Conn == nil {
		return errors.New("relay server cannot be nil")
	}

	for i := 0; i < handlers; i++ {
		go m.handler(ctx, i)
	}

	<-ctx.Done()

	return nil
}

func (m *UDPServerMux) handler(ctx context.Context, id int) {
	data := make([]byte, m.MaxPacketSize)

	for {
		numbytes, addr, _ := m.Conn.ReadFromUDP(data)
		if numbytes <= 0 {
			continue
		}
		log.Println("handler", id, "addr", addr.String(), "bytes", numbytes)

		var buf bytes.Buffer
		packet := UDPPacket{
			SourceAddr: addr,
			Data:       data[:numbytes],
		}

		switch data[0] {
		case PacketTypeServerUpdate:
			m.ServerUpdateHandlerFunc(&buf, &packet)
		case PacketTypeSessionUpdate:
			m.SessionUpdateHandlerFunc(&buf, &packet)
		}

		if buf.Len() > 0 {
			_, err := m.Conn.WriteToUDP(buf.Bytes(), addr)
			if err != nil {
				log.Println("addr", addr.String(), "msg", err.Error())
			}
		}
	}
}

type ServerCacheEntry struct {
	Sequence   uint64
	Server     routing.Server
	Datacenter routing.Datacenter
	SDKVersion SDKVersion
}

func (e *ServerCacheEntry) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, e)
}

func (e ServerCacheEntry) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(e)
}

type BuyerProvider interface {
	GetAndCheckBySdkVersion3PublicKeyId(key uint64) (*storage.Buyer, bool)
}

// ServerUpdateHandlerFunc ...
func ServerUpdateHandlerFunc(redisClient redis.Cmdable, bp BuyerProvider) UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		var packet ServerUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			log.Printf("failed to read server update packet: %v\n", err)
			return
		}

		// Drop the packet if version is older that the minimun sdk version
		psdkv := SDKVersion{packet.VersionMajor, packet.VersionMinor, packet.VersionPatch}
		if !incoming.SourceAddr.IP.IsLoopback() &&
			psdkv.Compare(SDKVersionMin) == SDKVersionOlder {
			log.Printf("sdk version is too old. Using %s but require at least %s", psdkv, SDKVersionMin)
			return
		}

		// Get the buyer information for the id in the packet
		buyer, ok := bp.GetAndCheckBySdkVersion3PublicKeyId(packet.CustomerId)
		if !ok {
			log.Printf("failed to get buyer '%d'", packet.CustomerId)
			return
		}

		// This was in the Router, but no sense having that buried in there when we already have
		// a Buyer to check before requesting a Route
		if !buyer.GetActive() {
			log.Printf("buyer '%s' is inactive", buyer.GetName())
			return
		}

		buyerPublicKey := buyer.SdkVersion3PublicKeyData
		// Drop the packet if the buyer is not an admin and they are using an internal build
		// if !buyer.GetA && psdkv.Compare(SDKVersionInternal) == SDKVersionEqual {
		// 	log.Printf("non-admin buyer using an internal sdk")
		// 	return
		// }

		// Drop the packet if the signed packet data cannot be verified with the buyers public key
		if !crypto.Verify(packet.GetSignData(), packet.Signature, buyerPublicKey) {
			log.Printf("ed25519: failed to verify server update signature")
			return
		}

		// Get the the old ServerCacheEntry if it exists, otherwise serverentry is in zero value state
		var serverentry ServerCacheEntry
		{
			result := redisClient.Get("SERVER-" + incoming.SourceAddr.String())
			if result.Err() != nil && result.Err() != redis.Nil {
				log.Printf("failed to load server %s from cache: %v", incoming.SourceAddr.String(), result.Err())
				return
			}
			serverdata, err := result.Bytes()
			if err != nil && result.Err() != redis.Nil {
				log.Printf("failed to get bytes from cache: %v", result.Err())
				return
			}
			if serverdata != nil {
				if err := serverentry.UnmarshalBinary(serverdata); err != nil {
					fmt.Printf("failed to unmarshal server cache entry: %v\n", err)
				}
			}
		}

		// Drop the packet if the sequence number is older than the previously cache sequence number
		// if packet.Sequence < serverentry.Sequence {
		// 	log.Printf("packet too old: (packet) %d < %d (Redis)", packet.Sequence, serverentry.Sequence)
		// 	return
		// }

		// Save some of the packet information to be used in SessionUpdateHandlerFunc
		serverentry = ServerCacheEntry{
			Sequence:   packet.Sequence,
			Server:     routing.Server{Addr: packet.ServerPrivateAddress, PublicKey: packet.ServerRoutePublicKey},
			Datacenter: routing.Datacenter{ID: packet.DatacenterId},
			SDKVersion: SDKVersion{packet.VersionMajor, packet.VersionMinor, packet.VersionPatch},
		}
		result := redisClient.Set("SERVER-"+incoming.SourceAddr.String(), serverentry, 5*time.Minute)
		if result.Err() != nil {
			log.Printf("failed to cache server %s: %v", incoming.SourceAddr.String(), result.Err())
			return
		}

		log.Printf("cached server '%s' for sequence '%d'\n", incoming.SourceAddr.String(), packet.Sequence)
	}
}

type SessionEntry struct {
	CustomerID uint64
	SessionID  uint64
	UserID     uint64
	PlatformID uint64

	NearRelays []routing.Relay

	DirectRTT        float64
	DirectJitter     float64
	DirectPacketLoss float64
	NextRTT          float64
	NextJitter       float64
	NextPacketLoss   float64

	ServerRoutePublicKey []byte
	ServerPrivateAddr    net.UDPAddr
	ServerAddr           net.UDPAddr
	ClientAddr           net.UDPAddr

	ConnectionType int32

	Latitude  float64
	Longitude float64

	Tag   uint64
	Flags uint32

	Flagged          bool
	TryBeforeYouBuy  bool
	OnNetworkNext    bool
	FallbackToDirect bool

	SDKVersion SDKVersion
}

func (e *SessionEntry) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, e)
}

func (e SessionEntry) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(e)
}

type RouteProvider interface {
	AllRoutes(routing.Datacenter, []routing.Relay) []routing.Route
}

// SessionUpdateHandlerFunc ...
func SessionUpdateHandlerFunc(redisClient redis.Cmdable, bp BuyerProvider, rp RouteProvider, iploc routing.IPLocator, geoClient *routing.GeoClient, encryptionPrivateKey []byte, signingPrivateKey []byte) UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		// Deserialize the Session packet
		var packet SessionUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			log.Printf("failed to read server update packet: %v\n", err)
			return
		}

		result := redisClient.Get("SERVER-" + incoming.SourceAddr.String())
		if result.Err() != nil {
			log.Fatalf("failed to get server entry from redis for '%s': %v", incoming.SourceAddr.String(), result.Err())
			return
		}

		serverdata, err := result.Bytes()
		if err != nil {
			log.Fatalf("failed to get server entry from redis for '%s': %v", incoming.SourceAddr.String(), err)
			return
		}

		var serverentry ServerCacheEntry
		if err := serverentry.UnmarshalBinary(serverdata); err != nil {
			log.Fatalf("failed to unmarshal server entry from redis for '%s': %v", incoming.SourceAddr.String(), err)
			return
		}

		buyer, ok := bp.GetAndCheckBySdkVersion3PublicKeyId(packet.CustomerId)
		if !ok {
			log.Printf("failed to get buyer '%d'", packet.CustomerId)
			return
		}
		buyerServerPublicKey := buyer.GetSdkVersion3PublicKeyData()

		// if !ed25519.Verify(buyerServerPublicKey, packet.GetSignData(serverentry.SDKVersion), packet.Signature) {
		// 	log.Printf("failed to verify session update signature")
		// 	return
		// }

		// if packet.Sequence < serverentry.Sequence {
		// 	log.Printf("packet too old: (packet) %d < %d (Redis)", packet.Sequence, serverentry.Sequence)
		// 	return
		// }

		location, err := iploc.LocateIP(packet.ClientAddress.IP)
		if err != nil {
			log.Printf("failed to lookup client ip '%s': %v", packet.ClientAddress.IP.String(), err)
			return
		}

		clientrelays, err := geoClient.RelaysWithin(location.Latitude, location.Longitude, 500, "mi")
		if err != nil {
			log.Printf("failed to lookup client ip '%s': %v", packet.ClientAddress.IP.String(), err)
			return
		}

		// Get a set of possible routes from the RouteProvider an on error ensure it falls back to direct
		routes := rp.AllRoutes(serverentry.Datacenter, clientrelays)
		if routes == nil {
			fmt.Println("failed to find routes")
			return
		}
		chosenRoute := routes[0] // Just take the first one it find regardless of optimizations

		// Build the next route with the client, server, and set of relays to use
		nextRoute := routing.NextRouteToken{
			Expires: uint64(time.Now().Add(10 * time.Second).Unix()),

			SessionId: packet.SessionId,

			SessionVersion: 0, // Haven't figured out what this is for
			SessionFlags:   0, // Haven't figured out what this is for

			Client: routing.Client{
				Addr:      packet.ClientAddress,
				PublicKey: packet.ClientRoutePublicKey,
			},

			Server: routing.Server{
				Addr:      packet.ServerAddress,
				PublicKey: buyerServerPublicKey,
			},

			Relays: chosenRoute.Relays,
		}

		// Encrypt the next route with the our private key
		routeTokens, err := nextRoute.Encrypt(encryptionPrivateKey)
		if err != nil {
			log.Fatalf("failed to encrypt route token: %v", err)
			return
		}

		// Create the Session Response for the server
		response := SessionResponsePacket{
			Sequence:             packet.Sequence,
			SessionId:            packet.SessionId,
			RouteType:            int32(routing.DecisionTypeNew),
			NumTokens:            int32(len(chosenRoute.Relays) + 2), // Num of relays + client + server
			Tokens:               routeTokens,
			Multipath:            false, // Haven't figured out what this is for
			ServerRoutePublicKey: serverentry.Server.PublicKey,
		}

		// Fill in the near relays
		response.NumNearRelays = int32(len(clientrelays))
		response.NearRelayIds = make([]uint64, len(clientrelays))
		response.NearRelayAddresses = make([]net.UDPAddr, len(clientrelays))
		for idx, relay := range clientrelays {
			response.NearRelayIds[idx] = relay.ID
			response.NearRelayAddresses[idx] = relay.Addr
		}

		// Sign the response
		response.Signature = ed25519.Sign(signingPrivateKey, response.GetSignData())

		// Send the Session Response back to the server
		res, err := response.MarshalBinary()
		if err != nil {
			log.Printf("failed to marshal session response to '%s': %v", incoming.SourceAddr.String(), err)
			return
		}

		if _, err := w.Write(res); err != nil {
			log.Printf("failed to write session response to '%s': %v", incoming.SourceAddr.String(), err)
		}
	}
}
