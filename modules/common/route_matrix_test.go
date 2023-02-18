package common_test

import (
	"net"
	"time"
	"testing"

	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"

	"github.com/stretchr/testify/assert"
)

func RouteMatrixReadWriteTest(writeMessage *common.RouteMatrix, readMessage *common.RouteMatrix, t *testing.T) {

	buffer, err := writeMessage.Write()
	assert.Nil(t, err)

	err = readMessage.Read(buffer)
	assert.Nil(t, err)

	assert.Equal(t, writeMessage, readMessage)
}

const NumRouteMatrixIterations = 10

func TestRouteMatrixReadWrite(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumRouteMatrixIterations; i++ {
		writeMessage := common.GenerateRandomRouteMatrix()
		readMessage := common.RouteMatrix{}
		RouteMatrixReadWriteTest(&writeMessage, &readMessage, t)
	}
}

// todo: func test all relay numbers [0,constants.MaxRelays] too slow for here

func TestRouteMatrixNoRelays(t *testing.T) {

	t.Parallel()

	routeMatrix := common.RouteMatrix{}

	routeMatrix.Version = common.RouteMatrixVersion_Write
	routeMatrix.CreatedAt = uint64(time.Now().Unix())
	routeMatrix.BinFileBytes = common.MaxDatabaseBinWrapperSize
	routeMatrix.BinFileData = make([]byte, routeMatrix.BinFileBytes)
	routeMatrix.RelayIds = make([]uint64, 0)
	routeMatrix.RelayIdToIndex = make(map[uint64]int32)
	routeMatrix.RelayAddresses = make([]net.UDPAddr, 0)
	routeMatrix.RelayNames = make([]string, 0)
	routeMatrix.RelayLatitudes = make([]float32, 0)
	routeMatrix.RelayLongitudes = make([]float32, 0)
	routeMatrix.RelayDatacenterIds = make([]uint64, 0)
	routeMatrix.DestRelays = make([]bool, 0)
	routeMatrix.RouteEntries = make([]core.RouteEntry, 0)

	readMessage := common.RouteMatrix{}

	RouteMatrixReadWriteTest(&routeMatrix, &readMessage, t)
}

func TestRouteMatrixMaxSize(t *testing.T) {

	t.Parallel()

	routeMatrix := common.RouteMatrix{}

	routeMatrix.Version = common.RouteMatrixVersion_Write
	routeMatrix.CreatedAt = uint64(time.Now().Unix())
	routeMatrix.BinFileBytes = common.MaxDatabaseBinWrapperSize
	routeMatrix.BinFileData = make([]byte, routeMatrix.BinFileBytes)
	routeMatrix.RelayIds = make([]uint64, constants.MaxRelays)
	routeMatrix.RelayIdToIndex = make(map[uint64]int32)
	routeMatrix.RelayAddresses = make([]net.UDPAddr, constants.MaxRelays)
	routeMatrix.RelayNames = make([]string, constants.MaxRelays)
	routeMatrix.RelayLatitudes = make([]float32, constants.MaxRelays)
	routeMatrix.RelayLongitudes = make([]float32, constants.MaxRelays)
	routeMatrix.RelayDatacenterIds = make([]uint64, constants.MaxRelays)
	routeMatrix.DestRelays = make([]bool, constants.MaxRelays)
	routeMatrix.RouteEntries = make([]core.RouteEntry, core.TriMatrixLength(constants.MaxRelays))

	for i := 0; i < constants.MaxRelays; i++ {
		routeMatrix.RelayIds[i] = uint64(i)
		routeMatrix.RelayIdToIndex[uint64(i)] = int32(i)
		routeMatrix.RelayAddresses[i] = common.RandomAddress()
		routeMatrix.RelayNames[i] = common.RandomString(constants.MaxRelayNameLength)
	}

	for i := range routeMatrix.RouteEntries {
		routeMatrix.RouteEntries[i].DirectCost = constants.MaxRouteCost
		routeMatrix.RouteEntries[i].NumRoutes = constants.MaxRoutesPerEntry
		for j := 0; j < int(routeMatrix.RouteEntries[i].NumRoutes); j++ {
			routeMatrix.RouteEntries[i].RouteCost[j] = int32(common.RandomInt(0, constants.MaxRouteCost))
			routeMatrix.RouteEntries[i].RouteNumRelays[j] = constants.MaxRouteRelays
			for k := 0; k < int(routeMatrix.RouteEntries[i].RouteNumRelays[j]); k++ {
				routeMatrix.RouteEntries[i].RouteRelays[j][k] = int32(k)
			}
			routeMatrix.RouteEntries[i].RouteHash[j] = core.RouteHash(routeMatrix.RouteEntries[i].RouteRelays[j][:]...)
		}
	}

	readMessage := common.RouteMatrix{}

	RouteMatrixReadWriteTest(&routeMatrix, &readMessage, t)
}
