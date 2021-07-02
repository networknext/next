package transport_test

import (
	"math/rand"
	"testing"
	"time"

	// "github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a random string of a specified length
// Useful for testing constant string lengths
// Adapted from: https://stackoverflow.com/a/22892986
func generateRandomStringSequence(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

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
		Name:    generateRandomStringSequence(rand.Intn(routing.MaxRelayNameLength)),
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
		Name:        generateRandomStringSequence(rand.Intn(routing.MaxRelayNameLength)),
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

func testSessionMeta() transport.SessionMeta {
	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	hops := make([]transport.RelayHop, rand.Intn(32))
	for i := 0; i < len(hops); i++ {
		hops[i] = testRelayHop()
	}

	nearRelays := make([]transport.NearRelayPortalData, rand.Intn(32))
	for i := 0; i < len(nearRelays); i++ {
		nearRelays[i] = testNearRelayPortalData()
	}

	data := transport.SessionMeta{
		Version:         transport.SessionMetaVersion,
		ID:              rand.Uint64(),
		UserHash:        rand.Uint64(),
		DatacenterName:  generateRandomStringSequence(rand.Intn(transport.MaxDatacenterNameLength)),
		DatacenterAlias: generateRandomStringSequence(rand.Intn(transport.MaxDatacenterNameLength)),
		OnNetworkNext:   true,
		NextRTT:         rand.Float64(),
		DirectRTT:       rand.Float64(),
		DeltaRTT:        rand.Float64(),
		Location: routing.Location{
			Continent:   generateRandomStringSequence(rand.Intn(routing.MaxContinentLength - 1)),
			Country:     generateRandomStringSequence(rand.Intn(routing.MaxCountryLength - 1)),
			CountryCode: generateRandomStringSequence(rand.Intn(routing.MaxCountryCodeLength - 1)),
			Region:      generateRandomStringSequence(rand.Intn(routing.MaxRegionLength - 1)),
			City:        generateRandomStringSequence(rand.Intn(routing.MaxCityLength - 1)),
			Latitude:    rand.Float32(),
			Longitude:   rand.Float32(),
			ISP:         generateRandomStringSequence(rand.Intn(routing.MaxISPNameLength - 1)),
			ASN:         rand.Int(),
		},
		ClientAddr:   generateRandomStringSequence(rand.Intn(transport.MaxAddressLength - 1)),
		ServerAddr:   generateRandomStringSequence(rand.Intn(transport.MaxAddressLength - 1)),
		Hops:         hops,
		SDK:          generateRandomStringSequence(rand.Intn(transport.MaxSDKVersionLength - 1)),
		Connection:   uint8(rand.Intn(256)),
		NearbyRelays: nearRelays,
		Platform:     uint8(rand.Intn(256)),
		BuyerID:      rand.Uint64(),
	}

	return data
}

func TestSessionMeta_Serialize(t *testing.T) {
	t.Parallel()

	t.Run("test serialize write", func(t *testing.T) {
		metaData := testSessionMeta()

		data, err := transport.WriteSessionMeta(&metaData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("test serialize read", func(t *testing.T) {
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

		// Some fields are not serialized
		sliceData.Predicted.Jitter = float64(0)
		sliceData.Predicted.PacketLoss = float64(0)

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

		// Some fields are not serialized
		sliceData.Predicted.Jitter = float64(0)
		sliceData.Predicted.PacketLoss = float64(0)
		sliceData.ClientToServerStats.RTT = float64(0)
		sliceData.ServerToClientStats.RTT = float64(0)

		data, err := transport.WriteSessionSlice(&sliceData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var readSliceData transport.SessionSlice

		err = transport.ReadSessionSlice(&readSliceData, data)

		assert.NoError(t, err)
		assert.Equal(t, sliceData, readSliceData)
	})
}
