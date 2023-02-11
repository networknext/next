package common_test

import (
	"testing"

	"github.com/networknext/backend/modules/common"

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

const NumCostMatrixIterations = 10000

func TestCostMatrixReadWrite(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumCostMatrixIterations; i++ {
		writeMessage := common.GenerateRandomCostMatrix()
		readMessage := common.CostMatrix{}
		CostMatrixReadWriteTest(&writeMessage, &readMessage, t)
	}
}
