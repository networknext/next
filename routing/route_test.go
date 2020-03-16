package routing_test

import (
	"testing"

	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestDecide(t *testing.T) {
	// Test case where we should upgrade to a nextwork next route
	{
		decisionFuncs := []routing.DecisionFunc{
			routing.DecideUpgradeRTT(float64(routing.DefaultRoutingRulesSettings.RTTThreshold)),
			routing.DecideDowngradeRTT(float64(routing.DefaultRoutingRulesSettings.RTTHysteresis)),
			routing.DecideVeto(float64(routing.DefaultRoutingRulesSettings.RTTVeto), routing.DefaultRoutingRulesSettings.EnablePacketLossSafety, routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce),
			routing.DecideCommitted(),
		}

		lastNNStats := routing.Stats{
			RTT:        0,
			Jitter:     0,
			PacketLoss: 0,
		}

		lastDirectStats := routing.Stats{
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

		startingDecision := routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionNoChange,
		}

		expected := routing.Decision{
			OnNetworkNext: true,
			Reason:        routing.DecisionRTTReduction,
		}

		// Loop through all permutations of the decision functions and test that the result is the same
		perms := permutations(decisionFuncs)
		for i := 0; i < len(perms); i++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, &metrics.EmptyDecisionMetrics, perms[i]...)
			assert.Equal(t, expected, decision)
		}
	}

	// Test case where we should stay on a nextwork next route
	{
		decisionFuncs := []routing.DecisionFunc{
			routing.DecideUpgradeRTT(float64(routing.DefaultRoutingRulesSettings.RTTThreshold)),
			routing.DecideDowngradeRTT(float64(routing.DefaultRoutingRulesSettings.RTTHysteresis)),
			routing.DecideVeto(float64(routing.DefaultRoutingRulesSettings.RTTVeto), routing.DefaultRoutingRulesSettings.EnablePacketLossSafety, routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce),
			routing.DecideCommitted(),
		}

		lastNNStats := routing.Stats{
			RTT:        30,
			Jitter:     0,
			PacketLoss: 0,
		}

		lastDirectStats := routing.Stats{
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
			Reason:        routing.DecisionNoChange,
		}

		expected := routing.Decision{
			OnNetworkNext: true,
			Reason:        routing.DecisionNoChange,
		}

		// Loop through all permutations of the decision functions and test that the result is the same
		perms := permutations(decisionFuncs)
		for i := 0; i < len(perms); i++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, &metrics.EmptyDecisionMetrics, perms[i]...)
			assert.Equal(t, expected, decision)
		}
	}

	// Test case where we should get off a nextwork next route due to hysteresis
	{
		decisionFuncs := []routing.DecisionFunc{
			routing.DecideUpgradeRTT(float64(routing.DefaultRoutingRulesSettings.RTTThreshold)),
			routing.DecideDowngradeRTT(float64(routing.DefaultRoutingRulesSettings.RTTHysteresis)),
			routing.DecideVeto(float64(routing.DefaultRoutingRulesSettings.RTTVeto), routing.DefaultRoutingRulesSettings.EnablePacketLossSafety, routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce),
			routing.DecideCommitted(),
		}

		lastNNStats := routing.Stats{
			RTT:        30,
			Jitter:     0,
			PacketLoss: 0,
		}

		lastDirectStats := routing.Stats{
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
			Reason:        routing.DecisionNoChange,
		}

		expected := routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionRTTIncrease,
		}

		// Loop through all permutations of the decision functions and test that the result is the same
		perms := permutations(decisionFuncs)
		for i := 0; i < len(perms); i++ {
			decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, &metrics.EmptyDecisionMetrics, perms[i]...)
			assert.Equal(t, expected, decision)
		}
	}

	// Test case where we should get off a network next route due to RTT veto
	{
		routingRulesSettings := routing.DefaultRoutingRulesSettings

		decisionFuncs := []routing.DecisionFunc{
			routing.DecideUpgradeRTT(float64(routingRulesSettings.RTTThreshold)),
			routing.DecideDowngradeRTT(float64(routingRulesSettings.RTTHysteresis)),
			routing.DecideCommitted(),
		}

		lastNNStats := routing.Stats{
			RTT:        61,
			Jitter:     0,
			PacketLoss: 0,
		}

		lastDirectStats := routing.Stats{
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
			Reason:        routing.DecisionNoChange,
		}

		expected := routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoRTT,
		}

		// Loop through all permutations and combinations of the decision functions and test that the result is the same
		// Veto is taken out of the permutations and inserted at the end because it has the highest priority
		// and should be at the end to work properly. All functions before it can be in an order though.
		combs := combinations(decisionFuncs)
		for i := 0; i < len(combs); i++ {
			perms := permutations(combs[i])

			for j := 0; j < len(perms); j++ {
				perms[j] = append(perms[j], routing.DecideVeto(float64(routingRulesSettings.RTTVeto), routingRulesSettings.EnablePacketLossSafety, routingRulesSettings.EnableYouOnlyLiveOnce))
				decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, &metrics.EmptyDecisionMetrics, perms[j]...)
				assert.Equal(t, expected, decision)
			}
		}
	}

	// Test case where we should downgrade to a direct route due to RTT veto and YOLO is enabled
	{
		routingRulesSettings := routing.DefaultRoutingRulesSettings
		routingRulesSettings.EnableYouOnlyLiveOnce = true

		decisionFuncs := []routing.DecisionFunc{
			routing.DecideUpgradeRTT(float64(routingRulesSettings.RTTThreshold)),
			routing.DecideDowngradeRTT(float64(routingRulesSettings.RTTHysteresis)),
			routing.DecideCommitted(),
		}

		lastNNStats := routing.Stats{
			RTT:        61,
			Jitter:     0,
			PacketLoss: 0,
		}

		lastDirectStats := routing.Stats{
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
			Reason:        routing.DecisionNoChange,
		}

		expected := routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoRTT | routing.DecisionVetoYOLO,
		}

		// Loop through all permutations and combinations of the decision functions and test that the result is the same
		// Veto is taken out of the permutations and inserted at the end because it has the highest priority
		// and should be at the end to work properly. All functions before it can be in an order though.
		combs := combinations(decisionFuncs)
		for i := 0; i < len(combs); i++ {
			perms := permutations(combs[i])

			for j := 0; j < len(perms); j++ {
				perms[j] = append(perms[j], routing.DecideVeto(float64(routingRulesSettings.RTTVeto), routingRulesSettings.EnablePacketLossSafety, routingRulesSettings.EnableYouOnlyLiveOnce))
				decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, &metrics.EmptyDecisionMetrics, perms[j]...)
				assert.Equal(t, expected, decision)
			}
		}
	}

	// Test case where we should downgrade to a direct route due to packet loss veto
	{
		routingRulesSettings := routing.DefaultRoutingRulesSettings
		routingRulesSettings.EnablePacketLossSafety = true

		decisionFuncs := []routing.DecisionFunc{
			routing.DecideUpgradeRTT(float64(routingRulesSettings.RTTThreshold)),
			routing.DecideDowngradeRTT(float64(routingRulesSettings.RTTHysteresis)),
			routing.DecideVeto(float64(routingRulesSettings.RTTVeto), routingRulesSettings.EnablePacketLossSafety, routingRulesSettings.EnableYouOnlyLiveOnce),
			routing.DecideCommitted(),
		}

		lastNNStats := routing.Stats{
			RTT:        40,
			Jitter:     0,
			PacketLoss: 10,
		}

		lastDirectStats := routing.Stats{
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
			Reason:        routing.DecisionNoChange,
		}

		expected := routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoPacketLoss,
		}

		// Loop through all permutations and combinations of the decision functions and test that the result is the same
		// Veto is taken out of the permutations and inserted at the end because it has the highest priority
		// and should be at the end to work properly. All functions before it can be in an order though.
		combs := combinations(decisionFuncs)
		for i := 0; i < len(combs); i++ {
			perms := permutations(combs[i])

			for j := 0; j < len(perms); j++ {
				perms[j] = append(perms[j], routing.DecideVeto(float64(routingRulesSettings.RTTVeto), routingRulesSettings.EnablePacketLossSafety, routingRulesSettings.EnableYouOnlyLiveOnce))
				decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, &metrics.EmptyDecisionMetrics, perms[j]...)
				assert.Equal(t, expected, decision)
			}
		}
	}

	// Test case where we should downgrade to a direct route due to packet loss veto and YOLO is enabled
	{
		routingRulesSettings := routing.DefaultRoutingRulesSettings
		routingRulesSettings.EnablePacketLossSafety = true
		routingRulesSettings.EnableYouOnlyLiveOnce = true

		decisionFuncs := []routing.DecisionFunc{
			routing.DecideUpgradeRTT(float64(routingRulesSettings.RTTThreshold)),
			routing.DecideDowngradeRTT(float64(routingRulesSettings.RTTHysteresis)),
			routing.DecideVeto(float64(routingRulesSettings.RTTVeto), routingRulesSettings.EnablePacketLossSafety, routingRulesSettings.EnableYouOnlyLiveOnce),
			routing.DecideCommitted(),
		}

		lastNNStats := routing.Stats{
			RTT:        40,
			Jitter:     0,
			PacketLoss: 10,
		}

		lastDirectStats := routing.Stats{
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
			Reason:        routing.DecisionNoChange,
		}

		expected := routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoPacketLoss | routing.DecisionVetoYOLO,
		}

		// Loop through all permutations and combinations of the decision functions and test that the result is the same
		// Veto is taken out of the permutations and inserted at the end because it has the highest priority
		// and should be at the end to work properly. All functions before it can be in an order though.
		combs := combinations(decisionFuncs)
		for i := 0; i < len(combs); i++ {
			perms := permutations(combs[i])

			for j := 0; j < len(perms); j++ {
				perms[j] = append(perms[j], routing.DecideVeto(float64(routingRulesSettings.RTTVeto), routingRulesSettings.EnablePacketLossSafety, routingRulesSettings.EnableYouOnlyLiveOnce))
				decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, &metrics.EmptyDecisionMetrics, perms[j]...)
				assert.Equal(t, expected, decision)
			}
		}
	}

	// Test case where we stay on direct with no change
	{
		decisionFuncs := []routing.DecisionFunc{
			routing.DecideUpgradeRTT(float64(routing.DefaultRoutingRulesSettings.RTTThreshold)),
			routing.DecideDowngradeRTT(float64(routing.DefaultRoutingRulesSettings.RTTHysteresis)),
			routing.DecideVeto(float64(routing.DefaultRoutingRulesSettings.RTTVeto), routing.DefaultRoutingRulesSettings.EnablePacketLossSafety, routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce),
			routing.DecideCommitted(),
		}

		lastNNStats := routing.Stats{
			RTT:        0,
			Jitter:     0,
			PacketLoss: 0,
		}

		lastDirectStats := routing.Stats{
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

		startingDecision := routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionNoChange,
		}

		expected := routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionNoChange,
		}

		// Loop through all permutations and combinations of the decision functions and test that the result is the same
		combs := combinations(decisionFuncs)
		for i := 0; i < len(combs); i++ {
			perms := permutations(combs[i])

			for j := 0; j < len(perms); j++ {
				decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, &metrics.EmptyDecisionMetrics, perms[j]...)
				assert.Equal(t, expected, decision)
			}
		}
	}

	// Test case where we stay on nextwork next with no change
	{
		decisionFuncs := []routing.DecisionFunc{
			routing.DecideUpgradeRTT(float64(routing.DefaultRoutingRulesSettings.RTTThreshold)),
			routing.DecideDowngradeRTT(float64(routing.DefaultRoutingRulesSettings.RTTHysteresis)),
			routing.DecideVeto(float64(routing.DefaultRoutingRulesSettings.RTTVeto), routing.DefaultRoutingRulesSettings.EnablePacketLossSafety, routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce),
			routing.DecideCommitted(),
		}

		lastNNStats := routing.Stats{
			RTT:        30,
			Jitter:     0,
			PacketLoss: 0,
		}

		lastDirectStats := routing.Stats{
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
			OnNetworkNext: true,
			Reason:        routing.DecisionNoChange,
		}

		expected := routing.Decision{
			OnNetworkNext: true,
			Reason:        routing.DecisionNoChange,
		}

		// Loop through all permutations and combinations of the decision functions and test that the result is the same
		combs := combinations(decisionFuncs)
		for i := 0; i < len(combs); i++ {
			perms := permutations(combs[i])

			for j := 0; j < len(perms); j++ {
				decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, &metrics.EmptyDecisionMetrics, perms[j]...)
				assert.Equal(t, expected, decision)
			}
		}
	}

	// Test case to check that DecisionNoChange really doesn't change the decision reason
	{
		decisionFuncs := []routing.DecisionFunc{
			routing.DecideUpgradeRTT(float64(routing.DefaultRoutingRulesSettings.RTTThreshold)),
			routing.DecideDowngradeRTT(float64(routing.DefaultRoutingRulesSettings.RTTHysteresis)),
			routing.DecideVeto(float64(routing.DefaultRoutingRulesSettings.RTTVeto), routing.DefaultRoutingRulesSettings.EnablePacketLossSafety, routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce),
			routing.DecideCommitted(),
		}

		lastNNStats := routing.Stats{
			RTT:        30,
			Jitter:     0,
			PacketLoss: 0,
		}

		lastDirectStats := routing.Stats{
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

		// Use a decision reason not handled by the decision functions so that
		// we can confirm DecisionNoChange was never set as the reason and that
		// the reason isn't changed
		startingDecision := routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionInitialSlice,
		}

		expected := routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionInitialSlice,
		}

		// Loop through all permutations and combinations of the decision functions and test that the result is the same
		combs := combinations(decisionFuncs)
		for i := 0; i < len(combs); i++ {
			perms := permutations(combs[i])

			for j := 0; j < len(perms); j++ {
				decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, &metrics.EmptyDecisionMetrics, perms[j]...)
				assert.Equal(t, expected, decision)
			}
		}

		// Run test again with OnNetworkNext true
		startingDecision = routing.Decision{
			OnNetworkNext: true,
			Reason:        routing.DecisionInitialSlice,
		}

		expected = routing.Decision{
			OnNetworkNext: true,
			Reason:        routing.DecisionInitialSlice,
		}

		// Loop through all permutations and combinations of the decision functions and test that the result is the same
		combs = combinations(decisionFuncs)
		for i := 0; i < len(combs); i++ {
			perms := permutations(combs[i])

			for j := 0; j < len(perms); j++ {
				decision := route.Decide(startingDecision, lastNNStats, lastDirectStats, &metrics.EmptyDecisionMetrics, perms[j]...)
				assert.Equal(t, expected, decision)
			}
		}
	}
}

// Algorithm adapted from https://stackoverflow.com/questions/45177692/getting-all-possible-combinations-of-an-array-of-objects
func combinations(decisionFuncs []routing.DecisionFunc) [][]routing.DecisionFunc {
	combs := make([][]routing.DecisionFunc, 1<<len(decisionFuncs))

	for i := 0; i < 1<<len(decisionFuncs); i++ {
		bits := 1
		for j := 0; j < len(decisionFuncs); j++ {
			if bits&i != 0 {
				combs[i] = append(combs[i], decisionFuncs[j])
			}

			bits <<= 1
		}
	}

	return combs
}

// Heaps algorithm adapted from https://en.wikipedia.org/wiki/Heap%27s_algorithm
func permutations(decisionFuncs []routing.DecisionFunc) [][]routing.DecisionFunc {
	length := len(decisionFuncs)
	c := make([]int, length)

	perms := make([][]routing.DecisionFunc, 0)

	decisionFuncsToAppend := make([]routing.DecisionFunc, len(decisionFuncs))
	copy(decisionFuncsToAppend, decisionFuncs)
	perms = append(perms, decisionFuncsToAppend)

	i := 0
	for i < length {
		if c[i] < i {
			if i%2 == 0 {
				swap(0, i, decisionFuncs)
			} else {
				swap(c[i], i, decisionFuncs)
			}

			decisionFuncsToAppend = make([]routing.DecisionFunc, len(decisionFuncs))
			copy(decisionFuncsToAppend, decisionFuncs)
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

func swap(i int, j int, decisionFuncs []routing.DecisionFunc) {
	temp := decisionFuncs[i]
	decisionFuncs[i] = decisionFuncs[j]
	decisionFuncs[j] = temp
}
