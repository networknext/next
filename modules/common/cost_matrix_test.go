package common_test

import (
	"testing"

	"github.com/networknext/backend/modules/common"

	"github.com/stretchr/testify/assert"
)

func GenerateRandomCostMatrix() common.CostMatrix {

	return common.CostMatrix{
		// todo
	}
}

func CostMatrixReadWriteTest(writeMessage *common.CostMatrix, readMessage *common.CostMatrix, t *testing.T) {

	const BufferSize = 10 * 1024

	buffer := make([]byte, BufferSize)

	buffer = writeMessage.Write(buffer[:])

	err := readMessage.Read(buffer)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

func TestCostMatrixReadWrite(t *testing.T) {

	t.Parallel()

	writeMessage := GenerateRandomCostMatrix()

	readMessage := common.CostMatrix{}

	CostMatrixReadWriteTest(&writeMessage, &readMessage, t)
}
