package routing

type Decision struct {
	OnNetworkNext bool
	Reason        DecisionReason
}

// Decision takes in whether or not the logic is currently considering staying on network next,
// the stats of the predicted network next route,
// the stats of the last network next route,
// and the stats of the direct route and decides whether or not to take the predicted network next route.
// A reason is also provided for billing.
type DecisionFunc func(prevDecision Decision, predictedNextStats Stats, lastNextStats Stats, directStats Stats) Decision

// DecisionReason is the reason why a Decision was made.
type DecisionReason uint64

// Route decision flags are required for billing, so this has to work the same for the billing entry to be correct
const (
	DecisionNoChange              DecisionReason = 0
	DecisionForceDirect           DecisionReason = 1 << 1
	DecisionForceNext             DecisionReason = 1 << 2
	DecisionNoNextRoute           DecisionReason = 1 << 3
	DecisionABTestDirect          DecisionReason = 1 << 4
	DecisionRTTReduction          DecisionReason = 1 << 5
	DecisionPacketLossMultipath   DecisionReason = 1 << 6
	DecisionJitterMultipath       DecisionReason = 1 << 7
	DecisionVetoRTT               DecisionReason = 1 << 8
	DecisionRTTMultipath          DecisionReason = 1 << 9
	DecisionVetoPacketLoss        DecisionReason = 1 << 10
	DecisionFallbackToDirect      DecisionReason = 1 << 11
	DecisionUnused                DecisionReason = 1 << 12
	DecisionVetoYOLO              DecisionReason = 1 << 13
	DecisionVetoNoRoute           DecisionReason = 1 << 14
	DecisionDatacenterHasNoRelays DecisionReason = 1 << 15
	DecisionInitialSlice          DecisionReason = 1 << 16
	DecisionNoNearRelays          DecisionReason = 1 << 17
)

// DecideUpgradeRTT will decide if the client should use the network next route if the RTT reduction is greater than the given threshold.
// This decision only upgrades direct routes, so network next routes aren't considered.
func DecideUpgradeRTT(rttThreshold float64) DecisionFunc {
	return func(prevDecision Decision, predictedNextStats Stats, lastNextStats Stats, directStats Stats) Decision {
		// If upgrading to a nextwork next route would reduce RTT by at least the given threshold, upgrade
		if !prevDecision.OnNetworkNext && directStats.RTT-predictedNextStats.RTT >= rttThreshold {
			return Decision{true, DecisionRTTReduction}
		}

		// If the RTT isn't reduced, return the original route consideration
		return Decision{prevDecision.OnNetworkNext, DecisionNoChange}
	}
}

// DecideDowngradeRTT will decide if the client should continue using the network next route if the network next RTT increase doesn't exceed the hysteresis value.
// This decision only downgrades network next routes, so direct routes aren't considered.
func DecideDowngradeRTT(rttHysteresis float64) DecisionFunc {
	return func(prevDecision Decision, predictedNextStats Stats, lastNextStats Stats, directStats Stats) Decision {
		// If staying on a nextwork next route doesn't increase RTT by more than the given hysteresis value, stay
		if prevDecision.OnNetworkNext {
			if predictedNextStats.RTT-directStats.RTT <= rttHysteresis {
				return Decision{true, DecisionRTTReduction}
			}

			// network next route increases RTT too much, switch back to direct
			return Decision{false, DecisionVetoRTT} // Wrong reason, but there isn't a reason for this situation
		}

		// If the route is already direct, don't touch it
		return Decision{false, DecisionNoChange}
	}
}

// DecideVeto will decide if a client should switch to a direct route if the network next route it's on increases
// RTT by more than the RTT veto value, or increases packet loss if packet loss safety is enabled.
// This decision only downgrades network next routes, so direct routes aren't considered.
func DecideVeto(rttVeto float64, packetLossSafety bool, yolo bool) DecisionFunc {
	return func(prevDecision Decision, predictedNextStats Stats, lastNextStats Stats, directStats Stats) Decision {
		if prevDecision.OnNetworkNext {
			// Whether or not the network next route made the RTT worse than the veto value
			if lastNextStats.RTT-directStats.RTT > rttVeto {
				// If the buyer has YouOnlyLiveOnce safety setting enabled, add that reason to the DecisionReason
				if yolo {
					return Decision{false, DecisionVetoRTT | DecisionVetoYOLO}
				}

				return Decision{false, DecisionVetoRTT}
			}

			// Whether or not the network next route made the packet loss worse, if the buyer has packet loss safety enabled
			if packetLossSafety && lastNextStats.PacketLoss > directStats.PacketLoss {
				// If the buyer has YouOnlyLiveOnce safety setting enabled, add that reason to the DecisionReason
				if yolo {
					return Decision{false, DecisionVetoPacketLoss | DecisionVetoYOLO}
				}

				return Decision{false, DecisionVetoPacketLoss}
			}

			// If the route isn't vetoed, then it stays on network next
			return Decision{true, DecisionNoChange}
		}

		// If the route isn't on network next yet, then this decision doesn't apply.
		return Decision{false, DecisionNoChange}
	}
}

// DecideCommitted is not yet implemented
func DecideCommitted() DecisionFunc {
	return func(prevDecision Decision, predictedNextStats Stats, lastNextStats Stats, directStats Stats) Decision {
		return Decision{prevDecision.OnNetworkNext, DecisionNoChange}
	}
}
