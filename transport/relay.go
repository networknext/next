package transport

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
)

const (
	PacketTypeServerUpdate = iota + 200
	PacketTypeSessionUpdate
	PacketTypeSessionResponse

	DefaultMaxPacketSize = 1500
)

// UDPHandlerFunc acts the same way http.HandlerFunc does, but for UDP packets and address
type UDPHandlerFunc func(*bytes.Buffer, net.Addr)

// RelayServer is a simple UDP router for specific packets and runs each UDPHandlerFunc based on the incoming packet type
type RelayServer struct {
	Conn          *net.UDPConn
	MaxPacketSize int

	ServerUpdateHandlerFunc  UDPHandlerFunc
	SessionUpdateHandlerFunc UDPHandlerFunc
}

// Start begins accepting UDP packets from the UDP connection and will block
func (rs *RelayServer) Start() error {
	if rs.Conn == nil {
		return errors.New("relay server cannot be nil")
	}

	packet := make([]byte, rs.MaxPacketSize)

	for {
		numbytes, addr, _ := rs.Conn.ReadFrom(packet)
		if numbytes <= 0 {
			continue
		}

		switch packet[0] {
		case PacketTypeServerUpdate:
			rs.ServerUpdateHandlerFunc(bytes.NewBuffer(packet[1:numbytes]), addr)
		case PacketTypeSessionUpdate:
			rs.SessionUpdateHandlerFunc(bytes.NewBuffer(packet[1:numbytes]), addr)
		}
	}
}

// ServerUpdatePacket ...
type ServerUpdatePacket struct{}

// MarshalBinary is the same as MarshalJSON but performs the binary format encoding we need
func (sup ServerUpdatePacket) MarshalBinary() ([]byte, error) {
	return nil, nil
}

// UnmarshalBinary is the same as UnmarshalJSON but performs the binary format decoding we need
func (sup *ServerUpdatePacket) UnmarshalBinary(data []byte) error {
	return nil
}

// ServerUpdateHandlerFunc ...
func ServerUpdateHandlerFunc(packet *bytes.Buffer, from net.Addr) {
	var sup ServerUpdatePacket
	if err := binary.Read(packet, binary.LittleEndian, &sup); err != nil {
		log.Println(err)
	}

	log.Println("not implemented")
}

// ServerUpdatePacket ...
type SessionUpdatePacket struct{}

// MarshalBinary is the same as MarshalJSON but performs the binary format encoding we need
func (sup *SessionUpdatePacket) MarshalBinary() ([]byte, error) {
	return nil, nil
}

// UnmarshalBinary is the same as UnmarshalJSON but performs the binary format decoding we need
func (sup *SessionUpdatePacket) UnmarshalBinary(data []byte) error {
	return nil
}

// SessionUpdateHandlerFunc ...
func SessionUpdateHandlerFunc(packet *bytes.Buffer, from net.Addr) {
	var sup SessionUpdatePacket
	if err := binary.Read(packet, binary.LittleEndian, &sup); err != nil {
		log.Println(err)
	}

	log.Println("not implemented")
}
