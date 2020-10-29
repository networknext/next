package routing_test

import (
	"net"
	"testing"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func getCostMatrix(t *testing.T) routing.CostMatrix {
	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	expected := routing.CostMatrix{
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		Costs:              []int32{10},
	}

	return expected
}

func TestCostMatrixSerialize(t *testing.T) {
	expected := getCostMatrix(t)

	ws, err := encoding.CreateWriteStream(10000)
	assert.NoError(t, err)
	err = expected.Serialize(ws)
	assert.NoError(t, err)

	ws.Flush()
	data := ws.GetData()[:ws.GetBytesProcessed()]

	var actual routing.CostMatrix
	rs := encoding.CreateReadStream(data)
	err = actual.Serialize(rs)
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}
