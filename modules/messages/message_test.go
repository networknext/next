package messages_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/messages"

	"github.com/stretchr/testify/assert"
)

func MessageReadWriteTest[M messages.Message](writeMessage M, readMessage M, t *testing.T) {

	const BufferSize = 10 * 1024

	buffer := make([]byte, BufferSize)

	buffer = writeMessage.Write(buffer[:])

	err := readMessage.Read(buffer)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

func GenerateRandomCostMatrixStatMessage() messages.CostMatrixStatsMessage {

	message := messages.CostMatrixStatsMessage{
		Version:        byte(common.RandomInt(messages.CostMatrixStatsMessageVersion_Min, messages.CostMatrixStatsMessageVersion_Max)),
		Timestamp:      uint64(time.Now().Unix()),
		Bytes:          rand.Int(),
		NumRelays:      rand.Int(),
		NumDestRelays:  rand.Int(),
		NumDatacenters: rand.Int(),
	}

	return message
}

func GenerateRandomRouteMatrixStatMessage() messages.RouteMatrixStatsMessage {

	message := messages.RouteMatrixStatsMessage{
		Version:                 byte(common.RandomInt(messages.RouteMatrixStatsMessageVersion_Min, messages.RouteMatrixStatsMessageVersion_Max)),
		Timestamp:               uint64(time.Now().Unix()),
		Bytes:                   rand.Int(),
		NumRelays:               rand.Int(),
		NumDestRelays:           rand.Int(),
		NumFullRelays:           rand.Int(),
		NumDatacenters:          rand.Int(),
		TotalRoutes:             rand.Int(),
		AverageNumRoutes:        rand.Float32(),
		AverageRouteLength:      rand.Float32(),
		NoRoutePercent:          rand.Float32(),
		OneRoutePercent:         rand.Float32(),
		NoDirectRoutePercent:    rand.Float32(),
		RTTBucket_NoImprovement: rand.Float32(),
		RTTBucket_0_5ms:         rand.Float32(),
		RTTBucket_5_10ms:        rand.Float32(),
		RTTBucket_10_15ms:       rand.Float32(),
		RTTBucket_15_20ms:       rand.Float32(),
		RTTBucket_20_25ms:       rand.Float32(),
		RTTBucket_25_30ms:       rand.Float32(),
		RTTBucket_30_35ms:       rand.Float32(),
		RTTBucket_35_40ms:       rand.Float32(),
		RTTBucket_40_45ms:       rand.Float32(),
		RTTBucket_45_50ms:       rand.Float32(),
		RTTBucket_50ms_Plus:     rand.Float32(),
	}

	return message
}

func GenerateRandomPingStatMessage() messages.PingStatsMessage {

	message := messages.PingStatsMessage{
		Version:    byte(common.RandomInt(messages.PingStatsMessageVersion_Min, messages.PingStatsMessageVersion_Max)),
		Timestamp:  uint64(time.Now().Unix()),
		RelayA:     rand.Uint64(),
		RelayB:     rand.Uint64(),
		RTT:        rand.Float32(),
		Jitter:     rand.Float32(),
		PacketLoss: rand.Float32(),
		Routable:   common.RandomBool(),
	}

	return message
}

func GenerateRandomRelayUpdateMessage() messages.RelayUpdateMessage {

	message := messages.RelayUpdateMessage{
		Version:                   byte(common.RandomInt(messages.RelayUpdateMessageVersion_Min, messages.RelayUpdateMessageVersion_Max)),
		Timestamp:                 uint64(time.Now().Unix()),
		RelayId:                   rand.Uint64(),
		SessionCount:              rand.Uint32(),
		MaxSessions:               rand.Uint32(),
		EnvelopeBandwidthUpKbps:   rand.Uint32(),
		EnvelopeBandwidthDownKbps: rand.Uint32(),
		ActualBandwidthUpKbps:     rand.Uint32(),
		ActualBandwidthDownKbps:   rand.Uint32(),
		RelayFlags:                rand.Uint64(),
		NumRelayCounters:          constants.NumRelayCounters,
	}

	for i := 0; i < constants.NumRelayCounters; i++ {
		message.RelayCounters[i] = rand.Uint64()
	}

	return message
}

func GenerateRandomDatabaseUpdateMessage() messages.DatabaseUpdateMessage {

	message := messages.DatabaseUpdateMessage{
		Version:        byte(common.RandomInt(messages.DatabaseUpdateMessageVersion_Min, messages.DatabaseUpdateMessageVersion_Max)),
		Timestamp:      uint64(time.Now().Unix()),
		DatabaseSize:   rand.Uint32(),
		NumRelays:      rand.Uint32(),
		NumDatacenters: rand.Uint32(),
		NumSellers:     rand.Uint32(),
		NumBuyers:      rand.Uint32(),
	}

	return message
}

func GenerateRandomMatchDataMessage() messages.MatchDataMessage {

	numMatchValues := rand.Intn(65)

	matchValues := [constants.MaxMatchValues]float64{}

	for i := 0; i < numMatchValues; i++ {
		matchValues[i] = rand.Float64()
	}

	return messages.MatchDataMessage{
		Version:        byte(common.RandomInt(messages.MatchDataMessageVersion_Min, messages.MatchDataMessageVersion_Max)),
		Timestamp:      uint64(time.Now().Unix()),
		Type:           rand.Uint64(),
		BuyerId:        rand.Uint64(),
		ServerAddress:  common.RandomAddress(),
		DatacenterId:   rand.Uint64(),
		UserHash:       rand.Uint64(),
		SessionId:      rand.Uint64(),
		MatchId:        rand.Uint64(),
		NumMatchValues: uint32(numMatchValues),
		MatchValues:    matchValues,
	}
}

func GenerateRandomServerInitMessage() messages.ServerInitMessage {

	message := messages.ServerInitMessage{
		Version:          byte(common.RandomInt(messages.ServerInitMessageVersion_Min, messages.ServerInitMessageVersion_Max)),
		SDKVersion_Major: 5,
		SDKVersion_Minor: 0,
		SDKVersion_Patch: 0,
		BuyerId:          rand.Uint64(),
		DatacenterId:     rand.Uint64(),
		DatacenterName:   common.RandomString(constants.MaxDatacenterNameLength),
	}

	return message
}

func GenerateRandomServerUpdateMessage() messages.ServerUpdateMessage {

	message := messages.ServerUpdateMessage{
		Version:          byte(common.RandomInt(messages.ServerUpdateMessageVersion_Min, messages.ServerUpdateMessageVersion_Max)),
		SDKVersion_Major: 5,
		SDKVersion_Minor: 0,
		SDKVersion_Patch: 0,
		BuyerId:          rand.Uint64(),
		DatacenterId:     rand.Uint64(),
	}

	return message
}

func GenerateRandomSessionUpdateMessage() messages.SessionUpdateMessage {

	message := messages.SessionUpdateMessage{

		Version: byte(common.RandomInt(messages.SessionUpdateMessageVersion_Min, messages.SessionUpdateMessageVersion_Max)),

		// always

		Timestamp:        rand.Uint64(),
		SessionId:        rand.Uint64(),
		SliceNumber:      rand.Uint32(),
		RealPacketLoss:   float32(common.RandomInt(0, 100)),
		RealJitter:       float32(common.RandomInt(0, 1000)),
		RealOutOfOrder:   float32(common.RandomInt(0, 100)),
		SessionFlags:     rand.Uint64(),
		GameEvents:       rand.Uint64(),
		DirectRTT:        float32(common.RandomInt(0, 1000)),
		DirectJitter:     float32(common.RandomInt(0, 1000)),
		DirectPacketLoss: float32(common.RandomInt(0, 100)),
		DirectKbpsUp:     rand.Uint32(),
		DirectKbpsDown:   rand.Uint32(),
	}

	// next only

	if (message.SessionFlags & constants.SessionFlags_Next) != 0 {
		message.NextRTT = float32(common.RandomInt(0, 1000))
		message.NextJitter = float32(common.RandomInt(0, 1000))
		message.NextPacketLoss = float32(common.RandomInt(0, 100))
		message.NextKbpsUp = rand.Uint32()
		message.NextKbpsDown = rand.Uint32()
		message.NextPredictedRTT = uint32(common.RandomInt(0, 1000))
		message.NextNumRouteRelays = uint32(common.RandomInt(0, constants.MaxRouteRelays))
		for i := 0; i < int(message.NextNumRouteRelays); i++ {
			message.NextRouteRelayId[i] = rand.Uint64()
		}
	}

	// first slice only

	if message.SliceNumber == 0 {
		message.NumTags = byte(common.RandomInt(0, constants.MaxTags))
		for i := 0; i < int(message.NumTags); i++ {
			message.Tags[i] = rand.Uint64()
		}
	}

	// summary only

	if (message.SessionFlags & constants.SessionFlags_Summary) != 0 {
		message.DatacenterId = rand.Uint64()
		message.BuyerId = rand.Uint64()
		message.UserHash = rand.Uint64()
		message.Latitude = float32(common.RandomInt(-90, +90))
		message.Longitude = float32(common.RandomInt(-180, +180))
		message.ClientAddress = common.RandomAddress()
		message.ServerAddress = common.RandomAddress()
		message.ConnectionType = uint8(common.RandomInt(0, 255))
		message.PlatformType = uint8(common.RandomInt(0, 255))
		message.SDKVersion_Major = uint8(common.RandomInt(0, 255))
		message.SDKVersion_Minor = uint8(common.RandomInt(0, 255))
		message.SDKVersion_Patch = uint8(common.RandomInt(0, 255))
		message.ClientToServerPacketsSent = rand.Uint64()
		message.ServerToClientPacketsSent = rand.Uint64()
		message.ClientToServerPacketsLost = rand.Uint64()
		message.ServerToClientPacketsLost = rand.Uint64()
		message.ClientToServerPacketsOutOfOrder = rand.Uint64()
		message.ServerToClientPacketsOutOfOrder = rand.Uint64()
		message.SessionDuration = rand.Uint32()
		message.TotalEnvelopeBytesUp = rand.Uint64()
		message.TotalEnvelopeBytesDown = rand.Uint64()
		message.DurationOnNext = rand.Uint32()
		message.StartTimestamp = rand.Uint64()
	}

	return message
}

func GenerateRandomPortalMessage() messages.PortalMessage {

	message := messages.PortalMessage{
		Version: byte(common.RandomInt(messages.PortalMessageVersion_Min, messages.PortalMessageVersion_Max)),

		SDKVersion_Major: uint8(common.RandomInt(0, 255)),
		SDKVersion_Minor: uint8(common.RandomInt(0, 255)),
		SDKVersion_Patch: uint8(common.RandomInt(0, 255)),
		SessionId:        rand.Uint64(),
		MatchId:          rand.Uint64(),
		BuyerId:          rand.Uint64(),
		DatacenterId:     rand.Uint64(),
		Latitude:         float32(common.RandomInt(-90, +90)),
		Longitude:        float32(common.RandomInt(-180, +180)),
		ClientAddress:    common.RandomAddress(),
		ServerAddress:    common.RandomAddress(),

		SliceNumber:      rand.Uint32(),
		DirectRTT:        float32(common.RandomInt(0, 1000)),
		DirectJitter:     float32(common.RandomInt(0, 1000)),
		DirectPacketLoss: float32(common.RandomInt(0, 100)),
		DirectKbpsUp:     rand.Uint32(),
		DirectKbpsDown:   rand.Uint32(),

		SessionFlags: rand.Uint64(),
		GameEvents:   rand.Uint64(),

		RealPacketLoss: float32(common.RandomInt(0, 100)),
		RealJitter:     float32(common.RandomInt(0, 1000)),
		RealOutOfOrder: float32(common.RandomInt(0, 100)),

		NumNearRelays: uint32(common.RandomInt(0, constants.MaxNearRelays)),
	}

	if (message.SessionFlags & constants.SessionFlags_Next) != 0 {
		message.NextRTT = float32(common.RandomInt(0, 1000))
		message.NextJitter = float32(common.RandomInt(0, 1000))
		message.NextPacketLoss = float32(common.RandomInt(0, 100))
		message.NextKbpsUp = rand.Uint32()
		message.NextKbpsDown = rand.Uint32()
		message.NextPredictedRTT = uint32(common.RandomInt(0, 1000))
		message.NextNumRouteRelays = uint32(common.RandomInt(0, constants.MaxRouteRelays))
		for i := 0; i < int(message.NextNumRouteRelays); i++ {
			message.NextRouteRelayId[i] = rand.Uint64()
		}
	}

	for i := 0; i < int(message.NumNearRelays); i++ {
		message.NearRelayId[i] = rand.Uint64()
		message.NearRelayRTT[i] = byte(common.RandomInt(0, 255))
		message.NearRelayJitter[i] = byte(common.RandomInt(0, 255))
		message.NearRelayPacketLoss[i] = float32(common.RandomInt(0, 100))
		message.NearRelayRoutable[i] = common.RandomBool()
	}

	return message
}

func GenerateRandomNearRelayPingsMessage() messages.NearRelayPingsMessage {

	message := messages.NearRelayPingsMessage{
		Version: byte(common.RandomInt(messages.NearRelayPingsMessageVersion_Min, messages.NearRelayPingsMessageVersion_Max)),

		Timestamp: rand.Uint64(),

		BuyerId:        rand.Uint64(),
		SessionId:      rand.Uint64(),
		MatchId:        rand.Uint64(),
		UserHash:       rand.Uint64(),
		Latitude:       float32(common.RandomInt(-90, +90)),
		Longitude:      float32(common.RandomInt(-180, +180)),
		ClientAddress:  common.RandomAddress(),
		ConnectionType: byte(common.RandomInt(0, 255)),
		PlatformType:   byte(common.RandomInt(0, 255)),

		NumNearRelays: uint32(common.RandomInt(0, constants.MaxNearRelays)),
	}

	for i := 0; i < int(message.NumNearRelays); i++ {
		message.NearRelayId[i] = rand.Uint64()
		message.NearRelayRTT[i] = byte(common.RandomInt(0, 255))
		message.NearRelayJitter[i] = byte(common.RandomInt(0, 255))
		message.NearRelayPacketLoss[i] = float32(common.RandomInt(0, 100))
	}

	return message
}

// -----------------------------------------------------------

const NumIterations = 1000

func TestCostMatrixStatsMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomCostMatrixStatMessage()
		readMessage := messages.CostMatrixStatsMessage{}
		MessageReadWriteTest[*messages.CostMatrixStatsMessage](&writeMessage, &readMessage, t)
	}
}

func TestRouteMatrixStatsMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomRouteMatrixStatMessage()
		readMessage := messages.RouteMatrixStatsMessage{}
		MessageReadWriteTest[*messages.RouteMatrixStatsMessage](&writeMessage, &readMessage, t)
	}
}

func TestPingStatsMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomPingStatMessage()
		readMessage := messages.PingStatsMessage{}
		MessageReadWriteTest[*messages.PingStatsMessage](&writeMessage, &readMessage, t)
	}
}

func TestRelayUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomRelayUpdateMessage()
		readMessage := messages.RelayUpdateMessage{}
		MessageReadWriteTest[*messages.RelayUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestDatabaseUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomDatabaseUpdateMessage()
		readMessage := messages.DatabaseUpdateMessage{}
		MessageReadWriteTest[*messages.DatabaseUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestServerInitMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomServerInitMessage()
		readMessage := messages.ServerInitMessage{}
		MessageReadWriteTest[*messages.ServerInitMessage](&writeMessage, &readMessage, t)
	}
}

func TestServerUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomServerUpdateMessage()
		readMessage := messages.ServerUpdateMessage{}
		MessageReadWriteTest[*messages.ServerUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestMatchDataMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomMatchDataMessage()
		readMessage := messages.MatchDataMessage{}
		MessageReadWriteTest[*messages.MatchDataMessage](&writeMessage, &readMessage, t)
	}
}

func TestSessionUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomSessionUpdateMessage()
		readMessage := messages.SessionUpdateMessage{}
		MessageReadWriteTest[*messages.SessionUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestPortalMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomPortalMessage()
		readMessage := messages.PortalMessage{}
		MessageReadWriteTest[*messages.PortalMessage](&writeMessage, &readMessage, t)
	}
}

func TestNearRelayPingsMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomNearRelayPingsMessage()
		readMessage := messages.NearRelayPingsMessage{}
		MessageReadWriteTest[*messages.NearRelayPingsMessage](&writeMessage, &readMessage, t)
	}
}
