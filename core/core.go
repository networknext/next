package main

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
    "fmt"
    "encoding/binary"
    "unsafe"
    "net"
    "runtime"
    "sync"
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

const MaxRelaysPerRoute = 5
const MaxRoutesPerEntry = 8

type RouteManager struct {
    NumRoutes       int
    RouteCost       [MaxRoutesPerEntry]int32
    RouteHash       [MaxRoutesPerEntry]uint32
    RouteNumRelays  [MaxRoutesPerEntry]int32
    RouteRelays     [MaxRoutesPerEntry][MaxRelaysPerRoute]int32
    RelayDatacenter []uint64
}

func (manager *RouteManager) AddRoute(cost int32, relays ...int32) {

    // IMPORTANT: Filter out any route with two relays in the same datacenter. These routes are redundant.
    datacenterCheck := make(map[uint64]int, len(relays))
    for i := range relays {
        if _, exists := datacenterCheck[manager.RelayDatacenter[relays[i]]]; exists {
            return
        }
        datacenterCheck[manager.RelayDatacenter[relays[i]]] = 1
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

func Optimize(numRelays int, cost []int32, costThreshold int32, relayDatacenter []uint64) []RouteEntry {

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

func GetCurrentRouteCost(routeMatrix []RouteEntry, routeNumRelays int32, routeRelays [MaxRelaysPerRoute]int32, sourceRelays []int32, sourceRelayCost[] int32, destRelays []int32) int32 {
    reversed := false
    if routeRelays[0] < routeRelays[routeNumRelays-1] {
        ReverseRoute(routeRelays[:routeNumRelays])
        destRelays, sourceRelays = sourceRelays, destRelays
        reversed = true
    }
    routeHash := RouteHash(routeRelays[:routeNumRelays]...)
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
                    if entry.RouteHash[k] == routeHash && entry.RouteNumRelays[k] == routeNumRelays {
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
                                actualSourceRelay := routeRelays[routeNumRelays-1]
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

func ReframeRoute(relayIdToIndex map[uint64]int32, routeRelayIds []uint64, out_routeRelays *[MaxRelaysPerRoute]int32) bool {
    for i := range routeRelayIds {
        relayIndex, ok := relayIdToIndex[routeRelayIds[i]]
        if !ok {
            return false
        }
        out_routeRelays[i] = relayIndex
    }
    return true
}

func ReframeRelays(relayIdToIndex map[uint64]int32, sourceRelayIds []uint64, sourceRelayLatency []int32, sourceRelayPacketLoss []float32, destRelayIds []uint64, out_numSourceRelays *int32, out_sourceRelays []int32, out_numDestRelays *int32, out_destRelays []int32) {
    
    numSourceRelays := int32(0)
    numDestRelays := int32(0)

    for i := range sourceRelayIds {
        if sourceRelayLatency[i] <= 0 {
            // you say your latency is 0ms? I don't believe you!
            continue
        }
        if sourceRelayPacketLoss[i] > 50.0 {
            // any source relay with > 50% PL in the last slice is bad news
            continue
        }
        sourceRelayIndex, ok := relayIdToIndex[sourceRelayIds[i]]
        if !ok {
            continue
        }
        out_sourceRelays[numSourceRelays] = sourceRelayIndex
        numSourceRelays++
    }

    for i := range destRelayIds {
        destRelayIndex, ok := relayIdToIndex[destRelayIds[i]]
        if !ok {
            continue
        }
        out_destRelays[numDestRelays] = destRelayIndex
        numDestRelays++
    }

    *out_numSourceRelays = numSourceRelays
    *out_numDestRelays = numDestRelays
}

func GetRandomBestRoute(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost[] int32, destRelays []int32, maxCost int32, out_bestRouteCost *int32, out_bestRouteNumRelays *int32, out_bestRouteRelays *[MaxRelaysPerRoute]int32) bool {
    
    if maxCost == -1 {
        return false
    }

    bestRouteCost := GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCost, destRelays)

    if bestRouteCost > maxCost {
        return false
    }

    numBestRoutes := 0
    bestRoutes := make([]BestRoute, 1024)
    GetBestRoutes(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, bestRoutes, &numBestRoutes)

    if numBestRoutes == 0 {
        return false
    }

    randomIndex := rand.Intn(numBestRoutes)

    *out_bestRouteCost = bestRoutes[randomIndex].Cost
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

func GetBestRoute_Initial(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost[] int32, destRelays []int32, maxCost int32, out_bestRouteCost *int32, out_bestRouteNumRelays *int32, out_bestRouteRelays *[MaxRelaysPerRoute]int32) bool {

    return GetRandomBestRoute(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, out_bestRouteCost, out_bestRouteNumRelays, out_bestRouteRelays)
}

func GetBestRoute_Update(routeMatrix []RouteEntry, sourceRelays []int32, sourceRelayCost[] int32, destRelays []int32, maxCost int32, costThreshold int32, currentRouteNumRelays int32, currentRouteRelays [MaxRelaysPerRoute]int32, out_updatedRouteCost *int32, out_updatedRouteNumRelays *int32, out_updatedRouteRelays *[MaxRelaysPerRoute]int32) bool {

    // if the current route no longer exists, pick a new route

    currentRouteCost := GetCurrentRouteCost(routeMatrix, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCost, destRelays)

    if currentRouteCost < 0 {
        GetRandomBestRoute(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, out_updatedRouteCost, out_updatedRouteNumRelays, out_updatedRouteRelays)
        return true
    }

    // if the current route is no longer within threshold of the best route, pick a new the route

    bestRouteCost := GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCost, destRelays)

    if currentRouteCost > bestRouteCost + costThreshold {
        GetRandomBestRoute(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, out_updatedRouteCost, out_updatedRouteNumRelays, out_updatedRouteRelays)
        return true
    }

    // hold current route

    *out_updatedRouteCost = currentRouteCost
    *out_updatedRouteNumRelays = currentRouteNumRelays
    copy(out_updatedRouteRelays[:], currentRouteRelays[:])

    return false
}

type RouteShader struct {
    DisableNetworkNext          bool
    SelectionPercent            int
    ABTest                      bool
    ProMode                     bool
    ReduceLatency               bool
    ReducePacketLoss            bool
    Multipath                   bool
    AcceptableLatency           int32
    LatencyThreshold            int32
    AcceptablePacketLoss        float32
    BandwidthEnvelopeUpKbps     int32
    BandwidthEnvelopeDownKbps   int32
}

func NewRouteShader() RouteShader {
    return RouteShader{
        DisableNetworkNext: false,
        SelectionPercent: 100,
        ABTest: false,
        ReduceLatency: true,
        ReducePacketLoss: true,
        Multipath: false,
        ProMode: false,
        AcceptableLatency: 25,
        LatencyThreshold: 5,
        AcceptablePacketLoss: 1.0,
        BandwidthEnvelopeUpKbps: 1024,
        BandwidthEnvelopeDownKbps: 1024,
    }
}

type RouteState struct {
    UserID uint64
    Next bool
    Veto bool
    Banned bool
    Disabled bool
    NotSelected bool
    ABTest bool
    A bool
    B bool
    ReduceLatency bool
    ReducePacketLoss bool
    ProMode bool
    Multipath bool
    LatencyWorse bool
    MultipathOverload bool
    NoRoute bool
}

type CustomerConfig struct {
    BannedUsers          map[uint64]bool
    MultipathVetoUsers   map[uint64]bool    
}

func NewCustomerConfig() CustomerConfig {
    return CustomerConfig {
        BannedUsers: make(map[uint64]bool),
        MultipathVetoUsers: make(map[uint64]bool),
    }
}

type InternalConfig struct {
    RouteSwitchThreshold int32
    MaxLatencyTradeOff int32
    RTTVeto_Default int32
    RTTVeto_PacketLoss int32
    RTTVeto_Multipath int32
    MultipathOverloadThreshold int32
}

func NewInternalConfig() InternalConfig {
    return InternalConfig{
        RouteSwitchThreshold: -5,
        MaxLatencyTradeOff: 10,
        RTTVeto_Default: -5,
        RTTVeto_PacketLoss: -20,
        RTTVeto_Multipath: -20,
        MultipathOverloadThreshold: 500,
    }
}

func EarlyOutDirect(routeShader *RouteShader, routeState *RouteState, customer *CustomerConfig) bool {

    if routeState.Veto || routeState.Banned || routeState.Disabled || routeState.NotSelected || routeState.B {
        return true
    }

    if routeShader.DisableNetworkNext {
        routeState.Disabled = true
        return true
    }

    if routeShader.SelectionPercent == 0 || (routeState.UserID % 100) > uint64(routeShader.SelectionPercent) {
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

    if customer.BannedUsers[routeState.UserID] {
        routeState.Banned = true
        return true
    }

    return false
}

func MakeRouteDecision_TakeNetworkNext(routeMatrix []RouteEntry, routeShader *RouteShader, routeState *RouteState, customer *CustomerConfig, internal *InternalConfig, directLatency int32, directPacketLoss float32, sourceRelays []int32, sourceRelayCost[]int32, destRelays []int32, out_routeCost *int32, out_routeNumRelays *int32, out_routeRelays []int32) bool {

    if routeState.Next {
        panic("only call MakeRouteDecision_TakeNetworkNext when *not* already taking network next")
    }

    if EarlyOutDirect(routeShader, routeState, customer) {
        return false
    }

    maxCost := directLatency

    // if we predict we can reduce latency, take network next

    reduceLatency := false
    if routeShader.ReduceLatency {
        if directLatency > routeShader.AcceptableLatency {
            maxCost = directLatency - routeShader.LatencyThreshold
            reduceLatency = true
        } else {
            maxCost = -1
        }
    }

    // if we predict we can reduce packet loss, take network next

    reducePacketLoss := false
    if routeShader.ReducePacketLoss && directPacketLoss > routeShader.AcceptablePacketLoss {
        maxCost = directLatency + internal.MaxLatencyTradeOff
        reducePacketLoss = true
    }

    // if we are in pro mode, take network next pro-actively in multipath before anything goes wrong

    userHasMultipathVeto := !customer.MultipathVetoUsers[routeState.UserID]

    proMode := false
    if routeShader.ProMode && !userHasMultipathVeto {
        maxCost = directLatency + internal.MaxLatencyTradeOff
        reduceLatency = true
        reducePacketLoss = true
        proMode = true
    }

    // get the initial best route

    bestRouteCost := int32(0)
    bestRouteNumRelays := int32(0)
    bestRouteRelays := [MaxRelaysPerRoute]int32{}

    GetBestRoute_Initial(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays)

    // if we don't find any network next route, we can't take network next

    if bestRouteNumRelays == 0 {
        return false
    }

    // take network next

    routeState.Next = true
    routeState.ReduceLatency = reduceLatency
    routeState.ReducePacketLoss = reducePacketLoss
    routeState.ProMode = proMode
    routeState.Multipath = routeShader.Multipath && userHasMultipathVeto

    *out_routeCost = bestRouteCost
    *out_routeNumRelays = bestRouteNumRelays
    copy(out_routeRelays, bestRouteRelays[:bestRouteNumRelays])

    return true
}

func MakeRouteDecision_StayOnNetworkNext_Internal(routeMatrix []RouteEntry, routeShader *RouteShader, routeState *RouteState, customer *CustomerConfig, internal *InternalConfig, directLatency int32, nextLatency int32, currentRouteNumRelays int32, currentRouteRelays [MaxRelaysPerRoute]int32, sourceRelays []int32, sourceRelayCost[]int32, destRelays []int32, out_updatedRouteCost *int32, out_updatedRouteNumRelays *int32, out_updatedRouteRelays []int32) bool {

    if EarlyOutDirect(routeShader, routeState, customer) {
        return false
    }

    // if we overload the connection in multipath, leave network next

    if routeState.Multipath && directLatency > internal.MultipathOverloadThreshold {
        routeState.MultipathOverload = true
        return false
    }

    // if we have made rtt significantly worse, leave network next

    rttVeto := internal.RTTVeto_Default

    if routeState.ReducePacketLoss {
        rttVeto = internal.RTTVeto_PacketLoss
    }

    if routeState.Multipath {
        rttVeto = internal.RTTVeto_Multipath
    }

    if nextLatency > directLatency + rttVeto {
        routeState.LatencyWorse = true
        return false
    }

    // update the current best route

    maxCost := directLatency + rttVeto

    bestRouteCost := int32(0)
    bestRouteNumRelays := int32(0)
    bestRouteRelays := [MaxRelaysPerRoute]int32{}

    GetBestRoute_Update(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, internal.RouteSwitchThreshold, currentRouteNumRelays, currentRouteRelays, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays)

    // if we no longer have a network next route, leave network next

    if bestRouteNumRelays == 0 {
        routeState.NoRoute = true
        return false
    }

    // stay on network next

    *out_updatedRouteCost = bestRouteCost
    *out_updatedRouteNumRelays = bestRouteNumRelays
    copy(out_updatedRouteRelays, bestRouteRelays[:bestRouteNumRelays])

    return true
}

func MakeRouteDecision_StayOnNetworkNext(routeMatrix []RouteEntry, routeShader *RouteShader, routeState *RouteState, customer *CustomerConfig, internal *InternalConfig, directLatency int32, nextLatency int32, currentRouteNumRelays int32, currentRouteRelays [MaxRelaysPerRoute]int32, sourceRelays []int32, sourceRelayCost[]int32, destRelays []int32, out_updatedRouteCost *int32, out_updatedRouteNumRelays *int32, out_updatedRouteRelays []int32) bool {

    if !routeState.Next {
        panic("only call MakeRouteDecision_TakeNetworkNext when session is on network next")
    }

    stayOnNetworkNext := MakeRouteDecision_StayOnNetworkNext_Internal(routeMatrix, routeShader, routeState, customer, internal, directLatency, nextLatency, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCost, destRelays, out_updatedRouteCost, out_updatedRouteNumRelays, out_updatedRouteRelays)

    if routeState.Next && !stayOnNetworkNext {
        routeState.Next = false
        routeState.Veto = true
    }

    return stayOnNetworkNext
}

// -------------------------------------------
