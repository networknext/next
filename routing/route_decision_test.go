package routing_test

import (
	"testing"

	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestDecideUpgradeRTT(t *testing.T) {
	// Test if a route gets upgraded to network next
	predictedStats := routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	directStats := routing.Stats{
		RTT:        50,
		Jitter:     0,
		PacketLoss: 0,
	}

	rttThreshold := float64(routing.DefaultRoutingRulesSettings.RTTThreshold)
	routeDecisionFunc := routing.DecideUpgradeRTT(rttThreshold)

	assert.Equal(
		t,
		routing.Decision{true, routing.DecisionRTTReduction},
		routeDecisionFunc(routing.Decision{false, routing.DecisionNoChange}, predictedStats, routing.Stats{}, directStats, &routing.EmptyDecisionMetrics),
	)

	// Now test if the route is left alone

	predictedStats.RTT = directStats.RTT
	assert.Equal(
		t,
		routing.Decision{false, routing.DecisionNoChange},
		routeDecisionFunc(routing.Decision{false, routing.DecisionNoChange}, predictedStats, routing.Stats{}, directStats, &routing.EmptyDecisionMetrics),
	)
}

func TestDecideDowngradeRTT(t *testing.T) {
	// Test if a route stays on the network next route
	predictedStats := routing.Stats{
		RTT:        40,
		Jitter:     0,
		PacketLoss: 0,
	}

	directStats := routing.Stats{
		RTT:        35,
		Jitter:     0,
		PacketLoss: 0,
	}

	rttHyteresis := float64(routing.DefaultRoutingRulesSettings.RTTHysteresis)
	routeDecisionFunc := routing.DecideDowngradeRTT(rttHyteresis)

	decision := routing.Decision{true, routing.DecisionNoChange}
	decision = routeDecisionFunc(decision, predictedStats, routing.Stats{}, directStats, &routing.EmptyDecisionMetrics)

	assert.Equal(t, routing.Decision{true, routing.DecisionNoChange}, decision)

	// Now test to see if the route gets downgraded to a direct route due to RTT
	predictedStats.RTT = directStats.RTT + rttHyteresis + 1.0

	decision = routeDecisionFunc(decision, predictedStats, routing.Stats{}, directStats, &routing.EmptyDecisionMetrics)
	assert.Equal(t, routing.Decision{false, routing.DecisionRTTIncrease}, decision)

	// Now test if a direct route is given
	decision = routeDecisionFunc(decision, predictedStats, routing.Stats{}, directStats, &routing.EmptyDecisionMetrics)
	assert.Equal(t, routing.Decision{false, routing.DecisionNoChange}, decision)
}

func TestDecideVeto(t *testing.T) {
	// Test if a route is vetoed for RTT increases
	lastNextStats := routing.Stats{
		RTT:        60,
		Jitter:     0,
		PacketLoss: 0,
	}

	directStats := routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	rttVeto := float64(routing.DefaultRoutingRulesSettings.RTTVeto)
	routeDecisionFunc := routing.DecideVeto(rttVeto, false, false)

	decision := routing.Decision{true, routing.DecisionNoChange}
	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats, &routing.EmptyDecisionMetrics)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT}, decision)

	// Now test for yolo reason
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideVeto(rttVeto, false, true)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats, &routing.EmptyDecisionMetrics)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT | routing.DecisionVetoYOLO}, decision)

	// Now test if the route is vetoed for packet loss increases
	lastNextStats.RTT = directStats.RTT
	lastNextStats.PacketLoss = directStats.PacketLoss + 1
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideVeto(rttVeto, true, false)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats, &routing.EmptyDecisionMetrics)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoPacketLoss}, decision)

	// Now test for yolo reason
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideVeto(rttVeto, true, true)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats, &routing.EmptyDecisionMetrics)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoPacketLoss | routing.DecisionVetoYOLO}, decision)

	// Test if route isn't vetoed
	lastNextStats.PacketLoss = directStats.PacketLoss
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideVeto(rttVeto, true, true)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats, &routing.EmptyDecisionMetrics)
	assert.Equal(t, routing.Decision{true, routing.DecisionNoChange}, decision)

	// Test if direct route isn't changed
	decision.OnNetworkNext = false

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats, &routing.EmptyDecisionMetrics)
	assert.Equal(t, routing.Decision{false, routing.DecisionNoChange}, decision)
}
