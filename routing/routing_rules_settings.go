package routing

const (
	ModeDefault     = 0 // Default behaviour - try to pick the best route
	ModeForceDirect = 1 // Force direct route (even if there is potential improvement)
	ModeForceNext   = 2 // Force a network next route (even if there is a degradement)
)

// Various settings a buyer can tweak to adjust the behaviour of Network Next route selection to their liking
type RoutingRulesSettings struct {
	// The maximum upstream bandwidth a customer is willing to pay for per slice
	EnvelopeKbpsUp int64

	// The maximum downstream bandwidth a customer is willing to pay for per slice
	EnvelopeKbpsDown int64

	// The router mode (see "mode" constants defined above)
	Mode int64

	// The maximum bid price in USD cents (¢) a customer is willing to pay per GB of traffic sent over network next
	// For example a value of 100 here would mean the customer is willing to pay $1.00 USD per GB of network next accelerated traffic
	MaxCentsPerGB int64

	// The maximum acceptable latency for the game. If we can't reduce the latency to be at least this then don't take network next
	// Note: not currently being used in old backend
	AcceptableLatency float32

	// How close to the best route in terms of latency routes need to be to be considered acceptable to take.
	// For example if RTTRouteSwitch was set to 20ms the best route in the matrix had an RTT of 60ms, routes with an RTT of more than 80ms would be filtered out
	RTTRouteSwitch float32

	// How many milliseconds the latency has to be improved by before going from a direct route to a network next route
	// For example if RTTThreshold was set to 20ms and the direct route had an RTT of 80ms, we would only take network next routes that have 60ms or lower latency
	RTTThreshold float32

	// How many milliseconds the latency has to be degraded by before going from a network next route to a direct route
	// For example if RTTHysteresis was set to 10ms, the direct route had an RTT of 80ms and we were on a network next route with a RTT of 85ms, we would not go back to direct
	// Not used when multipath enabled!
	RTTHysteresis float32

	// How much worse the latency of a network next route being taken needs to be than direct for the session to be "vetoed"
	// To be "vetoed" means that a particular session ID has been temporarily forced to take direct (times out after an hour)
	// Not used when multipath enabled!
	RTTVeto float32

	// If true, after being downgraded from a network next route to a direct route, the client will not be put back on a network next route for that session
	// Not used when multipath enabled!
	EnableYouOnlyLiveOnce bool

	// If true, causes sessions to be "vetoed" if network next packet loss is greater than direct packet loss
	// Not used when multipath enabled!
	EnablePacketLossSafety bool

	/* MULTIPATH */
	// Multipath means network traffic is sent over multiple network routes (any combination of direct and multiple network next routes)
	// Once a session has multipath enabled, it will stay on multipath until the session ends. As of such vetos are disabled

	// If true, enables multipath when there is 1% or more packet loss on the direct route
	EnableMultipathForPacketLoss bool

	// If true, enables multipath when there is 50ms or more jitter on the direct route
	EnableMultipathForJitter bool

	// If true, enables multipath when there is a next route that beats direct by the value specified in RTTThreshold
	EnableMultipathForRTT bool

	// If true, the customer is participating in an A/B test. Additional metrics will be recorded and half the sessions that would take network next will take direct instead
	EnableABTest bool
}

var DefaultRoutingRulesSettings = RoutingRulesSettings{
	MaxCentsPerGB:     25.0,
	EnvelopeKbpsUp:    256,
	EnvelopeKbpsDown:  256,
	AcceptableLatency: -1.0,
	RTTThreshold:      5.0,
	RTTRouteSwitch:    2.0,
}
