package messages

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

func TestCostMatrixStatsMessage(t *testing.T) {

	writeMessage := CostMatrixStatsMessage{
		Version:        1,
		Timestamp:      uint64(time.Now().Unix()),
		Bytes:          257,
		NumRelays:      23,
		NumDestRelays:  15,
		NumDatacenters: 7,
	}

	readMessage := CostMatrixStatsMessage{}

	MessageReadWriteTest[*CostMatrixStatsMessage](&writeMessage, &readMessage, t)
}

func TestRouteMatrixStatsMessage(t *testing.T) {

	writeMessage := RouteMatrixStatsMessage{
		Version:                 1,
		Timestamp:               uint64(time.Now().Unix()),
		Bytes:                   257,
		NumRelays:               23,
		NumDestRelays:           15,
		NumFullRelays:           5,
		NumDatacenters:          7,
		TotalRoutes:             1021412,
		AverageNumRoutes:        10.65,
		AverageRouteLength:      3.75,
		NoRoutePercent:          1.0,
		OneRoutePercent:         2.0,
		NoDirectRoutePercent:    3.0,
		RTTBucket_NoImprovement: 25.1,
		RTTBucket_0_5ms:         10.6,
		RTTBucket_5_10ms:        5.2,
		RTTBucket_10_15ms:       4.3,
		RTTBucket_15_20ms:       3.2,
		RTTBucket_20_25ms:       1.5,
		RTTBucket_25_30ms:       6.0,
		RTTBucket_30_35ms:       2.7,
		RTTBucket_35_40ms:       1.23,
		RTTBucket_40_45ms:       0.75,
		RTTBucket_45_50ms:       0.56,
		RTTBucket_50ms_Plus:     2.9,
	}

	readMessage := RouteMatrixStatsMessage{}

	MessageReadWriteTest[*RouteMatrixStatsMessage](&writeMessage, &readMessage, t)
}
