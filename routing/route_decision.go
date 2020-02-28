package routing

// RouteDecision takes in whether or not the logic is currently considering staying on network next,
// the stats of the predicted network next route,
// the stats of the last network next route,
// and the stats of the direct route and decides whether or not to take the predicted network next route.
// A reason is also provided for billing.
type RouteDecision func(onNextworkNext bool, predictedNextStats *Stats, lastNextStats *Stats, directStats *Stats) (bool, RouteDecisionReason)

// RouteDecisionReason is the reason why a RouteDecision was made.
type RouteDecisionReason uint64

// Route decision flags are required for billing, so this has to work the same for the billing entry to be correct
const (
	DecisionNoChange              RouteDecisionReason = 0
	DecisionForceDirect           RouteDecisionReason = 1 << 1
	DecisionForceNext             RouteDecisionReason = 1 << 2
	DecisionNoNextRoute           RouteDecisionReason = 1 << 3
	DecisionABTestDirect          RouteDecisionReason = 1 << 4
	DecisionRTTReduction          RouteDecisionReason = 1 << 5
	DecisionPacketLossMultipath   RouteDecisionReason = 1 << 6
	DecisionJitterMultipath       RouteDecisionReason = 1 << 7
	DecisionVetoRTT               RouteDecisionReason = 1 << 8
	DecisionRTTMultipath          RouteDecisionReason = 1 << 9
	DecisionVetoPacketLoss        RouteDecisionReason = 1 << 10
	DecisionFallbackToDirect      RouteDecisionReason = 1 << 11
	DecisionUnused                RouteDecisionReason = 1 << 12
	DecisionVetoYOLO              RouteDecisionReason = 1 << 13
	DecisionVetoNoRoute           RouteDecisionReason = 1 << 14
	DecisionDatacenterHasNoRelays RouteDecisionReason = 1 << 15
	DecisionInitialSlice          RouteDecisionReason = 1 << 16
	DecisionNoNearRelays          RouteDecisionReason = 1 << 17
)

// DecideUpgradeRTT will decide if the client should use the network next route if the RTT reduction is greater than the given threshold.
// This decision only upgrades direct routes, so network next routes aren't considered.
func DecideUpgradeRTT(rttThreshold float64) RouteDecision {
	return func(onNextworkNext bool, predictedNextStats *Stats, lastNextStats *Stats, directStats *Stats) (bool, RouteDecisionReason) {
		// If upgrading to a nextwork next route would reduce RTT by at least the given threshold, upgrade
		if !onNextworkNext && directStats.RTT-predictedNextStats.RTT >= rttThreshold {
			return true, DecisionRTTReduction
		}

		// If the RTT isn't reduced, return the original route consideration
		return onNextworkNext, DecisionNoChange
	}
}

// DecideDowngradeRTT will decide if the client should continue using the network next route if the network next RTT increase doesn't exceed the hysteresis value.
// This decision only downgrades network next routes, so direct routes aren't considered.
func DecideDowngradeRTT(rttHysteresis float64) RouteDecision {
	return func(onNextworkNext bool, predictedNextStats *Stats, lastNextStats *Stats, directStats *Stats) (bool, RouteDecisionReason) {
		// If staying on a nextwork next route doesn't increase RTT by more than the given hysteresis value, stay
		if onNextworkNext {
			if predictedNextStats.RTT-directStats.RTT <= rttHysteresis {
				return true, DecisionRTTReduction
			}

			// network next route increases RTT too much, switch back to direct
			return false, DecisionVetoRTT // Wrong reason, but there isn't a reason for this situation
		}

		// If the route is already direct, don't touch it
		return false, DecisionNoChange
	}
}

// DecideVeto will decide if a client should switch to a direct route if the network next route it's on increases
// RTT by more than the RTT veto value, or increases packet loss if packet loss safety is enabled.
// This decision only downgrades network next routes, so direct routes aren't considered.
func DecideVeto(rttVeto float64, packetLossSafety bool, yolo bool) RouteDecision {
	return func(onNetworkNext bool, predictedNextStats *Stats, lastNextStats *Stats, directStats *Stats) (bool, RouteDecisionReason) {
		if onNetworkNext {
			// Whether or not the network next route made the RTT worse than the veto value
			if lastNextStats.RTT-directStats.RTT > rttVeto {
				// If the buyer has YouOnlyLiveOnce safety setting enabled, add that reason to the RouteDecisionReason
				if yolo {
					return false, DecisionVetoRTT | DecisionVetoYOLO
				}

				return false, DecisionVetoRTT
			}

			// Whether or not the network next route made the packet loss worse, if the buyer has packet loss safety enabled
			if packetLossSafety && lastNextStats.PacketLoss > directStats.PacketLoss {
				// If the buyer has YouOnlyLiveOnce safety setting enabled, add that reason to the RouteDecisionReason
				if yolo {
					return false, DecisionVetoPacketLoss | DecisionVetoYOLO
				}

				return false, DecisionVetoPacketLoss
			}

			// If the route isn't vetoed, then it stays on network next
			return true, DecisionNoChange
		}

		// If the route isn't on network next yet, then this decision doesn't apply.
		return false, DecisionNoChange
	}
}

// DecideCommitted is not yet implemented
func DecideCommitted() RouteDecision {
	return func(onNetworkNext bool, predictedNextStats *Stats, lastNextStats *Stats, directStats *Stats) (bool, RouteDecisionReason) {
		return onNetworkNext, DecisionNoChange
	}
}
