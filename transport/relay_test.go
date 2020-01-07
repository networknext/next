package transport_test

import (
	"bytes"
	"github.com/networknext/backend/transport"
	"net"
	"testing"
)

func TestServerUpdateHandlerFunc(t *testing.T) {
	packet := bytes.NewBuffer(nil)
	addr, _ := net.ResolveUDPAddr("udp", ":30000")

	transport.ServerUpdateHandlerFunc(packet, addr)
}

func TestSessionUpdateHandlerFunc(t *testing.T) {
	packet := bytes.NewBuffer(nil)
	addr, _ := net.ResolveUDPAddr("udp", ":30000")

	transport.ServerUpdateHandlerFunc(packet, addr)
}
