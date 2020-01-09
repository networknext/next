package transport_test

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestServerUpdateHandlerFunc(t *testing.T) {
	// Get an in-memory redis server and a client that is connected to it
	redisServer, redisClient := NewTestRedis()

	// Create a ServerUpdatePacket and marshal it to binary so sent it into the UDP handler
	packet := core.ServerUpdatePacket{
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ServerRoutePublicKey: TestPublicKey,
		Signature:            []byte{0x00},

		DatacenterId: 13,

		VersionMajor: core.SDKVersionMajorMin,
		VersionMinor: core.SDKVersionMinorMin,
		VersionPatch: core.SDKVersionPatchMin,
	}
	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	// Initialize the UDP handler with the required redis client
	handler := transport.ServerUpdateHandlerFunc(redisClient)

	// Invoke the handler with the data packet and address it is coming from
	handler(nil, data, addr)

	// Get the server entry directly from the in-memory redis and assert there is no error
	ds, err := redisServer.Get("SERVER-0.0.0.0:13")
	assert.NoError(t, err)

	// Create an "expected" ServerEntry based on the incoming ServerUpdatePacket above
	expected := transport.ServerEntry{
		ServerRoutePublicKey: packet.ServerRoutePublicKey,
		ServerPrivateAddr:    *addr,

		DatacenterID: packet.DatacenterId,

		VersionMajor: packet.VersionMajor,
		VersionMinor: packet.VersionMinor,
		VersionPatch: packet.VersionPatch,
	}

	// Unmarshal the data in redis to the actual ServerEntry saved
	var actual transport.ServerEntry
	err = actual.UnmarshalBinary([]byte(ds))
	assert.NoError(t, err)

	// Finally compare both ServerEntry struct to ensure we saved the right data in redis
	assert.Equal(t, expected, actual)
}

func TestSessionUpdateHandlerFunc(t *testing.T) {
	t.Skip()

	redisServer, redisClient := NewTestRedis()

	ipStackClient := transport.IPStackClient{
		Client: NewTestHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body: ioutil.NopCloser(bytes.NewBufferString(`{
					"ip": "172.100.205.154",
					"continent_code": "NA",
					"country_code": "US",
					"region_code": "NY",
					"city": "Troy",
					"latitude": 43.05036163330078,
					"longitude": -73.75393676757812,
					"connection": {
						"asn": 11351,
						"isp": "Charter Communications Inc"
					}
				}`)),
			}
		}),
	}

	packet := core.SessionUpdatePacket{
		SessionId: 13,
		UserHash:  13,

		ClientRoutePublicKey: make([]byte, 1),
		Signature:            make([]byte, 1),
	}
	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	addr, _ := net.ResolveUDPAddr("udp", ":30000")

	handler := transport.SessionUpdateHandlerFunc(redisClient, &ipStackClient)
	handler(nil, data, addr)

	ds, err := redisServer.Get("SESSION-13")
	assert.NoError(t, err)

	var sessionentry transport.SessionEntry
	sessionentry.UnmarshalBinary([]byte(ds))

	assert.Equal(t, sessionentry, transport.SessionEntry{})
}
