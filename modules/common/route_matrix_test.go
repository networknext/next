package common_test

import (
	"testing"

	"github.com/networknext/accelerate/modules/common"

	"github.com/stretchr/testify/assert"
)

func RouteMatrixReadWriteTest(writeMessage *common.RouteMatrix, readMessage *common.RouteMatrix, t *testing.T) {

	buffer, err := writeMessage.Write()
	assert.Nil(t, err)

	err = readMessage.Read(buffer)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

func TestRouteMatrixReadWrite(t *testing.T) {
	t.Parallel()
	writeMessage := common.GenerateRandomRouteMatrix(32)
	readMessage := common.RouteMatrix{}
	RouteMatrixReadWriteTest(&writeMessage, &readMessage, t)
}
