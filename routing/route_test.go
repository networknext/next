package routing_test

import (
	"testing"

	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

// Test case where we should upgrade to a nextwork next route
func TestDecideUpgradeRoute(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        0,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        50,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        30,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{}

	expected := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionRTTReduction,
	}

	// Loop through all permutations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	perms := permutations(decisionFuncIndices)
	funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)
	for i := 0; i < len(funcs); i++ {
		decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[i]...)
		assert.Equal(t, expected, decision)
	}
}

// Test case where we should get off a nextwork next route due to hysteresis
func TestDecideDowngradeRTTHysteresis(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        36,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	expected := routing.Decision{
		OnNetworkNext: false,
		Reason:        routing.DecisionRTTHysteresis,
	}

	// Loop through all permutations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	perms := permutations(decisionFuncIndices)
	funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)
	for i := 0; i < len(funcs); i++ {
		decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[i]...)
		assert.Equal(t, expected, decision)
	}
}

// Test case where we should get off a nextwork next route due to hysteresis and YOLO is enabled, vetoing the session
func TestDecideDowngradeRTTHysteresisYOLO(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings
	rrs.EnableYouOnlyLiveOnce = true

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        36,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	expected := routing.Decision{
		OnNetworkNext: false,
		Reason:        routing.DecisionRTTHysteresis | routing.DecisionVetoYOLO,
	}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		perms = filterPermutations(perms, 1) // Remove all permutations that don't include DecideDowngradeRTT, since that's the function we're testing for
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(funcs); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case where we should get off a network next route due to RTT veto
func TestDecideRTTVeto(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        61,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        40,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        50,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	expected := routing.Decision{
		OnNetworkNext: false,
		Reason:        routing.DecisionVetoRTT,
	}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		perms = filterPermutations(perms, 2) // Remove all permutations that don't include DecideVeto, since that's the function we're testing for
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(funcs); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case where we should downgrade to a direct route due to RTT veto and YOLO is enabled
func TestDecideRTTVetoYOLO(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings
	rrs.EnableYouOnlyLiveOnce = true

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        61,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        40,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        50,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	expected := routing.Decision{
		OnNetworkNext: false,
		Reason:        routing.DecisionVetoRTT | routing.DecisionVetoYOLO,
	}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		perms = filterPermutations(perms, 2) // Remove all permutations that don't include DecideVeto, since that's the function we're testing for
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(funcs); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case where we shouldn't downgrade to a direct route due to packet loss veto yet since it's within the first 3 slices
func TestDecidePacketLossVetoEarlySlice(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings
	rrs.EnablePacketLossSafety = true

	onNNSliceCounter := uint64(2)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        40,
		Jitter:     0,
		PacketLoss: 10,
	}

	lastDirectStats := &routing.Stats{
		RTT:        50,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        45,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	expected := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		perms = filterPermutations(perms, 2) // Remove all permutations that don't include DecideVeto, since that's the function we're testing for
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(funcs); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case where we should downgrade to a direct route due to packet loss veto
func TestDecidePacketLossVeto(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings
	rrs.EnablePacketLossSafety = true

	onNNSliceCounter := uint64(3)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        40,
		Jitter:     0,
		PacketLoss: 10,
	}

	lastDirectStats := &routing.Stats{
		RTT:        50,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        45,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	expected := routing.Decision{
		OnNetworkNext: false,
		Reason:        routing.DecisionVetoPacketLoss,
	}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		perms = filterPermutations(perms, 2) // Remove all permutations that don't include DecideVeto, since that's the function we're testing for
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(funcs); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case where we should downgrade to a direct route due to packet loss veto and YOLO is enabled
func TestDecidePacketLossVetoYOLO(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings
	rrs.EnablePacketLossSafety = true
	rrs.EnableYouOnlyLiveOnce = true

	onNNSliceCounter := uint64(3)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        40,
		Jitter:     0,
		PacketLoss: 10,
	}

	lastDirectStats := &routing.Stats{
		RTT:        50,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        45,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	expected := routing.Decision{
		OnNetworkNext: false,
		Reason:        routing.DecisionVetoPacketLoss | routing.DecisionVetoYOLO,
	}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		perms = filterPermutations(perms, 2) // Remove all permutations that don't include DecideVeto, since that's the function we're testing for
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(funcs); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case where we stay on direct with no change
func TestDecideStayOnDirectRoute(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        0,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        35,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{}

	expected := routing.Decision{}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(funcs); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case where we stay on nextwork next with no change
func TestDecideStayOnNNRoute(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        40,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        30,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	expected := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(funcs); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case to check that DecisionInitialSlice is never the reason twice in a row
func TestValidateInitialSlice(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        30,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{
		OnNetworkNext: false,
		Reason:        routing.DecisionInitialSlice,
	}

	expected := routing.Decision{}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(perms); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case where we should get off a network next route due to commit veto
func TestDecideCommitVeto(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings
	rrs.EnableTryBeforeYouBuy = true

	onNNSliceCounter := uint64(0)

	committedData := &routing.CommittedData{}

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
		routing.DecideCommitted(true, uint8(rrs.TryBeforeYouBuyMaxSlices), rrs.EnableYouOnlyLiveOnce, committedData),
	}

	lastNNStats := &routing.Stats{
		RTT:        45,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        40,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        20,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	expected := routing.Decision{
		OnNetworkNext: false,
		Reason:        routing.DecisionVetoCommit,
	}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		perms = filterPermutations(perms, 4) // Remove all permutations that don't include DecideCommitted, since that's the function we're testing for
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(funcs); j++ {
			committedData.Pending = true
			committedData.ObservedSliceCounter = uint8(rrs.TryBeforeYouBuyMaxSlices)
			committedData.Committed = false

			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
			assert.Equal(t, false, committedData.Pending)
			assert.Equal(t, uint8(0), committedData.ObservedSliceCounter)
			assert.Equal(t, false, committedData.Committed)
		}
	}
}

// Test case to check that the committed flag from the decision function is being set correctly
func TestValidateCommitted(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings
	rrs.EnableTryBeforeYouBuy = true

	onNNSliceCounter := uint64(0)

	committedData := &routing.CommittedData{}

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
		routing.DecideCommitted(true, uint8(rrs.TryBeforeYouBuyMaxSlices), rrs.EnableYouOnlyLiveOnce, committedData),
	}

	lastNNStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        40,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        35,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	expected := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionNoReason,
	}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		perms = filterPermutations(perms, 4) // Remove all permutations that don't include DecideCommitted, since that's the function we're testing for
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(funcs); j++ {
			committedData.Pending = true
			committedData.ObservedSliceCounter = uint8(rrs.TryBeforeYouBuyMaxSlices)
			committedData.Committed = false

			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
			assert.Equal(t, false, committedData.Pending)
			assert.Equal(t, uint8(0), committedData.ObservedSliceCounter)
			assert.Equal(t, true, committedData.Committed)
		}
	}
}

// Test case to check that direct routes are still served if the initial next route isn't good enough for multipath
func TestDecideMultipathDirect(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings
	rrs.EnableMultipathForRTT = true
	rrs.EnableMultipathForJitter = true
	rrs.EnableMultipathForPacketLoss = true

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        0,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        40,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{}

	expected := routing.Decision{}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(perms); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case to check that direct routes are never served if multipath is already active
func TestDecideMultipathStayActive(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings
	rrs.EnableMultipathForRTT = true
	rrs.EnableMultipathForJitter = true
	rrs.EnableMultipathForPacketLoss = true

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        50,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        30,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        40,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{true, routing.DecisionRTTReductionMultipath}

	expected := routing.Decision{true, routing.DecisionRTTReductionMultipath}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(perms); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case to check that when multipath is enabled for RTT that the session is upgraded and the decision reason is always the multipath version
func TestValidateMultipathRTT(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings
	rrs.EnableMultipathForRTT = true

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        0,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        40,
		Jitter:     0,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        30,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{}

	expected := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionRTTReductionMultipath,
	}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		perms = filterPermutations(perms, 3) // Remove all permutations that don't include DecideMultipath, since that's the function we're testing for
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(perms); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case to check that when multipath is enabled for jitter that the session is upgraded and the decision reason is always the multipath version
func TestValidateMultipathJitter(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings
	rrs.EnableMultipathForJitter = true

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        0,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        0,
		Jitter:     50,
		PacketLoss: 0,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        0,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{}

	expected := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionHighJitterMultipath,
	}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		perms = filterPermutations(perms, 3) // Remove all permutations that don't include DecideMultipath, since that's the function we're testing for
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(perms); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Test case to check that when multipath is enabled for packet loss that the session is upgraded and the decision reason is always the multipath version
func TestValidateMultipathPacketLoss(t *testing.T) {
	rrs := routing.DefaultRoutingRulesSettings
	rrs.EnableMultipathForPacketLoss = true

	onNNSliceCounter := uint64(0)

	decisionFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(rrs.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(rrs.RTTHysteresis), rrs.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(rrs.RTTVeto), rrs.EnablePacketLossSafety, rrs.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(rrs.EnableMultipathForRTT, rrs.EnableMultipathForJitter, rrs.EnableMultipathForPacketLoss, float64(rrs.RTTThreshold), float64(rrs.MultipathPacketLossThreshold)),
	}

	lastNNStats := &routing.Stats{
		RTT:        0,
		Jitter:     0,
		PacketLoss: 0,
	}

	lastDirectStats := &routing.Stats{
		RTT:        0,
		Jitter:     0,
		PacketLoss: 2,
	}

	route := routing.Route{
		Stats: routing.Stats{
			RTT:        0,
			Jitter:     0,
			PacketLoss: 0,
		},
	}

	startingDecision := routing.Decision{}

	expected := routing.Decision{
		OnNetworkNext: true,
		Reason:        routing.DecisionHighPacketLossMultipath,
	}

	// Loop through all permutations and combinations of the decision functions and test that the result is the same
	decisionFuncIndices := createIndexSlice(decisionFuncs)
	combs := combinations(decisionFuncIndices)
	for i := 0; i < len(combs); i++ {
		perms := permutations(combs[i])
		perms = filterPermutations(perms, 3) // Remove all permutations that don't include DecideMultipath, since that's the function we're testing for
		funcs := replaceIndicesWithDecisionFuncs(perms, decisionFuncs)

		for j := 0; j < len(perms); j++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, funcs[j]...)
			assert.Equal(t, expected, decision)
		}
	}
}

// Algorithm adapted from https://stackoverflow.com/questions/45177692/getting-all-possible-combinations-of-an-array-of-objects
func combinations(decisionFuncIndices []int) [][]int {
	combs := make([][]int, 1<<len(decisionFuncIndices))

	for i := 0; i < 1<<len(decisionFuncIndices); i++ {
		bits := 1
		for j := 0; j < len(decisionFuncIndices); j++ {
			if bits&i != 0 {
				combs[i] = append(combs[i], decisionFuncIndices[j])
			}

			bits <<= 1
		}
	}

	return combs
}

// Heaps algorithm adapted from https://en.wikipedia.org/wiki/Heap%27s_algorithm
func permutations(decisionFuncIndices []int) [][]int {
	length := len(decisionFuncIndices)
	c := make([]int, length)

	perms := make([][]int, 0)

	decisionFuncsToAppend := make([]int, len(decisionFuncIndices))
	copy(decisionFuncsToAppend, decisionFuncIndices)
	perms = append(perms, decisionFuncsToAppend)

	i := 0
	for i < length {
		if c[i] < i {
			if i%2 == 0 {
				swap(0, i, decisionFuncIndices)
			} else {
				swap(c[i], i, decisionFuncIndices)
			}

			decisionFuncsToAppend = make([]int, len(decisionFuncIndices))
			copy(decisionFuncsToAppend, decisionFuncIndices)
			perms = append(perms, decisionFuncsToAppend)

			c[i]++
			i = 0
		} else {
			c[i] = 0
			i++
		}
	}

	return perms
}

func swap(i int, j int, decisionFuncIndices []int) {
	temp := decisionFuncIndices[i]
	decisionFuncIndices[i] = decisionFuncIndices[j]
	decisionFuncIndices[j] = temp
}

// Creates a slice of decision func indices
func createIndexSlice(decisionFuncs []routing.DecisionFunc) []int {
	indices := make([]int, len(decisionFuncs))
	for i := 0; i < len(indices); i++ {
		indices[i] = i
	}

	return indices
}

// Removes all permutations that do not contain the specified index
func filterPermutations(perms [][]int, index int) [][]int {
	permCount := len(perms)
	filtered := make([][]int, 0)

	for i := 0; i < permCount; i++ {
		addToFilter := false
		for j := 0; j < len(perms[i]); j++ {
			if perms[i][j] == index {
				addToFilter = true
				break
			}
		}

		if addToFilter {
			filtered = append(filtered, perms[i])
		}
	}

	return filtered
}

// Replaces index permutations with the corresponding decision funcs
func replaceIndicesWithDecisionFuncs(indices [][]int, key []routing.DecisionFunc) [][]routing.DecisionFunc {
	decisionFuncs := make([][]routing.DecisionFunc, 0)

	for i := 0; i < len(indices); i++ {
		decisionFuncs = append(decisionFuncs, make([]routing.DecisionFunc, len(indices[i])))
		for j := 0; j < len(indices[i]); j++ {
			index := indices[i][j]
			decisionFuncs[i][j] = key[index]
		}
	}

	return decisionFuncs
}
