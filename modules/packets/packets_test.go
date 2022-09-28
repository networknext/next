package packets

import (
	"fmt"
	"testing"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"

	"github.com/networknext/backend/modules-old/crypto"

	"github.com/stretchr/testify/assert"
)

// ------------------------------------------------------------------------

func TestVersionCompare(t *testing.T) {

	t.Parallel()

	t.Run("equal", func(t *testing.T) {
		a := SDKVersion{1, 2, 3}
		b := SDKVersion{1, 2, 3}

		assert.Equal(t, SDKVersionEqual, a.Compare(b))
	})

	t.Run("older", func(t *testing.T) {
		a := SDKVersion{1, 1, 1}
		b := SDKVersion{2, 0, 0}

		assert.Equal(t, SDKVersionOlder, a.Compare(b))

		a = SDKVersion{1, 1, 1}
		b = SDKVersion{1, 2, 0}

		assert.Equal(t, SDKVersionOlder, a.Compare(b))

		a = SDKVersion{1, 1, 1}
		b = SDKVersion{1, 1, 2}

		assert.Equal(t, SDKVersionOlder, a.Compare(b))
	})

	t.Run("newer", func(t *testing.T) {
		a := SDKVersion{1, 1, 1}
		b := SDKVersion{0, 0, 0}

		assert.Equal(t, SDKVersionNewer, a.Compare(b))

		a = SDKVersion{1, 2, 3}
		b = SDKVersion{1, 1, 3}

		assert.Equal(t, SDKVersionNewer, a.Compare(b))

		a = SDKVersion{1, 2, 3}
		b = SDKVersion{1, 2, 2}

		assert.Equal(t, SDKVersionNewer, a.Compare(b))
	})
}

func TestVersionAtLeast(t *testing.T) {

	t.Parallel()

	t.Run("equal", func(t *testing.T) {
		a := SDKVersion{0, 0, 0}
		b := SDKVersion{0, 0, 0}

		assert.True(t, a.AtLeast(b))
	})

	t.Run("newer", func(t *testing.T) {
		a := SDKVersion{0, 0, 1}
		b := SDKVersion{0, 0, 0}

		assert.True(t, a.AtLeast(b))
	})

	t.Run("older", func(t *testing.T) {
		a := SDKVersion{0, 0, 0}
		b := SDKVersion{0, 0, 1}

		assert.False(t, a.AtLeast(b))
	})
}

func PacketSerializationTest[P Packet](writePacket Packet, readPacket Packet, t *testing.T) {

	t.Parallel()

	const BufferSize = 1024

	buffer := [BufferSize]byte{}

	writeStream := common.CreateWriteStream(buffer[:])

	err := writePacket.Serialize(writeStream)
	assert.Nil(t, err)
	writeStream.Flush()
	packetBytes := writeStream.GetBytesProcessed()

	readStream := common.CreateReadStream(buffer[:packetBytes])
	err = readPacket.Serialize(readStream)
	assert.Nil(t, err)

	assert.Equal(t, writePacket, readPacket)
}

// ------------------------------------------------------------------------

func Test_SDK4_ServerInitRequestPacket(t *testing.T) {

	writePacket := SDK4_ServerInitRequestPacket{
		Version:        SDKVersion{1, 2, 3},
		BuyerId:        1234567,
		DatacenterId:   5124111,
		RequestId:      234198347,
		DatacenterName: "test",
	}

	readPacket := SDK4_ServerInitRequestPacket{}

	PacketSerializationTest[*SDK4_ServerInitRequestPacket](&writePacket, &readPacket, t)
}

func Test_SDK4_ServerInitResponsePacket(t *testing.T) {

	writePacket := SDK4_ServerInitResponsePacket{
		RequestId: 234198347,
		Response:  1,
	}

	readPacket := SDK4_ServerInitResponsePacket{}

	PacketSerializationTest[*SDK4_ServerInitResponsePacket](&writePacket, &readPacket, t)
}

func Test_SDK4_SessionUpdatePacket(t *testing.T) {

	writePacket := SDK4_SessionUpdatePacket{
		Version:                         SDKVersion{1, 2, 3},
		BuyerId:                         123414,
		DatacenterId:                    1234123491,
		SessionId:                       120394810984109,
		SliceNumber:                     5,
		RetryNumber:                     1,
		SessionDataBytes:                100,
		ClientAddress:                   *core.ParseAddress("127.0.0.1:50000"),
		ServerAddress:                   *core.ParseAddress("127.0.0.1:40000"),
		UserHash:                        12341298742,
		PlatformType:                    SDK4_PlatformTypePS4,
		ConnectionType:                  SDK4_ConnectionTypeWired,
		Next:                            true,
		Committed:                       true,
		Reported:                        false,
		FallbackToDirect:                false,
		ClientBandwidthOverLimit:        false,
		ServerBandwidthOverLimit:        false,
		ClientPingTimedOut:              false,
		NumTags:                         2,
		Flags:                           122,
		UserFlags:                       3152384721,
		DirectMinRTT:                    10.0,
		DirectMaxRTT:                    20.0,
		DirectPrimeRTT:                  19.0,
		DirectJitter:                    5.2,
		DirectPacketLoss:                0.1,
		NextRTT:                         5.0,
		NextJitter:                      0.5,
		NextPacketLoss:                  0.01,
		NumNearRelays:                   10,
		NextKbpsUp:                      100,
		NextKbpsDown:                    256,
		PacketsSentClientToServer:       10000,
		PacketsSentServerToClient:       10500,
		PacketsLostClientToServer:       5,
		PacketsLostServerToClient:       10,
		PacketsOutOfOrderClientToServer: 8,
		PacketsOutOfOrderServerToClient: 9,
		JitterClientToServer:            8.2,
		JitterServerToClient:            9.6,
	}

	for i := 0; i < int(writePacket.SessionDataBytes); i++ {
		writePacket.SessionData[i] = uint8((i + 17) % 256)
	}

	for i := 0; i < int(crypto.KeySize); i++ {
		writePacket.ClientRoutePublicKey[i] = uint8((i + 7) % 256)
		writePacket.ServerRoutePublicKey[i] = uint8((i + 13) % 256)
	}

	writePacket.Tags[0] = 12342151
	writePacket.Tags[1] = 134614111111

	for i := 0; i < int(writePacket.NumNearRelays); i++ {
		writePacket.NearRelayIds[i] = uint64(i * 32)
		writePacket.NearRelayRTT[i] = int32(i)
		writePacket.NearRelayJitter[i] = int32(i + 1)
		writePacket.NearRelayPacketLoss[i] = int32(i + 2)
	}

	readPacket := SDK4_SessionUpdatePacket{}

	PacketSerializationTest[*SDK4_SessionUpdatePacket](&writePacket, &readPacket, t)
}

func Test_SDK4_SessionResponsePacket_Direct(t *testing.T) {

	writePacket := SDK4_SessionResponsePacket{
		Version:            SDKVersion{1, 2, 3},
		SessionId:          123412341243,
		SliceNumber:        10234,
		SessionDataBytes:   100,
		RouteType:          SDK4_RouteTypeDirect,
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

	readPacket := SDK4_SessionResponsePacket{}

	PacketSerializationTest[*SDK4_SessionResponsePacket](&writePacket, &readPacket, t)
}

func Test_SDK4_SessionResponsePacket_NewRoute(t *testing.T) {

	writePacket := SDK4_SessionResponsePacket{
		Version:            SDKVersion{1, 2, 3},
		SessionId:          123412341243,
		SliceNumber:        10234,
		SessionDataBytes:   100,
		RouteType:          SDK4_RouteTypeNew,
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

	tokenBytes := writePacket.NumTokens * SDK4_EncryptedNextRouteTokenSize
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

	readPacket := SDK4_SessionResponsePacket{}

	PacketSerializationTest[*SDK4_SessionResponsePacket](&writePacket, &readPacket, t)
}

func Test_SDK4_SessionResponsePacket_ContinueRoute(t *testing.T) {

	writePacket := SDK4_SessionResponsePacket{
		Version:            SDKVersion{1, 2, 3},
		SessionId:          123412341243,
		SliceNumber:        10234,
		SessionDataBytes:   100,
		RouteType:          SDK4_RouteTypeContinue,
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

	tokenBytes := writePacket.NumTokens * SDK4_EncryptedContinueRouteTokenSize
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

	readPacket := SDK4_SessionResponsePacket{}

	PacketSerializationTest[*SDK4_SessionResponsePacket](&writePacket, &readPacket, t)
}

func Test_SDK4_SessionData(t *testing.T) {

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

	for i := 0; i < SDK4_MaxNearRelays; i++ {
		routeState.NearRelayRTT[i] = int32(i + 10)
		routeState.NearRelayJitter[i] = int32(i + 5)
		routeState.NearRelayPLHistory[i] = (uint32(1123414100) >> i) & 0xFF
		routeState.NearRelayPLCount[i] = uint32(500) + uint32(i)
	}

	routeState.DirectPLHistory = 127
	routeState.DirectPLCount = 5
	routeState.PLHistoryIndex = 3
	routeState.PLHistorySamples = 5

	writePacket := SDK4_SessionData{
		Version:                       SDK4_SessionDataVersion,
		SessionId:                     123123131,
		SessionVersion:                5,
		SliceNumber:                   10001,
		ExpireTimestamp:               3249823948198,
		Initial:                       false,
		Location:                      SDK4_LocationData{Latitude: 100.2, Longitude: 95.0, ISP: "Comcast", ASN: 12313},
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

	for i := 0; i < SDK4_MaxNearRelays; i++ {
		writePacket.HoldNearRelayRTT[i] = int32(i + 100)
	}

	readPacket := SDK4_SessionData{}

	PacketSerializationTest[*SDK4_SessionData](&writePacket, &readPacket, t)
}

func Test_SDK4_MatchDataRequestPacket(t *testing.T) {

	writePacket := SDK4_MatchDataRequestPacket{
		Version:        SDKVersion{1, 2, 3},
		BuyerId:        12341241,
		ServerAddress:  *core.ParseAddress("127.0.0.1:44444"),
		DatacenterId:   184283418,
		UserHash:       210987451,
		SessionId:      987249128471,
		RetryNumber:    4,
		MatchId:        1234209487198,
		NumMatchValues: 10,
	}

	for i := 0; i < int(writePacket.NumMatchValues); i++ {
		writePacket.MatchValues[i] = float64(i) * 34852.0
	}

	readPacket := SDK4_MatchDataRequestPacket{}

	PacketSerializationTest[*SDK4_MatchDataRequestPacket](&writePacket, &readPacket, t)
}

func Test_SDK4_MatchDataResponsePacket(t *testing.T) {

	writePacket := SDK4_MatchDataResponsePacket{
		SessionId: 1234141,
		Response:  1,
	}

	readPacket := SDK4_MatchDataResponsePacket{}

	PacketSerializationTest[*SDK4_MatchDataResponsePacket](&writePacket, &readPacket, t)
}

// ------------------------------------------------------------------------

func Test_SDK5_ServerInitRequestPacket(t *testing.T) {

	writePacket := SDK5_ServerInitRequestPacket{
		Version:        SDKVersion{1, 2, 3},
		BuyerId:        1234567,
		RequestId:      234198347,
		DatacenterId:   5124111,
		DatacenterName: "test",
	}

	readPacket := SDK5_ServerInitRequestPacket{}

	PacketSerializationTest[*SDK5_ServerInitRequestPacket](&writePacket, &readPacket, t)
}

func Test_SDK5_ServerInitResponsePacket(t *testing.T) {

	writePacket := SDK5_ServerInitResponsePacket{
		RequestId:     234198347,
		Response:      1,
		UpcomingMagic: [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		CurrentMagic:  [8]byte{2, 3, 4, 5, 6, 7, 8, 9},
		PreviousMagic: [8]byte{3, 4, 5, 6, 7, 8, 9, 10},
	}

	readPacket := SDK5_ServerInitResponsePacket{}

	PacketSerializationTest[*SDK5_ServerInitResponsePacket](&writePacket, &readPacket, t)
}

func Test_SDK5_ServerUpdateRequestPacket(t *testing.T) {

	writePacket := SDK5_ServerUpdateRequestPacket{
		Version:      SDKVersion{1, 2, 3},
		BuyerId:      1234567,
		RequestId:    234198347,
		DatacenterId: 5124111,
	}

	readPacket := SDK5_ServerUpdateRequestPacket{}

	PacketSerializationTest[*SDK5_ServerUpdateRequestPacket](&writePacket, &readPacket, t)
}

func Test_SDK5_ServerUpdateResponsePacket(t *testing.T) {

	writePacket := SDK5_ServerUpdateResponsePacket{
		RequestId:     234198347,
		UpcomingMagic: [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		CurrentMagic:  [8]byte{2, 3, 4, 5, 6, 7, 8, 9},
		PreviousMagic: [8]byte{3, 4, 5, 6, 7, 8, 9, 10},
	}

	readPacket := SDK5_ServerUpdateResponsePacket{}

	PacketSerializationTest[*SDK5_ServerUpdateResponsePacket](&writePacket, &readPacket, t)
}

func Test_SDK5_SessionUpdateRequestPacket(t *testing.T) {

	writePacket := SDK5_SessionUpdateRequestPacket{
		Version:                         SDKVersion{1, 2, 3},
		BuyerId:                         123414,
		DatacenterId:                    1234123491,
		SessionId:                       120394810984109,
		SliceNumber:                     5,
		RetryNumber:                     1,
		SessionDataBytes:                100,
		ClientAddress:                   *core.ParseAddress("127.0.0.1:50000"),
		ServerAddress:                   *core.ParseAddress("127.0.0.1:40000"),
		UserHash:                        12341298742,
		PlatformType:                    SDK5_PlatformTypePS4,
		ConnectionType:                  SDK5_ConnectionTypeWired,
		Next:                            true,
		Committed:                       true,
		Reported:                        false,
		FallbackToDirect:                false,
		ClientBandwidthOverLimit:        false,
		ServerBandwidthOverLimit:        false,
		ClientPingTimedOut:              false,
		NumTags:                         2,
		Flags:                           122,
		UserFlags:                       3152384721,
		DirectMinRTT:                    10.0,
		DirectMaxRTT:                    20.0,
		DirectPrimeRTT:                  19.0,
		DirectJitter:                    5.2,
		DirectPacketLoss:                0.1,
		NextRTT:                         5.0,
		NextJitter:                      0.5,
		NextPacketLoss:                  0.01,
		NumNearRelays:                   10,
		NextKbpsUp:                      100,
		NextKbpsDown:                    256,
		PacketsSentClientToServer:       10000,
		PacketsSentServerToClient:       10500,
		PacketsLostClientToServer:       5,
		PacketsLostServerToClient:       10,
		PacketsOutOfOrderClientToServer: 8,
		PacketsOutOfOrderServerToClient: 9,
		JitterClientToServer:            8.2,
		JitterServerToClient:            9.6,
	}

	for i := 0; i < int(writePacket.SessionDataBytes); i++ {
		writePacket.SessionData[i] = uint8((i + 17) % 256)
	}

	for i := 0; i < int(crypto.KeySize); i++ {
		writePacket.ClientRoutePublicKey[i] = uint8((i + 7) % 256)
		writePacket.ServerRoutePublicKey[i] = uint8((i + 13) % 256)
	}

	writePacket.Tags[0] = 12342151
	writePacket.Tags[1] = 134614111111

	for i := 0; i < int(writePacket.NumNearRelays); i++ {
		writePacket.NearRelayIds[i] = uint64(i * 32)
		writePacket.NearRelayRTT[i] = int32(i)
		writePacket.NearRelayJitter[i] = int32(i + 1)
		writePacket.NearRelayPacketLoss[i] = int32(i + 2)
	}

	readPacket := SDK5_SessionUpdateRequestPacket{}

	PacketSerializationTest[*SDK5_SessionUpdateRequestPacket](&writePacket, &readPacket, t)
}

func Test_SDK5_SessionUpdateResponsePacket_Direct(t *testing.T) {

	writePacket := SDK5_SessionUpdateResponsePacket{
		Version:            SDKVersion{1, 2, 3},
		SessionId:          123412341243,
		SliceNumber:        10234,
		SessionDataBytes:   100,
		RouteType:          SDK5_RouteTypeDirect,
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

	readPacket := SDK5_SessionUpdateResponsePacket{}

	PacketSerializationTest[*SDK5_SessionUpdateResponsePacket](&writePacket, &readPacket, t)
}

func Test_SDK5_SessionUpdateResponsePacket_NewRoute(t *testing.T) {

	writePacket := SDK5_SessionUpdateResponsePacket{
		Version:            SDKVersion{1, 2, 3},
		SessionId:          123412341243,
		SliceNumber:        10234,
		SessionDataBytes:   100,
		RouteType:          SDK5_RouteTypeNew,
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

	tokenBytes := writePacket.NumTokens * SDK5_EncryptedNextRouteTokenSize
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

	readPacket := SDK5_SessionUpdateResponsePacket{}

	PacketSerializationTest[*SDK5_SessionUpdateResponsePacket](&writePacket, &readPacket, t)
}

func Test_SDK5_SessionResponsePacket_ContinueRoute(t *testing.T) {

	writePacket := SDK5_SessionUpdateResponsePacket{
		Version:            SDKVersion{1, 2, 3},
		SessionId:          123412341243,
		SliceNumber:        10234,
		SessionDataBytes:   100,
		RouteType:          SDK5_RouteTypeContinue,
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

	tokenBytes := writePacket.NumTokens * SDK5_EncryptedContinueRouteTokenSize
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

	readPacket := SDK5_SessionUpdateResponsePacket{}

	PacketSerializationTest[*SDK5_SessionUpdateResponsePacket](&writePacket, &readPacket, t)
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

	for i := 0; i < SDK5_MaxNearRelays; i++ {
		routeState.NearRelayRTT[i] = int32(i + 10)
		routeState.NearRelayJitter[i] = int32(i + 5)
		routeState.NearRelayPLHistory[i] = (uint32(1123414100) >> i) & 0xFF
		routeState.NearRelayPLCount[i] = uint32(500) + uint32(i)
	}

	routeState.DirectPLHistory = 127
	routeState.DirectPLCount = 5
	routeState.PLHistoryIndex = 3
	routeState.PLHistorySamples = 5

	writePacket := SDK5_SessionData{
		Version:                       SDK5_SessionDataVersion,
		SessionId:                     123123131,
		SessionVersion:                5,
		SliceNumber:                   10001,
		ExpireTimestamp:               3249823948198,
		Initial:                       false,
		Location:                      SDK5_LocationData{Latitude: 100.2, Longitude: 95.0, ISP: "Comcast", ASN: 12313},
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

	for i := 0; i < SDK5_MaxNearRelays; i++ {
		writePacket.HoldNearRelayRTT[i] = int32(i + 100)
	}

	readPacket := SDK5_SessionData{}

	PacketSerializationTest[*SDK5_SessionData](&writePacket, &readPacket, t)
}

func Test_SDK5_MatchDataRequestPacket(t *testing.T) {

	writePacket := SDK5_MatchDataRequestPacket{
		Version:        SDKVersion{1, 2, 3},
		BuyerId:        12341241,
		ServerAddress:  *core.ParseAddress("127.0.0.1:44444"),
		DatacenterId:   184283418,
		UserHash:       210987451,
		SessionId:      987249128471,
		RetryNumber:    4,
		MatchId:        1234209487198,
		NumMatchValues: 10,
	}

	for i := 0; i < int(writePacket.NumMatchValues); i++ {
		writePacket.MatchValues[i] = float64(i) * 34852.0
	}

	readPacket := SDK5_MatchDataRequestPacket{}

	PacketSerializationTest[*SDK5_MatchDataRequestPacket](&writePacket, &readPacket, t)
}

func Test_SDK5_MatchDataResponsePacket(t *testing.T) {

	writePacket := SDK5_MatchDataResponsePacket{
		SessionId: 1234141,
		Response:  1,
	}

	readPacket := SDK5_MatchDataResponsePacket{}

	PacketSerializationTest[*SDK5_MatchDataResponsePacket](&writePacket, &readPacket, t)
}

// ------------------------------------------------------------------------
