package routing

import (
	"fmt"
)

// Decision is a representation of a whether or not a route should go over network next and why or why not.
type Decision struct {
	OnNetworkNext bool
	Reason        DecisionReason
}

// DecisionFunc is a decision making function that decides whether or not a route should go over network next or direct.
// Decision takes in whether or not the logic is currently considering staying on network next,
// the stats of the predicted network next route,
// the stats of the last network next route,
// and the stats of the direct route and decides whether or not to take the predicted network next route.
// A reason is also provided for billing.
type DecisionFunc func(prevDecision Decision, predictedNextStats Stats, lastNextStats Stats, directStats Stats) Decision

// DecisionReason is the reason why a Decision was made.
type DecisionReason uint64

func (dr DecisionReason) String() string {
	var reason string

	switch dr {
	case DecisionNoChange:
		reason = "No Change"
	case DecisionForceDirect:
		reason = "Force Direct"
	case DecisionForceNext:
		reason = "Force Next"
	case DecisionNoNextRoute:
		reason = "No Next Route"
	case DecisionABTestDirect:
		reason = "AB Test Direct"
	case DecisionRTTReduction:
		reason = "RTT Reduction"
	case DecisionPacketLossMultipath:
		reason = "Packet Loss Multipath"
	case DecisionJitterMultipath:
		reason = "Jitter Multipath"
	case DecisionVetoRTT:
		reason = "Veto RTT"
	case DecisionRTTMultipath:
		reason = "RTT Multipath"
	case DecisionVetoPacketLoss:
		reason = "Veto Packet Loss"
	case DecisionFallbackToDirect:
		reason = "Fallback to Direct"
	case DecisionVetoYOLO:
		reason = "Veto YOLO"
	case DecisionDatacenterHasNoRelays:
		reason = "Datacenter Has No Relays"
	case DecisionInitialSlice:
		reason = "Initial Slice"
	case DecisionNoNearRelays:
		reason = "No Near Relays"
	case DecisionVetoRTT | DecisionVetoYOLO:
		reason = "Veto RTT YOLO"
	case DecisionVetoPacketLoss | DecisionVetoYOLO:
		reason = "Veto Packet Loss YOLO"
	case DecisionRTTHysteresis:
		reason = "RTT Hysteresis"
	default:
		reason = "Unused"
	}

	return fmt.Sprintf("%s (%d)", reason, dr)
}

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
	DecisionVetoYOLO              DecisionReason = 1 << 13
	DecisionDatacenterHasNoRelays DecisionReason = 1 << 15
	DecisionInitialSlice          DecisionReason = 1 << 16
	DecisionNoNearRelays          DecisionReason = 1 << 17
	DecisionRTTHysteresis         DecisionReason = 1 << 18
	DecisionVetoCommit            DecisionReason = 1 << 19
)

// DecideUpgradeRTT will decide if the client should use the network next route if the RTT reduction is greater than the given threshold.
// This decision only upgrades direct routes, so network next routes aren't considered.
func DecideUpgradeRTT(rttThreshold float64) DecisionFunc {
	return func(prevDecision Decision, predictedNextStats Stats, lastNextStats Stats, directStats Stats) Decision {
		// If upgrading to a nextwork next route would reduce RTT by at least the given threshold, upgrade
		if !prevDecision.OnNetworkNext && !IsVetoed(prevDecision) && directStats.RTT-predictedNextStats.RTT >= rttThreshold {
			return Decision{true, DecisionRTTReduction}
		}

		// If the RTT isn't reduced, return the original route consideration
		return Decision{prevDecision.OnNetworkNext, DecisionNoChange}
	}
}

// DecideDowngradeRTT will decide if the client should continue using the network next route if the network next RTT increase doesn't exceed the hysteresis value.
// This decision only downgrades network next routes, so direct routes aren't considered.
func DecideDowngradeRTT(rttHysteresis float64, yolo bool) DecisionFunc {
	return func(prevDecision Decision, predictedNextStats Stats, lastNextStats Stats, directStats Stats) Decision {
		// If staying on a nextwork next route doesn't increase RTT by more than the given hysteresis value, stay
		if prevDecision.OnNetworkNext {
			if predictedNextStats.RTT-directStats.RTT <= rttHysteresis {
				return Decision{prevDecision.OnNetworkNext, DecisionNoChange}
			}

			// If the buyer has YouOnlyLiveOnce safety setting enabled, veto them and add that reason to the DecisionReason
			if yolo {
				return Decision{false, DecisionVetoRTT | DecisionVetoYOLO}
			}

			// Network next route increases RTT too much, switch back to direct
			return Decision{false, DecisionRTTHysteresis}
		}

		// If the route is already direct, don't touch it
		return Decision{prevDecision.OnNetworkNext, DecisionNoChange}
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
			return Decision{prevDecision.OnNetworkNext, DecisionNoChange}
		}

		// Handle the case where another decision function decided to switch back to direct due
		// to RTT increase, but the increase is so severe it should be vetoed.
		if prevDecision.Reason == DecisionRTTHysteresis {
			if lastNextStats.RTT-directStats.RTT > rttVeto {
				// If the buyer has YouOnlyLiveOnce safety setting enabled, add that reason to the DecisionReason
				if yolo {
					return Decision{false, DecisionVetoRTT | DecisionVetoYOLO}
				}

				return Decision{false, DecisionVetoRTT}
			}
		}

		// If the route isn't on network next yet, then this decision doesn't apply.
		return Decision{prevDecision.OnNetworkNext, DecisionNoChange}
	}
}

// DecideCommitted will decide if the route should be committed to the decided route through the committed out parameter.
// This function will not ever upgrade a route, it will only either keep it the same or veto it if it ends up being much worse
// than direct or if it takes too long to confidently decide.
// IN VARS
// onNNLastSlice: Whether or not the session was on NN during the last slice
// maxObservedSlices: The maximum number of slices to observe before vetoing an inconclusive session
// OUT VARS
// commitPending: Whether or not the logic is still considering to commit or not
// observedSliceCounter: How many slices have been observed while deciding whether or not to commit
// committed: Whether or not the route is committed
// The out vars describe the state of the committed logic to keep this function stateless.
func DecideCommitted(onNNLastSlice bool, maxObservedSlices uint8, commitPending *bool, observedSliceCounter *uint8, committed *bool) DecisionFunc {
	return func(prevDecision Decision, predictedNextStats, lastNextStats, directStats Stats) Decision {
		// Only consider committing a route if try before you buy is enabled and
		// the route decision logic has decided on a NN route in the first place
		if prevDecision.OnNetworkNext {
			// Check if the session ID is newly on NN
			if !onNNLastSlice {
				// Set the session to pending commit
				*commitPending = true
				*observedSliceCounter = 0
				*committed = false

				// Don't change the route deicison yet
				return Decision{prevDecision.OnNetworkNext, DecisionNoChange}
			} else if *commitPending { // See if the session is still pending
				if lastNextStats.RTT <= directStats.RTT && lastNextStats.PacketLoss <= directStats.PacketLoss {
					// The NN route was the same or better than direct, so commit to it
					*commitPending = false
					*observedSliceCounter = 0
					*committed = true

					// Don't actually change the route decision since it's good
					return Decision{prevDecision.OnNetworkNext, DecisionNoChange}
				} else if lastNextStats.RTT > directStats.RTT || lastNextStats.PacketLoss > directStats.PacketLoss {
					// The route wasn't so bad that it was vetoed, so continue to observe the route
					*commitPending = true
					*observedSliceCounter++
					*committed = false

					if *observedSliceCounter >= maxObservedSlices {
						// This session doesn't seem to be working out, just veto it
						*commitPending = false
						*observedSliceCounter = 0
						*committed = false
						return Decision{false, DecisionVetoCommit}
					}

					// Keep waiting for more data
					return Decision{prevDecision.OnNetworkNext, DecisionNoChange}
				}
			}
		}

		*commitPending = false
		*observedSliceCounter = 0
		*committed = false

		// Don't affect direct routes
		return Decision{prevDecision.OnNetworkNext, DecisionNoChange}
	}
}

// IsVetoed returns true if the given route decision was a veto.
func IsVetoed(decision Decision) bool {
	if !decision.OnNetworkNext {
		if decision.Reason&DecisionVetoPacketLoss != 0 || decision.Reason&DecisionVetoRTT != 0 || decision.Reason&DecisionVetoYOLO != 0 || decision.Reason&DecisionVetoCommit != 0 {
			return true
		}
	}

	return false
}
