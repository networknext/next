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
	"math/bits"
	"net"
	"os"
	"strconv"
	"sync"
	"unsafe"
	"runtime/debug"
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

func GetBestRouteCost(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32) int32 {
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
				cost := sourceRelayCost[i] + entry.RouteCost[0]
				if cost < bestRouteCost {
					bestRouteCost = cost
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

func GetBestRoutes(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, bestRoutes []BestRoute, numBestRoutes *int, routeDiversity *int32) {
	numRoutes := 0
	maxRoutes := len(bestRoutes)
	for i := range sourceRelays {
		// IMPORTANT: RTT = 255 signals the source relay is unroutable
		if sourceRelayCost[i] >= 255 {
			Debug("Source Relay is unroutable!")
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
			Debug("Number of routes in entry: %v", int(entry.NumRoutes))
			for k := 0; k < int(entry.NumRoutes); k++ {
				cost := entry.RouteCost[k] + sourceRelayCost[i]
				Debug("route cost: %v", cost)
				Debug("max cost: %v", cost)
				if cost > maxCost {
					break
				}
				bestRoutes[numRoutes].Cost = cost
				bestRoutes[numRoutes].NumRelays = entry.RouteNumRelays[k]
				for l := int32(0); l < bestRoutes[numRoutes].NumRelays; l++ {
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

func GetRandomBestRoute(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, threshold int32, out_bestRouteCost *int32, out_bestRouteNumRelays *int32, out_bestRouteRelays *[MaxRelaysPerRoute]int32, debug *string) (foundRoute bool, routeDiversity int32) {

	foundRoute = false
	routeDiversity = 0

	if maxCost == -1 {
		return
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
		return
	}

	numBestRoutes := 0
	bestRoutes := make([]BestRoute, 1024)
	GetBestRoutes(routeMatrix, sourceRelays, sourceRelayCost, destRelays, bestRouteCost+threshold, bestRoutes, &numBestRoutes, &routeDiversity)
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

func GetBestRoute_Initial(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, selectThreshold int32, out_bestRouteCost *int32, out_bestRouteNumRelays *int32, out_bestRouteRelays *[MaxRelaysPerRoute]int32, debug *string) (hasRoute bool, routeDiversity int32) {

	return GetRandomBestRoute(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, selectThreshold, out_bestRouteCost, out_bestRouteNumRelays, out_bestRouteRelays, debug)
}

func GetBestRoute_Update(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, maxCost int32, selectThreshold int32, switchThreshold int32, currentRouteNumRelays int32, currentRouteRelays [MaxRelaysPerRoute]int32, out_updatedRouteCost *int32, out_updatedRouteNumRelays *int32, out_updatedRouteRelays *[MaxRelaysPerRoute]int32, debug *string) (routeChanged bool, routeLost bool) {

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

	if currentRouteCost > bestRouteCost+switchThreshold {
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
}

type InternalConfig struct {
	RouteSelectThreshold       int32
	RouteSwitchThreshold       int32
	MaxLatencyTradeOff         int32
	RTTVeto_Default            int32
	RTTVeto_Multipath          int32
	RTTVeto_PacketLoss         int32
	MultipathOverloadThreshold int32
	TryBeforeYouBuy            bool
	ForceNext                  bool
	LargeCustomer              bool
	Uncommitted                bool
	MaxRTT                     int32
	HighFrequencyPings         bool
	RouteDiversity             int32
	MultipathThreshold         int32
	EnableVanityMetrics        bool
}

func NewInternalConfig() InternalConfig {
	return InternalConfig{
		RouteSelectThreshold:       2,
		RouteSwitchThreshold:       5,
		MaxLatencyTradeOff:         20,
		RTTVeto_Default:            -10,
		RTTVeto_Multipath:          -20,
		RTTVeto_PacketLoss:         -30,
		MultipathOverloadThreshold: 500,
		TryBeforeYouBuy:            false,
		ForceNext:                  false,
		LargeCustomer:              false,
		Uncommitted:                false,
		MaxRTT:                     300,
		HighFrequencyPings:         true,
		RouteDiversity:             0,
		MultipathThreshold:         25,
		EnableVanityMetrics:        false,
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

func MakeRouteDecision_TakeNetworkNext(routeMatrix []RouteEntry, routeShader *RouteShader, routeState *RouteState, multipathVetoUsers map[uint64]bool, internal *InternalConfig, directLatency int32, directPacketLoss float32, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, out_routeCost *int32, out_routeNumRelays *int32, out_routeRelays []int32, out_routeDiversity *int32, debug *string) bool {

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

	reducePacketLoss := false
	if routeShader.ReducePacketLoss && directPacketLoss > routeShader.AcceptablePacketLoss {
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

	hasRoute, routeDiversity := GetBestRoute_Initial(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, selectThreshold, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays, debug)

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

func MakeRouteDecision_StayOnNetworkNext_Internal(routeMatrix []RouteEntry, relayNames []string, routeShader *RouteShader, routeState *RouteState, internal *InternalConfig, directLatency int32, nextLatency int32, predictedLatency int32, directPacketLoss float32, nextPacketLoss float32, currentRouteNumRelays int32, currentRouteRelays [MaxRelaysPerRoute]int32, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, out_updatedRouteCost *int32, out_updatedRouteNumRelays *int32, out_updatedRouteRelays []int32, debug *string) (bool, bool) {

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

	routeSwitched, routeLost := GetBestRoute_Update(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, internal.RouteSelectThreshold, internal.RouteSwitchThreshold, currentRouteNumRelays, currentRouteRelays, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays, debug)

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

func MakeRouteDecision_StayOnNetworkNext(routeMatrix []RouteEntry, relayNames []string, routeShader *RouteShader, routeState *RouteState, internal *InternalConfig, directLatency int32, nextLatency int32, predictedLatency int32, directPacketLoss float32, nextPacketLoss float32, currentRouteNumRelays int32, currentRouteRelays [MaxRelaysPerRoute]int32, sourceRelays []int32, sourceRelayCost []int32, destRelays []int32, out_updatedRouteCost *int32, out_updatedRouteNumRelays *int32, out_updatedRouteRelays []int32, debug *string) (bool, bool) {

	stayOnNetworkNext, nextRouteSwitched := MakeRouteDecision_StayOnNetworkNext_Internal(routeMatrix, relayNames, routeShader, routeState, internal, directLatency, nextLatency, predictedLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCost, destRelays, out_updatedRouteCost, out_updatedRouteNumRelays, out_updatedRouteRelays, debug)

	if routeState.Next && !stayOnNetworkNext {
		routeState.Next = false
		routeState.Veto = true
	}

	return stayOnNetworkNext, nextRouteSwitched
}

// ---------------------------------------------------

/*
func uintsToBytes(data []uint32) []byte {
	return *(*[]byte)(unsafe.Pointer(&data[0]))
}

func bytesToUints(data []byte) []uint32 {
	return *(*[]uint32)(unsafe.Pointer(&data[0]))
}
*/

func Log2(x uint32) int {
	a := x | (x >> 1)
	b := a | (a >> 2)
	c := b | (b >> 4)
	d := c | (c >> 8)
	e := d | (d >> 16)
	f := e >> 1
	return bits.OnesCount32(f)
}

func BitsRequired(min uint32, max uint32) int {
	if min == max {
		return 0
	} else {
		return Log2(max-min) + 1
	}
}

func BitsRequiredSigned(min int32, max int32) int {
	if min == max {
		return 0
	} else {
		return Log2(uint32(max-min)) + 1
	}
}

func SequenceGreaterThan(s1 uint16, s2 uint16) bool {
	return ((s1 > s2) && (s1-s2 <= 32768)) ||
		((s1 < s2) && (s2-s1 > 32768))
}

func SequenceLessThan(s1 uint16, s2 uint16) bool {
	return SequenceGreaterThan(s2, s1)
}

func SignedToUnsigned(n int32) uint32 {
	return uint32((n << 1) ^ (n >> 31))
}

func UnsignedToSigned(n uint32) int32 {
	return int32(n>>1) ^ (-int32(n & 1))
}

type BitWriter struct {
	buffer      []byte
	scratch     uint64
	numBits     int
	bitsWritten int
	wordIndex   int
	scratchBits int
	numWords    int
}

func CreateBitWriter(buffer []byte) (*BitWriter, error) {
	bytes := len(buffer)
	if bytes%4 != 0 {
		return nil, fmt.Errorf("bitwriter bytes must be a multiple of 4")
	}
	numWords := bytes / 4
	return &BitWriter{
		buffer:   buffer,
		numBits:  numWords * 32,
		numWords: numWords,
	}, nil

}

func HostToNetwork(x uint32) uint32 {
	return x
}

func NetworkToHost(x uint32) uint32 {
	return x
}

func (writer *BitWriter) WriteBits(value uint32, bits int) error {

	writer.scratch |= uint64(value) << uint(writer.scratchBits)

	writer.scratchBits += bits

	if writer.scratchBits >= 32 {
		*(*uint32)(unsafe.Pointer(&writer.buffer[writer.wordIndex*4])) = HostToNetwork(uint32(writer.scratch & 0xFFFFFFFF))
		writer.scratch >>= 32
		writer.scratchBits -= 32
		writer.wordIndex++
	}

	writer.bitsWritten += bits

	return nil
}

func (writer *BitWriter) WriteAlign() error {
	remainderBits := writer.bitsWritten % 8
	if remainderBits != 0 {
		err := writer.WriteBits(uint32(0), 8-remainderBits)
		if err != nil {
			return err
		}
	}
	return nil
}

func (writer *BitWriter) WriteBytes(data []byte) error {
	
	headBytes := (4 - (writer.bitsWritten%32)/8) % 4
	if headBytes > len(data) {
		headBytes = len(data)
	}
	
	for i := 0; i < headBytes; i++ {
		writer.WriteBits(uint32(data[i]), 8)
	}
	
	if headBytes == len(data) {
		return nil
	}

	if err := writer.FlushBits(); err != nil {
		return err
	}
	numWords := (len(data) - headBytes) / 4
	if numWords > 0 {
		*(*uint32)(unsafe.Pointer(&writer.buffer[writer.wordIndex*4])) = *(*uint32)(unsafe.Pointer(&data[headBytes]))
		writer.bitsWritten += numWords * 32
		writer.wordIndex += numWords
		writer.scratch = 0
	}

	tailStart := headBytes + numWords*4
	tailBytes := len(data) - tailStart

	for i := 0; i < tailBytes; i++ {
		err := writer.WriteBits(uint32(data[tailStart+i]), 8)
		if err != nil {
			return err
		}
	}

	return nil
}

func (writer *BitWriter) FlushBits() error {
	if writer.scratchBits != 0 {
		*(*uint32)(unsafe.Pointer(&writer.buffer[writer.wordIndex*4])) = HostToNetwork(uint32(writer.scratch & 0xFFFFFFFF))
		writer.scratch >>= 32
		writer.scratchBits = 0
		writer.wordIndex++
	}
	return nil
}

func (writer *BitWriter) GetAlignBits() int {
	return (8 - (writer.bitsWritten % 8)) % 8
}

func (writer *BitWriter) GetBitsWritten() int {
	return writer.bitsWritten
}

func (writer *BitWriter) GetBitsAvailable() int {
	return writer.numBits - writer.bitsWritten
}

func (writer *BitWriter) GetData() []byte {
	return writer.buffer
}

func (writer *BitWriter) GetBytesWritten() int {
	return (writer.bitsWritten + 7) / 8
}

type BitReader struct {
	buffer      []byte
	numBits     int
	numBytes    int
	numWords    int
	bitsRead    int
	scratch     uint64
	scratchBits int
	wordIndex   int
}

func CreateBitReader(data []byte) *BitReader {
	return &BitReader{
		buffer:   data,
		numBits:  len(data) * 8,
		numBytes: len(data),
		numWords: (len(data) + 3) / 4,
	}
}

func (reader *BitReader) WouldReadPastEnd(bits int) bool {
	return reader.bitsRead+bits > reader.numBits
}

func (reader *BitReader) ReadBits(bits int) (uint32, error) {

	if reader.bitsRead+bits > reader.numBits {
		return 0, fmt.Errorf("call would read past end of buffer")
	}

	reader.bitsRead += bits

	if reader.scratchBits < bits {
		if reader.wordIndex >= reader.numWords {
			return 0, fmt.Errorf("would read past end of buffer")
		}
		reader.scratch |= uint64(NetworkToHost(*(*uint32)(unsafe.Pointer(&reader.buffer[reader.wordIndex*4])))) << uint(reader.scratchBits)
		reader.scratchBits += 32
		reader.wordIndex++
	}

	output := reader.scratch & ((uint64(1) << uint(bits)) - 1)

	reader.scratch >>= uint(bits)
	reader.scratchBits -= bits

	return uint32(output), nil
}

func (reader *BitReader) ReadAlign() error {
	remainderBits := reader.bitsRead % 8
	if remainderBits != 0 {
		_, err := reader.ReadBits(8 - remainderBits)
		if err != nil {
			return err
		}
	}
	return nil
}

func (reader *BitReader) ReadBytes(buffer []byte) error {

	if reader.bitsRead+len(buffer)*8 > reader.numBits {
		return fmt.Errorf("would read past end of buffer")
	}

	headBytes := (4 - (reader.bitsRead%32)/8) % 4
	if headBytes > len(buffer) {
		headBytes = len(buffer)
	}
	for i := 0; i < headBytes; i++ {
		value, err := reader.ReadBits(8)
		if err != nil {
			return err
		}
		buffer[i] = byte(value)
	}
	if headBytes == len(buffer) {
		return nil
	}

	numWords := (len(buffer) - headBytes) / 4

	if numWords > 0 {
		*(*uint32)(unsafe.Pointer(&buffer[headBytes])) = *(*uint32)(unsafe.Pointer(&reader.buffer[reader.wordIndex*4]))
		reader.bitsRead += numWords * 32
		reader.wordIndex += numWords
		reader.scratchBits = 0
	}

	tailStart := headBytes + numWords*4
	tailBytes := len(buffer) - tailStart

	for i := 0; i < tailBytes; i++ {
		value, err := reader.ReadBits(8)
		if err != nil {
			return err
		}
		buffer[tailStart+i] = byte(value)
	}

	return nil
}

func (reader *BitReader) GetAlignBits() int {
	return (8 - reader.bitsRead%8) % 8
}

func (reader *BitReader) GetBitsRead() int {
	return reader.bitsRead
}

func (reader *BitReader) GetBitsRemaining() int {
	return reader.numBits - reader.bitsRead
}

// ---------------------------------------------------

type Stream interface {
	IsWriting() bool
	IsReading() bool
	SerializeInteger(value *int32, min int32, max int32)
	SerializeBits(value *uint32, bits int)
	SerializeUint32(value *uint32)
	SerializeBool(value *bool)
	SerializeFloat32(value *float32)
	SerializeUint64(value *uint64)
	SerializeFloat64(value *float64)
	SerializeBytes(data []byte)
	SerializeString(value *string, maxSize int)
	SerializeIntRelative(previous *int32, current *int32)
	SerializeAckRelative(sequence uint16, ack *uint16)
	SerializeSequenceRelative(sequence1 uint16, sequence2 *uint16)
	SerializeAlign()
	SerializeAddress(addr *net.UDPAddr)
	GetAlignBits() int
	GetBytesProcessed() int
	GetBitsProcessed() int
	Error() error
	Flush()
}

type WriteStream struct {
	writer *BitWriter
	err    error
}

func CreateWriteStream(buffer []byte) (*WriteStream, error) {
	writer, err := CreateBitWriter(buffer)
	if err != nil {
		return nil, err
	}
	return &WriteStream{
		writer: writer,
	}, nil
}

func (stream *WriteStream) IsWriting() bool {
	return true
}

func (stream *WriteStream) IsReading() bool {
	return false
}

// todo: lame function name
func (stream *WriteStream) error(err error) {
	if err != nil && stream.err == nil {
		stream.err = fmt.Errorf("%v\n%s", err, string(debug.Stack()))
	}
}

func (stream *WriteStream) Error() error {
	return stream.err
}

func (stream *WriteStream) SerializeInteger(value *int32, min int32, max int32) {
	if stream.err != nil {
		return
	}
	if min >= max {
		stream.error(fmt.Errorf("min (%d) should be less than max (%d)", min, max))
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if *value < min {
		stream.error(fmt.Errorf("value (%d) should be at least min (%d)", *value, min))
		return
	}
	if *value > max {
		stream.error(fmt.Errorf("value (%d) should be no more than max (%d)", *value, max))
		return
	}
	bits := BitsRequired(uint32(min), uint32(max))
	unsignedValue := uint32(*value - min)
	stream.error(stream.writer.WriteBits(unsignedValue, bits))
}

func (stream *WriteStream) SerializeBits(value *uint32, bits int) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if bits < 0 || bits > 32 {
		stream.error(fmt.Errorf("bits (%d) should be between 0 and 32", bits))
		return
	}
	stream.error(stream.writer.WriteBits(*value, bits))
}

func (stream *WriteStream) SerializeUint32(value *uint32) {
	stream.SerializeBits(value, 32)
}

func (stream *WriteStream) SerializeBool(value *bool) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	uint32Value := uint32(0)
	if *value {
		uint32Value = 1
	}
	stream.error(stream.writer.WriteBits(uint32Value, 1))
}

func (stream *WriteStream) SerializeFloat32(value *float32) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	stream.error(stream.writer.WriteBits(math.Float32bits(*value), 32))
}

func (stream *WriteStream) SerializeUint64(value *uint64) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	lo := uint32(*value & 0xFFFFFFFF)
	hi := uint32(*value >> 32)
	stream.SerializeBits(&lo, 32)
	stream.SerializeBits(&hi, 32)
}

func (stream *WriteStream) SerializeFloat64(value *float64) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	uint64Value := math.Float64bits(*value)
	stream.SerializeUint64(&uint64Value)
}

func (stream *WriteStream) SerializeBytes(data []byte) {
	if stream.err != nil {
		return
	}
	if len(data) == 0 {
		stream.error(fmt.Errorf("byte buffer should have more than 0 bytes"))
		return
	}
	stream.SerializeAlign()
	stream.error(stream.writer.WriteBytes(data))
}

func (stream *WriteStream) SerializeString(value *string, maxSize int) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if maxSize <= 0 {
		stream.error(fmt.Errorf("maxSize (%d) should be > 0", maxSize))
		return
	}
	length := int32(len(*value))
	if length > int32(maxSize) {
		stream.error(fmt.Errorf("string is longer than maxSize"))
		return
	}
	min := int32(0)
	max := int32(maxSize - 1)
	stream.SerializeInteger(&length, min, max)	
	if length > 0 {
		stream.SerializeBytes([]byte(*value))
	}
}

func (stream *WriteStream) SerializeIntRelative(previous *int32, current *int32) {
	if stream.err != nil {
		return
	}
	if previous == nil {
		stream.error(fmt.Errorf("previous is nil"))
		return
	}
	if current == nil {
		stream.error(fmt.Errorf("current is nil"))
		return
	}
	if *previous >= *current {
		stream.error(fmt.Errorf("previous value should be less than current value"))
		return
	}

	difference := *current - *previous

	oneBit := difference == 1
	stream.SerializeBool(&oneBit)
	if oneBit {
		return
	}

	twoBits := difference <= 6
	stream.SerializeBool(&twoBits)
	if twoBits {
		min := int32(2)
		max := int32(6)
		stream.SerializeInteger(&difference, min, max)
		return
	}

	fourBits := difference <= 23
	stream.SerializeBool(&fourBits)
	if fourBits {
		min := int32(7)
		max := int32(23)
		stream.SerializeInteger(&difference, min, max)
		return
	}

	eightBits := difference <= 280
	stream.SerializeBool(&eightBits)
	if eightBits {
		min := int32(24)
		max := int32(280)
		stream.SerializeInteger(&difference, min, max)
		return
	}

	twelveBits := difference <= 4377
	stream.SerializeBool(&twelveBits)
	if twelveBits {
		min := int32(281)
		max := int32(4377)
		stream.SerializeInteger(&difference, min, max)
		return
	}

	sixteenBits := difference <= 4377
	stream.SerializeBool(&sixteenBits)
	if sixteenBits {
		min := int32(4378)
		max := int32(69914)
		stream.SerializeInteger(&difference, min, max)
		return
	}

	uint32Value := uint32(*current)
	stream.SerializeUint32(&uint32Value)
}

func (stream *WriteStream) SerializeAckRelative(sequence uint16, ack *uint16) {
	if ack == nil {
		stream.error(fmt.Errorf("ack is nil"))
		return
	}

	ackDelta := int32(0)
	if *ack < sequence {
		ackDelta = int32(sequence) - int32(*ack)
	} else {
		ackDelta = int32(sequence) + 65536 - int32(*ack)
	}

	if ackDelta < 0 {
		panic("ackDelta should never be < 0")
	}

	if ackDelta == 0 {
		stream.error(fmt.Errorf("ack should not equal sequence"))
		return
	}

	ackInRange := ackDelta <= 64
	stream.SerializeBool(&ackInRange)
	if stream.err != nil {
		return
	}
	if ackInRange {
		stream.SerializeInteger(&ackDelta, 1, 64)
	} else {
		uint32Value := uint32(*ack)
		stream.SerializeBits(&uint32Value, 16)
	}
}

func (stream *WriteStream) SerializeSequenceRelative(sequence1 uint16, sequence2 *uint16) {
	if stream.err != nil {
		return
	}
	if sequence2 == nil {
		stream.error(fmt.Errorf("sequence2 is nil"))
		return
	}
	a := int32(sequence1)
	b := int32(*sequence2)
	if sequence1 > *sequence2 {
		b += 65536
	}
	stream.SerializeIntRelative(&a, &b)
}

func (stream *WriteStream) SerializeAddress(addr *net.UDPAddr) {
	if stream.err != nil {
		return
	}
	if addr == nil {
		stream.error(fmt.Errorf("addr is nil"))
		return
	}

	addrType := uint32(0)
	if addr.IP == nil {
		addrType = IPAddressNone
	} else if addr.IP.To4() == nil {
		addrType = IPAddressIPv6
	} else {
		addrType = IPAddressIPv4
	}

	stream.SerializeBits(&addrType, 2)
	if stream.err != nil {
		return
	}
	if addrType == uint32(IPAddressIPv4) {
		stream.SerializeBytes(addr.IP[12:])
		if stream.err != nil {
			return
		}
		port := uint32(addr.Port)
		stream.SerializeBits(&port, 16)
	} else if addrType == uint32(IPAddressIPv6) {
		addr.IP = make([]byte, 16)
		for i := 0; i < 8; i++ {
			uint32Value := uint32(binary.BigEndian.Uint16(addr.IP[i*2:]))
			stream.SerializeBits(&uint32Value, 16)
			if stream.err != nil {
				return
			}
		}
		uint32Value := uint32(addr.Port)
		stream.SerializeBits(&uint32Value, 16)
	}
}

func (stream *WriteStream) SerializeAlign() {
	if stream.err != nil {
		return
	}
	stream.error(stream.writer.WriteAlign())
}

func (stream *WriteStream) GetAlignBits() int {
	return stream.writer.GetAlignBits()
}

func (stream *WriteStream) Flush() {
	if stream.err != nil {
		return
	}
	stream.error(stream.writer.FlushBits())
}

func (stream *WriteStream) GetData() []byte {
	return stream.writer.GetData()
}

func (stream *WriteStream) GetBytesProcessed() int {
	return stream.writer.GetBytesWritten()
}

func (stream *WriteStream) GetBitsProcessed() int {
	return stream.writer.GetBitsWritten()
}

type ReadStream struct {
	reader *BitReader
	err    error
}

func CreateReadStream(buffer []byte) *ReadStream {
	return &ReadStream{
		reader: CreateBitReader(buffer),
	}
}

// todo: stupid name
func (stream *ReadStream) error(err error) {
	if err != nil && stream.err == nil {
		stream.err = fmt.Errorf("%v\n%s", err, string(debug.Stack()))
	}
}

func (stream *ReadStream) Error() error {
	return stream.err
}

func (stream *ReadStream) IsWriting() bool {
	return false
}

func (stream *ReadStream) IsReading() bool {
	return true
}

func (stream *ReadStream) SerializeInteger(value *int32, min int32, max int32) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if min >= max {
		stream.error(fmt.Errorf("min (%d) should be less than max (%d)", min, max))
		return
	}
	bits := BitsRequiredSigned(min, max)
	if stream.reader.WouldReadPastEnd(bits) {
		stream.error(fmt.Errorf("would read past end of buffer"))
		return
	}
	unsignedValue, err := stream.reader.ReadBits(bits)
	if err != nil {
		stream.error(err)
		return
	}
	candidateValue := int32(unsignedValue) + min
	if candidateValue > max {
		stream.error(fmt.Errorf("value (%d) above max (%d)", candidateValue, max))
		return
	}
	*value = candidateValue
}

func (stream *ReadStream) SerializeBits(value *uint32, bits int) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if bits < 0 || bits > 32 {
		stream.error(fmt.Errorf("bits (%d) should be between 0 and 32 bits", bits))
		return
	}
	if stream.reader.WouldReadPastEnd(bits) {
		stream.error(fmt.Errorf("would read past end of buffer"))
		return
	}
	readValue, err := stream.reader.ReadBits(bits)
	if err != nil {
		stream.error(err)
		return
	}
	*value = readValue
}

func (stream *ReadStream) SerializeUint32(value *uint32) {
	stream.SerializeBits(value, 32)
}

func (stream *ReadStream) SerializeBool(value *bool) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if stream.reader.WouldReadPastEnd(1) {
		stream.error(fmt.Errorf("would read past end of buffer"))
		return
	}
	readValue, err := stream.reader.ReadBits(1)
	if err != nil {
		stream.error(err)
		return
	}
	*value = readValue != 0
}

func (stream *ReadStream) SerializeFloat32(value *float32) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if stream.reader.WouldReadPastEnd(32) {
		stream.error(fmt.Errorf("would read past end of buffer"))
		return
	}
	readValue, err := stream.reader.ReadBits(32)
	if err != nil {
		stream.error(err)
		return
	}
	*value = math.Float32frombits(readValue)
}

func (stream *ReadStream) SerializeUint64(value *uint64) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if stream.reader.WouldReadPastEnd(64) {
		stream.error(fmt.Errorf("would read past end of buffer"))
		return
	}
	lo, err := stream.reader.ReadBits(32)
	if err != nil {
		stream.error(err)
		return
	}
	hi, err := stream.reader.ReadBits(32)
	if err != nil {
		stream.error(err)
		return
	}
	*value = (uint64(hi) << 32) | uint64(lo)
}

func (stream *ReadStream) SerializeFloat64(value *float64) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	readValue := uint64(0)
	stream.SerializeUint64(&readValue)
	if stream.err != nil {
		return
	}
	*value = math.Float64frombits(readValue)
}

func (stream *ReadStream) SerializeBytes(data []byte) {
	if stream.err != nil {
		return
	}
	if len(data) == 0 {
		stream.error(fmt.Errorf("buffer should contain more than 0 bytes"))
		return
	}
	stream.SerializeAlign()
	if stream.err != nil {
		return
	}
	if stream.reader.WouldReadPastEnd(len(data) * 8) {
		stream.error(fmt.Errorf("SerializeBytes() would read past end of buffer"))
		return
	}
	stream.error(stream.reader.ReadBytes(data))
}

func (stream *ReadStream) SerializeString(value *string, maxSize int) {
	if stream.err != nil {
		return
	}
	if maxSize < 0 {
		stream.error(fmt.Errorf("maxSize (%d) should be > 0", maxSize))
		return
	}
	length := int32(0)
	min := int32(0)
	max := int32(maxSize - 1)
	stream.SerializeInteger(&length, min, max)
	if stream.err != nil {
		return
	}
	if length == 0 {
		*value = ""
		return
	}
	stringBytes := make([]byte, length)
	stream.SerializeBytes(stringBytes)
	*value = string(stringBytes)
}

func (stream *ReadStream) SerializeIntRelative(previous *int32, current *int32) {
	if stream.err != nil {
		return
	}
	if previous == nil {
		stream.error(fmt.Errorf("previous is nil"))
		return
	}
	if current == nil {
		stream.error(fmt.Errorf("current is nil"))
		return
	}
	oneBit := false
	stream.SerializeBool(&oneBit)
	if stream.err != nil {
		return
	}
	if oneBit {
		*current = *previous + 1
		return
	}

	twoBits := false
	stream.SerializeBool(&twoBits)
	if stream.err != nil {
		return
	}
	if twoBits {
		difference := int32(0)
		min := int32(2)
		max := int32(6)
		stream.SerializeInteger(&difference, min, max)
		*current = *previous + difference
		return
	}

	fourBits := false
	stream.SerializeBool(&fourBits)
	if stream.err != nil {
		return
	}
	if fourBits {
		difference := int32(0)
		min := int32(7)
		max := int32(32)
		stream.SerializeInteger(&difference, min, max)
		*current = *previous + difference
		return
	}

	eightBits := false
	stream.SerializeBool(&eightBits)
	if stream.err != nil {
		return
	}
	if eightBits {
		difference := int32(0)
		min := int32(24)
		max := int32(280)
		stream.SerializeInteger(&difference, min, max)
		*current = *previous + difference
		return
	}

	twelveBits := false
	stream.SerializeBool(&twelveBits)
	if stream.err != nil {
		return
	}
	if twelveBits {
		difference := int32(0)
		min := int32(281)
		max := int32(4377)
		stream.SerializeInteger(&difference, min, max)
		*current = *previous + difference
		return
	}

	sixteenBits := false
	stream.SerializeBool(&sixteenBits)
	if stream.err != nil {
		return
	}
	if sixteenBits {
		difference := int32(0)
		min := int32(4378)
		max := int32(69914)
		stream.SerializeInteger(&difference, min, max)
		*current = *previous + difference
		return
	}

	uint32Value := uint32(0)
	stream.SerializeUint32(&uint32Value)
	if stream.err != nil {
		return
	}
	*current = int32(uint32Value)
}

func (stream *ReadStream) SerializeAckRelative(sequence uint16, ack *uint16) {
	if ack == nil {
		stream.error(fmt.Errorf("ack is nil"))
		return
	}
	ackDelta := int32(0)
	ackInRange := false
	stream.SerializeBool(&ackInRange)
	if ackInRange {
		stream.SerializeInteger(&ackDelta, 1, 64)
		*ack = sequence - uint16(ackDelta)
	} else {
		uint32Value := uint32(0)
		stream.SerializeBits(&uint32Value, 16)
		*ack = uint16(uint32Value)
	}
}

func (stream *ReadStream) SerializeSequenceRelative(sequence1 uint16, sequence2 *uint16) {
	if stream.err != nil {
		return
	}
	if sequence2 == nil {
		stream.error(fmt.Errorf("sequence2 is nil"))
		return
	}
	a := int32(sequence1)
	b := int32(0)
	stream.SerializeIntRelative(&a, &b)
	if stream.err != nil {
		return
	}
	if b >= 65536 {
		b -= 65536
	}
	*sequence2 = uint16(b)
}

func (stream *ReadStream) SerializeAddress(addr *net.UDPAddr) {
	if stream.err != nil {
		return
	}
	if addr == nil {
		stream.error(fmt.Errorf("addr is nil"))
		return
	}
	addrType := uint32(0)
	stream.SerializeBits(&addrType, 2)
	if stream.err != nil {
		return
	}
	if addrType == uint32(IPAddressIPv4) {
		addr.IP = make([]byte, 16)
		addr.IP[10] = 255
		addr.IP[11] = 255
		stream.SerializeBytes(addr.IP[12:])
		if stream.err != nil {
			return
		}
		port := uint32(0)
		stream.SerializeBits(&port, 16)
		if stream.err != nil {
			return
		}
		addr.Port = int(port)
	} else if addrType == uint32(IPAddressIPv6) {
		addr.IP = make([]byte, 16)
		for i := 0; i < 8; i++ {
			uint32Value := uint32(0)
			stream.SerializeBits(&uint32Value, 16)
			if stream.err != nil {
				return
			}
			binary.BigEndian.PutUint16(addr.IP[i*2:], uint16(uint32Value))
		}
		uint32Value := uint32(0)
		stream.SerializeBits(&uint32Value, 16)
		if stream.err != nil {
			return
		}
		addr.Port = int(uint32Value)
	} else {
		*addr = net.UDPAddr{}
	}
}

func (stream *ReadStream) SerializeAlign() {
	alignBits := stream.reader.GetAlignBits()
	if stream.reader.WouldReadPastEnd(alignBits) {
		stream.error(fmt.Errorf("SerializeAlign() would read past end of buffer"))
		return
	}
	stream.error(stream.reader.ReadAlign())
}

func (stream *ReadStream) Flush() {
}

func (stream *ReadStream) GetAlignBits() int {
	return stream.reader.GetAlignBits()
}

func (stream *ReadStream) GetBitsProcessed() int {
	return stream.reader.GetBitsRead()
}

func (stream *ReadStream) GetBytesProcessed() int {
	return (stream.reader.GetBitsRead() + 7) / 8
}

// --------------------------------------------------------
