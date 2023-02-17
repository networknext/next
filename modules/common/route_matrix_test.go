package common_test

import (
	"testing"

	"github.com/networknext/backend/modules/common"

	"github.com/stretchr/testify/assert"
)

func RouteMatrixReadWriteTest(writeMessage *common.RouteMatrix, readMessage *common.RouteMatrix, t *testing.T) {

	buffer, err := writeMessage.Write()
	assert.Nil(t, err)

	err = readMessage.Read(buffer)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

const NumRouteMatrixIterations = 1

func TestRouteMatrixReadWrite(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumRouteMatrixIterations; i++ {
		writeMessage := common.GenerateRandomRouteMatrix()
		readMessage := common.RouteMatrix{}
		RouteMatrixReadWriteTest(&writeMessage, &readMessage, t)
	}
}
