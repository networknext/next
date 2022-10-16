package common_test

import (
	"net"
	"testing"
	"math/rand"

	"github.com/networknext/backend/modules/common"

	"github.com/stretchr/testify/assert"
)

func GenerateRandomCostMatrix() common.CostMatrix {

	costMatrix := common.CostMatrix{
		Version: uint32(common.RandomInt(common.CostMatrixVersion_Min, common.CostMatrixVersion_Max)),
	}

	numRelays := common.RandomInt(0, 64)

	costMatrix.RelayIds = make([]uint64, numRelays)
	costMatrix.RelayAddresses = make([]net.UDPAddr, numRelays)
	costMatrix.RelayNames = make([]string, numRelays)
	costMatrix.RelayLatitudes = make([]float32, numRelays)
	costMatrix.RelayLongitudes = make([]float32, numRelays)
	costMatrix.RelayDatacenterIds = make([]uint64, numRelays)
	costMatrix.DestRelays = make([]bool, numRelays)
	costMatrix.Costs = make([]int32, numRelays*numRelays)

	for i := 0; i < numRelays; i++ {
		costMatrix.RelayIds[i] = rand.Uint64()
		costMatrix.RelayAddresses[i] = common.RandomAddress()
		costMatrix.RelayNames[i] = common.RandomString(common.MaxRelayNameLength)
		costMatrix.RelayLatitudes[i] = rand.Float32()
		costMatrix.RelayLongitudes[i] = rand.Float32()
		costMatrix.RelayDatacenterIds[i] = rand.Uint64()
		costMatrix.DestRelays[i] = common.RandomBool()
	}

	for i := 0; i < numRelays*numRelays; i++ {
		costMatrix.Costs[i] = int32(common.RandomInt(-1, 1000))
	}

	return costMatrix
}

func CostMatrixReadWriteTest(writeMessage *common.CostMatrix, readMessage *common.CostMatrix, t *testing.T) {

	const BufferSize = 100 * 1024

	buffer, err := writeMessage.Write(BufferSize)
	assert.Nil(t, err)

	err = readMessage.Read(buffer)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

const NumCostMatrixIterations = 1000

func TestCostMatrixReadWrite(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumCostMatrixIterations; i++ {
		writeMessage := GenerateRandomCostMatrix()
		readMessage := common.CostMatrix{}
		CostMatrixReadWriteTest(&writeMessage, &readMessage, t)
	}
}
