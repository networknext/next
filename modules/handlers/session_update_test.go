package handlers_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	db "github.com/networknext/backend/modules/database"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/handlers"
	"github.com/networknext/backend/modules/packets"

	"github.com/stretchr/testify/assert"
)

func DummyLocateIP(ip net.IP) (packets.SDK5_LocationData, error) {
	location := packets.SDK5_LocationData{}
	location.Latitude = 43
	location.Longitude = -75
	return location, nil
}

func FailLocateIP(ip net.IP) (packets.SDK5_LocationData, error) {
	location := packets.SDK5_LocationData{}
	return location, fmt.Errorf("fail")
}

func CreateState() *handlers.SessionUpdateState {
	state := handlers.SessionUpdateState{}
	state.Request = &packets.SDK5_SessionUpdateRequestPacket{}
	state.LocateIP = DummyLocateIP
	state.RouteMatrix = &common.RouteMatrix{}
	state.RouteMatrix.CreatedAt = uint64(time.Now().Unix())
	state.Database = db.CreateDatabase()
	return &state
}

func generateRouteMatrix(relayIds []uint64, costMatrix []int32, relayDatacenters []uint64, database *db.Database) *common.RouteMatrix {

	numRelays := len(relayIds)

	numSegments := numRelays

	routeEntries := core.Optimize(numRelays, numSegments, costMatrix, 1, relayDatacenters[:])

	destRelays := make([]bool, numRelays)

	relayIdToIndex := make(map[uint64]int32)
	for i := range relayIds {
		relayIdToIndex[relayIds[i]] = int32(i)
	}

	relayAddresses := make([]net.UDPAddr, numRelays)
	for i := range relayAddresses {
		relayAddresses[i] = *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 40000+i))
	}

	relayNames := make([]string, numRelays)
	for i := range relayNames {
		relayNames[i] = fmt.Sprintf("relay-%d", i)
	}

	routeMatrix := &common.RouteMatrix{}

	routeMatrix.CreatedAt = uint64(time.Now().Unix())
	routeMatrix.Version = common.RouteMatrixVersion_Write
	routeMatrix.RouteEntries = routeEntries
	routeMatrix.RelayDatacenterIds = relayDatacenters
	routeMatrix.FullRelayIds = make([]uint64, 0)
	routeMatrix.FullRelayIndexSet = make(map[int32]bool)
	routeMatrix.DestRelays = destRelays
	routeMatrix.RelayIds = relayIds
	routeMatrix.RelayIdToIndex = relayIdToIndex
	routeMatrix.RelayNames = relayNames

	return routeMatrix
}

func Test_SessionUpdate_Pre_AnalysisOnly(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.RouteShader.AnalysisOnly = true

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.True(t, state.AnalysisOnly)
}

func Test_SessionUpdate_Pre_ClientPingTimedOut(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.ClientPingTimedOut = true

	result := handlers.SessionUpdate_Pre(state)

	assert.True(t, result)
	assert.True(t, state.ClientPingTimedOut)
}

func Test_SessionUpdate_Pre_LocatedIP(t *testing.T) {

	t.Parallel()

	state := CreateState()

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.True(t, state.LocatedIP)
	assert.False(t, state.LocationVeto)
	assert.Equal(t, state.Output.Location.Latitude, float32(43))
	assert.Equal(t, state.Output.Location.Longitude, float32(-75))
}

func Test_SessionUpdate_Pre_LocationVeto(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.LocateIP = FailLocateIP

	result := handlers.SessionUpdate_Pre(state)

	assert.True(t, result)
	assert.True(t, state.LocationVeto)
}

func Test_SessionUpdate_Pre_ReadLocation(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Request.SliceNumber = 1

	sessionData := packets.SDK5_SessionData{}

	sessionData.Version = packets.SDK5_SessionDataVersion_Write
	sessionData.Location.Version = packets.SDK5_LocationVersion_Write
	sessionData.Location.Latitude = 10
	sessionData.Location.Longitude = 20
	sessionData.Location.ISP = "Starlink"
	sessionData.Location.ASN = 5

	buffer := make([]byte, packets.SDK5_MaxSessionDataSize)
	writeStream := encoding.CreateWriteStream(buffer[:])
	err := sessionData.Serialize(writeStream)
	assert.Nil(t, err)
	writeStream.Flush()
	packetBytes := writeStream.GetBytesProcessed()
	packetData := buffer[:packetBytes]

	fmt.Printf("session data bytes = %d\n", len(packetData))

	state.Request.SessionDataBytes = int32(packetBytes)
	copy(state.Request.SessionData[:], packetData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(packetData, state.ServerBackendPrivateKey))

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.True(t, state.ReadSessionData)
	assert.Equal(t, float32(10), state.Output.Location.Latitude)
	assert.Equal(t, float32(20), state.Output.Location.Longitude)
	assert.Equal(t, "Starlink", state.Output.Location.ISP)
	assert.Equal(t, uint32(5), state.Output.Location.ASN)
}

func TestSessionUpdate_Pre_StaleRouteMatrix(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.RouteMatrix.CreatedAt = 0

	result := handlers.SessionUpdate_Pre(state)

	assert.True(t, result)
	assert.True(t, state.StaleRouteMatrix)
}

func Test_SessionUpdate_Pre_KnownDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Request.DatacenterId = 0x12345

	state.Database.DatacenterMap[0x12345] = db.Datacenter{}

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
}

func Test_SessionUpdate_Pre_UnknownDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Request.SliceNumber = 0
	state.Request.DatacenterId = 0x12345

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.True(t, state.UnknownDatacenter)
}

func Test_SessionUpdate_Pre_DatacenterNotEnabled(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Buyer.ID = 0x11111
	state.Request.SliceNumber = 0
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = db.Datacenter{}

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
	assert.True(t, state.DatacenterNotEnabled)
}

func Test_SessionUpdate_Pre_DatacenterEnabled(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Buyer.ID = 0x11111
	state.Request.BuyerId = 0x11111
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = db.Datacenter{}
	state.Database.DatacenterMaps[state.Buyer.ID] = make(map[uint64]db.DatacenterMap)
	state.Database.DatacenterMaps[state.Buyer.ID][state.Request.DatacenterId] = db.DatacenterMap{}

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
	assert.False(t, state.DatacenterNotEnabled)
}

func Test_SessionUpdate_Pre_SessionDataSignatureCheckFailed(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Request.SliceNumber = 1

	state.Request.SessionDataBytes = 100
	common.RandomBytes(state.Request.SessionData[:])

	result := handlers.SessionUpdate_Pre(state)

	assert.True(t, result)
	assert.True(t, state.SessionDataSignatureCheckFailed)
	assert.False(t, state.FailedToReadSessionData)
}

func Test_SessionUpdate_Pre_FailedToReadSessionData(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Request.SliceNumber = 1

	state.Request.SessionDataBytes = 100
	common.RandomBytes(state.Request.SessionData[:])
	copy(state.Request.SessionDataSignature[:], crypto.Sign(state.Request.SessionData[:state.Request.SessionDataBytes], state.ServerBackendPrivateKey))

	result := handlers.SessionUpdate_Pre(state)

	assert.True(t, result)
	assert.True(t, state.FailedToReadSessionData)
}

func Test_SessionUpdate_Pre_NoRelaysInDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Buyer.ID = 0x11111
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = db.Datacenter{}

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
	assert.True(t, state.NoRelaysInDatacenter)
}

func Test_SessionUpdate_Pre_RelaysInDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Buyer.ID = 0x11111
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = db.Datacenter{}

	const NumRelays = 10

	state.RouteMatrix.RelayIds = make([]uint64, NumRelays)
	state.RouteMatrix.RelayDatacenterIds = make([]uint64, NumRelays)

	for i := 0; i < NumRelays; i++ {
		state.RouteMatrix.RelayIds[i] = uint64(i)
		state.RouteMatrix.RelayDatacenterIds[i] = 0x12345
	}

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
	assert.False(t, state.NoRelaysInDatacenter)
}

func Test_SessionUpdate_Pre_Debug(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Buyer.Debug = true

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.NotNil(t, state.Debug)
}

// --------------------------------------------------------------

func Test_SessionUpdate_NewSession(t *testing.T) {

	t.Parallel()

	state := CreateState()

	sessionId := uint64(0x11223344556677)
	userHash := uint64(0x84731298749187)
	abTest := true

	state.Request.SessionId = sessionId
	state.Request.UserHash = userHash
	state.Buyer.RouteShader.ABTest = abTest

	handlers.SessionUpdate_NewSession(state)

	assert.Equal(t, state.Output.Version, uint32(packets.SDK5_SessionDataVersion_Write))
	assert.Equal(t, state.Output.SessionId, sessionId)
	assert.Equal(t, state.Output.SliceNumber, uint32(1))
	assert.Equal(t, state.Output.RouteState.UserID, userHash)
	assert.Equal(t, state.Output.RouteState.ABTest, abTest)
	assert.True(t, state.Output.ExpireTimestamp > uint64(time.Now().Unix()))

	assert.Equal(t, state.Input, state.Output)
}

// --------------------------------------------------------------

func Test_SessionUpdate_ExistingSession_FailedToReadSessionData(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	writeSessionData := make([]byte, 256)
	common.RandomBytes(writeSessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.FailedToReadSessionData)
}

func writeSessionData(sessionData packets.SDK5_SessionData) []byte {

	buffer := [packets.SDK5_MaxSessionDataSize]byte{}

	writeStream := encoding.CreateWriteStream(buffer[:])

	err := sessionData.Serialize(writeStream)
	if err != nil {
		panic(err)
	}

	writeStream.Flush()

	sessionDataBytes := writeStream.GetBytesProcessed()

	return buffer[:sessionDataBytes]
}

func Test_SessionUpdate_ExistingSession_ReadSessionData(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	sessionData := packets.GenerateRandomSessionData()

	writeSessionData := writeSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, state.FailedToReadSessionData)
}

func Test_SessionUpdate_ExistingSession_BadSessionId(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber

	writeSessionData := writeSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, state.FailedToReadSessionData)
	assert.True(t, state.BadSessionId)
	assert.False(t, state.BadSliceNumber)
}

func Test_SessionUpdate_ExistingSession_BadSliceNumber(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber

	writeSessionData := writeSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SessionId = sessionId

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, state.FailedToReadSessionData)
	assert.False(t, state.BadSessionId)
	assert.True(t, state.BadSliceNumber)
}

func Test_SessionUpdate_ExistingSession_PassConsistencyChecks(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber

	writeSessionData := writeSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SessionId = sessionId
	state.Request.SliceNumber = sliceNumber

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, state.FailedToReadSessionData)
	assert.False(t, state.BadSessionId)
	assert.False(t, state.BadSliceNumber)
}

func Test_SessionUpdate_ExistingSession_Output(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber

	writeSessionData := writeSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SessionId = sessionId
	state.Request.SliceNumber = sliceNumber

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, state.FailedToReadSessionData)
	assert.False(t, state.BadSessionId)
	assert.False(t, state.BadSliceNumber)
	assert.False(t, state.Output.Initial)
	assert.Equal(t, state.Output.SessionId, sessionId)
	assert.Equal(t, state.Output.SliceNumber, sliceNumber+1)
	assert.Equal(t, state.Output.ExpireTimestamp, state.Input.ExpireTimestamp+packets.SDK5_BillingSliceSeconds)
}

func Test_SessionUpdate_ExistingSession_RealPacketLoss(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber
	sessionData.PrevPacketsSentClientToServer = 1000
	sessionData.PrevPacketsSentServerToClient = 1000
	sessionData.PrevPacketsLostClientToServer = 0
	sessionData.PrevPacketsLostServerToClient = 0

	writeSessionData := writeSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SessionId = sessionId
	state.Request.SliceNumber = sliceNumber

	state.Request.PacketsSentClientToServer = 2000
	state.Request.PacketsSentServerToClient = 2000
	state.Request.PacketsLostClientToServer = 100
	state.Request.PacketsLostServerToClient = 50

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, state.FailedToReadSessionData)
	assert.False(t, state.BadSessionId)
	assert.False(t, state.BadSliceNumber)
	assert.False(t, state.Output.Initial)
	assert.Equal(t, state.Output.SessionId, sessionId)
	assert.Equal(t, state.Output.SliceNumber, sliceNumber+1)
	assert.Equal(t, state.Output.ExpireTimestamp, state.Input.ExpireTimestamp+packets.SDK5_BillingSliceSeconds)

	assert.Equal(t, state.RealPacketLoss, float32(10.0))
	assert.Equal(t, state.PostRealPacketLossServerToClient, float32(5.0))
	assert.Equal(t, state.PostRealPacketLossClientToServer, float32(10.0))
}

func Test_SessionUpdate_ExistingSession_RealJitter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber

	writeSessionData := writeSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SessionId = sessionId
	state.Request.SliceNumber = sliceNumber
	state.Request.JitterClientToServer = 50.0
	state.Request.JitterServerToClient = 100.0

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, state.FailedToReadSessionData)
	assert.False(t, state.BadSessionId)
	assert.False(t, state.BadSliceNumber)
	assert.False(t, state.Output.Initial)
	assert.Equal(t, state.Output.SessionId, sessionId)
	assert.Equal(t, state.Output.SliceNumber, sliceNumber+1)
	assert.Equal(t, state.Output.ExpireTimestamp, state.Input.ExpireTimestamp+packets.SDK5_BillingSliceSeconds)

	assert.Equal(t, state.RealJitter, float32(100.0))
}

// --------------------------------------------------------------

func Test_SessionUpdate_HandleFallbackToDirect_FallbackToDirect(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.FallbackToDirect = true

	result := handlers.SessionUpdate_HandleFallbackToDirect(state)

	assert.True(t, result)
	assert.True(t, state.FallbackToDirect)
	assert.True(t, state.Output.FallbackToDirect)
}

func Test_SessionUpdate_HandleFallbackToDirect_DoNotFallbackToDirect(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.FallbackToDirect = false

	result := handlers.SessionUpdate_HandleFallbackToDirect(state)

	assert.False(t, result)
	assert.False(t, state.FallbackToDirect)
	assert.False(t, state.Output.FallbackToDirect)
}

func Test_SessionUpdate_HandleFallbackToDirect_DontRepeat(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.FallbackToDirect = false
	state.Output.FallbackToDirect = true

	result := handlers.SessionUpdate_HandleFallbackToDirect(state)

	assert.False(t, result)
	assert.False(t, state.FallbackToDirect)
	assert.True(t, state.Output.FallbackToDirect)
}

// --------------------------------------------------------------

func Test_SessionUpdate_BuildNextTokens_PublicAddresses(t *testing.T) {

	t.Parallel()

	// initialize state

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, clientPrivateKey := crypto.Box_KeyPair()

	serverPublicKey, serverPrivateKey := crypto.Box_KeyPair()

	state.RoutingPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = serverAddress

	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, relay_private_key_a := crypto.Box_KeyPair()
	relay_public_key_b, relay_private_key_b := crypto.Box_KeyPair()
	relay_public_key_c, relay_private_key_c := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap["a"] = seller_a
	state.Database.SellerMap["b"] = seller_b
	state.Database.SellerMap["c"] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	// initialize route matrix

	state.RouteMatrix.RelayIds = make([]uint64, 3)
	state.RouteMatrix.RelayIds[0] = 1
	state.RouteMatrix.RelayIds[1] = 2
	state.RouteMatrix.RelayIds[2] = 3

	// initialize route relays

	routeNumRelays := int32(3)
	routeRelays := []int32{0, 1, 2}

	// build next tokens

	handlers.SessionUpdate_BuildNextTokens(state, routeNumRelays, routeRelays)

	// validate

	const NumTokens = 5

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK5_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK5_EncryptedNextRouteTokenSize)

	addresses := make([]*net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	privateKeys := make([][]byte, NumTokens)

	privateKeys[0] = clientPrivateKey
	privateKeys[1] = relay_private_key_a
	privateKeys[2] = relay_private_key_b
	privateKeys[3] = relay_private_key_c
	privateKeys[4] = serverPrivateKey

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK5_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK5_EncryptedNextRouteTokenSize]

		err := core.ReadEncryptedRouteToken(&token, tokenData, routingPublicKey, privateKeys[i])
		assert.Nil(t, err)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.KbpsUp, uint32(256))
		assert.Equal(t, token.KbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.PrivateKey {
			if token.PrivateKey[j] != 0 {
				found = true
				break
			}
		}
		assert.True(t, found)
	}
}

func Test_SessionUpdate_BuildNextTokens_PrivateAddresses(t *testing.T) {

	t.Parallel()

	// initialize state

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, clientPrivateKey := crypto.Box_KeyPair()

	serverPublicKey, serverPrivateKey := crypto.Box_KeyPair()

	state.RoutingPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = serverAddress

	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database

	seller := db.Seller{ID: "seller", Name: "seller"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_address_c_internal := core.ParseAddress("35.0.0.1:40002")

	relay_public_key_a, relay_private_key_a := crypto.Box_KeyPair()
	relay_public_key_b, relay_private_key_b := crypto.Box_KeyPair()
	relay_public_key_c, relay_private_key_c := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, InternalAddr: *relay_address_c_internal, Seller: seller, PublicKey: relay_public_key_c}

	state.Database.SellerMap["seller"] = seller

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	// initialize route matrix

	state.RouteMatrix.RelayIds = make([]uint64, 3)
	state.RouteMatrix.RelayIds[0] = 1
	state.RouteMatrix.RelayIds[1] = 2
	state.RouteMatrix.RelayIds[2] = 3

	// initialize route relays

	routeNumRelays := int32(3)
	routeRelays := []int32{0, 1, 2}

	// build next tokens

	handlers.SessionUpdate_BuildNextTokens(state, routeNumRelays, routeRelays)

	// validate

	const NumTokens = 5

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK5_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK5_EncryptedNextRouteTokenSize)

	addresses := make([]*net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	privateKeys := make([][]byte, NumTokens)

	privateKeys[0] = clientPrivateKey
	privateKeys[1] = relay_private_key_a
	privateKeys[2] = relay_private_key_b
	privateKeys[3] = relay_private_key_c
	privateKeys[4] = serverPrivateKey

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK5_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK5_EncryptedNextRouteTokenSize]

		err := core.ReadEncryptedRouteToken(&token, tokenData, routingPublicKey, privateKeys[i])
		assert.Nil(t, err)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.KbpsUp, uint32(256))
		assert.Equal(t, token.KbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else if i == 2 {
			assert.Equal(t, token.NextAddress.String(), relay_address_c_internal.String())
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.PrivateKey {
			if token.PrivateKey[j] != 0 {
				found = true
				break
			}
		}
		assert.True(t, found)
	}
}

// --------------------------------------------------------------

func Test_SessionUpdate_BuildContinueTokens(t *testing.T) {

	t.Parallel()

	// initialize state

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, clientPrivateKey := crypto.Box_KeyPair()

	serverPublicKey, serverPrivateKey := crypto.Box_KeyPair()

	state.RoutingPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, relay_private_key_a := crypto.Box_KeyPair()
	relay_public_key_b, relay_private_key_b := crypto.Box_KeyPair()
	relay_public_key_c, relay_private_key_c := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap["a"] = seller_a
	state.Database.SellerMap["b"] = seller_b
	state.Database.SellerMap["c"] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	// initialize route matrix

	state.RouteMatrix.RelayIds = make([]uint64, 3)
	state.RouteMatrix.RelayIds[0] = 1
	state.RouteMatrix.RelayIds[1] = 2
	state.RouteMatrix.RelayIds[2] = 3

	// initialize route relays

	routeNumRelays := int32(3)
	routeRelays := []int32{0, 1, 2}

	// build next tokens

	handlers.SessionUpdate_BuildContinueTokens(state, routeNumRelays, routeRelays)

	// validate

	const NumTokens = 5

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK5_RouteTypeContinue))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK5_EncryptedContinueRouteTokenSize)

	privateKeys := make([][]byte, NumTokens)

	privateKeys[0] = clientPrivateKey
	privateKeys[1] = relay_private_key_a
	privateKeys[2] = relay_private_key_b
	privateKeys[3] = relay_private_key_c
	privateKeys[4] = serverPrivateKey

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK5_EncryptedContinueRouteTokenSize * i

		token := core.ContinueToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK5_EncryptedContinueRouteTokenSize]

		err := core.ReadEncryptedContinueToken(&token, tokenData, routingPublicKey, privateKeys[i])
		assert.Nil(t, err)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
	}
}

// --------------------------------------------------------------

func Test_SessionUpdate_MakeRouteDecision_NoRouteRelays(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Input.RouteState.Next = true
	state.Input.RouteNumRelays = 0

	handlers.SessionUpdate_MakeRouteDecision(state)

	assert.False(t, state.Output.RouteState.Next)
	assert.True(t, state.Output.RouteState.Veto)
	assert.True(t, state.NoRouteRelays)
}

func Test_SessionUpdate_MakeRouteDecision_StayDirect(t *testing.T) {

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = false
	state.Request.DirectRTT = 100
	state.Request.SliceNumber = 100
	state.Debug = new(string)

	// initialize database with three relays

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller_c}

	state.Database.SellerMap["a"] = seller_a
	state.Database.SellerMap["b"] = seller_b
	state.Database.SellerMap["c"] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	// setup route shader

	state.Buyer.RouteShader = core.NewRouteShader()

	// setup internal config

	state.Buyer.InternalConfig = core.NewInternalConfig()

	// setup near relays

	state.SourceRelays = []int32{0, 1, 2}
	state.SourceRelayRTT = []int32{10, 10, 10}

	// setup dest relays

	state.DestRelays = []int32{0, 1, 2}

	// make the route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// verify

	assert.True(t, state.StayDirect)
	assert.False(t, state.TakeNetworkNext)
	assert.False(t, state.Output.RouteState.Next)
}

func Test_SessionUpdate_MakeRouteDecision_TakeNetworkNext(t *testing.T) {

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = false
	state.Request.DirectRTT = 100
	state.Request.SliceNumber = 100
	state.Debug = new(string)

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, clientPrivateKey := crypto.Box_KeyPair()

	serverPublicKey, serverPrivateKey := crypto.Box_KeyPair()

	state.RoutingPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, relay_private_key_a := crypto.Box_KeyPair()
	relay_public_key_b, relay_private_key_b := crypto.Box_KeyPair()
	relay_public_key_c, relay_private_key_c := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap["a"] = seller_a
	state.Database.SellerMap["b"] = seller_b
	state.Database.SellerMap["c"] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]int32, entryCount)

	for i := range costMatrix {
		costMatrix[i] = -1
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	// generate route matrix

	relayIds := [...]uint64{1, 2, 3}

	relayDatacenters := [...]uint64{1, 2, 3}

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	// setup route shader

	state.Buyer.RouteShader = core.NewRouteShader()

	state.Buyer.RouteShader.DisableNetworkNext = false
	state.Buyer.RouteShader.AnalysisOnly = false
	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	// setup internal config

	state.Buyer.InternalConfig = core.NewInternalConfig()

	// setup near relays

	state.SourceRelays = []int32{0, 1, 2}
	state.SourceRelayRTT = []int32{1, 100, 100}

	// setup dest relays

	state.DestRelays = []int32{2}

	// make the route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// verify output state

	assert.True(t, state.TakeNetworkNext)
	assert.True(t, state.Output.RouteState.Next)
	assert.True(t, state.Response.Multipath)

	assert.Equal(t, state.Output.RouteCost, int32(24))
	assert.False(t, state.Output.RouteChanged)
	assert.Equal(t, state.Output.RouteNumRelays, int32(3))

	assert.Equal(t, state.Output.RouteRelayIds[0], uint64(1))
	assert.Equal(t, state.Output.RouteRelayIds[1], uint64(2))
	assert.Equal(t, state.Output.RouteRelayIds[2], uint64(3))

	// verify route tokens

	const NumTokens = 5

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK5_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK5_EncryptedNextRouteTokenSize)

	addresses := make([]*net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	privateKeys := make([][]byte, NumTokens)

	privateKeys[0] = clientPrivateKey
	privateKeys[1] = relay_private_key_a
	privateKeys[2] = relay_private_key_b
	privateKeys[3] = relay_private_key_c
	privateKeys[4] = serverPrivateKey

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK5_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK5_EncryptedNextRouteTokenSize]

		err := core.ReadEncryptedRouteToken(&token, tokenData, routingPublicKey, privateKeys[i])
		assert.Nil(t, err)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.KbpsUp, uint32(256))
		assert.Equal(t, token.KbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.PrivateKey {
			if token.PrivateKey[j] != 0 {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	// verify debug string is set

	assert.NotEqual(t, *state.Debug, "")
}

func Test_SessionUpdate_MakeRouteDecision_Aborted(t *testing.T) {

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = true
	state.Input.RouteNumRelays = 5
	state.Request.Next = false

	// make the route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// verify output state

	assert.True(t, state.Aborted)
	assert.True(t, state.Output.RouteState.Veto)
	assert.False(t, state.Output.RouteState.Next)
}

func Test_SessionUpdate_MakeRouteDecision_RouteContinued(t *testing.T) {

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = false
	state.Request.DirectRTT = 100
	state.Request.SliceNumber = 100
	state.Debug = new(string)

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, clientPrivateKey := crypto.Box_KeyPair()

	serverPublicKey, serverPrivateKey := crypto.Box_KeyPair()

	state.RoutingPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, relay_private_key_a := crypto.Box_KeyPair()
	relay_public_key_b, relay_private_key_b := crypto.Box_KeyPair()
	relay_public_key_c, relay_private_key_c := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap["a"] = seller_a
	state.Database.SellerMap["b"] = seller_b
	state.Database.SellerMap["c"] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]int32, entryCount)

	for i := range costMatrix {
		costMatrix[i] = -1
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	// generate route matrix

	relayIds := [...]uint64{1, 2, 3}

	relayDatacenters := [...]uint64{1, 2, 3}

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	// setup route shader

	state.Buyer.RouteShader = core.NewRouteShader()

	state.Buyer.RouteShader.DisableNetworkNext = false
	state.Buyer.RouteShader.AnalysisOnly = false
	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	// setup internal config

	state.Buyer.InternalConfig = core.NewInternalConfig()

	// setup near relays

	state.SourceRelays = []int32{0, 1, 2}
	state.SourceRelayRTT = []int32{1, 100, 100}

	// setup dest relays

	state.DestRelays = []int32{2}

	// make the route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// verify output state

	assert.True(t, state.TakeNetworkNext)
	assert.True(t, state.Output.RouteState.Next)
	assert.True(t, state.Response.Multipath)

	assert.Equal(t, state.Output.RouteCost, int32(24))
	assert.False(t, state.Output.RouteChanged)
	assert.Equal(t, state.Output.RouteNumRelays, int32(3))

	assert.Equal(t, state.Output.RouteRelayIds[0], uint64(1))
	assert.Equal(t, state.Output.RouteRelayIds[1], uint64(2))
	assert.Equal(t, state.Output.RouteRelayIds[2], uint64(3))

	// verify route tokens

	const NumTokens = 5

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK5_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK5_EncryptedNextRouteTokenSize)

	addresses := make([]*net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	privateKeys := make([][]byte, NumTokens)

	privateKeys[0] = clientPrivateKey
	privateKeys[1] = relay_private_key_a
	privateKeys[2] = relay_private_key_b
	privateKeys[3] = relay_private_key_c
	privateKeys[4] = serverPrivateKey

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK5_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK5_EncryptedNextRouteTokenSize]

		err := core.ReadEncryptedRouteToken(&token, tokenData, routingPublicKey, privateKeys[i])
		assert.Nil(t, err)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.KbpsUp, uint32(256))
		assert.Equal(t, token.KbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.PrivateKey {
			if token.PrivateKey[j] != 0 {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	// verify debug string is set

	assert.NotEqual(t, *state.Debug, "")

	// setup to continue the route

	state.Input = state.Output

	state.Input.RouteState.Next = true
	state.Input.RouteNumRelays = 3
	state.Input.RouteRelayIds[0] = 1
	state.Input.RouteRelayIds[1] = 2
	state.Input.RouteRelayIds[2] = 3
	state.Request.Next = true

	// make route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// validate continue

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK5_RouteTypeContinue))

	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK5_EncryptedContinueRouteTokenSize)

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK5_EncryptedContinueRouteTokenSize * i

		token := core.ContinueToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK5_EncryptedContinueRouteTokenSize]

		err := core.ReadEncryptedContinueToken(&token, tokenData, routingPublicKey, privateKeys[i])
		assert.Nil(t, err)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
	}
}

func Test_SessionUpdate_MakeRouteDecision_RouteChanged(t *testing.T) {

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = false
	state.Request.DirectRTT = 100
	state.Request.SliceNumber = 100
	state.Debug = new(string)

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, clientPrivateKey := crypto.Box_KeyPair()

	serverPublicKey, serverPrivateKey := crypto.Box_KeyPair()

	state.RoutingPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, relay_private_key_a := crypto.Box_KeyPair()
	relay_public_key_b, relay_private_key_b := crypto.Box_KeyPair()
	relay_public_key_c, relay_private_key_c := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap["a"] = seller_a
	state.Database.SellerMap["b"] = seller_b
	state.Database.SellerMap["c"] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]int32, entryCount)

	for i := range costMatrix {
		costMatrix[i] = -1
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	// generate route matrix

	relayIds := [...]uint64{1, 2, 3}

	relayDatacenters := [...]uint64{1, 2, 3}

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	// setup route shader

	state.Buyer.RouteShader = core.NewRouteShader()

	state.Buyer.RouteShader.DisableNetworkNext = false
	state.Buyer.RouteShader.AnalysisOnly = false
	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	// setup internal config

	state.Buyer.InternalConfig = core.NewInternalConfig()

	// setup near relays

	state.SourceRelays = []int32{0, 1, 2}
	state.SourceRelayRTT = []int32{1, 100, 100}

	// setup dest relays

	state.DestRelays = []int32{2}

	// make the route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// verify output state

	assert.True(t, state.TakeNetworkNext)
	assert.True(t, state.Output.RouteState.Next)
	assert.True(t, state.Response.Multipath)

	assert.Equal(t, state.Output.RouteCost, int32(24))
	assert.False(t, state.Output.RouteChanged)
	assert.Equal(t, state.Output.RouteNumRelays, int32(3))

	assert.Equal(t, state.Output.RouteRelayIds[0], uint64(1))
	assert.Equal(t, state.Output.RouteRelayIds[1], uint64(2))
	assert.Equal(t, state.Output.RouteRelayIds[2], uint64(3))

	// verify route tokens

	const NumTokens = 5

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK5_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK5_EncryptedNextRouteTokenSize)

	addresses := make([]*net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	privateKeys := make([][]byte, NumTokens)

	privateKeys[0] = clientPrivateKey
	privateKeys[1] = relay_private_key_a
	privateKeys[2] = relay_private_key_b
	privateKeys[3] = relay_private_key_c
	privateKeys[4] = serverPrivateKey

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK5_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK5_EncryptedNextRouteTokenSize]

		err := core.ReadEncryptedRouteToken(&token, tokenData, routingPublicKey, privateKeys[i])
		assert.Nil(t, err)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.KbpsUp, uint32(256))
		assert.Equal(t, token.KbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.PrivateKey {
			if token.PrivateKey[j] != 0 {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	// verify debug string is set

	assert.NotEqual(t, *state.Debug, "")

	// setup so the route will change

	state.Input = state.Output

	state.Input.RouteState.Next = true
	state.Input.RouteNumRelays = 3
	state.Input.RouteRelayIds[0] = 1
	state.Input.RouteRelayIds[1] = 2
	state.Input.RouteRelayIds[2] = 3
	state.Request.Next = true

	costMatrix = make([]int32, entryCount)

	costMatrix[core.TriMatrixIndex(0, 1)] = 100
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	state.SourceRelayRTT[0] = 100
	state.SourceRelayRTT[1] = 1
	state.SourceRelayRTT[2] = 100

	// make route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// validate new route

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK5_RouteTypeNew))

	const NewTokens = 4

	assert.Equal(t, state.Response.NumTokens, int32(NewTokens))
	assert.Equal(t, len(state.Response.Tokens), NewTokens*packets.SDK5_EncryptedNextRouteTokenSize)

	addresses = make([]*net.UDPAddr, NewTokens)
	addresses[1] = relay_address_b
	addresses[2] = relay_address_c
	addresses[3] = serverAddress

	privateKeys = make([][]byte, NewTokens)

	privateKeys[0] = clientPrivateKey
	privateKeys[1] = relay_private_key_b
	privateKeys[2] = relay_private_key_c
	privateKeys[3] = serverPrivateKey

	for i := 0; i < NewTokens; i++ {

		index := packets.SDK5_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK5_EncryptedNextRouteTokenSize]

		err := core.ReadEncryptedRouteToken(&token, tokenData, routingPublicKey, privateKeys[i])
		assert.Nil(t, err)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.KbpsUp, uint32(256))
		assert.Equal(t, token.KbpsDown, uint32(1024))

		if i == 3 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.PrivateKey {
			if token.PrivateKey[j] != 0 {
				found = true
				break
			}
		}
		assert.True(t, found)
	}
}

func Test_SessionUpdate_MakeRouteDecision_RouteRelayNoLongerExists(t *testing.T) {

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = false
	state.Request.DirectRTT = 100
	state.Request.SliceNumber = 100
	state.Debug = new(string)

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, clientPrivateKey := crypto.Box_KeyPair()

	serverPublicKey, serverPrivateKey := crypto.Box_KeyPair()

	state.RoutingPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, relay_private_key_a := crypto.Box_KeyPair()
	relay_public_key_b, relay_private_key_b := crypto.Box_KeyPair()
	relay_public_key_c, relay_private_key_c := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap["a"] = seller_a
	state.Database.SellerMap["b"] = seller_b
	state.Database.SellerMap["c"] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	state.DestRelays = []int32{0, 1, 2}

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]int32, entryCount)

	for i := range costMatrix {
		costMatrix[i] = -1
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	// generate route matrix

	relayIds := make([]uint64, 3)
	relayIds[0] = 1
	relayIds[1] = 2
	relayIds[2] = 3

	relayDatacenters := make([]uint64, 3)
	relayDatacenters[0] = 1
	relayDatacenters[1] = 2
	relayDatacenters[2] = 3

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	// setup route shader

	state.Buyer.RouteShader = core.NewRouteShader()

	state.Buyer.RouteShader.DisableNetworkNext = false
	state.Buyer.RouteShader.AnalysisOnly = false
	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	// setup internal config

	state.Buyer.InternalConfig = core.NewInternalConfig()

	// setup near relays

	state.SourceRelays = []int32{0, 1, 2}
	state.SourceRelayRTT = []int32{1, 100, 100}

	// setup dest relays

	state.DestRelays = []int32{2}

	// make the route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// verify output state

	assert.True(t, state.TakeNetworkNext)
	assert.True(t, state.Output.RouteState.Next)
	assert.True(t, state.Response.Multipath)

	assert.Equal(t, state.Output.RouteCost, int32(24))
	assert.False(t, state.Output.RouteChanged)
	assert.Equal(t, state.Output.RouteNumRelays, int32(3))

	assert.Equal(t, state.Output.RouteRelayIds[0], uint64(1))
	assert.Equal(t, state.Output.RouteRelayIds[1], uint64(2))
	assert.Equal(t, state.Output.RouteRelayIds[2], uint64(3))

	// verify route tokens

	const NumTokens = 5

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK5_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK5_EncryptedNextRouteTokenSize)

	addresses := make([]*net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	privateKeys := make([][]byte, NumTokens)

	privateKeys[0] = clientPrivateKey
	privateKeys[1] = relay_private_key_a
	privateKeys[2] = relay_private_key_b
	privateKeys[3] = relay_private_key_c
	privateKeys[4] = serverPrivateKey

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK5_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK5_EncryptedNextRouteTokenSize]

		err := core.ReadEncryptedRouteToken(&token, tokenData, routingPublicKey, privateKeys[i])
		assert.Nil(t, err)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.KbpsUp, uint32(256))
		assert.Equal(t, token.KbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.PrivateKey {
			if token.PrivateKey[j] != 0 {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	// verify debug string is set

	assert.NotEqual(t, *state.Debug, "")

	// dummy out the route matrix so it is empty

	state.Input = state.Output

	state.Input.RouteState.Next = true
	state.Input.RouteNumRelays = 3
	state.Input.RouteRelayIds[0] = 1
	state.Input.RouteRelayIds[1] = 2
	state.Input.RouteRelayIds[2] = 3

	state.Request.Next = true

	state.SourceRelays = make([]int32, 0)
	state.SourceRelayRTT = make([]int32, 0)

	state.RouteMatrix = &common.RouteMatrix{}
	state.RouteMatrix.CreatedAt = uint64(time.Now().Unix())

	// make route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// validate that we tripped "route relay no longer exists" *and* "route no longer exists"

	assert.True(t, state.RouteRelayNoLongerExists)
	assert.True(t, state.RouteNoLongerExists)
}

func Test_SessionUpdate_MakeRouteDecision_RouteNoLongerExists_NearRelays(t *testing.T) {

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = false
	state.Request.DirectRTT = 100
	state.Request.SliceNumber = 100
	state.Debug = new(string)

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, clientPrivateKey := crypto.Box_KeyPair()

	serverPublicKey, serverPrivateKey := crypto.Box_KeyPair()

	state.RoutingPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, relay_private_key_a := crypto.Box_KeyPair()
	relay_public_key_b, relay_private_key_b := crypto.Box_KeyPair()
	relay_public_key_c, relay_private_key_c := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap["a"] = seller_a
	state.Database.SellerMap["b"] = seller_b
	state.Database.SellerMap["c"] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]int32, entryCount)

	for i := range costMatrix {
		costMatrix[i] = -1
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	// generate route matrix

	relayIds := make([]uint64, 3)
	relayIds[0] = 1
	relayIds[1] = 2
	relayIds[2] = 3

	relayDatacenters := make([]uint64, 3)
	relayDatacenters[0] = 1
	relayDatacenters[1] = 2
	relayDatacenters[2] = 3

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	// setup route shader

	state.Buyer.RouteShader = core.NewRouteShader()

	state.Buyer.RouteShader.DisableNetworkNext = false
	state.Buyer.RouteShader.AnalysisOnly = false
	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	// setup internal config

	state.Buyer.InternalConfig = core.NewInternalConfig()

	// setup near relays

	state.SourceRelays = []int32{0, 1, 2}
	state.SourceRelayRTT = []int32{1, 100, 100}

	// setup dest relays

	state.DestRelays = []int32{2}

	// make the route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// verify output state

	assert.True(t, state.TakeNetworkNext)
	assert.True(t, state.Output.RouteState.Next)
	assert.True(t, state.Response.Multipath)

	assert.Equal(t, state.Output.RouteCost, int32(24))
	assert.False(t, state.Output.RouteChanged)
	assert.Equal(t, state.Output.RouteNumRelays, int32(3))

	assert.Equal(t, state.Output.RouteRelayIds[0], uint64(1))
	assert.Equal(t, state.Output.RouteRelayIds[1], uint64(2))
	assert.Equal(t, state.Output.RouteRelayIds[2], uint64(3))

	// verify route tokens

	const NumTokens = 5

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK5_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK5_EncryptedNextRouteTokenSize)

	addresses := make([]*net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	privateKeys := make([][]byte, NumTokens)

	privateKeys[0] = clientPrivateKey
	privateKeys[1] = relay_private_key_a
	privateKeys[2] = relay_private_key_b
	privateKeys[3] = relay_private_key_c
	privateKeys[4] = serverPrivateKey

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK5_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK5_EncryptedNextRouteTokenSize]

		err := core.ReadEncryptedRouteToken(&token, tokenData, routingPublicKey, privateKeys[i])
		assert.Nil(t, err)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.KbpsUp, uint32(256))
		assert.Equal(t, token.KbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.PrivateKey {
			if token.PrivateKey[j] != 0 {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	// verify debug string is set

	assert.NotEqual(t, *state.Debug, "")

	// make all near relays unroutable

	state.Input = state.Output

	state.Input.RouteState.Next = true
	state.Input.RouteNumRelays = 3
	state.Input.RouteRelayIds[0] = 1
	state.Input.RouteRelayIds[1] = 2
	state.Input.RouteRelayIds[2] = 3

	state.Request.Next = true

	state.SourceRelays = make([]int32, 0)

	// make route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// validate that we tripped "route no longer exists"

	assert.False(t, state.RouteRelayNoLongerExists)
	assert.True(t, state.RouteNoLongerExists)
	assert.False(t, state.Output.RouteState.Next)
}

func Test_SessionUpdate_MakeRouteDecision_RouteNoLongerExists_MidRelay(t *testing.T) {

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = false
	state.Request.DirectRTT = 100
	state.Request.SliceNumber = 100
	state.Debug = new(string)

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, clientPrivateKey := crypto.Box_KeyPair()

	serverPublicKey, serverPrivateKey := crypto.Box_KeyPair()

	state.RoutingPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, relay_private_key_a := crypto.Box_KeyPair()
	relay_public_key_b, relay_private_key_b := crypto.Box_KeyPair()
	relay_public_key_c, relay_private_key_c := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap["a"] = seller_a
	state.Database.SellerMap["b"] = seller_b
	state.Database.SellerMap["c"] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]int32, entryCount)

	for i := range costMatrix {
		costMatrix[i] = -1
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	// generate route matrix

	relayIds := make([]uint64, 3)
	relayIds[0] = 1
	relayIds[1] = 2
	relayIds[2] = 3

	relayDatacenters := make([]uint64, 3)
	relayDatacenters[0] = 1
	relayDatacenters[1] = 2
	relayDatacenters[2] = 3

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	// setup route shader

	state.Buyer.RouteShader = core.NewRouteShader()

	state.Buyer.RouteShader.DisableNetworkNext = false
	state.Buyer.RouteShader.AnalysisOnly = false
	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	// setup internal config

	state.Buyer.InternalConfig = core.NewInternalConfig()

	// setup near relays

	state.SourceRelays = []int32{0, 1, 2}
	state.SourceRelayRTT = []int32{1, 100, 100}

	// setup dest relays

	state.DestRelays = []int32{2}

	// make the route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// verify output state

	assert.True(t, state.TakeNetworkNext)
	assert.True(t, state.Output.RouteState.Next)
	assert.True(t, state.Response.Multipath)

	assert.Equal(t, state.Output.RouteCost, int32(24))
	assert.False(t, state.Output.RouteChanged)
	assert.Equal(t, state.Output.RouteNumRelays, int32(3))

	assert.Equal(t, state.Output.RouteRelayIds[0], uint64(1))
	assert.Equal(t, state.Output.RouteRelayIds[1], uint64(2))
	assert.Equal(t, state.Output.RouteRelayIds[2], uint64(3))

	// verify route tokens

	const NumTokens = 5

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK5_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK5_EncryptedNextRouteTokenSize)

	addresses := make([]*net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	privateKeys := make([][]byte, NumTokens)

	privateKeys[0] = clientPrivateKey
	privateKeys[1] = relay_private_key_a
	privateKeys[2] = relay_private_key_b
	privateKeys[3] = relay_private_key_c
	privateKeys[4] = serverPrivateKey

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK5_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK5_EncryptedNextRouteTokenSize]

		err := core.ReadEncryptedRouteToken(&token, tokenData, routingPublicKey, privateKeys[i])
		assert.Nil(t, err)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.KbpsUp, uint32(256))
		assert.Equal(t, token.KbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.PrivateKey {
			if token.PrivateKey[j] != 0 {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	// verify debug string is set

	assert.NotEqual(t, *state.Debug, "")

	// generate new route matrix without the current route

	state.Input = state.Output

	state.Input.RouteState.Next = true
	state.Input.RouteNumRelays = 3
	state.Input.RouteRelayIds[0] = 1
	state.Input.RouteRelayIds[1] = 2
	state.Input.RouteRelayIds[2] = 3

	state.Request.Next = true

	costMatrix = make([]int32, entryCount)

	for i := range costMatrix {
		costMatrix[i] = -1
	}

	costMatrix[core.TriMatrixIndex(0, 2)] = 1000

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	// make route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// validate that we tripped "route no longer exists"

	assert.False(t, state.RouteRelayNoLongerExists)
	assert.True(t, state.RouteNoLongerExists)
	assert.False(t, state.Output.RouteState.Next)
}

func Test_SessionUpdate_MakeRouteDecision_Mispredict(t *testing.T) {

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = false
	state.Request.DirectRTT = 100
	state.Request.SliceNumber = 100
	state.Debug = new(string)

	_, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RoutingPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap["a"] = seller_a
	state.Database.SellerMap["b"] = seller_b
	state.Database.SellerMap["c"] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]int32, entryCount)

	for i := range costMatrix {
		costMatrix[i] = -1
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	// generate route matrix

	relayIds := make([]uint64, 3)
	relayIds[0] = 1
	relayIds[1] = 2
	relayIds[2] = 3

	relayDatacenters := make([]uint64, 3)
	relayDatacenters[0] = 1
	relayDatacenters[1] = 2
	relayDatacenters[2] = 3

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	// setup route shader

	state.Buyer.RouteShader = core.NewRouteShader()

	state.Buyer.RouteShader.DisableNetworkNext = false
	state.Buyer.RouteShader.AnalysisOnly = false
	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	// setup internal config

	state.Buyer.InternalConfig = core.NewInternalConfig()

	// setup near relays

	state.SourceRelays = []int32{0, 1, 2}
	state.SourceRelayRTT = []int32{1, 100, 100}

	// setup dest relays

	state.DestRelays = []int32{2}

	// make the route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// verify output state (we should be on next now)

	assert.True(t, state.TakeNetworkNext)
	assert.True(t, state.Output.RouteState.Next)
	assert.True(t, state.Response.Multipath)

	// mispredict 3 times

	state.Request.Next = true
	state.Request.NextRTT = 100

	for i := 0; i < 3; i++ {
		state.Input = state.Output
		handlers.SessionUpdate_MakeRouteDecision(state)
		if i < 2 {
			assert.False(t, state.Mispredict)
		}
	}

	// validate that we tripped "mispredicted"

	assert.True(t, state.Mispredict)
	assert.False(t, state.Output.RouteState.Next)
}

func Test_SessionUpdate_MakeRouteDecision_LatencyWorse(t *testing.T) {

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = false
	state.Request.DirectRTT = 100
	state.Request.SliceNumber = 100
	state.Debug = new(string)

	_, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RoutingPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap["a"] = seller_a
	state.Database.SellerMap["b"] = seller_b
	state.Database.SellerMap["c"] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]int32, entryCount)

	for i := range costMatrix {
		costMatrix[i] = -1
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	// generate route matrix

	relayIds := make([]uint64, 3)
	relayIds[0] = 1
	relayIds[1] = 2
	relayIds[2] = 3

	relayDatacenters := make([]uint64, 3)
	relayDatacenters[0] = 1
	relayDatacenters[1] = 2
	relayDatacenters[2] = 3

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	// setup route shader

	state.Buyer.RouteShader = core.NewRouteShader()

	state.Buyer.RouteShader.DisableNetworkNext = false
	state.Buyer.RouteShader.AnalysisOnly = false
	state.Buyer.RouteShader.Multipath = false
	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	// setup internal config

	state.Buyer.InternalConfig = core.NewInternalConfig()

	// setup near relays

	state.SourceRelays = []int32{0, 1, 2}
	state.SourceRelayRTT = []int32{1, 100, 100}

	// setup dest relays

	state.DestRelays = []int32{2}

	// make the route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// verify output state (we should be on next now)

	assert.True(t, state.TakeNetworkNext)
	assert.True(t, state.Output.RouteState.Next)
	assert.False(t, state.Response.Multipath)

	// make all near relays very expensive

	state.SourceRelayRTT = []int32{100, 100, 100}

	// make route decision

	state.Request.Next = true
	state.Request.NextRTT = 100
	state.Request.DirectRTT = 1

	state.Input = state.Output

	handlers.SessionUpdate_MakeRouteDecision(state)

	// validate that we tripped "latency worse"

	assert.True(t, state.LatencyWorse)
	assert.False(t, state.Output.RouteState.Next)
}

// --------------------------------------------------------------

func Test_SessionUpdate_GetNearRelays_AnalysisOnly(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.AnalysisOnly = true

	result := handlers.SessionUpdate_GetNearRelays(state)

	assert.False(t, result)
	assert.True(t, state.NotGettingNearRelaysAnalysisOnly)
	assert.Equal(t, state.Response.NumNearRelays, int32(0))
}

func Test_SessionUpdate_GetNearRelays_DatacenterNotEnabled(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.DatacenterNotEnabled = true

	result := handlers.SessionUpdate_GetNearRelays(state)

	assert.False(t, result)
	assert.True(t, state.NotGettingNearRelaysDatacenterNotEnabled)
	assert.Equal(t, state.Response.NumNearRelays, int32(0))
}

func Test_SessionUpdate_GetNearRelays_NoNearRelays(t *testing.T) {

	t.Parallel()

	state := CreateState()

	result := handlers.SessionUpdate_GetNearRelays(state)

	assert.False(t, result)
	assert.True(t, state.NoNearRelays)
	assert.Equal(t, state.Response.NumNearRelays, int32(0))
}

func Test_SessionUpdate_GetNearRelays_Success(t *testing.T) {

	t.Parallel()

	state := CreateState()

	// initialize database with three relays

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap["a"] = seller_a
	state.Database.SellerMap["b"] = seller_b
	state.Database.SellerMap["c"] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]int32, entryCount)

	for i := range costMatrix {
		costMatrix[i] = -1
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	// generate route matrix

	relayIds := make([]uint64, 3)
	relayIds[0] = 1
	relayIds[1] = 2
	relayIds[2] = 3

	relayDatacenters := make([]uint64, 3)
	relayDatacenters[0] = 1
	relayDatacenters[1] = 2
	relayDatacenters[2] = 3

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	state.RouteMatrix.RelayAddresses = make([]net.UDPAddr, NumRelays)
	state.RouteMatrix.RelayLatitudes = make([]float32, NumRelays)
	state.RouteMatrix.RelayLongitudes = make([]float32, NumRelays)

	state.RouteMatrix.RelayAddresses[0] = *relay_address_a
	state.RouteMatrix.RelayAddresses[1] = *relay_address_b
	state.RouteMatrix.RelayAddresses[2] = *relay_address_c

	// get near relays

	result := handlers.SessionUpdate_GetNearRelays(state)

	// validate

	assert.True(t, result)
	assert.False(t, state.NotGettingNearRelaysAnalysisOnly)
	assert.False(t, state.NotGettingNearRelaysDatacenterNotEnabled)
	assert.False(t, state.NoNearRelays)
	assert.Equal(t, state.Response.NumNearRelays, int32(3))
	assert.True(t, state.Response.HasNearRelays)

	contains_1 := false
	contains_2 := false
	contains_3 := false

	for i := 0; i < int(state.Response.NumNearRelays); i++ {
		if state.Response.NearRelayIds[i] == 1 {
			contains_1 = true
		}
		if state.Response.NearRelayIds[i] == 2 {
			contains_2 = true
		}
		if state.Response.NearRelayIds[i] == 3 {
			contains_3 = true
		}
	}

	assert.True(t, contains_1)
	assert.True(t, contains_2)
	assert.True(t, contains_3)
}

// --------------------------------------------------------------

func Test_SessionUpdate_UpdateNearRelays_AnalysisOnly(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.AnalysisOnly = true

	result := handlers.SessionUpdate_UpdateNearRelays(state)

	assert.False(t, result)
	assert.True(t, state.NotUpdatingNearRelaysAnalysisOnly)
	assert.Equal(t, state.Response.NumNearRelays, int32(0))
	assert.False(t, state.Response.HasNearRelays)
}

func Test_SessionUpdate_UpdateNearRelays_DatacenterNotEnabled(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.DatacenterNotEnabled = true

	result := handlers.SessionUpdate_UpdateNearRelays(state)

	assert.False(t, result)
	assert.True(t, state.NotUpdatingNearRelaysDatacenterNotEnabled)
	assert.Equal(t, state.Response.NumNearRelays, int32(0))
	assert.False(t, state.Response.HasNearRelays)
}

func Test_SessionUpdate_UpdateNearRelays_SliceOne(t *testing.T) {

	t.Parallel()

	state := CreateState()

	// initialize database with three relays

	seller := db.Seller{ID: "seller", Name: "seller"}

	datacenter := db.Datacenter{ID: 1, Name: "datacenter"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller, PublicKey: relay_public_key_c}

	state.Database.SellerMap["seller"] = seller

	state.Database.DatacenterMap[1] = datacenter

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	state.DestRelayIds = []uint64{1, 2, 3}

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]int32, entryCount)

	for i := range costMatrix {
		costMatrix[i] = -1
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	// generate route matrix

	relayIds := make([]uint64, 3)
	relayIds[0] = 1
	relayIds[1] = 2
	relayIds[2] = 3

	relayDatacenters := make([]uint64, 3)
	relayDatacenters[0] = 1
	relayDatacenters[1] = 2
	relayDatacenters[2] = 3

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	state.RouteMatrix.RelayAddresses = make([]net.UDPAddr, NumRelays)
	state.RouteMatrix.RelayLatitudes = make([]float32, NumRelays)
	state.RouteMatrix.RelayLongitudes = make([]float32, NumRelays)

	state.RouteMatrix.RelayAddresses[0] = *relay_address_a
	state.RouteMatrix.RelayAddresses[1] = *relay_address_b
	state.RouteMatrix.RelayAddresses[2] = *relay_address_c

	// setup near relays

	state.Request.NumNearRelays = 3
	copy(state.Request.NearRelayIds[:], []uint64{1, 2, 3})
	copy(state.Request.NearRelayRTT[:], []int32{10, 20, 30})
	copy(state.Request.NearRelayJitter[:], []int32{0, 0, 0})
	copy(state.Request.NearRelayPacketLoss[:], []float32{0, 0, 0})

	// update near relays

	state.Input.SliceNumber = 1

	result := handlers.SessionUpdate_UpdateNearRelays(state)

	// validate

	assert.True(t, result)
	assert.False(t, state.NotUpdatingNearRelaysAnalysisOnly)
	assert.False(t, state.NotUpdatingNearRelaysDatacenterNotEnabled)

	assert.Equal(t, len(state.DestRelays), 3)
	assert.Equal(t, state.DestRelays[0], int32(0))
	assert.Equal(t, state.DestRelays[1], int32(1))
	assert.Equal(t, state.DestRelays[2], int32(2))

	assert.Equal(t, state.Output.HeldNumNearRelays, int32(3))

	assert.Equal(t, state.Output.HeldNearRelayIds[0], uint64(1))
	assert.Equal(t, state.Output.HeldNearRelayIds[1], uint64(2))
	assert.Equal(t, state.Output.HeldNearRelayIds[2], uint64(3))

	assert.Equal(t, state.Output.HeldNearRelayRTT[0], int32(10))
	assert.Equal(t, state.Output.HeldNearRelayRTT[1], int32(20))
	assert.Equal(t, state.Output.HeldNearRelayRTT[2], int32(30))

	assert.Equal(t, len(state.SourceRelays), 3)
	assert.Equal(t, len(state.SourceRelayRTT), 3)

	assert.Equal(t, state.SourceRelays[0], int32(0))
	assert.Equal(t, state.SourceRelays[1], int32(1))
	assert.Equal(t, state.SourceRelays[2], int32(2))

	assert.Equal(t, state.SourceRelayRTT[0], int32(10))
	assert.Equal(t, state.SourceRelayRTT[1], int32(20))
	assert.Equal(t, state.SourceRelayRTT[2], int32(30))

	assert.Equal(t, state.Response.NumNearRelays, int32(0))
	assert.False(t, state.Response.HasNearRelays)
}

func Test_SessionUpdate_UpdateNearRelays_SliceTwo(t *testing.T) {

	t.Parallel()

	state := CreateState()

	// initialize database with three relays

	seller := db.Seller{ID: "seller", Name: "seller"}

	datacenter := db.Datacenter{ID: 1, Name: "datacenter"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	relay_a := db.Relay{ID: 1, Name: "a", Addr: *relay_address_a, Seller: seller, PublicKey: relay_public_key_a}
	relay_b := db.Relay{ID: 2, Name: "b", Addr: *relay_address_b, Seller: seller, PublicKey: relay_public_key_b}
	relay_c := db.Relay{ID: 3, Name: "c", Addr: *relay_address_c, Seller: seller, PublicKey: relay_public_key_c}

	state.Database.SellerMap["seller"] = seller

	state.Database.DatacenterMap[1] = datacenter

	state.Database.RelayMap[1] = relay_a
	state.Database.RelayMap[2] = relay_b
	state.Database.RelayMap[3] = relay_c

	state.DestRelayIds = []uint64{1, 2, 3}

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]int32, entryCount)

	for i := range costMatrix {
		costMatrix[i] = -1
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	// generate route matrix

	relayIds := make([]uint64, 3)
	relayIds[0] = 1
	relayIds[1] = 2
	relayIds[2] = 3

	relayDatacenters := make([]uint64, 3)
	relayDatacenters[0] = 1
	relayDatacenters[1] = 2
	relayDatacenters[2] = 3

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	state.RouteMatrix.RelayAddresses = make([]net.UDPAddr, NumRelays)
	state.RouteMatrix.RelayLatitudes = make([]float32, NumRelays)
	state.RouteMatrix.RelayLongitudes = make([]float32, NumRelays)

	state.RouteMatrix.RelayAddresses[0] = *relay_address_a
	state.RouteMatrix.RelayAddresses[1] = *relay_address_b
	state.RouteMatrix.RelayAddresses[2] = *relay_address_c

	// setup held near relays

	state.Output.HeldNumNearRelays = 3
	copy(state.Output.HeldNearRelayIds[:], []uint64{1, 2, 3})
	copy(state.Output.HeldNearRelayRTT[:], []int32{10, 20, 30})

	// update near relays

	state.Input.SliceNumber = 2

	result := handlers.SessionUpdate_UpdateNearRelays(state)

	// validate

	assert.True(t, result)
	assert.False(t, state.NotUpdatingNearRelaysAnalysisOnly)
	assert.False(t, state.NotUpdatingNearRelaysDatacenterNotEnabled)

	assert.Equal(t, len(state.DestRelays), 3)
	assert.Equal(t, state.DestRelays[0], int32(0))
	assert.Equal(t, state.DestRelays[1], int32(1))
	assert.Equal(t, state.DestRelays[2], int32(2))

	assert.Equal(t, state.Output.HeldNumNearRelays, int32(3))

	assert.Equal(t, state.Output.HeldNearRelayIds[0], uint64(1))
	assert.Equal(t, state.Output.HeldNearRelayIds[1], uint64(2))
	assert.Equal(t, state.Output.HeldNearRelayIds[2], uint64(3))

	assert.Equal(t, state.Output.HeldNearRelayRTT[0], int32(10))
	assert.Equal(t, state.Output.HeldNearRelayRTT[1], int32(20))
	assert.Equal(t, state.Output.HeldNearRelayRTT[2], int32(30))

	assert.Equal(t, len(state.SourceRelays), 3)
	assert.Equal(t, len(state.SourceRelayRTT), 3)

	assert.Equal(t, state.SourceRelays[0], int32(0))
	assert.Equal(t, state.SourceRelays[1], int32(1))
	assert.Equal(t, state.SourceRelays[2], int32(2))

	assert.Equal(t, state.SourceRelayRTT[0], int32(10))
	assert.Equal(t, state.SourceRelayRTT[1], int32(20))
	assert.Equal(t, state.SourceRelayRTT[2], int32(30))

	assert.Equal(t, state.Response.NumNearRelays, int32(0))
	assert.False(t, state.Response.HasNearRelays)
}

// --------------------------------------------------------------

func Test_SessionUpdate_Post_SliceZero(t *testing.T) {

	t.Parallel()

	state := CreateState()

	_, routingPrivateKey := crypto.Box_KeyPair()

	var serverBackendPublicKey [packets.SDK5_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var serverBackendPrivateKey [packets.SDK5_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK5_SignKeypair(serverBackendPublicKey[:], serverBackendPublicKey[:])

	state.RoutingPrivateKey = routingPrivateKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	state.From = core.ParseAddress("127.0.0.1:40000")
	state.ServerBackendAddress = core.ParseAddress("127.0.0.1:50000")

	state.Request.SliceNumber = 0

	handlers.SessionUpdate_Post(state)

	assert.True(t, state.GetNearRelays)
	assert.False(t, state.Response.HasNearRelays)
}

func Test_SessionUpdate_Post_DurationOnNext(t *testing.T) {

	t.Parallel()

	state := CreateState()

	_, routingPrivateKey := crypto.Box_KeyPair()

	var serverBackendPublicKey [packets.SDK5_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var serverBackendPrivateKey [packets.SDK5_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK5_SignKeypair(serverBackendPublicKey[:], serverBackendPublicKey[:])

	state.RoutingPrivateKey = routingPrivateKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	state.From = core.ParseAddress("127.0.0.1:40000")
	state.ServerBackendAddress = core.ParseAddress("127.0.0.1:50000")

	state.Request.Next = true
	state.Request.SliceNumber = 1

	handlers.SessionUpdate_Post(state)

	assert.False(t, state.GetNearRelays)
	assert.True(t, state.Output.EverOnNext)
	assert.Equal(t, state.Output.DurationOnNext, uint32(packets.SDK5_BillingSliceSeconds))
}

func Test_SessionUpdate_Post_PacketsSentPacketsLost(t *testing.T) {

	t.Parallel()

	state := CreateState()

	_, routingPrivateKey := crypto.Box_KeyPair()

	var serverBackendPublicKey [packets.SDK5_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var serverBackendPrivateKey [packets.SDK5_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK5_SignKeypair(serverBackendPublicKey[:], serverBackendPublicKey[:])

	state.RoutingPrivateKey = routingPrivateKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	state.From = core.ParseAddress("127.0.0.1:40000")
	state.ServerBackendAddress = core.ParseAddress("127.0.0.1:50000")

	state.Request.SliceNumber = 2

	state.Request.PacketsSentClientToServer = 10001
	state.Request.PacketsSentServerToClient = 10002
	state.Request.PacketsLostClientToServer = 10003
	state.Request.PacketsLostServerToClient = 10004

	handlers.SessionUpdate_Post(state)

	assert.Equal(t, state.Output.PrevPacketsSentClientToServer, state.Request.PacketsSentClientToServer)
	assert.Equal(t, state.Output.PrevPacketsSentServerToClient, state.Request.PacketsSentServerToClient)
	assert.Equal(t, state.Output.PrevPacketsLostClientToServer, state.Request.PacketsLostClientToServer)
	assert.Equal(t, state.Output.PrevPacketsLostServerToClient, state.Request.PacketsLostServerToClient)
}

func Test_SessionUpdate_Post_Debug(t *testing.T) {

	t.Parallel()

	state := CreateState()

	_, routingPrivateKey := crypto.Box_KeyPair()

	var serverBackendPublicKey [packets.SDK5_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var serverBackendPrivateKey [packets.SDK5_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK5_SignKeypair(serverBackendPublicKey[:], serverBackendPublicKey[:])

	state.RoutingPrivateKey = routingPrivateKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	state.From = core.ParseAddress("127.0.0.1:40000")
	state.ServerBackendAddress = core.ParseAddress("127.0.0.1:50000")

	state.Request.SliceNumber = 2

	debugString := "it's debug time"

	state.Debug = &debugString

	handlers.SessionUpdate_Post(state)

	assert.True(t, state.Response.HasDebug)
	assert.Equal(t, state.Response.Debug, *state.Debug)
	assert.False(t, state.Response.HasNearRelays)
}

func Test_SessionUpdate_Post_WriteSummary(t *testing.T) {

	t.Parallel()

	state := CreateState()

	_, routingPrivateKey := crypto.Box_KeyPair()

	var serverBackendPublicKey [packets.SDK5_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var serverBackendPrivateKey [packets.SDK5_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK5_SignKeypair(serverBackendPublicKey[:], serverBackendPublicKey[:])

	state.RoutingPrivateKey = routingPrivateKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	state.From = core.ParseAddress("127.0.0.1:40000")
	state.ServerBackendAddress = core.ParseAddress("127.0.0.1:50000")

	state.Request.SliceNumber = 100
	state.Request.ClientPingTimedOut = true

	handlers.SessionUpdate_Post(state)

	assert.True(t, state.Output.WriteSummary)
	assert.False(t, state.Output.WroteSummary)
	assert.False(t, state.Response.HasNearRelays)
}

func Test_SessionUpdate_Post_WroteSummary(t *testing.T) {

	t.Parallel()

	state := CreateState()

	_, routingPrivateKey := crypto.Box_KeyPair()

	var serverBackendPublicKey [packets.SDK5_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var serverBackendPrivateKey [packets.SDK5_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK5_SignKeypair(serverBackendPublicKey[:], serverBackendPublicKey[:])

	state.RoutingPrivateKey = routingPrivateKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	state.From = core.ParseAddress("127.0.0.1:40000")
	state.ServerBackendAddress = core.ParseAddress("127.0.0.1:50000")

	state.Request.SliceNumber = 100
	state.Request.ClientPingTimedOut = true
	state.Output.WriteSummary = true

	handlers.SessionUpdate_Post(state)

	assert.False(t, state.Output.WriteSummary)
	assert.True(t, state.Output.WroteSummary)
	assert.False(t, state.Response.HasNearRelays)
}

func Test_SessionUpdate_Post_Response(t *testing.T) {

	t.Parallel()

	state := CreateState()

	// setup so we write a response with random session data in the post

	_, routingPrivateKey := crypto.Box_KeyPair()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.RoutingPrivateKey = routingPrivateKey
	state.ServerBackendPublicKey = serverBackendPublicKey[:]
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	state.From = core.ParseAddress("127.0.0.1:40000")
	state.ServerBackendAddress = core.ParseAddress("127.0.0.1:50000")

	state.Input = packets.GenerateRandomSessionData()
	state.Output = state.Input

	// run session update post

	handlers.SessionUpdate_Post(state)

	// verify we wrote the session data and response packet without error

	assert.True(t, state.WroteResponsePacket)
	assert.False(t, state.FailedToWriteSessionData)
	assert.False(t, state.FailedToWriteResponsePacket)
	assert.True(t, len(state.ResponsePacket) > 0)

	// make sure the basic packet filter passes

	packetData := state.ResponsePacket

	assert.True(t, core.BasicPacketFilter(packetData[:], len(packetData)))

	// make sure the advanced packet filter passes

	to := state.From
	from := state.ServerBackendAddress

	var emptyMagic [8]byte

	var fromAddressBuffer [32]byte
	var toAddressBuffer [32]byte

	fromAddressData, fromAddressPort := core.GetAddressData(from, fromAddressBuffer[:])
	toAddressData, toAddressPort := core.GetAddressData(to, toAddressBuffer[:])

	assert.True(t, core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(packetData)))

	// check packet signature

	assert.True(t, packets.SDK5_CheckPacketSignature(packetData, state.ServerBackendPublicKey[:]))

	// verify we can read the response packet

	packetData = packetData[16:]

	packet := packets.SDK5_SessionUpdateResponsePacket{}
	err := packets.ReadPacket(packetData, &packet)
	assert.Nil(t, err)

	// verify the response packet is equal to the response in state

	assert.Equal(t, packet, state.Response)

	// verify that the signature check passes on the session data inside the response

	assert.True(t, crypto.Verify(packet.SessionData[:packet.SessionDataBytes], state.ServerBackendPublicKey[:], packet.SessionDataSignature[:]))

	// verify that we can serialize read the session data inside the response

	sessionData := packets.SDK5_SessionData{}
	err = packets.ReadPacket(packet.SessionData[:packet.SessionDataBytes], &sessionData)
	assert.Nil(t, err)

	// verify that the session data we read matches what was written

	assert.Equal(t, state.Output, sessionData)
}

// --------------------------------------------------------------
