package common_test

import (
	"math/rand"
	"net"
	"testing"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"

	"github.com/stretchr/testify/assert"
)

func GenerateRandomRouteMatrix() common.RouteMatrix {

	routeMatrix := common.RouteMatrix{
		Version: uint32(common.RandomInt(common.RouteMatrixVersion_Min, common.RouteMatrixVersion_Max)),
	}

	numRelays := common.RandomInt(0, 64)

	routeMatrix.RelayIds = make([]uint64, numRelays)
	routeMatrix.RelayAddresses = make([]net.UDPAddr, numRelays)
	routeMatrix.RelayNames = make([]string, numRelays)
	routeMatrix.RelayLatitudes = make([]float32, numRelays)
	routeMatrix.RelayLongitudes = make([]float32, numRelays)
	routeMatrix.RelayDatacenterIds = make([]uint64, numRelays)
	routeMatrix.DestRelays = make([]bool, numRelays)

	for i := 0; i < numRelays; i++ {
		routeMatrix.RelayIds[i] = rand.Uint64()
		routeMatrix.RelayAddresses[i] = common.RandomAddress()
		routeMatrix.RelayNames[i] = common.RandomString(common.MaxRelayNameLength)
		routeMatrix.RelayLatitudes[i] = rand.Float32()
		routeMatrix.RelayLongitudes[i] = rand.Float32()
		routeMatrix.RelayDatacenterIds[i] = rand.Uint64()
		routeMatrix.DestRelays[i] = common.RandomBool()
	}

	routeMatrix.RelayIdToIndex = make(map[uint64]int32)
	for i := range routeMatrix.RelayIds {
		routeMatrix.RelayIdToIndex[routeMatrix.RelayIds[i]] = int32(i)
	}

	routeMatrix.BinFileBytes = int32(common.RandomInt(100, 10000))
	routeMatrix.BinFileData = make([]byte, routeMatrix.BinFileBytes)
	common.RandomBytes(routeMatrix.BinFileData)

	routeMatrix.CreatedAt = rand.Uint64()
	routeMatrix.Version = uint32(common.RandomInt(common.RouteMatrixVersion_Min, common.RouteMatrixVersion_Max))

	numFullRelays := common.RandomInt(0, numRelays)

	routeMatrix.FullRelayIds = make([]uint64, numFullRelays)
	routeMatrix.FullRelayIndexSet = make(map[int32]bool)
	for i := range routeMatrix.FullRelayIds {
		routeMatrix.FullRelayIds[i] = routeMatrix.RelayIds[i]
		routeMatrix.FullRelayIndexSet[int32(i)] = true
	}

	numEntries := common.RandomInt(0, 1000)
	routeMatrix.RouteEntries = make([]core.RouteEntry, numEntries)
	for i := range routeMatrix.RouteEntries {
		// todo: random route entry
		_ = i
	}

	return routeMatrix
}

func RouteMatrixReadWriteTest(writeMessage *common.RouteMatrix, readMessage *common.RouteMatrix, t *testing.T) {

	const BufferSize = 100 * 1024

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
		writeMessage := GenerateRandomRouteMatrix()
		readMessage := common.RouteMatrix{}
		RouteMatrixReadWriteTest(&writeMessage, &readMessage, t)
	}
}
