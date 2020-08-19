package routing_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/networknext/backend/routing"
)

func TestSelectBestRTT(t *testing.T) {
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
				RTT:        1,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			Stats: routing.Stats{
				RTT:        3,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
	}

	selectedRoutes := routing.SelectBestRTT()(routes)

	assert.Equal(t, 1, len(selectedRoutes))
	assert.Equal(t, float64(1), selectedRoutes[0].Stats.RTT)
}

func TestSelectAcceptableRoutesFromBestRTT(t *testing.T) {
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

	selectedRoutes := routing.SelectAcceptableRoutesFromBestRTT(.5)(routes)

	assert.Equal(t, 3, len(selectedRoutes))
	assert.Equal(t, float64(5), selectedRoutes[0].Stats.RTT)
	assert.Equal(t, float64(4.7), selectedRoutes[1].Stats.RTT)
	assert.Equal(t, float64(5.2), selectedRoutes[2].Stats.RTT)
}

func TestSelectContainsRouteHash(t *testing.T) {
	routes := []routing.Route{
		{
			RelayIDs: []uint64{
				1, 2, 3,
			},
		},
		{
			RelayIDs: []uint64{
				4, 1, 2,
			},
		},
	}

	selectedRoutes := routing.SelectContainsRouteHash(routes[0].Hash64())(routes)

	assert.Equal(t, 1, len(selectedRoutes))
	assert.Equal(t, uint64(1), selectedRoutes[0].RelayIDs[0])
	assert.Equal(t, uint64(2), selectedRoutes[0].RelayIDs[1])
	assert.Equal(t, uint64(3), selectedRoutes[0].RelayIDs[2])
}

func TestSelectUnencumberedRoutes(t *testing.T) {
	routes := []routing.Route{
		{
			RelayIDs: []uint64{
				1, 2, 3,
			},
			RelaySessions: []uint32{
				150, 450, 300,
			},
			RelayMaxSessions: []uint32{
				3000, 3000, 3000,
			},
			Stats: routing.Stats{
				RTT:        5,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			RelayIDs: []uint64{
				4, 2, 5, 3,
			},
			RelaySessions: []uint32{
				3000,
				450,
				4500,
				300,
			},
			RelayMaxSessions: []uint32{
				3000, 3000, 6000, 3000,
			},
			Stats: routing.Stats{
				RTT:        4.7,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
	}

	selectedRoutes := routing.SelectUnencumberedRoutes(0.8)(routes)

	assert.Equal(t, 1, len(selectedRoutes))
	assert.Equal(t, routes[0], selectedRoutes[0])
}

func TestSelectRoutesByRandomDestRelay(t *testing.T) {
	routes := []routing.Route{
		{
			RelayIDs: []uint64{
				1, 2, 3,
			},
			Stats: routing.Stats{
				RTT:        5,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			RelayIDs: []uint64{
				4, 2, 5, 3,
			},
			Stats: routing.Stats{
				RTT:        4.7,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			RelayIDs: []uint64{
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
