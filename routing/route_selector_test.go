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
		{
			Stats: routing.Stats{
				RTT:        5,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
	}

	selectedRoutes := routing.SelectBestRTT()(routes)

	assert.Equal(t, 1, len(selectedRoutes))
	assert.Equal(t, float64(1), selectedRoutes[0].Stats.RTT)

	for _, route := range selectedRoutes {
		assert.NotEmpty(t, route)
	}
}

func TestSelectAcceptableRoutesFromBestRTT(t *testing.T) {
	routes := []routing.Route{
		{
			Stats: routing.Stats{
				RTT:        4.7,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			Stats: routing.Stats{
				RTT:        5,
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

	selectedRoutes := routing.SelectAcceptableRoutesFromBestRTT(0.5)(routes)

	assert.Equal(t, 3, len(selectedRoutes))
	assert.Equal(t, float64(4.7), selectedRoutes[0].Stats.RTT)
	assert.Equal(t, float64(5), selectedRoutes[1].Stats.RTT)
	assert.Equal(t, float64(5.2), selectedRoutes[2].Stats.RTT)

	for _, route := range selectedRoutes {
		assert.NotEmpty(t, route)
	}
}

func TestSelectContainsRouteHash(t *testing.T) {
	routes := []routing.Route{
		{
			Relays: []routing.Relay{
				{ID: 1}, {ID: 2}, {ID: 3},
			},
		},
		{
			Relays: []routing.Relay{
				{ID: 4}, {ID: 1}, {ID: 2},
			},
		},
	}

	selectedRoutes := routing.SelectContainsRouteHash(routes[0].Hash64())(routes)

	assert.Equal(t, 1, len(selectedRoutes))
	assert.Equal(t, uint64(1), selectedRoutes[0].Relays[0].ID)
	assert.Equal(t, uint64(2), selectedRoutes[0].Relays[1].ID)
	assert.Equal(t, uint64(3), selectedRoutes[0].Relays[2].ID)

	for _, route := range selectedRoutes {
		assert.NotEmpty(t, route)
	}
}

func TestSelectUnencumberedRoutes(t *testing.T) {
	routes := []routing.Route{
		{
			Relays: []routing.Relay{
				{
					ID: 1,
					TrafficStats: routing.RelayTrafficStats{
						SessionCount: 150,
					},
					MaxSessions: 3000,
				},
				{
					ID: 2,
					TrafficStats: routing.RelayTrafficStats{
						SessionCount: 450,
					},
					MaxSessions: 3000,
				},
				{
					ID: 3,
					TrafficStats: routing.RelayTrafficStats{
						SessionCount: 300,
					},
					MaxSessions: 3000,
				},
			},
			Stats: routing.Stats{
				RTT:        5,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			Relays: []routing.Relay{
				{
					ID: 4,
					TrafficStats: routing.RelayTrafficStats{
						SessionCount: 3000,
					},
					MaxSessions: 3000,
				},
				{
					ID: 2,
					TrafficStats: routing.RelayTrafficStats{
						SessionCount: 450,
					},
					MaxSessions: 3000,
				},
				{
					ID: 5,
					TrafficStats: routing.RelayTrafficStats{
						SessionCount: 4500,
					},
					MaxSessions: 6000,
				},
				{
					ID: 3,
					TrafficStats: routing.RelayTrafficStats{
						SessionCount: 300,
					},
					MaxSessions: 3000,
				},
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

	for _, route := range selectedRoutes {
		assert.NotEmpty(t, route)
	}
}

func TestSelectRoutesByRandomDestRelay(t *testing.T) {
	routes := []routing.Route{
		{
			Relays: []routing.Relay{
				{ID: 4}, {ID: 2}, {ID: 5}, {ID: 3},
			},
			Stats: routing.Stats{
				RTT:        4.7,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			Relays: []routing.Relay{
				{ID: 1}, {ID: 2}, {ID: 3},
			},
			Stats: routing.Stats{
				RTT:        5,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			Relays: []routing.Relay{
				{ID: 1}, {ID: 3},
			},
			Stats: routing.Stats{
				RTT:        5.2,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
	}

	randsrc := rand.NewSource(0)
	selectedRoutes := routing.SelectRoutesByRandomDestRelay(randsrc)(routes)

	assert.Equal(t, 3, len(selectedRoutes))
	assert.Equal(t, float64(4.7), selectedRoutes[0].Stats.RTT)
	assert.Equal(t, float64(5), selectedRoutes[1].Stats.RTT)
	assert.Equal(t, float64(5.2), selectedRoutes[2].Stats.RTT)

	for _, route := range selectedRoutes {
		assert.NotEmpty(t, route)
	}
}

func TestSelectRandomRoute(t *testing.T) {
	routes := []routing.Route{
		{
			Stats: routing.Stats{
				RTT:        4.7,
				Jitter:     0,
				PacketLoss: 0,
			},
		},
		{
			Stats: routing.Stats{
				RTT:        5,
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

	for _, route := range selectedRoutes {
		assert.NotEmpty(t, route)
	}
}
