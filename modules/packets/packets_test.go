package packets_test

import (
	"fmt"
	"testing"
	"math/rand"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/packets"

	"github.com/networknext/backend/modules-old/crypto"

	"github.com/stretchr/testify/assert"
)

// ------------------------------------------------------------------------

func TestVersionCompare(t *testing.T) {

	t.Parallel()

	t.Run("equal", func(t *testing.T) {
		a := packets.SDKVersion{1, 2, 3}
		b := packets.SDKVersion{1, 2, 3}

		assert.Equal(t, packets.SDKVersionEqual, a.Compare(b))
	})

	t.Run("older", func(t *testing.T) {
		a := packets.SDKVersion{1, 1, 1}
		b := packets.SDKVersion{2, 0, 0}

		assert.Equal(t, packets.SDKVersionOlder, a.Compare(b))

		a = packets.SDKVersion{1, 1, 1}
		b = packets.SDKVersion{1, 2, 0}

		assert.Equal(t, packets.SDKVersionOlder, a.Compare(b))

		a = packets.SDKVersion{1, 1, 1}
		b = packets.SDKVersion{1, 1, 2}

		assert.Equal(t, packets.SDKVersionOlder, a.Compare(b))
	})

	t.Run("newer", func(t *testing.T) {
		a := packets.SDKVersion{1, 1, 1}
		b := packets.SDKVersion{0, 0, 0}

		assert.Equal(t, packets.SDKVersionNewer, a.Compare(b))

		a = packets.SDKVersion{1, 2, 3}
		b = packets.SDKVersion{1, 1, 3}

		assert.Equal(t, packets.SDKVersionNewer, a.Compare(b))

		a = packets.SDKVersion{1, 2, 3}
		b = packets.SDKVersion{1, 2, 2}

		assert.Equal(t, packets.SDKVersionNewer, a.Compare(b))
	})
}

func TestVersionAtLeast(t *testing.T) {

	t.Run("equal", func(t *testing.T) {
		a := packets.SDKVersion{0, 0, 0}
		b := packets.SDKVersion{0, 0, 0}

		assert.True(t, a.AtLeast(b))
	})

	t.Run("newer", func(t *testing.T) {
		a := packets.SDKVersion{0, 0, 1}
		b := packets.SDKVersion{0, 0, 0}

		assert.True(t, a.AtLeast(b))
	})

	t.Run("older", func(t *testing.T) {
		a := packets.SDKVersion{0, 0, 0}
		b := packets.SDKVersion{0, 0, 1}

		assert.False(t, a.AtLeast(b))
	})
}

// -------------------------------------------------------------------------

func PacketSerializationTest[P packets.Packet](writePacket P, readPacket P, t *testing.T) {

	const BufferSize = 10 * 1024

	buffer := [BufferSize]byte{}

	writeStream := encoding.CreateWriteStream(buffer[:])

	err := writePacket.Serialize(writeStream)
	assert.Nil(t, err)
	writeStream.Flush()
	packetBytes := writeStream.GetBytesProcessed()

	readStream := encoding.CreateReadStream(buffer[:packetBytes])
	err = readPacket.Serialize(readStream)
	assert.Nil(t, err)

	assert.Equal(t, writePacket, readPacket)
}

func GenerateRandomServerInitRequestPacket() packets.SDK5_ServerInitRequestPacket {

	return packets.SDK5_ServerInitRequestPacket{
		Version:        packets.SDKVersion{5, 0, 0},
		BuyerId:        rand.Uint64(),
		RequestId:      rand.Uint64(),
		DatacenterId:   rand.Uint64(),
		DatacenterName: common.RandomString(packets.SDK5_MaxDatacenterNameLength),
	}
}

func GenerateRandomServerInitResponsePacket() packets.SDK5_ServerInitResponsePacket {

	packet := packets.SDK5_ServerInitResponsePacket{
		RequestId:     rand.Uint64(),
		Response:      uint32(common.RandomInt(0,255)),
	}

	common.RandomBytes(packet.UpcomingMagic[:])
	common.RandomBytes(packet.CurrentMagic[:])
	common.RandomBytes(packet.PreviousMagic[:])

	return packet
}

func GenerateRandomServerUpdateRequestPacket() packets.SDK5_ServerUpdateRequestPacket {

	return packets.SDK5_ServerUpdateRequestPacket{
		Version:        packets.SDKVersion{5, 0, 0},
		BuyerId:        rand.Uint64(),
		RequestId:      rand.Uint64(),
		DatacenterId:   rand.Uint64(),
	}
}

func GenerateRandomServerUpdateResponsePacket() packets.SDK5_ServerUpdateResponsePacket {

	packet := packets.SDK5_ServerUpdateResponsePacket{
		RequestId:     rand.Uint64(),
	}

	common.RandomBytes(packet.UpcomingMagic[:])
	common.RandomBytes(packet.CurrentMagic[:])
	common.RandomBytes(packet.PreviousMagic[:])

	return packet
}

func GenerateRandomMatchDataRequestPacket() packets.SDK5_MatchDataRequestPacket {

	packet := packets.SDK5_MatchDataRequestPacket{
		Version:        packets.SDKVersion{1, 2, 3},
		BuyerId:        12341241,
		ServerAddress:  *core.ParseAddress("127.0.0.1:44444"),
		DatacenterId:   184283418,
		UserHash:       210987451,
		SessionId:      987249128471,
		RetryNumber:    4,
		MatchId:        1234209487198,
		NumMatchValues: 10,
	}

	for i := 0; i < int(packet.NumMatchValues); i++ {
		packet.MatchValues[i] = float64(i) * 34852.0
	}

	return packet
}

func GenerateRandomMatchDataResponsePacket() packets.SDK5_MatchDataResponsePacket {

	return packets.SDK5_MatchDataResponsePacket{
		SessionId:     rand.Uint64(),
	}
}

func GenerateRandomSessionUpdateRequestPacket() packets.SDK5_SessionUpdateRequestPacket {

	packet := packets.SDK5_SessionUpdateRequestPacket{
		Version:                         packets.SDKVersion{1, 2, 3},
		BuyerId:                         rand.Uint64(),
		DatacenterId:                    rand.Uint64(),
		SessionId:                       rand.Uint64(),
		SliceNumber:                     rand.Uint32(),
		RetryNumber:                     int32(common.RandomInt(0, packets.SDK5_MaxSessionUpdateRetries)),
		SessionDataBytes:                int32(common.RandomInt(0, packets.SDK5_MaxSessionDataSize)),
		ClientAddress:                   *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", common.RandomInt(0, 65535))),
		ServerAddress:                   *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", common.RandomInt(0, 65535))),
		UserHash:                        rand.Uint64(),
		HasNearRelayPings:               common.RandomBool(),
		Next:                            common.RandomBool(),
		Committed:                       common.RandomBool(),
		Reported:                        common.RandomBool(),
		FallbackToDirect:                common.RandomBool(),
		ClientBandwidthOverLimit:        common.RandomBool(),
		ServerBandwidthOverLimit:        common.RandomBool(),
		ClientPingTimedOut:              common.RandomBool(),
		PlatformType:                    int32(common.RandomInt(0, packets.SDK5_PlatformTypeMax)),
		ConnectionType:                  int32(common.RandomInt(0, packets.SDK5_ConnectionTypeMax)),
		ServerEvents:                    rand.Uint64(),
		NumNearRelays:                   int32(common.RandomInt(0, packets.SDK5_MaxNearRelays)),
		DirectMinRTT:                    rand.Float32(),
		DirectMaxRTT:                    rand.Float32(),
		DirectPrimeRTT:                  rand.Float32(),
		DirectJitter:                    rand.Float32(),
		DirectPacketLoss:                rand.Float32(),
		PacketsSentClientToServer:       rand.Uint64(),
		PacketsSentServerToClient:       rand.Uint64(),
		PacketsLostClientToServer:       rand.Uint64(),
		PacketsLostServerToClient:       rand.Uint64(),
		PacketsOutOfOrderClientToServer: rand.Uint64(),
		PacketsOutOfOrderServerToClient: rand.Uint64(),
		JitterClientToServer:            rand.Float32(),
		JitterServerToClient:            rand.Float32(),
	}

	if packet.SliceNumber == 0 {
		packet.NumTags = int32(common.RandomInt(1, packets.SDK5_MaxTags))
		for i := 0; i < int(packet.NumTags); i++ {
			packet.Tags[i] = rand.Uint64()
		}
	}

	for i := 0; i < int(packet.SessionDataBytes); i++ {
		packet.SessionData[i] = uint8((i + 17) % 256)
	}

	for i := 0; i < int(crypto.KeySize); i++ {
		packet.ClientRoutePublicKey[i] = uint8((i + 7) % 256)
		packet.ServerRoutePublicKey[i] = uint8((i + 13) % 256)
	}

	for i := 0; i < int(packet.NumNearRelays); i++ {
		packet.NearRelayIds[i] = rand.Uint64()
		if packet.HasNearRelayPings {
			packet.NearRelayRTT[i] = int32(common.RandomInt(1, packets.SDK5_MaxNearRelayRTT))
			packet.NearRelayJitter[i] = int32(common.RandomInt(1, packets.SDK5_MaxNearRelayJitter))
			packet.NearRelayPacketLoss[i] = int32(common.RandomInt(1, packets.SDK5_MaxNearRelayPacketLoss))
		}
	}

	if packet.Next {
		packet.NextRTT = rand.Float32()
		packet.NextJitter = rand.Float32()
		packet.NextPacketLoss = rand.Float32()
		packet.NextKbpsUp = rand.Uint32()
		packet.NextKbpsDown = rand.Uint32()
	}

	return packet
}

// ------------------------------------------------------------

const NumIterations = 10000 // todo

func Test_SDK5_ServerInitRequestPacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomServerInitRequestPacket()

		readPacket := packets.SDK5_ServerInitRequestPacket{}

		PacketSerializationTest[*packets.SDK5_ServerInitRequestPacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK5_ServerInitResponsePacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomServerInitResponsePacket()

		readPacket := packets.SDK5_ServerInitResponsePacket{}

		PacketSerializationTest[*packets.SDK5_ServerInitResponsePacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK5_ServerUpdateRequestPacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomServerUpdateRequestPacket()

		readPacket := packets.SDK5_ServerUpdateRequestPacket{}

		PacketSerializationTest[*packets.SDK5_ServerUpdateRequestPacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK5_ServerUpdateResponsePacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomServerUpdateResponsePacket()

		readPacket := packets.SDK5_ServerUpdateResponsePacket{}

		PacketSerializationTest[*packets.SDK5_ServerUpdateResponsePacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK5_MatchDataRequestPacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomMatchDataRequestPacket()

		readPacket := packets.SDK5_MatchDataRequestPacket{}

		PacketSerializationTest[*packets.SDK5_MatchDataRequestPacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK5_MatchDataResponsePacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomMatchDataResponsePacket()

		readPacket := packets.SDK5_MatchDataResponsePacket{}

		PacketSerializationTest[*packets.SDK5_MatchDataResponsePacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK5_SessionUpdateRequestPacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomSessionUpdateRequestPacket()

		readPacket := packets.SDK5_SessionUpdateRequestPacket{}

		PacketSerializationTest[*packets.SDK5_SessionUpdateRequestPacket](&writePacket, &readPacket, t)
	}
}

// -----------------------------

// dragons below...

/*
func Test_SDK5_SessionUpdateResponsePacket_Direct(t *testing.T) {

	writePacket := packets.SDK5_SessionUpdateResponsePacket{
		SessionId:          123412341243,
		SliceNumber:        10234,
		SessionDataBytes:   100,
		RouteType:          packets.SDK5_RouteTypeDirect,
		NearRelaysChanged:  true,
		NumNearRelays:      10,
		HasDebug:           true,
		Debug:              "I am a debug string",
		ExcludeNearRelays:  true,
		HighFrequencyPings: true,
	}

	for i := 0; i < int(writePacket.SessionDataBytes); i++ {
		writePacket.SessionData[i] = uint8((i + 17) % 256)
	}

	for i := 0; i < int(writePacket.NumNearRelays); i++ {
		writePacket.NearRelayIds[i] = uint64(i * 32)
		writePacket.NearRelayAddresses[i] = *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", i+5000))
		writePacket.NearRelayExcluded[i] = (i % 2) == 0
	}

	readPacket := packets.SDK5_SessionUpdateResponsePacket{}

	PacketSerializationTest[*packets.SDK5_SessionUpdateResponsePacket](&writePacket, &readPacket, t)
}

func Test_SDK5_SessionUpdateResponsePacket_NewRoute(t *testing.T) {

	writePacket := packets.SDK5_SessionUpdateResponsePacket{
		SessionId:          123412341243,
		SliceNumber:        10234,
		SessionDataBytes:   100,
		RouteType:          packets.SDK5_RouteTypeNew,
		Multipath:          true,
		Committed:          true,
		NumTokens:          5,
		NearRelaysChanged:  true,
		NumNearRelays:      10,
		HasDebug:           true,
		Debug:              "I am a debug string",
		ExcludeNearRelays:  true,
		HighFrequencyPings: true,
	}

	tokenBytes := writePacket.NumTokens * packets.SDK5_EncryptedNextRouteTokenSize
	writePacket.Tokens = make([]byte, tokenBytes)
	for i := 0; i < int(tokenBytes); i++ {
		writePacket.Tokens[i] = uint8(i + 3)
	}

	for i := 0; i < int(writePacket.SessionDataBytes); i++ {
		writePacket.SessionData[i] = uint8((i + 17) % 256)
	}

	for i := 0; i < int(writePacket.NumNearRelays); i++ {
		writePacket.NearRelayIds[i] = uint64(i * 32)
		writePacket.NearRelayAddresses[i] = *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", i+5000))
		writePacket.NearRelayExcluded[i] = (i % 2) == 0
	}

	readPacket := packets.SDK5_SessionUpdateResponsePacket{}

	PacketSerializationTest[*packets.SDK5_SessionUpdateResponsePacket](&writePacket, &readPacket, t)
}

func Test_SDK5_SessionResponsePacket_ContinueRoute(t *testing.T) {

	writePacket := packets.SDK5_SessionUpdateResponsePacket{
		SessionId:          123412341243,
		SliceNumber:        10234,
		SessionDataBytes:   100,
		RouteType:          packets.SDK5_RouteTypeContinue,
		Multipath:          true,
		Committed:          true,
		NumTokens:          5,
		NearRelaysChanged:  true,
		NumNearRelays:      10,
		HasDebug:           true,
		Debug:              "I am a debug string",
		ExcludeNearRelays:  true,
		HighFrequencyPings: true,
	}

	tokenBytes := writePacket.NumTokens * packets.SDK5_EncryptedContinueRouteTokenSize
	writePacket.Tokens = make([]byte, tokenBytes)
	for i := 0; i < int(tokenBytes); i++ {
		writePacket.Tokens[i] = uint8(i + 3)
	}

	for i := 0; i < int(writePacket.SessionDataBytes); i++ {
		writePacket.SessionData[i] = uint8((i + 17) % 256)
	}

	for i := 0; i < int(writePacket.NumNearRelays); i++ {
		writePacket.NearRelayIds[i] = uint64(i * 32)
		writePacket.NearRelayAddresses[i] = *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", i+5000))
		writePacket.NearRelayExcluded[i] = (i % 2) == 0
	}

	readPacket := packets.SDK5_SessionUpdateResponsePacket{}

	PacketSerializationTest[*packets.SDK5_SessionUpdateResponsePacket](&writePacket, &readPacket, t)
}

func Test_SDK5_SessionData(t *testing.T) {

	routeState := core.RouteState{
		UserID:              123213131,
		Next:                true,
		Veto:                false,
		Banned:              false,
		Disabled:            false,
		NotSelected:         false,
		ABTest:              true,
		A:                   true,
		B:                   false,
		ForcedNext:          false,
		ReduceLatency:       true,
		ReducePacketLoss:    true,
		ProMode:             false,
		Multipath:           true,
		Committed:           true,
		CommitVeto:          false,
		CommitCounter:       0,
		LatencyWorse:        false,
		LocationVeto:        false,
		MultipathOverload:   false,
		NoRoute:             false,
		NextLatencyTooHigh:  false,
		NumNearRelays:       32,
		RelayWentAway:       false,
		RouteLost:           false,
		DirectJitter:        5,
		Mispredict:          false,
		LackOfDiversity:     false,
		MispredictCounter:   0,
		LatencyWorseCounter: 0,
		MultipathRestricted: false,
		PLSustainedCounter:  0,
	}

	for i := 0; i < packets.SDK5_MaxNearRelays; i++ {
		routeState.NearRelayRTT[i] = int32(i + 10)
		routeState.NearRelayJitter[i] = int32(i + 5)
		routeState.NearRelayPLHistory[i] = (uint32(1123414100) >> i) & 0xFF
		routeState.NearRelayPLCount[i] = uint32(500) + uint32(i)
	}

	routeState.DirectPLHistory = 127
	routeState.DirectPLCount = 5
	routeState.PLHistoryIndex = 3
	routeState.PLHistorySamples = 5

	writePacket := packets.SDK5_SessionData{
		Version:                       packets.SDK5_SessionDataVersion,
		SessionId:                     123123131,
		SessionVersion:                5,
		SliceNumber:                   10001,
		ExpireTimestamp:               3249823948198,
		Initial:                       false,
		Location:                      packets.SDK5_LocationData{Latitude: 100.2, Longitude: 95.0, ISP: "Comcast", ASN: 12313},
		RouteChanged:                  true,
		RouteNumRelays:                5,
		RouteCost:                     105,
		RouteState:                    routeState,
		EverOnNext:                    true,
		FellBackToDirect:              false,
		PrevPacketsSentClientToServer: 100000,
		PrevPacketsSentServerToClient: 100234,
		PrevPacketsLostClientToServer: 100021,
		PrevPacketsLostServerToClient: 100005,
		HoldNearRelays:                true,
		WroteSummary:                  false,
		TotalPriceSum:                 123213111,
		NextEnvelopeBytesUpSum:        12313123,
		NextEnvelopeBytesDownSum:      238129381,
		DurationOnNext:                5000,
	}

	for i := 0; i < int(writePacket.RouteNumRelays); i++ {
		writePacket.RouteRelayIds[i] = uint64(i + 1000)
	}

	for i := 0; i < packets.SDK5_MaxNearRelays; i++ {
		writePacket.HoldNearRelayRTT[i] = int32(i + 100)
	}

	readPacket := packets.SDK5_SessionData{}

	PacketSerializationTest[*packets.SDK5_SessionData](&writePacket, &readPacket, t)
}
*/

// ------------------------------------------------------------------------
