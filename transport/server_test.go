package transport_test

import (
	"net"
	"testing"

	"github.com/networknext/backend/transport"
)

func TestServerUpdateHandlerFunc(t *testing.T) {
	t.Skip()

	packet := make([]byte, 0)
	addr, _ := net.ResolveUDPAddr("udp", ":30000")

	handler := transport.ServerUpdateHandlerFunc(nil)
	handler(nil, packet, addr)
}

func TestSessionUpdateHandlerFunc(t *testing.T) {
	t.Skip()

	packet := make([]byte, 0)
	addr, _ := net.ResolveUDPAddr("udp", ":30000")

	handler := transport.ServerUpdateHandlerFunc(nil)
	handler(nil, packet, addr)
}
