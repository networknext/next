package messages

import (
	"math/rand"
	"testing"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/stretchr/testify/assert"
)

func MessageReadWriteTest[M Message](writeMessage Message, readMessage Message, t *testing.T) {

	t.Parallel()

	const BufferSize = 10 * 1024

	buffer := make([]byte, BufferSize)

	buffer = writeMessage.Write(buffer[:])

	err := readMessage.Read(buffer)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

func GenerateRandomCostMatrixStatMessage() CostMatrixStatsMessage {

	rand.Seed(time.Now().UnixNano())

	return CostMatrixStatsMessage{
		Version:        CostMatrixStatsMessageVersion,
		Timestamp:      uint64(time.Now().Unix()),
		Bytes:          rand.Int(),
		NumRelays:      rand.Int(),
		NumDestRelays:  rand.Int(),
		NumDatacenters: rand.Int(),
	}
}

func GenerateRandomRouteMatrixStatMessage() RouteMatrixStatsMessage {

	rand.Seed(time.Now().UnixNano())

	return RouteMatrixStatsMessage{
		Version:                 RouteMatrixStatsMessageVersion,
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

func GenerateRandomPingStatMessage(routeable bool) PingStatsMessage {

	rand.Seed(time.Now().UnixNano())

	return PingStatsMessage{
		Version:    PingStatsMessageVersion,
		Timestamp:  uint64(time.Now().Unix()),
		RelayA:     rand.Uint64(),
		RelayB:     rand.Uint64(),
		RTT:        rand.Float32(),
		Jitter:     rand.Float32(),
		PacketLoss: rand.Float32(),
		Routable:   routeable,
	}
}

func GenerateRandomRelayStatMessage(full bool) RelayStatsMessage {

	rand.Seed(time.Now().UnixNano())

	return RelayStatsMessage{
		Version:                   RelayStatsMessageVersion,
		Timestamp:                 uint64(time.Now().Unix()),
		ID:                        rand.Uint64(),
		NumSessions:               rand.Uint32(),
		MaxSessions:               rand.Uint32(),
		NumRoutable:               rand.Uint32(),
		NumUnroutable:             rand.Uint32(),
		Full:                      full,
		CPUUsage:                  rand.Float32(),
		BandwidthSentPercent:      rand.Float32(),
		BandwidthReceivedPercent:  rand.Float32(),
		EnvelopeSentPercent:       rand.Float32(),
		EnvelopeReceivedPercent:   rand.Float32(),
		BandwidthSentMbps:         rand.Float32(),
		BandwidthReceivedMbps:     rand.Float32(),
		EnvelopeSentMbps:          rand.Float32(),
		EnvelopeReceivedMbps:      rand.Float32(),
		Tx:                        rand.Uint64(),
		Rx:                        rand.Uint64(),
		PeakSessions:              rand.Uint64(),
		PeakSentBandwidthMbps:     rand.Float32(),
		PeakReceivedBandwidthMbps: rand.Float32(),
		MemUsage:                  rand.Float32(),
	}
}

func GenerateRandomMatchDataMessage() MatchDataMessage {

	rand.Seed(time.Now().UnixNano())

	matchValues := make([]float64, MatchDataMaxMatchValues)

	for i := 0; i < MatchDataMaxMatchValues; i++ {
		matchValues[i] = rand.Float64()
	}

	return MatchDataMessage{
		Version:        MatchDataMessageVersion,
		Timestamp:      uint32(time.Now().Unix()),
		BuyerID:        rand.Uint64(),
		ServerAddress:  backend.GenerateRandomStringSequence(MatchDataMaxAddressLength),
		DatacenterID:   rand.Uint64(),
		UserHash:       rand.Uint64(),
		SessionID:      rand.Uint64(),
		MatchID:        rand.Uint64(),
		NumMatchValues: rand.Int31(),
		MatchValues:    matchValues,
	}
}

func TestCostMatrixStatsMessage(t *testing.T) {

	writeMessage := GenerateRandomCostMatrixStatMessage()

	readMessage := CostMatrixStatsMessage{}

	MessageReadWriteTest[*CostMatrixStatsMessage](&writeMessage, &readMessage, t)
}

func TestRouteMatrixStatsMessage(t *testing.T) {

	writeMessage := GenerateRandomRouteMatrixStatMessage()

	readMessage := RouteMatrixStatsMessage{}

	MessageReadWriteTest[*RouteMatrixStatsMessage](&writeMessage, &readMessage, t)
}

func TestPingStatsMessage(t *testing.T) {

	writeMessage := GenerateRandomPingStatMessage(true)

	readMessage := PingStatsMessage{}

	MessageReadWriteTest[*PingStatsMessage](&writeMessage, &readMessage, t)
}

func TestRelayStatsMessage(t *testing.T) {

	writeMessage := GenerateRandomRelayStatMessage(true)

	readMessage := RelayStatsMessage{}

	MessageReadWriteTest[*RelayStatsMessage](&writeMessage, &readMessage, t)
}

func TestMatchDataMessage(t *testing.T) {

	writeMessage := GenerateRandomMatchDataMessage()

	readMessage := MatchDataMessage{}

	MessageReadWriteTest[*MatchDataMessage](&writeMessage, &readMessage, t)
}

func TestBillingMessage(t *testing.T) {

	writeMessage := BillingMessage{}

	readMessage := BillingMessage{}

	MessageReadWriteTest[*BillingMessage](&writeMessage, &readMessage, t)
}

func TestSummaryMessage(t *testing.T) {

	writeMessage := SummaryMessage{}

	readMessage := SummaryMessage{}

	MessageReadWriteTest[*SummaryMessage](&writeMessage, &readMessage, t)
}
