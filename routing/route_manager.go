package routing

const (
	MaxRelays = 5
	MaxRoutesPerRelayPair = 8
)

type RouteManager struct {
	NumRoutes      int
	RouteRTT       [MaxRoutesPerRelayPair]int32
	RouteHash      [MaxRoutesPerRelayPair]uint64
	RouteNumRelays [MaxRoutesPerRelayPair]int32
	RouteRelays    [MaxRoutesPerRelayPair][MaxRelays]uint64
}

func (manager *RouteManager) AddRoute(rtt int32, relays ...uint64) {
	
	if rtt < 0 {
		return
	}

	// IMPORTANT: Filter out routes with loops. They can happen *very* occasionally.
	loopCheck := make(map[uint64]int, len(relays))
	for i := range relays {
		if _, exists := loopCheck[relays[i]]; exists {
			return
		}
		loopCheck[relays[i]] = 1
	}

	if manager.NumRoutes == 0 {

		// no routes yet. add the route

		manager.NumRoutes = 1
		manager.RouteRTT[0] = rtt
		manager.RouteHash[0] = routeHash(relays...)
		manager.RouteNumRelays[0] = int32(len(relays))
		for i := range relays {
			manager.RouteRelays[0][i] = relays[i]
		}

	} else if manager.NumRoutes < MaxRoutesPerRelayPair {

		// not at max routes yet. insert according RTT sort order

		hash := routeHash(relays...)
		for i := 0; i < manager.NumRoutes; i++ {
			if hash == manager.RouteHash[i] {
				return
			}
		}

		if rtt >= manager.RouteRTT[manager.NumRoutes-1] {

			// RTT is greater than existing entries. append.

			manager.RouteRTT[manager.NumRoutes] = rtt
			manager.RouteHash[manager.NumRoutes] = hash
			manager.RouteNumRelays[manager.NumRoutes] = int32(len(relays))
			for i := range relays {
				manager.RouteRelays[manager.NumRoutes][i] = relays[i]
			}
			manager.NumRoutes++

		} else {

			// RTT is lower than at least one entry. insert.

			insertIndex := manager.NumRoutes - 1
			for {
				if insertIndex == 0 || rtt > manager.RouteRTT[insertIndex-1] {
					break
				}
				insertIndex--
			}
			manager.NumRoutes++
			for i := manager.NumRoutes - 1; i > insertIndex; i-- {
				manager.RouteRTT[i] = manager.RouteRTT[i-1]
				manager.RouteHash[i] = manager.RouteHash[i-1]
				manager.RouteNumRelays[i] = manager.RouteNumRelays[i-1]
				for j := 0; j < int(manager.RouteNumRelays[i]); j++ {
					manager.RouteRelays[i][j] = manager.RouteRelays[i-1][j]
				}
			}
			manager.RouteRTT[insertIndex] = rtt
			manager.RouteHash[insertIndex] = hash
			manager.RouteNumRelays[insertIndex] = int32(len(relays))
			for i := range relays {
				manager.RouteRelays[insertIndex][i] = relays[i]
			}

		}

	} else {

		// route set is full. only insert if lower RTT than at least one current route.

		if rtt >= manager.RouteRTT[manager.NumRoutes-1] {
			return
		}

		hash := routeHash(relays...)
		for i := 0; i < manager.NumRoutes; i++ {
			if hash == manager.RouteHash[i] {
				return
			}
		}

		insertIndex := manager.NumRoutes - 1
		for {
			if insertIndex == 0 || rtt > manager.RouteRTT[insertIndex-1] {
				break
			}
			insertIndex--
		}

		for i := manager.NumRoutes - 1; i > insertIndex; i-- {
			manager.RouteRTT[i] = manager.RouteRTT[i-1]
			manager.RouteHash[i] = manager.RouteHash[i-1]
			manager.RouteNumRelays[i] = manager.RouteNumRelays[i-1]
			for j := 0; j < int(manager.RouteNumRelays[i]); j++ {
				manager.RouteRelays[i][j] = manager.RouteRelays[i-1][j]
			}
		}

		manager.RouteRTT[insertIndex] = rtt
		manager.RouteHash[insertIndex] = hash
		manager.RouteNumRelays[insertIndex] = int32(len(relays))

		for i := range relays {
			manager.RouteRelays[insertIndex][i] = relays[i]
		}

	}
}

func routeHash(relays ...uint64) uint64 {
	// http://www.isthe.com/chongo/tech/comp/fnv/
	const fnv64OffsetBasis = uint64(0xCBF29CE484222325)
	hash := uint64(0)
	for i := range relays {
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 56) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 48) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 40) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 32) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 24) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 16) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 8) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= relays[i] & 0xFF
	}
	return hash
}
