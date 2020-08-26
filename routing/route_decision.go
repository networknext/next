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
type DecisionFunc func(prevDecision Decision, predictedNextStats, lastNextStats, lastDirectStats *Stats) Decision

// DecisionReason is the reason why a Decision was made.
type DecisionReason uint64

func (dr DecisionReason) String() string {
	var reason string

	switch dr {
	case DecisionNoReason:
		reason = "No Reason"
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
	case DecisionHighPacketLossMultipath:
		reason = "Packet Loss Multipath"
	case DecisionHighJitterMultipath:
		reason = "Jitter Multipath"
	case DecisionVetoRTT:
		reason = "Veto RTT"
	case DecisionRTTReductionMultipath:
		reason = "RTT Reduction Multipath"
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
	case DecisionRTTHysteresis | DecisionVetoYOLO:
		reason = "RTT Hysteresis YOLO"
	case DecisionVetoCommit:
		reason = "Veto Commit"
	case DecisionVetoCommit | DecisionVetoYOLO:
		reason = "Veto Commit YOLO"
	case DecisionDatacenterDisabled:
		reason = "Datacenter Disabled"
	case DecisionVetoNoRoute:
		reason = "Vetoed No Route"
	case DecisionNoLocation:
		reason = "No Location"
	case DecisionBuyerNotLive:
		reason = "Buyer Not Live"
	case DecisionMultipathVetoRTT:
		reason = "Multipath Veto RTT"
	case DecisionMultipathVetoRTT | DecisionVetoYOLO:
		reason = "Multipath Veto RTT YOLO"
	default:
		reason = "Unknown"
	}

	return fmt.Sprintf("%s (%d)", reason, dr)
}

// Route decision flags are required for billing, so this has to work the same for the billing entry to be correct
const (
	DecisionNoReason                DecisionReason = 0
	DecisionForceDirect             DecisionReason = 1 << 1
	DecisionForceNext               DecisionReason = 1 << 2
	DecisionNoNextRoute             DecisionReason = 1 << 3
	DecisionABTestDirect            DecisionReason = 1 << 4
	DecisionRTTReduction            DecisionReason = 1 << 5
	DecisionHighPacketLossMultipath DecisionReason = 1 << 6
	DecisionHighJitterMultipath     DecisionReason = 1 << 7
	DecisionVetoRTT                 DecisionReason = 1 << 8
	DecisionRTTReductionMultipath   DecisionReason = 1 << 9
	DecisionVetoPacketLoss          DecisionReason = 1 << 10
	DecisionFallbackToDirect        DecisionReason = 1 << 11
	DecisionVetoYOLO                DecisionReason = 1 << 13
	DecisionDatacenterHasNoRelays   DecisionReason = 1 << 15
	DecisionInitialSlice            DecisionReason = 1 << 16
	DecisionNoNearRelays            DecisionReason = 1 << 17
	DecisionRTTHysteresis           DecisionReason = 1 << 18
	DecisionVetoCommit              DecisionReason = 1 << 19
	DecisionDatacenterDisabled      DecisionReason = 1 << 20
	DecisionVetoNoRoute             DecisionReason = 1 << 21
	DecisionNoLocation              DecisionReason = 1 << 22
	DecisionBuyerNotLive            DecisionReason = 1 << 23
	DecisionMultipathVetoRTT        DecisionReason = 1 << 24
)

// DecideUpgradeRTT will decide if the client should use the network next route if the RTT reduction is greater than the given threshold.
// This decision only upgrades direct routes, so network next routes aren't considered.
// Multipath sessions aren't considered.
func DecideUpgradeRTT(rttThreshold float64) DecisionFunc {
	return func(prevDecision Decision, predictedNextStats, lastNextStats, lastDirectStats *Stats) Decision {
		// If we've already decided on multipath, then don't change the reason
		if IsMultipath(prevDecision) {
			return prevDecision
		}

		predictedImprovement := lastDirectStats.RTT - predictedNextStats.RTT

		// If upgrading to a nextwork next route would reduce RTT by at least the given threshold, upgrade
		if !prevDecision.OnNetworkNext && !IsVetoed(prevDecision) && predictedImprovement >= rttThreshold {
			return Decision{true, DecisionRTTReduction}
		}

		// If the RTT isn't reduced, return the original route consideration
		return prevDecision
	}
}

// DecideDowngradeRTT will decide if the client should continue using the network next route if the network next RTT increase doesn't exceed the hysteresis value.
// This decision only downgrades network next routes, so direct routes aren't considered.
// Multipath sessions aren't considered.
func DecideDowngradeRTT(rttHysteresis float64, yolo bool) DecisionFunc {
	return func(prevDecision Decision, predictedNextStats, lastNextStats, lastDirectStats *Stats) Decision {
		// If we've already decided on multipath, then don't change the reason
		if IsMultipath(prevDecision) {
			return prevDecision
		}

		// If we are on network next and we are improving RTT by less than the given hysteresis value, go direct
		if prevDecision.OnNetworkNext {
			predictedImprovement := lastDirectStats.RTT - predictedNextStats.RTT

			if predictedImprovement < rttHysteresis {
				// If the buyer has YouOnlyLiveOnce safety setting enabled, veto them and add that reason to the DecisionReason
				if yolo {
					return Decision{false, DecisionRTTHysteresis | DecisionVetoYOLO}
				}

				// Network next route increases RTT too much, switch back to direct
				return Decision{false, DecisionRTTHysteresis}
			}

			// network next route is still good, so keep it
			return prevDecision
		}

		// If the route is already direct, don't touch it
		return prevDecision
	}
}

// DecideVeto will decide if a client should switch to a direct route if the network next route it's on increases
// RTT by more than the RTT veto value, or increases packet loss if packet loss safety is enabled.
// This decision only downgrades network next routes, so direct routes aren't considered.
// Multipath sessions aren't considered.
func DecideVeto(onNNSliceCounter uint64, rttVeto float64, packetLossSafety bool, yolo bool) DecisionFunc {
	return func(prevDecision Decision, predictedNextStats, lastNextStats, lastDirectStats *Stats) Decision {
		// If we've already decided on multipath, then don't change the reason
		if IsMultipath(prevDecision) {
			return prevDecision
		}

		actualImprovement := lastDirectStats.RTT - lastNextStats.RTT

		if prevDecision.OnNetworkNext {
			// Whether or not the network next route made the RTT worse than the veto value
			if actualImprovement < rttVeto {
				// If the buyer has YouOnlyLiveOnce safety setting enabled, add that reason to the DecisionReason
				if yolo {
					return Decision{false, DecisionVetoRTT | DecisionVetoYOLO}
				}

				return Decision{false, DecisionVetoRTT}
			}

			// Whether or not the network next route made the packet loss worse, if the buyer has packet loss safety enabled
			if onNNSliceCounter > 2 && packetLossSafety && lastNextStats.PacketLoss > lastDirectStats.PacketLoss {
				// If the buyer has YouOnlyLiveOnce safety setting enabled, add that reason to the DecisionReason
				if yolo {
					return Decision{false, DecisionVetoPacketLoss | DecisionVetoYOLO}
				}

				return Decision{false, DecisionVetoPacketLoss}
			}

			// If the route isn't vetoed, then it stays on network next
			return prevDecision
		}

		// Handle the case where another decision function decided to switch back to direct due
		// to RTT hysteresis, but the increase is so severe it should be vetoed.
		if prevDecision.Reason == DecisionRTTHysteresis {
			if actualImprovement < rttVeto {
				// If the buyer has YouOnlyLiveOnce safety setting enabled, add that reason to the DecisionReason
				if yolo {
					return Decision{false, DecisionVetoRTT | DecisionVetoYOLO}
				}

				return Decision{false, DecisionVetoRTT}
			}
		}

		// If the route isn't on network next yet, then this decision doesn't apply.
		return prevDecision
	}
}

type CommittedData struct {
	Pending              bool
	ObservedSliceCounter uint8
	Committed            bool
}

// DecideCommitted will decide if the route should be committed to the decided route through the committed out parameter.
// This function will not ever upgrade a route, it will only either keep it the same or veto it if it ends up being much worse
// than direct or if it takes too long to confidently decide.
// IN VARS
// onNNLastSlice: Whether or not the session was on NN during the last slice
// maxObservedSlices: The maximum number of slices to observe before vetoing an inconclusive session
// yolo: Whether or not the buyer has YOLO enabled
// OUT VARS
// committedData: container struct for the following fields:
// 	commitPending: Whether or not the logic is still considering to commit or not
// 	observedSliceCounter: How many slices have been observed while deciding whether or not to commit
// 	committed: Whether or not the route is committed
// The out vars describe the state of the committed logic to keep this function stateless.
// This decision only downgrades network next routes, so direct routes aren't considered.
// Multipath sessions aren't considered.
func DecideCommitted(onNNLastSlice bool, maxObservedSlices uint8, yolo bool, committedData *CommittedData) DecisionFunc {
	return func(prevDecision Decision, predictedNextStats, lastNextStats, lastDirectStats *Stats) Decision {
		// If we've already decided on multipath, then don't change the reason
		if IsMultipath(prevDecision) {
			committedData.Pending = false
			committedData.ObservedSliceCounter = 0
			committedData.Committed = true // Since network next routes will always be used in multipath, always commit to them
			return prevDecision
		}

		// Only consider committing a route if try before you buy is enabled and
		// the route decision logic has decided on a NN route in the first place
		if prevDecision.OnNetworkNext {
			// Check if the session ID is newly on NN
			if !onNNLastSlice {
				// Set the session to pending commit
				committedData.Pending = true
				committedData.ObservedSliceCounter = 0
				committedData.Committed = false

				// Don't change the route deicison yet
				return prevDecision
			} else if committedData.Pending { // See if the session is still pending
				if lastNextStats.RTT <= lastDirectStats.RTT && lastNextStats.PacketLoss <= lastDirectStats.PacketLoss {
					// The NN route was the same or better than direct, so commit to it
					committedData.Pending = false
					committedData.ObservedSliceCounter = 0
					committedData.Committed = true

					// Don't actually change the route decision since it's good
					return prevDecision
				} else if lastNextStats.RTT > lastDirectStats.RTT || lastNextStats.PacketLoss > lastDirectStats.PacketLoss {
					// The route wasn't so bad that it was vetoed, so continue to observe the route
					committedData.Pending = true
					committedData.ObservedSliceCounter++
					committedData.Committed = false

					if committedData.ObservedSliceCounter >= maxObservedSlices {
						// This session doesn't seem to be working out, just veto it
						committedData.Pending = false
						committedData.ObservedSliceCounter = 0
						committedData.Committed = false

						// Add yolo reason if yolo is enabled
						if yolo {
							return Decision{false, DecisionVetoCommit | DecisionVetoYOLO}
						}

						return Decision{false, DecisionVetoCommit}
					}

					// Keep waiting for more data
					return prevDecision
				}
			}
		}

		committedData.Pending = false
		committedData.ObservedSliceCounter = 0
		committedData.Committed = false

		// Don't affect direct routes
		return prevDecision
	}
}

// DecideMultipath will decide if we should serve a network next route to be used for multipath
// If the decision function can't find a good enough reason to send a network next route, then it decides to go direct
// If multipath isn't enabled then the decision isn't affected
func DecideMultipath(rttMultipath bool, jitterMultipath bool, packetLossMultipath bool, rttThreshold float64, packetLossThreshold float64) DecisionFunc {
	return func(prevDecision Decision, predictedNextStats, lastNextStats, lastDirectStats *Stats) Decision {
		// If we've already decided on multipath, then don't change the reason
		// This is to make sure that the session can't go back to direct, since multipath always needs a next route
		if IsMultipath(prevDecision) {
			return prevDecision
		}

		decision := prevDecision

		// Reset the decision reason if multipath is enabled
		if rttMultipath || jitterMultipath || packetLossMultipath {
			decision.OnNetworkNext = false // Start with false, then if we should use multipath then go to true
			decision.Reason = DecisionNoReason
		}

		// If the RTT reduction would result in direct -> next, then use the RTT multipath decision reason
		if rttMultipath && lastDirectStats.RTT-predictedNextStats.RTT >= rttThreshold {
			decision.OnNetworkNext = true
			decision.Reason |= DecisionRTTReductionMultipath
		}

		// If the direct jitter is too high, then use multipath for jitter
		if jitterMultipath && lastDirectStats.Jitter >= 50.0 {
			decision.OnNetworkNext = true
			decision.Reason |= DecisionHighJitterMultipath
		}

		// If the direct packet loss is more than 1%, then use multipath for packet loss
		if packetLossMultipath && lastDirectStats.PacketLoss >= packetLossThreshold {
			decision.OnNetworkNext = true
			decision.Reason |= DecisionHighPacketLossMultipath
		}

		// There was probably a ping spike due to an overloaded connection for 2x multipath bandwidth,
		// so "multipath veto" this user
		if lastDirectStats.RTT > 500 || lastNextStats.RTT > 500 {
			decision.OnNetworkNext = false
			decision.Reason = DecisionMultipathVetoRTT
		}

		return decision
	}
}

// IsVetoed returns true if the given route decision was a veto.
func IsVetoed(decision Decision) bool {
	if !decision.OnNetworkNext {
		if decision.Reason&DecisionVetoPacketLoss != 0 || decision.Reason&DecisionVetoRTT != 0 || decision.Reason&DecisionVetoYOLO != 0 || decision.Reason&DecisionVetoCommit != 0 || decision.Reason&DecisionVetoNoRoute != 0 {
			return true
		}
	}

	return false
}

// IsMultipath returns true if the given route decision decided to enable multipath
func IsMultipath(decision Decision) bool {
	if decision.Reason&DecisionRTTReductionMultipath != 0 || decision.Reason&DecisionHighJitterMultipath != 0 || decision.Reason&DecisionHighPacketLossMultipath != 0 {
		return true
	}

	return false
}
