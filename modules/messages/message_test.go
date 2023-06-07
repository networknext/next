package messages_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/networknext/accelerate/modules/common"
	"github.com/networknext/accelerate/modules/constants"
	"github.com/networknext/accelerate/modules/messages"

	"github.com/stretchr/testify/assert"
)

func MessageReadWriteTest[M messages.Message](writeMessage M, readMessage M, t *testing.T) {

	buffer := make([]byte, writeMessage.GetMaxSize())

	messageData := writeMessage.Write(buffer[:])

	err := readMessage.Read(messageData)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

func GenerateRandomAnalyticsCostMatrixUpdateMessage() messages.AnalyticsCostMatrixUpdateMessage {

	message := messages.AnalyticsCostMatrixUpdateMessage{
		Version:        byte(common.RandomInt(messages.AnalyticsCostMatrixUpdateMessageVersion_Min, messages.AnalyticsCostMatrixUpdateMessageVersion_Max)),
		Timestamp:      uint64(time.Now().Unix()),
		CostMatrixSize: rand.Int(),
		NumRelays:      rand.Int(),
		NumDestRelays:  rand.Int(),
		NumDatacenters: rand.Int(),
	}

	return message
}

func GenerateRandomAnalyticsRouteMatrixUpdateMessage() messages.AnalyticsRouteMatrixUpdateMessage {

	message := messages.AnalyticsRouteMatrixUpdateMessage{
		Version:                 byte(common.RandomInt(messages.AnalyticsRouteMatrixUpdateMessageVersion_Min, messages.AnalyticsRouteMatrixUpdateMessageVersion_Max)),
		Timestamp:               uint64(time.Now().Unix()),
		RouteMatrixSize:         rand.Int(),
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

func GenerateRandomAnalyticsRelayToRelayPingMessage() messages.AnalyticsRelayToRelayPingMessage {

	message := messages.AnalyticsRelayToRelayPingMessage{
		Version:    byte(common.RandomInt(messages.AnalyticsRelayToRelayPingMessageVersion_Min, messages.AnalyticsRelayToRelayPingMessageVersion_Max)),
		Timestamp:  uint64(time.Now().Unix()),
		RelayA:     rand.Uint64(),
		RelayB:     rand.Uint64(),
		RTT:        uint8(common.RandomInt(0, 255)),
		Jitter:     uint8(common.RandomInt(0, 255)),
		PacketLoss: rand.Float32(),
	}

	return message
}

func GenerateRandomAnalyticsRelayUpdateMessage() messages.AnalyticsRelayUpdateMessage {

	message := messages.AnalyticsRelayUpdateMessage{

		Version:                   byte(common.RandomInt(messages.AnalyticsRelayUpdateMessageVersion_Min, messages.AnalyticsRelayUpdateMessageVersion_Max)),
		Timestamp:                 uint64(time.Now().Unix()),
		RelayId:                   rand.Uint64(),
		SessionCount:              rand.Uint32(),
		MaxSessions:               rand.Uint32(),

		EnvelopeBandwidthUpKbps:   rand.Uint32(),
		EnvelopeBandwidthDownKbps: rand.Uint32(),
		PacketsSentPerSecond:      float32(common.RandomInt(0,1000)),
		PacketsReceivedPerSecond:  float32(common.RandomInt(0,1000)),
		BandwidthSentKbps:         float32(common.RandomInt(0,1000)),
		BandwidthReceivedKbps:     float32(common.RandomInt(0,1000)),
		NearPingsPerSecond:		   float32(common.RandomInt(0,1000)),
		RelayPingsPerSecond:       float32(common.RandomInt(0,1000)),

		RelayFlags:                rand.Uint64(),
		StartTime:                 rand.Uint64(),
		CurrentTime:               rand.Uint64(),

		NumRelayCounters:          constants.NumRelayCounters,
	}

	for i := 0; i < constants.NumRelayCounters; i++ {
		message.RelayCounters[i] = rand.Uint64()
	}

	return message
}

func GenerateRandomAnalyticsDatabaseUpdateMessage() messages.AnalyticsDatabaseUpdateMessage {

	message := messages.AnalyticsDatabaseUpdateMessage{
		Version:        byte(common.RandomInt(messages.AnalyticsDatabaseUpdateMessageVersion_Min, messages.AnalyticsDatabaseUpdateMessageVersion_Max)),
		Timestamp:      uint64(time.Now().Unix()),
		DatabaseSize:   rand.Uint32(),
		NumRelays:      rand.Uint32(),
		NumDatacenters: rand.Uint32(),
		NumSellers:     rand.Uint32(),
		NumBuyers:      rand.Uint32(),
	}

	return message
}

func GenerateRandomAnalyticsMatchDataMessage() messages.AnalyticsMatchDataMessage {

	numMatchValues := rand.Intn(65)

	matchValues := [constants.MaxMatchValues]float64{}

	for i := 0; i < numMatchValues; i++ {
		matchValues[i] = rand.Float64()
	}

	return messages.AnalyticsMatchDataMessage{
		Version:        byte(common.RandomInt(messages.AnalyticsMatchDataMessageVersion_Min, messages.AnalyticsMatchDataMessageVersion_Max)),
		Timestamp:      uint64(time.Now().Unix()),
		Type:           rand.Uint64(),
		BuyerId:        rand.Uint64(),
		ServerAddress:  common.RandomAddress(),
		DatacenterId:   rand.Uint64(),
		SessionId:      rand.Uint64(),
		MatchId:        rand.Uint64(),
		NumMatchValues: uint32(numMatchValues),
		MatchValues:    matchValues,
	}
}

func GenerateRandomAnalyticsServerInitMessage() messages.AnalyticsServerInitMessage {

	message := messages.AnalyticsServerInitMessage{
		Version:          byte(common.RandomInt(messages.AnalyticsServerInitMessageVersion_Min, messages.AnalyticsServerInitMessageVersion_Max)),
		Timestamp:        uint64(time.Now().Unix()),
		SDKVersion_Major: 5,
		SDKVersion_Minor: 0,
		SDKVersion_Patch: 0,
		BuyerId:          rand.Uint64(),
		DatacenterId:     rand.Uint64(),
		DatacenterName:   common.RandomString(constants.MaxDatacenterNameLength),
	}

	return message
}

func GenerateRandomAnalyticsServerUpdateMessage() messages.AnalyticsServerUpdateMessage {

	message := messages.AnalyticsServerUpdateMessage{
		Version:          byte(common.RandomInt(messages.AnalyticsServerUpdateMessageVersion_Min, messages.AnalyticsServerUpdateMessageVersion_Max)),
		SDKVersion_Major: 5,
		SDKVersion_Minor: 0,
		SDKVersion_Patch: 0,
		BuyerId:          rand.Uint64(),
		DatacenterId:     rand.Uint64(),
		ServerAddress:    common.RandomAddress(),
	}

	return message
}

func GenerateRandomAnalyticsSessionUpdateMessage() messages.AnalyticsSessionUpdateMessage {

	message := messages.AnalyticsSessionUpdateMessage{

		Version: byte(common.RandomInt(messages.AnalyticsSessionUpdateMessageVersion_Min, messages.AnalyticsSessionUpdateMessageVersion_Max)),

		// always

		Timestamp:        rand.Uint64(),
		SessionId:        rand.Uint64(),
		SliceNumber:      rand.Uint32(),
		RealPacketLoss:   float32(common.RandomInt(0, 100)),
		RealJitter:       float32(common.RandomInt(0, 1000)),
		RealOutOfOrder:   float32(common.RandomInt(0, 100)),
		SessionFlags:     rand.Uint64(),
		SessionEvents:    rand.Uint64(),
		InternalEvents:   rand.Uint64(),
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

	return message
}

func GenerateRandomAnalyticsSessionSummaryMessage() messages.AnalyticsSessionSummaryMessage {

	message := messages.AnalyticsSessionSummaryMessage{

		Version:                         byte(common.RandomInt(messages.AnalyticsSessionSummaryMessageVersion_Min, messages.AnalyticsSessionSummaryMessageVersion_Max)),
		Timestamp:                       rand.Uint64(),
		SessionId:                       rand.Uint64(),
		DatacenterId:                    rand.Uint64(),
		BuyerId:                         rand.Uint64(),
		MatchId:                         rand.Uint64(),
		UserHash:                        rand.Uint64(),
		Latitude:                        float32(common.RandomInt(-90, +90)),
		Longitude:                       float32(common.RandomInt(-180, +180)),
		ClientAddress:                   common.RandomAddress(),
		ServerAddress:                   common.RandomAddress(),
		ConnectionType:                  uint8(common.RandomInt(0, 255)),
		PlatformType:                    uint8(common.RandomInt(0, 255)),
		SDKVersion_Major:                uint8(common.RandomInt(0, 255)),
		SDKVersion_Minor:                uint8(common.RandomInt(0, 255)),
		SDKVersion_Patch:                uint8(common.RandomInt(0, 255)),
		ClientToServerPacketsSent:       rand.Uint64(),
		ServerToClientPacketsSent:       rand.Uint64(),
		ClientToServerPacketsLost:       rand.Uint64(),
		ServerToClientPacketsLost:       rand.Uint64(),
		ClientToServerPacketsOutOfOrder: rand.Uint64(),
		ServerToClientPacketsOutOfOrder: rand.Uint64(),
		SessionDuration:                 rand.Uint32(),
		TotalEnvelopeBytesUp:            rand.Uint64(),
		TotalEnvelopeBytesDown:          rand.Uint64(),
		DurationOnNext:                  rand.Uint32(),
		StartTimestamp:                  rand.Uint64(),
	}

	return message
}

func GenerateRandomPortalServerUpdateMessage() messages.PortalServerUpdateMessage {

	message := messages.PortalServerUpdateMessage{
		Version:          byte(common.RandomInt(messages.PortalServerUpdateMessageVersion_Min, messages.PortalServerUpdateMessageVersion_Max)),
		SDKVersion_Major: uint8(common.RandomInt(0, 255)),
		SDKVersion_Minor: uint8(common.RandomInt(0, 255)),
		SDKVersion_Patch: uint8(common.RandomInt(0, 255)),
		MatchId:          rand.Uint64(),
		BuyerId:          rand.Uint64(),
		DatacenterId:     rand.Uint64(),
		NumSessions:      rand.Uint32(),
		ServerAddress:    common.RandomAddress(),
	}

	return message
}

func GenerateRandomPortalSessionUpdateMessage() messages.PortalSessionUpdateMessage {

	message := messages.PortalSessionUpdateMessage{
		Version: byte(common.RandomInt(messages.PortalSessionUpdateMessageVersion_Min, messages.PortalSessionUpdateMessageVersion_Max)),

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

		SessionFlags:   rand.Uint64(),
		SessionEvents:  rand.Uint64(),
		InternalEvents: rand.Uint64(),

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

func GenerateRandomPortalRelayUpdateMessage() messages.PortalRelayUpdateMessage {

	message := messages.PortalRelayUpdateMessage{
		Version:                   byte(common.RandomInt(messages.PortalRelayUpdateMessageVersion_Min, messages.PortalRelayUpdateMessageVersion_Max)),
		Timestamp:                 uint64(time.Now().Unix()),
		RelayId:                   rand.Uint64(),
		SessionCount:              rand.Uint32(),
		MaxSessions:               rand.Uint32(),
		EnvelopeBandwidthUpKbps:   rand.Uint32(),
		EnvelopeBandwidthDownKbps: rand.Uint32(),
		// todo: add new stats here
		RelayFlags:                rand.Uint64(),
		RelayAddress:              common.RandomAddress(),
		RelayVersion:              common.RandomString(constants.MaxRelayVersionLength),
		StartTime:                 rand.Uint64(),
		CurrentTime:               rand.Uint64(),
	}

	return message
}

func GenerateRandomPortalNearRelayUpdateMessage() messages.PortalNearRelayUpdateMessage {

	message := messages.PortalNearRelayUpdateMessage{
		Version:       byte(common.RandomInt(messages.PortalNearRelayUpdateMessageVersion_Min, messages.PortalNearRelayUpdateMessageVersion_Max)),
		Timestamp:     rand.Uint64(),
		BuyerId:       rand.Uint64(),
		SessionId:     rand.Uint64(),
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

func GenerateRandomPortalMapUpdateMessage() messages.PortalMapUpdateMessage {

	message := messages.PortalMapUpdateMessage{
		Version:   byte(common.RandomInt(messages.PortalMapUpdateMessageVersion_Min, messages.PortalMapUpdateMessageVersion_Max)),
		SessionId: rand.Uint64(),
		Latitude:  float32(common.RandomInt(-90, +90)),
		Longitude: float32(common.RandomInt(-180, +180)),
		Next:      common.RandomBool(),
	}

	return message
}

func GenerateRandomAnalyticsNearRelayUpdateMessage() messages.AnalyticsNearRelayUpdateMessage {

	message := messages.AnalyticsNearRelayUpdateMessage{
		Version:        byte(common.RandomInt(messages.AnalyticsNearRelayUpdateMessageVersion_Min, messages.AnalyticsNearRelayUpdateMessageVersion_Max)),
		Timestamp:      rand.Uint64(),
		BuyerId:        rand.Uint64(),
		SessionId:      rand.Uint64(),
		MatchId:        rand.Uint64(),
		UserHash:       rand.Uint64(),
		Latitude:       float32(common.RandomInt(-90, +90)),
		Longitude:      float32(common.RandomInt(-180, +180)),
		ClientAddress:  common.RandomAddress(),
		ConnectionType: byte(common.RandomInt(0, 255)),
		PlatformType:   byte(common.RandomInt(0, 255)),
		NumNearRelays:  uint32(common.RandomInt(0, constants.MaxNearRelays)),
	}

	for i := 0; i < int(message.NumNearRelays); i++ {
		message.NearRelayId[i] = rand.Uint64()
		message.NearRelayRTT[i] = byte(common.RandomInt(0, 255))
		message.NearRelayJitter[i] = byte(common.RandomInt(0, 255))
		message.NearRelayPacketLoss[i] = float32(common.RandomInt(0, 100))
	}

	return message
}

// -------------------------------------------------------------------------------------------------------------------

const NumIterations = 1000

func TestPortalServerUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomPortalServerUpdateMessage()
		readMessage := messages.PortalServerUpdateMessage{}
		MessageReadWriteTest[*messages.PortalServerUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestPortalSessionUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomPortalSessionUpdateMessage()
		readMessage := messages.PortalSessionUpdateMessage{}
		MessageReadWriteTest[*messages.PortalSessionUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestPortalRelayUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomPortalRelayUpdateMessage()
		readMessage := messages.PortalRelayUpdateMessage{}
		MessageReadWriteTest[*messages.PortalRelayUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestPortalNearRelayUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomPortalNearRelayUpdateMessage()
		readMessage := messages.PortalNearRelayUpdateMessage{}
		MessageReadWriteTest[*messages.PortalNearRelayUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestPortalMapUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomPortalMapUpdateMessage()
		readMessage := messages.PortalMapUpdateMessage{}
		MessageReadWriteTest[*messages.PortalMapUpdateMessage](&writeMessage, &readMessage, t)
	}
}

// ------------------------------------------------------------------------------------------------------------------

func TestAnalyticsCostMatrixUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomAnalyticsCostMatrixUpdateMessage()
		readMessage := messages.AnalyticsCostMatrixUpdateMessage{}
		MessageReadWriteTest[*messages.AnalyticsCostMatrixUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestRouteMatrixStatsMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomAnalyticsRouteMatrixUpdateMessage()
		readMessage := messages.AnalyticsRouteMatrixUpdateMessage{}
		MessageReadWriteTest[*messages.AnalyticsRouteMatrixUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestAnalyticsRelayToRelayPingMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomAnalyticsRelayToRelayPingMessage()
		readMessage := messages.AnalyticsRelayToRelayPingMessage{}
		MessageReadWriteTest[*messages.AnalyticsRelayToRelayPingMessage](&writeMessage, &readMessage, t)
	}
}

func TestAnalyticsRelayUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomAnalyticsRelayUpdateMessage()
		readMessage := messages.AnalyticsRelayUpdateMessage{}
		MessageReadWriteTest[*messages.AnalyticsRelayUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestAnalyticsDatabaseUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomAnalyticsDatabaseUpdateMessage()
		readMessage := messages.AnalyticsDatabaseUpdateMessage{}
		MessageReadWriteTest[*messages.AnalyticsDatabaseUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestAnalyticsServerInitMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomAnalyticsServerInitMessage()
		readMessage := messages.AnalyticsServerInitMessage{}
		MessageReadWriteTest[*messages.AnalyticsServerInitMessage](&writeMessage, &readMessage, t)
	}
}

func TestAnalyticsServerUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomAnalyticsServerUpdateMessage()
		readMessage := messages.AnalyticsServerUpdateMessage{}
		MessageReadWriteTest[*messages.AnalyticsServerUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestAnalyticsMatchDataMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomAnalyticsMatchDataMessage()
		readMessage := messages.AnalyticsMatchDataMessage{}
		MessageReadWriteTest[*messages.AnalyticsMatchDataMessage](&writeMessage, &readMessage, t)
	}
}

func TestAnalyticsSessionUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomAnalyticsSessionUpdateMessage()
		readMessage := messages.AnalyticsSessionUpdateMessage{}
		MessageReadWriteTest[*messages.AnalyticsSessionUpdateMessage](&writeMessage, &readMessage, t)
	}
}

func TestAnalyticsNearRelayUpdateMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomAnalyticsNearRelayUpdateMessage()
		readMessage := messages.AnalyticsNearRelayUpdateMessage{}
		MessageReadWriteTest[*messages.AnalyticsNearRelayUpdateMessage](&writeMessage, &readMessage, t)
	}
}
