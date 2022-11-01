package common_test

import (
	"testing"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/encoding"

	"github.com/networknext/backend/modules-old/routing"

	"github.com/stretchr/testify/assert"
)

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
		writeMessage := common.GenerateRandomCostMatrix()
		readMessage := common.CostMatrix{}
		CostMatrixReadWriteTest(&writeMessage, &readMessage, t)
	}
}

// todo: remove this once we remove the old cost matrix
func CostMatrixReadWriteTest_NewToOld(writeMessage *common.CostMatrix, readMessage *common.CostMatrix, t *testing.T) {

	const BufferSize = 100 * 1024

	buffer, err := writeMessage.Write(BufferSize)
	assert.Nil(t, err)

	var oldCostMatrix routing.CostMatrix
	readStream := encoding.CreateReadStream(buffer)
	err = oldCostMatrix.Serialize(readStream)
	assert.Nil(t, err)

	buffer2 := make([]byte, BufferSize)
	writeStream := encoding.CreateWriteStream(buffer2)
	err = oldCostMatrix.Serialize(writeStream)
	assert.Nil(t, err)

	err = readMessage.Read(buffer2)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

func TestCostMatrixReadWrite_NewToOld(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumCostMatrixIterations; i++ {
		writeMessage := common.GenerateRandomCostMatrix()
		readMessage := common.CostMatrix{}
		CostMatrixReadWriteTest(&writeMessage, &readMessage, t)
	}
}
