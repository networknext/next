package routing

import "fmt"

type RouteDecision func(onNextworkNext bool, nextRoute *Route, directRoute *Route) (RouteDecisionResult, string)

type RouteDecisionResult uint64

// Route decision flags are required for billing, so this has to work the same for the billing entry to be correct
const RouteDecisionForceDirect = (uint64(1) << 1)
const RouteDecisionForceNext = (uint64(1) << 2)
const RouteDecisionNoNextRoute = (uint64(1) << 3)
const RouteDecisionABTestDirect = (uint64(1) << 4)
const RouteDecisionRTTReduction = (uint64(1) << 5)
const RouteDecisionPacketLossMultipath = (uint64(1) << 6)
const RouteDecisionJitterMultipath = (uint64(1) << 7)
const RouteDecisionVetoRTT = (uint64(1) << 8)
const RouteDecisionRTTMultipath = (uint64(1) << 9)
const RouteDecisionVetoPacketLoss = (uint64(1) << 10)
const RouteDecisionFallbackToDirect = (uint64(1) << 11)
const ROUTE_DECISION____REUSE_ME_PLEASE3__ = (uint64(1) << 12)
const RouteDecisionVetoYOLO = (uint64(1) << 13)
const RouteDecisionVetoNoRoute = (uint64(1) << 14)
const RouteDecisionDatacenterHasNoRelays = (uint64(1) << 15)
const RouteDecisionInitialSlice = (uint64(1) << 16)
const RouteDecisionNoNearRelays = (uint64(1) << 17)

func RouteDecisionRTTThreshold(rttThreshold float64) RouteDecision {
	return func(onNextworkNext bool, nextRoute, directRoute *Route) (RouteDecisionResult, string) {
		if !onNextworkNext && directRoute.Stats.RTT-nextRoute.Stats.RTT > rttThreshold {
			return RouteDecisionResult_GoOnNN, fmt.Sprintf("Network Next route improves RTT by more than %vms, session upgraded.", rttThreshold)
		}
	}
}

func RouteDecisionRTTHysteresis(rttHysteresis float64) RouteDecision {
	return func(onNextworkNext bool, nextRoute, directRoute *Route) (RouteDecisionResult, string) {
		if onNextworkNext && nextRoute.Stats.RTT-directRoute.Stats.RTT > rttHysteresis {
			return RouteDecisionResult_GoToDirect, fmt.Sprintf("Network Next route degrades RTT by more than %vms, session falls back to direct.", rttHysteresis)
		}
	}
}

func RouteDecisionVeto(rttVeto int) RouteDecision {
	return func(onNextworkNext bool, nextRoute, directRoute *Route) (RouteDecisionResult, string) {
		if onNextworkNext && packet.NextMeanRtt-packet.DirectMeanRtt > rttVeto {

		}
	}
}

func RouteDecisionCommitted() RouteDecision {
	return func(onNextworkNext bool, nextRoute, directRoute *Route) (RouteDecisionResult, string) {

	}
}
