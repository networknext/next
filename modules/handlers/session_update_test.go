package handlers_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/database"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/handlers"
	"github.com/networknext/backend/modules/packets"
	db "github.com/networknext/backend/modules/database"

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
	state.Database = database.CreateDatabase()
	return &state
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
	state.LocateIP = FailLocateIP

	result := handlers.SessionUpdate_Pre(state)

	assert.True(t, result)
	assert.True(t, state.LocationVeto)
}

func Test_SessionUpdate_Pre_ReadLocation(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.SliceNumber = 1

	sessionData := packets.SDK5_SessionData{}

	sessionData.Version = packets.SDK5_SessionDataVersion_Write
	sessionData.Location.Version = packets.SDK5_LocationVersion_Write
	sessionData.Location.Latitude = 10
	sessionData.Location.Longitude = 20
	sessionData.Location.ISP = "Starlink"
	sessionData.Location.ASN = 5

	buffer := make([]byte, packets.SDK5_MaxPacketBytes)
	writeStream := encoding.CreateWriteStream(buffer[:])
	err := sessionData.Serialize(writeStream)
	assert.Nil(t, err)
	writeStream.Flush()
	packetBytes := writeStream.GetBytesProcessed()
	packetData := buffer[:packetBytes]

	state.Request.SessionDataBytes = int32(packetBytes)
	copy(state.Request.SessionData[:], packetData)

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

	state.RouteMatrix.CreatedAt = 0

	result := handlers.SessionUpdate_Pre(state)

	assert.True(t, result)
	assert.True(t, state.StaleRouteMatrix)
}

func Test_SessionUpdate_Pre_KnownDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.DatacenterId = 0x12345

	state.Database.DatacenterMap[0x12345] = database.Datacenter{}

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
}

func Test_SessionUpdate_Pre_UnknownDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.DatacenterId = 0x12345

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.True(t, state.UnknownDatacenter)
}

func Test_SessionUpdate_Pre_DatacenterNotEnabled(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.ID = 0x11111
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = database.Datacenter{}

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
	assert.True(t, state.DatacenterNotEnabled)
}

func Test_SessionUpdate_Pre_DatacenterEnabled(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.ID = 0x11111
	state.Request.BuyerId = 0x11111
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = database.Datacenter{}
	state.Database.DatacenterMaps[state.Buyer.ID] = make(map[uint64]database.DatacenterMap)
	state.Database.DatacenterMaps[state.Buyer.ID][state.Request.DatacenterId] = database.DatacenterMap{}

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
	assert.False(t, state.DatacenterNotEnabled)
}

func Test_SessionUpdate_Pre_FailedToReadSessionData(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.SliceNumber = 1

	result := handlers.SessionUpdate_Pre(state)

	assert.True(t, result)
	assert.True(t, state.FailedToReadSessionData)
}

func Test_SessionUpdate_Pre_NoRelaysInDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.ID = 0x11111
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = database.Datacenter{}

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
	assert.True(t, state.NoRelaysInDatacenter)
}

func Test_SessionUpdate_Pre_RelaysInDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.ID = 0x11111
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = database.Datacenter{}

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

func Test_SessionUpdate_Pre_Pro(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.NumTags = 1
	state.Request.Tags[0] = common.HashTag("pro")

	result := handlers.SessionUpdate_Pre(state)

	assert.False(t, result)
	assert.True(t, state.Pro)
}

func Test_SessionUpdate_Pre_Debug(t *testing.T) {

	t.Parallel()

	state := CreateState()

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

	sessionData := packets.GenerateRandomSessionData()

	copy(state.Request.SessionData[:], writeSessionData(sessionData))

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, state.FailedToReadSessionData)
}

func Test_SessionUpdate_ExistingSession_BadSessionId(t *testing.T) {

	t.Parallel()

	state := CreateState()

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber

	copy(state.Request.SessionData[:], writeSessionData(sessionData))

	handlers.SessionUpdate_ExistingSession(state)

	assert.True(t, state.ReadSessionData)
	assert.False(t, state.FailedToReadSessionData)
	assert.True(t, state.BadSessionId)
	assert.False(t, state.BadSliceNumber)
}

func Test_SessionUpdate_ExistingSession_BadSliceNumber(t *testing.T) {

	t.Parallel()

	state := CreateState()

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber

	copy(state.Request.SessionData[:], writeSessionData(sessionData))

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

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber

	copy(state.Request.SessionData[:], writeSessionData(sessionData))

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

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber

	copy(state.Request.SessionData[:], writeSessionData(sessionData))

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
	assert.Equal(t, state.Output.ExpireTimestamp, state.Input.ExpireTimestamp + packets.SDK5_BillingSliceSeconds)
}

func Test_SessionUpdate_ExistingSession_RealPacketLoss(t *testing.T) {

	t.Parallel()

	state := CreateState()

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber
	sessionData.PrevPacketsSentClientToServer = 1000
	sessionData.PrevPacketsSentServerToClient = 1000
	sessionData.PrevPacketsLostClientToServer = 0
	sessionData.PrevPacketsLostServerToClient = 0

	copy(state.Request.SessionData[:], writeSessionData(sessionData))

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
	assert.Equal(t, state.Output.ExpireTimestamp, state.Input.ExpireTimestamp + packets.SDK5_BillingSliceSeconds)

	assert.Equal(t, state.RealPacketLoss, float32(10.0))
	assert.Equal(t, state.PostRealPacketLossServerToClient, float32(5.0))
	assert.Equal(t, state.PostRealPacketLossClientToServer, float32(10.0))
}

func Test_SessionUpdate_ExistingSession_RealJitter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	sessionId := uint64(0x1234556134512)
	sliceNumber := uint32(100)

	sessionData := packets.GenerateRandomSessionData()
	sessionData.SessionId = sessionId
	sessionData.SliceNumber = sliceNumber

	copy(state.Request.SessionData[:], writeSessionData(sessionData))

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
	assert.Equal(t, state.Output.ExpireTimestamp, state.Input.ExpireTimestamp + packets.SDK5_BillingSliceSeconds)

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

// todo: SessionUpdate_GetNearRelays

// --------------------------------------------------------------

// todo: SessionUpdate_UpdateNearRelays

// --------------------------------------------------------------

// todo: SessionUpdate_FilterNearRelays

// --------------------------------------------------------------

func Test_SessionUpdate_BuildNextTokens_PublicAddresses(t *testing.T) {

	t.Parallel()

	// initialize state

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.GenerateRoutingKeyPair()	

	copy(state.RoutingPrivateKey, routingPrivateKey)

	_ = routingPublicKey

	// initialize database

	seller_a := db.Seller{ID: "a", Name: "a"}
	seller_b := db.Seller{ID: "b", Name: "b"}
	seller_c := db.Seller{ID: "c", Name: "c"}

	datacenter_a := db.Datacenter{ID: 1, Name: "a"}
	datacenter_b := db.Datacenter{ID: 2, Name: "b"}
	datacenter_c := db.Datacenter{ID: 3, Name: "c"}

	// todo: we need keypairs for each relay

	relay_a := db.Relay{ID: 1, Name: "a", Addr: core.ParseAddress("127.0.0.1:40000"), Seller: seller_a} // todo: set relay public key
	relay_b := db.Relay{ID: 2, Name: "a", Addr: core.ParseAddress("127.0.0.1:40001"), Seller: seller_b}
	relay_c := db.Relay{ID: 3, Name: "a", Addr: core.ParseAddress("127.0.0.1:40002"), Seller: seller_c}

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

	// todo: build route matrix with relays a,b,c and stick it in state

	// initialize route relays

	routeNumRelays := int32(3)
	routeRelays := []int32{1,2,3}

	// build next tokens

	SessionUpdate_BuildNextTokens(state, routeNumRelays, routeRelays)

	// validate

	// todo: actually decrypt the tokens and verify they contain what we expect
}

func Test_SessionUpdate_BuildNextTokens_PrivateAddresses(t *testing.T) {

	t.Parallel()

	state := CreateState()

	_ = state

	// todo: same as above, but we're going to have two relays using the internal address for comms w. same supplier
}

// --------------------------------------------------------------

// todo: SessionUpdate_BuildContinueTokens

// --------------------------------------------------------------

// todo: SessionUpdate_MakeRouteDecision

// --------------------------------------------------------------

// todo: SessionUpdate_Post

// --------------------------------------------------------------
