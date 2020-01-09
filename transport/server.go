package transport

import (
	"errors"
	"fmt"
	"log"
	"net"

	jsoniter "github.com/json-iterator/go"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/core"
)

const (
	PacketTypeServerUpdate = iota + 200
	PacketTypeSessionUpdate
	PacketTypeSessionResponse

	DefaultMaxPacketSize = 1500
)

// UDPHandlerFunc acts the same way http.HandlerFunc does, but for UDP packets and address
type UDPHandlerFunc func(*net.UDPConn, []byte, *net.UDPAddr)

// ServerIngress is a simple UDP router for specific packets and runs each UDPHandlerFunc based on the incoming packet type
type UDPServerMux struct {
	Conn          *net.UDPConn
	MaxPacketSize int

	ServerUpdateHandlerFunc  UDPHandlerFunc
	SessionUpdateHandlerFunc UDPHandlerFunc
}

// Start begins accepting UDP packets from the UDP connection and will block
func (m *UDPServerMux) Start() error {
	if m.Conn == nil {
		return errors.New("relay server cannot be nil")
	}

	packet := make([]byte, m.MaxPacketSize)

	for {
		numbytes, addr, _ := m.Conn.ReadFromUDP(packet)
		if numbytes <= 0 {
			continue
		}

		switch packet[0] {
		case PacketTypeServerUpdate:
			m.ServerUpdateHandlerFunc(m.Conn, packet[1:], addr)
		case PacketTypeSessionUpdate:
			m.SessionUpdateHandlerFunc(m.Conn, packet[1:], addr)
		}
	}
}

type ServerEntry struct {
	ServerRoutePublicKey []byte
	ServerPrivateAddr    net.UDPAddr

	DatacenterID      uint64
	DatacenterName    string
	DatacenterEnabled bool

	VersionMajor int32
	VersionMinor int32
	VersionPatch int32
}

func (e *ServerEntry) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, e)
}

func (e ServerEntry) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(e)
}

// ServerUpdateHandlerFunc ...
func ServerUpdateHandlerFunc(redisClient *redis.Client) UDPHandlerFunc {
	return func(conn *net.UDPConn, data []byte, from *net.UDPAddr) {
		var packet core.ServerUpdatePacket
		if err := packet.UnmarshalBinary(data); err != nil {
			fmt.Printf("failed to read server update packet: %v\n", err)
			return
		}

		// Verify the Session packet version
		if !core.ProtocolVersionAtLeast(packet.VersionMajor, packet.VersionMinor, packet.VersionPatch, core.SDKVersionMajorMin, core.SDKVersionMinorMin, core.SDKVersionPatchMin) {
			log.Printf("sdk version is too old. Using %d.%d.%d but require at least %d.%d.%d", packet.VersionMajor, packet.VersionMinor, packet.VersionPatch, core.SDKVersionMajorMin, core.SDKVersionMinorMin, core.SDKVersionPatchMin)
			return
		}

		// Get the the old ServerEntry from Redis
		serverentry := ServerEntry{}
		{
			result := redisClient.Get("SERVER-" + from.String())
			if result.Err() != redis.Nil {
				log.Printf("failed to get server %s from redis: %v", from.String(), result.Err())
				return
			}
			serverdata, err := result.Bytes()
			if err != redis.Nil {
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
			fmt.Println(serverentry.ServerPrivateAddr)
			result := redisClient.Set("SERVER-"+from.String(), serverentry, 0)
			if result.Err() != nil {
				log.Printf("failed to register server %s: %v", from.String(), result.Err())
				return
			}
		}
	}
}

type SessionEntry struct {
	ServerRoutePublicKey []byte
	ServerPrivateAddr    net.UDPAddr

	DatacenterID      uint64
	DatacenterName    string
	DatacenterEnabled bool

	VersionMajor int32
	VersionMinor int32
	VersionPatch int32
}

func (e *SessionEntry) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, e)
}

func (e SessionEntry) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(e)
}

// SessionUpdateHandlerFunc ...
func SessionUpdateHandlerFunc(redisClient *redis.Client) UDPHandlerFunc {
	return func(conn *net.UDPConn, data []byte, from *net.UDPAddr) {
		// Deserialize the Session packet
		var packet core.SessionUpdatePacket
		if err := packet.UnmarshalBinary(data); err != nil {
			fmt.Printf("failed to read server update packet: %v\n", err)
			return
		}

		// Change Session Packet

		// Save the Session packet to Redis
		var sessionentry SessionEntry
		{
			result := redisClient.Set(fmt.Sprintf("SESSION-%d", packet.SessionId), sessionentry, 0)
			if result.Err() != nil {
				log.Printf("failed to save session db entry for %s: %v", from.String(), result.Err())
				return
			}
		}
	}
}
