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

	// Create an in-memory buffer to give to the hander since it implements io.Writer
	var buf bytes.Buffer

	// Create a UDPPacket for the handler
	incoming := transport.UDPPacket{
		SourceAddr: addr,
		Data:       data,
	}

	// Initialize the UDP handler with the required redis client
	handler := transport.ServerUpdateHandlerFunc(redisClient)

	// Invoke the handler with the data packet and address it is coming from
	handler(&buf, &incoming)

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

	assert.Equal(t, 0, buf.Len())
}

func TestSessionUpdateHandlerFunc(t *testing.T) {
	redisServer, redisClient := NewTestRedis()

	// New IPStackClient that mocks a successful response
	ipStackClient := transport.IPStackClient{
		Client: NewTestHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body: ioutil.NopCloser(bytes.NewBufferString(`{
					"ip": "1.1.1.1",
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

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:12345")
	assert.NoError(t, err)

	// Create a ServerEntry to put into redis for a SessionUpdate to read
	serverentry := transport.ServerEntry{
		ServerRoutePublicKey: TestPublicKey,
		ServerPrivateAddr:    *addr,

		VersionMajor: core.SDKVersionMajorMin,
		VersionMinor: core.SDKVersionMinorMin,
		VersionPatch: core.SDKVersionPatchMin,
	}
	serverdata, err := serverentry.MarshalBinary()
	assert.NoError(t, err)

	// Set the ServerEntry in redis
	err = redisServer.Set("SERVER-0.0.0.0:12345", string(serverdata))
	assert.NoError(t, err)

	// Create an incoming SessionUpdatePacket for the handler
	packet := core.SessionUpdatePacket{
		Sequence:   13,
		CustomerId: 13,
		SessionId:  13,
		UserHash:   13,
		PlatformId: core.PlatformUnknown,

		ConnectionType: core.ConnectionTypeUnknown,

		Tag:   13,
		Flags: 0,

		Flagged:          true,
		FallbackToDirect: true,
		TryBeforeYouBuy:  true,
		OnNetworkNext:    true,

		DirectMinRtt:     1.0,
		DirectMaxRtt:     2.0,
		DirectMeanRtt:    1.5,
		DirectJitter:     3.0,
		DirectPacketLoss: 4.0,
		NextMinRtt:       1.0,
		NextMaxRtt:       2.0,
		NextMeanRtt:      1.5,
		NextJitter:       3.0,
		NextPacketLoss:   4.0,

		KbpsUp:   10,
		KbpsDown: 20,

		PacketsLostServerToClient: 0,
		PacketsLostClientToServer: 0,

		NumNearRelays:       1,
		NearRelayIds:        []uint64{1},
		NearRelayMinRtt:     []float32{1.0},
		NearRelayMaxRtt:     []float32{2.0},
		NearRelayMeanRtt:    []float32{1.5},
		NearRelayJitter:     []float32{3.0},
		NearRelayPacketLoss: []float32{4.0},

		ServerAddress:        *addr,
		ClientAddress:        *addr,
		ClientRoutePublicKey: TestPublicKey,
		Signature:            make([]byte, 5),
	}
	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var buf bytes.Buffer
	incoming := transport.UDPPacket{
		SourceAddr: addr,
		Data:       data,
	}

	// Create and invoke the handler with the packet and from addr
	handler := transport.SessionUpdateHandlerFunc(redisClient, &ipStackClient)
	handler(&buf, &incoming)

	// Get the SessionEntry from redis based on the SessionUpdatePacket.Sequence number
	ds, err := redisServer.Get("SESSION-13")
	assert.NoError(t, err)

	// Create the expected SessionEntry
	expected := transport.SessionEntry{
		SessionID:  packet.SessionId,
		UserID:     packet.UserHash,
		PlatformID: packet.PlatformId,

		DirectRTT:        float64(packet.DirectMinRtt),
		DirectJitter:     float64(packet.DirectJitter),
		DirectPacketLoss: float64(packet.DirectPacketLoss),
		NextRTT:          float64(packet.NextMinRtt),
		NextJitter:       float64(packet.NextJitter),
		NextPacketLoss:   float64(packet.NextPacketLoss),

		ServerRoutePublicKey: serverentry.ServerRoutePublicKey,
		ServerPrivateAddr:    serverentry.ServerPrivateAddr,
		ServerAddr:           packet.ServerAddress,
		ClientAddr:           packet.ClientAddress,

		ConnectionType: packet.ConnectionType,

		GeoLocation: transport.IPStackResponse{
			IP:            "1.1.1.1",
			ContinentCode: "NA",
			CountryCode:   "US",
			RegionCode:    "NY",
			City:          "Troy",
			Latitude:      43.05036163330078,
			Longitude:     -73.75393676757812,
			Connection: struct {
				ASN int    `json:"asn"`
				ISP string `json:"isp"`
			}{
				ASN: 11351,
				ISP: "Charter Communications Inc",
			},
		},

		Tag:              packet.Tag,
		Flagged:          packet.Flagged,
		FallbackToDirect: packet.FallbackToDirect,
		OnNetworkNext:    packet.OnNetworkNext,

		VersionMajor: serverentry.VersionMajor,
		VersionMinor: serverentry.VersionMinor,
		VersionPatch: serverentry.VersionPatch,
	}

	var actual transport.SessionEntry
	actual.UnmarshalBinary([]byte(ds))

	// Finally compare that the SessionUpdatePacket created the right SessionEntry in redis
	assert.Equal(t, expected, actual)

	// Need to test the buf SessionResponsePacket
	// log.Fatal(buf.Bytes())
}
