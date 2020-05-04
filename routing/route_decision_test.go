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
		routeDecisionFunc(routing.Decision{false, routing.DecisionNoChange}, predictedStats, routing.Stats{}, directStats),
	)

	// Now test if the route is left alone

	predictedStats.RTT = directStats.RTT
	assert.Equal(
		t,
		routing.Decision{false, routing.DecisionNoChange},
		routeDecisionFunc(routing.Decision{false, routing.DecisionNoChange}, predictedStats, routing.Stats{}, directStats),
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
	routeDecisionFunc := routing.DecideDowngradeRTT(rttHyteresis, false)

	decision := routing.Decision{true, routing.DecisionNoChange}
	decision = routeDecisionFunc(decision, predictedStats, routing.Stats{}, directStats)

	assert.Equal(t, routing.Decision{true, routing.DecisionNoChange}, decision)

	// Now test to see if the route gets downgraded to a direct route due to RTT
	predictedStats.RTT = directStats.RTT + rttHyteresis + 1.0

	decision = routeDecisionFunc(decision, predictedStats, routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionRTTIncrease}, decision)

	// Now test if a direct route is given
	decision = routeDecisionFunc(decision, predictedStats, routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionNoChange}, decision)

	// Now test if the route is vetoed with YOLO enabled
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideDowngradeRTT(rttHyteresis, true)
	decision = routeDecisionFunc(decision, predictedStats, routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT | routing.DecisionVetoYOLO}, decision)
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
	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT}, decision)

	// Now test for yolo reason
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideVeto(rttVeto, false, true)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT | routing.DecisionVetoYOLO}, decision)

	// Now test if the route is vetoed for packet loss increases
	lastNextStats.RTT = directStats.RTT
	lastNextStats.PacketLoss = directStats.PacketLoss + 1
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideVeto(rttVeto, true, false)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoPacketLoss}, decision)

	// Now test for yolo reason
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideVeto(rttVeto, true, true)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoPacketLoss | routing.DecisionVetoYOLO}, decision)

	// Test if route isn't vetoed
	lastNextStats.PacketLoss = directStats.PacketLoss
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideVeto(rttVeto, true, true)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionNoChange}, decision)

	// Test if route was changed to direct from another function, but the RTT increase was so severe that it should be vetoed
	lastNextStats.RTT = 60
	decision = routing.Decision{OnNetworkNext: false, Reason: routing.DecisionRTTIncrease}
	routeDecisionFunc = routing.DecideVeto(rttVeto, true, false)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT}, decision)

	// Now with yolo
	decision = routing.Decision{OnNetworkNext: false, Reason: routing.DecisionRTTIncrease}
	routeDecisionFunc = routing.DecideVeto(rttVeto, true, true)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT | routing.DecisionVetoYOLO}, decision)

	// Test if direct route isn't changed
	decision.OnNetworkNext = false

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionNoChange}, decision)
}

func TestDecideCommitted(t *testing.T) {
	decision := routing.Decision{false, routing.DecisionNoChange}
	committedRouteCount := uint64(0)
	routeCountThreshold := uint64(3)

	// Test if a route is unchanged when try before you buy is disabled
	routeDecisionFunc := routing.DecideCommitted(false, &committedRouteCount, routeCountThreshold)

	decision = routeDecisionFunc(decision, routing.Stats{}, routing.Stats{}, routing.Stats{})
	assert.Equal(t, routing.Decision{false, routing.DecisionNoChange}, decision)

	// Now test if the route has try before you buy enabled but the route is already direct
	routeDecisionFunc = routing.DecideCommitted(true, &committedRouteCount, routeCountThreshold)

	decision = routeDecisionFunc(decision, routing.Stats{}, routing.Stats{}, routing.Stats{})
	assert.Equal(t, routing.Decision{false, routing.DecisionNoChange}, decision)

	// Test if the route was an improvement, but not enough routes have been observed yet
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideCommitted(true, &committedRouteCount, routeCountThreshold)

	decision = routeDecisionFunc(decision, routing.Stats{}, routing.Stats{}, routing.Stats{})
	assert.Equal(t, routing.Decision{false, routing.DecisionCommittedPending}, decision)
	assert.Equal(t, uint64(1), committedRouteCount)

	// Test if the route was an improvement, and the SDK should finally be committed to the route
	decision.OnNetworkNext = true
	committedRouteCount = routeCountThreshold
	routeDecisionFunc = routing.DecideCommitted(true, &committedRouteCount, routeCountThreshold)

	decision = routeDecisionFunc(decision, routing.Stats{}, routing.Stats{}, routing.Stats{})
	assert.Equal(t, routing.Decision{true, routing.DecisionNoChange}, decision)
}
