package transport_test

import (
	"net"
	"testing"

	"github.com/networknext/backend/transport"
)

func TestServerUpdateHandlerFunc(t *testing.T) {
	packet := make([]byte, 0)
	addr, _ := net.ResolveUDPAddr("udp", ":30000")

	transport.ServerUpdateHandlerFunc(packet, addr)
}

func TestSessionUpdateHandlerFunc(t *testing.T) {
	packet := make([]byte, 0)
	addr, _ := net.ResolveUDPAddr("udp", ":30000")

	transport.ServerUpdateHandlerFunc(packet, addr)
}
