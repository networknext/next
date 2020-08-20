package routing

import (
	"math/rand"
	"sort"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// SelectorFunc reduces a slice of routes according to the selector function.
// Takes in a slice of routes as input and returns a new slice of selected routes.
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

// SelectAcceptableRoutesFromBestRTT will return a slice of acceptable routes, which is defined as all routes whose RTT is within the given threshold of the best RTT.
// Returns nil if there are no acceptable routes.
func SelectAcceptableRoutesFromBestRTT(rttEpsilon float64) SelectorFunc {
	// Use SelectBestRTT() to get the best RTT
	bestRTTSelector := SelectBestRTT()
	return func(routes []Route) []Route {
		bestRoutes := bestRTTSelector(routes)
		if bestRoutes == nil {
			return nil // This selector needs the best RTT to work correctly
		}

		bestRTT := bestRoutes[0].Stats.RTT
		acceptableRoutes := make([]Route, 0)
		for _, route := range routes {
			rttDifference := route.Stats.RTT - bestRTT
			if rttDifference <= rttEpsilon {
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

// SelectContainsRouteHash returns the route if its route hash matches a route in the current list of routes.
// If the route has doesn't match any of the routes, then it will return the existing list of routes to not break route selection.
func SelectContainsRouteHash(routeHash uint64) SelectorFunc {
	return func(routes []Route) []Route {
		for _, route := range routes {
			sameRoute := routeHash == route.Hash64()
			if sameRoute {
				return []Route{route}
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
			if len(route.RelayIDs) == 0 {
				continue
			}

			// Get the destination relay
			destRelayID := route.RelayIDs[len(route.RelayIDs)-1]

			// If the relay isn't in the map yet, add an empty slice to add routes to
			if _, ok := destRelayRouteMap[destRelayID]; !ok {
				destRelayRouteMap[destRelayID] = nil
			}

			// Append the route to the relay's entry in the map
			destRelayRouteMap[destRelayID] = append(destRelayRouteMap[destRelayID], route)
		}

		// Don't continue if there are no routes in the map
		if len(destRelayRouteMap) == 0 {
			return nil
		}

		// Get a slice of all destination relay IDs
		var destinationRelayIDs []uint64
		for destRelayID := range destRelayRouteMap {
			destinationRelayIDs = append(destinationRelayIDs, destRelayID)
		}

		// NOTE - Why does this need to be sorted if a random destination relay is chosen anyway?
		sort.Slice(destinationRelayIDs, func(i, j int) bool {
			return destinationRelayIDs[i] < destinationRelayIDs[j]
		})

		// choose a random destination relay, and use the routes from that
		relayRoutes := destRelayRouteMap[destinationRelayIDs[randgen.Intn(len(destinationRelayIDs))]]
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

		return []Route{routes[randgen.Intn(len(routes))]}
	}
}
