package messages_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/networknext/backend/modules/common"
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

	return messages.CostMatrixStatsMessage{
		Version:        messages.CostMatrixStatsMessageVersion,
		Timestamp:      uint64(time.Now().Unix()),
		Bytes:          rand.Int(),
		NumRelays:      rand.Int(),
		NumDestRelays:  rand.Int(),
		NumDatacenters: rand.Int(),
	}
}

func GenerateRandomRouteMatrixStatMessage() messages.RouteMatrixStatsMessage {

	return messages.RouteMatrixStatsMessage{
		Version:                 messages.RouteMatrixStatsMessageVersion,
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
}

func GenerateRandomPingStatMessage() messages.PingStatsMessage {

	return messages.PingStatsMessage{
		Version:    messages.PingStatsMessageVersion,
		Timestamp:  uint64(time.Now().Unix()),
		RelayA:     rand.Uint64(),
		RelayB:     rand.Uint64(),
		RTT:        rand.Float32(),
		Jitter:     rand.Float32(),
		PacketLoss: rand.Float32(),
		Routable:   common.RandomBool(),
	}
}

func GenerateRandomRelayStatMessage() messages.RelayStatsMessage {

	return messages.RelayStatsMessage{
		Version:                  messages.RelayStatsMessageVersion,
		Timestamp:                uint64(time.Now().Unix()),
		ID:                       rand.Uint64(),
		NumSessions:              rand.Uint32(),
		MaxSessions:              rand.Uint32(),
		NumRoutable:              rand.Uint32(),
		NumUnroutable:            rand.Uint32(),
		Full:                     common.RandomBool(),
		CPUUsage:                 rand.Float32(),
		BandwidthSentPercent:     rand.Float32(),
		BandwidthReceivedPercent: rand.Float32(),
		EnvelopeSentPercent:      rand.Float32(),
		EnvelopeReceivedPercent:  rand.Float32(),
		BandwidthSentMbps:        rand.Float32(),
		BandwidthReceivedMbps:    rand.Float32(),
		EnvelopeSentMbps:         rand.Float32(),
		EnvelopeReceivedMbps:     rand.Float32(),
	}
}

func GenerateRandomUptimeStatMessage() messages.UptimeStatsMessage {

	return messages.UptimeStatsMessage{
		Version:      messages.UptimeStatsMessageVersion,
		Timestamp:    uint64(time.Now().Unix()),
		ServiceName:  common.RandomString(messages.MaxServiceNameLength),
		Up:           common.RandomBool(),
		ResponseTime: common.RandomInt(0, 10000),
	}
}

func GenerateRandomMatchDataMessage() messages.MatchDataMessage {

	numMatchValues := rand.Intn(65)

	matchValues := [messages.MatchDataMaxMatchValues]float64{}

	for i := 0; i < numMatchValues; i++ {
		matchValues[i] = rand.Float64()
	}

	return messages.MatchDataMessage{
		Version:        messages.MatchDataMessageVersion,
		Timestamp:      uint64(time.Now().Unix()),
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

	return messages.ServerInitMessage{
		MessageVersion:   messages.ServerInitMessageVersion,
		SDKVersion_Major: 5,
		SDKVersion_Minor: 0,
		SDKVersion_Patch: 0,
		BuyerId:          rand.Uint64(),
		DatacenterId:     rand.Uint64(),
		DatacenterName:   common.RandomString(messages.ServerInitMaxDatacenterNameLength),
	}
}

func GenerateRandomServerUpdateMessage() messages.ServerUpdateMessage {

	return messages.ServerUpdateMessage{
		MessageVersion:   messages.ServerInitMessageVersion,
		SDKVersion_Major: 5,
		SDKVersion_Minor: 0,
		SDKVersion_Patch: 0,
		BuyerId:          rand.Uint64(),
		DatacenterId:     rand.Uint64(),
	}
}

func GenerateRandomSessionUpdateMessage() messages.SessionUpdateMessage {

	message := messages.SessionUpdateMessage{

		// always

		Version:             messages.SessionUpdateMessageVersion,
		Timestamp:           rand.Uint64(),
		SessionId:           rand.Uint64(),
		SliceNumber:         rand.Uint32(),
		DirectMinRTT:        int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxRTT+10)),
		DirectMaxRTT:        int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxRTT+10)),
		DirectPrimeRTT:      int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxRTT+10)),
		DirectJitter:        int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxJitter+10)),
		DirectPacketLoss:    int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxPacketLoss+10)),
		RealPacketLoss:      int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxPacketLoss+10)),
		RealPacketLoss_Frac: uint32(common.RandomInt(0, 300)),
		RealJitter:          uint32(common.RandomInt(0, messages.SessionUpdateMessageMaxJitter+10)),
		Next:                common.RandomBool(),
		Flagged:             common.RandomBool(),
		Summary:             common.RandomBool(),
		UseDebug:            common.RandomBool(),
		Debug:               common.RandomString(messages.SessionUpdateMessageMaxDebugLength + 10),
		RouteDiversity:      int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxRouteDiversity+10)),
		UserFlags:           rand.Uint64(),
		TryBeforeYouBuy:     common.RandomBool(),

		// error state only

		FallbackToDirect:     common.RandomBool(),
		MultipathVetoed:      common.RandomBool(),
		Mispredicted:         common.RandomBool(),
		Vetoed:               common.RandomBool(),
		LatencyWorse:         common.RandomBool(),
		NoRoute:              common.RandomBool(),
		NextLatencyTooHigh:   common.RandomBool(),
		CommitVeto:           common.RandomBool(),
		UnknownDatacenter:    common.RandomBool(),
		DatacenterNotEnabled: common.RandomBool(),
		BuyerNotLive:         common.RandomBool(),
		StaleRouteMatrix:     common.RandomBool(),
	}

	// first slice and summary slice

	if message.SliceNumber == 0 || message.Summary {

		message.DatacenterId = rand.Uint64()
		message.BuyerId = rand.Uint64()
		message.UserHash = rand.Uint64()
		message.EnvelopeBytesUp = rand.Uint64()
		message.EnvelopeBytesDown = rand.Uint64()
		message.Latitude = rand.Float32()
		message.Longitude = rand.Float32()
		message.ClientAddress = common.RandomAddress()
		message.ServerAddress = common.RandomAddress()
		message.ISP = common.RandomString(messages.SessionUpdateMessageMaxISPLength)
		message.ConnectionType = int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxConnectionType+10))
		message.PlatformType = int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxPlatformType+10))
		message.NumTags = int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxTags+10))
		message.ABTest = common.RandomBool()
		message.Pro = common.RandomBool()
	}

	// summary slice only

	if message.Summary {

		message.ClientToServerPacketsSent = rand.Uint64()
		message.ServerToClientPacketsSent = rand.Uint64()
		message.ClientToServerPacketsLost = rand.Uint64()
		message.ServerToClientPacketsLost = rand.Uint64()
		message.ClientToServerPacketsOutOfOrder = rand.Uint64()
		message.ServerToClientPacketsOutOfOrder = rand.Uint64()
		message.NumNearRelays = int32(common.RandomInt(0, messages.SessionUpdateMessageMaxNearRelays))
		message.EverOnNext = common.RandomBool()
		message.SessionDuration = rand.Uint32()

		if message.EverOnNext {
			message.TotalPriceSum = rand.Uint64()
			message.EnvelopeBytesUpSum = rand.Uint64()
			message.EnvelopeBytesDownSum = rand.Uint64()
			message.DurationOnNext = rand.Uint32()
		}

		message.StartTimestamp = rand.Uint64()
	}

	// next only

	if message.Next {

		message.NextRTT = int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxRTT+10))
		message.NextJitter = int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxJitter+10))
		message.NextPacketLoss = int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxPacketLoss+10))
		message.PredictedNextRTT = int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxRTT+10))
		message.NearRelayRTT = int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxNearRelayRTT+10))
		message.NumNextRelays = int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxRelays+10))
		message.TotalPrice = rand.Uint64()
		message.Uncommitted = common.RandomBool()
		message.Multipath = common.RandomBool()
		message.RTTReduction = common.RandomBool()
		message.PacketLossReduction = common.RandomBool()
		message.RouteChanged = common.RandomBool()
		message.NextBytesUp = rand.Uint64()
		message.NextBytesDown = rand.Uint64()
	}

	message.Clamp()

	// post clamp

	if message.SliceNumber == 0 || message.Summary {

		for i := 0; i < int(message.NumTags); i++ {
			message.Tags[i] = rand.Uint64()
		}
	}

	if message.Summary {

		for i := 0; i < int(message.NumNearRelays); i++ {
			message.NearRelayIds[i] = rand.Uint64()
			message.NearRelayRTTs[i] = int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxNearRelayRTT+10))
			message.NearRelayJitters[i] = int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxJitter+10))
			message.NearRelayPacketLosses[i] = int32(common.RandomInt(-10, messages.SessionUpdateMessageMaxPacketLoss+10))
		}
	}

	if message.Next {

		for i := 0; i < int(message.NumNextRelays); i++ {
			message.NextRelays[i] = rand.Uint64()
			message.NextRelayPrice[i] = rand.Uint64()
		}
	}

	// clamp again for array entries

	message.Clamp()

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

func TestRelayStatsMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomRelayStatMessage()
		readMessage := messages.RelayStatsMessage{}
		MessageReadWriteTest[*messages.RelayStatsMessage](&writeMessage, &readMessage, t)
	}
}

func TestUptimeStatsMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomUptimeStatMessage()
		readMessage := messages.UptimeStatsMessage{}
		MessageReadWriteTest[*messages.UptimeStatsMessage](&writeMessage, &readMessage, t)
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
