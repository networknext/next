package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"

	crypto_rand "crypto/rand"
	math_rand "math/rand"

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/crypto"
	"golang.org/x/crypto/chacha20poly1305"
)

var DebugLogs bool

func init() {
	value, ok := os.LookupEnv("DEBUG_LOGS")
	if ok && value == "1" {
		DebugLogs = true
	}
}

func Error(s string, params ...interface{}) {
	fmt.Printf("error: "+s+"\n", params...)
}

func Warn(s string, params ...interface{}) {
	fmt.Printf("warning: "+s+"\n", params...)
}

func Log(s string, params ...interface{}) {
	fmt.Printf(s+"\n", params...)
}

func Debug(s string, params ...interface{}) {
	if DebugLogs {
		fmt.Printf(s+"\n", params...)
	}
}

func ProtocolVersionAtLeast(serverMajor uint32, serverMinor uint32, serverPatch uint32, targetMajor uint32, targetMinor uint32, targetPatch uint32) bool {
	serverVersion := ((serverMajor & 0xFF) << 16) | ((serverMinor & 0xFF) << 8) | (serverPatch & 0xFF)
	targetVersion := ((targetMajor & 0xFF) << 16) | ((targetMinor & 0xFF) << 8) | (targetPatch & 0xFF)
	return serverVersion >= targetVersion
}

func HaversineDistance(lat1 float64, long1 float64, lat2 float64, long2 float64) float64 {
	lat1 *= math.Pi / 180
	lat2 *= math.Pi / 180
	long1 *= math.Pi / 180
	long2 *= math.Pi / 180
	delta_lat := lat2 - lat1
	delta_long := long2 - long1
	lat_sine := math.Sin(delta_lat / 2)
	long_sine := math.Sin(delta_long / 2)
	a := lat_sine*lat_sine + math.Cos(lat1)*math.Cos(lat2)*long_sine*long_sine
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	r := 6371.0
	d := r * c
	return d // kilometers
}

func SpeedOfLightTimeMilliseconds_AB(a_lat float64, a_long float64, b_lat float64, b_long float64) float64 {
	ab_distance_kilometers := HaversineDistance(a_lat, a_long, b_lat, b_long)
	speed_of_light_time_milliseconds := ab_distance_kilometers / 299792.458 * 1000.0
	return speed_of_light_time_milliseconds
}

func SpeedOfLightTimeMilliseconds_ABC(a_lat float64, a_long float64, b_lat float64, b_long float64, c_lat float64, c_long float64) float64 {
	ab_distance_kilometers := HaversineDistance(a_lat, a_long, b_lat, b_long)
	bc_distance_kilometers := HaversineDistance(b_lat, b_long, c_lat, c_long)
	total_distance_kilometers := ab_distance_kilometers + bc_distance_kilometers
	speed_of_light_time_milliseconds := total_distance_kilometers / 299792.458 * 1000.0
	return speed_of_light_time_milliseconds
}

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

func RandomBytes(buffer []byte) {
	crypto_rand.Read(buffer)
}

// -----------------------------------------------------

const (
	ADDRESS_NONE = 0
	ADDRESS_IPV4 = 1
	ADDRESS_IPV6 = 2
)

func ParseAddress(input string) net.UDPAddr {
	address := net.UDPAddr{}
	ip_string, port_string, err := net.SplitHostPort(input)
	if err != nil {
		address.IP = net.ParseIP(input)
		address.Port = 0
		return address
	}
	address.IP = net.ParseIP(ip_string)
	address.Port, _ = strconv.Atoi(port_string)
	return address
}

func WriteAddress(buffer []byte, address *net.UDPAddr) {
	if address == nil {
		buffer[0] = ADDRESS_NONE
		return
	}
	ipv4 := address.IP.To4()
	port := address.Port
	if ipv4 != nil {
		buffer[0] = ADDRESS_IPV4
		buffer[1] = ipv4[0]
		buffer[2] = ipv4[1]
		buffer[3] = ipv4[2]
		buffer[4] = ipv4[3]
		buffer[5] = (byte)(port & 0xFF)
		buffer[6] = (byte)(port >> 8)
	} else {
		buffer[0] = ADDRESS_IPV6
		copy(buffer[1:], address.IP)
		buffer[17] = (byte)(port & 0xFF)
		buffer[18] = (byte)(port >> 8)
	}
}

func ReadAddress(buffer []byte) net.UDPAddr {
	addressType := buffer[0]
	if addressType == ADDRESS_IPV4 {
		return net.UDPAddr{IP: net.IPv4(buffer[1], buffer[2], buffer[3], buffer[4]), Port: ((int)(binary.LittleEndian.Uint16(buffer[5:])))}
	} else if addressType == ADDRESS_IPV6 {
		return net.UDPAddr{IP: buffer[1:17], Port: ((int)(binary.LittleEndian.Uint16(buffer[17:19])))}
	}
	return net.UDPAddr{}
}

func WriteAddress_IPv4(buffer []byte, address *net.UDPAddr) {
	ipv4 := address.IP.To4()
	port := address.Port
	buffer[0] = ipv4[0]
	buffer[1] = ipv4[1]
	buffer[2] = ipv4[2]
	buffer[3] = ipv4[3]
	buffer[4] = (byte)(port & 0xFF)
	buffer[5] = (byte)(port >> 8)
}

func ReadAddress_IPv4(buffer []byte) net.UDPAddr {
	return net.UDPAddr{IP: net.IPv4(buffer[0], buffer[1], buffer[2], buffer[3]), Port: ((int)(binary.LittleEndian.Uint16(buffer[4:])))}
}

func AnonymizeAddress(address net.UDPAddr) net.UDPAddr {
	ipv4 := address.IP.To4()
	if ipv4 != nil {
		return net.UDPAddr{IP: net.IPv4(ipv4[0], ipv4[1], ipv4[2], 0), Port: 0}
	} else {
		address.Port = 0
		address.IP[6] = 0
		address.IP[7] = 0
		address.IP[8] = 0
		address.IP[9] = 0
		address.IP[10] = 0
		address.IP[11] = 0
		address.IP[12] = 0
		address.IP[13] = 0
		address.IP[14] = 0
		address.IP[15] = 0
		return address
	}
}

// ---------------------------------------------------

type RouteManager struct {
	NumRoutes       int
	RouteCost       [constants.MaxRoutesPerEntry]int32
	RoutePrice      [constants.MaxRoutesPerEntry]int32
	RouteHash       [constants.MaxRoutesPerEntry]uint32
	RouteNumRelays  [constants.MaxRoutesPerEntry]int32
	RouteRelays     [constants.MaxRoutesPerEntry][constants.MaxRouteRelays]int32
	RelayDatacenter []uint64
}

func (manager *RouteManager) AddRoute(cost int32, price int32, relays ...int32) {

	// no routes above cost 255 are allowed
	if cost >= 255 {
		return
	}

	// filter out any loops (yes, they can happen...)
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
		manager.RoutePrice[0] = price
		manager.RouteHash[0] = RouteHash(relays...)
		manager.RouteNumRelays[0] = int32(len(relays))
		for i := range relays {
			manager.RouteRelays[0][i] = relays[i]
		}

	} else if manager.NumRoutes < constants.MaxRoutesPerEntry {

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
			manager.RoutePrice[manager.NumRoutes] = price
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
				manager.RoutePrice[i] = manager.RoutePrice[i-1]
				manager.RouteHash[i] = manager.RouteHash[i-1]
				manager.RouteNumRelays[i] = manager.RouteNumRelays[i-1]
				for j := 0; j < int(manager.RouteNumRelays[i]); j++ {
					manager.RouteRelays[i][j] = manager.RouteRelays[i-1][j]
				}
			}
			manager.RouteCost[insertIndex] = cost
			manager.RoutePrice[insertIndex] = price
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
			manager.RoutePrice[i] = manager.RoutePrice[i-1]
			manager.RouteHash[i] = manager.RouteHash[i-1]
			manager.RouteNumRelays[i] = manager.RouteNumRelays[i-1]
			for j := 0; j < int(manager.RouteNumRelays[i]); j++ {
				manager.RouteRelays[i][j] = manager.RouteRelays[i-1][j]
			}
		}

		manager.RouteCost[insertIndex] = cost
		manager.RoutePrice[insertIndex] = price
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

type RouteEntry struct {
	DirectCost     int32
	NumRoutes      int32
	RouteCost      [constants.MaxRoutesPerEntry]int32
	RoutePrice     [constants.MaxRoutesPerEntry]int32
	RouteHash      [constants.MaxRoutesPerEntry]uint32
	RouteNumRelays [constants.MaxRoutesPerEntry]int32
	RouteRelays    [constants.MaxRoutesPerEntry][constants.MaxRouteRelays]int32
}

func Optimize(numRelays int, numSegments int, cost []uint8, relayPrice []uint8, relayDatacenter []uint64) []RouteEntry {

	// build a matrix of indirect routes from relays i -> j that have lower cost than direct, eg. i -> (x) -> j, where x is every other relay

	type Indirect struct {
		relay int32
		cost  uint32
	}

	indirect := make([][][]Indirect, numRelays)

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

					ijIndex := TriMatrixIndex(i, j)

					numRoutes := 0
					costDirect := uint32(cost[ijIndex])

					for x := 0; x < numRelays; x++ {
						if x == i || x == j {
							continue
						}
						ixIndex := TriMatrixIndex(i, x)
						ixCost := uint32(cost[ixIndex])
						xjIndex := TriMatrixIndex(x, j)
						xjCost := uint32(cost[xjIndex])
						indirectCost := uint32(ixCost) + uint32(xjCost)
						if indirectCost >= costDirect {
							continue
						}
						working[numRoutes].relay = int32(x)
						working[numRoutes].cost = indirectCost
						numRoutes++
					}

					if numRoutes > constants.MaxIndirects {
						sort.SliceStable(working, func(i, j int) bool { return working[i].cost < working[j].cost })
						copy(indirect[i][j], working[:constants.MaxIndirects])
					} else if numRoutes > 0 {
						indirect[i][j] = make([]Indirect, numRoutes)
						copy(indirect[i][j], working)
					}
				}
			}

		}(startIndex, endIndex)
	}

	wg.Wait()

	// use the indirect matrix to subdivide routes

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

					var routeManager RouteManager

					routeManager.RelayDatacenter = relayDatacenter

					// add the direct route

					index := TriMatrixIndex(i, j)

					directCost := int32(cost[index])

					if directCost < 255 {
						routeManager.AddRoute(directCost, int32(relayPrice[i])+int32(relayPrice[j]), int32(i), int32(j))
					}

					// add subdivided routes

					for k_index := range indirect[i][j] {

						k := int(indirect[i][j][k_index].relay)

						ik_cost := cost[TriMatrixIndex(i, k)]
						kj_cost := cost[TriMatrixIndex(k, j)]

						// i -> (k) -> j
						{
							ikj_cost := indirect[i][j][k_index].cost
							cost := int32(ikj_cost)
							if cost < directCost {
								routeManager.AddRoute(cost, int32(relayPrice[i])+int32(relayPrice[k])+int32(relayPrice[j]), int32(i), int32(k), int32(j))
							}
						}

						// i -> (x) -> k    ->     j

						for x_index := range indirect[i][k] {

							x := indirect[i][k][x_index].relay
							ixk_cost := indirect[i][k][x_index].cost
							cost := int32(ixk_cost) + int32(kj_cost)
							if cost < directCost {
								routeManager.AddRoute(cost, int32(relayPrice[i])+int32(relayPrice[x])+int32(relayPrice[k])+int32(relayPrice[j]), int32(i), int32(x), int32(k), int32(j))
							}
						}

						// i        -> k -> (y) -> j

						for y_index := range indirect[k][j] {
							kyj_cost := indirect[k][j][y_index].cost
							y := indirect[k][j][y_index].relay
							cost := int32(ik_cost) + int32(kyj_cost)
							if cost < directCost {
								routeManager.AddRoute(cost, int32(relayPrice[i])+int32(relayPrice[k])+int32(relayPrice[y])+int32(relayPrice[j]), int32(i), int32(k), int32(y), int32(j))
							}
						}

						// i -> (x) -> k -> (y) -> j

						for x_index := range indirect[i][k] {
							ixk_cost := indirect[i][k][x_index].cost
							x := int(indirect[i][k][x_index].relay)
							for y_index := range indirect[k][j] {
								kyj_cost := indirect[k][j][y_index].cost
								y := int(indirect[k][j][y_index].relay)
								cost := int32(ixk_cost) + int32(kyj_cost)
								if cost < directCost {
									routeManager.AddRoute(cost, int32(relayPrice[i])+int32(relayPrice[x])+int32(relayPrice[k])+int32(relayPrice[y])+int32(relayPrice[j]), int32(i), int32(x), int32(k), int32(y), int32(j))
								}
							}
						}
					}

					// store the best routes in order of lowest to highest cost

					numRoutes := int(routeManager.NumRoutes)

					routes[index].DirectCost = int32(cost[index])
					routes[index].NumRoutes = int32(numRoutes)

					for u := 0; u < numRoutes; u++ {
						routes[index].RouteCost[u] = routeManager.RouteCost[u]
						routes[index].RoutePrice[u] = routeManager.RoutePrice[u]
						routes[index].RouteNumRelays[u] = routeManager.RouteNumRelays[u]
						numRelays := int(routes[index].RouteNumRelays[u])
						for v := 0; v < numRelays; v++ {
							routes[index].RouteRelays[u][v] = routeManager.RouteRelays[u][v]
						}
						routes[index].RouteHash[u] = routeManager.RouteHash[u]
					}
				}
			}

		}(startIndex, endIndex)
	}

	wg.Wait()

	return routes
}

func Optimize2(numRelays int, numSegments int, cost []uint8, relayPrice []uint8, relayDatacenter []uint64, destinationRelay []bool) []RouteEntry {

	// Same as "Optimize", but it only optimizes to relays marked as destination relays.

	// build a matrix of indirect routes from relays i -> j that have lower cost than direct, eg. i -> (x) -> j, where x is every other relay

	type Indirect struct {
		relay int32
		cost  uint32
	}

	indirect := make([][][]Indirect, numRelays)

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

					if !destinationRelay[i] && !destinationRelay[j] {
						continue
					}

					ijIndex := TriMatrixIndex(i, j)

					numRoutes := 0
					costDirect := uint32(cost[ijIndex])

					for x := 0; x < numRelays; x++ {
						if x == i || x == j {
							continue
						}
						ixIndex := TriMatrixIndex(i, x)
						ixCost := uint32(cost[ixIndex])
						xjIndex := TriMatrixIndex(x, j)
						xjCost := uint32(cost[xjIndex])
						indirectCost := uint32(ixCost) + uint32(xjCost)
						if indirectCost >= costDirect {
							continue
						}
						working[numRoutes].relay = int32(x)
						working[numRoutes].cost = indirectCost
						numRoutes++
					}

					if numRoutes > constants.MaxIndirects {
						sort.SliceStable(working, func(i, j int) bool { return working[i].cost < working[j].cost })
						copy(indirect[i][j], working[:constants.MaxIndirects])
					} else if numRoutes > 0 {
						indirect[i][j] = make([]Indirect, numRoutes)
						copy(indirect[i][j], working)
					}
				}
			}

		}(startIndex, endIndex)
	}

	wg.Wait()

	// use the indirect matrix to subdivide routes

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

					var routeManager RouteManager

					routeManager.RelayDatacenter = relayDatacenter

					// add the direct route

					index := TriMatrixIndex(i, j)

					directCost := int32(cost[index])

					if directCost < 255 {
						routeManager.AddRoute(directCost, int32(relayPrice[i])+int32(relayPrice[j]), int32(i), int32(j))
					}

					if destinationRelay[i] || destinationRelay[j] {

						// add subdivided routes

						for k_index := range indirect[i][j] {

							k := int(indirect[i][j][k_index].relay)

							ik_cost := cost[TriMatrixIndex(i, k)]
							kj_cost := cost[TriMatrixIndex(k, j)]

							// i -> (k) -> j
							{
								cost := int32(indirect[i][j][k_index].cost)
								if cost < directCost {
									routeManager.AddRoute(int32(cost), int32(relayPrice[i])+int32(relayPrice[k])+int32(relayPrice[j]), int32(i), int32(k), int32(j))
								}
							}

							// i -> (x) -> k    ->     j

							for x_index := range indirect[i][k] {

								x := indirect[i][k][x_index].relay
								cost := int32(indirect[i][k][x_index].cost)
								if cost < directCost {
									routeManager.AddRoute(int32(cost)+int32(kj_cost), int32(relayPrice[i])+int32(relayPrice[x])+int32(relayPrice[k])+int32(relayPrice[j]), int32(i), int32(x), int32(k), int32(j))
								}
							}

							// i        -> k -> (y) -> j

							for y_index := range indirect[k][j] {
								kyj_cost := indirect[k][j][y_index].cost
								y := indirect[k][j][y_index].relay
								cost := int32(ik_cost) + int32(kyj_cost)
								if cost < directCost {
									routeManager.AddRoute(cost, int32(relayPrice[i])+int32(relayPrice[k])+int32(relayPrice[y])+int32(relayPrice[j]), int32(i), int32(k), int32(y), int32(j))
								}
							}

							// i -> (x) -> k -> (y) -> j

							for x_index := range indirect[i][k] {
								ixk_cost := indirect[i][k][x_index].cost
								x := int(indirect[i][k][x_index].relay)
								for y_index := range indirect[k][j] {
									kyj_cost := indirect[k][j][y_index].cost
									y := int(indirect[k][j][y_index].relay)
									cost := int32(ixk_cost) + int32(kyj_cost)
									if cost < directCost {
										routeManager.AddRoute(cost, int32(relayPrice[i])+int32(relayPrice[x])+int32(relayPrice[k])+int32(relayPrice[y])+int32(relayPrice[j]), int32(i), int32(x), int32(k), int32(y), int32(j))
									}
								}
							}
						}
					}

					// store the best routes in order of lowest to highest cost

					numRoutes := int(routeManager.NumRoutes)

					routes[index].DirectCost = int32(cost[index])
					routes[index].NumRoutes = int32(numRoutes)

					for u := 0; u < numRoutes; u++ {
						routes[index].RouteCost[u] = routeManager.RouteCost[u]
						routes[index].RoutePrice[u] = routeManager.RoutePrice[u]
						routes[index].RouteNumRelays[u] = routeManager.RouteNumRelays[u]
						numRelays := int(routes[index].RouteNumRelays[u])
						for v := 0; v < numRelays; v++ {
							routes[index].RouteRelays[u][v] = routeManager.RouteRelays[u][v]
						}
						routes[index].RouteHash[u] = routeManager.RouteHash[u]
					}
				}
			}

		}(startIndex, endIndex)
	}

	wg.Wait()

	return routes
}

// ---------------------------------------------------

type RouteToken struct {
	ExpireTimestamp   uint64
	SessionId         uint64
	SessionVersion    uint8
	EnvelopeKbpsUp    uint32
	EnvelopeKbpsDown  uint32
	NextAddress       net.UDPAddr
	PrevAddress       net.UDPAddr
	NextInternal      uint8
	PrevInternal      uint8
	SessionPrivateKey [crypto.Box_PrivateKeySize]byte
}

type ContinueToken struct {
	ExpireTimestamp uint64
	SessionId       uint64
	SessionVersion  uint8
}

// -----------------------------------------------------------------------------

func WriteRouteToken(data *RouteToken, buffer []byte) {

	index := 0

	copy(buffer[index:], data.SessionPrivateKey[:])
	index += 32

	binary.LittleEndian.PutUint64(buffer[index:], data.ExpireTimestamp)
	index += 8

	binary.LittleEndian.PutUint64(buffer[index:], data.SessionId)
	index += 8

	binary.LittleEndian.PutUint32(buffer[index:], data.EnvelopeKbpsUp)
	index += 4

	binary.LittleEndian.PutUint32(buffer[index:], data.EnvelopeKbpsDown)
	index += 4

	copy(buffer[index:], GetAddressData(&data.NextAddress))
	index += 4

	copy(buffer[index:], GetAddressData(&data.PrevAddress))
	index += 4

	binary.BigEndian.PutUint16(buffer[index:], uint16(data.NextAddress.Port))
	index += 2

	binary.BigEndian.PutUint16(buffer[index:], uint16(data.PrevAddress.Port))
	index += 2

	buffer[index] = data.SessionVersion
	index += 1

	buffer[index] = data.NextInternal
	index += 1

	buffer[index] = data.PrevInternal
	index += 1
}

func ReadRouteToken(token *RouteToken, buffer []byte) {

	index := 0

	copy(token.SessionPrivateKey[:], buffer[index:])
	index += 32

	token.ExpireTimestamp = binary.LittleEndian.Uint64(buffer[index:])
	index += 8

	token.SessionId = binary.LittleEndian.Uint64(buffer[index:])
	index += 8

	token.EnvelopeKbpsUp = binary.LittleEndian.Uint32(buffer[index:])
	index += 4

	token.EnvelopeKbpsDown = binary.LittleEndian.Uint32(buffer[index:])
	index += 4

	nextAddress := binary.LittleEndian.Uint32(buffer[index:])
	index += 4

	prevAddress := binary.LittleEndian.Uint32(buffer[index:])
	index += 4

	nextPort := binary.BigEndian.Uint16(buffer[index:])
	index += 2

	prevPort := binary.BigEndian.Uint16(buffer[index:])
	index += 2

	token.SessionVersion = buffer[index]
	index += 1

	token.NextInternal = buffer[index]
	index += 1

	token.PrevInternal = buffer[index]
	index += 1

	token.NextAddress = net.UDPAddr{IP: net.IPv4(uint8(nextAddress&0xFF), uint8((nextAddress>>8)&0xFF), uint8((nextAddress>>16)&0xFF), uint8((nextAddress>>24)&0xFF)), Port: int(nextPort)}
	token.PrevAddress = net.UDPAddr{IP: net.IPv4(uint8(prevAddress&0xFF), uint8((prevAddress>>8)&0xFF), uint8((prevAddress>>16)&0xFF), uint8((prevAddress>>24)&0xFF)), Port: int(prevPort)}
}

func WriteEncryptedRouteToken(token *RouteToken, tokenData []byte, secretKey []byte) bool {

	data := make([]byte, constants.RouteTokenBytes)

	WriteRouteToken(token, data)

	aead, err := chacha20poly1305.NewX(secretKey)
	if err != nil {
		return false
	}

	nonce := make([]byte, aead.NonceSize(), constants.EncryptedRouteTokenBytes)
	if _, err := crypto_rand.Read(nonce[:aead.NonceSize()]); err != nil {
		return false
	}

	dest := nonce

	encryptedRouteToken := aead.Seal(dest, nonce, data, nil)

	copy(tokenData, encryptedRouteToken)

	return true
}

func ReadEncryptedRouteToken(token *RouteToken, tokenData []byte, secretKey []byte) bool {

	aead, err := chacha20poly1305.NewX(secretKey)
	if err != nil {
		return false
	}

	nonceSize := aead.NonceSize()

	tokenData = tokenData[:constants.EncryptedRouteTokenBytes]

	nonce, encrypted := tokenData[:nonceSize], tokenData[nonceSize:]

	output := make([]byte, 0, constants.RouteTokenBytes)

	decrypted, err := aead.Open(output, nonce, encrypted, nil)
	if err != nil {
		return false
	}

	ReadRouteToken(token, decrypted)

	return true
}

func WriteRouteTokens(tokenData []byte, expireTimestamp uint64, sessionId uint64, sessionVersion uint8, kbpsUp uint32, kbpsDown uint32, numNodes int, publicAddresses []net.UDPAddr, hasInternalAddress []bool, internalAddresses []net.UDPAddr, internalGroups []uint64, sellers []int, secretKeys [][]byte) {
	privateKey := [crypto.Box_PrivateKeySize]byte{}
	RandomBytes(privateKey[:])
	for i := 0; i < numNodes; i++ {
		var token RouteToken
		token.ExpireTimestamp = expireTimestamp
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.EnvelopeKbpsUp = kbpsUp
		token.EnvelopeKbpsDown = kbpsDown
		if i != 0 {
			if hasInternalAddress[i] && hasInternalAddress[i-1] && sellers[i] == sellers[i-1] && internalGroups[i] == internalGroups[i-1] {
				token.PrevAddress = internalAddresses[i-1]
				token.PrevInternal = 1
			} else {
				token.PrevAddress = publicAddresses[i-1]
			}
		} else {
			token.PrevAddress = net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 0}
		}
		if i != numNodes-1 {
			if hasInternalAddress[i] && hasInternalAddress[i+1] && sellers[i] == sellers[i+1] && internalGroups[i] == internalGroups[i+1] {
				token.NextAddress = internalAddresses[i+1]
				token.NextInternal = 1
			} else {
				token.NextAddress = publicAddresses[i+1]
			}
		} else {
			token.NextAddress = net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 0}
		}
		copy(token.SessionPrivateKey[:], privateKey[:])
		WriteEncryptedRouteToken(&token, tokenData[i*constants.EncryptedRouteTokenBytes:(i+1)*constants.EncryptedRouteTokenBytes], secretKeys[i])
	}
}

// -----------------------------------------------------------------------------

func WriteContinueToken(token *ContinueToken, buffer []byte) {
	binary.LittleEndian.PutUint64(buffer[0:], token.ExpireTimestamp)
	binary.LittleEndian.PutUint64(buffer[8:], token.SessionId)
	buffer[8+8] = token.SessionVersion
}

func ReadContinueToken(token *ContinueToken, buffer []byte) {
	token.ExpireTimestamp = binary.LittleEndian.Uint64(buffer[0:])
	token.SessionId = binary.LittleEndian.Uint64(buffer[8:])
	token.SessionVersion = buffer[8+8]
}

func WriteEncryptedContinueToken(token *ContinueToken, tokenData []byte, secretKey []byte) bool {

	data := make([]byte, constants.ContinueTokenBytes)

	WriteContinueToken(token, data)

	aead, err := chacha20poly1305.NewX(secretKey)
	if err != nil {
		return false
	}

	nonce := make([]byte, aead.NonceSize(), constants.EncryptedContinueTokenBytes)
	if _, err := crypto_rand.Read(nonce[:aead.NonceSize()]); err != nil {
		return false
	}

	dest := nonce

	encryptedContinueToken := aead.Seal(dest, nonce, data, nil)

	copy(tokenData, encryptedContinueToken)

	return true
}

func ReadEncryptedContinueToken(token *ContinueToken, tokenData []byte, secretKey []byte) bool {

	aead, err := chacha20poly1305.NewX(secretKey)
	if err != nil {
		return false
	}

	nonceSize := aead.NonceSize()

	tokenData = tokenData[:constants.EncryptedContinueTokenBytes]

	nonce, encrypted := tokenData[:nonceSize], tokenData[nonceSize:]

	output := make([]byte, 0, constants.ContinueTokenBytes)

	decrypted, err := aead.Open(output, nonce, encrypted, nil)
	if err != nil {
		return false
	}

	ReadContinueToken(token, decrypted)

	return true
}

func WriteContinueTokens(tokenData []byte, expireTimestamp uint64, sessionId uint64, sessionVersion uint8, numNodes int, secretKeys [][]byte) {
	for i := 0; i < numNodes; i++ {
		var token ContinueToken
		token.ExpireTimestamp = expireTimestamp
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		WriteEncryptedContinueToken(&token, tokenData[i*constants.EncryptedContinueTokenBytes:], secretKeys[i])
	}
}

// -----------------------------------------------------------------------------

func GetBestRouteCost(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32) int32 {

	bestRouteCost := int32(math.MaxInt32)

	if len(routeMatrix) == 0 {
		return bestRouteCost
	}

	for i := range sourceRelays {

		// IMPORTANT: RTT=255 is used to signal an unroutable source relay
		if sourceRelayCost[i] >= 255 {
			continue
		}

		sourceRelayIndex := sourceRelays[i]

		for j := range destRelays {

			destRelayIndex := destRelays[j]
			if sourceRelayIndex == destRelayIndex {
				continue
			}

			index := TriMatrixIndex(int(sourceRelayIndex), int(destRelayIndex))

			entry := &routeMatrix[index]

			if entry.NumRoutes > 0 {
				cost := sourceRelayCost[i] + entry.RouteCost[0] // first entry is always lowest cost
				if cost < bestRouteCost {
					bestRouteCost = cost
				}
			}
		}
	}

	if bestRouteCost == int32(math.MaxInt32) {
		return bestRouteCost
	}

	return bestRouteCost + constants.CostBias
}

func ReverseRoute(route []int32) {
	for i, j := 0, len(route)-1; i < j; i, j = i+1, j-1 {
		route[i], route[j] = route[j], route[i]
	}
}

func RouteExists(routeMatrix []RouteEntry, routeNumRelays int32, routeRelays [constants.MaxRouteRelays]int32, debug *string) bool {
	if len(routeMatrix) == 0 {
		return false
	}
	if routeRelays[0] < routeRelays[routeNumRelays-1] {
		ReverseRoute(routeRelays[:routeNumRelays])
	}
	sourceRelayIndex := routeRelays[0]
	destRelayIndex := routeRelays[routeNumRelays-1]
	index := TriMatrixIndex(int(sourceRelayIndex), int(destRelayIndex))
	entry := &routeMatrix[index]
	for i := 0; i < int(entry.NumRoutes); i++ {
		if entry.RouteNumRelays[i] == routeNumRelays {
			found := true
			for j := range routeRelays {
				if entry.RouteRelays[i][j] != routeRelays[j] {
					found = false
					break
				}
			}
			if found {
				return true
			}
		}
	}
	return false
}

func GetCurrentRouteCost(routeMatrix []RouteEntry, routeNumRelays int32, routeRelays [constants.MaxRouteRelays]int32, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, debug *string) int32 {

	// IMPORTANT: This shouldn't happen. Triaging...
	if len(routeRelays) == 0 {
		if debug != nil {
			*debug += "no route relays?\n"
		}
		return -1
	}

	// IMPORTANT: This can happen. Make sure we handle it without exploding
	if len(routeMatrix) == 0 {
		if debug != nil {
			*debug += "route matrix is empty\n"
		}
		return -1
	}

	// Find the cost to first relay in the route
	// IMPORTANT: A cost of 255 means that the source relay is not routable
	sourceCost := int32(1000)
	for i := range sourceRelays {
		if routeRelays[0] == sourceRelays[i] {
			sourceCost = sourceRelayCost[i]
			break
		}
	}
	if sourceCost >= 255 {
		if debug != nil {
			*debug += "source relay for route is no longer routable\n"
		}
		return -1
	}

	// The route matrix is triangular, so depending on the indices for the
	// source and dest relays in the route, we need to reverse the route
	if routeRelays[0] < routeRelays[routeNumRelays-1] {
		ReverseRoute(routeRelays[:routeNumRelays])
		destRelays, sourceRelays = sourceRelays, destRelays
	}

	// IMPORTANT: We have to handle this. If it's passed in we'll crash out otherwise
	sourceRelayIndex := routeRelays[0]
	destRelayIndex := routeRelays[routeNumRelays-1]
	if sourceRelayIndex == destRelayIndex {
		if debug != nil {
			*debug += "source and dest relays are the same\n"
		}
		return -1
	}

	// Speed things up by hashing the route and comparing that vs. checking route relays manually
	routeHash := RouteHash(routeRelays[:routeNumRelays]...)
	index := TriMatrixIndex(int(sourceRelayIndex), int(destRelayIndex))
	entry := &routeMatrix[index]
	for i := 0; i < int(entry.NumRoutes); i++ {
		if entry.RouteHash[i] != routeHash {
			continue
		}
		if entry.RouteNumRelays[i] != routeNumRelays {
			continue
		}
		return sourceCost + entry.RouteCost[i] + constants.CostBias
	}

	// We didn't find the route :(
	if debug != nil {
		*debug += "could not find route\n"
	}
	return -1
}

type BestRoute struct {
	Cost          int32
	Price         int32
	NumRelays     int32
	Relays        [constants.MaxRouteRelays]int32
	NeedToReverse bool
}

func GetBestRoutes(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, bestRoutes []BestRoute, numBestRoutes *int) {

	if len(routeMatrix) == 0 {
		*numBestRoutes = 0
		return
	}

	numRoutes := 0

	maxRoutes := len(bestRoutes)

	for i := range sourceRelays {

		// IMPORTANT: RTT = 255 signals the source relay is unroutable
		if sourceRelayCost[i] >= 255 {
			continue
		}

		firstRouteFromThisRelay := true

		sourceRelayIndex := sourceRelays[i]

		for j := range destRelays {

			destRelayIndex := destRelays[j]

			if sourceRelayIndex == destRelayIndex {
				continue
			}

			index := TriMatrixIndex(int(sourceRelayIndex), int(destRelayIndex))

			entry := &routeMatrix[index]

			for k := 0; k < int(entry.NumRoutes); k++ {

				cost := entry.RouteCost[k] + sourceRelayCost[i]

				if cost > maxCost {
					break
				}

				bestRoutes[numRoutes].Cost = cost
				bestRoutes[numRoutes].Price = entry.RoutePrice[k]
				bestRoutes[numRoutes].NumRelays = entry.RouteNumRelays[k]

				for l := 0; l < len(entry.RouteRelays[0]); l++ {
					bestRoutes[numRoutes].Relays[l] = entry.RouteRelays[k][l]
				}

				bestRoutes[numRoutes].NeedToReverse = sourceRelayIndex < destRelayIndex

				numRoutes++

				if firstRouteFromThisRelay {
					firstRouteFromThisRelay = false
				}

				if numRoutes == maxRoutes {
					*numBestRoutes = numRoutes
					return
				}
			}
		}
	}

	*numBestRoutes = numRoutes
}

// -------------------------------------------

func ReframeRoute(relayIdToIndex map[uint64]int32, routeRelayIds []uint64, out_routeRelays *[constants.MaxRouteRelays]int32) bool {
	for i := range routeRelayIds {
		relayIndex, ok := relayIdToIndex[routeRelayIds[i]]
		if !ok {
			return false
		}
		out_routeRelays[i] = relayIndex
	}
	return true
}

// -------------------------------------------

func FilterSourceRelays(directLatency int32, directJitter int32, directPacketLoss float32, sourceRelayId []uint64, sourceRelayLatency []int32, sourceRelayJitter []int32, sourceRelayPacketLoss []float32, filterSourceRelay []bool) {

	// IMPORTANT: Just to be sure...
	for i := range filterSourceRelay {
		filterSourceRelay[i] = false
	}

	// if direct has high packet loss, and *most* source relays have high packet loss
	// it's a temporary packet loss spike on the edge, and we should ignore it.

	directHasHighPacketLoss := directPacketLoss >= 1.0

	numRelaysWithHighPacketLoss := 0
	for i := range sourceRelayPacketLoss {
		if sourceRelayPacketLoss[i] >= 1.0 {
			numRelaysWithHighPacketLoss++
		}
	}

	if directHasHighPacketLoss && numRelaysWithHighPacketLoss > len(sourceRelayId)*2.0/3.0 {
		for i := range sourceRelayPacketLoss {
			sourceRelayPacketLoss[i] = 0.0
		}
	}

	// exclude relays with higher packet loss than direct

	for i := range sourceRelayPacketLoss {
		if sourceRelayPacketLoss[i] > directPacketLoss {
			filterSourceRelay[i] = true
		}
	}

	// exclude unsuitable source relays

	for i := range sourceRelayLatency {

		// IMPORTANT: In the past we've had problems where relays with no pings back (100% PL) reported as 0ms RTT
		// This no longer occurs but it makes me very nervous, so we have this check here historically just in case
		// Any latency with real RTT runs through ceil and thus must have an RTT value of 1ms or greater.

		// you say your latency is 0ms? I don't believe you!
		if sourceRelayLatency[i] <= 0 {
			filterSourceRelay[i] = true
			continue
		}
	}
}

func ReframeSourceRelays(relayIdToIndex map[uint64]int32, sourceRelayId []uint64, sourceRelayLatency []int32, excludeSourceRelay []bool, out_sourceRelays []int32, out_sourceRelayLatency []int32) {

	for i := range sourceRelayId {

		// any excluded relay cannot be routed through
		if excludeSourceRelay[i] {
			out_sourceRelayLatency[i] = 255
			out_sourceRelays[i] = -1
			continue
		}

		// you say your latency is 0ms? I don't believe you!
		if sourceRelayLatency[i] <= 0 {
			out_sourceRelayLatency[i] = 255
			out_sourceRelays[i] = -1
			continue
		}

		// clamp latency above 255ms
		if sourceRelayLatency[i] > 255 {
			out_sourceRelayLatency[i] = 255
			out_sourceRelays[i] = -1
			continue
		}

		// any source relay that no longer exists cannot be routed through
		relayIndex, ok := relayIdToIndex[sourceRelayId[i]]
		if !ok {
			out_sourceRelayLatency[i] = 255
			out_sourceRelays[i] = -1
			continue
		}

		out_sourceRelays[i] = relayIndex
		out_sourceRelayLatency[i] = sourceRelayLatency[i]
	}
}

// ----------------------------------------------

func FilterDestRelays(destRelayId []uint64, destRelayLatency []int32, destRelayJitter []int32, destRelayPacketLoss []float32, filterDestRelay []bool) {

	// exclude dest relays with high latency
	// IMPORTANT: dest relays should be in the same physical datacenter as the game server
	// if the dest relay has more than 2ms of round trip time, it's obviously not!

	for i := range destRelayLatency {
		if destRelayLatency[i] > 2 {
			filterDestRelay[i] = true
		}
	}

	// exclude dest relays with high jitter

	for i := range destRelayJitter {
		if destRelayJitter[i] > 10 {
			filterDestRelay[i] = true
		}
	}

	// exclude dest relays with packet loss

	for i := range destRelayPacketLoss {
		if destRelayPacketLoss[i] > 0.1 {
			filterDestRelay[i] = true
		}
	}
}

func ReframeDestRelays(relayIdToIndex map[uint64]int32, destRelayId []uint64, excludeDestRelay []bool, out_destRelays *[]int32) {

	for i := range destRelayId {

		// any excluded relay cannot be routed through
		if excludeDestRelay[i] {
			continue
		}

		// any dest relay that no longer exists cannot be routed through
		relayIndex, ok := relayIdToIndex[destRelayId[i]]
		if !ok {
			continue
		}

		*out_destRelays = append(*out_destRelays, relayIndex)
	}
}

// ----------------------------------------------

func GetRandomBestRoute(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, threshold int32, out_bestRouteCost *int32, out_bestRouteNumRelays *int32, out_bestRouteRelays *[constants.MaxRouteRelays]int32, debug *string) bool {

	if maxCost == -1 {
		return false
	}

	bestRouteCost := GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCost, destRelays)
	if debug != nil {
		*debug += fmt.Sprintf("best route cost is %d\n", bestRouteCost)
	}

	if bestRouteCost > maxCost {
		if debug != nil {
			*debug += fmt.Sprintf("could not find any next route <= max cost %d\n", maxCost)
		}
		*out_bestRouteCost = bestRouteCost
		return false
	}

	numBestRoutes := 0
	bestRoutes := make([]BestRoute, 1024)
	GetBestRoutes(routeMatrix, sourceRelays, sourceRelayCost, destRelays, bestRouteCost+threshold, bestRoutes, &numBestRoutes)
	if numBestRoutes == 0 {
		if debug != nil {
			*debug += "could not find any next routes\n"
		}
		return false
	}

	if debug != nil {
		numClientRelays := 0
		for i := range sourceRelays {
			if sourceRelayCost[i] != 255 {
				numClientRelays++
			}
		}
		*debug += fmt.Sprintf("found %d suitable routes in [%d,%d] from %d/%d client relays\n", numBestRoutes, bestRouteCost, bestRouteCost+threshold, numClientRelays, len(sourceRelays))
	}

	randomIndex := math_rand.Intn(numBestRoutes)

	*out_bestRouteCost = bestRoutes[randomIndex].Cost + constants.CostBias
	*out_bestRouteNumRelays = bestRoutes[randomIndex].NumRelays

	if !bestRoutes[randomIndex].NeedToReverse {
		copy(out_bestRouteRelays[:], bestRoutes[randomIndex].Relays[:bestRoutes[randomIndex].NumRelays])
	} else {
		numRouteRelays := bestRoutes[randomIndex].NumRelays
		for i := int32(0); i < numRouteRelays; i++ {
			out_bestRouteRelays[numRouteRelays-1-i] = bestRoutes[randomIndex].Relays[i]
		}
	}

	return true
}

func GetRandomBestRoute_LowestPrice(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, threshold int32, out_bestRouteCost *int32, out_bestRouteNumRelays *int32, out_bestRouteRelays *[constants.MaxRouteRelays]int32, debug *string) bool {

	if maxCost == -1 {
		return false
	}

	bestRouteCost := GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCost, destRelays)
	if debug != nil {
		*debug += fmt.Sprintf("best route cost is %d\n", bestRouteCost)
	}

	if bestRouteCost > maxCost {
		if debug != nil {
			*debug += fmt.Sprintf("could not find any next route <= max cost %d\n", maxCost)
		}
		*out_bestRouteCost = bestRouteCost
		return false
	}

	numBestRoutes := 0
	bestRoutes := make([]BestRoute, 1024)
	GetBestRoutes(routeMatrix, sourceRelays, sourceRelayCost, destRelays, bestRouteCost+threshold, bestRoutes, &numBestRoutes)
	if numBestRoutes == 0 {
		if debug != nil {
			*debug += "could not find any next routes\n"
		}
		return false
	}

	bestRoutes = bestRoutes[:numBestRoutes]

	if debug != nil {
		numClientRelays := 0
		for i := range sourceRelays {
			if sourceRelayCost[i] != 255 {
				numClientRelays++
			}
		}
		*debug += fmt.Sprintf("found %d suitable routes in [%d,%d] from %d/%d client relays\n", numBestRoutes, bestRouteCost, bestRouteCost+threshold, numClientRelays, len(sourceRelays))
	}

	// find only routes with the lowest price

	filteredBestRoutes := make([]BestRoute, 0, len(bestRoutes))

	lowestPrice := int32(1000000)

	for i := range bestRoutes {
		if bestRoutes[i].Price < lowestPrice {
			filteredBestRoutes = filteredBestRoutes[:0]
			filteredBestRoutes = append(filteredBestRoutes, bestRoutes[i])
			lowestPrice = bestRoutes[i].Price
		} else if bestRoutes[i].Price == lowestPrice {
			filteredBestRoutes = append(filteredBestRoutes, bestRoutes[i])
		}
	}

	numBestRoutes = len(filteredBestRoutes)

	// any route with price >= 255 is not selectable for a new route

	if lowestPrice >= 255 {
		*debug += fmt.Sprintf("lowest price is >= 255, found no selectable routes")
		return false
	}

	// randomly select between lowest price routes

	randomIndex := math_rand.Intn(numBestRoutes)

	*out_bestRouteCost = filteredBestRoutes[randomIndex].Cost + constants.CostBias
	*out_bestRouteNumRelays = filteredBestRoutes[randomIndex].NumRelays

	if !filteredBestRoutes[randomIndex].NeedToReverse {
		copy(out_bestRouteRelays[:], filteredBestRoutes[randomIndex].Relays[:filteredBestRoutes[randomIndex].NumRelays])
	} else {
		numRouteRelays := filteredBestRoutes[randomIndex].NumRelays
		for i := int32(0); i < numRouteRelays; i++ {
			out_bestRouteRelays[numRouteRelays-1-i] = filteredBestRoutes[randomIndex].Relays[i]
		}
	}

	return true
}

// --------------------------------------------------------------------------------------------------------------------

func GetBestRoute_Initial(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, selectThreshold int32, out_bestRouteCost *int32, out_bestRouteNumRelays *int32, out_bestRouteRelays *[constants.MaxRouteRelays]int32, debug *string) bool {

	return GetRandomBestRoute_LowestPrice(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, selectThreshold, out_bestRouteCost, out_bestRouteNumRelays, out_bestRouteRelays, debug)
}

func GetBestRoute_Update(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, selectThreshold int32, switchThreshold int32, currentRouteNumRelays int32, currentRouteRelays [constants.MaxRouteRelays]int32, out_updatedRouteCost *int32, out_updatedRouteNumRelays *int32, out_updatedRouteRelays *[constants.MaxRouteRelays]int32, debug *string) (routeChanged bool, routeLost bool) {

	// if the current route no longer exists, pick a new route

	currentRouteCost := GetCurrentRouteCost(routeMatrix, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCost, destRelays, debug)

	if currentRouteCost < 0 {
		if debug != nil {
			*debug += "current route no longer exists. picking a new random route\n"
		}
		GetRandomBestRoute(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, selectThreshold, out_updatedRouteCost, out_updatedRouteNumRelays, out_updatedRouteRelays, debug)
		routeChanged = true
		routeLost = true
		return
	}

	// if the current route is no longer within threshold of the best route, pick a new the route

	bestRouteCost := GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCost, destRelays)

	if int64(currentRouteCost) > int64(bestRouteCost)+int64(switchThreshold) {
		if debug != nil {
			*debug += fmt.Sprintf("current route no longer within switch threshold of best route. picking a new random route.\ncurrent route cost = %d, best route cost = %d, route switch threshold = %d\n", currentRouteCost, bestRouteCost, switchThreshold)
		}
		GetRandomBestRoute(routeMatrix, sourceRelays, sourceRelayCost, destRelays, bestRouteCost, selectThreshold, out_updatedRouteCost, out_updatedRouteNumRelays, out_updatedRouteRelays, debug)
		routeChanged = true
		return
	}

	// hold current route

	*out_updatedRouteCost = currentRouteCost
	*out_updatedRouteNumRelays = currentRouteNumRelays
	copy(out_updatedRouteRelays[:], currentRouteRelays[:])
	return
}

type RouteShader struct {
	DisableNetworkNext            bool    `json:"disable_network_next"`
	SelectionPercent              int     `json:"selection_percentage"`
	ABTest                        bool    `json:"ab_test"`
	AcceptableLatency             int32   `json:"acceptable_latency"`
	LatencyReductionThreshold     int32   `json:"latency_reduction_threshold"`
	AcceptablePacketLossInstant   float32 `json:"acceptable_packet_loss_instant"`
	AcceptablePacketLossSustained float32 `json:"acceptable_packet_loss_sustained"`
	BandwidthEnvelopeUpKbps       int32   `json:"bandwidth_envelope_up_kbps"`
	BandwidthEnvelopeDownKbps     int32   `json:"bandwidth_envelope_down_kbps"`
	RouteSelectThreshold          int32   `json:"route_select_threshold"`
	RouteSwitchThreshold          int32   `json:"route_switch_threshold"`
	MaxLatencyTradeOff            int32   `json:"max_latency_trade_off"`
	RTTVeto                       int32   `json:"rtt_veto"`
	MaxNextRTT                    int32   `json:"max_next_rtt"`
	ForceNext                     bool    `json:"force_next"`
}

func NewRouteShader() RouteShader {
	return RouteShader{
		DisableNetworkNext:            false,
		SelectionPercent:              100,
		ABTest:                        false,
		AcceptableLatency:             0,
		LatencyReductionThreshold:     10,
		AcceptablePacketLossInstant:   0.1,
		AcceptablePacketLossSustained: 0.01,
		BandwidthEnvelopeUpKbps:       1024,
		BandwidthEnvelopeDownKbps:     1024,
		RouteSelectThreshold:          5,
		RouteSwitchThreshold:          10,
		MaxLatencyTradeOff:            20,
		RTTVeto:                       10,
		MaxNextRTT:                    250,
		ForceNext:                     false,
	}
}

type RouteState struct {
	Next                bool
	Veto                bool
	Disabled            bool
	NotSelected         bool
	ABTest              bool
	A                   bool
	B                   bool
	ForcedNext          bool
	ReduceLatency       bool
	ReducePacketLoss    bool
	LatencyWorse        bool
	NoRoute             bool
	NextLatencyTooHigh  bool
	RouteLost           bool
	Mispredict          bool
	LackOfDiversity     bool
	MispredictCounter   uint32
	LatencyWorseCounter uint32
	PLSustainedCounter  uint32
}

func EarlyOutDirect(userId uint64, routeShader *RouteShader, routeState *RouteState, debug *string) bool {

	// high frequency

	if routeState.Disabled {
		if debug != nil {
			*debug += "disabled\n"
		}
		return true
	}

	if routeState.Veto {
		if debug != nil {
			*debug += "veto\n"
		}
		return true
	}

	if routeState.NotSelected {
		if debug != nil {
			*debug += "not selected\n"
		}
		return true
	}

	if routeState.B {
		if debug != nil {
			*debug += "b group in ab test\n"
		}
		return true
	}

	// low frequency

	if routeShader.DisableNetworkNext {
		if debug != nil {
			*debug += "network next is disabled\n"
		}
		routeState.Disabled = true
		return true
	}

	if routeShader.SelectionPercent == 0 {
		if debug != nil {
			*debug += "selection percent is zero\n"
		}
		routeState.NotSelected = true
		return true
	}

	if (userId % 100) > uint64(routeShader.SelectionPercent) {
		if debug != nil {
			*debug += "user is not selected\n"
		}
		routeState.NotSelected = true
		return true
	}

	if routeShader.ABTest {
		routeState.ABTest = true
		if (userId % 2) == 1 {
			routeState.B = true
			if debug != nil {
				*debug += "ab test\n"
			}
			return true
		} else {
			routeState.A = true
		}
	}

	return false
}

func MakeRouteDecision_TakeNetworkNext(userId uint64, routeMatrix []RouteEntry, routeShader *RouteShader, routeState *RouteState, directLatency int32, directPacketLoss float32, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, out_routeCost *int32, out_routeNumRelays *int32, out_routeRelays []int32, debug *string, sliceNumber int32) bool {

	if EarlyOutDirect(userId, routeShader, routeState, debug) {
		if debug != nil {
			*debug += "early out direct\n"
		}
		return false
	}

	maxCost := directLatency

	reduceLatency := false
	reducePacketLoss := false

	if routeShader.ForceNext {

		// if we are forcing a network next route, set the max cost to max 32 bit integer to accept all routes

		if debug != nil {
			*debug += "forcing network next\n"
		}

		routeState.ForcedNext = true
		maxCost = math.MaxInt32

	} else {

		// otherwise, let's see if we should take network next, according to the route shader settings...

		// apply safety to source relay cost

		for i := range sourceRelayCost {
			if sourceRelayCost[i] <= 0 {
				sourceRelayCost[i] = 255
			}
		}

		// print out number of source relays that are routable + dest relays

		if debug != nil {
			numSourceRelays := len(sourceRelays)
			numRoutableSourceRelays := 0
			for i := range sourceRelays {
				if sourceRelayCost[i] != 255 {
					numRoutableSourceRelays++
				}
			}
			if sliceNumber != 0 {
				*debug += fmt.Sprintf("%d/%d source relays are routable\n", numRoutableSourceRelays, numSourceRelays)
			} else {
				*debug += "first slice. sending down client relays to ping\n"
				return false
			}
			numDestRelays := len(destRelays)
			if numDestRelays == 1 {
				*debug += fmt.Sprintf("1 dest relay\n")
			} else {
				*debug += fmt.Sprintf("%d dest relays\n", numDestRelays)
			}
		}

		// should we try to reduce latency?

		if directLatency > routeShader.AcceptableLatency {
			if debug != nil {
				*debug += fmt.Sprintf("latency is above acceptable latency %dms. try to reduce it\n", routeShader.AcceptableLatency)
			}
			maxCost = directLatency - routeShader.LatencyReductionThreshold
			reduceLatency = true
		} else {
			if debug != nil {
				*debug += fmt.Sprintf("direct latency is already acceptable. direct latency = %dms, acceptable latency = %dms\n", directLatency, routeShader.AcceptableLatency)
			}
			maxCost = -1
		}

		// should we try to reduce packet loss?

		if directPacketLoss >= routeShader.AcceptablePacketLossSustained {
			if routeState.PLSustainedCounter < 3 {
				routeState.PLSustainedCounter = routeState.PLSustainedCounter + 1
			}
		}

		if directPacketLoss < routeShader.AcceptablePacketLossSustained {
			routeState.PLSustainedCounter = 0
		}

		if directPacketLoss > routeShader.AcceptablePacketLossInstant {
			if debug != nil {
				*debug += fmt.Sprintf("packet loss is > %.2f%%. try to reduce it\n", routeShader.AcceptablePacketLossInstant)
			}
			maxCost = directLatency + routeShader.MaxLatencyTradeOff
			reducePacketLoss = true
		} else if routeState.PLSustainedCounter == 3 {
			if debug != nil {
				*debug += fmt.Sprintf("sustained packet loss > %.2f%%. try to reduce it\n", routeShader.AcceptablePacketLossSustained)
			}
			maxCost = directLatency + routeShader.MaxLatencyTradeOff
			reducePacketLoss = true
		}
	}

	// get the initial best route

	bestRouteCost := int32(0)
	bestRouteNumRelays := int32(0)
	bestRouteRelays := [constants.MaxRouteRelays]int32{}

	selectThreshold := routeShader.RouteSelectThreshold

	hasRoute := GetBestRoute_Initial(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, selectThreshold, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays, debug)

	*out_routeCost = bestRouteCost
	*out_routeNumRelays = bestRouteNumRelays
	copy(out_routeRelays, bestRouteRelays[:bestRouteNumRelays])

	// if we don't have a network next route, we can't take network next

	if !hasRoute {
		if debug != nil {
			*debug += "not taking network next. no network next route is available\n"
		}
		return false
	}

	// if the next route RTT is too high, don't take it

	if routeShader.MaxNextRTT > 0 && bestRouteCost > routeShader.MaxNextRTT {
		if debug != nil {
			*debug += fmt.Sprintf("not taking network next. best route is higher than max allowable next rtt %d\n", routeShader.MaxNextRTT)
		}
		return false
	}

	// take the network next route

	routeState.Next = true
	routeState.ReduceLatency = routeState.ReduceLatency || reduceLatency
	routeState.ReducePacketLoss = routeState.ReducePacketLoss || reducePacketLoss

	return true
}

func MakeRouteDecision_StayOnNetworkNext_Internal(userId uint64, routeMatrix []RouteEntry, relayNames []string, routeShader *RouteShader, routeState *RouteState, directLatency int32, nextLatency int32, predictedLatency int32, directPacketLoss float32, nextPacketLoss float32, currentRouteNumRelays int32, currentRouteRelays [constants.MaxRouteRelays]int32, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, out_updatedRouteCost *int32, out_updatedRouteNumRelays *int32, out_updatedRouteRelays []int32, debug *string) (bool, bool) {

	Debug("direct latency = %d", directLatency)
	Debug("next latency = %d", nextLatency)
	Debug("predicted latency = %d", predictedLatency)

	// if we early out, go direct

	if EarlyOutDirect(userId, routeShader, routeState, debug) {
		if debug != nil {
			*debug += "early out direct\n"
		}
		return false, false
	}

	// apply safety to source relay cost

	for i := range sourceRelayCost {
		if sourceRelayCost[i] <= 0 {
			sourceRelayCost[i] = 255
		}
	}

	// print out number of source relays that are routable + dest relays

	if debug != nil {
		numSourceRelays := len(sourceRelays)
		numRoutableSourceRelays := 0
		for i := range sourceRelays {
			if sourceRelayCost[i] != 255 {
				numRoutableSourceRelays++
			}
		}
		*debug += fmt.Sprintf("stay on network next: %d/%d source relays are routable\n", numRoutableSourceRelays, numSourceRelays)
		numDestRelays := len(destRelays)
		if numDestRelays == 1 {
			*debug += fmt.Sprintf("1 dest relay\n")
		} else {
			*debug += fmt.Sprintf("%d dest relays\n", numDestRelays)
		}
	}

	// if we make rtt significantly worse leave network next

	maxCost := int32(math.MaxInt32)

	if !routeShader.ForceNext {

		rttVeto := routeShader.RTTVeto

		if routeState.ReducePacketLoss {
			rttVeto += routeShader.MaxLatencyTradeOff
		}

		maxCost = directLatency - routeShader.LatencyReductionThreshold
		if routeState.ReducePacketLoss {
			maxCost += routeShader.MaxLatencyTradeOff
		}
	}

	// update the current best route

	bestRouteCost := int32(0)
	bestRouteNumRelays := int32(0)
	bestRouteRelays := [constants.MaxRouteRelays]int32{}

	routeSwitched, routeLost := GetBestRoute_Update(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, routeShader.RouteSelectThreshold, routeShader.RouteSwitchThreshold, currentRouteNumRelays, currentRouteRelays, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays, debug)

	routeState.RouteLost = routeLost

	// if we don't have a network next route, leave network next

	if bestRouteNumRelays == 0 {
		if debug != nil {
			*debug += fmt.Sprintf("leaving network next because we no longer have a suitable next route\n")
		}
		routeState.NoRoute = true
		return false, false
	}

	// stay on network next

	*out_updatedRouteCost = bestRouteCost
	*out_updatedRouteNumRelays = bestRouteNumRelays
	copy(out_updatedRouteRelays, bestRouteRelays[:bestRouteNumRelays])

	// print the network next route to debug

	if debug != nil {
		for i := 0; i < int(bestRouteNumRelays); i++ {
			if i != int(bestRouteNumRelays-1) {
				*debug += fmt.Sprintf("%s - ", relayNames[bestRouteRelays[i]])
			} else {
				*debug += fmt.Sprintf("%s\n", relayNames[bestRouteRelays[i]])
			}
		}
	}

	return true, routeSwitched
}

func MakeRouteDecision_StayOnNetworkNext(userId uint64, routeMatrix []RouteEntry, relayNames []string, routeShader *RouteShader, routeState *RouteState, directLatency int32, nextLatency int32, predictedLatency int32, directPacketLoss float32, nextPacketLoss float32, currentRouteNumRelays int32, currentRouteRelays [constants.MaxRouteRelays]int32, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, out_updatedRouteCost *int32, out_updatedRouteNumRelays *int32, out_updatedRouteRelays []int32, debug *string) (bool, bool) {

	stayOnNetworkNext, nextRouteSwitched := MakeRouteDecision_StayOnNetworkNext_Internal(userId, routeMatrix, relayNames, routeShader, routeState, directLatency, nextLatency, predictedLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCost, destRelays, out_updatedRouteCost, out_updatedRouteNumRelays, out_updatedRouteRelays, debug)

	if routeState.Next && !stayOnNetworkNext {
		routeState.Next = false
		routeState.Veto = true
	}

	return stayOnNetworkNext, nextRouteSwitched
}

// ------------------------------------------------------

func GeneratePittle(output []byte, fromAddress []byte, toAddress []byte, packetLength int) {

	var packetLengthData [2]byte
	binary.LittleEndian.PutUint16(packetLengthData[:], uint16(packetLength))

	sum := uint16(0)

	for i := 0; i < len(fromAddress); i++ {
		sum += uint16(fromAddress[i])
	}

	for i := 0; i < len(toAddress); i++ {
		sum += uint16(toAddress[i])
	}

	sum += uint16(packetLengthData[0])
	sum += uint16(packetLengthData[1])

	var sumData [2]byte
	binary.LittleEndian.PutUint16(sumData[:], sum)

	output[0] = 1 | (sumData[0] ^ sumData[1] ^ 193)
	output[1] = 1 | ((255 - output[0]) ^ 113)
}

func GenerateChonkle(output []byte, magic []byte, fromAddressData []byte, toAddressData []byte, packetLength int) {

	var packetLengthData [2]byte
	binary.LittleEndian.PutUint16(packetLengthData[:], uint16(packetLength))

	hash := fnv.New64a()
	hash.Write(magic)
	hash.Write(fromAddressData)
	hash.Write(toAddressData)
	hash.Write(packetLengthData[:])
	hashValue := hash.Sum64()

	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], uint64(hashValue))

	output[0] = ((data[6] & 0xC0) >> 6) + 42
	output[1] = (data[3] & 0x1F) + 200
	output[2] = ((data[2] & 0xFC) >> 2) + 5
	output[3] = data[0]
	output[4] = (data[2] & 0x03) + 78
	output[5] = (data[4] & 0x7F) + 96
	output[6] = ((data[1] & 0xFC) >> 2) + 100
	if (data[7] & 1) == 0 {
		output[7] = 79
	} else {
		output[7] = 7
	}
	if (data[4] & 0x80) == 0 {
		output[8] = 37
	} else {
		output[8] = 83
	}
	output[9] = (data[5] & 0x07) + 124
	output[10] = ((data[1] & 0xE0) >> 5) + 175
	output[11] = (data[6] & 0x3F) + 33
	value := (data[1] & 0x03)
	if value == 0 {
		output[12] = 97
	} else if value == 1 {
		output[12] = 5
	} else if value == 2 {
		output[12] = 43
	} else {
		output[12] = 13
	}
	output[13] = ((data[5] & 0xF8) >> 3) + 210
	output[14] = ((data[7] & 0xFE) >> 1) + 17
}

func BasicPacketFilter(data []byte, packetLength int) bool {

	if packetLength < 18 {
		return false
	}

	if data[0] < 0x32 || data[0] > 0x3C {
		return false
	}

	if data[2] != (1 | ((255 - data[1]) ^ 113)) {
		return false
	}

	if data[3] < 0x2A || data[3] > 0x2D {
		return false
	}

	if data[4] < 0xC8 || data[4] > 0xE7 {
		return false
	}

	if data[5] < 0x05 || data[5] > 0x44 {
		return false
	}

	if data[7] < 0x4E || data[7] > 0x51 {
		return false
	}

	if data[8] < 0x60 || data[8] > 0xDF {
		return false
	}

	if data[9] < 0x64 || data[9] > 0xE3 {
		return false
	}

	if data[10] != 0x07 && data[10] != 0x4F {
		return false
	}

	if data[11] != 0x25 && data[11] != 0x53 {
		return false
	}

	if data[12] < 0x7C || data[12] > 0x83 {
		return false
	}

	if data[13] < 0xAF || data[13] > 0xB6 {
		return false
	}

	if data[14] < 0x21 || data[14] > 0x60 {
		return false
	}

	if data[15] != 0x61 && data[15] != 0x05 && data[15] != 0x2B && data[15] != 0x0D {
		return false
	}

	if data[16] < 0xD2 || data[16] > 0xF1 {
		return false
	}

	if data[17] < 0x11 || data[17] > 0x90 {
		return false
	}

	return true
}

func AdvancedPacketFilter(data []byte, magic []byte, fromAddress []byte, toAddress []byte, packetLength int) bool {
	if packetLength < 18 {
		return false
	}
	var a [2]byte
	var b [15]byte
	GeneratePittle(a[:], fromAddress, toAddress, packetLength)
	GenerateChonkle(b[:], magic, fromAddress, toAddress, packetLength)
	if bytes.Compare(a[:], data[1:3]) != 0 {
		return false
	}
	if bytes.Compare(b[:], data[3:18]) != 0 {
		return false
	}
	return true
}

func GetAddressData(address *net.UDPAddr) []byte {
	return address.IP.To4()
}

func GeneratePingToken(expireTimestamp uint64, from *net.UDPAddr, to *net.UDPAddr, key []byte, output []byte) {
	data := [32 + 20]byte{}
	index := 0
	copy(data[index:], key)
	index += 32
	binary.LittleEndian.PutUint64(data[index:], expireTimestamp)
	index += 8
	copy(data[index:], from.IP.To4())
	index += 4
	copy(data[index:], to.IP.To4())
	index += 4
	binary.BigEndian.PutUint16(data[index:], uint16(from.Port))
	index += 2
	binary.BigEndian.PutUint16(data[index:], uint16(to.Port))
	index += 2
	hash := sha256.Sum256(data[:index])
	copy(output, hash[:])
}

// ------------------------------------------------------

func GetSessionScore(directRTT int32, nextRTT int32) uint32 {
	var score uint32
	if nextRTT > 0 {
		improvement := directRTT - nextRTT
		if improvement < 0 {
			improvement = 0
		}
		if improvement > 254 {
			improvement = 254
		}
		score = uint32(254 - improvement)
	} else {
		if directRTT > constants.MaxScore-255 {
			directRTT = constants.MaxScore - 255
		}
		score = 255 + uint32(constants.MaxScore-255-directRTT)
		if score < 255 {
			score = 255
		} else if score > constants.MaxScore {
			score = constants.MaxScore
		}
	}
	return score
}

func DoPagination(page int, length int) (begin, end, outputPage, numPages int) {
	begin = 0
	end = 100
	outputPage = page
	numPages = length / 100
	if length%100 != 0 {
		numPages += 1
	}
	if length > 100 {
		if page > 0 {
			begin = page * 100
			end = (page + 1) * 100
			if end > length {
				outputPage = -1
				end = length
				begin = end - 100
			}
		} else if page < 0 {
			end = length - (-page)*100
			begin = end - 100
			if begin < 0 {
				outputPage = 0
				begin = 0
				end = 100
			}
		}
	} else {
		end = length
		outputPage = 0
	}
	return
}

func DoPagination_Simple(page int, length int) (begin, end, outputPage, numPages int) {
	begin = 0
	end = 100
	outputPage = page
	numPages = length / 100
	if length%100 != 0 {
		numPages += 1
	}
	if numPages < 1 {
		numPages = 1
	}
	if page < 0 {
		page = 0
	}
	if page > numPages-1 {
		page = numPages - 1
	}
	begin = page * 100
	end = (page + 1) * 100
	if end > length {
		end = length
	}
	outputPage = page
	return
}

// ------------------------------------------------------
