package transport_test

import (
	"context"
	"net"
	"testing"

	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestBuildRouteRequest(t *testing.T) {
	expected := &billing.RouteRequest{
		BuyerID:          transport.NewEntityID("Buyer", 1),
		SessionID:        2,
		UserHash:         3,
		PlatformID:       4,
		DirectRTT:        5,
		DirectJitter:     6,
		DirectPacketLoss: 7,
		NextRTT:          8,
		NextJitter:       9,
		NextPacketLoss:   10,
		ClientIpAddress: transport.NewBillingAddress(net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 11,
		}),
		ServerIpAddress: transport.NewBillingAddress(net.UDPAddr{
			IP:   net.ParseIP("127.0.0.2"),
			Port: 12,
		}),
		ServerPrivateIpAddress: transport.NewBillingAddress(net.UDPAddr{
			IP:   net.ParseIP("127.0.0.3"),
			Port: 13,
		}),
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
		ServerRoutePublicKey: TestBuyersServerPublicKey[:],
		Tag:                  14,
		NearRelays: []*billing.NearRelay{
			&billing.NearRelay{
				RelayID:    transport.NewEntityID("Relay", 100),
				RTT:        1,
				Jitter:     2,
				PacketLoss: 3,
			},
			&billing.NearRelay{
				RelayID:    transport.NewEntityID("Relay", 200),
				RTT:        4,
				Jitter:     5,
				PacketLoss: 6,
			}, &billing.NearRelay{
				RelayID:    transport.NewEntityID("Relay", 300),
				RTT:        7,
				Jitter:     8,
				PacketLoss: 9,
			},
		},
		IssuedNearRelays: []*billing.IssuedNearRelay{
			&billing.IssuedNearRelay{
				Index:   0,
				RelayID: transport.NewEntityID("Relay", 100),
				RelayIpAddress: transport.NewBillingAddress(net.UDPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 1000,
				}),
			},
			&billing.IssuedNearRelay{
				Index:   1,
				RelayID: transport.NewEntityID("Relay", 200),
				RelayIpAddress: transport.NewBillingAddress(net.UDPAddr{
					IP:   net.ParseIP("127.0.0.2"),
					Port: 2000,
				}),
			},
			&billing.IssuedNearRelay{
				Index:   2,
				RelayID: transport.NewEntityID("Relay", 300),
				RelayIpAddress: transport.NewBillingAddress(net.UDPAddr{
					IP:   net.ParseIP("127.0.0.3"),
					Port: 3000,
				}),
			},
		},
		ConnectionType: billing.SessionConnectionType_SESSION_CONNECTION_TYPE_WIFI,
		Location: &billing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  42.7328415,
			Longitude: -73.6835691,
		},
		DatacenterID:              transport.NewEntityID("Datacenter", 15),
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

	buyer := &routing.Buyer{
		ID: 1,
	}

	updatePacket := transport.SessionUpdatePacket{
		SessionID:        2,
		UserHash:         3,
		PlatformID:       4,
		DirectMinRTT:     5,
		DirectJitter:     6,
		DirectPacketLoss: 7,
		NextMinRTT:       8,
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
		NumNearRelays:        3,
		NearRelayIDs:         []uint64{100, 200, 300},
		NearRelayMinRTT:      []float32{1, 4, 7},
		NearRelayJitter:      []float32{2, 5, 8},
		NearRelayPacketLoss:  []float32{3, 6, 9},
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
		Continent: "NA",
		Country:   "US",
		Region:    "NY",
		City:      "Troy",
		Latitude:  42.7328415,
		Longitude: -73.6835691,
	}

	clientRelays := []routing.Relay{
		routing.Relay{
			ID: 100,
			Addr: net.UDPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 1000,
			},
		},
		routing.Relay{
			ID: 200,
			Addr: net.UDPAddr{
				IP:   net.ParseIP("127.0.0.2"),
				Port: 2000,
			},
		},
		routing.Relay{
			ID: 300,
			Addr: net.UDPAddr{
				IP:   net.ParseIP("127.0.0.3"),
				Port: 3000,
			},
		},
	}

	storer := storage.InMemory{}
	for _, clientRelay := range clientRelays {
		storer.AddRelay(context.Background(), clientRelay)
	}

	actual := transport.NewRouteRequest(updatePacket, buyer, serverData, location, &storer, clientRelays)

	assert.Equal(t, expected, actual)
}
