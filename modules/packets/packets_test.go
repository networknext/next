package packets_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
	"github.com/networknext/next/modules/encoding"
	"github.com/networknext/next/modules/packets"

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

	const BufferSize = constants.MaxPacketBytes

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

func GenerateRandomServerInitRequestPacket() packets.SDK_ServerInitRequestPacket {

	return packets.SDK_ServerInitRequestPacket{
		Version:        packets.SDKVersion{1, 0, 0},
		BuyerId:        rand.Uint64(),
		RequestId:      rand.Uint64(),
		DatacenterId:   rand.Uint64(),
		DatacenterName: common.RandomString(packets.SDK_MaxDatacenterNameLength),
	}
}

func GenerateRandomServerInitResponsePacket() packets.SDK_ServerInitResponsePacket {

	packet := packets.SDK_ServerInitResponsePacket{
		RequestId: rand.Uint64(),
		Response:  uint32(common.RandomInt(0, 255)),
	}

	common.RandomBytes(packet.UpcomingMagic[:])
	common.RandomBytes(packet.CurrentMagic[:])
	common.RandomBytes(packet.PreviousMagic[:])

	return packet
}

func GenerateRandomServerUpdateRequestPacket() packets.SDK_ServerUpdateRequestPacket {

	return packets.SDK_ServerUpdateRequestPacket{
		Version:      packets.SDKVersion{1, 2, 6},
		BuyerId:      rand.Uint64(),
		RequestId:    rand.Uint64(),
		DatacenterId: rand.Uint64(),
		Uptime:       rand.Uint64(),
	}
}

func GenerateRandomServerUpdateResponsePacket() packets.SDK_ServerUpdateResponsePacket {

	packet := packets.SDK_ServerUpdateResponsePacket{
		RequestId: rand.Uint64(),
	}

	common.RandomBytes(packet.UpcomingMagic[:])
	common.RandomBytes(packet.CurrentMagic[:])
	common.RandomBytes(packet.PreviousMagic[:])

	return packet
}

func GenerateRandomClientRelayRequestPacket() packets.SDK_ClientRelayRequestPacket {

	packet := packets.SDK_ClientRelayRequestPacket{
		Version:       packets.SDKVersion{1, 0, 0},
		BuyerId:       rand.Uint64(),
		DatacenterId:  rand.Uint64(),
		ClientAddress: common.RandomAddress(),
	}

	return packet
}

func GenerateRandomClientRelayResponsePacket() packets.SDK_ClientRelayResponsePacket {

	packet := packets.SDK_ClientRelayResponsePacket{
		ClientAddress:   common.RandomAddress(),
		RequestId:       rand.Uint64(),
		Latitude:        rand.Float32(),
		Longitude:       rand.Float32(),
		NumClientRelays: int32(common.RandomInt(0, constants.MaxClientRelays)),
		ExpireTimestamp: rand.Uint64(),
	}

	for i := 0; i < int(packet.NumClientRelays); i++ {
		packet.ClientRelayIds[i] = rand.Uint64()
		packet.ClientRelayAddresses[i] = common.RandomAddress()
		common.RandomBytes(packet.ClientRelayPingTokens[i][:])
	}

	return packet
}

func GenerateRandomServerRelayRequestPacket() packets.SDK_ServerRelayRequestPacket {

	packet := packets.SDK_ServerRelayRequestPacket{
		Version:      packets.SDKVersion{1, 0, 0},
		BuyerId:      rand.Uint64(),
		RequestId:    rand.Uint64(),
		DatacenterId: rand.Uint64(),
	}

	return packet
}

func GenerateRandomServerRelayResponsePacket() packets.SDK_ServerRelayResponsePacket {

	packet := packets.SDK_ServerRelayResponsePacket{
		RequestId:       rand.Uint64(),
		NumServerRelays: int32(common.RandomInt(0, constants.MaxServerRelays)),
		ExpireTimestamp: rand.Uint64(),
	}

	for i := 0; i < int(packet.NumServerRelays); i++ {
		packet.ServerRelayIds[i] = rand.Uint64()
		packet.ServerRelayAddresses[i] = common.RandomAddress()
		common.RandomBytes(packet.ServerRelayPingTokens[i][:])
	}

	return packet
}

func GenerateRandomSessionUpdateRequestPacket() packets.SDK_SessionUpdateRequestPacket {

	packet := packets.SDK_SessionUpdateRequestPacket{
		Version:                         packets.SDKVersion{1, 2, 6},
		BuyerId:                         rand.Uint64(),
		DatacenterId:                    rand.Uint64(),
		SessionId:                       rand.Uint64(),
		SliceNumber:                     rand.Uint32(),
		RetryNumber:                     int32(common.RandomInt(0, packets.SDK_MaxSessionUpdateRetries)),
		SessionDataBytes:                int32(common.RandomInt(0, packets.SDK_MaxSessionDataSize)),
		ClientAddress:                   core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", common.RandomInt(0, 65535))),
		ServerAddress:                   core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", common.RandomInt(0, 65535))),
		UserHash:                        rand.Uint64(),
		HasClientRelayPings:             common.RandomBool(),
		HasServerRelayPings:             common.RandomBool(),
		ClientRelayPingsHaveChanged:     common.RandomBool(),
		ServerRelayPingsHaveChanged:     common.RandomBool(),
		Next:                            common.RandomBool(),
		Reported:                        common.RandomBool(),
		FallbackToDirect:                common.RandomBool(),
		ClientNextBandwidthOverLimit:    common.RandomBool(),
		ServerNextBandwidthOverLimit:    common.RandomBool(),
		ClientPingTimedOut:              common.RandomBool(),
		PlatformType:                    int32(common.RandomInt(0, packets.SDK_PlatformTypeMax)),
		ConnectionType:                  int32(common.RandomInt(0, packets.SDK_ConnectionTypeMax)),
		SessionEvents:                   rand.Uint64(),
		InternalEvents:                  rand.Uint64(),
		DirectRTT:                       rand.Float32(),
		DirectJitter:                    rand.Float32(),
		DirectPacketLoss:                rand.Float32(),
		DirectMaxPacketLossSeen:         rand.Float32(),
		PacketsSentClientToServer:       rand.Uint64(),
		PacketsSentServerToClient:       rand.Uint64(),
		PacketsLostClientToServer:       rand.Uint64(),
		PacketsLostServerToClient:       rand.Uint64(),
		PacketsOutOfOrderClientToServer: rand.Uint64(),
		PacketsOutOfOrderServerToClient: rand.Uint64(),
		JitterClientToServer:            rand.Float32(),
		JitterServerToClient:            rand.Float32(),
		DeltaTimeMin:                    rand.Float32(),
		DeltaTimeMax:                    rand.Float32(),
		DeltaTimeAvg:                    rand.Float32(),
	}

	for i := 0; i < int(packet.SessionDataBytes); i++ {
		packet.SessionData[i] = uint8((i + 17) % 256)
	}

	if packet.SessionDataBytes > 0 {
		common.RandomBytes(packet.SessionData[:packet.SessionDataBytes])
	}

	for i := 0; i < int(crypto.Box_PublicKeySize); i++ {
		packet.ClientRoutePublicKey[i] = uint8((i + 7) % 256)
		packet.ServerRoutePublicKey[i] = uint8((i + 13) % 256)
	}

	if packet.HasClientRelayPings {
		packet.NumClientRelays = int32(common.RandomInt(0, packets.SDK_MaxClientRelays))
		for i := 0; i < int(packet.NumClientRelays); i++ {
			packet.ClientRelayIds[i] = rand.Uint64()
			if packet.HasClientRelayPings {
				packet.ClientRelayRTT[i] = int32(common.RandomInt(1, packets.SDK_MaxRelayRTT))
				packet.ClientRelayJitter[i] = int32(common.RandomInt(1, packets.SDK_MaxRelayJitter))
				packet.ClientRelayPacketLoss[i] = rand.Float32()
			}
		}
	}

	packet.DirectKbpsUp = rand.Uint32()
	packet.DirectKbpsDown = rand.Uint32()

	if packet.Next {
		packet.NextRTT = rand.Float32()
		packet.NextJitter = rand.Float32()
		packet.NextPacketLoss = rand.Float32()
		packet.NextKbpsUp = rand.Uint32()
		packet.NextKbpsDown = rand.Uint32()
	}

	return packet
}

func GenerateRandomSessionUpdateResponsePacket() packets.SDK_SessionUpdateResponsePacket {

	packet := packets.SDK_SessionUpdateResponsePacket{
		SessionId:        rand.Uint64(),
		SliceNumber:      rand.Uint32(),
		SessionDataBytes: int32(common.RandomInt(0, packets.SDK_MaxSessionDataSize)),
	}

	for i := 0; i < int(packet.SessionDataBytes); i++ {
		packet.SessionData[i] = uint8((i + 17) % 256)
	}

	if packet.SessionDataBytes > 0 {
		common.RandomBytes(packet.SessionData[:packet.SessionDataBytes])
		common.RandomBytes(packet.SessionDataSignature[:])
	}

	packet.RouteType = int32(common.RandomInt(packets.SDK_RouteTypeDirect, packets.SDK_RouteTypeContinue))

	if packet.RouteType != packets.SDK_RouteTypeDirect {
		packet.NumTokens = int32(common.RandomInt(1, packets.SDK_MaxTokens))
	}

	if packet.RouteType == packets.SDK_RouteTypeNew {
		packet.Tokens = make([]byte, packet.NumTokens*packets.SDK_EncryptedNextRouteTokenSize)
		for i := range packet.Tokens {
			packet.Tokens[i] = byte(common.RandomInt(0, 255))
		}
	}

	if packet.RouteType == packets.SDK_RouteTypeContinue {
		packet.Tokens = make([]byte, packet.NumTokens*packets.SDK_EncryptedContinueRouteTokenSize)
		for i := range packet.Tokens {
			packet.Tokens[i] = byte(common.RandomInt(0, 255))
		}
	}

	return packet
}

// ------------------------------------------------------------

const NumIterations = 1000

func Test_SDK_ServerInitRequestPacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomServerInitRequestPacket()

		readPacket := packets.SDK_ServerInitRequestPacket{}

		PacketSerializationTest[*packets.SDK_ServerInitRequestPacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK_ServerInitResponsePacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomServerInitResponsePacket()

		readPacket := packets.SDK_ServerInitResponsePacket{}

		PacketSerializationTest[*packets.SDK_ServerInitResponsePacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK_ServerUpdateRequestPacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomServerUpdateRequestPacket()

		readPacket := packets.SDK_ServerUpdateRequestPacket{}

		PacketSerializationTest[*packets.SDK_ServerUpdateRequestPacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK_ServerUpdateResponsePacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomServerUpdateResponsePacket()

		readPacket := packets.SDK_ServerUpdateResponsePacket{}

		PacketSerializationTest[*packets.SDK_ServerUpdateResponsePacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK_ClientRelayRequestPacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomClientRelayRequestPacket()

		readPacket := packets.SDK_ClientRelayRequestPacket{}

		PacketSerializationTest[*packets.SDK_ClientRelayRequestPacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK_ClientRelayResponsePacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomClientRelayResponsePacket()

		readPacket := packets.SDK_ClientRelayResponsePacket{}

		PacketSerializationTest[*packets.SDK_ClientRelayResponsePacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK_ServerRelayRequestPacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomServerRelayRequestPacket()

		readPacket := packets.SDK_ServerRelayRequestPacket{}

		PacketSerializationTest[*packets.SDK_ServerRelayRequestPacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK_ServerRelayResponsePacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomServerRelayResponsePacket()

		readPacket := packets.SDK_ServerRelayResponsePacket{}

		PacketSerializationTest[*packets.SDK_ServerRelayResponsePacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK_SessionUpdateRequestPacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomSessionUpdateRequestPacket()

		readPacket := packets.SDK_SessionUpdateRequestPacket{}

		PacketSerializationTest[*packets.SDK_SessionUpdateRequestPacket](&writePacket, &readPacket, t)
	}
}

func Test_SDK_SessionUpdateResponsePacket(t *testing.T) {

	t.Parallel()

	for i := 0; i < NumIterations; i++ {

		writePacket := GenerateRandomSessionUpdateResponsePacket()

		readPacket := packets.SDK_SessionUpdateResponsePacket{}

		PacketSerializationTest[*packets.SDK_SessionUpdateResponsePacket](&writePacket, &readPacket, t)
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
		Version:     uint8(common.RandomInt(packets.RelayUpdateRequestPacket_VersionMin, packets.RelayUpdateRequestPacket_VersionMax)),
		CurrentTime: rand.Uint64(),
		StartTime:   rand.Uint64(),
		Address:     common.RandomAddress(),
		NumSamples:  uint32(common.RandomInt(0, constants.MaxRelays-1)),
	}

	for i := 0; i < int(packet.NumSamples); i++ {
		packet.SampleRelayId[i] = rand.Uint64()
		packet.SampleRTT[i] = uint8(common.RandomInt(0, 255))
		packet.SampleJitter[i] = uint8(common.RandomInt(0, 255))
		packet.SamplePacketLoss[i] = uint16(common.RandomInt(0, 65535))
	}

	packet.SessionCount = rand.Uint32()
	packet.EnvelopeBandwidthUpKbps = rand.Uint32()
	packet.EnvelopeBandwidthDownKbps = rand.Uint32()
	packet.PacketsSentPerSecond = float32(common.RandomInt(0, 1000))
	packet.PacketsReceivedPerSecond = float32(common.RandomInt(0, 1000))
	packet.BandwidthSentKbps = float32(common.RandomInt(0, 1000))
	packet.BandwidthReceivedKbps = float32(common.RandomInt(0, 1000))
	packet.ClientPingsPerSecond = float32(common.RandomInt(0, 1000))
	packet.ServerPingsPerSecond = float32(common.RandomInt(0, 1000))
	packet.RelayPingsPerSecond = float32(common.RandomInt(0, 1000))

	packet.RelayFlags = rand.Uint64()
	packet.RelayVersion = common.RandomString(constants.MaxRelayVersionLength)

	packet.NumRelayCounters = constants.NumRelayCounters
	for i := 0; i < constants.NumRelayCounters; i++ {
		packet.RelayCounters[i] = rand.Uint64()
	}

	return packet
}

func GenerateRandomRelayUpdateResponsePacket() packets.RelayUpdateResponsePacket {

	packet := packets.RelayUpdateResponsePacket{
		Version:   uint8(common.RandomInt(packets.RelayUpdateResponsePacket_VersionMin, packets.RelayUpdateResponsePacket_VersionMax)),
		Timestamp: rand.Uint64(),
		NumRelays: uint32(common.RandomInt(0, constants.MaxRelays)),
	}

	for i := 0; i < int(packet.NumRelays); i++ {
		packet.RelayId[i] = rand.Uint64()
		packet.RelayAddress[i] = common.RandomAddress()
		if common.RandomBool() {
			packet.RelayInternal[i] = 1
		}
	}

	packet.TargetVersion = common.RandomString(constants.MaxRelayVersionLength)

	common.RandomBytes(packet.UpcomingMagic[:])
	common.RandomBytes(packet.CurrentMagic[:])
	common.RandomBytes(packet.PreviousMagic[:])

	packet.ExpectedPublicAddress = common.RandomAddress()
	if common.RandomBool() {
		packet.ExpectedHasInternalAddress = 1
		packet.ExpectedInternalAddress = common.RandomAddress()
	}
	common.RandomBytes(packet.ExpectedRelayPublicKey[:])
	common.RandomBytes(packet.ExpectedRelayBackendPublicKey[:])
	common.RandomBytes(packet.TestToken[:])
	common.RandomBytes(packet.PingKey[:])

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

const NumSessionDataIterations = 1000

func TestSessionUpdate(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumSessionDataIterations; i++ {
		writeMessage := packets.GenerateRandomSessionData()
		readMessage := packets.SDK_SessionData{}
		PacketSerializationTest[*packets.SDK_SessionData](&writeMessage, &readMessage, t)
	}
}

// ------------------------------------------------------------------
