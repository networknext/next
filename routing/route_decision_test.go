package routing_test

import (
	"testing"

	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestDecideUpgradeRTT(t *testing.T) {
	// Test if a route gets upgraded to network next
	predictedStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	directStats := &routing.Stats{
		RTT:        50,
		Jitter:     0,
		PacketLoss: 0,
	}

	stayOnNN := false
	decisionReason := routing.DecisionNone
	rttThreshold := float64(routing.DefaultRoutingRulesSettings.RTTThreshold)

	routeDecision := routing.DecideUpgradeRTT(rttThreshold)
	stayOnNN, decisionReason = routeDecision(stayOnNN, predictedStats, nil, directStats)

	assert.True(t, stayOnNN)
	assert.Equal(t, routing.DecisionRTTReduction, decisionReason)

	// Now test if the route is left alone
	stayOnNN = false
	predictedStats.RTT = directStats.RTT

	stayOnNN, decisionReason = routeDecision(stayOnNN, predictedStats, nil, directStats)

	assert.False(t, stayOnNN)
	assert.Equal(t, routing.DecisionNone, decisionReason)
}

func TestDecideDowngradeRTT(t *testing.T) {
	// Test if a route stays on the network next route
	predictedStats := &routing.Stats{
		RTT:        40,
		Jitter:     0,
		PacketLoss: 0,
	}

	directStats := &routing.Stats{
		RTT:        35,
		Jitter:     0,
		PacketLoss: 0,
	}

	stayOnNN := true
	decisionReason := routing.DecisionNone
	rttHyteresis := float64(routing.DefaultRoutingRulesSettings.RTTHysteresis)

	routeDecision := routing.DecideDowngradeRTT(rttHyteresis)
	stayOnNN, decisionReason = routeDecision(stayOnNN, predictedStats, nil, directStats)

	assert.True(t, stayOnNN)
	assert.Equal(t, routing.DecisionRTTReduction, decisionReason)

	// Now test to see if the route gets downgraded to a direct route due to RTT
	predictedStats.RTT = directStats.RTT + rttHyteresis + 1.0

	stayOnNN, decisionReason = routeDecision(stayOnNN, predictedStats, nil, directStats)

	assert.False(t, stayOnNN)
	assert.Equal(t, routing.DecisionNone, decisionReason)

	// Now test if a direct route is given
	stayOnNN, decisionReason = routeDecision(stayOnNN, predictedStats, nil, directStats)

	assert.False(t, stayOnNN)
	assert.Equal(t, routing.DecisionNone, decisionReason)
}

func TestDecideVeto(t *testing.T) {
	// Test if a route is vetoed for RTT increases
	lastNextStats := &routing.Stats{
		RTT:        60,
		Jitter:     0,
		PacketLoss: 0,
	}

	directStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	stayOnNN := true
	decisionReason := routing.DecisionNone
	rttVeto := float64(routing.DefaultRoutingRulesSettings.RTTVeto)

	routeDecision := routing.DecideVeto(rttVeto, false, false)
	stayOnNN, decisionReason = routeDecision(stayOnNN, nil, lastNextStats, directStats)

	assert.False(t, stayOnNN)
	assert.Equal(t, routing.DecisionVetoRTT, decisionReason)

	// Now test for yolo reason
	stayOnNN = true
	routeDecision = routing.DecideVeto(rttVeto, false, true)

	stayOnNN, decisionReason = routeDecision(stayOnNN, nil, lastNextStats, directStats)

	assert.False(t, stayOnNN)
	assert.Equal(t, routing.DecisionVetoRTT|routing.DecisionVetoYOLO, decisionReason)

	// Now test if the route is vetoed for packet loss increases
	stayOnNN = true
	lastNextStats.RTT = directStats.RTT
	lastNextStats.PacketLoss = directStats.PacketLoss + 1
	routeDecision = routing.DecideVeto(rttVeto, true, false)

	stayOnNN, decisionReason = routeDecision(stayOnNN, nil, lastNextStats, directStats)

	assert.False(t, stayOnNN)
	assert.Equal(t, routing.DecisionVetoPacketLoss, decisionReason)

	// Now test for yolo reason
	stayOnNN = true
	routeDecision = routing.DecideVeto(rttVeto, true, true)

	stayOnNN, decisionReason = routeDecision(stayOnNN, nil, lastNextStats, directStats)

	assert.False(t, stayOnNN)
	assert.Equal(t, routing.DecisionVetoPacketLoss|routing.DecisionVetoYOLO, decisionReason)

	// Test if route isn't vetoed
	stayOnNN = true
	lastNextStats.PacketLoss = directStats.PacketLoss
	routeDecision = routing.DecideVeto(rttVeto, true, true)

	stayOnNN, decisionReason = routeDecision(stayOnNN, nil, lastNextStats, directStats)

	assert.True(t, stayOnNN)
	assert.Equal(t, routing.DecisionNone, decisionReason)

	// Test if direct route isn't changed
	stayOnNN = false

	stayOnNN, decisionReason = routeDecision(stayOnNN, nil, lastNextStats, directStats)

	assert.False(t, stayOnNN)
	assert.Equal(t, routing.DecisionNone, decisionReason)
}
