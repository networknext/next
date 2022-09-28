package messages_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/messages"

	"github.com/stretchr/testify/assert"
)

func MessageReadWriteTest[M messages.Message](writeMessage messages.Message, readMessage messages.Message, t *testing.T) {

	const BufferSize = 10 * 1024

	buffer := make([]byte, BufferSize)

	buffer = writeMessage.Write(buffer[:])

	err := readMessage.Read(buffer)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

func GenerateRandomCostMatrixStatMessage() messages.CostMatrixStatsMessage {

	rand.Seed(time.Now().UnixNano())

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

	rand.Seed(time.Now().UnixNano())

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

func GenerateRandomPingStatMessage(routeable bool) messages.PingStatsMessage {

	rand.Seed(time.Now().UnixNano())

	return messages.PingStatsMessage{
		Version:    messages.PingStatsMessageVersion,
		Timestamp:  uint64(time.Now().Unix()),
		RelayA:     rand.Uint64(),
		RelayB:     rand.Uint64(),
		RTT:        rand.Float32(),
		Jitter:     rand.Float32(),
		PacketLoss: rand.Float32(),
		Routable:   routeable,
	}
}

func GenerateRandomRelayStatMessage(full bool) messages.RelayStatsMessage {

	rand.Seed(time.Now().UnixNano())

	return messages.RelayStatsMessage{
		Version:                  messages.RelayStatsMessageVersion,
		Timestamp:                uint64(time.Now().Unix()),
		ID:                       rand.Uint64(),
		NumSessions:              rand.Uint32(),
		MaxSessions:              rand.Uint32(),
		NumRoutable:              rand.Uint32(),
		NumUnroutable:            rand.Uint32(),
		Full:                     full,
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

func GenerateRandomMatchDataMessage() messages.MatchDataMessage {

	rand.Seed(time.Now().UnixNano())

	matchValues := [messages.MatchDataMaxMatchValues]float64{}

	for i := 0; i < messages.MatchDataMaxMatchValues; i++ {
		matchValues[i] = rand.Float64()
	}

	return messages.MatchDataMessage{
		Version:        messages.MatchDataMessageVersion,
		Timestamp:      uint64(time.Now().Unix()),
		BuyerId:        rand.Uint64(),
		ServerAddress:  common.RandomString(messages.MatchDataMaxAddressLength),
		DatacenterId:   rand.Uint64(),
		UserHash:       rand.Uint64(),
		SessionId:      rand.Uint64(),
		MatchId:        rand.Uint64(),
		NumMatchValues: rand.Int31(),
		MatchValues:    matchValues,
	}
}

const NumIterations = 10000

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
		writeMessage := GenerateRandomPingStatMessage(true)
		readMessage := messages.PingStatsMessage{}
		MessageReadWriteTest[*messages.PingStatsMessage](&writeMessage, &readMessage, t)
	}
}

func TestRelayStatsMessage(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeMessage := GenerateRandomRelayStatMessage(true)
		readMessage := messages.RelayStatsMessage{}
		MessageReadWriteTest[*messages.RelayStatsMessage](&writeMessage, &readMessage, t)
	}
}

// todo: finish testing the messages

/*
func TestMatchDataMessage(t *testing.T) {

	writeMessage := GenerateRandomMatchDataMessage()

	readMessage := messages.MatchDataMessage{}

	MessageReadWriteTest[*messages.MatchDataMessage](&writeMessage, &readMessage, t)
}

func TestBillingMessage(t *testing.T) {

	writeMessage := messages.BillingMessage{}

	readMessage := messages.BillingMessage{}

	MessageReadWriteTest[*messages.BillingMessage](&writeMessage, &readMessage, t)
}

func TestSummaryMessage(t *testing.T) {

	writeMessage := messages.SummaryMessage{}

	readMessage := messages.SummaryMessage{}

	MessageReadWriteTest[*messages.SummaryMessage](&writeMessage, &readMessage, t)
}
*/

// todo: test the server init message

// todo: test the server update message
