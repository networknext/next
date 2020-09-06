
package main

import (
    "runtime"
    "sync"
    "sort"
)

const MaxRelays = 5
const MaxRoutesPerRelayPair = 8

func TriMatrixLength(size int) int {
	return (size * (size - 1)) / 2
}

func TriMatrixIndex(i, j int) int {
	if i > j {
        return i*(i+1)/2 - i + j
	} else {
        return j*(j+1)/2 - j + i        
    }
}

type RouteEntry struct {
    DirectCost     int32
    NumRoutes      int32
    RouteCost      [MaxRoutesPerRelayPair]int32
    RouteNumRelays [MaxRoutesPerRelayPair]int32
    RouteRelays    [MaxRoutesPerRelayPair][MaxRelays]int32
}

type RouteManager struct {
    NumRoutes      int
    RouteCost      [MaxRoutesPerRelayPair]int32
    RouteHash      [MaxRoutesPerRelayPair]uint32
    RouteNumRelays [MaxRoutesPerRelayPair]int32
    RouteRelays    [MaxRoutesPerRelayPair][MaxRelays]int32
}

func (manager *RouteManager) AddRoute(cost int32, relays ...int32) {

    // todo: need to bring back the code to filter out routes with two relays in the same datacenter

    // IMPORTANT: Filter out routes with loops. They can happen *very* occasionally.
    loopCheck := make(map[int32]int, len(relays))
    for i := range relays {
        if _, exists := loopCheck[relays[i]]; exists {
            return
        }
        loopCheck[relays[i]] = 1
    }

    if manager.NumRoutes == 0 {

        // no routes yet. add the route

        manager.NumRoutes = 1
        manager.RouteCost[0] = cost
        manager.RouteHash[0] = RouteHash(relays...)
        manager.RouteNumRelays[0] = int32(len(relays))
        for i := range relays {
            manager.RouteRelays[0][i] = relays[i]
        }

    } else if manager.NumRoutes < MaxRoutesPerRelayPair {

        // not at max routes yet. insert according cost sort order

        hash := RouteHash(relays...)
        for i := 0; i < manager.NumRoutes; i++ {
            if hash == manager.RouteHash[i] {
                return
            }
        }

        if cost >= manager.RouteCost[manager.NumRoutes-1] {

            // cost is greater than existing entries. append.

            manager.RouteCost[manager.NumRoutes] = cost
            manager.RouteHash[manager.NumRoutes] = hash
            manager.RouteNumRelays[manager.NumRoutes] = int32(len(relays))
            for i := range relays {
                manager.RouteRelays[manager.NumRoutes][i] = relays[i]
            }
            manager.NumRoutes++

        } else {

            // cost is lower than at least one entry. insert.

            insertIndex := manager.NumRoutes - 1
            for {
                if insertIndex == 0 || cost > manager.RouteCost[insertIndex-1] {
                    break
                }
                insertIndex--
            }
            manager.NumRoutes++
            for i := manager.NumRoutes - 1; i > insertIndex; i-- {
                manager.RouteCost[i] = manager.RouteCost[i-1]
                manager.RouteHash[i] = manager.RouteHash[i-1]
                manager.RouteNumRelays[i] = manager.RouteNumRelays[i-1]
                for j := 0; j < int(manager.RouteNumRelays[i]); j++ {
                    manager.RouteRelays[i][j] = manager.RouteRelays[i-1][j]
                }
            }
            manager.RouteCost[insertIndex] = cost
            manager.RouteHash[insertIndex] = hash
            manager.RouteNumRelays[insertIndex] = int32(len(relays))
            for i := range relays {
                manager.RouteRelays[insertIndex][i] = relays[i]
            }

        }

    } else {

        // route set is full. only insert if lower cost than at least one current route.

        if cost >= manager.RouteCost[manager.NumRoutes-1] {
            return
        }

        hash := RouteHash(relays...)
        for i := 0; i < manager.NumRoutes; i++ {
            if hash == manager.RouteHash[i] {
                return
            }
        }

        insertIndex := manager.NumRoutes - 1
        for {
            if insertIndex == 0 || cost > manager.RouteCost[insertIndex-1] {
                break
            }
            insertIndex--
        }

        for i := manager.NumRoutes - 1; i > insertIndex; i-- {
            manager.RouteCost[i] = manager.RouteCost[i-1]
            manager.RouteHash[i] = manager.RouteHash[i-1]
            manager.RouteNumRelays[i] = manager.RouteNumRelays[i-1]
            for j := 0; j < int(manager.RouteNumRelays[i]); j++ {
                manager.RouteRelays[i][j] = manager.RouteRelays[i-1][j]
            }
        }

        manager.RouteCost[insertIndex] = cost
        manager.RouteHash[insertIndex] = hash
        manager.RouteNumRelays[insertIndex] = int32(len(relays))

        for i := range relays {
            manager.RouteRelays[insertIndex][i] = relays[i]
        }

    }
}

func RouteHash(relays ...int32) uint32 {
    const prime = uint32(16777619) 
    const offset = uint32(2166136261)
    hash := uint32(0)
    for i := range relays {
        hash ^= uint32(relays[i]>>24) & 0xFF
        hash *= prime
        hash ^= uint32(relays[i]>>16) & 0xFF
        hash *= prime
        hash ^= uint32(relays[i]>>8) & 0xFF
        hash *= prime
        hash ^= uint32(relays[i]) & 0xFF
        hash *= prime
    }
    return hash
}

func Optimize(numRelays int, cost []int32) []RouteEntry {

    // build a matrix of indirect routes from relays i -> j that have lower cost than direct, eg. i -> (x) -> j, where x is every other relay

    type Indirect struct {
        relay int32
        cost  int32
    }

    indirect := make([][][]Indirect, numRelays)

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

    costThreshold := int32(5)

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

                    ijIndex := TriMatrixIndex(i, j)

                    numRoutes := 0
                    costDirect := cost[ijIndex]

                    if costDirect < 0 {

                        // no direct route exists between i,j. subdivide valid routes so we don't miss indirect paths.

                        for k := 0; k < numRelays; k++ {
                            if k == i || k == j {
                                continue
                            }
                            ikIndex := TriMatrixIndex(i, k)
                            kjIndex := TriMatrixIndex(k, j)
                            ikCost := cost[ikIndex]
                            kjCost := cost[kjIndex]
                            if ikCost < 0 || kjCost < 0 {
                                continue
                            }
                            working[numRoutes].relay = int32(k)
                            working[numRoutes].cost = int32(ikCost + kjCost)
                            numRoutes++
                        }

                    } else {

                        // direct route exists between i,j. subdivide only when a significant cost reduction occurs.

                        for k := 0; k < numRelays; k++ {
                            if k == i || k == j {
                                continue
                            }
                            ikIndex := TriMatrixIndex(i, k)
                            ikCost := cost[ikIndex]
                            if ikCost < 0 {
                                continue
                            }
                            kjIndex := TriMatrixIndex(k, j)
                            kjCost := cost[kjIndex]
                            if kjCost < 0 {
                                continue
                            }
                            indirectCost := ikCost + kjCost
                            if indirectCost > costDirect-costThreshold {
                                continue
                            }
                            working[numRoutes].relay = int32(k)
                            working[numRoutes].cost = indirectCost
                            numRoutes++
                        }

                    }

                    if numRoutes > 0 {
                        indirect[i][j] = make([]Indirect, numRoutes)
                        copy(indirect[i][j], working)
                        sort.Slice(indirect[i][j], func(a, b int) bool { return indirect[i][j][a].cost < indirect[i][j][b].cost })
                    }
                }
            }

        }(startIndex, endIndex)
    }

    wg.Wait()

    // use the indirect matrix to subdivide a route up to 5 hops

    entryCount := TriMatrixLength(numRelays)

    routes := make([]RouteEntry, entryCount)

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

                    ijIndex := TriMatrixIndex(i, j)

                    if indirect[i][j] == nil {

                        if cost[ijIndex] >= 0 {

                            // only direct route from i -> j exists, and it is suitable

                            routes[ijIndex].DirectCost = cost[ijIndex]
                            routes[ijIndex].NumRoutes = 1
                            routes[ijIndex].RouteCost[0] = cost[ijIndex]
                            routes[ijIndex].RouteNumRelays[0] = 2
                            routes[ijIndex].RouteRelays[0][0] = int32(i)
                            routes[ijIndex].RouteRelays[0][1] = int32(j)

                        } else {

                            // no route exists from i -> j

                        }

                    } else {

                        // subdivide routes from i -> j as follows: i -> (x) -> (y) -> (z) -> j, where the subdivision improves significantly on cost

                        var routeManager RouteManager

                        for k := range indirect[i][j] {

                            if cost[ijIndex] >= 0 {
                                routeManager.AddRoute(cost[ijIndex], int32(i), int32(j))
                            }

                            y := indirect[i][j][k]

                            routeManager.AddRoute(y.cost, int32(i), y.relay, int32(j))

                            var x *Indirect
                            if indirect[i][y.relay] != nil {
                                x = &indirect[i][y.relay][0]
                            }

                            var z *Indirect
                            if indirect[j][y.relay] != nil {
                                z = &indirect[j][y.relay][0]
                            }

                            if x != nil {
                                ixIndex := TriMatrixIndex(i, int(x.relay))
                                xyIndex := TriMatrixIndex(int(x.relay), int(y.relay))
                                yjIndex := TriMatrixIndex(int(y.relay), j)

                                routeManager.AddRoute(cost[ixIndex]+cost[xyIndex]+cost[yjIndex], int32(i), x.relay, y.relay, int32(j))
                            }

                            if z != nil {
                                iyIndex := TriMatrixIndex(i, int(y.relay))
                                yzIndex := TriMatrixIndex(int(y.relay), int(z.relay))
                                zjIndex := TriMatrixIndex(int(z.relay), j)

                                routeManager.AddRoute(cost[iyIndex]+cost[yzIndex]+cost[zjIndex], int32(i), y.relay, z.relay, int32(j))
                            }

                            if x != nil && z != nil {
                                ixIndex := TriMatrixIndex(i, int(x.relay))
                                xyIndex := TriMatrixIndex(int(x.relay), int(y.relay))
                                yzIndex := TriMatrixIndex(int(y.relay), int(z.relay))
                                zjIndex := TriMatrixIndex(int(z.relay), j)

                                routeManager.AddRoute(cost[ixIndex]+cost[xyIndex]+cost[yzIndex]+cost[zjIndex], int32(i), x.relay, y.relay, z.relay, int32(j))
                            }

                            numRoutes := routeManager.NumRoutes

                            routes[ijIndex].DirectCost = cost[ijIndex]

                            routes[ijIndex].NumRoutes = int32(numRoutes)

                            for u := 0; u < numRoutes; u++ {
                                routes[ijIndex].RouteCost[u] = routeManager.RouteCost[u]
                                routes[ijIndex].RouteNumRelays[u] = routeManager.RouteNumRelays[u]
                                numRelays := int(routes[ijIndex].RouteNumRelays[u])
                                for v := 0; v < numRelays; v++ {
                                    routes[ijIndex].RouteRelays[u][v] = routeManager.RouteRelays[u][v]
                                }
                            }
                        }
                    }
                }
            }

        }(startIndex, endIndex)
    }

    wg.Wait()

    return routes
}

func Analyze(numRelays int, routes []RouteEntry) []int {

    buckets := make([]int, 6)

    for i := 0; i < numRelays; i++ {
        for j := 0; j < numRelays; j++ {
            if j < i {
                abFlatIndex := TriMatrixIndex(i, j)
                if len(routes[abFlatIndex].RouteCost) > 0 {
                    improvement := routes[abFlatIndex].DirectCost - routes[abFlatIndex].RouteCost[0]
                    if improvement <= 10 {
                        buckets[0]++
                    } else if improvement <= 20 {
                        buckets[1]++
                    } else if improvement <= 30 {
                        buckets[2]++
                    } else if improvement <= 40 {
                        buckets[3]++
                    } else if improvement <= 50 {
                        buckets[4]++
                    } else {
                        buckets[5]++
                    }
                }
            }
        }
    }

    return buckets

}
