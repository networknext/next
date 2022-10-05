/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
    // "bytes"
    "context"
    "encoding/binary"
    "fmt"
    "io/ioutil"
    "math"
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
    "github.com/networknext/backend/modules/packets"
    "github.com/networknext/backend/modules/handlers"

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
    serverDatabase  map[string]ServerEntry
    sessionDatabase map[uint64]SessionCacheEntry
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

// todo: these should not be duplicated here

func ReadBool(data []byte, index *int, value *bool) bool {
    if *index+1 > len(data) {
        return false
    }

    if data[*index] > 0 {
        *value = true
    } else {
        *value = false
    }

    *index += 1
    return true
}

func ReadUint8(data []byte, index *int, value *uint8) bool {
    if *index+1 > len(data) {
        return false
    }
    *value = data[*index]
    *index += 1
    return true
}

func ReadUint32(data []byte, index *int, value *uint32) bool {
    if *index+4 > len(data) {
        return false
    }
    *value = binary.LittleEndian.Uint32(data[*index:])
    *index += 4
    return true
}

func ReadUint64(data []byte, index *int, value *uint64) bool {
    if *index+8 > len(data) {
        return false
    }
    *value = binary.LittleEndian.Uint64(data[*index:])
    *index += 8
    return true
}

func ReadFloat32(data []byte, index *int, value *float32) bool {
    var int_value uint32
    if !ReadUint32(data, index, &int_value) {
        return false
    }
    *value = math.Float32frombits(int_value)
    return true
}

func ReadString(data []byte, index *int, value *string, maxStringLength uint32) bool {
    var stringLength uint32
    if !ReadUint32(data, index, &stringLength) {
        return false
    }
    if stringLength > maxStringLength {
        return false
    }
    if *index+int(stringLength) > len(data) {
        return false
    }
    stringData := make([]byte, stringLength)
    for i := uint32(0); i < stringLength; i++ {
        stringData[i] = data[*index]
        *index += 1
    }
    *value = string(stringData)
    return true
}

func ReadBytes(data []byte, index *int, value *[]byte, bytes uint32) bool {
    if *index+int(bytes) > len(data) {
        return false
    }
    *value = make([]byte, bytes)
    for i := uint32(0); i < bytes; i++ {
        (*value)[i] = data[*index]
        *index += 1
    }
    return true
}

func WriteUint32(data []byte, index *int, value uint32) {
    binary.LittleEndian.PutUint32(data[*index:], value)
    *index += 4
}

func WriteUint64(data []byte, index *int, value uint64) {
    binary.LittleEndian.PutUint64(data[*index:], value)
    *index += 8
}

func WriteString(data []byte, index *int, value string, maxStringLength uint32) {
    stringLength := uint32(len(value))
    if stringLength > maxStringLength {
        panic("string is too long!\n")
    }
    binary.LittleEndian.PutUint32(data[*index:], stringLength)
    *index += 4
    for i := 0; i < int(stringLength); i++ {
        data[*index] = value[i]
        *index++
    }
}

func WriteBytes(data []byte, index *int, value []byte, numBytes int) {
    for i := 0; i < numBytes; i++ {
        data[*index] = value[i]
        *index++
    }
}

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
    if !ReadUint32(body, &index, &version) || version != UpdateRequestVersion {
        fmt.Printf("bad version\n")
        return
    }

    var relay_address string
    if !ReadString(body, &index, &relay_address, MaxRelayAddressLength) {
        fmt.Printf("address\n")
        return
    }

    var token []byte
    if !ReadBytes(body, &index, &token, RelayTokenBytes) {
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
    if !ReadUint32(body, &index, &numRelays) {
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
        if !ReadUint64(body, &index, &id) {
            fmt.Printf("bad relay id\n")
            return
        }
        if !ReadFloat32(body, &index, &rtt) {
            fmt.Printf("bad relay rtt\n")
            return
        }
        if !ReadFloat32(body, &index, &jitter) {
            fmt.Printf("bad relay jitter\n")
            return
        }
        if !ReadFloat32(body, &index, &packetLoss) {
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
    if !ReadUint64(body, &index, &sessionCount) {
        fmt.Printf("could not read session count\n")
        return
    }

    var shutdown bool
    if !ReadBool(body, &index, &shutdown) {
        fmt.Printf("could not read shutdown\n")
        return
    }

    var relayVersion string
    if !ReadString(body, &index, &relayVersion, uint32(32)) {
        fmt.Printf("could not read relay version\n")
        return
    }

    var cpu uint8
    if !ReadUint8(body, &index, &cpu) {
        fmt.Printf("could not read cpu\n")
        return
    }

    var envelopeUpKbps uint64
    if !ReadUint64(body, &index, &envelopeUpKbps) {
        fmt.Printf("could not read envelope up kbps\n")
        return
    }

    var envelopeDownKbps uint64
    if !ReadUint64(body, &index, &envelopeDownKbps) {
        fmt.Printf("could not read envelope down kbps\n")
        return
    }

    var bandwidthSentKbps uint64
    if !ReadUint64(body, &index, &bandwidthSentKbps) {
        fmt.Printf("could not read bandwidth sent kbps\n")
        return
    }

    var bandwidthRecvKbps uint64
    if !ReadUint64(body, &index, &bandwidthRecvKbps) {
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

    WriteUint32(responseData, &index, UpdateResponseVersion)

    WriteUint64(responseData, &index, uint64(time.Now().Unix()))

    WriteUint32(responseData, &index, uint32(len(relaysToPing)))

    for i := range relaysToPing {
        WriteUint64(responseData, &index, relaysToPing[i].ID)
        WriteString(responseData, &index, relaysToPing[i].Address, MaxRelayAddressLength)
    }

    WriteString(responseData, &index, relayVersion, uint32(32))

    WriteBytes(responseData, &index, magicUpcoming[:], 8)

    WriteBytes(responseData, &index, magicCurrent[:], 8)

    WriteBytes(responseData, &index, magicPrevious[:], 8)

    WriteUint32(responseData, &index, 0)

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

	// todo: sdk5 write packet should move into "packets" module

	// todo: we should not be running the func test with the real backend private key...

    packetData, err := handlers.SDK5_WritePacket(packet, packetType, 4096, &serverBackendAddress, to, crypto.BackendPrivateKey)
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
	
	// todo
	
	fmt.Printf("server init request\n")
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

// todo
/*
func excludeNearRelays(sessionResponse *transport.SessionResponsePacketSDK5, routeState core.RouteState) {

    numExcluded := 0

    for i := 0; i < int(routeState.NumNearRelays); i++ {
        if routeState.NearRelayRTT[i] == 255 {
            sessionResponse.NearRelayExcluded[i] = true
        }
    }

    sessionResponse.ExcludeNearRelays = numExcluded > 0
}
*/

// todo
/*
func SessionUpdateHandlerFunc(w io.Writer, incoming *transport.UDPPacket) {

	var sessionUpdate transport.SessionUpdatePacketSDK5
	if err := transport.UnmarshalPacketSDK5(&sessionUpdate, core.GetPacketDataSDK5(incoming.Data)); err != nil {
	   fmt.Printf("error: failed to read session update packet: %v\n", err)
	   return
	}

	if backend.mode == BACKEND_MODE_FORCE_RETRY && sessionUpdate.RetryNumber < 4 {
	   return
	}

	if sessionUpdate.PlatformType == transport.PlatformTypeUnknown {
	   panic("platform type is unknown")
	}

	if sessionUpdate.ConnectionType == transport.ConnectionTypeUnknown {
	   panic("connection type is unknown")
	}

	if sessionUpdate.FallbackToDirect {
	   fmt.Printf("error: fallback to direct %s\n", incoming.From.String())
	   return
	}

	if sessionUpdate.Reported {
	   fmt.Printf("client reported session\n")
	}

	if sessionUpdate.ClientPingTimedOut {
	   fmt.Printf("client ping timed out\n")
	}

	if sessionUpdate.ClientBandwidthOverLimit {
	   fmt.Printf("client bandwidth over limit\n")
	}

	if sessionUpdate.ServerBandwidthOverLimit {
	   fmt.Printf("server bandwidth over limit\n")
	}

	if sessionUpdate.PacketsLostClientToServer > 0 {
	   fmt.Printf("%d client to server packets lost\n", sessionUpdate.PacketsLostClientToServer)
	}

	if sessionUpdate.PacketsLostServerToClient > 0 {
	   fmt.Printf("%d server to client packets lost\n", sessionUpdate.PacketsLostServerToClient)
	}

	if backend.mode == BACKEND_MODE_BANDWIDTH {
	   if sessionUpdate.NextKbpsUp > 0 {
	       fmt.Printf("%d kbps up\n", sessionUpdate.NextKbpsUp)
	   }
	   if sessionUpdate.NextKbpsDown > 0 {
	       fmt.Printf("%d kbps down\n", sessionUpdate.NextKbpsDown)
	   }
	}

	if backend.mode == BACKEND_MODE_JITTER {
	   if sessionUpdate.JitterClientToServer > 0 {
	       fmt.Printf("%f jitter up\n", sessionUpdate.JitterClientToServer)
	       if sessionUpdate.JitterClientToServer > 100 {
	           panic("jitter up too high")
	       }
	   }
	   if sessionUpdate.JitterServerToClient > 0 {
	       fmt.Printf("%f jitter down\n", sessionUpdate.JitterServerToClient)
	       if sessionUpdate.JitterServerToClient > 100 {
	           panic("jitter down too high")
	       }
	   }
	}

	if backend.mode == BACKEND_MODE_TAGS {
	   if sessionUpdate.NumTags > 0 {
	       for i := 0; i < int(sessionUpdate.NumTags); i++ {
	           fmt.Printf("tag %x\n", sessionUpdate.Tags[i])
	       }
	   } else {
	       fmt.Printf("tag cleared\n")
	   }
	}

	if backend.mode == BACKEND_MODE_DIRECT_STATS {
	   if sessionUpdate.DirectMinRTT > 0 && sessionUpdate.DirectJitter > 0 && sessionUpdate.DirectPacketLoss > 0 {
	       fmt.Printf("direct rtt = %f, direct jitter = %f, direct packet loss = %f\n", sessionUpdate.DirectMinRTT, sessionUpdate.DirectJitter, sessionUpdate.DirectPacketLoss)
	   }
	}

	if backend.mode == BACKEND_MODE_NEXT_STATS {
	   if sessionUpdate.NextRTT > 0 && sessionUpdate.NextJitter > 0 && sessionUpdate.NextPacketLoss > 0 {
	       fmt.Printf("next rtt = %f, next jitter = %f, next packet loss = %f\n", sessionUpdate.NextRTT, sessionUpdate.NextJitter, sessionUpdate.NextPacketLoss)
	   }
	}

	if backend.mode == BACKEND_MODE_NEAR_RELAY_STATS {
	   for i := 0; i <= int(sessionUpdate.NumNearRelays); i++ {
	       fmt.Printf("near relay: id = %x, rtt = %d, jitter = %d, packet loss = %d\n", sessionUpdate.NearRelayIDs[i], sessionUpdate.NearRelayRTT[i], sessionUpdate.NearRelayJitter[i], sessionUpdate.NearRelayPacketLoss[i])
	   }
	}

	newSession := sessionUpdate.SliceNumber == 0

	var sessionData transport.SessionDataSDK5
	if newSession {
	   sessionData.Version = transport.SessionDataVersionSDK5
	   sessionData.SessionID = sessionUpdate.SessionID
	   sessionData.SliceNumber = uint32(sessionUpdate.SliceNumber + 1)
	   sessionData.ExpireTimestamp = uint64(time.Now().Unix()) + billing.BillingSliceSeconds
	   sessionData.RouteState.UserID = sessionUpdate.UserHash
	   sessionData.Location = routing.LocationNullIsland
	} else {
	   if err := transport.UnmarshalSessionDataSDK5(&sessionData, sessionUpdate.SessionData[:]); err != nil {
	       fmt.Printf("could not read session data in session update packet: %v\n", err)
	       return
	   }

	   sessionData.SliceNumber = uint32(sessionUpdate.SliceNumber + 1)
	   sessionData.ExpireTimestamp += billing.BillingSliceSeconds
	}

	nearRelays := backend.GetNearRelays()

	var sessionResponse *transport.SessionResponsePacketSDK5

	takeNetworkNext := len(nearRelays) > 0

	if backend.mode == BACKEND_MODE_FORCE_DIRECT {
	   takeNetworkNext = false
	}

	if backend.mode == BACKEND_MODE_RANDOM {
	   takeNetworkNext = takeNetworkNext && rand.Float32() > 0.5
	}

	if backend.mode == BACKEND_MODE_ON_OFF {
	   // Alternate between direct and next routes for each slice
	   if (sessionUpdate.SliceNumber & 1) == 0 {
	       takeNetworkNext = false
	   }
	}

	if backend.mode == BACKEND_MODE_ON_ON_OFF {
	   // Alternate between direct, a new route token and a continue route token for every 3 slices
	   if (sessionUpdate.SliceNumber & 2) == 0 {
	       takeNetworkNext = false
	   }
	}

	if backend.mode == BACKEND_MODE_ROUTE_SWITCHING {
	   rand.Shuffle(len(nearRelays), func(i, j int) {
	       nearRelays[i], nearRelays[j] = nearRelays[j], nearRelays[i]
	   })
	}

	multipath := len(nearRelays) > 0 && backend.mode == BACKEND_MODE_MULTIPATH

	committed := true

	if backend.mode == BACKEND_MODE_UNCOMMITTED {
	   committed = false
	   if sessionUpdate.Committed {
	       panic("slices must not be committed in this mode")
	   }
	}

	if backend.mode == BACKEND_MODE_UNCOMMITTED_TO_COMMITTED {
	   committed = sessionUpdate.SliceNumber > 2
	   if sessionUpdate.SliceNumber <= 2 && sessionUpdate.Committed {
	       panic("slices 0,1,2,3 should not be committed")
	   }
	   if sessionUpdate.SliceNumber >= 4 && !sessionUpdate.Committed {
	       panic("slices 4 and greater should be committed")
	   }
	}

	if backend.mode == BACKEND_MODE_SERVER_EVENTS {
	   if sessionUpdate.SliceNumber >= 2 && sessionUpdate.ServerEvents != 0x123 {
	       panic("server flags not set on session update")
	   }
	}

	if sessionUpdate.ServerEvents > 0 {
	   fmt.Printf("server events %x\n", sessionUpdate.ServerEvents)
	}

	// Extract ids and addresses into own list to make response
	var nearRelayIDs = [MaxRelays]uint64{}
	var nearRelayAddresses = [MaxRelays]net.UDPAddr{}
	var nearRelayPublicKeys = [MaxRelays][]byte{}
	for i, relay := range nearRelays {
	   nearRelayIDs[i] = relay.ID
	   nearRelayAddresses[i] = relay.Addr
	   nearRelayPublicKeys[i] = relay.PublicKey
	}

	if !takeNetworkNext {

	   // direct route
	   sessionResponse = &transport.SessionResponsePacketSDK5{
	       SessionID:          sessionUpdate.SessionID,
	       SliceNumber:        sessionUpdate.SliceNumber,
	       NumNearRelays:      int32(len(nearRelays)),
	       NearRelayIDs:       nearRelayIDs[:len(nearRelays)],
	       NearRelayAddresses: nearRelayAddresses[:len(nearRelays)],
	       RouteType:          int32(routing.RouteTypeDirect),
	       NumTokens:          0,
	       Tokens:             nil,
	       HighFrequencyPings: true,
	   }

	} else {

	   // Make next route from near relays (but respect hop limit)
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
	}

	if sessionResponse == nil {
	   fmt.Printf("error: nil session response\n")
	   return
	}

	sessionResponse.Version = sessionUpdate.Version

	excludeNearRelays(sessionResponse, sessionData.RouteState)

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

	fromAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", NEXT_SERVER_BACKEND_PORT))
	toAddress := &incoming.From

	var emptyMagic [8]byte
	sessionResponseData, err := transport.MarshalPacketSDK5(transport.PacketTypeSessionResponseSDK5, sessionResponse, emptyMagic[:], fromAddress, toAddress, crypto.BackendPrivateKey)
	if err != nil {
	   fmt.Printf("error: failed to marshal session response: %v\n", err)
	   return
	}

	if !core.BasicPacketFilter(sessionResponseData[:], len(sessionResponseData)) {
	   panic("basic packet filter failed on session response?")
	}

	{
	   var emptyMagic [8]byte
	   var fromAddressBuffer [32]byte
	   var toAddressBuffer [32]byte

	   fromAddressData, fromAddressPort := core.GetAddressData(fromAddress, fromAddressBuffer[:])
	   toAddressData, toAddressPort := core.GetAddressData(toAddress, toAddressBuffer[:])

	   if !core.AdvancedPacketFilter(sessionResponseData, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(sessionResponseData)) {
	       panic("advanced packet filter failed on session response\n")
	   }
	}

	if _, err := w.Write(sessionResponseData); err != nil {
	   fmt.Printf("error: failed to write session response: %v\n", err)
	   return
	}
}
*/

// -----------------------------------------------

func main() {

    rand.Seed(time.Now().UnixNano())

    backend.serverDatabase = make(map[string]ServerEntry)
    backend.sessionDatabase = make(map[uint64]SessionCacheEntry)
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
