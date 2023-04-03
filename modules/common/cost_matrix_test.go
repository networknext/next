package common_test

import (
	"testing"

	"github.com/networknext/accelerate/modules/common"

	"github.com/stretchr/testify/assert"
)

func CostMatrixReadWriteTest(writeMessage *common.CostMatrix, readMessage *common.CostMatrix, t *testing.T) {

	buffer, err := writeMessage.Write()
	assert.Nil(t, err)

	err = readMessage.Read(buffer)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

func TestCostMatrixReadWrite(t *testing.T) {
	t.Parallel()
	writeMessage := common.GenerateRandomCostMatrix(32)
	readMessage := common.CostMatrix{}
	CostMatrixReadWriteTest(&writeMessage, &readMessage, t)
}
