package transport

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/gomodule/redigo/redis"
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
			m.ServerUpdateHandlerFunc(m.Conn, packet[1:numbytes], addr)
		case PacketTypeSessionUpdate:
			m.SessionUpdateHandlerFunc(m.Conn, packet[1:numbytes], addr)
		}
	}
}

// ServerUpdateHandlerFunc ...
func ServerUpdateHandlerFunc(redisConn redis.Conn) UDPHandlerFunc {
	return func(conn *net.UDPConn, packet []byte, from *net.UDPAddr) {
		// Deserialize the Session packet
		var sup core.ServerUpdatePacket
		{
			if err := sup.Serialize(core.CreateReadStream(packet)); err != nil {
				fmt.Printf("failed to read server update packet: %v\n", err)
				return
			}
		}

		// Verify the Session packet version
		if !core.ProtocolVersionAtLeast(sup.VersionMajor, sup.VersionMinor, sup.VersionPatch, core.SDKVersionMajorMin, core.SDKVersionMinorMin, core.SDKVersionPatchMin) {
			log.Printf("sdk version is too old. Using %d.%d.%d but require at least %d.%d.%d", sup.VersionMajor, sup.VersionMinor, sup.VersionPatch, core.SDKVersionMajorMin, core.SDKVersionMinorMin, core.SDKVersionPatchMin)
			return
		}

		// Get the the old Server packet from Redis
		var serverentry core.ServerUpdatePacket
		{
			serverdata, err := redis.Bytes(redisConn.Do("GET", "SERVER-"+from.String()))
			if err != nil {
				log.Printf("failed to register server %s: %v", from.String(), err)
				return
			}

			if err := serverentry.Serialize(core.CreateReadStream(serverdata)); err != nil {
				fmt.Printf("failed to read server entry: %v\n", err)
				return
			}
		}

		// TODO 1. Get Buyer and Customer information from ConfigStore

		// TODO 2. Check server packet version for customer and don't let them use 0.0.0

		// signdata := sup.GetSignData()

		// Save the Server packet to Redis
		{
			ws, err := core.CreateWriteStream(DefaultMaxPacketSize)
			if err != nil {
				fmt.Printf("failed to create server entry read stream: %v\n", err)
				return
			}

			if err := serverentry.Serialize(ws); err != nil {
				fmt.Printf("failed to read server entry: %v\n", err)
				return
			}
			ws.Flush()

			serverdata := ws.GetData()

			if _, err := redisConn.Do("SET", "SERVER-"+from.String(), serverdata[:ws.GetBytesProcessed()]); err != nil {
				log.Printf("failed to save server db entry for %s: %v", from.String(), err)
			}
		}
	}
}

// SessionUpdateHandlerFunc ...
func SessionUpdateHandlerFunc(redisConn redis.Conn) UDPHandlerFunc {
	return func(conn *net.UDPConn, packet []byte, from *net.UDPAddr) {
		// Deserialize the Session packet
		sup := core.SessionUpdatePacket{}
		{
			if err := sup.Serialize(core.CreateReadStream(packet), core.SDKVersionMajorMax, core.SDKVersionMinorMax, core.SDKVersionPatchMax); err != nil {
				fmt.Printf("failed to read session update packet: %v\n", err)
				return
			}
		}

		// Change Session Packet

		// Save the Session packet to Redis
		{
			ws, err := core.CreateWriteStream(DefaultMaxPacketSize)
			if err != nil {
				fmt.Printf("failed to create session entry read stream: %v\n", err)
				return
			}

			if err := sup.Serialize(ws, core.SDKVersionMajorMin, core.SDKVersionMinorMin, core.SDKVersionPatchMin); err != nil {
				fmt.Printf("failed to read session entry: %v\n", err)
				return
			}
			ws.Flush()

			sessiondata := ws.GetData()

			if _, err := redisConn.Do("SET", fmt.Sprintf("SESSION-%d", sup.SessionId), sessiondata[:ws.GetBytesProcessed()]); err != nil {
				log.Printf("failed to save session db entry for %s: %v", from.String(), err)
			}
		}
	}
}
