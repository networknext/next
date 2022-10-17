package packets_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/packets"

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
		RequestId: rand.Uint64(),
		Response:  uint32(common.RandomInt(0, 255)),
	}

	common.RandomBytes(packet.UpcomingMagic[:])
	common.RandomBytes(packet.CurrentMagic[:])
	common.RandomBytes(packet.PreviousMagic[:])

	return packet
}

func GenerateRandomServerUpdateRequestPacket() packets.SDK5_ServerUpdateRequestPacket {

	return packets.SDK5_ServerUpdateRequestPacket{
		Version:      packets.SDKVersion{5, 0, 0},
		BuyerId:      rand.Uint64(),
		RequestId:    rand.Uint64(),
		DatacenterId: rand.Uint64(),
	}
}

func GenerateRandomServerUpdateResponsePacket() packets.SDK5_ServerUpdateResponsePacket {

	packet := packets.SDK5_ServerUpdateResponsePacket{
		RequestId: rand.Uint64(),
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
		SessionId: rand.Uint64(),
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

	for i := 0; i < int(crypto.Box_KeySize); i++ {
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

func GenerateRandomSessionUpdateResponsePacket() packets.SDK5_SessionUpdateResponsePacket {

	packet := packets.SDK5_SessionUpdateResponsePacket{
		SessionId:          rand.Uint64(),
		SliceNumber:        rand.Uint32(),
		SessionDataBytes:   int32(common.RandomInt(0, packets.SDK5_MaxSessionDataSize)),
		NearRelaysChanged:  common.RandomBool(),
		HasDebug:           common.RandomBool(),
		ExcludeNearRelays:  common.RandomBool(),
		HighFrequencyPings: common.RandomBool(),
	}

	if packet.HasDebug {
		packet.Debug = common.RandomString(packets.SDK5_MaxSessionDebug)
	}

	for i := 0; i < int(packet.SessionDataBytes); i++ {
		packet.SessionData[i] = uint8((i + 17) % 256)
	}

	if packet.NearRelaysChanged {
		packet.NumNearRelays = int32(common.RandomInt(0, packets.SDK5_MaxNearRelays))
		for i := 0; i < int(packet.NumNearRelays); i++ {
			packet.NearRelayIds[i] = uint64(i * 32)
			packet.NearRelayAddresses[i] = *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", i+5000))
		}
	}

	if packet.ExcludeNearRelays {
		for i := 0; i < int(packets.SDK5_MaxNearRelays); i++ {
			packet.NearRelayExcluded[i] = common.RandomBool()
		}
	}

	packet.RouteType = int32(common.RandomInt(packets.SDK5_RouteTypeDirect, packets.SDK5_RouteTypeContinue))

	if packet.RouteType != packets.SDK5_RouteTypeDirect {
		packet.Multipath = common.RandomBool()
		packet.Committed = common.RandomBool()
		packet.NumTokens = int32(common.RandomInt(1, packets.SDK5_MaxTokens))
	}

	if packet.RouteType == packets.SDK5_RouteTypeNew {
		packet.Tokens = make([]byte, packet.NumTokens*packets.SDK5_EncryptedNextRouteTokenSize)
		for i := range packet.Tokens {
			packet.Tokens[i] = byte(common.RandomInt(0, 255))
		}
	}

	if packet.RouteType == packets.SDK5_RouteTypeContinue {
		packet.Tokens = make([]byte, packet.NumTokens*packets.SDK5_EncryptedContinueRouteTokenSize)
		for i := range packet.Tokens {
			packet.Tokens[i] = byte(common.RandomInt(0, 255))
		}
	}

	return packet
}

// ------------------------------------------------------------

const NumIterations = 1000

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

func Test_SDK5_SessionUpdateResponsePacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomSessionUpdateResponsePacket()

		readPacket := packets.SDK5_SessionUpdateResponsePacket{}

		PacketSerializationTest[*packets.SDK5_SessionUpdateResponsePacket](&writePacket, &readPacket, t)
	}
}

// -------------------------------------------------------------------

const NumRelayPacketIterations = 1000

func RelayPacketReadWriteTest[P packets.RelayPacket](writePacket P, readPacket P, t *testing.T) {

	const BufferSize = 150 * 1024

	var buffer [BufferSize]byte

	output := writePacket.Write(buffer[:])

	err := readPacket.Read(output)
	assert.Nil(t, err)

	assert.Equal(t, writePacket, readPacket)
}

func GenerateRandomRelayUpdateRequestPacket() packets.RelayUpdateRequestPacket {

	packet := packets.RelayUpdateRequestPacket{
		Version:    packets.VersionNumberRelayUpdateRequest,
		Address:    common.RandomAddress(),
		Token:      make([]byte, packets.RelayTokenSize),
		NumSamples: uint32(common.RandomInt(0, packets.MaxRelays-1)),
	}

	for i := 0; i < int(packet.NumSamples); i++ {
		packet.SampleRelayId[i] = rand.Uint64()
		packet.SampleRTT[i] = rand.Float32()
		packet.SampleJitter[i] = rand.Float32()
		packet.SamplePacketLoss[i] = rand.Float32()
	}

	packet.SessionCount = rand.Uint64()
	packet.ShuttingDown = common.RandomBool()
	packet.RelayVersion = common.RandomString(packets.MaxRelayVersionStringLength)
	packet.CPU = uint8(common.RandomInt(0, 100))
	packet.EnvelopeUpKbps = rand.Uint64()
	packet.EnvelopeDownKbps = rand.Uint64()
	packet.BandwidthSentKbps = rand.Uint64()
	packet.BandwidthRecvKbps = rand.Uint64()

	return packet
}

func GenerateRandomRelayUpdateResponsePacket() packets.RelayUpdateResponsePacket {

	packet := packets.RelayUpdateResponsePacket{
		Version:       packets.VersionNumberRelayUpdateResponse,
		Timestamp:     rand.Uint64(),
		NumRelays:     uint32(common.RandomInt(0, packets.MaxRelays)),
		UpcomingMagic: make([]byte, 8),
		CurrentMagic:  make([]byte, 8),
		PreviousMagic: make([]byte, 8),
	}

	for i := 0; i < int(packet.NumRelays); i++ {
		packet.RelayId[i] = rand.Uint64()
		packet.RelayAddress[i] = common.RandomString(packets.MaxRelayAddressLength)
	}

	packet.TargetVersion = common.RandomString(packets.MaxRelayVersionStringLength)

	common.RandomBytes(packet.UpcomingMagic)
	common.RandomBytes(packet.CurrentMagic)
	common.RandomBytes(packet.PreviousMagic)

	return packet
}

func TestRelayUpdateRequestPacket(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumRelayPacketIterations; i++ {
		writeMessage := GenerateRandomRelayUpdateRequestPacket()
		readMessage := packets.RelayUpdateRequestPacket{}
		RelayPacketReadWriteTest[*packets.RelayUpdateRequestPacket](&writeMessage, &readMessage, t)
	}
}

func TestRelayUpdateResponsePacket(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumRelayPacketIterations; i++ {
		writeMessage := GenerateRandomRelayUpdateResponsePacket()
		readMessage := packets.RelayUpdateResponsePacket{}
		RelayPacketReadWriteTest[*packets.RelayUpdateResponsePacket](&writeMessage, &readMessage, t)
	}
}

// ------------------------------------------------------------------

func GenerateRandomSessionData() packets.SDK5_SessionData {

	sessionData := packets.SDK5_SessionData{
		Version:                       uint32(common.RandomInt(packets.SDK5_SessionDataVersion_Min, packets.SDK5_SessionDataVersion_Max)),
		SessionId:                     rand.Uint64(),
		SessionVersion:                uint32(common.RandomInt(0, 255)),
		SliceNumber:                   rand.Uint32(),
		ExpireTimestamp:               rand.Uint64(),
		Initial:                       common.RandomBool(),
		RouteChanged:                  common.RandomBool(),
		RouteNumRelays:                int32(common.RandomInt(0, packets.SDK5_MaxRelaysPerRoute)),
		RouteCost:                     int32(common.RandomInt(0, packets.SDK5_InvalidRouteValue)),
		EverOnNext:                    common.RandomBool(),
		FallbackToDirect:              common.RandomBool(),
		PrevPacketsSentClientToServer: rand.Uint64(),
		PrevPacketsSentServerToClient: rand.Uint64(),
		PrevPacketsLostClientToServer: rand.Uint64(),
		PrevPacketsLostServerToClient: rand.Uint64(),
		HoldNearRelays:                common.RandomBool(),
		WroteSummary:                  common.RandomBool(),
		TotalPriceSum:                 rand.Uint64(),
		NextEnvelopeBytesUpSum:        rand.Uint64(),
		NextEnvelopeBytesDownSum:      rand.Uint64(),
		DurationOnNext:                rand.Uint32(),
	}

	for i := 0; i < int(sessionData.RouteNumRelays); i++ {
		sessionData.RouteRelayIds[i] = rand.Uint64()
	}

	if sessionData.HoldNearRelays {
		for i := 0; i < core.MaxNearRelays; i++ {
			sessionData.HoldNearRelayRTT[i] = int32(common.RandomInt(0, 255))
		}
	}

	// todo: Location

	// todo: RouteState (big)

	return sessionData
}

const NumSessionDataIterations = 1

func TestSessionUpdate(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumSessionDataIterations; i++ {
		writeMessage := GenerateRandomSessionData()
		readMessage := packets.SDK5_SessionData{}
		PacketSerializationTest[*packets.SDK5_SessionData](&writeMessage, &readMessage, t)
	}
}
