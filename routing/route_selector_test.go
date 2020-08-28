package routing_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/networknext/backend/routing"
)

func TestSelectRoutesByRandomDestRelay(t *testing.T) {
	routes := []routing.Route{
		{
			NumRelays: 3,
			RelayIDs: [routing.MaxRelays]uint64{
				1, 2, 3,
			},
			Stats: routing.Stats{
				RTT:        5,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			NumRelays: 4,
			RelayIDs: [routing.MaxRelays]uint64{
				4, 2, 5, 3,
			},
			Stats: routing.Stats{
				RTT:        4.7,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			NumRelays: 2,
			RelayIDs: [routing.MaxRelays]uint64{
				1, 3,
			},
			Stats: routing.Stats{
				RTT:        5.2,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
	}

	randsrc := rand.NewSource(0)
	selectedRoutes := routing.SelectRandomRoute(randsrc)(routes)

	assert.Equal(t, 1, len(selectedRoutes))
	assert.Equal(t, float64(5), selectedRoutes[0].Stats.RTT)
}

func TestSelectRandomRoute(t *testing.T) {
	routes := []routing.Route{
		{
			Stats: routing.Stats{
				RTT:        5,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			Stats: routing.Stats{
				RTT:        4.7,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			Stats: routing.Stats{
				RTT:        5.2,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			Stats: routing.Stats{
				RTT:        8,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
	}

	randsrc := rand.NewSource(0)
	selectedRoutes := routing.SelectRandomRoute(randsrc)(routes)

	assert.Equal(t, 1, len(selectedRoutes))
	assert.Equal(t, float64(5.2), selectedRoutes[0].Stats.RTT)
}
