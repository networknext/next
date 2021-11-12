package transport_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func testSessionCountData() transport.SessionCountData {
	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	data := transport.SessionCountData{
		Version:     transport.SessionCountDataVersion,
		ServerID:    rand.Uint64(),
		BuyerID:     rand.Uint64(),
		NumSessions: rand.Uint32(),
	}

	return data
}

func TestSessionCountData_Serialize(t *testing.T) {
	t.Parallel()

	t.Run("test serialize write", func(t *testing.T) {
		countData := testSessionCountData()

		data, err := transport.WriteSessionCountData(&countData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("test serialize read", func(t *testing.T) {
		countData := testSessionCountData()

		data, err := transport.WriteSessionCountData(&countData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var readCountData transport.SessionCountData

		err = transport.ReadSessionCountData(&readCountData, data)

		assert.NoError(t, err)
		assert.Equal(t, countData, readCountData)
	})
}

func testRelayHop() transport.RelayHop {
	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	data := transport.RelayHop{
		Version: transport.RelayHopVersion,
		ID:      rand.Uint64(),
		Name:    backend.GenerateRandomStringSequence(rand.Intn(routing.MaxRelayNameLength)),
	}

	return data
}

func TestRelayHop_Serialize(t *testing.T) {
	t.Parallel()

	t.Run("test serialize write", func(t *testing.T) {
		hopData := testRelayHop()

		data, err := transport.WriteRelayHop(&hopData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("test serialize read", func(t *testing.T) {
		hopData := testRelayHop()

		data, err := transport.WriteRelayHop(&hopData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var readhopData transport.RelayHop

		err = transport.ReadRelayHop(&readhopData, data)

		assert.NoError(t, err)
		assert.Equal(t, hopData, readhopData)
	})
}

func testRoutingStats() routing.Stats {
	data := routing.Stats{
		RTT:        rand.Float64(),
		Jitter:     rand.Float64(),
		PacketLoss: rand.Float64(),
	}

	return data
}

func testNearRelayPortalData() transport.NearRelayPortalData {
	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	data := transport.NearRelayPortalData{
		Version:     transport.NearRelayPortalDataVersion,
		ID:          rand.Uint64(),
		Name:        backend.GenerateRandomStringSequence(rand.Intn(routing.MaxRelayNameLength)),
		ClientStats: testRoutingStats(),
	}

	return data
}

func TestNearRelayPortalData_Serialize(t *testing.T) {
	t.Parallel()

	t.Run("test serialize write", func(t *testing.T) {
		nearPortalData := testNearRelayPortalData()

		data, err := transport.WriteNearRelayPortalData(&nearPortalData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("test serialize read", func(t *testing.T) {
		nearPortalData := testNearRelayPortalData()

		data, err := transport.WriteNearRelayPortalData(&nearPortalData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var readNearPortalData transport.NearRelayPortalData

		err = transport.ReadNearRelayPortalData(&readNearPortalData, data)

		assert.NoError(t, err)
		assert.Equal(t, nearPortalData, readNearPortalData)
	})
}

func testLocation() routing.Location {
	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	return routing.Location{
		Continent:   backend.GenerateRandomStringSequence(rand.Intn(routing.MaxContinentLength - 1)),
		Country:     backend.GenerateRandomStringSequence(rand.Intn(routing.MaxCountryLength - 1)),
		CountryCode: backend.GenerateRandomStringSequence(rand.Intn(routing.MaxCountryCodeLength - 1)),
		Region:      backend.GenerateRandomStringSequence(rand.Intn(routing.MaxRegionLength - 1)),
		City:        backend.GenerateRandomStringSequence(rand.Intn(routing.MaxCityLength - 1)),
		Latitude:    rand.Float32(),
		Longitude:   rand.Float32(),
		ISP:         backend.GenerateRandomStringSequence(rand.Intn(routing.MaxISPNameLength - 1)),
		ASN:         rand.Int(),
	}
}

func testSessionMeta() transport.SessionMeta {
	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	hops := make([]transport.RelayHop, rand.Intn(6))
	for i := 0; i < len(hops); i++ {
		hops[i] = testRelayHop()
	}

	nearRelays := make([]transport.NearRelayPortalData, rand.Intn(33))
	for i := 0; i < len(nearRelays); i++ {
		nearRelays[i] = testNearRelayPortalData()
	}

	data := transport.SessionMeta{
		Version:         transport.SessionMetaVersion,
		ID:              rand.Uint64(),
		UserHash:        rand.Uint64(),
		DatacenterName:  backend.GenerateRandomStringSequence(rand.Intn(transport.MaxDatacenterNameLength)),
		DatacenterAlias: backend.GenerateRandomStringSequence(rand.Intn(transport.MaxDatacenterNameLength)),
		OnNetworkNext:   true,
		NextRTT:         rand.Float64(),
		DirectRTT:       rand.Float64(),
		DeltaRTT:        rand.Float64(),
		Location:        testLocation(),
		ClientAddr:      backend.GenerateRandomStringSequence(rand.Intn(transport.MaxAddressLength - 1)),
		ServerAddr:      backend.GenerateRandomStringSequence(rand.Intn(transport.MaxAddressLength - 1)),
		Hops:            hops,
		SDK:             backend.GenerateRandomStringSequence(rand.Intn(transport.MaxSDKVersionLength - 1)),
		Connection:      uint8(rand.Intn(256)),
		NearbyRelays:    nearRelays,
		Platform:        uint8(rand.Intn(256)),
		BuyerID:         rand.Uint64(),
	}

	// Zero out unused Location fields
	data.Location.Continent = ""
	data.Location.Country = ""
	data.Location.CountryCode = ""
	data.Location.Region = ""
	data.Location.City = ""

	return data
}

func TestSessionMeta_Serialize(t *testing.T) {
	t.Parallel()

	t.Run("test serialize write v0 and v1", func(t *testing.T) {
		metaData := testSessionMeta()
		metaData.Version = uint32(1)
		for i := 0; i < len(metaData.Hops); i++ {
			metaData.Hops[i].Version = uint32(0)
		}
		for i := 0; i < len(metaData.NearbyRelays); i++ {
			metaData.NearbyRelays[i].Version = uint32(0)
		}

		data, err := transport.WriteSessionMeta(&metaData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("test serialize read v0 and v1", func(t *testing.T) {
		metaData := testSessionMeta()
		metaData.Version = uint32(1)
		for i := 0; i < len(metaData.Hops); i++ {
			metaData.Hops[i].Version = uint32(0)
		}
		for i := 0; i < len(metaData.NearbyRelays); i++ {
			metaData.NearbyRelays[i].Version = uint32(0)
		}

		data, err := transport.WriteSessionMeta(&metaData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var readMetaData transport.SessionMeta

		err = transport.ReadSessionMeta(&readMetaData, data)

		assert.NoError(t, err)
		assert.Equal(t, metaData, readMetaData)
	})

	t.Run("test serialize write v2", func(t *testing.T) {
		metaData := testSessionMeta()

		data, err := transport.WriteSessionMeta(&metaData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("test serialize read v2", func(t *testing.T) {
		metaData := testSessionMeta()

		data, err := transport.WriteSessionMeta(&metaData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var readMetaData transport.SessionMeta

		err = transport.ReadSessionMeta(&readMetaData, data)

		assert.NoError(t, err)
		assert.Equal(t, metaData, readMetaData)
	})
}

func testSessionSlice() transport.SessionSlice {
	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	data := transport.SessionSlice{
		Version:             transport.SessionSliceVersion,
		Timestamp:           time.Unix(0, rand.Int63()),
		Next:                testRoutingStats(),
		Direct:              testRoutingStats(),
		Predicted:           testRoutingStats(),
		ClientToServerStats: testRoutingStats(),
		ServerToClientStats: testRoutingStats(),
		RouteDiversity:      rand.Uint32(),
		Envelope: routing.Envelope{
			Up:   rand.Int63(),
			Down: rand.Int63(),
		},
		OnNetworkNext:     true,
		IsMultiPath:       true,
		IsTryBeforeYouBuy: false,
	}

	// Some fields are not serialized
	data.Predicted.Jitter = float64(0)
	data.Predicted.PacketLoss = float64(0)
	data.ClientToServerStats.RTT = float64(0)
	data.ServerToClientStats.RTT = float64(0)

	return data
}

func TestSessionSlice_Serialize(t *testing.T) {
	t.Parallel()

	t.Run("test serialize write v0", func(t *testing.T) {
		sliceData := testSessionSlice()
		sliceData.Version = uint32(0)

		data, err := transport.WriteSessionSlice(&sliceData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("test serialize read v0", func(t *testing.T) {
		sliceData := testSessionSlice()
		sliceData.Version = uint32(0)
		sliceData.Predicted = routing.Stats{}
		sliceData.ClientToServerStats = routing.Stats{}
		sliceData.ServerToClientStats = routing.Stats{}
		sliceData.RouteDiversity = uint32(0)

		data, err := transport.WriteSessionSlice(&sliceData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var readSliceData transport.SessionSlice

		err = transport.ReadSessionSlice(&readSliceData, data)

		assert.NoError(t, err)
		assert.Equal(t, sliceData, readSliceData)
	})

	t.Run("test serialize write v1", func(t *testing.T) {
		sliceData := testSessionSlice()
		sliceData.Version = uint32(1)

		data, err := transport.WriteSessionSlice(&sliceData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("test serialize read v1", func(t *testing.T) {
		sliceData := testSessionSlice()
		sliceData.Version = uint32(1)
		sliceData.ClientToServerStats = routing.Stats{}
		sliceData.ServerToClientStats = routing.Stats{}
		sliceData.RouteDiversity = uint32(0)

		data, err := transport.WriteSessionSlice(&sliceData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var readSliceData transport.SessionSlice

		err = transport.ReadSessionSlice(&readSliceData, data)

		assert.NoError(t, err)
		assert.Equal(t, sliceData, readSliceData)
	})

	t.Run("test serialize write v2", func(t *testing.T) {
		sliceData := testSessionSlice()

		data, err := transport.WriteSessionSlice(&sliceData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("test serialize read v2", func(t *testing.T) {
		sliceData := testSessionSlice()

		data, err := transport.WriteSessionSlice(&sliceData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var readSliceData transport.SessionSlice

		err = transport.ReadSessionSlice(&readSliceData, data)

		assert.NoError(t, err)
		assert.Equal(t, sliceData, readSliceData)
	})
}

func testSessionMapPoint() transport.SessionMapPoint {
	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	data := transport.SessionMapPoint{
		Version:   transport.SessionMapPointVersion,
		Latitude:  rand.Float64(),
		Longitude: rand.Float64(),
		SessionID: rand.Uint64(),
	}

	return data
}

func TestSessionMapPoint_Serialize(t *testing.T) {
	t.Parallel()

	t.Run("test serialize write", func(t *testing.T) {
		mapData := testSessionMapPoint()

		data, err := transport.WriteSessionMapPoint(&mapData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("test serialize read", func(t *testing.T) {
		mapData := testSessionMapPoint()

		data, err := transport.WriteSessionMapPoint(&mapData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var readMapData transport.SessionMapPoint

		err = transport.ReadSessionMapPoint(&readMapData, data)

		assert.NoError(t, err)
		assert.Equal(t, mapData, readMapData)
	})
}

func testSessionPortalData() transport.SessionPortalData {
	data := transport.SessionPortalData{
		Version:       transport.SessionPortalDataVersion,
		Meta:          testSessionMeta(),
		Slice:         testSessionSlice(),
		Point:         testSessionMapPoint(),
		LargeCustomer: true,
		EverOnNext:    true,
	}

	return data
}

func TestSessionPortalData_Serialize(t *testing.T) {
	t.Parallel()

	t.Run("test serialize write", func(t *testing.T) {
		portalData := testSessionPortalData()

		data, err := transport.WriteSessionPortalData(&portalData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("test serialize read", func(t *testing.T) {
		portalData := testSessionPortalData()

		data, err := transport.WriteSessionPortalData(&portalData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var readPortalData transport.SessionPortalData

		err = transport.ReadSessionPortalData(&readPortalData, data)

		assert.NoError(t, err)
		assert.Equal(t, portalData, readPortalData)
	})
}
