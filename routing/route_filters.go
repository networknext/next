package routing

import (
	"math"
	"math/rand"
)

// RouteFilter reduces a slice of routes according to the filter function.
// Takes in an array of routes and returns the filtered array of routes.
// If the filter couldn't reduce the list of routes, then it returns nil.
type RouteFilter func(routes []Route) []Route

// FilterBestRTT returns the best routes based on lowest RTT, or nil if no best route is found.
// This will return multiple routes if the routes have the same RTT.
func FilterBestRTT() RouteFilter {
	return func(routes []Route) []Route {
		bestRoutes := make([]Route, 0)
		for _, route := range routes {
			if len(bestRoutes) == 0 || route.Stats.RTT < bestRoutes[0].Stats.RTT {
				bestRoutes = make([]Route, 1)
				bestRoutes[0] = route
			} else if route.Stats.RTT == bestRoutes[0].Stats.RTT {
				bestRoutes = append(bestRoutes, route)
			}
		}

		// Returns nil if there are no acceptable routes
		if len(bestRoutes) == 0 {
			return nil
		}

		return bestRoutes
	}
}

// FilterAcceptableRoutesFromRTT will return a list of acceptable routes, which is defined as all routes whose RTT is within the given threshold of the base route's RTT.
// This filter uses the firrt route in the list as the base route.
// This filter actually grows the list of routes.
// Returns nil if there are no acceptable routes.
func FilterAcceptableRoutesFromRTT(rttSwitchThreshold float32) RouteFilter {
	return func(routes []Route) []Route {
		var baseRoute *Route
		if len(routes) > 0 {
			baseRoute = &routes[0]
		} else {
			return nil // FilterAcceptableRoutesFromRTT needs a base route to filter correctly
		}

		acceptableRoutes := make([]Route, 0)
		for _, route := range routes {
			rttDifference := baseRoute.Stats.RTT - route.Stats.RTT
			if math.Abs(rttDifference) < float64(rttSwitchThreshold) {
				acceptableRoutes = append(acceptableRoutes, route)
			}
		}

		// Return nil if there are no acceptable routes
		if len(acceptableRoutes) == 0 {
			return nil
		}

		return acceptableRoutes
	}
}

// FilterContainsRouteHash returns the route if its route hash matches a route in the list of routes, or nil if it is not.
func FilterContainsRouteHash(routeHash uint64) RouteFilter {
	return func(routes []Route) []Route {
		for _, route := range routes {
			sameRoute := routeHash == route.Hash64()
			if sameRoute {
				return []Route{route}
			}
		}

		return nil
	}
}

// FilterRoutesByRandomDestRelay will group the routes by their destination relays, then choose a random relay to return routes from.
func FilterRoutesByRandomDestRelay() RouteFilter {
	return func(routes []Route) []Route {
		// Group routes by destination relay
		destRelayRouteMap := make(map[uint64][]Route)
		for _, route := range routes {
			// In case the route has zero relays, ignore it
			if len(route.Relays) == 0 {
				continue
			}

			// Get the destination relay
			destRelay := route.Relays[len(route.Relays)-1]

			// If the relay isn't in the map yet, add an empty slice to add routes to
			if _, ok := destRelayRouteMap[destRelay.ID]; !ok {
				destRelayRouteMap[destRelay.ID] = nil
			}

			// Append the route to the relay's entry in the map
			destRelayRouteMap[destRelay.ID] = append(destRelayRouteMap[destRelay.ID], route)
		}

		// Get a slice of all destination relay IDs
		var destinationRelayIDs []uint64
		for destRelayID := range destRelayRouteMap {
			destinationRelayIDs = append(destinationRelayIDs, destRelayID)
		}

		// NOTE - Why does this need to be sorted if a random destination relay is chosen anyway?
		// sort.Slice(destinationRelayIDs, func(i, j int) bool {
		// 	return destinationRelayIDs[i] < destinationRelayIDs[j]
		// })

		// choose a random destination relay, and use the routes from that
		relayRoutes := destRelayRouteMap[destinationRelayIDs[rand.Intn(len(destinationRelayIDs))]]
		return relayRoutes
	}
}

// FilterRandomRoute returns a random route from the list of routes.
func FilterRandomRoute() RouteFilter {
	return func(routes []Route) []Route {
		return []Route{routes[rand.Intn(len(routes))]}
	}
}
