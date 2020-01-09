package transport_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/core"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestServerUpdateHandlerFunc(t *testing.T) {
	serverentry := core.ServerUpdatePacket{
		Sequence:             13,
		ServerRoutePublicKey: make([]byte, 1),
		Signature:            make([]byte, 1),
	}
	packet, err := serverentry.MarshalBinary()
	if err != nil {
		fmt.Printf("failed to marshal server entry: %v\n", err)
		return
	}

	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	addr, _ := net.ResolveUDPAddr("udp", ":30000")

	handler := transport.ServerUpdateHandlerFunc(redisClient)
	handler(nil, packet, addr)

	ds, _ := redisServer.Get("SERVER-:30000")

	assert.Equal(t, uint8(0xD), uint8(ds[0]))
}

func TestSessionUpdateHandlerFunc(t *testing.T) {
	t.Skip()

	packet := make([]byte, 0)
	addr, _ := net.ResolveUDPAddr("udp", ":30000")

	handler := transport.ServerUpdateHandlerFunc(nil)
	handler(nil, packet, addr)
}
