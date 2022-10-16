package common_test

import (
	"testing"

	"github.com/networknext/backend/modules/common"

	"github.com/stretchr/testify/assert"
)

func GenerateRandomCostMatrix() common.CostMatrix {

	// todo: we actually need to allocate and fill the cost matrix arrays before this will work

	return common.CostMatrix{
		Version: uint32(common.RandomInt(common.CostMatrixVersion_Min, common.CostMatrixVersion_Max)),
		// todo
	}
}

func CostMatrixReadWriteTest(writeMessage *common.CostMatrix, readMessage *common.CostMatrix, t *testing.T) {

	const BufferSize = 100 * 1024

	buffer, err := writeMessage.Write(BufferSize)
	assert.Nil(t, err)

	err = readMessage.Read(buffer)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

func TestCostMatrixReadWrite(t *testing.T) {

	t.Parallel()

	// todo
	/*
	writeMessage := GenerateRandomCostMatrix()

	readMessage := common.CostMatrix{}

	CostMatrixReadWriteTest(&writeMessage, &readMessage, t)
	*/
}
