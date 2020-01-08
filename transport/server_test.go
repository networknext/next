package transport_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/gomodule/redigo/redis"
	"github.com/networknext/backend/core"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestServerUpdateHandlerFunc(t *testing.T) {
	var packet []byte
	{
		serverentry := core.ServerUpdatePacket{
			Sequence:             13,
			ServerRoutePublicKey: make([]byte, 1),
			Signature:            make([]byte, 1),
		}
		ws, err := core.CreateWriteStream(transport.DefaultMaxPacketSize)
		if err != nil {
			fmt.Printf("failed to create server entry read stream: %v\n", err)
			return
		}

		if err := serverentry.Serialize(ws); err != nil {
			fmt.Printf("failed to read server entry: %v\n", err)
			return
		}
		ws.Flush()
		packet = ws.GetData()
	}

	redisServer, _ := miniredis.Run()
	redisConn, _ := redis.Dial("tcp", redisServer.Addr())

	addr, _ := net.ResolveUDPAddr("udp", ":30000")

	handler := transport.ServerUpdateHandlerFunc(redisConn)
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
