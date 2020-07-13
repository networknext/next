package routing

import (
	"math/rand"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// SelectorFunc reduces a slice of routes according to the selector function.
// Takes in a slice of routes sorted by lowest RTT as input and returns a new slice of selected routes.
// A RouteSelector never modifies the input.
// If the selector couldn't produce a non-empty slice of routes, then it returns nil.
type SelectorFunc func(routes []Route) []Route

// SelectLogger logs all of the routes currently being selected
func SelectLogger(logger log.Logger) SelectorFunc {
	return func(routes []Route) []Route {
		for _, route := range routes {
			level.Debug(logger).Log("route", route.String())
		}

		return routes
	}
}

// SelectBestRTT returns the best routes based on lowest RTT, or nil if no best route is found.
// This will return multiple routes if the routes have the same RTT.
func SelectBestRTT() SelectorFunc {
	return func(routes []Route) []Route {
		if len(routes) == 0 {
			return nil
		}

		bestRouteCount := 1

		for i := range routes {

			if i+1 < len(routes) && routes[i+1].Stats.RTT == routes[i].Stats.RTT {
				bestRouteCount++
			} else {
				break
			}
		}

		return routes[:bestRouteCount]
	}
}

// SelectAcceptableRoutesFromBestRTT will return a slice of acceptable routes, which is defined as all routes whose RTT is within the given threshold of the best RTT.
// Returns nil if there are no acceptable routes.
func SelectAcceptableRoutesFromBestRTT(rttEpsilon float64) SelectorFunc {
	return func(routes []Route) []Route {
		if len(routes) == 0 {
			return nil
		}

		routeCount := 0
		bestRTT := routes[0].Stats.RTT

		for i := range routes {
			rttDifference := routes[i].Stats.RTT - bestRTT
			if rttDifference <= rttEpsilon {
				routeCount++
			} else {
				break
			}
		}

		// Return nil if there are no acceptable routes
		if routeCount == 0 {
			return nil
		}

		return routes[:routeCount]
	}
}

// SelectContainsRouteHash returns the route if its route hash matches a route in the current list of routes.
// If the route has doesn't match any of the routes, then it will return the existing list of routes to not break route selection.
func SelectContainsRouteHash(routeHash uint64) SelectorFunc {
	return func(routes []Route) []Route {
		for i := range routes {
			if routeHash == routes[i].Hash64() {
				return routes[i : i+1]
			}
		}

		return routes
	}
}

// SelectRoutesByRandomDestRelay will group the current routes by their destination relays, then choose a random relay to return routes from.
func SelectRoutesByRandomDestRelay(source rand.Source) SelectorFunc {
	randgen := rand.New(source)

	return func(routes []Route) []Route {
		// Group routes by destination relay
		destRelayRouteMap := make(map[uint64][]Route)
		for _, route := range routes {
			// In case the route has zero relays, ignore it
			if len(route.Relays) == 0 {
				continue
			}

			// Get the destination relay
			destRelay := &route.Relays[len(route.Relays)-1]

			// If the relay isn't in the map yet, add an empty slice to add routes to
			if _, ok := destRelayRouteMap[destRelay.ID]; !ok {
				destRelayRouteMap[destRelay.ID] = nil
			}

			// Append the route to the relay's entry in the map
			destRelayRouteMap[destRelay.ID] = append(destRelayRouteMap[destRelay.ID], route)
		}

		// Don't continue if there are no routes in the map
		if len(destRelayRouteMap) == 0 {
			return nil
		}

		// Get a slice of all destination relay IDs
		destinationRelayIDs := make([]uint64, len(destRelayRouteMap))
		var i int
		for destRelayID := range destRelayRouteMap {
			destinationRelayIDs[i] = destRelayID
			i++
		}

		temp := randgen.Intn(len(destinationRelayIDs))

		// choose a random destination relay, and use the routes from that
		relayRoutes := destRelayRouteMap[destinationRelayIDs[temp]]
		return relayRoutes
	}
}

// SelectRandomRoute returns a random route from the current list of routes.
func SelectRandomRoute(source rand.Source) SelectorFunc {
	randgen := rand.New(source)

	return func(routes []Route) []Route {
		// Don't select a random route if there are no routes to select
		if len(routes) == 0 {
			return nil
		}

		randomIndex := randgen.Intn(len(routes))
		return routes[randomIndex : randomIndex+1]
	}
}

// SelectUnencumberedRoutes returns routes whose relays don't have too many sessions using them
// For example if sessionThreshold was 0.8 (80%), then all routes that have at least one relay whose session count
// is 80% or more of its max allowed session count would be filtered out
func SelectUnencumberedRoutes(sessionThreshold float64) SelectorFunc {
	return func(routes []Route) []Route {
		unencumberedRoutes := make([]Route, len(routes))
		unencumberedRouteCount := 0

		for i, route := range routes {
			isUnencumbered := true
			for _, relay := range route.Relays {
				if relay.MaxSessions == 0 || float64(relay.TrafficStats.SessionCount)/float64(relay.MaxSessions) >= sessionThreshold {
					isUnencumbered = false
					break
				}
			}

			if isUnencumbered {
				unencumberedRoutes[unencumberedRouteCount] = routes[i]
				unencumberedRouteCount++
			}
		}

		if unencumberedRouteCount == 0 {
			return nil
		}

		return unencumberedRoutes[:unencumberedRouteCount]
	}
}
