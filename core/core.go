
package main

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
    "fmt"
    "errors"
    "encoding/binary"
    "unsafe"
    "net"
    "runtime"
    "sync"
    "sort"
    "math"
    "math/rand"
    "strconv"
    "crypto/ed25519"
)

const NEXT_MAX_NODES = 7
const NEXT_ADDRESS_BYTES = 19
const NEXT_ROUTE_TOKEN_BYTES = 76
const NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES = 116
const NEXT_CONTINUE_TOKEN_BYTES = 17
const NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES = 57
const NEXT_PRIVATE_KEY_BYTES = 32

func ProtocolVersionAtLeast(serverMajor int32, serverMinor int32, serverPatch int32, targetMajor int32, targetMinor int32, targetPatch int32) bool {
    serverVersion := ( (serverMajor&0xFF) << 16 ) | ( (serverMinor&0xFF) << 8 ) | (serverPatch&0xFF);
    targetVersion := ( (targetMajor&0xFF) << 16 ) | ( (targetMinor&0xFF) << 8 ) | (targetPatch&0xFF);
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

func GenerateCustomerKeyPair() ([]byte, []byte, error) {
    customerId := make([]byte, 8)
    rand.Read(customerId)
    publicKey, privateKey, err := ed25519.GenerateKey(nil)
    if err != nil {
        return nil, nil, err
    }
    customerPublicKey := make([]byte, 0)
    customerPublicKey = append(customerPublicKey, customerId...)
    customerPublicKey = append(customerPublicKey, publicKey...)
    customerPrivateKey := make([]byte, 0)
    customerPrivateKey = append(customerPrivateKey, customerId...)
    customerPrivateKey = append(customerPrivateKey, privateKey...)
    return customerPublicKey, customerPrivateKey, nil
}

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
        return &net.UDPAddr{IP: buffer[1:], Port: ((int)(binary.LittleEndian.Uint16(buffer[17:])))}
    }
    return nil
}

// ---------------------------------------------------

const MaxRelaysPerRoute = 5
const MaxRoutesPerEntry = 8

type RouteManager struct {
    NumRoutes      int
    RouteCost      [MaxRoutesPerEntry]int32
    RouteHash      [MaxRoutesPerEntry]uint32
    RouteNumRelays [MaxRoutesPerEntry]int32
    RouteRelays    [MaxRoutesPerEntry][MaxRelaysPerRoute]int32
}

func (manager *RouteManager) AddRoute(cost int32, relayDatacenter []uint64, relays ...int32) {

    // IMPORTANT: Filter out any route with two relays in the same datacenter. These routes are redundant.
    datacenterCheck := make(map[uint64]int, len(relays))
    for i := range relays {
        if _, exists := datacenterCheck[relayDatacenter[relays[i]]]; exists {
            return
        }
        datacenterCheck[relayDatacenter[relays[i]]] = 1
    }

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

func Optimize(numRelays int, cost []int32, costThreshold int32) []RouteEntry {

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
                            routes[ijIndex].RouteHash[0] = RouteHash(int32(i), int32(j))

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

// ---------------------------------------------------

type RouteToken struct {
    expireTimestamp uint64
    sessionId       uint64
    sessionVersion  uint8
    kbpsUp          uint32
    kbpsDown        uint32
    nextAddress     *net.UDPAddr
    privateKey      []byte
}

type ContinueToken struct {
    expireTimestamp uint64
    sessionId       uint64
    sessionVersion  uint8
}

const Crypto_kx_PUBLICKEYBYTES = C.crypto_kx_PUBLICKEYBYTES
const Crypto_box_PUBLICKEYBYTES = C.crypto_box_PUBLICKEYBYTES

const KeyBytes = 32
const NonceBytes = 24
const SignatureBytes = C.crypto_sign_BYTES
const PublicKeyBytes = C.crypto_sign_PUBLICKEYBYTES

func Encrypt(senderPrivateKey []byte, receiverPublicKey []byte, nonce []byte, buffer []byte, bytes int) error {
    result := C.crypto_box_easy((*C.uchar)(&buffer[0]),
        (*C.uchar)(&buffer[0]),
        C.ulonglong(bytes),
        (*C.uchar)(&nonce[0]),
        (*C.uchar)(&receiverPublicKey[0]),
        (*C.uchar)(&senderPrivateKey[0]))
    if result != 0 {
        return fmt.Errorf("failed to encrypt: result = %d", result)
    } else {
        return nil
    }
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

func Encrypt_ChaCha20(buffer []byte, additional []byte, privateKey []byte) ([]byte, []byte, error) {
    nonce := RandomBytes(C.crypto_aead_xchacha20poly1305_ietf_NPUBBYTES)
    encrypted := make([]byte, len(buffer)+C.crypto_aead_xchacha20poly1305_ietf_ABYTES)
    var encryptedLength = C.ulonglong(0)
    result := C.crypto_aead_xchacha20poly1305_ietf_encrypt((*C.uchar)(&encrypted[0]), &encryptedLength,
        (*C.uchar)(&buffer[0]), C.ulonglong(len(buffer)),
        (*C.uchar)(&additional[0]), C.ulonglong(len(additional)),
        nil, (*C.uchar)(&nonce[0]), (*C.uchar)(&privateKey[0]))
    if result != 0 {
        return nil, nil, fmt.Errorf("failed to encrypt chacha20: result = %d", result)
    } else {
        return encrypted, nonce, nil
    }
}

func Decrypt_ChaCha20(encrypted []byte, additional []byte, nonce []byte, privateKey []byte) ([]byte, error) {
    if len(encrypted) <= C.crypto_aead_xchacha20poly1305_ietf_ABYTES {
        return nil, fmt.Errorf("failed to decrypt chacha20: encrypted data is too small")
    }
    decrypted := make([]byte, len(encrypted)-C.crypto_aead_xchacha20poly1305_ietf_ABYTES)
    var decryptedLength = C.ulonglong(0)
    result := C.crypto_aead_xchacha20poly1305_ietf_decrypt((*C.uchar)(&decrypted[0]), &decryptedLength, nil,
        (*C.uchar)(&encrypted[0]), C.ulonglong(len(encrypted)),
        (*C.uchar)(&additional[0]), C.ulonglong(len(additional)),
        (*C.uchar)(&nonce[0]), (*C.uchar)(&privateKey[0]))
    if result != 0 {
        return nil, fmt.Errorf("failed to decrypt chacha20: result = %d", result)
    } else {
        return decrypted, nil
    }
}

func RandomBytes(bytes int) []byte {
    buffer := make([]byte, bytes)
    C.randombytes_buf(unsafe.Pointer(&buffer[0]), C.size_t(bytes))
    return buffer
}

func WriteRouteToken(token *RouteToken, buffer []byte) {
    binary.LittleEndian.PutUint64(buffer[0:], token.expireTimestamp)
    binary.LittleEndian.PutUint64(buffer[8:], token.sessionId)
    buffer[8+8] = token.sessionVersion
    binary.LittleEndian.PutUint32(buffer[8+8+1:], token.kbpsUp)
    binary.LittleEndian.PutUint32(buffer[8+8+1+4:], token.kbpsDown)
    WriteAddress(buffer[8+8+1+4+4:], token.nextAddress)
    copy(buffer[8+8+1+4+4+NEXT_ADDRESS_BYTES:], token.privateKey)
}

func ReadRouteToken(buffer []byte) (*RouteToken, error) {
    if len(buffer) < NEXT_ROUTE_TOKEN_BYTES {
        return nil, fmt.Errorf("buffer too small to read route token")
    }
    token := &RouteToken{}
    token.expireTimestamp = binary.LittleEndian.Uint64(buffer[0:])
    token.sessionId = binary.LittleEndian.Uint64(buffer[8:])
    token.sessionVersion = buffer[8+8]
    token.kbpsUp = binary.LittleEndian.Uint32(buffer[8+8+1:])
    token.kbpsDown = binary.LittleEndian.Uint32(buffer[8+8+1+4:])
    token.nextAddress = ReadAddress(buffer[8+8+1+4+4:])
    token.privateKey = make([]byte, NEXT_PRIVATE_KEY_BYTES)
    copy(token.privateKey, buffer[8+8+1+4+4+NEXT_ADDRESS_BYTES:])
    return token, nil
}

func WriteEncryptedRouteToken(buffer []byte, token *RouteToken, senderPrivateKey []byte, receiverPublicKey []byte) error {
    nonce := RandomBytes(NonceBytes)
    copy(buffer, nonce)
    WriteRouteToken(token, buffer[NonceBytes:])
    result := Encrypt(senderPrivateKey, receiverPublicKey, nonce, buffer[NonceBytes:], NEXT_ROUTE_TOKEN_BYTES)
    return result
}

func ReadEncryptedRouteToken(tokenData []byte, senderPublicKey []byte, receiverPrivateKey []byte) (*RouteToken, error) {
    if len(tokenData) < NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES {
        return nil, fmt.Errorf("not enough bytes for encrypted route token")
    }
    nonce := tokenData[0 : C.crypto_box_NONCEBYTES-1]
    tokenData = tokenData[C.crypto_box_NONCEBYTES:]
    if err := Decrypt(senderPublicKey, receiverPrivateKey, nonce, tokenData, NEXT_ROUTE_TOKEN_BYTES+C.crypto_box_MACBYTES); err != nil {
        return nil, err
    }
    return ReadRouteToken(tokenData)
}

func WriteRouteTokens(expireTimestamp uint64, sessionId uint64, sessionVersion uint8, kbpsUp uint32, kbpsDown uint32, numNodes int, addresses []*net.UDPAddr, publicKeys [][]byte, masterPrivateKey [KeyBytes]byte) ([]byte, error) {
    if numNodes < 1 || numNodes > NEXT_MAX_NODES {
        return nil, fmt.Errorf("invalid numNodes %d. expected value in range [1,%d]", numNodes, NEXT_MAX_NODES)
    }
    privateKey := RandomBytes(KeyBytes)
    tokenData := make([]byte, numNodes*NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)
    for i := 0; i < numNodes; i++ {
        token := &RouteToken{}
        token.expireTimestamp = expireTimestamp
        token.sessionId = sessionId
        token.sessionVersion = sessionVersion
        token.kbpsUp = kbpsUp
        token.kbpsDown = kbpsDown
        if i != numNodes-1 {
            token.nextAddress = addresses[i+1]
        }
        token.privateKey = privateKey
        err := WriteEncryptedRouteToken(tokenData[i*NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], token, masterPrivateKey[:], publicKeys[i])
        if err != nil {
            return nil, err
        }
    }
    return tokenData, nil
}

func WriteContinueToken(token *ContinueToken, buffer []byte) {
    binary.LittleEndian.PutUint64(buffer[0:], token.expireTimestamp)
    binary.LittleEndian.PutUint64(buffer[8:], token.sessionId)
    buffer[8+8] = token.sessionVersion
}

func ReadContinueToken(buffer []byte) (*ContinueToken, error) {
    if len(buffer) < NEXT_CONTINUE_TOKEN_BYTES {
        return nil, fmt.Errorf("buffer too small to read continue token")
    }
    token := &ContinueToken{}
    token.expireTimestamp = binary.LittleEndian.Uint64(buffer[0:])
    token.sessionId = binary.LittleEndian.Uint64(buffer[8:])
    token.sessionVersion = buffer[8+8]
    return token, nil
}

func WriteEncryptedContinueToken(buffer []byte, token *ContinueToken, senderPrivateKey []byte, receiverPublicKey []byte) error {
    nonce := RandomBytes(NonceBytes)
    copy(buffer, nonce)
    WriteContinueToken(token, buffer[NonceBytes:])
    result := Encrypt(senderPrivateKey, receiverPublicKey, nonce, buffer[NonceBytes:], NEXT_CONTINUE_TOKEN_BYTES)
    return result
}

func ReadEncryptedContinueToken(tokenData []byte, senderPublicKey []byte, receiverPrivateKey []byte) (*ContinueToken, error) {
    if len(tokenData) < NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES {
        return nil, fmt.Errorf("not enough bytes for encrypted continue token")
    }
    nonce := tokenData[0 : C.crypto_box_NONCEBYTES-1]
    tokenData = tokenData[C.crypto_box_NONCEBYTES:]
    if err := Decrypt(senderPublicKey, receiverPrivateKey, nonce, tokenData, NEXT_CONTINUE_TOKEN_BYTES+C.crypto_box_MACBYTES); err != nil {
        return nil, err
    }
    return ReadContinueToken(tokenData)
}

func WriteContinueTokens(expireTimestamp uint64, sessionId uint64, sessionVersion uint8, numNodes int, publicKeys [][]byte, masterPrivateKey [KeyBytes]byte) ([]byte, error) {
    if numNodes < 1 || numNodes > NEXT_MAX_NODES {
        return nil, fmt.Errorf("invalid numNodes %d. expected value in range [1,%d]", numNodes, NEXT_MAX_NODES)
    }
    tokenData := make([]byte, numNodes*NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES)
    for i := 0; i < numNodes; i++ {
        token := &ContinueToken{}
        token.expireTimestamp = expireTimestamp
        token.sessionId = sessionId
        token.sessionVersion = sessionVersion
        err := WriteEncryptedContinueToken(tokenData[i*NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES:], token, masterPrivateKey[:], publicKeys[i])
        if err != nil {
            return nil, err
        }
    }
    return tokenData, nil
}

// -------------------------------------------

func GetBestRouteCost(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost[] int32, destRelays []int32) int32 {
    bestRouteCost := int32(math.MaxInt32)
    for i := range sourceRelays {
        if sourceRelayCost[i] < int32(0) {
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
    return bestRouteCost
}

func ReverseRoute(route []int32) {
    for i, j := 0, len(route)-1; i < j; i, j = i+1, j-1 {
        route[i], route[j] = route[j], route[i]
    }    
}

func GetCurrentRouteCost(routeMatrix []RouteEntry, routeRelays []int32, sourceRelays []int32, sourceRelayCost[] int32, destRelays []int32) int32 {
    if len(routeRelays) == 0 {
        return -1
    }
    reversed := false
    if routeRelays[0] < routeRelays[len(routeRelays)-1] {
        ReverseRoute(routeRelays)
        destRelays, sourceRelays = sourceRelays, destRelays
        reversed = true
    }
    routeHash := RouteHash(routeRelays...)
    firstRouteRelay := routeRelays[0]
    for i := range sourceRelays {
        if sourceRelayCost[i] < int32(0) {
            continue
        }
        if sourceRelays[i] == firstRouteRelay {
            for j := range destRelays {
                sourceRelayIndex := sourceRelays[i]
                destRelayIndex := destRelays[j]
                if sourceRelayIndex == destRelayIndex {
                    continue
                }
                index := TriMatrixIndex(int(sourceRelayIndex), int(destRelayIndex))
                entry := &routeMatrix[index]
                for k := 0; k < int(entry.NumRoutes); k++ {
                    if entry.RouteHash[k] == routeHash && int(entry.RouteNumRelays[k]) == len(routeRelays) {
                        found := true
                        for l := range routeRelays {
                            if entry.RouteRelays[k][l] != routeRelays[l] {
                                found = false
                                break
                            }
                        }
                        if found {
                            sourceCost := int32(math.MaxInt32)
                            if reversed {
                                sourceRelays = destRelays
                                actualSourceRelay := routeRelays[len(routeRelays)-1]
                                for m := range sourceRelays {
                                    if sourceRelays[m] == actualSourceRelay {
                                        sourceCost = sourceRelayCost[m]
                                        break
                                    }
                                }
                            } else {
                                for m := range sourceRelays {
                                    if sourceRelays[m] == firstRouteRelay {
                                        sourceCost = sourceRelayCost[m]
                                        break
                                    }
                                }
                            }
                            if sourceCost == int32(math.MaxInt32) {
                                panic("this should never happen")
                            }
                            return sourceCost + entry.RouteCost[k]
                        }
                    }
                }
            }
        }
    }
    return -1
}

type BestRoute struct {
    Cost            int32    
    NumRelays       int32
    Relays          [MaxRelaysPerRoute]int32
    NeedToReverse   bool
}

func GetBestRoutes(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost[] int32, destRelays []int32, maxCost int32, bestRoutes []BestRoute, numBestRoutes *int) {
    numRoutes := 0
    maxRoutes := len(bestRoutes)
    for i := range sourceRelays {
        if sourceRelayCost[i] < int32(0) {
            continue
        }
        for j := range destRelays {
            sourceRelayIndex := sourceRelays[i]
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
                bestRoutes[numRoutes].NumRelays = entry.RouteNumRelays[k]
                for l := int32(0); l < bestRoutes[numRoutes].NumRelays; l++ {
                    bestRoutes[numRoutes].Relays[l] = entry.RouteRelays[k][l]
                }
                bestRoutes[numRoutes].NeedToReverse = sourceRelayIndex < destRelayIndex
                numRoutes++
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

func ReframeRoute(routeRelayIds []uint64, relayIdToIndex map[uint64]int32) ([]int32, error) {
    routeRelays := make([]int32, len(routeRelayIds))
    for i := range routeRelayIds {
        relayIndex, ok := relayIdToIndex[routeRelayIds[i]]
        if !ok {
            return nil, errors.New("relay id does not exist")
        }
        routeRelays[i] = relayIndex
    }
    return routeRelays, nil
}

// todo: need a function to convert a route expressed in terms of uint64 relayIds into a route in terms of relax indices in the current route matrix
// this function *can* return error, if one of the relays in the route no longer exists.

// todo
/*
func GetBestRoute_Initial(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost[] int32, destRelays []int32, maxCost int32, costThreshold int32, bestRoute *BestRoute) {
    
    bestRouteCost := GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, )

    if bestRouteCost > maxCost {
        bestRoute.NumRelays = 0
        bestRoute.Cost = -1
        return 
    }

    // todo: GetBestRoutes -- make sure to pass in a maxCost above which we won't even consider the routes as valid or useful

    // todo: randomly pick one of the best routes

}
*/

func GetBestRoute_Update() {

    /*
    get route in terms of indices in current route matrix (relayId -> relayIndex)

    if any relay in the current route does not exist in current route matrix, get best route initial

    routeCost := GetCurrentRouteCost
    
    if routeValid && Fabs(routeCost - bestRouteCost) > costThreshold {
        routeValid = false
    }
    
    if !routeValid {
        GetBestRoutes
        randomly pick out of best routes
    }

    hold current route
    */
}

func MakeRouteDecision_TakeNetworkNext() {
    // todo
}

func MakeRouteDecision_StayOnNetworkNext() {
    // todo
}

// -------------------------------------------
