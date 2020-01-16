package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	jsoniter "github.com/json-iterator/go"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/core"
	"github.com/networknext/backend/routing"
)

const (
	PacketTypeServerUpdate = iota + 200
	PacketTypeSessionUpdate
	PacketTypeSessionResponse
)

const (
	DefaultMaxPacketSize = 1500
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
			Data:       data,
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

type ServerEntry struct {
	ServerRoutePublicKey []byte
	ServerPrivateAddr    net.UDPAddr

	DatacenterID      uint64
	DatacenterName    string
	DatacenterEnabled bool

	SDKVersion SDKVersion
}

func (e *ServerEntry) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, e)
}

func (e ServerEntry) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(e)
}

// ServerUpdateHandlerFunc ...
func ServerUpdateHandlerFunc(redisClient redis.Cmdable) UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		var packet core.ServerUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			fmt.Printf("failed to read server update packet: %v\n", err)
			return
		}

		// Drop the packet if version is older that the minimun sdk version
		psdkv := SDKVersion{packet.VersionMajor, packet.VersionMinor, packet.VersionPatch}
		if psdkv.Compare(SDKVersionMin) == SDKVersionOlder {
			log.Printf("sdk version is too old. Using %s but require at least %s", psdkv, SDKVersionMin)
			return
		}
			return
		}

		// Get the the old ServerEntry from Redis
		var serverentry ServerEntry
		{
			result := redisClient.Get("SERVER-" + incoming.SourceAddr.String())
			if result.Err() != nil && result.Err() != redis.Nil {
				log.Printf("failed to get server %s from redis: %v", incoming.SourceAddr.String(), result.Err())
				return
			}
			serverdata, err := result.Bytes()
			if err != nil && result.Err() != redis.Nil {
				log.Printf("failed to get bytes from redis: %v", result.Err())
				return
			}
			if serverdata != nil {
				if err := serverentry.UnmarshalBinary(serverdata); err != nil {
					fmt.Printf("failed to read server entry: %v\n", err)
				}
			}
		}

		// TODO 1. Get Buyer and Customer information from ConfigStore

		// TODO 2. Check server packet version for customer and don't let them use 0.0.0

		// signdata := sup.GetSignData()

		// Save the ServerEntry to Redis
		{
			serverentry = ServerEntry{
				ServerRoutePublicKey: packet.ServerRoutePublicKey,
				ServerPrivateAddr:    packet.ServerPrivateAddress,

				DatacenterID: packet.DatacenterId,

				VersionMajor: packet.VersionMajor,
				VersionMinor: packet.VersionMinor,
				VersionPatch: packet.VersionPatch,
			}
			result := redisClient.Set("SERVER-"+incoming.SourceAddr.String(), serverentry, 0)
			if result.Err() != nil {
				log.Printf("failed to register server %s: %v", incoming.SourceAddr.String(), result.Err())
				return
			}
			SDKVersion: SDKVersion{packet.VersionMajor, packet.VersionMinor, packet.VersionPatch},
		}
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

// SessionUpdateHandlerFunc ...
func SessionUpdateHandlerFunc(redisClient redis.Cmdable, iploc routing.IPLocator, geoClient *routing.GeoClient) UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		// Deserialize the Session packet
		var packet core.SessionUpdatePacket
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

		var serverentry ServerEntry
		if err := serverentry.UnmarshalBinary(serverdata); err != nil {
			log.Fatalf("failed to unmarshal server entry from redis for '%s': %v", incoming.SourceAddr.String(), err)
			return
		}

		location, err := iploc.LocateIP(packet.ClientAddress.IP)
		if err != nil {
			log.Printf("failed to lookup client ip '%s': %v", packet.ClientAddress.IP.String(), err)
			return
		}

		relays, err := geoClient.RelaysWithin(location.Latitude, location.Longitude, 500, "mi")
		if err != nil {
			log.Printf("failed to lookup client ip '%s': %v", packet.ClientAddress.IP.String(), err)
			return
		}

		// Save the Session packet to Redis
		sessionentry := SessionEntry{
			SessionID:  packet.SessionId,
			UserID:     packet.UserHash,
			PlatformID: packet.PlatformId,

			DirectRTT:        float64(packet.DirectMinRtt),
			DirectJitter:     float64(packet.DirectJitter),
			DirectPacketLoss: float64(packet.DirectPacketLoss),
			NextRTT:          float64(packet.NextMinRtt),
			NextJitter:       float64(packet.NextJitter),
			NextPacketLoss:   float64(packet.NextPacketLoss),

			ServerRoutePublicKey: serverentry.ServerRoutePublicKey,
			ServerPrivateAddr:    serverentry.ServerPrivateAddr,
			ServerAddr:           packet.ServerAddress,
			ClientAddr:           packet.ClientAddress,

			ConnectionType: packet.ConnectionType,

			Latitude:  location.Latitude,
			Longitude: location.Longitude,

			Tag:              packet.Tag,
			Flagged:          packet.Flagged,
			FallbackToDirect: packet.FallbackToDirect,
			OnNetworkNext:    packet.OnNetworkNext,

			SDKVersion: serverentry.SDKVersion,
		}
		{
			result := redisClient.Set(fmt.Sprintf("SESSION-%d", packet.SessionId), sessionentry, 0)
			if result.Err() != nil {
				log.Printf("failed to save session db entry for %s: %v", incoming.SourceAddr.String(), result.Err())
				return
			}
		}

		// Create a zeo-value route for now until we get this information from the optimized route matrix/router
		route := routing.Route{
			Type:      routing.RouteTypeDirect,
			NumTokens: 0,
			Tokens:    nil,
			Multipath: false,
		}

		// Create the Session Response for the server
		response := core.SessionResponsePacket{
			Sequence:             packet.Sequence,
			SessionId:            packet.SessionId,
			RouteType:            int32(route.Type),
			NumTokens:            int32(route.NumTokens),
			Tokens:               route.Tokens,
			Multipath:            route.Multipath,
			ServerRoutePublicKey: serverentry.ServerRoutePublicKey,
		}

		// Fill in the near relays
		response.NumNearRelays = int32(len(relays))
		response.NearRelayIds = make([]uint64, len(relays))
		response.NearRelayAddresses = make([]net.UDPAddr, len(relays))
		for idx, relay := range relays {
			response.NearRelayIds[idx] = relay.ID
			response.NearRelayAddresses[idx] = relay.Addr
		}

		// Sign the response
		response.Sign(serverentry.SDKVersion.Major, serverentry.SDKVersion.Minor, serverentry.SDKVersion.Patch)

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
