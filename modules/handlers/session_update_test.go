package handlers_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/crypto"
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

func Test_SessionUpdate_BuildNextTokens_PublicAddresses(t *testing.T) {

	t.Parallel()

	// initialize state

	state := CreateState()

	routingPublicKey, routingPrivateKey := crypto.Box_KeyPair()

	clientPublicKey, clientPrivateKey := crypto.Box_KeyPair()

	serverPublicKey, serverPrivateKey := crypto.Box_KeyPair()

	copy(state.RoutingPrivateKey[:], routingPrivateKey)
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
	routeRelays := []int32{0,1,2}

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

		tokenData := state.Response.Tokens[index:index+packets.SDK5_EncryptedNextRouteTokenSize]

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

	copy(state.RoutingPrivateKey[:], routingPrivateKey)
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
	routeRelays := []int32{0,1,2}

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

		tokenData := state.Response.Tokens[index:index+packets.SDK5_EncryptedNextRouteTokenSize]

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

	copy(state.RoutingPrivateKey[:], routingPrivateKey)
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
	routeRelays := []int32{0,1,2}

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

		tokenData := state.Response.Tokens[index:index+packets.SDK5_EncryptedContinueRouteTokenSize]

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
	state.Request.DirectMinRTT = 100
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

	state.NumNearRelays = 3

	state.NearRelayIndices[0] = 0
	state.NearRelayIndices[1] = 1
	state.NearRelayIndices[2] = 2

	state.NearRelayRTTs[0] = 10
	state.NearRelayRTTs[1] = 10
	state.NearRelayRTTs[2] = 10

	// setup dest relays

	state.NumDestRelays = 3

	state.DestRelays = make([]int32, state.NumDestRelays)

	state.DestRelays[0] = 0
	state.DestRelays[1] = 1
	state.DestRelays[2] = 2

	// make the route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// verify

	assert.True(t, state.StayDirect)
	assert.False(t, state.TakeNetworkNext)
	assert.False(t, state.Output.RouteState.Next)
}

func generateRouteMatrix(numRelays int, costMatrix []int32, relayDatacenters []uint64, database *db.Database) *common.RouteMatrix {

	numSegments := numRelays

	routeEntries := core.Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters[:])

	destRelays := make([]bool, numRelays)
	for i := range destRelays {
		destRelays[i] = true
	}	

	routeMatrix := &common.RouteMatrix{}

	routeMatrix.CreatedAt = uint64(time.Now().Unix())
	routeMatrix.Version = common.RouteMatrixVersion_Write
	routeMatrix.RouteEntries = routeEntries
	routeMatrix.RelayDatacenterIds = relayDatacenters
	routeMatrix.FullRelayIds = make([]uint64, 0)
	routeMatrix.FullRelayIndexSet = make(map[int32]bool)
	routeMatrix.DestRelays = destRelays

	// todo
	/*
		RelayIds           []uint64
		RelayIdToIndex     map[uint64]int32
		RelayAddresses     []net.UDPAddr // external IPs only
		RelayNames         []string
	*/

	return routeMatrix
}

func Test_SessionUpdate_MakeRouteDecision_TakeNetworkNext(t *testing.T) {

	t.Parallel()

	// setup state

	state := CreateState()

	state.Input.RouteState.Next = false
	state.Request.DirectMinRTT = 100
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

	// setup cost matrix with route through relays a -> b -> c

	const NumRelays = 3

	entryCount := core.TriMatrixLength(NumRelays)
	
	costMatrix := make([]int32, entryCount)
	
	costMatrix[core.TriMatrixIndex(0,1)] = 10
	costMatrix[core.TriMatrixIndex(1,2)] = 10
	costMatrix[core.TriMatrixIndex(0,2)] = 50

	// generate route matrix

	relayDatacenters := [...]uint64{1,2,3}

	state.RouteMatrix = generateRouteMatrix(NumRelays, costMatrix, relayDatacenters[:], state.Database)

	// setup route shader

	state.Buyer.RouteShader = core.NewRouteShader()

	// setup internal config

	state.Buyer.InternalConfig = core.NewInternalConfig()

	// setup near relays

	state.NumNearRelays = 3

	state.NearRelayIndices[0] = 0
	state.NearRelayIndices[1] = 1
	state.NearRelayIndices[2] = 2

	state.NearRelayRTTs[0] = 10
	state.NearRelayRTTs[1] = 10
	state.NearRelayRTTs[2] = 10

	// setup dest relays

	state.NumDestRelays = 3

	state.DestRelays = make([]int32, state.NumDestRelays)

	state.DestRelays[0] = 0
	state.DestRelays[1] = 1
	state.DestRelays[2] = 2

	// make the route decision

	handlers.SessionUpdate_MakeRouteDecision(state)

	// verify

	assert.True(t, state.TakeNetworkNext)
	assert.True(t, state.Output.RouteState.Next)

	// todo: verify route is as expected, rest of output etc...
}

/*
	outputs:

		state.Response.Committed = state.Output.RouteState.Committed
		state.Response.Multipath = state.Output.RouteState.Multipath

		state.Output.RouteCost = routeCost
		state.Output.RouteChanged = routeChanged
		state.Output.RouteNumRelays = routeNumRelays

		for i := int32(0); i < routeNumRelays; i++ {
			relayId := state.RouteMatrix.RelayIds[routeRelays[i]]
			state.Output.RouteRelayIds[i] = relayId
		}
*/

// todo: Test_SessionUpdate_MakeRouteDecision_TakeNetworkNext       (debug true... verify debug string...)

// todo: Test_SessionUpdate_MakeRouteDecision_Aborted

// todo: Test_SessionUpdate_MakeRouteDecision_RouteRelayNoLongerExists

// todo: Test_SessionUpdate_MakeRouteDecision_RouteChanged

// todo: Test_SessionUpdate_MakeRouteDecision_RouteContinued

// todo: Test_SessionUpdate_MakeRouteDecision_RouteContinued

// todo: Test_SessionUpdate_MakeRouteDecision_RouteNoLongerExists

// todo: Test_SessionUpdate_MakeRouteDecision_Mispredict

// todo: Test_SessionUpdate_MakeRouteDecision_LatencyWorse

// --------------------------------------------------------------

// todo: SessionUpdate_Post

// --------------------------------------------------------------

// todo: SessionUpdate_GetNearRelays

// --------------------------------------------------------------

// todo: SessionUpdate_UpdateNearRelays

// --------------------------------------------------------------

// todo: SessionUpdate_FilterNearRelays
