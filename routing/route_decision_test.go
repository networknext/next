package routing_test

import (
	"testing"

	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestDecideUpgradeRTT(t *testing.T) {
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

	// Test if multipath is enabled
	assert.Equal(
		t,
		routing.Decision{true, routing.DecisionRTTReductionMultipath},
		routeDecisionFunc(routing.Decision{true, routing.DecisionRTTReductionMultipath}, predictedStats, routing.Stats{}, directStats),
	)

	// Now test if a route gets upgraded to network next
	assert.Equal(
		t,
		routing.Decision{true, routing.DecisionRTTReduction},
		routeDecisionFunc(routing.Decision{}, predictedStats, routing.Stats{}, directStats),
	)

	// Now test if the route is left alone

	predictedStats.RTT = directStats.RTT
	assert.Equal(
		t,
		routing.Decision{},
		routeDecisionFunc(routing.Decision{}, predictedStats, routing.Stats{}, directStats),
	)
}

func TestDecideDowngradeRTT(t *testing.T) {
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

	// Test if multipath is enabled
	decision := routing.Decision{true, routing.DecisionRTTReductionMultipath}
	decision = routeDecisionFunc(decision, predictedStats, routing.Stats{}, directStats)

	assert.Equal(t, routing.Decision{true, routing.DecisionRTTReductionMultipath}, decision)

	// Now test if a route stays on the network next route
	decision = routing.Decision{true, routing.DecisionNoReason}
	decision = routeDecisionFunc(decision, predictedStats, routing.Stats{}, directStats)

	assert.Equal(t, routing.Decision{true, routing.DecisionNoReason}, decision)

	// Now test to see if the route gets downgraded to a direct route due to RTT
	predictedStats.RTT = directStats.RTT + rttHyteresis + 1.0

	decision = routeDecisionFunc(decision, predictedStats, routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionRTTHysteresis}, decision)

	// Now test if a direct route is given
	decision = routing.Decision{}
	decision = routeDecisionFunc(decision, predictedStats, routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{}, decision)

	// Now test if the route is vetoed with YOLO enabled
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideDowngradeRTT(rttHyteresis, true)
	decision = routeDecisionFunc(decision, predictedStats, routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionRTTHysteresis | routing.DecisionVetoYOLO}, decision)
}

func TestDecideVeto(t *testing.T) {
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

	// Test if multipath is enabled
	decision := routing.Decision{true, routing.DecisionRTTReductionMultipath}
	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionRTTReductionMultipath}, decision)

	// Now test if a route is vetoed for RTT increases
	decision = routing.Decision{true, routing.DecisionNoReason}
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
	decision = routing.Decision{true, routing.DecisionNoReason}
	routeDecisionFunc = routing.DecideVeto(rttVeto, true, true)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionNoReason}, decision)

	// Test if route was changed to direct from another function, but the RTT increase was so severe that it should be vetoed
	lastNextStats.RTT = 60
	decision = routing.Decision{OnNetworkNext: false, Reason: routing.DecisionRTTHysteresis}
	routeDecisionFunc = routing.DecideVeto(rttVeto, true, false)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT}, decision)

	// Now with yolo
	decision = routing.Decision{OnNetworkNext: false, Reason: routing.DecisionRTTHysteresis}
	routeDecisionFunc = routing.DecideVeto(rttVeto, true, true)

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT | routing.DecisionVetoYOLO}, decision)

	// Test if direct route isn't changed
	decision = routing.Decision{false, routing.DecisionNoReason}

	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{}, decision)
}

func TestDecideCommitted(t *testing.T) {
	lastNextStats := routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	directStats := routing.Stats{
		RTT:        60,
		Jitter:     0,
		PacketLoss: 0,
	}

	maxSlices := uint8(routing.DefaultRoutingRulesSettings.TryBeforeYouBuyMaxSlices)

	var commitPending bool
	var observedSliceCounter uint8
	var committed bool

	routeDecisionFunc := routing.DecideCommitted(false, maxSlices, false, &commitPending, &observedSliceCounter, &committed)

	// Test if multipath is enabled
	decision := routing.Decision{true, routing.DecisionRTTReductionMultipath}
	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionRTTReductionMultipath}, decision)
	assert.Equal(t, false, commitPending)
	assert.Equal(t, uint8(0), observedSliceCounter)
	assert.Equal(t, true, committed)

	// Check direct routes aren't affected
	decision = routing.Decision{}
	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{}, decision)
	assert.Equal(t, false, commitPending)
	assert.Equal(t, uint8(0), observedSliceCounter)
	assert.Equal(t, false, committed)

	// Check if a slice newly on NN is "initialized" properly
	decision.OnNetworkNext = true
	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionNoReason}, decision)
	assert.Equal(t, true, commitPending)
	assert.Equal(t, uint8(0), observedSliceCounter)
	assert.Equal(t, false, committed)

	// Now check the case where the route is better
	routeDecisionFunc = routing.DecideCommitted(true, maxSlices, false, &commitPending, &observedSliceCounter, &committed)
	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionNoReason}, decision)
	assert.Equal(t, false, commitPending)
	assert.Equal(t, uint8(0), observedSliceCounter)
	assert.Equal(t, true, committed)

	// Now check the case where the route isn't bad enough yet to veto
	commitPending = true
	lastNextStats.RTT = 65
	decision.OnNetworkNext = true
	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionNoReason}, decision)
	assert.Equal(t, true, commitPending)
	assert.Equal(t, uint8(1), observedSliceCounter)
	assert.Equal(t, false, committed)

	// Now check the case where the route has taken too long to decide and should veto
	observedSliceCounter = maxSlices
	decision.OnNetworkNext = true
	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoCommit}, decision)
	assert.Equal(t, false, commitPending)
	assert.Equal(t, uint8(0), observedSliceCounter)
	assert.Equal(t, false, committed)

	// Now check the case where the route has taken too long to decide and should veto and yolo is enabled
	decision.OnNetworkNext = true
	commitPending = true
	observedSliceCounter = maxSlices
	routeDecisionFunc = routing.DecideCommitted(true, maxSlices, true, &commitPending, &observedSliceCounter, &committed)
	decision = routeDecisionFunc(decision, routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoCommit | routing.DecisionVetoYOLO}, decision)
	assert.Equal(t, false, commitPending)
	assert.Equal(t, uint8(0), observedSliceCounter)
	assert.Equal(t, false, committed)
}

func TestDecideMultipath(t *testing.T) {
	rttThreshold := float64(routing.LocalRoutingRulesSettings.RTTThreshold)

	// Test if multipath isn't enabled
	routeDecisionFunc := routing.DecideMultipath(false, false, false, rttThreshold)
	predictedNNStats := routing.Stats{RTT: 30}
	directStats := routing.Stats{RTT: 60}
	decision := routing.Decision{}
	decision = routeDecisionFunc(decision, predictedNNStats, routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{}, decision)

	// Test if multipath is already active
	routeDecisionFunc = routing.DecideMultipath(true, true, true, rttThreshold)
	decision = routing.Decision{true, routing.DecisionRTTReductionMultipath}
	decision = routeDecisionFunc(decision, predictedNNStats, routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionRTTReductionMultipath}, decision)

	// Test when multipath reason is RTT reduction
	decision = routing.Decision{}
	decision = routeDecisionFunc(decision, predictedNNStats, routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionRTTReductionMultipath}, decision)

	// Test when multipath reason is high jitter
	routeDecisionFunc = routing.DecideMultipath(true, true, true, rttThreshold)
	decision = routing.Decision{}
	predictedNNStats = routing.Stats{}
	directStats = routing.Stats{Jitter: 50}
	decision = routeDecisionFunc(decision, predictedNNStats, routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionHighJitterMultipath}, decision)

	// Test when multipath reason is high packet loss
	routeDecisionFunc = routing.DecideMultipath(true, true, true, rttThreshold)
	decision = routing.Decision{}
	predictedNNStats = routing.Stats{}
	directStats = routing.Stats{PacketLoss: 1}
	decision = routeDecisionFunc(decision, predictedNNStats, routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionHighPacketLossMultipath}, decision)
}
