package core

import (
	"runtime"
	"sort"
	"sync"
)

func RouteHash(relays ...uint32) uint32 {
	hash := uint32(0)
	for i := range relays {
		hash *= uint32(0x811C9DC5)
		hash ^= (relays[i] >> 24) & 0xFF
		hash *= uint32(0x811C9DC5)
		hash ^= (relays[i] >> 16) & 0xFF
		hash *= uint32(0x811C9DC5)
		hash ^= (relays[i] >> 8) & 0xFF
		hash *= uint32(0x811C9DC5)
		hash ^= relays[i] & 0xFF
	}
	return hash
}

type RouteManager struct {
	NumRoutes      int
	RouteRTT       [MaxRoutesPerRelayPair]int32
	RouteHash      [MaxRoutesPerRelayPair]uint32
	RouteNumRelays [MaxRoutesPerRelayPair]int32
	RouteRelays    [MaxRoutesPerRelayPair][MaxRelays]uint32
}

func NewRouteManager() *RouteManager {
	manager := &RouteManager{}
	return manager
}

func (manager *RouteManager) AddRoute(rtt int32, relays ...uint32) {
	if rtt < 0 {
		return
	}
	if manager.NumRoutes == 0 {

		// no routes yet. add the route

		manager.NumRoutes = 1
		manager.RouteRTT[0] = rtt
		manager.RouteHash[0] = RouteHash(relays...)
		manager.RouteNumRelays[0] = int32(len(relays))
		for i := range relays {
			manager.RouteRelays[0][i] = relays[i]
		}

	} else if manager.NumRoutes < MaxRoutesPerRelayPair {

		// not at max routes yet. insert according RTT sort order

		routeHash := RouteHash(relays...)
		for i := 0; i < manager.NumRoutes; i++ {
			if routeHash == manager.RouteHash[i] {
				return
			}
		}

		if rtt >= manager.RouteRTT[manager.NumRoutes-1] {

			// RTT is greater than existing entries. append.

			manager.RouteRTT[manager.NumRoutes] = rtt
			manager.RouteHash[manager.NumRoutes] = routeHash
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
			manager.RouteHash[insertIndex] = routeHash
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

		routeHash := RouteHash(relays...)
		for i := 0; i < manager.NumRoutes; i++ {
			if routeHash == manager.RouteHash[i] {
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
		manager.RouteHash[insertIndex] = routeHash
		manager.RouteNumRelays[insertIndex] = int32(len(relays))

		for i := range relays {
			manager.RouteRelays[insertIndex][i] = relays[i]
		}

	}
}

func Optimize(costMatrix *CostMatrix, thresholdRTT int32) *RouteMatrix {

	numRelays := len(costMatrix.RelayIds)

	entryCount := TriMatrixLength(numRelays)

	result := &RouteMatrix{}
	result.RelayIds = costMatrix.RelayIds
	result.RelayNames = costMatrix.RelayNames
	result.RelayAddresses = costMatrix.RelayAddresses
	result.RelayPublicKeys = costMatrix.RelayPublicKeys
	result.DatacenterIds = costMatrix.DatacenterIds
	result.DatacenterNames = costMatrix.DatacenterNames
	result.DatacenterRelays = costMatrix.DatacenterRelays
	result.Entries = make([]RouteMatrixEntry, entryCount)

	type Indirect struct {
		relay int32
		rtt   int32
	}

	rtt := costMatrix.RTT

	indirect := make([][][]Indirect, numRelays)

	// phase 1: build a matrix of indirect routes from relays i -> j that have lower rtt than direct, eg. i -> (x) -> j, where x is every other relay

	numCPUs := runtime.NumCPU()

	numSegments := numRelays
	if numCPUs < numRelays {
		numSegments = numRelays / 5
		if numSegments == 0 {
			numSegments = 1
		}
	}

	var wg sync.WaitGroup

	wg.Add(numSegments)

	for segment := 0; segment < numSegments; segment++ {

		startIndex := segment * numRelays / numSegments
		endIndex := (segment+1)*numRelays/numSegments - 1
		if segment == numSegments-1 {
			endIndex = numRelays - 1
		}

		go func(startIndex int, endIndex int) {

			defer wg.Done()

			working := make([]Indirect, numRelays)

			for i := startIndex; i <= endIndex; i++ {

				indirect[i] = make([][]Indirect, numRelays)

				for j := 0; j < numRelays; j++ {

					// can't route to self
					if i == j {
						continue
					}

					ij_index := TriMatrixIndex(i, j)

					numRoutes := 0
					rtt_direct := rtt[ij_index]

					if rtt_direct < 0 {

						// no direct route exists between i,j. subdivide valid routes so we don't miss indirect paths.

						for k := 0; k < numRelays; k++ {
							if k == i || k == j {
								continue
							}
							ik_index := TriMatrixIndex(i, k)
							kj_index := TriMatrixIndex(k, j)
							ik_rtt := rtt[ik_index]
							kj_rtt := rtt[kj_index]
							if ik_rtt < 0 || kj_rtt < 0 {
								continue
							}
							working[numRoutes].relay = int32(k)
							working[numRoutes].rtt = ik_rtt + kj_rtt
							numRoutes++
						}

					} else {

						// direct route exists between i,j. subdivide only when a significant rtt reduction occurs.

						for k := 0; k < numRelays; k++ {
							if k == i || k == j {
								continue
							}
							ik_index := TriMatrixIndex(i, k)
							kj_index := TriMatrixIndex(k, j)
							ik_rtt := rtt[ik_index]
							kj_rtt := rtt[kj_index]
							if ik_rtt < 0 || kj_rtt < 0 {
								continue
							}
							indirectRTT := ik_rtt + kj_rtt
							if indirectRTT > rtt_direct-thresholdRTT {
								continue
							}
							working[numRoutes].relay = int32(k)
							working[numRoutes].rtt = indirectRTT
							numRoutes++
						}

					}

					if numRoutes > 0 {
						indirect[i][j] = make([]Indirect, numRoutes)
						copy(indirect[i][j], working)
						sort.Slice(indirect[i][j], func(a, b int) bool { return indirect[i][j][a].rtt < indirect[i][j][b].rtt })
					}
				}
			}

		}(startIndex, endIndex)
	}

	wg.Wait()

	// phase 2: use the indirect matrix to subdivide a route up to 5 hops

	wg.Add(numSegments)

	for segment := 0; segment < numSegments; segment++ {

		startIndex := segment * numRelays / numSegments
		endIndex := (segment+1)*numRelays/numSegments - 1
		if segment == numSegments-1 {
			endIndex = numRelays - 1
		}

		go func(startIndex int, endIndex int) {

			defer wg.Done()

			for i := startIndex; i <= endIndex; i++ {

				for j := 0; j < i; j++ {

					ij_index := TriMatrixIndex(i, j)

					if indirect[i][j] == nil {

						if rtt[ij_index] >= 0 {

							// only direct route from i -> j exists, and it is suitable

							result.Entries[ij_index].DirectRTT = rtt[ij_index]
							result.Entries[ij_index].NumRoutes = 1
							result.Entries[ij_index].RouteRTT[0] = rtt[ij_index]
							result.Entries[ij_index].RouteNumRelays[0] = 2
							result.Entries[ij_index].RouteRelays[0][0] = uint32(i)
							result.Entries[ij_index].RouteRelays[0][1] = uint32(j)

						}

					} else {

						// subdivide routes from i -> j as follows: i -> (x) -> (y) -> (z) -> j, where the subdivision improves significantly on RTT

						routeManager := NewRouteManager()

						for k := range indirect[i][j] {

							routeManager.AddRoute(rtt[ij_index], uint32(i), uint32(j))

							y := indirect[i][j][k]

							routeManager.AddRoute(y.rtt, uint32(i), uint32(y.relay), uint32(j))

							var x *Indirect
							if indirect[i][y.relay] != nil {
								x = &indirect[i][y.relay][0]
							}

							var z *Indirect
							if indirect[j][y.relay] != nil {
								z = &indirect[j][y.relay][0]
							}

							if x != nil {
								ix_index := TriMatrixIndex(i, int(x.relay))
								xy_index := TriMatrixIndex(int(x.relay), int(y.relay))
								yj_index := TriMatrixIndex(int(y.relay), j)

								routeManager.AddRoute(rtt[ix_index]+rtt[xy_index]+rtt[yj_index],
									uint32(i), uint32(x.relay), uint32(y.relay), uint32(j))
							}

							if z != nil {
								iy_index := TriMatrixIndex(i, int(y.relay))
								yz_index := TriMatrixIndex(int(y.relay), int(z.relay))
								zj_index := TriMatrixIndex(int(z.relay), j)

								routeManager.AddRoute(rtt[iy_index]+rtt[yz_index]+rtt[zj_index],
									uint32(i), uint32(y.relay), uint32(z.relay), uint32(j))
							}

							if x != nil && z != nil {
								ix_index := TriMatrixIndex(i, int(x.relay))
								xy_index := TriMatrixIndex(int(x.relay), int(y.relay))
								yz_index := TriMatrixIndex(int(y.relay), int(z.relay))
								zj_index := TriMatrixIndex(int(z.relay), j)

								routeManager.AddRoute(rtt[ix_index]+rtt[xy_index]+rtt[yz_index]+rtt[zj_index],
									uint32(i), uint32(x.relay), uint32(y.relay), uint32(z.relay), uint32(j))
							}

							numRoutes := routeManager.NumRoutes

							result.Entries[ij_index].DirectRTT = rtt[ij_index]
							result.Entries[ij_index].NumRoutes = int32(numRoutes)

							for u := 0; u < numRoutes; u++ {
								result.Entries[ij_index].RouteRTT[u] = routeManager.RouteRTT[u]
								result.Entries[ij_index].RouteNumRelays[u] = routeManager.RouteNumRelays[u]
								numRelays := int(result.Entries[ij_index].RouteNumRelays[u])
								for v := 0; v < numRelays; v++ {
									result.Entries[ij_index].RouteRelays[u][v] = routeManager.RouteRelays[u][v]
								}
							}
						}
					}
				}
			}

		}(startIndex, endIndex)
	}

	wg.Wait()

	return result
}
