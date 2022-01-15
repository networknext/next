package core

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"unsafe"
	"bytes"
	"hash/fnv"
)

const CostBias = 3
const MaxNearRelays = 32
const MaxRelaysPerRoute = 5
const MaxRoutesPerEntry = 16
const JitterThreshold = 15

const NEXT_MAX_NODES = 7
const NEXT_ADDRESS_BYTES = 19
const NEXT_ROUTE_TOKEN_BYTES = 76
const NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES = 116
const NEXT_CONTINUE_TOKEN_BYTES = 17
const NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES = 57
const NEXT_PRIVATE_KEY_BYTES = 32

var debugLogs bool

func init() {
	value, ok := os.LookupEnv("NEXT_DEBUG_LOGS")
	if ok && value == "1" {
		debugLogs = true
	}
}

func Error(s string, params ...interface{}) {
	fmt.Printf("error: "+s+"\n", params...)
}

func Debug(s string, params ...interface{}) {
	if debugLogs {
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

func SpeedOfLightTimeMilliseconds(a_lat float64, a_long float64, b_lat float64, b_long float64, c_lat float64, c_long float64) float64 {
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

func GenerateRelayKeyPair() ([]byte, []byte, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	return publicKey, privateKey, err
}

// -----------------------------------------------------

const (
	IPAddressNone = 0
	IPAddressIPv4 = 1
	IPAddressIPv6 = 2
)

func ParseAddress(input string) *net.UDPAddr {
	address := &net.UDPAddr{}
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

const (
	ADDRESS_NONE = 0
	ADDRESS_IPV4 = 1
	ADDRESS_IPV6 = 2
)

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

func ReadAddress(buffer []byte) *net.UDPAddr {
	addressType := buffer[0]
	if addressType == ADDRESS_IPV4 {
		return &net.UDPAddr{IP: net.IPv4(buffer[1], buffer[2], buffer[3], buffer[4]), Port: ((int)(binary.LittleEndian.Uint16(buffer[5:])))}
	} else if addressType == ADDRESS_IPV6 {
		return &net.UDPAddr{IP: buffer[1:17], Port: ((int)(binary.LittleEndian.Uint16(buffer[17:19])))}
	}
	return nil
}

// ---------------------------------------------------

type RouteManager struct {
	NumRoutes       int
	RouteCost       [MaxRoutesPerEntry]int32
	RouteHash       [MaxRoutesPerEntry]uint32
	RouteNumRelays  [MaxRoutesPerEntry]int32
	RouteRelays     [MaxRoutesPerEntry][MaxRelaysPerRoute]int32
	RelayDatacenter []uint64
}

func (manager *RouteManager) AddRoute(cost int32, relays ...int32) {

	// IMPORTANT: Filter out routes with loops. They can happen *very* occasionally.
	loopCheck := make(map[int32]int, len(relays))
	for i := range relays {
		if _, exists := loopCheck[relays[i]]; exists {
			return
		}
		loopCheck[relays[i]] = 1
	}

	// IMPORTANT: Filter out any route with two relays in the same datacenter. These routes are redundant.
	datacenterCheck := make(map[uint64]int, len(relays))
	for i := range relays {
		if _, exists := datacenterCheck[manager.RelayDatacenter[relays[i]]]; exists {
			return
		}
		datacenterCheck[manager.RelayDatacenter[relays[i]]] = 1
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

	} else if manager.NumRoutes < MaxRoutesPerEntry {

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

type RouteEntry struct {
	DirectCost     int32
	NumRoutes      int32
	RouteCost      [MaxRoutesPerEntry]int32
	RouteNumRelays [MaxRoutesPerEntry]int32
	RouteRelays    [MaxRoutesPerEntry][MaxRelaysPerRoute]int32
	RouteHash      [MaxRoutesPerEntry]uint32
}

func Optimize(numRelays int, numSegments int, cost []int32, costThreshold int32, relayDatacenter []uint64) []RouteEntry {

	// build a matrix of indirect routes from relays i -> j that have lower cost than direct, eg. i -> (x) -> j, where x is every other relay

	type Indirect struct {
		relay int32
		cost  int32
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
							routes[ijIndex].RouteHash[0] = RouteHash(int32(i), int32(j))

						} else {

							// no route exists from i -> j

						}

					} else {

						// subdivide routes from i -> j as follows: i -> (x) -> (y) -> (z) -> j, where the subdivision improves significantly on cost

						var routeManager RouteManager

						routeManager.RelayDatacenter = relayDatacenter

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
								routes[ijIndex].RouteHash[u] = routeManager.RouteHash[u]
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

func Optimize2(numRelays int, numSegments int, cost []int32, costThreshold int32, relayDatacenter []uint64, destinationRelay []bool) []RouteEntry {

	// build a matrix of indirect routes from relays i -> j that have lower cost than direct, eg. i -> (x) -> j, where x is every other relay

	type Indirect struct {
		relay int32
		cost  int32
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

					if !destinationRelay[i] && !destinationRelay[j] {
						continue
					}

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
							routes[ijIndex].RouteHash[0] = RouteHash(int32(i), int32(j))

						} else {

							// no route exists from i -> j

						}

					} else {

						// subdivide routes from i -> j as follows: i -> (x) -> (y) -> (z) -> j, where the subdivision improves significantly on cost

						var routeManager RouteManager

						routeManager.RelayDatacenter = relayDatacenter

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
								routes[ijIndex].RouteHash[u] = routeManager.RouteHash[u]
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

// ---------------------------------------------------

type RouteToken struct {
	ExpireTimestamp uint64
	SessionId       uint64
	SessionVersion  uint8
	KbpsUp          uint32
	KbpsDown        uint32
	NextAddress     *net.UDPAddr
	PrivateKey      [NEXT_PRIVATE_KEY_BYTES]byte
}

type ContinueToken struct {
	ExpireTimestamp uint64
	SessionId       uint64
	SessionVersion  uint8
}

const Crypto_kx_PUBLICKEYBYTES = C.crypto_kx_PUBLICKEYBYTES
const Crypto_box_PUBLICKEYBYTES = C.crypto_box_PUBLICKEYBYTES

const KeyBytes = 32
const NonceBytes = 24
const MacBytes = C.crypto_box_MACBYTES
const SignatureBytes = C.crypto_sign_BYTES
const PublicKeyBytes = C.crypto_sign_PUBLICKEYBYTES

func Encrypt(senderPrivateKey []byte, receiverPublicKey []byte, nonce []byte, buffer []byte, bytes int) int {
	C.crypto_box_easy((*C.uchar)(&buffer[0]),
		(*C.uchar)(&buffer[0]),
		C.ulonglong(bytes),
		(*C.uchar)(&nonce[0]),
		(*C.uchar)(&receiverPublicKey[0]),
		(*C.uchar)(&senderPrivateKey[0]))
	return bytes + C.crypto_box_MACBYTES
}

func Decrypt(senderPublicKey []byte, receiverPrivateKey []byte, nonce []byte, buffer []byte, bytes int) error {
	result := C.crypto_box_open_easy(
		(*C.uchar)(&buffer[0]),
		(*C.uchar)(&buffer[0]),
		C.ulonglong(bytes),
		(*C.uchar)(&nonce[0]),
		(*C.uchar)(&senderPublicKey[0]),
		(*C.uchar)(&receiverPrivateKey[0]))
	if result != 0 {
		return fmt.Errorf("failed to decrypt: result = %d", result)
	} else {
		return nil
	}
}

func RandomBytes(buffer []byte) {
	C.randombytes_buf(unsafe.Pointer(&buffer[0]), C.size_t(len(buffer)))
}

// -----------------------------------------------------------------------------

func WriteRouteToken(token *RouteToken, buffer []byte) {
	binary.LittleEndian.PutUint64(buffer[0:], token.ExpireTimestamp)
	binary.LittleEndian.PutUint64(buffer[8:], token.SessionId)
	buffer[8+8] = token.SessionVersion
	binary.LittleEndian.PutUint32(buffer[8+8+1:], token.KbpsUp)
	binary.LittleEndian.PutUint32(buffer[8+8+1+4:], token.KbpsDown)
	WriteAddress(buffer[8+8+1+4+4:], token.NextAddress)
	copy(buffer[8+8+1+4+4+NEXT_ADDRESS_BYTES:], token.PrivateKey[:])
}

func ReadRouteToken(token *RouteToken, buffer []byte) error {
	if len(buffer) < NEXT_ROUTE_TOKEN_BYTES {
		return fmt.Errorf("buffer too small to read route token")
	}
	token.ExpireTimestamp = binary.LittleEndian.Uint64(buffer[0:])
	token.SessionId = binary.LittleEndian.Uint64(buffer[8:])
	token.SessionVersion = buffer[8+8]
	token.KbpsUp = binary.LittleEndian.Uint32(buffer[8+8+1:])
	token.KbpsDown = binary.LittleEndian.Uint32(buffer[8+8+1+4:])
	token.NextAddress = ReadAddress(buffer[8+8+1+4+4:])
	copy(token.PrivateKey[:], buffer[8+8+1+4+4+NEXT_ADDRESS_BYTES:])
	return nil
}

func WriteEncryptedRouteToken(token *RouteToken, tokenData []byte, senderPrivateKey []byte, receiverPublicKey []byte) {
	RandomBytes(tokenData[:NonceBytes])
	WriteRouteToken(token, tokenData[NonceBytes:])
	Encrypt(senderPrivateKey, receiverPublicKey, tokenData[0:NonceBytes], tokenData[NonceBytes:], NEXT_ROUTE_TOKEN_BYTES)
}

func ReadEncryptedRouteToken(token *RouteToken, tokenData []byte, senderPublicKey []byte, receiverPrivateKey []byte) error {
	if len(tokenData) < NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES {
		return fmt.Errorf("not enough bytes for encrypted route token")
	}
	nonce := tokenData[0 : C.crypto_box_NONCEBYTES-1]
	tokenData = tokenData[C.crypto_box_NONCEBYTES:]
	if err := Decrypt(senderPublicKey, receiverPrivateKey, nonce, tokenData, NEXT_ROUTE_TOKEN_BYTES+C.crypto_box_MACBYTES); err != nil {
		return err
	}
	return ReadRouteToken(token, tokenData)
}

func WriteRouteTokens(tokenData []byte, expireTimestamp uint64, sessionId uint64, sessionVersion uint8, kbpsUp uint32, kbpsDown uint32, numNodes int, addresses []*net.UDPAddr, publicKeys [][]byte, masterPrivateKey [KeyBytes]byte) {
	privateKey := [KeyBytes]byte{}
	RandomBytes(privateKey[:])
	for i := 0; i < numNodes; i++ {
		var token RouteToken
		token.ExpireTimestamp = expireTimestamp
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.KbpsUp = kbpsUp
		token.KbpsDown = kbpsDown
		if i != numNodes-1 {
			token.NextAddress = addresses[i+1]
		}
		copy(token.PrivateKey[:], privateKey[:])
		WriteEncryptedRouteToken(&token, tokenData[i*NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], masterPrivateKey[:], publicKeys[i])
	}
}

// -----------------------------------------------------------------------------

func WriteContinueToken(token *ContinueToken, buffer []byte) {
	binary.LittleEndian.PutUint64(buffer[0:], token.ExpireTimestamp)
	binary.LittleEndian.PutUint64(buffer[8:], token.SessionId)
	buffer[8+8] = token.SessionVersion
}

func ReadContinueToken(token *ContinueToken, buffer []byte) error {
	if len(buffer) < NEXT_CONTINUE_TOKEN_BYTES {
		return fmt.Errorf("buffer too small to read continue token")
	}
	token.ExpireTimestamp = binary.LittleEndian.Uint64(buffer[0:])
	token.SessionId = binary.LittleEndian.Uint64(buffer[8:])
	token.SessionVersion = buffer[8+8]
	return nil
}

func WriteEncryptedContinueToken(token *ContinueToken, buffer []byte, senderPrivateKey []byte, receiverPublicKey []byte) {
	RandomBytes(buffer[:NonceBytes])
	WriteContinueToken(token, buffer[NonceBytes:])
	Encrypt(senderPrivateKey, receiverPublicKey, buffer[:NonceBytes], buffer[NonceBytes:], NEXT_CONTINUE_TOKEN_BYTES)
}

func ReadEncryptedContinueToken(token *ContinueToken, tokenData []byte, senderPublicKey []byte, receiverPrivateKey []byte) error {
	if len(tokenData) < NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES {
		return fmt.Errorf("not enough bytes for encrypted continue token")
	}
	nonce := tokenData[0 : C.crypto_box_NONCEBYTES-1]
	tokenData = tokenData[C.crypto_box_NONCEBYTES:]
	if err := Decrypt(senderPublicKey, receiverPrivateKey, nonce, tokenData, NEXT_CONTINUE_TOKEN_BYTES+C.crypto_box_MACBYTES); err != nil {
		return err
	}
	return ReadContinueToken(token, tokenData)
}

func WriteContinueTokens(tokenData []byte, expireTimestamp uint64, sessionId uint64, sessionVersion uint8, numNodes int, publicKeys [][]byte, masterPrivateKey [KeyBytes]byte) {
	for i := 0; i < numNodes; i++ {
		var token ContinueToken
		token.ExpireTimestamp = expireTimestamp
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		WriteEncryptedContinueToken(&token, tokenData[i*NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES:], masterPrivateKey[:], publicKeys[i])
	}
}

// -----------------------------------------------------------------------------

func GetBestRouteCost(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32) int32 {
	bestRouteCost := int32(math.MaxInt32)
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

			routeRelayLoop:
				for k := int32(0); k < entry.NumRoutes; k++ {
					for l := 0; l < len(entry.RouteRelays[0]); l++ {

						// Do not consider routes with full relays
						if _, isRelayFull := fullRelaySet[entry.RouteRelays[k][l]]; isRelayFull {
							continue routeRelayLoop
						}
					}

					cost := sourceRelayCost[i] + entry.RouteCost[k]
					if cost < bestRouteCost {
						bestRouteCost = cost
					}
				}
			}
		}
	}
	if bestRouteCost == int32(math.MaxInt32) {
		return bestRouteCost
	}

	return bestRouteCost + CostBias
}

func ReverseRoute(route []int32) {
	for i, j := 0, len(route)-1; i < j; i, j = i+1, j-1 {
		route[i], route[j] = route[j], route[i]
	}
}

func RouteExists(routeMatrix []RouteEntry, routeNumRelays int32, routeRelays [MaxRelaysPerRoute]int32, debug *string) bool {
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

func GetCurrentRouteCost(routeMatrix []RouteEntry, routeNumRelays int32, routeRelays [MaxRelaysPerRoute]int32, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, debug *string) int32 {

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
		return sourceCost + entry.RouteCost[i] + CostBias
	}

	// We didn't find the route :(
	if debug != nil {
		*debug += "could not find route\n"
	}
	return -1
}

type BestRoute struct {
	Cost          int32
	NumRelays     int32
	Relays        [MaxRelaysPerRoute]int32
	NeedToReverse bool
}

func GetBestRoutes(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, bestRoutes []BestRoute, numBestRoutes *int, routeDiversity *int32) {
	numRoutes := 0
	maxRoutes := len(bestRoutes)
	for i := range sourceRelays {
		// IMPORTANT: RTT = 255 signals the source relay is unroutable
		if sourceRelayCost[i] >= 255 {
			continue
		}
		firstRouteFromThisRelay := true
		for j := range destRelays {
			sourceRelayIndex := sourceRelays[i]
			destRelayIndex := destRelays[j]
			if sourceRelayIndex == destRelayIndex {
				continue
			}

			index := TriMatrixIndex(int(sourceRelayIndex), int(destRelayIndex))
			entry := &routeMatrix[index]

		routeEntryLoop:
			for k := 0; k < int(entry.NumRoutes); k++ {
				cost := entry.RouteCost[k] + sourceRelayCost[i]
				if cost > maxCost {
					break
				}
				bestRoutes[numRoutes].Cost = cost
				bestRoutes[numRoutes].NumRelays = entry.RouteNumRelays[k]

				for l := 0; l < len(entry.RouteRelays[0]); l++ {

					// Skip over any relays that are considered full
					if _, isRelayFull := fullRelaySet[entry.RouteRelays[k][l]]; isRelayFull {
						continue routeEntryLoop
					}

					bestRoutes[numRoutes].Relays[l] = entry.RouteRelays[k][l]
				}
				bestRoutes[numRoutes].NeedToReverse = sourceRelayIndex < destRelayIndex
				numRoutes++
				if firstRouteFromThisRelay {
					*routeDiversity++
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

func ReframeRoute(routeState *RouteState, relayIDToIndex map[uint64]int32, routeRelayIds []uint64, out_routeRelays *[MaxRelaysPerRoute]int32) bool {
	for i := range routeRelayIds {
		relayIndex, ok := relayIDToIndex[routeRelayIds[i]]
		if !ok {
			routeState.RelayWentAway = true
			return false
		}
		out_routeRelays[i] = relayIndex
	}
	routeState.RelayWentAway = false
	return true
}

func ReframeRelays(routeShader *RouteShader, routeState *RouteState, relayIDToIndex map[uint64]int32, directLatency int32, directJitter int32, directPacketLoss int32, nextPacketLoss int32, sliceNumber int32, sourceRelayId []uint64, sourceRelayLatency []int32, sourceRelayJitter []int32, sourceRelayPacketLoss []int32, destRelayIds []uint64, out_sourceRelayLatency []int32, out_sourceRelayJitter []int32, out_numDestRelays *int32, out_destRelays []int32) {

	if routeState.NumNearRelays == 0 {
		routeState.NumNearRelays = int32(len(sourceRelayId))
	}

	if directJitter > 255 {
		directJitter = 255
	}

	if directJitter > routeState.DirectJitter {
		routeState.DirectJitter = directJitter
	}

	for i := range sourceRelayLatency {

		// you say your latency is 0ms? I don't believe you!
		if sourceRelayLatency[i] <= 0 {
			routeState.NearRelayRTT[i] = 255
			out_sourceRelayLatency[i] = 255
			continue
		}

		// any source relay with >= 50% PL in the last slice is bad news
		if sourceRelayPacketLoss[i] >= 50 {
			routeState.NearRelayRTT[i] = 255
			out_sourceRelayLatency[i] = 255
			continue
		}

		// any source relay with latency > direct is not helpful to us
		if routeState.NearRelayRTT[i] != 255 && routeState.NearRelayRTT[i] > directLatency+10 {
			routeState.NearRelayRTT[i] = 255
			out_sourceRelayLatency[i] = 255
			continue
		}

		// any source relay that no longer exists cannot be routed through
		_, ok := relayIDToIndex[sourceRelayId[i]]
		if !ok {
			routeState.NearRelayRTT[i] = 255
			out_sourceRelayLatency[i] = 255
			continue
		}

		rtt := sourceRelayLatency[i]
		jitter := sourceRelayJitter[i]

		if rtt > 255 {
			rtt = 255
		}

		if jitter > 255 {
			jitter = 255
		}

		if rtt > routeState.NearRelayRTT[i] {
			routeState.NearRelayRTT[i] = rtt
		}

		if jitter > routeState.NearRelayJitter[i] {
			routeState.NearRelayJitter[i] = jitter
		}

		out_sourceRelayLatency[i] = routeState.NearRelayRTT[i]
		out_sourceRelayJitter[i] = routeState.NearRelayJitter[i]
	}

	// exclude near relays with higher number of packet loss events than direct (sporadic packet loss)

	if directPacketLoss > 0 {
		routeState.DirectPLCount++
	}

	// IMPORTANT: Only run for nonexistent or sporadic direct PL
	if int32(routeState.DirectPLCount*10) <= sliceNumber {

		for i := range sourceRelayPacketLoss {

			if sourceRelayPacketLoss[i] > 0 {
				routeState.NearRelayPLCount[i]++
			}

			if routeState.NearRelayPLCount[i] > routeState.DirectPLCount {
				out_sourceRelayLatency[i] = 255
			}
		}
	}

	// exclude near relays with a history of packet loss values worse than direct (continuous packet loss)

	routeState.PLHistorySamples++
	if routeState.PLHistorySamples > 8 {
		routeState.PLHistorySamples = 8
	}

	index := routeState.PLHistoryIndex

	samples := routeState.PLHistorySamples

	temp_threshold := samples / 2

	if directPacketLoss > 0 {
		routeState.DirectPLHistory |= (1 << index)
	} else {
		routeState.DirectPLHistory &= ^(1 << index)
	}

	for i := range sourceRelayPacketLoss {

		if sourceRelayPacketLoss[i] > directPacketLoss {
			routeState.NearRelayPLHistory[i] |= (1 << index)
		} else {
			routeState.NearRelayPLHistory[i] &= ^(1 << index)
		}

		plCount := int32(0)
		for j := 0; j < int(samples); j++ {
			if (routeState.NearRelayPLHistory[i] & (1 << j)) != 0 {
				plCount++
			}
		}

		if plCount > temp_threshold {
			out_sourceRelayLatency[i] = 255
		}
	}

	routeState.PLHistoryIndex = (routeState.PLHistoryIndex + 1) % 8

	// exclude near relays with (significantly) higher jitter than direct

	for i := range sourceRelayLatency {

		if routeState.NearRelayJitter[i] > routeState.DirectJitter+JitterThreshold {
			out_sourceRelayLatency[i] = 255
		}
	}

	// exclude near relays with (significantly) higher than average jitter

	count := 0
	totalJitter := 0.0
	for i := range sourceRelayLatency {
		if out_sourceRelayLatency[i] != 255 {
			totalJitter += float64(out_sourceRelayJitter[i])
			count++
		}
	}

	if count > 0 {
		averageJitter := int32(math.Ceil(totalJitter / float64(count)))
		for i := range sourceRelayLatency {
			if out_sourceRelayLatency[i] == 255 {
				continue
			}
			if out_sourceRelayJitter[i] > averageJitter+JitterThreshold {
				out_sourceRelayLatency[i] = 255
			}
		}
	}

	// extra safety. don't let any relay report latency of zero

	for i := range sourceRelayLatency {

		if sourceRelayLatency[i] <= 0 || out_sourceRelayLatency[i] <= 0 {
			routeState.NearRelayRTT[i] = 255
			out_sourceRelayLatency[i] = 255
			continue
		}
	}

	// exclude any dest relays that no longer exist in the route matrix

	numDestRelays := int32(0)

	for i := range destRelayIds {
		destRelayIndex, ok := relayIDToIndex[destRelayIds[i]]
		if !ok {
			continue
		}
		out_destRelays[numDestRelays] = destRelayIndex
		numDestRelays++
	}

	*out_numDestRelays = numDestRelays
}

// ----------------------------------------------

func GetRandomBestRoute(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, threshold int32, out_bestRouteCost *int32, out_bestRouteNumRelays *int32, out_bestRouteRelays *[MaxRelaysPerRoute]int32, debug *string) (foundRoute bool, routeDiversity int32) {

	foundRoute = false
	routeDiversity = 0

	if maxCost == -1 {
		return
	}

	bestRouteCost := GetBestRouteCost(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCost, destRelays)
	if debug != nil {
		*debug += fmt.Sprintf("best route cost is %d\n", bestRouteCost)
	}

	if bestRouteCost > maxCost {
		if debug != nil {
			*debug += fmt.Sprintf("could not find any next route <= max cost %d\n", maxCost)
		}
		*out_bestRouteCost = bestRouteCost
		return
	}

	numBestRoutes := 0
	bestRoutes := make([]BestRoute, 1024)
	GetBestRoutes(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCost, destRelays, bestRouteCost+threshold, bestRoutes, &numBestRoutes, &routeDiversity)
	if numBestRoutes == 0 {
		if debug != nil {
			*debug += "could not find any next routes\n"
		}
		return
	}

	if debug != nil {
		numNearRelays := 0
		for i := range sourceRelays {
			if sourceRelayCost[i] != 255 {
				numNearRelays++
			}
		}
		*debug += fmt.Sprintf("found %d suitable routes in [%d,%d] from %d/%d near relays\n", numBestRoutes, bestRouteCost, bestRouteCost+threshold, numNearRelays, len(sourceRelays))
	}

	randomIndex := rand.Intn(numBestRoutes)

	*out_bestRouteCost = bestRoutes[randomIndex].Cost + CostBias
	*out_bestRouteNumRelays = bestRoutes[randomIndex].NumRelays

	if !bestRoutes[randomIndex].NeedToReverse {
		copy(out_bestRouteRelays[:], bestRoutes[randomIndex].Relays[:bestRoutes[randomIndex].NumRelays])
	} else {
		numRouteRelays := bestRoutes[randomIndex].NumRelays
		for i := int32(0); i < numRouteRelays; i++ {
			out_bestRouteRelays[numRouteRelays-1-i] = bestRoutes[randomIndex].Relays[i]
		}
	}

	foundRoute = true

	return
}

func GetBestRoute_Initial(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, selectThreshold int32, out_bestRouteCost *int32, out_bestRouteNumRelays *int32, out_bestRouteRelays *[MaxRelaysPerRoute]int32, debug *string) (hasRoute bool, routeDiversity int32) {

	return GetRandomBestRoute(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCost, destRelays, maxCost, selectThreshold, out_bestRouteCost, out_bestRouteNumRelays, out_bestRouteRelays, debug)
}

func GetBestRoute_Update(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, selectThreshold int32, switchThreshold int32, currentRouteNumRelays int32, currentRouteRelays [MaxRelaysPerRoute]int32, out_updatedRouteCost *int32, out_updatedRouteNumRelays *int32, out_updatedRouteRelays *[MaxRelaysPerRoute]int32, debug *string) (routeChanged bool, routeLost bool) {

	// if the current route no longer exists, pick a new route

	currentRouteCost := GetCurrentRouteCost(routeMatrix, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCost, destRelays, debug)

	if currentRouteCost < 0 {
		if debug != nil {
			*debug += "current route no longer exists. picking a new random route\n"
		}
		GetRandomBestRoute(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCost, destRelays, maxCost, selectThreshold, out_updatedRouteCost, out_updatedRouteNumRelays, out_updatedRouteRelays, debug)
		routeChanged = true
		routeLost = true
		return
	}

	// if the current route is no longer within threshold of the best route, pick a new the route

	bestRouteCost := GetBestRouteCost(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCost, destRelays)

	if int64(currentRouteCost) > int64(bestRouteCost)+int64(switchThreshold) {
		if debug != nil {
			*debug += fmt.Sprintf("current route no longer within switch threshold of best route. picking a new random route.\ncurrent route cost = %d, best route cost = %d, route switch threshold = %d\n", currentRouteCost, bestRouteCost, switchThreshold)
		}
		GetRandomBestRoute(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCost, destRelays, bestRouteCost, selectThreshold, out_updatedRouteCost, out_updatedRouteNumRelays, out_updatedRouteRelays, debug)
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
	DisableNetworkNext        bool
	SelectionPercent          int
	ABTest                    bool
	ProMode                   bool
	ReduceLatency             bool
	ReduceJitter              bool
	ReducePacketLoss          bool
	Multipath                 bool
	AcceptableLatency         int32
	LatencyThreshold          int32
	AcceptablePacketLoss      float32
	BandwidthEnvelopeUpKbps   int32
	BandwidthEnvelopeDownKbps int32
	BannedUsers               map[uint64]bool
	PacketLossSustained       float32
}

func NewRouteShader() RouteShader {
	return RouteShader{
		DisableNetworkNext:        false,
		SelectionPercent:          100,
		ABTest:                    false,
		ReduceLatency:             true,
		ReduceJitter:              true,
		ReducePacketLoss:          true,
		Multipath:                 false,
		ProMode:                   false,
		AcceptableLatency:         0,
		LatencyThreshold:          10,
		AcceptablePacketLoss:      1.0,
		BandwidthEnvelopeUpKbps:   1024,
		BandwidthEnvelopeDownKbps: 1024,
		BannedUsers:               make(map[uint64]bool),
		PacketLossSustained:       100,
	}
}

type RouteState struct {
	UserID              uint64
	Next                bool
	Veto                bool
	Banned              bool
	Disabled            bool
	NotSelected         bool
	ABTest              bool
	A                   bool
	B                   bool
	ForcedNext          bool
	ReduceLatency       bool
	ReducePacketLoss    bool
	ProMode             bool
	Multipath           bool
	Committed           bool
	CommitVeto          bool
	CommitCounter       int32
	LatencyWorse        bool
	LocationVeto        bool
	MultipathOverload   bool
	NoRoute             bool
	NextLatencyTooHigh  bool
	NumNearRelays       int32
	NearRelayRTT        [MaxNearRelays]int32
	NearRelayJitter     [MaxNearRelays]int32
	NearRelayPLHistory  [MaxNearRelays]uint32
	NearRelayPLCount    [MaxNearRelays]uint32
	DirectPLHistory     uint32
	DirectPLCount       uint32
	PLHistoryIndex      int32
	PLHistorySamples    int32
	RelayWentAway       bool
	RouteLost           bool
	DirectJitter        int32
	Mispredict          bool
	LackOfDiversity     bool
	MispredictCounter   uint32
	LatencyWorseCounter uint32
	MultipathRestricted bool
	PLSustainedCounter  int32
}

type InternalConfig struct {
	RouteSelectThreshold           int32
	RouteSwitchThreshold           int32
	MaxLatencyTradeOff             int32
	RTTVeto_Default                int32
	RTTVeto_Multipath              int32
	RTTVeto_PacketLoss             int32
	MultipathOverloadThreshold     int32
	TryBeforeYouBuy                bool
	ForceNext                      bool
	LargeCustomer                  bool
	Uncommitted                    bool
	MaxRTT                         int32
	HighFrequencyPings             bool
	RouteDiversity                 int32
	MultipathThreshold             int32
	EnableVanityMetrics            bool
	ReducePacketLossMinSliceNumber int32
}

func NewInternalConfig() InternalConfig {
	return InternalConfig{
		RouteSelectThreshold:           2,
		RouteSwitchThreshold:           5,
		MaxLatencyTradeOff:             20,
		RTTVeto_Default:                -10,
		RTTVeto_Multipath:              -20,
		RTTVeto_PacketLoss:             -30,
		MultipathOverloadThreshold:     500,
		TryBeforeYouBuy:                false,
		ForceNext:                      false,
		LargeCustomer:                  false,
		Uncommitted:                    false,
		MaxRTT:                         300,
		HighFrequencyPings:             true,
		RouteDiversity:                 0,
		MultipathThreshold:             25,
		EnableVanityMetrics:            false,
		ReducePacketLossMinSliceNumber: 0,
	}
}

func EarlyOutDirect(routeShader *RouteShader, routeState *RouteState) bool {

	if routeState.Veto || routeState.LocationVeto || routeState.Banned || routeState.Disabled || routeState.NotSelected || routeState.B {
		return true
	}

	if routeShader.DisableNetworkNext {
		routeState.Disabled = true
		return true
	}

	if routeShader.SelectionPercent == 0 || (routeState.UserID%100) > uint64(routeShader.SelectionPercent) {
		routeState.NotSelected = true
		return true
	}

	if routeShader.ABTest {
		routeState.ABTest = true
		if (routeState.UserID % 2) == 1 {
			routeState.B = true
			return true
		} else {
			routeState.A = true
		}
	}

	if routeShader.BannedUsers[routeState.UserID] {
		routeState.Banned = true
		return true
	}

	return false
}

func TryBeforeYouBuy(routeState *RouteState, internal *InternalConfig, directLatency int32, nextLatency int32, directPacketLoss float32, nextPacketLoss float32) bool {

	// don't do anything unless try before you buy is enabled

	if !internal.TryBeforeYouBuy {
		return true
	}

	// don't do anything if we have already committed

	if routeState.Committed {
		return true
	}

	// veto the route if we don't see improvement after three slices

	routeState.CommitCounter++
	if routeState.CommitCounter > 3 {
		routeState.CommitVeto = true
		return false
	}

	// if we are reducing packet loss. commit if RTT is within tolerance and packet loss is not worse

	if routeState.ReducePacketLoss {
		if nextLatency <= directLatency-internal.RTTVeto_PacketLoss && nextPacketLoss <= directPacketLoss {
			routeState.Committed = true
		}
		return true
	}

	// we are reducing latency. commit if latency and packet loss are not worse.

	if nextLatency <= directLatency && nextPacketLoss <= directPacketLoss {
		routeState.Committed = true
		return true
	}

	return true
}

func MakeRouteDecision_TakeNetworkNext(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, routeShader *RouteShader, routeState *RouteState, multipathVetoUsers map[uint64]bool, internal *InternalConfig, directLatency int32, directPacketLoss float32, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, out_routeCost *int32, out_routeNumRelays *int32, out_routeRelays []int32, out_routeDiversity *int32, debug *string, sliceNumber int32) bool {

	if EarlyOutDirect(routeShader, routeState) {
		return false
	}

	maxCost := directLatency

	// apply safety to source relay cost

	for i := range sourceRelayCost {
		if sourceRelayCost[i] <= 0 {
			sourceRelayCost[i] = 255
		}
	}

	// should we try to reduce latency?

	reduceLatency := false
	if routeShader.ReduceLatency {
		if directLatency > routeShader.AcceptableLatency {
			if debug != nil {
				*debug += "try to reduce latency\n"
			}
			maxCost = directLatency - (routeShader.LatencyThreshold + internal.RouteSelectThreshold)
			reduceLatency = true
		} else {
			if debug != nil {
				*debug += fmt.Sprintf("direct latency is already acceptable. direct latency = %d, latency threshold = %d\n", directLatency, routeShader.LatencyThreshold)
			}
			maxCost = -1
		}
	}

	// should we try to reduce packet loss?

	// Check if the session is seeing sustained packet loss and increment/reset the counter

	if directPacketLoss >= routeShader.PacketLossSustained {
		if routeState.PLSustainedCounter < 3 {
			routeState.PLSustainedCounter = routeState.PLSustainedCounter + 1
		}
	}

	if directPacketLoss < routeShader.PacketLossSustained {
		routeState.PLSustainedCounter = 0
	}

	reducePacketLoss := false
	if routeShader.ReducePacketLoss && ((directPacketLoss > routeShader.AcceptablePacketLoss && sliceNumber >= internal.ReducePacketLossMinSliceNumber) || routeState.PLSustainedCounter == 3) {
		if debug != nil {
			*debug += "try to reduce packet loss\n"
		}
		maxCost = directLatency + internal.MaxLatencyTradeOff - internal.RouteSelectThreshold
		reducePacketLoss = true
	}

	// should we enable pro mode?

	routeState.MultipathRestricted = multipathVetoUsers[routeState.UserID]

	proMode := false
	if routeShader.ProMode && !routeState.MultipathRestricted {
		if debug != nil {
			*debug += "pro mode\n"
		}
		maxCost = directLatency + internal.MaxLatencyTradeOff - internal.RouteSelectThreshold
		proMode = true
		reduceLatency = false
		reducePacketLoss = false
	}

	// if we are forcing a network next route, set the max cost to max 32 bit integer to accept all routes

	if internal.ForceNext {
		if debug != nil {
			*debug += "forcing network next\n"
		}
		maxCost = math.MaxInt32
		routeState.ForcedNext = true
	}

	// get the initial best route

	bestRouteCost := int32(0)
	bestRouteNumRelays := int32(0)
	bestRouteRelays := [MaxRelaysPerRoute]int32{}

	selectThreshold := internal.RouteSelectThreshold

	hasRoute, routeDiversity := GetBestRoute_Initial(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCost, destRelays, maxCost, selectThreshold, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays, debug)

	*out_routeCost = bestRouteCost
	*out_routeNumRelays = bestRouteNumRelays
	*out_routeDiversity = routeDiversity
	copy(out_routeRelays, bestRouteRelays[:bestRouteNumRelays])

	if debug != nil && hasRoute {
		*debug += fmt.Sprintf("route diversity %d\n", routeDiversity)
	}

	// if we don't have enough route diversity, we can't take network next

	if routeDiversity < internal.RouteDiversity {
		if debug != nil {
			*debug += fmt.Sprintf("not enough route diversity. %d < %d\n", routeDiversity, internal.RouteDiversity)
		}
		routeState.LackOfDiversity = true
		return false
	}

	// if we don't have a network next route, we can't take network next

	if !hasRoute {
		if debug != nil {
			*debug += "not taking network next. no next route available within parameters\n"
		}
		return false
	}

	// if the next route RTT is too high, don't take it

	if bestRouteCost > internal.MaxRTT {
		if debug != nil {
			*debug += fmt.Sprintf("not taking network next. best route is higher than max rtt %d\n", internal.MaxRTT)
		}
		return false
	}

	// don't multipath if we are reducing latency more than the multipath threshold

	multipath := (proMode || routeShader.Multipath) && !routeState.MultipathRestricted

	if internal.MultipathThreshold > 0 {
		difference := directLatency - bestRouteCost
		if difference > internal.MultipathThreshold {
			multipath = false
		}
	}

	// take the network next route

	routeState.Next = true
	routeState.ReduceLatency = reduceLatency
	routeState.ReducePacketLoss = reducePacketLoss
	routeState.ProMode = proMode
	routeState.Multipath = multipath

	// should we commit to sending packets across network next?

	routeState.Committed = !internal.Uncommitted && (!internal.TryBeforeYouBuy || routeState.Multipath)

	return true
}

func MakeRouteDecision_StayOnNetworkNext_Internal(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, relayNames []string, routeShader *RouteShader, routeState *RouteState, internal *InternalConfig, directLatency int32, nextLatency int32, predictedLatency int32, directPacketLoss float32, nextPacketLoss float32, currentRouteNumRelays int32, currentRouteRelays [MaxRelaysPerRoute]int32, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, out_updatedRouteCost *int32, out_updatedRouteNumRelays *int32, out_updatedRouteRelays []int32, debug *string) (bool, bool) {

	// if we early out, go direct

	if EarlyOutDirect(routeShader, routeState) {
		return false, false
	}

	// apply safety to source relay cost

	for i := range sourceRelayCost {
		if sourceRelayCost[i] <= 0 {
			sourceRelayCost[i] = 255
		}
	}

	// if we mispredict RTT by 10ms or more, 3 slices in a row, leave network next

	if predictedLatency > 0 && nextLatency >= predictedLatency+10 {
		routeState.MispredictCounter++
		if routeState.MispredictCounter == 3 {
			if debug != nil {
				*debug += fmt.Sprintf("mispredict: next rtt = %d, predicted rtt = %d\n", nextLatency, predictedLatency)
			}
			routeState.Mispredict = true
			return false, false
		}
	} else {
		routeState.MispredictCounter = 0
	}

	// if we overload the connection in multipath, leave network next

	if routeState.Multipath && directLatency >= internal.MultipathOverloadThreshold {
		if debug != nil {
			*debug += fmt.Sprintf("multipath overload: direct rtt = %d > threshold %d\n", directLatency, internal.MultipathOverloadThreshold)
		}
		routeState.MultipathOverload = true
		return false, false
	}

	// if we make rtt significantly worse leave network next

	maxCost := int32(math.MaxInt32)

	if !internal.ForceNext {

		rttVeto := internal.RTTVeto_Default

		if routeState.ReducePacketLoss {
			rttVeto = internal.RTTVeto_PacketLoss
		}

		if routeState.Multipath {
			rttVeto = internal.RTTVeto_Multipath
		}

		// IMPORTANT: Here is where we abort the network next route if we see that we have
		// made latency worse on the previous slice. This is disabled while we are not committed,
		// so we can properly evaluate the route in try before you buy instead of vetoing it right away

		if routeState.Committed {

			if !routeState.Multipath {

				// If we make latency worse and we are not in multipath, leave network next right away

				if nextLatency > (directLatency - rttVeto) {
					if debug != nil {
						*debug += fmt.Sprintf("aborting route because we made latency worse: next rtt = %d, direct rtt = %d, veto rtt = %d\n", nextLatency, directLatency, directLatency-rttVeto)
					}
					routeState.LatencyWorse = true
					return false, false
				}

			} else {

				// If we are in multipath, only leave network next if we make latency worse three slices in a row

				if nextLatency > (directLatency - rttVeto) {
					routeState.LatencyWorseCounter++
					if routeState.LatencyWorseCounter == 3 {
						if debug != nil {
							*debug += fmt.Sprintf("aborting route because we made latency worse 3X: next rtt = %d, direct rtt = %d, veto rtt = %d\n", nextLatency, directLatency, directLatency-rttVeto)
						}
						routeState.LatencyWorse = true
						return false, false
					}
				} else {
					routeState.LatencyWorseCounter = 0
				}

			}
		}

		maxCost = directLatency - rttVeto
	}

	// update the current best route

	bestRouteCost := int32(0)
	bestRouteNumRelays := int32(0)
	bestRouteRelays := [MaxRelaysPerRoute]int32{}

	routeSwitched, routeLost := GetBestRoute_Update(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCost, destRelays, maxCost, internal.RouteSelectThreshold, internal.RouteSwitchThreshold, currentRouteNumRelays, currentRouteRelays, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays, debug)

	routeState.RouteLost = routeLost

	// if we don't have a network next route, leave network next

	if bestRouteNumRelays == 0 {
		if debug != nil {
			*debug += fmt.Sprintf("leaving network next because we no longer have a suitable next route\n")
		}
		routeState.NoRoute = true
		return false, false
	}

	// if the next route RTT is too high, leave network next

	if bestRouteCost > internal.MaxRTT {
		if debug != nil {
			*debug += fmt.Sprintf("next latency is too high. next rtt = %d, threshold = %d\n", bestRouteCost, internal.MaxRTT)
		}
		routeState.NextLatencyTooHigh = true
		return false, false
	}

	// run try before you buy logic

	if !TryBeforeYouBuy(routeState, internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss) {
		if debug != nil {
			*debug += "leaving network next because try before you buy vetoed the session\n"
		}
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

func MakeRouteDecision_StayOnNetworkNext(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, relayNames []string, routeShader *RouteShader, routeState *RouteState, internal *InternalConfig, directLatency int32, nextLatency int32, predictedLatency int32, directPacketLoss float32, nextPacketLoss float32, currentRouteNumRelays int32, currentRouteRelays [MaxRelaysPerRoute]int32, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, out_updatedRouteCost *int32, out_updatedRouteNumRelays *int32, out_updatedRouteRelays []int32, debug *string) (bool, bool) {

	stayOnNetworkNext, nextRouteSwitched := MakeRouteDecision_StayOnNetworkNext_Internal(routeMatrix, fullRelaySet, relayNames, routeShader, routeState, internal, directLatency, nextLatency, predictedLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCost, destRelays, out_updatedRouteCost, out_updatedRouteNumRelays, out_updatedRouteRelays, debug)

	if routeState.Next && !stayOnNetworkNext {
		routeState.Next = false
		routeState.Veto = true
	}

	return stayOnNetworkNext, nextRouteSwitched
}

// ------------------------------------------------------

func GeneratePittle(output []byte, fromAddress []byte, fromPort uint16, toAddress []byte, toPort uint16, packetLength int) {

	var fromPortData [2]byte
	binary.LittleEndian.PutUint16(fromPortData[:], fromPort)

	var toPortData [2]byte
	binary.LittleEndian.PutUint16(toPortData[:], toPort)

	var packetLengthData [4]byte
	binary.LittleEndian.PutUint32(packetLengthData[:], uint32(packetLength))

	sum := uint16(0)

    for i := 0; i < len(fromAddress); i++ {
    	sum += uint16(fromAddress[i])
    }

    sum += uint16(fromPortData[0])
    sum += uint16(fromPortData[1])

    for i := 0; i < len(toAddress); i++ {
    	sum += uint16(toAddress[i])
    }

    sum += uint16(toPortData[0])
    sum += uint16(toPortData[1])

    sum += uint16(packetLengthData[0])
    sum += uint16(packetLengthData[1])
    sum += uint16(packetLengthData[2])
    sum += uint16(packetLengthData[3])

	var sumData [2]byte
	binary.LittleEndian.PutUint16(sumData[:], sum)

    output[0] = 1 | ( sumData[0] ^ sumData[1] ^ 193 );
    output[1] = 1 | ( ( 255 - output[0] ) ^ 113 );
}

func GenerateChonkle(output []byte, magic []byte, fromAddressData []byte, fromPort uint16, toAddressData []byte, toPort uint16, packetLength int) {

	var fromPortData [2]byte
	binary.LittleEndian.PutUint16(fromPortData[:], fromPort)

	var toPortData [2]byte
	binary.LittleEndian.PutUint16(toPortData[:], toPort)

	var packetLengthData [4]byte
	binary.LittleEndian.PutUint32(packetLengthData[:], uint32(packetLength))

	hash := fnv.New64a()
	hash.Write(magic)
	hash.Write(fromAddressData)
	hash.Write(fromPortData[:])
	hash.Write(toAddressData)
	hash.Write(toPortData[:])
	hash.Write(packetLengthData[:])
	hashValue := hash.Sum64()

	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], uint64(hashValue))

    output[0] = ( ( data[6] & 0xC0 ) >> 6 ) + 42
    output[1] = ( data[3] & 0x1F ) + 200
    output[2] = ( ( data[2] & 0xFC ) >> 2 ) + 5
    output[3] = data[0]
    output[4] = ( data[2] & 0x03 ) + 78
    output[5] = ( data[4] & 0x7F ) + 96
    output[6] = ( ( data[1] & 0xFC ) >> 2 ) + 100
    if ( data[7] & 1 ) == 0 { 
    	output[7] = 79
    } else { 
    	output[7] = 7 
    }
    if ( data[4] & 0x80 ) == 0 {
    	output[8] = 37
    } else { 
    	output[8] = 83
    }
    output[9] = ( data[5] & 0x07 ) + 124
    output[10] = ( ( data[1] & 0xE0 ) >> 5 ) + 175
    output[11] = ( data[6] & 0x3F ) + 33
    value := ( data[1] & 0x03 ); 
    if value == 0 { 
    	output[12] = 97
    } else if value == 1 { 
    	output[12] = 5
    } else if value == 2 { 
    	output[12] = 43
    } else { 
    	output[12] = 13
    }
    output[13] = ( ( data[5] & 0xF8 ) >> 3 ) + 210
    output[14] = ( ( data[7] & 0xFE ) >> 1 ) + 17
}

func BasicPacketFilter(data []byte, packetLength int) bool {

    if packetLength < 18 {
        return false
    }

    if data[0] < 0x01 || data[0] > 0x63 {
        return false
    }

    if data[1] < 0x2A || data[1] > 0x2D {
        return false
    }

    if data[2] < 0xC8 || data[2] > 0xE7 {
        return false
    }

    if data[3] < 0x05 || data[3] > 0x44 {
        return false
    }

    if data[5] < 0x4E || data[5] > 0x51 {
        return false
    }

    if data[6] < 0x60 || data[6] > 0xDF {
        return false
    }

    if data[7] < 0x64 || data[7] > 0xE3 {
        return false
    }

    if data[8] != 0x07 && data[8] != 0x4F {
        return false
    }

    if data[9] != 0x25 && data[9] != 0x53 {
        return false
    }
    
    if data[10] < 0x7C || data[10] > 0x83 {
        return false
    }

    if data[11] < 0xAF || data[11] > 0xB6 {
        return false
    }

    if data[12] < 0x21 || data[12] > 0x60 {
        return false
    }

    if data[13] != 0x61 && data[13] != 0x05 && data[13] != 0x2B && data[13] != 0x0D {
        return false
    }

    if data[14] < 0xD2 || data[14] > 0xF1 {
        return false
    }

    if data[15] < 0x11 || data[15] > 0x90 {
        return false
    }

    return true
}

func AdvancedPacketFilter(data []byte, magic []byte, fromAddress []byte, fromPort uint16, toAddress []byte, toPort uint16, packetLength int) bool {
    if packetLength < 18 {
        return false;
    }
    var a [15]byte
    var b [2]byte
    GenerateChonkle(a[:], magic, fromAddress, fromPort, toAddress, toPort, packetLength)
    GeneratePittle(b[:], fromAddress, fromPort, toAddress, toPort, packetLength)
    if bytes.Compare(a[0:15], data[1:16]) != 0 {
        return false
    }
    if bytes.Compare(b[0:2], data[packetLength-2:packetLength]) != 0 {
        return false
    }
    return true;
}

func GetAddressData(address *net.UDPAddr, addressData []byte, addressPort *uint16, addressBytes *int) {

	// todo
	*addressPort = 0
	*addressBytes = 0

	/*
    next_assert( address );
    if ( address->type == NEXT_ADDRESS_IPV4 )
    {
        address_data[0] = address->data.ipv4[0];
        address_data[1] = address->data.ipv4[1];
        address_data[2] = address->data.ipv4[2];
        address_data[3] = address->data.ipv4[3];
        *address_bytes = 4;
    }
    else if ( address->type == NEXT_ADDRESS_IPV6 )
    {
        for ( int i = 0; i < 8; ++i )
        {
            address_data[i*2]   = address->data.ipv6[i] >> 8;
            address_data[i*2+1] = address->data.ipv6[i] & 0xFF;
        }
        *address_bytes = 16;
    }
    else
    {
        *address_bytes = 0;
    }
    *address_port = address->port;
    */
}
