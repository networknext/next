package handlers_test

import (
	"net"
	"time"
	"testing"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/packets"
	"github.com/networknext/backend/modules/handlers"
	"github.com/networknext/backend/modules/database"

	"github.com/stretchr/testify/assert"
)

func DummyLocateIP(ip net.IP) (packets.SDK5_LocationData, error) {
	location := packets.SDK5_LocationData{}
	location.Latitude = 43
	location.Longitude = -75
	return location, nil
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

	// todo: stick in a locate IP func that fails

	result := handlers.SessionPre(state)

	assert.False(t, result)
	assert.False(t, result)
}

func TestSessionPre_ReadLocation(t *testing.T) {

	t.Parallel()

	state := CreateState()

	// todo: make slice non-zero

	// todo: stick in session data that has previous location stored in it

	result := handlers.SessionPre(state)

	assert.False(t, result)

	// todo: ReadSessionData = true
	// todo: check output location lat/long
	// todo: check input location == output location
}

func TestSessionPre_StaleRouteMatrix(t *testing.T) {

	t.Parallel()

	state := CreateState()

	state.RouteMatrix.CreatedAt = 0

	result := handlers.SessionPre(state)

	assert.True(t, result)
	assert.True(t, state.StaleRouteMatrix)
}

// UnknownDatacenter
// DatacenterNotEnabled
// getDatacenter
// NoRelaysInDatacenter
// (relays are in datacenter)
// pro tag
// optout tag
// debug string
