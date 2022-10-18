package common_test

import (
	"testing"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/encoding"

	"github.com/networknext/backend/modules-old/routing"

	"github.com/stretchr/testify/assert"
)

func RouteMatrixReadWriteTest(writeMessage *common.RouteMatrix, readMessage *common.RouteMatrix, t *testing.T) {

	const BufferSize = 1024 * 1024

	buffer, err := writeMessage.Write(BufferSize)
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

// todo: remove this once we remove the old route matrix

func RouteMatrixReadWriteTest_NewToOld(writeMessage *common.RouteMatrix, readMessage *common.RouteMatrix, t *testing.T) {

	const BufferSize = 100 * 1024

	buffer, err := writeMessage.Write(BufferSize)
	assert.Nil(t, err)

	var oldRouteMatrix routing.RouteMatrix
	readStream := encoding.CreateReadStream(buffer)
	err = oldRouteMatrix.Serialize(readStream)
	assert.Nil(t, err)

	buffer2 := make([]byte, BufferSize)
	writeStream := encoding.CreateWriteStream(buffer2)
	err = oldRouteMatrix.Serialize(writeStream)
	assert.Nil(t, err)

	err = readMessage.Read(buffer2)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

func TestRouteMatrixReadWrite_NewToOld(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumRouteMatrixIterations; i++ {
		writeMessage := common.GenerateRandomRouteMatrix()
		readMessage := common.RouteMatrix{}
		RouteMatrixReadWriteTest(&writeMessage, &readMessage, t)
	}
}
