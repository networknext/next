package routing_test

import (
	"testing"

	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestDecideUpgradeRTT(t *testing.T) {
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

	rttThreshold := float64(routing.DefaultRoutingRulesSettings.RTTThreshold)
	routeDecisionFunc := routing.DecideUpgradeRTT(rttThreshold)

	// Test if multipath is enabled
	assert.Equal(
		t,
		routing.Decision{true, routing.DecisionRTTReductionMultipath},
		routeDecisionFunc(routing.Decision{true, routing.DecisionRTTReductionMultipath}, predictedStats, &routing.Stats{}, directStats),
	)

	// Now test if a route gets upgraded to network next
	assert.Equal(
		t,
		routing.Decision{true, routing.DecisionRTTReduction},
		routeDecisionFunc(routing.Decision{}, predictedStats, &routing.Stats{}, directStats),
	)

	// Now test if the route is left alone

	predictedStats.RTT = directStats.RTT
	assert.Equal(
		t,
		routing.Decision{},
		routeDecisionFunc(routing.Decision{}, predictedStats, &routing.Stats{}, directStats),
	)
}

func TestDecideDowngradeRTT(t *testing.T) {
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

	rttHyteresis := float64(routing.DefaultRoutingRulesSettings.RTTHysteresis)
	routeDecisionFunc := routing.DecideDowngradeRTT(rttHyteresis, false)

	// Test if multipath is enabled
	decision := routing.Decision{true, routing.DecisionRTTReductionMultipath}
	decision = routeDecisionFunc(decision, predictedStats, &routing.Stats{}, directStats)

	assert.Equal(t, routing.Decision{true, routing.DecisionRTTReductionMultipath}, decision)

	// Now test if a route stays on the network next route
	decision = routing.Decision{true, routing.DecisionNoReason}
	decision = routeDecisionFunc(decision, predictedStats, &routing.Stats{}, directStats)

	assert.Equal(t, routing.Decision{true, routing.DecisionNoReason}, decision)

	// Now test to see if the route gets downgraded to a direct route due to RTT
	predictedStats.RTT = directStats.RTT - rttHyteresis + 1.0

	decision = routeDecisionFunc(decision, predictedStats, &routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionRTTHysteresis}, decision)

	// Now test if a direct route is given
	decision = routing.Decision{}
	decision = routeDecisionFunc(decision, predictedStats, &routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{}, decision)

	// Now test if the route is vetoed with YOLO enabled
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideDowngradeRTT(rttHyteresis, true)
	decision = routeDecisionFunc(decision, predictedStats, &routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionRTTHysteresis | routing.DecisionVetoYOLO}, decision)
}

func TestDecideVeto(t *testing.T) {
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

	onNNSliceCounter := uint64(0)

	rttVeto := float64(routing.DefaultRoutingRulesSettings.RTTVeto)
	routeDecisionFunc := routing.DecideVeto(onNNSliceCounter, rttVeto, false, false)

	// Test if multipath is enabled
	decision := routing.Decision{true, routing.DecisionRTTReductionMultipath}
	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionRTTReductionMultipath}, decision)

	// Now test if a route is vetoed for RTT increases
	decision = routing.Decision{true, routing.DecisionNoReason}
	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT}, decision)

	// Now test for yolo reason
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideVeto(onNNSliceCounter, rttVeto, false, true)

	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT | routing.DecisionVetoYOLO}, decision)

	// Now test that the route won't be vetoed yet for packet loss increases since it's withing the first 3 slices
	lastNextStats.RTT = directStats.RTT
	lastNextStats.PacketLoss = directStats.PacketLoss + 1
	decision = routing.Decision{OnNetworkNext: true, Reason: routing.DecisionRTTReduction}
	routeDecisionFunc = routing.DecideVeto(onNNSliceCounter, rttVeto, true, false)

	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionRTTReduction}, decision)

	// Now test if the route is vetoed for packet loss increases
	onNNSliceCounter = 3
	routeDecisionFunc = routing.DecideVeto(onNNSliceCounter, rttVeto, true, false)

	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoPacketLoss}, decision)

	// Now test for yolo reason
	decision.OnNetworkNext = true
	routeDecisionFunc = routing.DecideVeto(onNNSliceCounter, rttVeto, true, true)

	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoPacketLoss | routing.DecisionVetoYOLO}, decision)

	// Test if route isn't vetoed
	lastNextStats.PacketLoss = directStats.PacketLoss
	decision = routing.Decision{true, routing.DecisionNoReason}
	routeDecisionFunc = routing.DecideVeto(onNNSliceCounter, rttVeto, true, true)

	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionNoReason}, decision)

	// Test if route was changed to direct from another function, but the RTT increase was so severe that it should be vetoed
	lastNextStats.RTT = 60
	decision = routing.Decision{OnNetworkNext: false, Reason: routing.DecisionRTTHysteresis}
	routeDecisionFunc = routing.DecideVeto(onNNSliceCounter, rttVeto, true, false)

	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT}, decision)

	// Now with yolo
	decision = routing.Decision{OnNetworkNext: false, Reason: routing.DecisionRTTHysteresis}
	routeDecisionFunc = routing.DecideVeto(onNNSliceCounter, rttVeto, true, true)

	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoRTT | routing.DecisionVetoYOLO}, decision)

	// Test if direct route isn't changed
	decision = routing.Decision{false, routing.DecisionNoReason}

	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{}, decision)
}

func TestDecideCommitted(t *testing.T) {
	lastNextStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	directStats := &routing.Stats{
		RTT:        60,
		Jitter:     0,
		PacketLoss: 0,
	}

	maxSlices := uint8(routing.DefaultRoutingRulesSettings.TryBeforeYouBuyMaxSlices)

	committedData := &routing.CommittedData{}

	routeDecisionFunc := routing.DecideCommitted(false, maxSlices, false, committedData)

	// Test if multipath is enabled
	decision := routing.Decision{true, routing.DecisionRTTReductionMultipath}
	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionRTTReductionMultipath}, decision)
	assert.Equal(t, false, committedData.Pending)
	assert.Equal(t, uint8(0), committedData.ObservedSliceCounter)
	assert.Equal(t, true, committedData.Committed)

	// Check direct routes aren't affected
	decision = routing.Decision{}
	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{}, decision)
	assert.Equal(t, false, committedData.Pending)
	assert.Equal(t, uint8(0), committedData.ObservedSliceCounter)
	assert.Equal(t, false, committedData.Committed)

	// Check if a slice newly on NN is "initialized" properly
	decision.OnNetworkNext = true
	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionNoReason}, decision)
	assert.Equal(t, true, committedData.Pending)
	assert.Equal(t, uint8(0), committedData.ObservedSliceCounter)
	assert.Equal(t, false, committedData.Committed)

	// Now check the case where the route is better
	routeDecisionFunc = routing.DecideCommitted(true, maxSlices, false, committedData)
	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionNoReason}, decision)
	assert.Equal(t, false, committedData.Pending)
	assert.Equal(t, uint8(0), committedData.ObservedSliceCounter)
	assert.Equal(t, true, committedData.Committed)

	// Now check the case where the route isn't bad enough yet to veto
	committedData.Pending = true
	lastNextStats.RTT = 65
	decision.OnNetworkNext = true
	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionNoReason}, decision)
	assert.Equal(t, true, committedData.Pending)
	assert.Equal(t, uint8(1), committedData.ObservedSliceCounter)
	assert.Equal(t, false, committedData.Committed)

	// Now check the case where the route has taken too long to decide and should veto
	committedData.ObservedSliceCounter = maxSlices
	decision.OnNetworkNext = true
	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoCommit}, decision)
	assert.Equal(t, false, committedData.Pending)
	assert.Equal(t, uint8(0), committedData.ObservedSliceCounter)
	assert.Equal(t, false, committedData.Committed)

	// Now check the case where the route has taken too long to decide and should veto and yolo is enabled
	decision.OnNetworkNext = true
	committedData.Pending = true
	committedData.ObservedSliceCounter = maxSlices
	routeDecisionFunc = routing.DecideCommitted(true, maxSlices, true, committedData)
	decision = routeDecisionFunc(decision, &routing.Stats{}, lastNextStats, directStats)
	assert.Equal(t, routing.Decision{false, routing.DecisionVetoCommit | routing.DecisionVetoYOLO}, decision)
	assert.Equal(t, false, committedData.Pending)
	assert.Equal(t, uint8(0), committedData.ObservedSliceCounter)
	assert.Equal(t, false, committedData.Committed)
}

func TestDecideMultipath(t *testing.T) {
	rttThreshold := float64(routing.LocalRoutingRulesSettings.RTTThreshold)
	packetLossThreshold := float64(routing.LocalRoutingRulesSettings.MultipathPacketLossThreshold)

	// Test if multipath isn't enabled
	routeDecisionFunc := routing.DecideMultipath(false, false, false, rttThreshold, packetLossThreshold)
	predictedNNStats := &routing.Stats{RTT: 30}
	directStats := &routing.Stats{RTT: 60}
	decision := routing.Decision{}
	decision = routeDecisionFunc(decision, predictedNNStats, &routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{}, decision)

	// Test if multipath is already active
	routeDecisionFunc = routing.DecideMultipath(true, true, true, rttThreshold, packetLossThreshold)
	decision = routing.Decision{true, routing.DecisionRTTReductionMultipath}
	decision = routeDecisionFunc(decision, predictedNNStats, &routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionRTTReductionMultipath}, decision)

	// Test when multipath reason is RTT reduction
	decision = routing.Decision{}
	decision = routeDecisionFunc(decision, predictedNNStats, &routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionRTTReductionMultipath}, decision)

	// Test when multipath reason is high jitter
	// directStats.RTT set to -6.0 as it must be less than LocalROutingRulesSettings.RTTTHreshold
	// if predictedNNStats.RTT = 0 (default) for happy path to run and force NN.
	routeDecisionFunc = routing.DecideMultipath(true, true, true, rttThreshold, packetLossThreshold)
	decision = routing.Decision{}
	predictedNNStats = &routing.Stats{}
	directStats = &routing.Stats{Jitter: 50, RTT: -6.0}
	decision = routeDecisionFunc(decision, predictedNNStats, &routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionHighJitterMultipath}, decision)

	// Test when multipath reason is high packet loss
	// directStats.RTT set to -5 as it must be less than LocalROutingRulesSettings.RTTTHreshold
	// if predictedNNStats.RTT = 0 (default) for happy path to run and force NN.
	routeDecisionFunc = routing.DecideMultipath(true, true, true, rttThreshold, packetLossThreshold)
	decision = routing.Decision{}
	predictedNNStats = &routing.Stats{}
	directStats = &routing.Stats{PacketLoss: 1, RTT: -6.0}
	decision = routeDecisionFunc(decision, predictedNNStats, &routing.Stats{}, directStats)
	assert.Equal(t, routing.Decision{true, routing.DecisionHighPacketLossMultipath}, decision)
}
