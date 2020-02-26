package transport_test

import (
	"net"
	"testing"

	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestBuildRouteRequest(t *testing.T) {
	expected := billing.RouteRequest{
		BuyerId:          billing.MakeEntityID("Buyer", 1),
		SessionId:        2,
		UserHash:         3,
		PlatformId:       4,
		DirectRtt:        5,
		DirectJitter:     6,
		DirectPacketLoss: 7,
		NextRtt:          8,
		NextJitter:       9,
		NextPacketLoss:   10,
		ClientIpAddress: billing.UdpAddrToAddress(net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 11,
		}),
		ServerIpAddress: billing.UdpAddrToAddress(net.UDPAddr{
			IP:   net.ParseIP("127.0.0.2"),
			Port: 12,
		}),
		ServerPrivateIpAddress: billing.UdpAddrToAddress(net.UDPAddr{
			IP:   net.ParseIP("127.0.0.3"),
			Port: 13,
		}),
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
		ServerRoutePublicKey: TestBuyersServerPublicKey[:],
		Tag:                  14,
		//NearRelays:           make([]*billing.NearRelay, 0),       TODO: come back and populate these
		//IssuedNearRelays:     make([]*billing.IssuedNearRelay, 0), TODO: come back and populate these
		ConnectionType: billing.SessionConnectionType_SESSION_CONNECTION_TYPE_WIFI,
		Location: &billing.Location{
			// Isp: TODO,
			// Asn: TODO,
			// CountryCode: TODO,
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  42.7328415,
			Longitude: -73.6835691,
		},
		DatacenterId:              billing.MakeEntityID("Datacenter", 15),
		SequenceNumber:            16,
		FallbackToDirect:          true,
		VersionMajor:              17,
		VersionMinor:              18,
		VersionPatch:              19,
		UsageKbpsUp:               20,
		UsageKbpsDown:             21,
		Flagged:                   true,
		TryBeforeYouBuy:           true,
		OnNetworkNext:             true,
		PacketsLostClientToServer: 22,
		PacketsLostServerToClient: 23,
		FallbackFlags:             24,
		Committed:                 true,
	}

	buyer := routing.Buyer{
		ID: 1,
	}

	updatePacket := transport.SessionUpdatePacket{
		SessionId:        2,
		UserHash:         3,
		PlatformId:       4,
		DirectMinRtt:     5,
		DirectJitter:     6,
		DirectPacketLoss: 7,
		NextMinRtt:       8,
		NextJitter:       9,
		NextPacketLoss:   10,
		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 11,
		},
		ServerAddress: net.UDPAddr{
			IP:   net.ParseIP("127.0.0.2"),
			Port: 12,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
		Tag:                  14,
		ConnectionType:       2, // billing.SessionConnectionType_SESSION_CONNECTION_TYPE_WIFI
		Sequence:             16,
		FallbackToDirect:     true,
		//Version:                 This comes from server cache entry
		KbpsUp:                    20,
		KbpsDown:                  21,
		Flagged:                   true,
		TryBeforeYouBuy:           true,
		OnNetworkNext:             true,
		PacketsLostClientToServer: 22,
		PacketsLostServerToClient: 23,
		Flags:                     24,
		Committed:                 true,
	}

	serverData := transport.ServerCacheEntry{
		Datacenter: routing.Datacenter{
			ID: 15,
		},
		Server: routing.Server{
			Addr: net.UDPAddr{
				IP:   net.ParseIP("127.0.0.3"),
				Port: 13,
			},
			PublicKey: TestBuyersServerPublicKey[:],
		},
		SDKVersion: transport.SDKVersion{17, 18, 19},
	}

	location := routing.Location{
		// Isp: TODO,
		// Asn: TODO,
		// CountryCode: TODO,
		Continent: "NA",
		Country:   "US",
		Region:    "NY",
		City:      "Troy",
		Latitude:  42.7328415,
		Longitude: -73.6835691,
	}

	storer := storage.InMemory{}

	clientRelays := []routing.Relay{}

	actual := transport.BuildRouteRequest(updatePacket, buyer, serverData, location, &storer, clientRelays)

	assert.Equal(t, expected, actual)
}
