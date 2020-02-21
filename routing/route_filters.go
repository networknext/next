package routing

import (
	"math"
)

// RouteFilter reduces a set of routes according to the filter function.
// Takes in an array of routes and returns the filtered array of routes.
// If the filter couldn't find a new set of routes, then it returns nil.
type RouteFilter func(routes []Route) []Route

// FilterBestRTT returns the best route based on lowest RTT, or nil if no best route is found.
func FilterBestRTT() RouteFilter {
	return func(routes []Route) []Route {
		var bestRoute *Route
		for _, route := range routes {
			if bestRoute == nil || route.Stats.RTT < bestRoute.Stats.RTT {
				bestRoute = &route
			}
		}

		if bestRoute == nil {
			return nil
		}

		return []Route{*bestRoute}
	}
}

// FilterRTTSwitchThreshold will get a list of acceptable routes, which is defined as all routes whose RTT is within the given threshold, and return a Route based on that list.
// If the current route is within this list of acceptable routes, then use that. Otherwise, use the one with the smallest RTT difference.
func FilterRTTSwitchThreshold(currentRoute *Route, rttSwitchThreshold float32) RouteFilter {
	return func(routes []Route) []Route {
		acceptableRoutes := make([]Route, 0)
		for _, route := range routes {
			rttDifference := currentRoute.Stats.RTT - route.Stats.RTT
			if math.Abs(rttDifference) < float64(rttSwitchThreshold) {
				acceptableRoutes = append(acceptableRoutes, route)
			}
		}

		// prefer to hold the same route if it is still acceptable
		for _, route := range acceptableRoutes {
			sameRoute := lastRouteHash == GetRouteHash(route.RelayIds)
			if sameRoute {
				return route, routes, acceptableRoutes
			}
		}
	}
}
