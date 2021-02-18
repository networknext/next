package routing_test

import (
	"net"
	"testing"
	"time"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
	"github.com/stretchr/testify/assert"
)

func getCostMatrix() (routing.CostMatrix, error) {
	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	if err != nil {
		return routing.CostMatrix{}, err
	}
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	if err != nil {
		return routing.CostMatrix{}, err
	}

	expected := routing.CostMatrix{
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		Costs:              []int32{10},
	}

	return expected, nil
}

func TestCostMatrixSerialize(t *testing.T) {
	expected, err := getCostMatrix()
	assert.NoError(t, err)

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

func BenchmarkCostMatrixSerialize(b *testing.B) {
	expected, err := getCostMatrix()
	assert.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws, err := encoding.CreateWriteStream(10000)
		assert.NoError(b, err)
		err = expected.Serialize(ws)
		assert.NoError(b, err)

		ws.Flush()
		data := ws.GetData()[:ws.GetBytesProcessed()]

		var actual routing.CostMatrix
		rs := encoding.CreateReadStream(data)
		err = actual.Serialize(rs)
		assert.NoError(b, err)

		assert.Equal(b, expected, actual)
	}
}

func TestCostMatrixEncodeDecodeBinary(t *testing.T) {
	expected, err := getCostMatrix()
	assert.NoError(t, err)
	expected.Version = routing.CostMatrixV1
	expected.CreatedAt = uint64(time.Now().Unix())
	data := make([]byte, 10000)
	err = expected.WriteToBinary(data)
	assert.NoError(t, err)

	var actual routing.CostMatrix
	err = actual.ReadFromBinary(data)
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func BenchmarkCostMatrixEncodeDecodeBinary(b *testing.B) {
	expected, err := getCostMatrix()
	assert.NoError(b, err)
	expected.Version = routing.CostMatrixV1
	expected.CreatedAt = uint64(time.Now().Unix())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := make([]byte, 10000)
		err = expected.WriteToBinary(data)
		assert.NoError(b, err)

		var actual routing.CostMatrix
		err = actual.ReadFromBinary(data)
		assert.NoError(b, err)

		assert.Equal(b, expected, actual)
	}
}
