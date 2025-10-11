package handlers_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
	db "github.com/networknext/next/modules/database"
	"github.com/networknext/next/modules/encoding"
	"github.com/networknext/next/modules/handlers"
	"github.com/networknext/next/modules/packets"

	"github.com/stretchr/testify/assert"
)

func DummyLocateIP(ip net.IP) (float32, float32) {
	return 43.0, -75.0
}

func FailLocateIP(ip net.IP) (float32, float32) {
	return 0.0, 0.0
}

func CreateState() *handlers.SessionUpdateState {
	state := handlers.SessionUpdateState{}
	state.Request = &packets.SDK_SessionUpdateRequestPacket{}
	state.Request.ClientAddress = core.ParseAddress("127.0.0.1:5000")
	state.RouteMatrix = &common.RouteMatrix{}
	state.RouteMatrix.CreatedAt = uint64(time.Now().Unix())
	state.Database = db.CreateDatabase()
	state.Input.Latitude = 35.0
	state.Input.Longitude = -75.0
	state.Buyer = &db.Buyer{}
	state.Datacenter = &db.Datacenter{}
	state.PingKey = make([]byte, crypto.Auth_KeySize)
	common.RandomBytes(state.PingKey)
	return &state
}

func generateRouteMatrix(relayIds []uint64, costMatrix []uint8, relayDatacenters []uint64, database *db.Database) *common.RouteMatrix {

	numRelays := len(relayIds)

	numSegments := numRelays

	relayPrice := make([]uint8, numRelays)

	routeEntries := core.Optimize(numRelays, numSegments, costMatrix, relayPrice, relayDatacenters[:])

	destRelays := make([]bool, numRelays)

	relayIdToIndex := make(map[uint64]int32)
	for i := range relayIds {
		relayIdToIndex[relayIds[i]] = int32(i)
	}

	relayAddresses := make([]net.UDPAddr, numRelays)
	for i := range relayAddresses {
		relayAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 40000+i))
	}

	relayNames := make([]string, numRelays)
	for i := range relayNames {
		relayNames[i] = fmt.Sprintf("relay-%d", i)
	}

	routeMatrix := &common.RouteMatrix{}

	routeMatrix.CreatedAt = uint64(time.Now().Unix())
	routeMatrix.Version = common.RouteMatrixVersion_Write
	routeMatrix.RelayDatacenterIds = relayDatacenters
	routeMatrix.DestRelays = destRelays
	routeMatrix.RelayIds = relayIds
	routeMatrix.RelayIdToIndex = relayIdToIndex
	routeMatrix.RelayNames = relayNames
	routeMatrix.RouteEntries = routeEntries

	return routeMatrix
}

func WriteSessionData(sessionData packets.SDK_SessionData) []byte {

	buffer := [packets.SDK_MaxSessionDataSize]byte{}

	writeStream := encoding.CreateWriteStream(buffer[:])

	err := sessionData.Serialize(writeStream)
	if err != nil {
		panic(err)
	}

	writeStream.Flush()

	sessionDataBytes := writeStream.GetBytesProcessed()

	return buffer[:sessionDataBytes]
}

func Test_SessionUpdate_Pre_FallbackToDirect(t *testing.T) {

	t.Parallel()

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	state.ServerBackendPublicKey = serverBackendPublicKey[:]
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	from := core.ParseAddress("127.0.0.1:40000")
	state.From = &from
	serverBackendAddress := core.ParseAddress("127.0.0.1:50000")
	state.ServerBackendAddress = &serverBackendAddress

	state.Request.SliceNumber = 100
	state.Request.ClientPingTimedOut = true
	state.Output.WriteSummary = true

	state.Request.FallbackToDirect = true

	sessionData := packets.GenerateRandomSessionData()
	writeSessionData := WriteSessionData(sessionData)
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))
	state.Request.SessionDataBytes = int32(len(writeSessionData))

	result := handlers.SessionUpdate_Pre(state)

	assert.True(t, result)
	assert.True(t, (state.Error&constants.SessionError_FallbackToDirect) != 0)
}

func Test_SessionUpdate_Pre_AnalysisOnly(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.RouteShader.AnalysisOnly = true

	result := handlers.SessionUpdate_Pre(state)

	assert.True(t, result)
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

func TestSessionUpdate_Pre_StaleRouteMatrix(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.RouteMatrix.CreatedAt = 0

	result := handlers.SessionUpdate_Pre(state)

	assert.True(t, result)
	assert.True(t, (state.Error&constants.SessionError_StaleRouteMatrix) != 0)
}

func Test_SessionUpdate_Pre_KnownDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Request.DatacenterId = 0x12345

	state.Database.DatacenterMap[0x12345] = &db.Datacenter{}

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.False(t, (state.Error&constants.SessionError_UnknownDatacenter) != 0)
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
	assert.True(t, (state.Error&constants.SessionError_UnknownDatacenter) != 0)
}

func Test_SessionUpdate_Pre_DatacenterNotEnabled(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Buyer.Id = 0x11111
	state.Request.SliceNumber = 0
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = &db.Datacenter{}

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.False(t, (state.Error&constants.SessionError_UnknownDatacenter) != 0)
	assert.True(t, (state.Error&constants.SessionError_DatacenterNotEnabled) != 0)
}

func Test_SessionUpdate_Pre_DatacenterEnabled(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	state.Buyer.Id = 0x11111
	state.Request.BuyerId = 0x11111
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = &db.Datacenter{}
	state.Database.BuyerDatacenterSettings[state.Buyer.Id] = make(map[uint64]*db.BuyerDatacenterSettings)
	state.Database.BuyerDatacenterSettings[state.Buyer.Id][state.Request.DatacenterId] = &db.BuyerDatacenterSettings{EnableAcceleration: true, BuyerId: 0x11111, DatacenterId: 0x12345}

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.False(t, (state.Error&constants.SessionError_UnknownDatacenter) != 0)
	assert.False(t, (state.Error&constants.SessionError_DatacenterNotEnabled) != 0)
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

	assert.Equal(t, state.Output.Version, uint32(packets.SDK_SessionDataVersion_Write))
	assert.Equal(t, state.Output.SessionId, sessionId)
	assert.Equal(t, state.Output.SliceNumber, uint32(1))
	assert.Equal(t, state.Output.RouteState.ABTest, abTest)
	assert.True(t, state.Output.ExpireTimestamp > uint64(time.Now().Unix()))
}

// --------------------------------------------------------------

func Test_SessionUpdate_Pre_FailedToReadSessionData(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	writeSessionData := make([]byte, 1)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SliceNumber = 10

	handlers.SessionUpdate_Pre(state)

	assert.True(t, (state.Error&constants.SessionError_FailedToReadSessionData) != 0)
}

func Test_SessionUpdate_Pre_ReadSessionData(t *testing.T) {

	t.Parallel()

	state := CreateState()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SliceNumber = 10

	writeSessionData := WriteSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SliceNumber = 10

	handlers.SessionUpdate_Pre(state)

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, (state.Error&constants.SessionError_FailedToReadSessionData) != 0)
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

	writeSessionData := WriteSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SliceNumber = 5

	handlers.SessionUpdate_Pre(state)

	assert.True(t, state.ReadSessionData)

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, (state.Error&constants.SessionError_FailedToReadSessionData) != 0)
	assert.True(t, (state.Error&constants.SessionError_BadSessionId) != 0)
	assert.False(t, (state.Error&constants.SessionError_BadSliceNumber) != 0)
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

	writeSessionData := WriteSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SliceNumber = 5
	state.Request.SessionId = sessionId

	handlers.SessionUpdate_Pre(state)

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, (state.Error&constants.SessionError_FailedToReadSessionData) != 0)
	assert.False(t, (state.Error&constants.SessionError_BadSessionId) != 0)
	assert.True(t, (state.Error&constants.SessionError_BadSliceNumber) != 0)
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

	writeSessionData := WriteSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SessionId = sessionId
	state.Request.SliceNumber = sliceNumber

	handlers.SessionUpdate_Pre(state)

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, (state.Error&constants.SessionError_FailedToReadSessionData) != 0)
	assert.False(t, (state.Error&constants.SessionError_BadSessionId) != 0)
	assert.False(t, (state.Error&constants.SessionError_BadSliceNumber) != 0)
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

	writeSessionData := WriteSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SessionId = sessionId
	state.Request.SliceNumber = sliceNumber

	handlers.SessionUpdate_Pre(state)

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.Equal(t, state.Output.SessionId, sessionId)
	assert.Equal(t, state.Output.SliceNumber, sliceNumber+1)
	assert.Equal(t, state.Output.ExpireTimestamp, state.Input.ExpireTimestamp+packets.SDK_SliceSeconds)
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

	writeSessionData := WriteSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SessionId = sessionId
	state.Request.SliceNumber = sliceNumber

	state.Request.PacketsSentClientToServer = 2000
	state.Request.PacketsSentServerToClient = 2000
	state.Request.PacketsLostClientToServer = 100
	state.Request.PacketsLostServerToClient = 50

	handlers.SessionUpdate_Pre(state)

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.Equal(t, state.Output.SessionId, sessionId)
	assert.Equal(t, state.Output.SliceNumber, sliceNumber+1)
	assert.Equal(t, state.Output.ExpireTimestamp, state.Input.ExpireTimestamp+packets.SDK_SliceSeconds)

	assert.Equal(t, state.RealPacketLoss, float32(10.0))
}

func Test_SessionUpdate_ExistingSession_RealOutOfOrder(t *testing.T) {

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

	writeSessionData := WriteSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SessionId = sessionId
	state.Request.SliceNumber = sliceNumber

	state.Request.PacketsSentClientToServer = 2000
	state.Request.PacketsSentServerToClient = 2000
	state.Request.PacketsOutOfOrderClientToServer = 100
	state.Request.PacketsOutOfOrderServerToClient = 50

	handlers.SessionUpdate_Pre(state)

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.Equal(t, state.Output.SessionId, sessionId)
	assert.Equal(t, state.Output.SliceNumber, sliceNumber+1)
	assert.Equal(t, state.Output.ExpireTimestamp, state.Input.ExpireTimestamp+packets.SDK_SliceSeconds)

	assert.Equal(t, state.RealOutOfOrder, float32(10.0))
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

	writeSessionData := WriteSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SessionId = sessionId
	state.Request.SliceNumber = sliceNumber
	state.Request.JitterClientToServer = 50.0
	state.Request.JitterServerToClient = 100.0

	handlers.SessionUpdate_Pre(state)

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.Equal(t, state.Output.SessionId, sessionId)
	assert.Equal(t, state.Output.SliceNumber, sliceNumber+1)
	assert.Equal(t, state.Output.ExpireTimestamp, state.Input.ExpireTimestamp+packets.SDK_SliceSeconds)

	assert.Equal(t, state.RealJitter, float32(100.0))
}

func Test_SessionUpdate_ExistingSession_EnvelopeBandwidth(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber
	sessionData.RouteState.Next = true
	sessionData.NextEnvelopeBytesUpSum = 1000
	sessionData.NextEnvelopeBytesDownSum = 1000

	writeSessionData := WriteSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.Next = true
	state.Request.SessionId = sessionId
	state.Request.SliceNumber = sliceNumber
	state.Request.JitterClientToServer = 50.0
	state.Request.JitterServerToClient = 100.0

	handlers.SessionUpdate_Pre(state)

	handlers.SessionUpdate_ExistingSession(state)

	assert.Equal(t, state.Output.NextEnvelopeBytesUpSum, uint64(321000))
	assert.Equal(t, state.Output.NextEnvelopeBytesDownSum, uint64(1281000))
}

func Test_SessionUpdate_ExistingSession_EnvelopeBandwidthOnlyOnNext(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.ServerBackendPublicKey = serverBackendPublicKey
	state.ServerBackendPrivateKey = serverBackendPrivateKey

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.RouteState.Next = false
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber
	sessionData.NextEnvelopeBytesUpSum = 0
	sessionData.NextEnvelopeBytesDownSum = 0

	writeSessionData := WriteSessionData(sessionData)

	state.Request.SessionDataBytes = int32(len(writeSessionData))
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))

	state.Request.SessionId = sessionId
	state.Request.SliceNumber = sliceNumber
	state.Request.JitterClientToServer = 50.0
	state.Request.JitterServerToClient = 100.0

	handlers.SessionUpdate_Pre(state)

	handlers.SessionUpdate_ExistingSession(state)

	assert.Equal(t, state.Output.NextEnvelopeBytesUpSum, uint64(0))
	assert.Equal(t, state.Output.NextEnvelopeBytesDownSum, uint64(0))
}

// --------------------------------------------------------------

func Test_SessionUpdate_BuildNextTokens_PublicAddresses(t *testing.T) {

	t.Parallel()

	// initialize state

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = &serverAddress

	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database

	seller_a := &db.Seller{Id: 1, Name: "a"}
	seller_b := &db.Seller{Id: 2, Name: "b"}
	seller_c := &db.Seller{Id: 3, Name: "c"}

	datacenter_a := &db.Datacenter{Id: 1, Name: "a"}
	datacenter_b := &db.Datacenter{Id: 2, Name: "b"}
	datacenter_c := &db.Datacenter{Id: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)

	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller_a
	state.Database.SellerMap[2] = seller_b
	state.Database.SellerMap[3] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.Database.GenerateRelaySecretKeys(routingPublicKey, routingPrivateKey)

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

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK_EncryptedNextRouteTokenSize)

	addresses := make([]net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	secretKeys := make([][]byte, NumTokens)

	secretKeys[0], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, clientPublicKey)
	secretKeys[1], _ = state.Database.RelaySecretKeys[1]
	secretKeys[2], _ = state.Database.RelaySecretKeys[2]
	secretKeys[3], _ = state.Database.RelaySecretKeys[3]
	secretKeys[4], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, serverPublicKey)

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK_EncryptedNextRouteTokenSize]

		result := core.ReadEncryptedRouteToken(&token, tokenData, secretKeys[i])
		assert.True(t, result)
		if !result {
			return
		}

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.EnvelopeKbpsUp, uint32(256))
		assert.Equal(t, token.EnvelopeKbpsDown, uint32(1024))

		if i > 1 {
			assert.Equal(t, token.PrevAddress.String(), addresses[i-1].String())
		}

		if i != 4 {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.SessionPrivateKey {
			if token.SessionPrivateKey[j] != 0 {
				found = true
				break
			}
		}
		assert.True(t, found)
	}
}

func Test_SessionUpdate_BuildNextTokens_InternalAddresses(t *testing.T) {

	t.Parallel()

	// initialize state

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = &serverAddress

	state.Buyer.RouteShader.BandwidthEnvelopeUpKbps = 256
	state.Buyer.RouteShader.BandwidthEnvelopeDownKbps = 1024

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database

	seller := &db.Seller{Id: 1, Name: "seller"}

	datacenter_a := &db.Datacenter{Id: 1, Name: "a"}
	datacenter_b := &db.Datacenter{Id: 2, Name: "b"}
	datacenter_c := &db.Datacenter{Id: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_address_a_internal := core.ParseAddress("35.0.0.1:40002")
	relay_address_b_internal := core.ParseAddress("35.0.0.1:40003")
	relay_address_c_internal := core.ParseAddress("35.0.0.1:40004")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, InternalAddress: relay_address_a_internal, HasInternalAddress: true, Seller: seller, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, InternalAddress: relay_address_b_internal, HasInternalAddress: true, Seller: seller, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, InternalAddress: relay_address_c_internal, HasInternalAddress: true, Seller: seller, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.Database.GenerateRelaySecretKeys(routingPublicKey, routingPrivateKey)

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

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK_EncryptedNextRouteTokenSize)

	addresses := make([]net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b_internal
	addresses[3] = relay_address_c_internal
	addresses[4] = serverAddress

	secretKeys := make([][]byte, NumTokens)

	secretKeys[0], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, clientPublicKey)
	secretKeys[1], _ = state.Database.RelaySecretKeys[1]
	secretKeys[2], _ = state.Database.RelaySecretKeys[2]
	secretKeys[3], _ = state.Database.RelaySecretKeys[3]
	secretKeys[4], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, serverPublicKey)

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK_EncryptedNextRouteTokenSize]

		result := core.ReadEncryptedRouteToken(&token, tokenData, secretKeys[i])
		assert.True(t, result)
		if !result {
			return
		}

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.EnvelopeKbpsUp, uint32(256))
		assert.Equal(t, token.EnvelopeKbpsDown, uint32(1024))

		if i == 0 || i == NumTokens-1 {
			assert.Equal(t, token.PrevInternal, uint8(0))
			assert.Equal(t, token.NextInternal, uint8(0))
		}

		if i == 1 {
			assert.Equal(t, token.PrevInternal, uint8(0))
			assert.Equal(t, token.NextInternal, uint8(1))
		}

		if i == 2 {
			assert.Equal(t, token.PrevInternal, uint8(1))
			assert.Equal(t, token.NextInternal, uint8(1))
		}

		if i == 3 {
			assert.Equal(t, token.PrevInternal, uint8(1))
			assert.Equal(t, token.NextInternal, uint8(0))
		}

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.SessionPrivateKey {
			if token.SessionPrivateKey[j] != 0 {
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

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database

	seller_a := &db.Seller{Id: 1, Name: "a"}
	seller_b := &db.Seller{Id: 2, Name: "b"}
	seller_c := &db.Seller{Id: 3, Name: "c"}

	datacenter_a := &db.Datacenter{Id: 1, Name: "a"}
	datacenter_b := &db.Datacenter{Id: 2, Name: "b"}
	datacenter_c := &db.Datacenter{Id: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller_a
	state.Database.SellerMap[2] = seller_b
	state.Database.SellerMap[3] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.Database.GenerateRelaySecretKeys(routingPublicKey, routingPrivateKey)

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

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK_RouteTypeContinue))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK_EncryptedContinueRouteTokenSize)

	secretKeys := make([][]byte, NumTokens)

	secretKeys[0], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, clientPublicKey)
	secretKeys[1], _ = state.Database.RelaySecretKeys[1]
	secretKeys[2], _ = state.Database.RelaySecretKeys[2]
	secretKeys[3], _ = state.Database.RelaySecretKeys[3]
	secretKeys[4], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, serverPublicKey)

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK_EncryptedContinueRouteTokenSize * i

		token := core.ContinueToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK_EncryptedContinueRouteTokenSize]

		result := core.ReadEncryptedContinueToken(&token, tokenData, secretKeys[i])
		assert.True(t, result)
		if !result {
			return
		}

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
	assert.True(t, (state.Error&constants.SessionError_NoRouteRelays) != 0)
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

	seller_a := &db.Seller{Id: 1, Name: "a"}
	seller_b := &db.Seller{Id: 2, Name: "b"}
	seller_c := &db.Seller{Id: 3, Name: "c"}

	datacenter_a := &db.Datacenter{Id: 1, Name: "a"}
	datacenter_b := &db.Datacenter{Id: 2, Name: "b"}
	datacenter_c := &db.Datacenter{Id: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c}

	state.Database.SellerMap[1] = seller_a
	state.Database.SellerMap[2] = seller_b
	state.Database.SellerMap[3] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	// setup route shader

	state.Buyer.RouteShader = core.NewRouteShader()

	// setup source relays

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

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = &serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := &db.Seller{Id: 1, Name: "a"}
	seller_b := &db.Seller{Id: 1, Name: "b"}
	seller_c := &db.Seller{Id: 1, Name: "c"}

	datacenter_a := &db.Datacenter{Id: 1, Name: "a"}
	datacenter_b := &db.Datacenter{Id: 2, Name: "b"}
	datacenter_c := &db.Datacenter{Id: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller_a
	state.Database.SellerMap[2] = seller_b
	state.Database.SellerMap[3] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.Database.GenerateRelaySecretKeys(routingPublicKey, routingPrivateKey)

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]uint8, entryCount)

	for i := range costMatrix {
		costMatrix[i] = 255
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

	// setup source relays

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

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK_EncryptedNextRouteTokenSize)

	addresses := make([]net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	secretKeys := make([][]byte, NumTokens)

	secretKeys[0], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, clientPublicKey)
	secretKeys[1], _ = state.Database.RelaySecretKeys[1]
	secretKeys[2], _ = state.Database.RelaySecretKeys[2]
	secretKeys[3], _ = state.Database.RelaySecretKeys[3]
	secretKeys[4], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, serverPublicKey)

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK_EncryptedNextRouteTokenSize]

		result := core.ReadEncryptedRouteToken(&token, tokenData, secretKeys[i])
		assert.True(t, result)
		if !result {
			return
		}

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.EnvelopeKbpsUp, uint32(256))
		assert.Equal(t, token.EnvelopeKbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.SessionPrivateKey {
			if token.SessionPrivateKey[j] != 0 {
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

	assert.True(t, (state.Error&constants.SessionError_Aborted) != 0)
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

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = &serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := &db.Seller{Id: 1, Name: "a"}
	seller_b := &db.Seller{Id: 2, Name: "b"}
	seller_c := &db.Seller{Id: 3, Name: "c"}

	datacenter_a := &db.Datacenter{Id: 1, Name: "a"}
	datacenter_b := &db.Datacenter{Id: 2, Name: "b"}
	datacenter_c := &db.Datacenter{Id: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller_a
	state.Database.SellerMap[2] = seller_b
	state.Database.SellerMap[3] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.Database.GenerateRelaySecretKeys(routingPublicKey, routingPrivateKey)

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]uint8, entryCount)

	for i := range costMatrix {
		costMatrix[i] = 255
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

	// setup source relays

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

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK_EncryptedNextRouteTokenSize)

	addresses := make([]net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	secretKeys := make([][]byte, NumTokens)

	secretKeys[0], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, clientPublicKey)
	secretKeys[1], _ = state.Database.RelaySecretKeys[1]
	secretKeys[2], _ = state.Database.RelaySecretKeys[2]
	secretKeys[3], _ = state.Database.RelaySecretKeys[3]
	secretKeys[4], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, serverPublicKey)

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK_EncryptedNextRouteTokenSize]

		result := core.ReadEncryptedRouteToken(&token, tokenData, secretKeys[i])
		assert.True(t, result)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.EnvelopeKbpsUp, uint32(256))
		assert.Equal(t, token.EnvelopeKbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.SessionPrivateKey {
			if token.SessionPrivateKey[j] != 0 {
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

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK_RouteTypeContinue))

	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK_EncryptedContinueRouteTokenSize)

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK_EncryptedContinueRouteTokenSize * i

		token := core.ContinueToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK_EncryptedContinueRouteTokenSize]

		result := core.ReadEncryptedContinueToken(&token, tokenData, secretKeys[i])
		assert.True(t, result)
		if !result {
			return
		}

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

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = &serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := &db.Seller{Id: 1, Name: "a"}
	seller_b := &db.Seller{Id: 2, Name: "b"}
	seller_c := &db.Seller{Id: 3, Name: "c"}

	datacenter_a := &db.Datacenter{Id: 1, Name: "a"}
	datacenter_b := &db.Datacenter{Id: 2, Name: "b"}
	datacenter_c := &db.Datacenter{Id: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller_a
	state.Database.SellerMap[2] = seller_b
	state.Database.SellerMap[3] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.Database.GenerateRelaySecretKeys(routingPublicKey, routingPrivateKey)

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]uint8, entryCount)

	for i := range costMatrix {
		costMatrix[i] = 255
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

	// setup source relays

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

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK_EncryptedNextRouteTokenSize)

	addresses := make([]net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	secretKeys := make([][]byte, NumTokens)

	secretKeys[0], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, clientPublicKey)
	secretKeys[1], _ = state.Database.RelaySecretKeys[1]
	secretKeys[2], _ = state.Database.RelaySecretKeys[2]
	secretKeys[3], _ = state.Database.RelaySecretKeys[3]
	secretKeys[4], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, serverPublicKey)

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK_EncryptedNextRouteTokenSize]

		result := core.ReadEncryptedRouteToken(&token, tokenData, secretKeys[i])
		assert.True(t, result)
		if !result {
			return
		}

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.EnvelopeKbpsUp, uint32(256))
		assert.Equal(t, token.EnvelopeKbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.SessionPrivateKey {
			if token.SessionPrivateKey[j] != 0 {
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

	costMatrix = make([]uint8, entryCount)

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

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK_RouteTypeNew))

	const NewTokens = 4

	assert.Equal(t, state.Response.NumTokens, int32(NewTokens))
	assert.Equal(t, len(state.Response.Tokens), NewTokens*packets.SDK_EncryptedNextRouteTokenSize)

	addresses = make([]net.UDPAddr, NewTokens)
	addresses[1] = relay_address_b
	addresses[2] = relay_address_c
	addresses[3] = serverAddress

	secretKeys = make([][]byte, NewTokens)

	secretKeys[0], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, clientPublicKey)
	secretKeys[1], _ = state.Database.RelaySecretKeys[2]
	secretKeys[2], _ = state.Database.RelaySecretKeys[3]
	secretKeys[3], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, serverPublicKey)

	for i := 0; i < NewTokens; i++ {

		index := packets.SDK_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK_EncryptedNextRouteTokenSize]

		result := core.ReadEncryptedRouteToken(&token, tokenData, secretKeys[i])
		assert.True(t, result)
		if !result {
			return
		}

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.EnvelopeKbpsUp, uint32(256))
		assert.Equal(t, token.EnvelopeKbpsDown, uint32(1024))

		if i == 3 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.SessionPrivateKey {
			if token.SessionPrivateKey[j] != 0 {
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

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = &serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := &db.Seller{Id: 1, Name: "a"}
	seller_b := &db.Seller{Id: 2, Name: "b"}
	seller_c := &db.Seller{Id: 3, Name: "c"}

	datacenter_a := &db.Datacenter{Id: 1, Name: "a"}
	datacenter_b := &db.Datacenter{Id: 2, Name: "b"}
	datacenter_c := &db.Datacenter{Id: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller_a
	state.Database.SellerMap[2] = seller_b
	state.Database.SellerMap[3] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.Database.GenerateRelaySecretKeys(routingPublicKey, routingPrivateKey)

	state.DestRelays = []int32{0, 1, 2}

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]uint8, entryCount)

	for i := range costMatrix {
		costMatrix[i] = 255
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

	// setup source relays

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

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK_EncryptedNextRouteTokenSize)

	addresses := make([]net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	secretKeys := make([][]byte, NumTokens)

	secretKeys[0], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, clientPublicKey)
	secretKeys[1], _ = state.Database.RelaySecretKeys[1]
	secretKeys[2], _ = state.Database.RelaySecretKeys[2]
	secretKeys[3], _ = state.Database.RelaySecretKeys[3]
	secretKeys[4], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, serverPublicKey)

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK_EncryptedNextRouteTokenSize]

		result := core.ReadEncryptedRouteToken(&token, tokenData, secretKeys[i])
		assert.True(t, result)
		if !result {
			return
		}

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.EnvelopeKbpsUp, uint32(256))
		assert.Equal(t, token.EnvelopeKbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.SessionPrivateKey {
			if token.SessionPrivateKey[j] != 0 {
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

	assert.True(t, (state.Error&constants.SessionError_RouteRelayNoLongerExists) != 0)
	assert.True(t, (state.Error&constants.SessionError_RouteNoLongerExists) != 0)
}

func Test_SessionUpdate_MakeRouteDecision_RouteNoLongerExists_ClientRelays(t *testing.T) {

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = false
	state.Request.DirectRTT = 100
	state.Request.SliceNumber = 100
	state.Debug = new(string)

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = &serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := &db.Seller{Id: 1, Name: "a"}
	seller_b := &db.Seller{Id: 2, Name: "b"}
	seller_c := &db.Seller{Id: 3, Name: "c"}

	datacenter_a := &db.Datacenter{Id: 1, Name: "a"}
	datacenter_b := &db.Datacenter{Id: 2, Name: "b"}
	datacenter_c := &db.Datacenter{Id: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller_a
	state.Database.SellerMap[2] = seller_b
	state.Database.SellerMap[3] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.Database.GenerateRelaySecretKeys(routingPublicKey, routingPrivateKey)

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]uint8, entryCount)

	for i := range costMatrix {
		costMatrix[i] = 255
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

	// setup source relays

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

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK_EncryptedNextRouteTokenSize)

	addresses := make([]net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	secretKeys := make([][]byte, NumTokens)

	secretKeys[0], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, clientPublicKey)
	secretKeys[1], _ = state.Database.RelaySecretKeys[1]
	secretKeys[2], _ = state.Database.RelaySecretKeys[2]
	secretKeys[3], _ = state.Database.RelaySecretKeys[3]
	secretKeys[4], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, serverPublicKey)

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK_EncryptedNextRouteTokenSize]

		result := core.ReadEncryptedRouteToken(&token, tokenData, secretKeys[i])
		assert.True(t, result)

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.EnvelopeKbpsUp, uint32(256))
		assert.Equal(t, token.EnvelopeKbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.SessionPrivateKey {
			if token.SessionPrivateKey[j] != 0 {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	// verify debug string is set

	assert.NotEqual(t, *state.Debug, "")

	// make all source relays unroutable

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

	assert.False(t, (state.Error&constants.SessionError_RouteRelayNoLongerExists) != 0)
	assert.True(t, (state.Error&constants.SessionError_RouteNoLongerExists) != 0)
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

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = &serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := &db.Seller{Id: 1, Name: "a"}
	seller_b := &db.Seller{Id: 2, Name: "b"}
	seller_c := &db.Seller{Id: 3, Name: "c"}

	datacenter_a := &db.Datacenter{Id: 1, Name: "a"}
	datacenter_b := &db.Datacenter{Id: 2, Name: "b"}
	datacenter_c := &db.Datacenter{Id: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller_a
	state.Database.SellerMap[2] = seller_b
	state.Database.SellerMap[3] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.Database.GenerateRelaySecretKeys(routingPublicKey, routingPrivateKey)

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]uint8, entryCount)

	for i := range costMatrix {
		costMatrix[i] = 255
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

	// setup source relays

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

	assert.Equal(t, state.Response.RouteType, int32(packets.SDK_RouteTypeNew))
	assert.Equal(t, state.Response.NumTokens, int32(NumTokens))
	assert.Equal(t, len(state.Response.Tokens), NumTokens*packets.SDK_EncryptedNextRouteTokenSize)

	addresses := make([]net.UDPAddr, NumTokens)
	addresses[1] = relay_address_a
	addresses[2] = relay_address_b
	addresses[3] = relay_address_c
	addresses[4] = serverAddress

	secretKeys := make([][]byte, NumTokens)

	secretKeys[0], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, clientPublicKey)
	secretKeys[1], _ = state.Database.RelaySecretKeys[1]
	secretKeys[2], _ = state.Database.RelaySecretKeys[2]
	secretKeys[3], _ = state.Database.RelaySecretKeys[3]
	secretKeys[4], _ = crypto.SecretKey_GenerateRemote(routingPublicKey, routingPrivateKey, serverPublicKey)

	for i := 0; i < NumTokens; i++ {

		index := packets.SDK_EncryptedNextRouteTokenSize * i

		token := core.RouteToken{}

		tokenData := state.Response.Tokens[index : index+packets.SDK_EncryptedNextRouteTokenSize]

		result := core.ReadEncryptedRouteToken(&token, tokenData, secretKeys[i])
		assert.True(t, result)
		if !result {
			return
		}

		assert.Equal(t, token.ExpireTimestamp, state.Output.ExpireTimestamp)
		assert.Equal(t, token.SessionId, state.Output.SessionId)
		assert.Equal(t, token.SessionVersion, uint8(state.Output.SessionVersion))
		assert.Equal(t, token.EnvelopeKbpsUp, uint32(256))
		assert.Equal(t, token.EnvelopeKbpsDown, uint32(1024))

		if i == 4 {
			assert.Nil(t, nil)
		} else {
			assert.Equal(t, token.NextAddress.String(), addresses[i+1].String())
		}

		found := false
		for j := range token.SessionPrivateKey {
			if token.SessionPrivateKey[j] != 0 {
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

	costMatrix = make([]uint8, entryCount)

	for i := range costMatrix {
		costMatrix[i] = 255
	}

	costMatrix[core.TriMatrixIndex(0, 2)] = 254

	state.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:], state.Database)

	// make route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// validate that we tripped "route no longer exists"

	assert.False(t, (state.Error&constants.SessionError_RouteRelayNoLongerExists) != 0)
	assert.True(t, (state.Error&constants.SessionError_RouteNoLongerExists) != 0)
	assert.False(t, state.Output.RouteState.Next)
}

func Test_SessionUpdate_MakeRouteDecision_Mispredict(t *testing.T) {

	if core.Relax {
		return
	}

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = false
	state.Request.DirectRTT = 100
	state.Request.SliceNumber = 100
	state.Debug = new(string)

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = &serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := &db.Seller{Id: 1, Name: "a"}
	seller_b := &db.Seller{Id: 2, Name: "b"}
	seller_c := &db.Seller{Id: 3, Name: "c"}

	datacenter_a := &db.Datacenter{Id: 1, Name: "a"}
	datacenter_b := &db.Datacenter{Id: 2, Name: "b"}
	datacenter_c := &db.Datacenter{Id: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller_a
	state.Database.SellerMap[2] = seller_b
	state.Database.SellerMap[3] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.Database.GenerateRelaySecretKeys(routingPublicKey, routingPrivateKey)

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]uint8, entryCount)

	for i := range costMatrix {
		costMatrix[i] = 255
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

	// setup source relays

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
			assert.False(t, state.Output.RouteState.Mispredict)
		}
	}

	// validate that we tripped "mispredict"

	assert.True(t, state.Output.RouteState.Mispredict)
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

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, _ := crypto.Box_KeyPair()

	serverPublicKey, _ := crypto.Box_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	copy(state.Request.ClientRoutePublicKey[:], clientPublicKey)
	copy(state.Request.ServerRoutePublicKey[:], serverPublicKey)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	state.From = &serverAddress

	state.Output.SessionId = 0x123457
	state.Output.SessionVersion = 100

	// initialize database with three relays

	seller_a := &db.Seller{Id: 1, Name: "a"}
	seller_b := &db.Seller{Id: 2, Name: "b"}
	seller_c := &db.Seller{Id: 3, Name: "c"}

	datacenter_a := &db.Datacenter{Id: 1, Name: "a"}
	datacenter_b := &db.Datacenter{Id: 2, Name: "b"}
	datacenter_c := &db.Datacenter{Id: 3, Name: "c"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller_a
	state.Database.SellerMap[2] = seller_b
	state.Database.SellerMap[3] = seller_c

	state.Database.DatacenterMap[1] = datacenter_a
	state.Database.DatacenterMap[2] = datacenter_b
	state.Database.DatacenterMap[3] = datacenter_c

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.Database.GenerateRelaySecretKeys(routingPublicKey, routingPrivateKey)

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]uint8, entryCount)

	for i := range costMatrix {
		costMatrix[i] = 255
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

	// setup source relays

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

	// make all source relays very expensive

	state.SourceRelayRTT = []int32{100, 100, 100}

	// make route decision

	state.Request.Next = true
	state.Request.NextRTT = 100
	state.Request.DirectRTT = 1

	state.Input = state.Output

	handlers.SessionUpdate_MakeRouteDecision(state)

	// validate that we tripped "latency worse"

	assert.True(t, state.Output.RouteState.LatencyWorse)
	assert.False(t, state.Output.RouteState.Next)
}

// --------------------------------------------------------------

func Test_SessionUpdate_UpdateClientRelays_AnalysisOnly(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.RouteShader.AnalysisOnly = true

	result := handlers.SessionUpdate_UpdateClientRelays(state)

	assert.False(t, result)
	assert.True(t, state.NotUpdatingClientRelaysAnalysisOnly)
}

func Test_SessionUpdate_UpdateClientRelays_DatacenterNotEnabled(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Error |= constants.SessionError_DatacenterNotEnabled

	result := handlers.SessionUpdate_UpdateClientRelays(state)

	assert.False(t, result)
	assert.True(t, state.NotUpdatingClientRelaysDatacenterNotEnabled)
}

func Test_SessionUpdate_UpdateClientRelays(t *testing.T) {

	t.Parallel()

	state := CreateState()

	// initialize database with three relays

	seller := &db.Seller{Id: 1, Name: "seller"}

	datacenter := &db.Datacenter{Id: 1, Name: "datacenter"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller

	state.Database.DatacenterMap[1] = datacenter

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.DestRelayIds = []uint64{1, 2, 3}

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]uint8, entryCount)

	for i := range costMatrix {
		costMatrix[i] = 255
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

	state.RouteMatrix.RelayAddresses[0] = relay_address_a
	state.RouteMatrix.RelayAddresses[1] = relay_address_b
	state.RouteMatrix.RelayAddresses[2] = relay_address_c

	// setup client relays

	state.Request.NumClientRelays = 3
	state.Request.ClientRelayPingsHaveChanged = true
	copy(state.Request.ClientRelayIds[:], []uint64{1, 2, 3})
	copy(state.Request.ClientRelayRTT[:], []int32{10, 20, 30})
	copy(state.Request.ClientRelayJitter[:], []int32{0, 0, 0})
	copy(state.Request.ClientRelayPacketLoss[:], []float32{0, 0, 0})

	// update client relays

	state.Input.SliceNumber = 1

	result := handlers.SessionUpdate_UpdateClientRelays(state)

	// validate

	assert.True(t, result)
	assert.False(t, state.NotUpdatingClientRelaysAnalysisOnly)
	assert.False(t, state.NotUpdatingClientRelaysDatacenterNotEnabled)

	assert.Equal(t, len(state.SourceRelays), 3)
	assert.Equal(t, len(state.SourceRelayRTT), 3)

	assert.Equal(t, state.SourceRelays[0], int32(0))
	assert.Equal(t, state.SourceRelays[1], int32(1))
	assert.Equal(t, state.SourceRelays[2], int32(2))

	assert.Equal(t, state.SourceRelayRTT[0], int32(10))
	assert.Equal(t, state.SourceRelayRTT[1], int32(20))
	assert.Equal(t, state.SourceRelayRTT[2], int32(30))
}

// --------------------------------------------------------------

func Test_SessionUpdate_UpdateServerRelays(t *testing.T) {

	t.Parallel()

	state := CreateState()

	// initialize database with three relays in the same datacenter

	seller := &db.Seller{Id: 1, Name: "seller"}

	datacenter := &db.Datacenter{Id: 1, Name: "datacenter"}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	state.Database.Relays = make([]db.Relay, 3)
	state.Database.Relays[0] = db.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller, PublicKey: relay_public_key_a}
	state.Database.Relays[1] = db.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller, PublicKey: relay_public_key_b}
	state.Database.Relays[2] = db.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller, PublicKey: relay_public_key_c}

	state.Database.SellerMap[1] = seller

	state.Database.DatacenterMap[1] = datacenter

	state.Database.RelayMap[1] = &state.Database.Relays[0]
	state.Database.RelayMap[2] = &state.Database.Relays[1]
	state.Database.RelayMap[3] = &state.Database.Relays[2]

	state.DestRelayIds = []uint64{1, 2, 3}

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)

	costMatrix := make([]uint8, entryCount)

	for i := range costMatrix {
		costMatrix[i] = 255
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

	state.RouteMatrix.RelayAddresses[0] = relay_address_a
	state.RouteMatrix.RelayAddresses[1] = relay_address_b
	state.RouteMatrix.RelayAddresses[2] = relay_address_c

	// setup client relays

	state.Request.NumClientRelays = 3
	state.Request.ClientRelayPingsHaveChanged = true
	copy(state.Request.ClientRelayIds[:], []uint64{1, 2, 3})
	copy(state.Request.ClientRelayRTT[:], []int32{10, 20, 30})
	copy(state.Request.ClientRelayJitter[:], []int32{0, 0, 0})
	copy(state.Request.ClientRelayPacketLoss[:], []float32{0, 0, 0})

	// update client relays

	state.Input.SliceNumber = 1

	result := handlers.SessionUpdate_UpdateClientRelays(state)

	// validate

	assert.True(t, result)
	assert.False(t, state.NotUpdatingClientRelaysAnalysisOnly)
	assert.False(t, state.NotUpdatingClientRelaysDatacenterNotEnabled)

	assert.Equal(t, len(state.SourceRelays), 3)
	assert.Equal(t, len(state.SourceRelayRTT), 3)

	assert.Equal(t, state.SourceRelays[0], int32(0))
	assert.Equal(t, state.SourceRelays[1], int32(1))
	assert.Equal(t, state.SourceRelays[2], int32(2))

	assert.Equal(t, state.SourceRelayRTT[0], int32(10))
	assert.Equal(t, state.SourceRelayRTT[1], int32(20))
	assert.Equal(t, state.SourceRelayRTT[2], int32(30))
}

// --------------------------------------------------------------

func Test_SessionUpdate_Post_SliceZero(t *testing.T) {

	t.Parallel()

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	var serverBackendPublicKey [packets.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var serverBackendPrivateKey [packets.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK_SignKeypair(serverBackendPublicKey[:], serverBackendPublicKey[:])

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	state.ServerBackendPublicKey = serverBackendPublicKey[:]
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	from := core.ParseAddress("127.0.0.1:40000")
	state.From = &from
	serverBackendAddress := core.ParseAddress("127.0.0.1:50000")
	state.ServerBackendAddress = &serverBackendAddress

	state.Request.SliceNumber = 0

	handlers.SessionUpdate_Post(state)
}

func Test_SessionUpdate_Post_DurationOnNext(t *testing.T) {

	t.Parallel()

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	state.ServerBackendPublicKey = serverBackendPublicKey[:]
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	from := core.ParseAddress("127.0.0.1:40000")
	state.From = &from
	serverBackendAddress := core.ParseAddress("127.0.0.1:50000")
	state.ServerBackendAddress = &serverBackendAddress

	state.Request.SliceNumber = 1

	sessionData := packets.GenerateRandomSessionData()
	sessionData.WroteSummary = false
	sessionData.RouteState.Next = true
	writeSessionData := WriteSessionData(sessionData)
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))
	state.Request.SessionDataBytes = int32(len(writeSessionData))

	handlers.SessionUpdate_Pre(state)

	handlers.SessionUpdate_Post(state)

	assert.Equal(t, state.Output.DurationOnNext, uint32(packets.SDK_SliceSeconds))
}

func Test_SessionUpdate_Post_PacketsSentPacketsLost(t *testing.T) {

	t.Parallel()

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	state.ServerBackendPublicKey = serverBackendPublicKey[:]
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	from := core.ParseAddress("127.0.0.1:40000")
	state.From = &from
	serverBackendAddress := core.ParseAddress("127.0.0.1:50000")
	state.ServerBackendAddress = &serverBackendAddress

	state.Request.SliceNumber = 2

	state.Request.PacketsSentClientToServer = 10001
	state.Request.PacketsSentServerToClient = 10002
	state.Request.PacketsLostClientToServer = 10003
	state.Request.PacketsLostServerToClient = 10004

	sessionData := packets.GenerateRandomSessionData()
	sessionData.WroteSummary = false
	sessionData.RouteState.Next = true
	writeSessionData := WriteSessionData(sessionData)
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))
	state.Request.SessionDataBytes = int32(len(writeSessionData))

	handlers.SessionUpdate_Post(state)

	assert.Equal(t, state.Output.PrevPacketsSentClientToServer, state.Request.PacketsSentClientToServer)
	assert.Equal(t, state.Output.PrevPacketsSentServerToClient, state.Request.PacketsSentServerToClient)
	assert.Equal(t, state.Output.PrevPacketsLostClientToServer, state.Request.PacketsLostClientToServer)
	assert.Equal(t, state.Output.PrevPacketsLostServerToClient, state.Request.PacketsLostServerToClient)
}

func Test_SessionUpdate_Post_WriteSummary(t *testing.T) {

	t.Parallel()

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	state.ServerBackendPublicKey = serverBackendPublicKey[:]
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	from := core.ParseAddress("127.0.0.1:40000")
	state.From = &from
	serverBackendAddress := core.ParseAddress("127.0.0.1:50000")
	state.ServerBackendAddress = &serverBackendAddress

	state.Request.SliceNumber = 100
	state.Request.ClientPingTimedOut = true

	sessionData := packets.GenerateRandomSessionData()
	writeSessionData := WriteSessionData(sessionData)
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))
	state.Request.SessionDataBytes = int32(len(writeSessionData))

	handlers.SessionUpdate_Post(state)

	assert.True(t, state.Output.WriteSummary)
	assert.False(t, state.Output.WroteSummary)
}

func Test_SessionUpdate_Post_WroteSummary(t *testing.T) {

	t.Parallel()

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	serverBackendPublicKey, serverBackendPrivateKey := crypto.Sign_KeyPair()

	state.RelayBackendPublicKey = routingPublicKey
	state.RelayBackendPrivateKey = routingPrivateKey
	state.ServerBackendPublicKey = serverBackendPublicKey[:]
	state.ServerBackendPrivateKey = serverBackendPrivateKey[:]

	from := core.ParseAddress("127.0.0.1:40000")
	state.From = &from
	serverBackendAddress := core.ParseAddress("127.0.0.1:50000")
	state.ServerBackendAddress = &serverBackendAddress

	state.Request.SliceNumber = 100
	state.Request.ClientPingTimedOut = true
	state.Output.WriteSummary = true

	sessionData := packets.GenerateRandomSessionData()
	writeSessionData := WriteSessionData(sessionData)
	copy(state.Request.SessionData[:], writeSessionData)
	copy(state.Request.SessionDataSignature[:], crypto.Sign(writeSessionData, state.ServerBackendPrivateKey))
	state.Request.SessionDataBytes = int32(len(writeSessionData))

	handlers.SessionUpdate_Post(state)

	assert.False(t, state.Output.WriteSummary)
	assert.True(t, state.Output.WroteSummary)
}

// --------------------------------------------------------------
