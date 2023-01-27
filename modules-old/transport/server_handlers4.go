package transport

import (
	"net"
	"io"
)

const (
	UDPIPPacketHeaderSize = 28 // IP: 20, UDP: 8
)

type UDPPacket struct {
	From net.UDPAddr
	Data []byte
}

type UDPHandlerFunc func(io.Writer, *UDPPacket)
