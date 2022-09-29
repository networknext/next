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

	return messages.SessionUpdateMessage{
		Version:     messages.SessionUpdateMessageVersion,
		Timestamp:   rand.Uint64(),
		SessionId:   rand.Uint64(),
		SliceNumber: rand.Uint32(),
		// ...
	}
}

/*
type SessionUpdateMessage struct {

	// always

	Timestamp           uint32
	SessionID           uint64
	SliceNumber         uint32
	DirectMinRTT        int32
	DirectMaxRTT        int32
	DirectPrimeRTT      int32
	DirectJitter        int32
	DirectPacketLoss    int32
	RealPacketLoss      int32
	RealPacketLoss_Frac uint32
	RealJitter          uint32
	Next                bool
	Flagged             bool
	Summary             bool
	UseDebug            bool
	Debug               string
	RouteDiversity      int32
	UserFlags           uint64
	TryBeforeYouBuy     bool

	// first slice and summary slice only

	DatacenterID      uint64
	BuyerID           uint64
	UserHash          uint64
	EnvelopeBytesUp   uint64
	EnvelopeBytesDown uint64
	Latitude          float32
	Longitude         float32
	ClientAddress     string
	ServerAddress     string
	ISP               string
	ConnectionType    int32
	PlatformType      int32
	SDKVersion        string
	NumTags           int32
	Tags              [SessionUpdateMessageMaxTags]uint64
	ABTest            bool
	Pro               bool

	// summary slice only

	ClientToServerPacketsSent       uint64
	ServerToClientPacketsSent       uint64
	ClientToServerPacketsLost       uint64
	ServerToClientPacketsLost       uint64
	ClientToServerPacketsOutOfOrder uint64
	ServerToClientPacketsOutOfOrder uint64
	NumNearRelays                   int32
	NearRelayIDs                    [SessionUpdateMessageMaxNearRelays]uint64
	NearRelayRTTs                   [SessionUpdateMessageMaxNearRelays]int32
	NearRelayJitters                [SessionUpdateMessageMaxNearRelays]int32
	NearRelayPacketLosses           [SessionUpdateMessageMaxNearRelays]int32
	EverOnNext                      bool
	SessionDuration                 uint32
	TotalPriceSum                   uint64
	EnvelopeBytesUpSum              uint64
	EnvelopeBytesDownSum            uint64
	DurationOnNext                  uint32
	StartTimestamp                  uint32

	// network next only

	NextRTT             int32
	NextJitter          int32
	NextPacketLoss      int32
	PredictedNextRTT    int32
	NearRelayRTT        int32
	NumNextRelays       int32
	NextRelays          [SessionUpdateMessageMaxRelays]uint64
	NextRelayPrice      [SessionUpdateMessageMaxRelays]uint64
	TotalPrice          uint64
	Uncommitted         bool
	Multipath           bool
	RTTReduction        bool
	PacketLossReduction bool
	RouteChanged        bool
	NextBytesUp         uint64
	NextBytesDown       uint64

	// error state only

	FallbackToDirect     bool
	MultipathVetoed      bool
	Mispredicted         bool
	Vetoed               bool
	LatencyWorse         bool
	NoRoute              bool
	NextLatencyTooHigh   bool
	CommitVeto           bool
	UnknownDatacenter    bool
	DatacenterNotEnabled bool
	BuyerNotLive         bool
	StaleRouteMatrix     bool
}
*/

// -----------------------------------------------------------

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
	// for i := 0; i < NumIterations; i++ {
	writeMessage := GenerateRandomMatchDataMessage()
	readMessage := messages.MatchDataMessage{}
	MessageReadWriteTest[*messages.MatchDataMessage](&writeMessage, &readMessage, t)
	// }
}

// todo: test the session update message
