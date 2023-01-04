package handlers_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/database"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/handlers"
	"github.com/networknext/backend/modules/packets"

	"github.com/stretchr/testify/assert"
)

func DummyLocateIP(ip net.IP) (packets.SDK5_LocationData, error) {
	location := packets.SDK5_LocationData{}
	location.Latitude = 43
	location.Longitude = -75
	return location, nil
}

func FailLocateIP(ip net.IP) (packets.SDK5_LocationData, error) {
	location := packets.SDK5_LocationData{}
	return location, fmt.Errorf("fail")
}

func CreateState() *handlers.SessionUpdateState {
	state := handlers.SessionUpdateState{}
	state.Request = &packets.SDK5_SessionUpdateRequestPacket{}
	state.LocateIP = DummyLocateIP
	state.RouteMatrix = &common.RouteMatrix{}
	state.RouteMatrix.CreatedAt = uint64(time.Now().Unix())
	state.Database = database.CreateDatabase()
	return &state
}

func TestSessionPre_AnalysisOnly(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.RouteShader.AnalysisOnly = true

	result := handlers.SessionPre(state)

	assert.False(t, result)
	assert.True(t, state.AnalysisOnly)
}

func TestSessionPre_ClientPingTimedOut(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.ClientPingTimedOut = true

	result := handlers.SessionPre(state)

	assert.True(t, result)
	assert.True(t, state.ClientPingTimedOut)
}

func TestSessionPre_LocatedIP(t *testing.T) {

	t.Parallel()

	state := CreateState()

	result := handlers.SessionPre(state)

	assert.False(t, result)
	assert.True(t, state.LocatedIP)
	assert.False(t, state.LocationVeto)
	assert.Equal(t, state.Output.Location.Latitude, float32(43))
	assert.Equal(t, state.Output.Location.Longitude, float32(-75))
}

func TestSessionPre_LocationVeto(t *testing.T) {

	t.Parallel()

	state := CreateState()
	state.LocateIP = FailLocateIP

	result := handlers.SessionPre(state)

	assert.True(t, result)
	assert.True(t, state.LocationVeto)
}

func TestSessionPre_ReadLocation(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.SliceNumber = 1

	sessionData := packets.SDK5_SessionData{}

	sessionData.Version = packets.SDK5_SessionDataVersion_Write
	sessionData.Location.Version = packets.SDK5_LocationVersion_Write
	sessionData.Location.Latitude = 10
	sessionData.Location.Longitude = 20
	sessionData.Location.ISP = "Starlink"
	sessionData.Location.ASN = 5

	buffer := make([]byte, packets.SDK5_MaxPacketBytes)
	writeStream := encoding.CreateWriteStream(buffer[:])
	err := sessionData.Serialize(writeStream)
	assert.Nil(t, err)
	writeStream.Flush()
	packetBytes := writeStream.GetBytesProcessed()
	packetData := buffer[:packetBytes]

	state.Request.SessionDataBytes = int32(packetBytes)
	copy(state.Request.SessionData[:], packetData)

	result := handlers.SessionPre(state)

	assert.False(t, result)
	assert.True(t, state.ReadSessionData)
	assert.Equal(t, float32(10), state.Output.Location.Latitude)
	assert.Equal(t, float32(20), state.Output.Location.Longitude)
	assert.Equal(t, "Starlink", state.Output.Location.ISP)
	assert.Equal(t, uint32(5), state.Output.Location.ASN)
}

func TestSessionPre_StaleRouteMatrix(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.RouteMatrix.CreatedAt = 0

	result := handlers.SessionPre(state)

	assert.True(t, result)
	assert.True(t, state.StaleRouteMatrix)
}

func TestSessionPre_KnownDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.DatacenterId = 0x12345

	state.Database.DatacenterMap[0x12345] = database.Datacenter{}

	result := handlers.SessionPre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
}

func TestSessionPre_UnknownDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.DatacenterId = 0x12345

	result := handlers.SessionPre(state)

	assert.False(t, result)
	assert.True(t, state.UnknownDatacenter)
}

func TestSessionPre_DatacenterNotEnabled(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.ID = 0x11111
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = database.Datacenter{}

	result := handlers.SessionPre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
	assert.True(t, state.DatacenterNotEnabled)
}

func TestSessionPre_DatacenterEnabled(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.ID = 0x11111
	state.Request.BuyerId = 0x11111
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = database.Datacenter{}
	state.Database.DatacenterMaps[state.Buyer.ID] = make(map[uint64]database.DatacenterMap)
	state.Database.DatacenterMaps[state.Buyer.ID][state.Request.DatacenterId] = database.DatacenterMap{}

	result := handlers.SessionPre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
	assert.False(t, state.DatacenterNotEnabled)
}

func TestSessionPre_FailedToReadSessionData(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.SliceNumber = 1

	result := handlers.SessionPre(state)

	assert.True(t, result)
	assert.True(t, state.FailedToReadSessionData)
}

func TestSessionPre_NoRelaysInDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.ID = 0x11111
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = database.Datacenter{}

	result := handlers.SessionPre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
	assert.True(t, state.NoRelaysInDatacenter)
}

func TestSessionPre_RelaysInDatacenter(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.ID = 0x11111
	state.Request.DatacenterId = 0x12345
	state.Database.DatacenterMap[0x12345] = database.Datacenter{}

	const NumRelays = 10

	state.RouteMatrix.RelayIds = make([]uint64, NumRelays)
	state.RouteMatrix.RelayDatacenterIds = make([]uint64, NumRelays)

	for i := 0; i < NumRelays; i++ {
		state.RouteMatrix.RelayIds[i] = uint64(i)
		state.RouteMatrix.RelayDatacenterIds[i] = 0x12345
	}

	result := handlers.SessionPre(state)

	assert.False(t, result)
	assert.False(t, state.UnknownDatacenter)
	assert.False(t, state.NoRelaysInDatacenter)
}

// getDatacenter

func TestSessionPre_Pro(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.NumTags = 1
	state.Request.Tags[0] = common.HashTag("pro")

	result := handlers.SessionPre(state)

	assert.False(t, result)
	assert.True(t, state.Pro)
	assert.False(t, state.OptOut)
}

func TestSessionPre_OptOut(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Request.NumTags = 1
	state.Request.Tags[0] = common.HashTag("optout")

	result := handlers.SessionPre(state)

	assert.True(t, result)
	assert.True(t, state.OptOut)
	assert.False(t, state.Pro)
}

func TestSessionPre_Debug(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.Buyer.Debug = true

	result := handlers.SessionPre(state)

	assert.False(t, result)
	assert.NotNil(t, state.Debug)
}
