/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
    "context"
    "fmt"
    "io/ioutil"
    "math/rand"
    "net"
    "net/http"
    "os"
    "os/signal"
    "runtime"
    "sort"
    "sync"
    "syscall"
    "time"

    "github.com/gorilla/mux"

    "github.com/networknext/backend/modules/common"
    "github.com/networknext/backend/modules/core"
    "github.com/networknext/backend/modules/encoding"
    "github.com/networknext/backend/modules/packets"

    "github.com/networknext/backend/modules-old/crypto"
    "github.com/networknext/backend/modules-old/routing"
)

const NEXT_RELAY_BACKEND_PORT = 30000
const NEXT_SERVER_BACKEND_PORT = 45000

const BACKEND_MODE_FORCE_DIRECT = 1
const BACKEND_MODE_RANDOM = 2
const BACKEND_MODE_MULTIPATH = 3
const BACKEND_MODE_ON_OFF = 4
const BACKEND_MODE_ON_ON_OFF = 5
const BACKEND_MODE_ROUTE_SWITCHING = 6
const BACKEND_MODE_UNCOMMITTED = 7
const BACKEND_MODE_UNCOMMITTED_TO_COMMITTED = 8
const BACKEND_MODE_SERVER_EVENTS = 9
const BACKEND_MODE_FORCE_RETRY = 10
const BACKEND_MODE_BANDWIDTH = 11
const BACKEND_MODE_JITTER = 12
const BACKEND_MODE_TAGS = 13
const BACKEND_MODE_DIRECT_STATS = 14
const BACKEND_MODE_NEXT_STATS = 15
const BACKEND_MODE_NEAR_RELAY_STATS = 16
const BACKEND_MODE_MATCH_ID = 17
const BACKEND_MODE_MATCH_VALUES = 18

type Backend struct {
    mutex           sync.RWMutex
    dirty           bool
    mode            int
    relayManager    *common.RelayManager
    routeMatrix     *common.RouteMatrix
}

var backend Backend

type ServerEntry struct {
    address    *net.UDPAddr
    publicKey  []byte
    lastUpdate int64
}

type SessionCacheEntry struct {
    CustomerID                 uint64
    SessionID                  uint64
    UserHash                   uint64
    Sequence                   uint64
    RouteHash                  uint64
    OnNNSliceCounter           uint64
    CommitPending              bool
    CommitObservedSliceCounter uint8
    Committed                  bool
    TimestampStart             time.Time
    TimestampExpire            time.Time
    Version                    uint8
    DirectRTT                  float64
    NextRTT                    float64
    Location                   packets.SDK5_LocationData
    Response                   []byte
}

const ThresholdRTT = 1.0
const MaxRTT = float32(100)
const MaxJitter = float32(10.0)
const MaxPacketLoss = float32(0.1)

func OptimizeThread() {

    for {

        backend.mutex.Lock()

        relayIds := make([]uint64, 0)
        relayDatacenterIds := make([]uint64, 0)

        // todo: get set of relay ids and datacenter ids from relay manager

        /*
            for _, relayData := range backend.relayMap.GetAllRelayData() {
               relayIDs = append(relayIDs, relayData.ID)
               relayDatacenterIDs = append(relayDatacenterIDs, common.DatacenterId("local"))
            }
        */

        costMatrix := backend.relayManager.GetCosts(relayIds, MaxRTT, MaxJitter, MaxPacketLoss, false)

        numRelays := len(relayIds)

        numCPUs := runtime.NumCPU()
        numSegments := numRelays
        if numCPUs < numRelays {
            numSegments = numRelays / 5
            if numSegments == 0 {
                numSegments = 1
            }
        }

        destRelays := make([]bool, numRelays)
        for i := 0; i < numRelays; i++ {
            destRelays[i] = true
        }

        core.Optimize2(numRelays, numSegments, costMatrix, 1, relayDatacenterIds, destRelays)

        backend.mutex.Unlock()

        time.Sleep(1 * time.Second)
    }
}

var (
    magicMutex    sync.RWMutex
    magicUpcoming [8]byte
    magicCurrent  [8]byte
    magicPrevious [8]byte
)

func GetMagic() ([8]byte, [8]byte, [8]byte) {
    magicMutex.RLock()
    upcoming := magicUpcoming
    current := magicCurrent
    previous := magicPrevious
    magicMutex.RUnlock()
    return upcoming, current, previous
}

func GenerateMagic(magic []byte) {
    newMagic := make([]byte, 8)
    core.RandomBytes(newMagic)
    for i := range newMagic {
        magic[i] = newMagic[i]
    }
}

func UpdateMagic() {
    ticker := time.NewTicker(time.Second * 60)
    for {
        select {
        case <-ticker.C:
            magicMutex.Lock()
            magicPrevious = magicCurrent
            magicCurrent = magicUpcoming
            GenerateMagic(magicUpcoming[:])
            magicMutex.Unlock()
        }
    }
}

func (backend *Backend) GetNearRelays() []routing.RelayData {
    backend.mutex.Lock()
    // todo: fill all relay data from relay manager
    allRelayData := make([]routing.RelayData, 0)
    backend.mutex.Unlock()
    sort.SliceStable(allRelayData, func(i, j int) bool { return allRelayData[i].ID < allRelayData[j].ID })
    if len(allRelayData) > int(core.MaxNearRelays) {
        allRelayData = allRelayData[:core.MaxNearRelays]
    }
    return allRelayData
}

// -----------------------------------------------------------

const InitRequestMagic = uint32(0x9083708f)
const InitRequestVersion = 0
const UpdateRequestVersion = 5
const UpdateResponseVersion = 1
const MaxRelayAddressLength = 256
const RelayTokenBytes = 32
const MaxRelays = 5

func RelayUpdateHandler(writer http.ResponseWriter, request *http.Request) {

    // parse the relay update request

    body, err := ioutil.ReadAll(request.Body)
    if err != nil {
        fmt.Printf("could not read body\n")
        return
    }
    defer request.Body.Close()

    index := 0

    var version uint32
    if !encoding.ReadUint32(body, &index, &version) || version != UpdateRequestVersion {
        fmt.Printf("bad version\n")
        return
    }

    var relay_address string
    if !encoding.ReadString(body, &index, &relay_address, MaxRelayAddressLength) {
        fmt.Printf("address\n")
        return
    }

    var token []byte
    if !encoding.ReadBytes(body, &index, &token, RelayTokenBytes) {
        fmt.Printf("bad token\n")
        return
    }

    udpAddr, err := net.ResolveUDPAddr("udp", relay_address)
    if err != nil {
        fmt.Printf("bad resolve addr %s\n", relay_address)
        return
    }

    relay := &routing.RelayData{
        ID:             common.RelayId(relay_address),
        Addr:           *udpAddr,
        PublicKey:      token,
        LastUpdateTime: time.Now(),
    }

    var numRelays uint32
    if !encoding.ReadUint32(body, &index, &numRelays) {
        fmt.Printf("could not read num relays\n")
        return
    }

    if numRelays > MaxRelays {
        fmt.Printf("too many relays\n")
        return
    }

    statsUpdate := &routing.RelayStatsUpdate{}
    statsUpdate.ID = relay.ID

    for i := 0; i < int(numRelays); i++ {
        var id uint64
        var rtt, jitter, packetLoss float32
        if !encoding.ReadUint64(body, &index, &id) {
            fmt.Printf("bad relay id\n")
            return
        }
        if !encoding.ReadFloat32(body, &index, &rtt) {
            fmt.Printf("bad relay rtt\n")
            return
        }
        if !encoding.ReadFloat32(body, &index, &jitter) {
            fmt.Printf("bad relay jitter\n")
            return
        }
        if !encoding.ReadFloat32(body, &index, &packetLoss) {
            fmt.Printf("bad relay packet loss\n")
            return
        }
        ping := routing.RelayStatsPing{}
        ping.RelayID = id
        ping.RTT = rtt
        ping.Jitter = jitter
        ping.PacketLoss = packetLoss
        statsUpdate.PingStats = append(statsUpdate.PingStats, ping)
    }

    var sessionCount uint64
    if !encoding.ReadUint64(body, &index, &sessionCount) {
        fmt.Printf("could not read session count\n")
        return
    }

    var shutdown bool
    if !encoding.ReadBool(body, &index, &shutdown) {
        fmt.Printf("could not read shutdown\n")
        return
    }

    var relayVersion string
    if !encoding.ReadString(body, &index, &relayVersion, uint32(32)) {
        fmt.Printf("could not read relay version\n")
        return
    }

    var cpu uint8
    if !encoding.ReadUint8(body, &index, &cpu) {
        fmt.Printf("could not read cpu\n")
        return
    }

    var envelopeUpKbps uint64
    if !encoding.ReadUint64(body, &index, &envelopeUpKbps) {
        fmt.Printf("could not read envelope up kbps\n")
        return
    }

    var envelopeDownKbps uint64
    if !encoding.ReadUint64(body, &index, &envelopeDownKbps) {
        fmt.Printf("could not read envelope down kbps\n")
        return
    }

    var bandwidthSentKbps uint64
    if !encoding.ReadUint64(body, &index, &bandwidthSentKbps) {
        fmt.Printf("could not read bandwidth sent kbps\n")
        return
    }

    var bandwidthRecvKbps uint64
    if !encoding.ReadUint64(body, &index, &bandwidthRecvKbps) {
        fmt.Printf("could not read bandwidth recv kbps\n")
        return
    }

    // process the relay update

    // todo

    // ...

    // get relays to ping

    relaysToPing := make([]routing.RelayPingData, 0)

    // todo

    // ...

    // write response packet

    magicUpcoming, magicCurrent, magicPrevious := GetMagic()

    responseData := make([]byte, 10*1024)

    index = 0

    encoding.WriteUint32(responseData, &index, UpdateResponseVersion)

    encoding.WriteUint64(responseData, &index, uint64(time.Now().Unix()))

    encoding.WriteUint32(responseData, &index, uint32(len(relaysToPing)))

    for i := range relaysToPing {
        encoding.WriteUint64(responseData, &index, relaysToPing[i].ID)
        encoding.WriteString(responseData, &index, relaysToPing[i].Address, MaxRelayAddressLength)
    }

    encoding.WriteString(responseData, &index, relayVersion, uint32(32))

    encoding.WriteBytes(responseData, &index, magicUpcoming[:], 8)

    encoding.WriteBytes(responseData, &index, magicCurrent[:], 8)

    encoding.WriteBytes(responseData, &index, magicPrevious[:], 8)

    encoding.WriteUint32(responseData, &index, 0)

    responseLength := index

    responseData = responseData[:responseLength]

    writer.Header().Set("Content-Type", "application/octet-stream")

    writer.Write(responseData)
}

func StartWebServer() {
    router := mux.NewRouter()
    router.HandleFunc("/relay_update", RelayUpdateHandler).Methods("POST")
    http.ListenAndServe(fmt.Sprintf(":%d", NEXT_RELAY_BACKEND_PORT), router)
}

func StartUDPServer() {

    serverBackendAddress = *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", NEXT_SERVER_BACKEND_PORT))

    config := common.UDPServerConfig{}
    config.Port = NEXT_SERVER_BACKEND_PORT
    config.NumThreads = 2
    config.SocketReadBuffer = 1024 * 1024
    config.SocketWriteBuffer = 1024 * 1024
    config.MaxPacketSize = 4096

    common.CreateUDPServer(context.Background(), config, packetHandler)
}

// -------------------------------------------------

var serverBackendAddress net.UDPAddr

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

    fmt.Printf("received %d byte packet from %s\n", len(packetData), from.String())

    // ignore packets that are too small

    if len(packetData) < 16+3+4+packets.NEXT_CRYPTO_SIGN_BYTES+2 {
        fmt.Printf("packet is too small")
        return
    }

    // ignore packet types we don't support

    packetType := packetData[0]

    if packetType != packets.SDK5_SERVER_INIT_REQUEST_PACKET && packetType != packets.SDK5_SERVER_UPDATE_REQUEST_PACKET && packetType != packets.SDK5_SESSION_UPDATE_REQUEST_PACKET && packetType != packets.SDK5_MATCH_DATA_REQUEST_PACKET {
        fmt.Printf("unsupported packet type %d", packetType)
        return
    }

    // check basic packet filetr

    if !core.BasicPacketFilter(packetData, len(packetData)) {
        fmt.Printf("basic packet filter failed\n")
        return
    }

    // check advanced packet filter

    var emptyMagic [8]byte

    var fromAddressBuffer [32]byte
    var toAddressBuffer [32]byte

    fromAddressData, fromAddressPort := core.GetAddressData(from, fromAddressBuffer[:])
    toAddressData, toAddressPort := core.GetAddressData(&serverBackendAddress, toAddressBuffer[:])

    if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(packetData)) {
        fmt.Printf("advanced packet filter failed\n")
        return
    }

    // todo: check buyer id and packet signature

    // ...

    // process the packet according to type

    packetData = packetData[16:]

    switch packetType {

    case packets.SDK5_SERVER_INIT_REQUEST_PACKET:
        packet := packets.SDK5_ServerInitRequestPacket{}
        if err := packets.ReadPacket(packetData, &packet); err != nil {
            core.Error("could not read server init request packet from %s: %v", from.String(), err)
            return
        }
        ProcessServerInitRequestPacket(conn, from, &packet)
        break

    case packets.SDK5_SERVER_UPDATE_REQUEST_PACKET:
        packet := packets.SDK5_ServerUpdateRequestPacket{}
        if err := packets.ReadPacket(packetData, &packet); err != nil {
            core.Error("could not read server update request packet from %s", from.String())
            return
        }
        ProcessServerUpdateRequestPacket(conn, from, &packet)
        break

    case packets.SDK5_SESSION_UPDATE_REQUEST_PACKET:
        packet := packets.SDK5_SessionUpdateRequestPacket{}
        if err := packets.ReadPacket(packetData, &packet); err != nil {
            core.Error("could not read session update request packet from %s", from.String())
            return
        }
        ProcessSessionUpdateRequestPacket(conn, from, &packet)
        break

    case packets.SDK5_MATCH_DATA_REQUEST_PACKET:
        packet := packets.SDK5_MatchDataRequestPacket{}
        if err := packets.ReadPacket(packetData, &packet); err != nil {
            core.Error("could not read match data request packet from %s", from.String())
            return
        }
        ProcessMatchDataRequestPacket(conn, from, &packet)
        break

    default:
        panic("unknown packet type")
    }
}

func SendResponsePacket[P packets.Packet](conn *net.UDPConn, to *net.UDPAddr, packetType int, packet P) {

    // todo: we should not be running the func test with the real backend private key...

    packetData, err := packets.SDK5_WritePacket(packet, packetType, 4096, &serverBackendAddress, to, crypto.BackendPrivateKey)
    if err != nil {
        core.Error("failed to write response packet: %v", err)
        return
    }

    if _, err := conn.WriteToUDP(packetData, to); err != nil {
        core.Error("failed to send response packet: %v", err)
        return
    }
}

func ProcessServerInitRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_ServerInitRequestPacket) {

    fmt.Printf("server init request\n")

    responsePacket := &packets.SDK5_ServerInitResponsePacket{
        RequestId: requestPacket.RequestId,
        Response:  packets.SDK5_ServerInitResponseOK,
    }

    responsePacket.UpcomingMagic, responsePacket.CurrentMagic, responsePacket.PreviousMagic = GetMagic()

    SendResponsePacket(conn, from, packets.SDK5_SERVER_INIT_RESPONSE_PACKET, responsePacket)
}

func ProcessServerUpdateRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_ServerUpdateRequestPacket) {

    fmt.Printf("server update request\n")

    responsePacket := &packets.SDK5_ServerUpdateResponsePacket{
        RequestId: requestPacket.RequestId,
    }

    responsePacket.UpcomingMagic, responsePacket.CurrentMagic, responsePacket.PreviousMagic = GetMagic()

    SendResponsePacket(conn, from, packets.SDK5_SERVER_UPDATE_RESPONSE_PACKET, responsePacket)
}

func ProcessSessionUpdateRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_SessionUpdateRequestPacket) {

    fmt.Printf("server init request\n")

    if backend.mode == BACKEND_MODE_FORCE_RETRY && requestPacket.RetryNumber < 4 {
        return
    }

    if requestPacket.PlatformType == packets.SDK5_PlatformTypeUnknown {
        panic("platform type is unknown")
    }

    if requestPacket.ConnectionType == packets.SDK5_ConnectionTypeUnknown {
        panic("connection type is unknown")
    }

    if requestPacket.FallbackToDirect {
        fmt.Printf("error: fallback to direct %s\n", from.String())
        return
    }

    if requestPacket.Reported {
        fmt.Printf("client reported session\n")
    }

    if requestPacket.ClientPingTimedOut {
        fmt.Printf("client ping timed out\n")
    }

    if requestPacket.ClientBandwidthOverLimit {
        fmt.Printf("client bandwidth over limit\n")
    }

    if requestPacket.ServerBandwidthOverLimit {
        fmt.Printf("server bandwidth over limit\n")
    }

    if requestPacket.PacketsLostClientToServer > 0 {
        fmt.Printf("%d client to server packets lost\n", requestPacket.PacketsLostClientToServer)
    }

    if requestPacket.PacketsLostServerToClient > 0 {
        fmt.Printf("%d server to client packets lost\n", requestPacket.PacketsLostServerToClient)
    }

    if backend.mode == BACKEND_MODE_BANDWIDTH {
        if requestPacket.NextKbpsUp > 0 {
            fmt.Printf("%d kbps up\n", requestPacket.NextKbpsUp)
        }
        if requestPacket.NextKbpsDown > 0 {
            fmt.Printf("%d kbps down\n", requestPacket.NextKbpsDown)
        }
    }

    if backend.mode == BACKEND_MODE_JITTER {
        if requestPacket.JitterClientToServer > 0 {
            fmt.Printf("%f jitter up\n", requestPacket.JitterClientToServer)
            if requestPacket.JitterClientToServer > 100 {
                panic("jitter up too high")
            }
        }
        if requestPacket.JitterServerToClient > 0 {
            fmt.Printf("%f jitter down\n", requestPacket.JitterServerToClient)
            if requestPacket.JitterServerToClient > 100 {
                panic("jitter down too high")
            }
        }
    }

    if backend.mode == BACKEND_MODE_TAGS {
        if requestPacket.NumTags > 0 {
            for i := 0; i < int(requestPacket.NumTags); i++ {
                fmt.Printf("tag %x\n", requestPacket.Tags[i])
            }
        } else {
            fmt.Printf("tag cleared\n")
        }
    }

    if backend.mode == BACKEND_MODE_DIRECT_STATS {
        if requestPacket.DirectMinRTT > 0 && requestPacket.DirectJitter > 0 && requestPacket.DirectPacketLoss > 0 {
            fmt.Printf("direct rtt = %f, direct jitter = %f, direct packet loss = %f\n", requestPacket.DirectMinRTT, requestPacket.DirectJitter, requestPacket.DirectPacketLoss)
        }
    }

    if backend.mode == BACKEND_MODE_NEXT_STATS {
        if requestPacket.NextRTT > 0 && requestPacket.NextJitter > 0 && requestPacket.NextPacketLoss > 0 {
            fmt.Printf("next rtt = %f, next jitter = %f, next packet loss = %f\n", requestPacket.NextRTT, requestPacket.NextJitter, requestPacket.NextPacketLoss)
        }
    }

    if backend.mode == BACKEND_MODE_NEAR_RELAY_STATS {
        for i := 0; i <= int(requestPacket.NumNearRelays); i++ {
            fmt.Printf("near relay: id = %x, rtt = %d, jitter = %d, packet loss = %d\n", requestPacket.NearRelayIds[i], requestPacket.NearRelayRTT[i], requestPacket.NearRelayJitter[i], requestPacket.NearRelayPacketLoss[i])
        }
    }

    // read the session data

    newSession := requestPacket.SliceNumber == 0

    var sessionData packets.SDK5_SessionData

    if newSession {

        sessionData.Version = packets.SDK5_SessionDataVersion
        sessionData.SessionId = requestPacket.SessionId
        sessionData.SliceNumber = uint32(requestPacket.SliceNumber + 1)
        sessionData.ExpireTimestamp = uint64(time.Now().Unix()) + packets.SDK5_BillingSliceSeconds
        sessionData.RouteState.UserID = requestPacket.UserHash

    } else {

        readStream := encoding.CreateReadStream(requestPacket.SessionData[:])

        err := sessionData.Serialize(readStream)
        if err != nil {
            fmt.Printf("could not read session data in session update packet: %v\n", err)
            return
        }

        sessionData.SliceNumber = uint32(requestPacket.SliceNumber + 1)
        sessionData.ExpireTimestamp += packets.SDK5_BillingSliceSeconds
    }

    // decide if we should take network next or not

    nearRelays := backend.GetNearRelays()

    takeNetworkNext := len(nearRelays) > 0

    if backend.mode == BACKEND_MODE_FORCE_DIRECT {
        takeNetworkNext = false
    }

    if backend.mode == BACKEND_MODE_RANDOM {
        takeNetworkNext = takeNetworkNext && rand.Float32() > 0.5
    }

    if backend.mode == BACKEND_MODE_ON_OFF {
        // Alternate between direct and next routes for each slice
        if (requestPacket.SliceNumber & 1) == 0 {
            takeNetworkNext = false
        }
    }

    if backend.mode == BACKEND_MODE_ON_ON_OFF {
        // Alternate between direct, a new route token and a continue route token for every 3 slices
        if (requestPacket.SliceNumber & 2) == 0 {
            takeNetworkNext = false
        }
    }

    if backend.mode == BACKEND_MODE_ROUTE_SWITCHING {
        rand.Shuffle(len(nearRelays), func(i, j int) {
            nearRelays[i], nearRelays[j] = nearRelays[j], nearRelays[i]
        })
    }

    // run various checks and prints for special func test modes

    multipath := len(nearRelays) > 0 && backend.mode == BACKEND_MODE_MULTIPATH

    committed := true

    if backend.mode == BACKEND_MODE_UNCOMMITTED {
        committed = false
        if requestPacket.Committed {
            panic("slices must not be committed in this mode")
        }
    }

    if backend.mode == BACKEND_MODE_UNCOMMITTED_TO_COMMITTED {
        committed = requestPacket.SliceNumber > 2
        if requestPacket.SliceNumber <= 2 && requestPacket.Committed {
            panic("slices 0,1,2,3 should not be committed")
        }
        if requestPacket.SliceNumber >= 4 && !requestPacket.Committed {
            panic("slices 4 and greater should be committed")
        }
    }

    if backend.mode == BACKEND_MODE_SERVER_EVENTS {
        if requestPacket.SliceNumber >= 2 && requestPacket.UserFlags != 0x123 {
            panic("server events not set on session update")
        }
    }

    if requestPacket.UserFlags > 0 {
        fmt.Printf("server events %x\n", requestPacket.UserFlags)
    }

    // todo
    _ = multipath
    _ = committed

    // build response packet

    var responsePacket *packets.SDK5_SessionUpdateResponsePacket

    var nearRelayIds = [MaxRelays]uint64{}
    var nearRelayAddresses = [MaxRelays]net.UDPAddr{}
    var nearRelayPublicKeys = [MaxRelays][]byte{}
    for i, relay := range nearRelays {
        nearRelayIds[i] = relay.ID
        nearRelayAddresses[i] = relay.Addr
        nearRelayPublicKeys[i] = relay.PublicKey
    }

    if !takeNetworkNext {

        // direct route

        responsePacket = &packets.SDK5_SessionUpdateResponsePacket{
            SessionId:     requestPacket.SessionId,
            SliceNumber:   requestPacket.SliceNumber,
            NumNearRelays: int32(len(nearRelays)),
            // todo: need to fill out near relay ids etc.
            // NearRelayIds:       nearRelayIds[:len(nearRelays)],
            // NearRelayAddresses: nearRelayAddresses[:len(nearRelays)],
            RouteType:          int32(packets.SDK5_RouteTypeDirect),
            NumTokens:          0,
            Tokens:             nil,
            HighFrequencyPings: true,
        }

    } else {

        // todo: get processing for when on next
        /*
            // next

            numRelays := len(nearRelays)
            if numRelays > core.MaxRelaysPerRoute {
               numRelays = core.MaxRelaysPerRoute
            }
            nextRoute := routing.Route{
               NumRelays:       numRelays,
               RelayIDs:        nearRelayIDs,
               RelayAddrs:      nearRelayAddresses,
               RelayPublicKeys: nearRelayPublicKeys,
            }

            relayTokens := make([]routing.RelayToken, nextRoute.NumRelays)
            for i := range relayTokens {
               relayTokens[i] = routing.RelayToken{
                   ID:        nextRoute.RelayIDs[i],
                   Addr:      nextRoute.RelayAddrs[i],
                   PublicKey: nextRoute.RelayPublicKeys[i],
               }
            }

            var routeType int32
            sameRoute := nextRoute.NumRelays == int(sessionData.RouteNumRelays) && nextRoute.RelayIDs == sessionData.RouteRelayIDs

            routerPrivateKey := [crypto.KeySize]byte{}
            copy(routerPrivateKey[:], crypto.RouterPrivateKey)

            tokenAddresses := make([]*net.UDPAddr, nextRoute.NumRelays+2)
            tokenAddresses[0] = &sessionUpdate.ClientAddress
            tokenAddresses[len(tokenAddresses)-1] = &sessionUpdate.ServerAddress
            for i := 0; i < nextRoute.NumRelays; i++ {
               tokenAddresses[1+i] = &nearRelayAddresses[i]
            }

            tokenPublicKeys := make([][]byte, nextRoute.NumRelays+2)
            tokenPublicKeys[0] = sessionUpdate.ClientRoutePublicKey
            tokenPublicKeys[len(tokenPublicKeys)-1] = sessionUpdate.ServerRoutePublicKey
            for i := 0; i < nextRoute.NumRelays; i++ {
               tokenPublicKeys[1+i] = nearRelayPublicKeys[i]
            }

            numTokens := nextRoute.NumRelays + 2

            var tokenData []byte
            if sameRoute {
               tokenData = make([]byte, numTokens*routing.EncryptedContinueRouteTokenSize)
               core.WriteContinueTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionID, uint8(sessionData.SessionVersion), int(numTokens), nextRoute.RelayPublicKeys[:], routerPrivateKey)
               routeType = routing.RouteTypeContinue
            } else {
               sessionData.ExpireTimestamp += billing.BillingSliceSeconds
               sessionData.SessionVersion++

               tokenData = make([]byte, numTokens*routing.EncryptedNextRouteTokenSize)
               core.WriteRouteTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionID, uint8(sessionData.SessionVersion), 256, 256, int(numTokens), tokenAddresses, tokenPublicKeys, routerPrivateKey)
               routeType = routing.RouteTypeNew
            }

            sessionResponse = &transport.SessionResponsePacketSDK5{
                SessionID:          sessionUpdate.SessionID,
                SliceNumber:        sessionUpdate.SliceNumber,
                NumNearRelays:      int32(len(nearRelays)),
                NearRelayIDs:       nearRelayIDs[:len(nearRelays)],
                NearRelayAddresses: nearRelayAddresses[:len(nearRelays)],
                RouteType:          routeType,
                Multipath:          multipath,
                Committed:          committed,
                NumTokens:          int32(numTokens),
                Tokens:             tokenData,
                HighFrequencyPings: true,
            }
        */
    }

    if responsePacket == nil {
        fmt.Printf("error: nil session response\n")
        return
    }

    responsePacket.Version = requestPacket.Version

    excludeNearRelays(responsePacket, sessionData.RouteState)

    // todo: write session data to buffer

    // todo: stash session data in response packet

    SendResponsePacket(conn, from, packets.SDK5_SESSION_UPDATE_RESPONSE_PACKET, responsePacket)

    /*
        sessionDataBuffer, err := transport.MarshalSessionDataSDK5(&sessionData)
        if err != nil {
           fmt.Printf("error: failed to marshal session data: %v\n", err)
           return
        }

        if len(sessionDataBuffer) > transport.MaxSessionDataSize {
           fmt.Printf("session data too large\n")
        }

        sessionResponse.SessionDataBytes = int32(len(sessionDataBuffer))
        copy(sessionResponse.SessionData[:], sessionDataBuffer)
    */
}

func ProcessMatchDataRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_MatchDataRequestPacket) {

    fmt.Printf("server match data request\n")

    if backend.mode == BACKEND_MODE_FORCE_RETRY && requestPacket.RetryNumber < 4 {
        fmt.Printf("force retry for match data request packet\n")
        return
    }

    if backend.mode == BACKEND_MODE_MATCH_ID || backend.mode == BACKEND_MODE_FORCE_RETRY {
        fmt.Printf("match id %x\n", requestPacket.MatchId)
    }

    if backend.mode == BACKEND_MODE_MATCH_VALUES || backend.mode == BACKEND_MODE_FORCE_RETRY {
        for i := 0; i < int(requestPacket.NumMatchValues); i++ {
            fmt.Printf("match value %.2f\n", requestPacket.MatchValues[i])
        }
    }

    responsePacket := &packets.SDK5_MatchDataResponsePacket{
        SessionId: requestPacket.SessionId,
    }

    SendResponsePacket(conn, from, packets.SDK5_MATCH_DATA_RESPONSE_PACKET, responsePacket)
}

// todo: incorporate
/*
   // check packet signature

   var buyerId uint64
   index := 16 + 3
   encoding.ReadUint64(packetData, &index, &buyerId)

   buyer, ok := handler.Database.BuyerMap[buyerId]
   if !ok {
       core.Error("unknown buyer id: %016x", buyerId)
       handler.Events[SDK5_HandlerEvent_UnknownBuyer] = true
       return
   }

   publicKey := buyer.PublicKey

   if !SDK5_CheckPacketSignature(packetData, publicKey) {
       core.Debug("packet signature check failed")
       handler.Events[SDK5_HandlerEvent_SignatureCheckFailed] = true
       return
   }
*/

func excludeNearRelays(sessionResponse *packets.SDK5_SessionUpdateResponsePacket, routeState core.RouteState) {
    numExcluded := 0
    for i := 0; i < int(routeState.NumNearRelays); i++ {
        if routeState.NearRelayRTT[i] == 255 {
            sessionResponse.NearRelayExcluded[i] = true
        }
    }
    sessionResponse.ExcludeNearRelays = numExcluded > 0
}

// -----------------------------------------------

func main() {

    rand.Seed(time.Now().UnixNano())

    backend.relayManager = common.CreateRelayManager()
    
    backend.routeMatrix = &common.RouteMatrix{}

    if os.Getenv("BACKEND_MODE") == "FORCE_DIRECT" {
        backend.mode = BACKEND_MODE_FORCE_DIRECT
    }

    if os.Getenv("BACKEND_MODE") == "RANDOM" {
        backend.mode = BACKEND_MODE_RANDOM
    }

    if os.Getenv("BACKEND_MODE") == "MULTIPATH" {
        backend.mode = BACKEND_MODE_MULTIPATH
    }

    if os.Getenv("BACKEND_MODE") == "ON_OFF" {
        backend.mode = BACKEND_MODE_ON_OFF
    }

    if os.Getenv("BACKEND_MODE") == "ON_ON_OFF" {
        backend.mode = BACKEND_MODE_ON_ON_OFF
    }

    if os.Getenv("BACKEND_MODE") == "ROUTE_SWITCHING" {
        backend.mode = BACKEND_MODE_ROUTE_SWITCHING
    }

    if os.Getenv("BACKEND_MODE") == "UNCOMMITTED" {
        backend.mode = BACKEND_MODE_UNCOMMITTED
    }

    if os.Getenv("BACKEND_MODE") == "UNCOMMITTED_TO_COMMITTED" {
        backend.mode = BACKEND_MODE_UNCOMMITTED_TO_COMMITTED
    }

    if os.Getenv("BACKEND_MODE") == "SERVER_EVENTS" {
        backend.mode = BACKEND_MODE_SERVER_EVENTS
    }

    if os.Getenv("BACKEND_MODE") == "FORCE_RETRY" {
        backend.mode = BACKEND_MODE_FORCE_RETRY
    }

    if os.Getenv("BACKEND_MODE") == "BANDWIDTH" {
        backend.mode = BACKEND_MODE_BANDWIDTH
    }

    if os.Getenv("BACKEND_MODE") == "JITTER" {
        backend.mode = BACKEND_MODE_JITTER
    }

    if os.Getenv("BACKEND_MODE") == "TAGS" {
        backend.mode = BACKEND_MODE_TAGS
    }

    if os.Getenv("BACKEND_MODE") == "DIRECT_STATS" {
        backend.mode = BACKEND_MODE_DIRECT_STATS
    }

    if os.Getenv("BACKEND_MODE") == "NEXT_STATS" {
        backend.mode = BACKEND_MODE_NEXT_STATS
    }

    if os.Getenv("BACKEND_MODE") == "MATCH_ID" {
        backend.mode = BACKEND_MODE_MATCH_ID
    }

    if os.Getenv("BACKEND_MODE") == "MATCH_VALUES" {
        backend.mode = BACKEND_MODE_MATCH_VALUES
    }

    GenerateMagic(magicUpcoming[:])
    GenerateMagic(magicCurrent[:])
    GenerateMagic(magicPrevious[:])

    go OptimizeThread()

    go UpdateMagic()

    go StartWebServer()

    go StartUDPServer()

    fmt.Printf("started functional backend on ports %d and %d (sdk5)\n", NEXT_RELAY_BACKEND_PORT, NEXT_SERVER_BACKEND_PORT)

    // Wait for shutdown signal
    termChan := make(chan os.Signal, 1)
    signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

    select {
    case <-termChan:
        core.Debug("received shutdown signal")
        return
    }
}
