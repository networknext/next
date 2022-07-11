package transport_test

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
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

func TestNearRelayHopData_RedisString(t *testing.T) {
	t.Parallel()

	t.Run("test to redis string", func(t *testing.T) {
		relayHopData := testRelayHop()
		relayHopRedisString := relayHopData.RedisString()
		assert.NotEqual(t, "", relayHopRedisString)

		relayHopStringTokens := strings.Split(relayHopRedisString, "|")

		index := 0
		assert.Equal(t, fmt.Sprintf("%d", relayHopData.Version), relayHopStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%016x", relayHopData.ID), relayHopStringTokens[index])
		index++
		assert.Equal(t, relayHopData.Name, relayHopStringTokens[index])
		index++

		assert.Equal(t, index, len(relayHopStringTokens))
	})

	t.Run("test parse redis string", func(t *testing.T) {
		relayHopData := testRelayHop()
		expectedRelayHopData := transport.RelayHop{}

		relayHopRedisString := relayHopData.RedisString()

		relayHopStringTokens := strings.Split(relayHopRedisString, "|")
		err := expectedRelayHopData.ParseRedisString(relayHopStringTokens)
		assert.NoError(t, err)

		assert.Equal(t, expectedRelayHopData.Version, relayHopData.Version)
		assert.Equal(t, expectedRelayHopData.ID, relayHopData.ID)
		assert.Equal(t, expectedRelayHopData.Name, relayHopData.Name)
	})
}

// TODO: Move this somewhere more accesible
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

func TestNearRelayPortalData_RedisString(t *testing.T) {
	t.Parallel()

	t.Run("test to redis string", func(t *testing.T) {
		nearRelayData := testNearRelayPortalData()
		nearRelayRedisString := nearRelayData.RedisString()
		assert.NotEqual(t, "", nearRelayRedisString)

		nearRelayStringTokens := strings.Split(nearRelayRedisString, "|")
		clientStatTokens := strings.Split(nearRelayData.ClientStats.RedisString(), "|")

		index := 0
		assert.Equal(t, fmt.Sprintf("%d", nearRelayData.Version), nearRelayStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%016x", nearRelayData.ID), nearRelayStringTokens[index])
		index++
		assert.Equal(t, nearRelayData.Name, nearRelayStringTokens[index])
		index++

		for i := 0; i < len(clientStatTokens); i++ {
			assert.Equal(t, clientStatTokens[i], nearRelayStringTokens[index])
			index++
		}

		assert.Equal(t, index, len(nearRelayStringTokens))
	})

	t.Run("test parse redis string", func(t *testing.T) {
		nearRelayData := testNearRelayPortalData()
		expectedNearRelayData := transport.NearRelayPortalData{}

		nearRelayRedisString := nearRelayData.RedisString()

		nearRelayStringTokens := strings.Split(nearRelayRedisString, "|")
		err := expectedNearRelayData.ParseRedisString(nearRelayStringTokens)
		assert.NoError(t, err)

		assert.Equal(t, expectedNearRelayData.Version, nearRelayData.Version)
		assert.Equal(t, expectedNearRelayData.Name, nearRelayData.Name)
		assert.Equal(t, fmt.Sprintf("%.2f", expectedNearRelayData.ClientStats.RTT), fmt.Sprintf("%.2f", nearRelayData.ClientStats.RTT))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedNearRelayData.ClientStats.Jitter), fmt.Sprintf("%.2f", nearRelayData.ClientStats.Jitter))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedNearRelayData.ClientStats.PacketLoss), fmt.Sprintf("%.2f", nearRelayData.ClientStats.PacketLoss))
	})
}

func testLocation() routing.Location {
	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	return routing.Location{
		Continent:   generateRandomStringSequence(rand.Intn(routing.MaxContinentLength - 1)),
		Country:     generateRandomStringSequence(rand.Intn(routing.MaxCountryLength - 1)),
		CountryCode: generateRandomStringSequence(rand.Intn(routing.MaxCountryCodeLength - 1)),
		Region:      generateRandomStringSequence(rand.Intn(routing.MaxRegionLength - 1)),
		City:        generateRandomStringSequence(rand.Intn(routing.MaxCityLength - 1)),
		Latitude:    rand.Float32(),
		Longitude:   rand.Float32(),
		ISP:         generateRandomStringSequence(rand.Intn(routing.MaxISPNameLength - 1)),
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
		DatacenterName:  generateRandomStringSequence(rand.Intn(transport.MaxDatacenterNameLength)),
		DatacenterAlias: generateRandomStringSequence(rand.Intn(transport.MaxDatacenterNameLength)),
		OnNetworkNext:   true,
		NextRTT:         rand.Float64(),
		DirectRTT:       rand.Float64(),
		DeltaRTT:        rand.Float64(),
		Location:        testLocation(),
		ClientAddr:      generateRandomStringSequence(rand.Intn(transport.MaxAddressLength - 1)),
		ServerAddr:      generateRandomStringSequence(rand.Intn(transport.MaxAddressLength - 1)),
		Hops:            hops,
		SDK:             generateRandomStringSequence(rand.Intn(transport.MaxSDKVersionLength - 1)),
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

func TestSessionMeta_RedisString(t *testing.T) {
	t.Parallel()

	t.Run("test to redis string", func(t *testing.T) {
		metaData := testSessionMeta()
		metaRedisString := metaData.RedisString()
		assert.NotEqual(t, "", metaRedisString)

		locationTokens := strings.Split(metaData.Location.RedisString(), "|")

		metaStringTokens := strings.Split(metaRedisString, "|")

		index := 0
		assert.Equal(t, fmt.Sprintf("%d", metaData.Version), metaStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%016x", metaData.ID), metaStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%016x", metaData.UserHash), metaStringTokens[index])
		index++
		assert.Equal(t, metaData.DatacenterName, metaStringTokens[index])
		index++
		assert.Equal(t, metaData.DatacenterAlias, metaStringTokens[index])
		index++

		onNetworkNextString := "0"
		if metaData.OnNetworkNext {
			onNetworkNextString = "1"
		}

		assert.Equal(t, onNetworkNextString, metaStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%.2f", metaData.NextRTT), metaStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%.2f", metaData.DirectRTT), metaStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%.2f", metaData.DeltaRTT), metaStringTokens[index])
		index++

		for i := 0; i < len(locationTokens); i++ {
			assert.Equal(t, locationTokens[i], metaStringTokens[index])
			index++
		}

		assert.Equal(t, metaData.ClientAddr, metaStringTokens[index])
		index++
		assert.Equal(t, metaData.ServerAddr, metaStringTokens[index])
		index++

		assert.Equal(t, fmt.Sprintf("%d", len(metaData.Hops)), metaStringTokens[index])
		index++
		for i := 0; i < len(metaData.Hops); i++ {
			hopTokens := strings.Split(metaData.Hops[i].RedisString(), "|")
			for j := 0; j < len(hopTokens); j++ {
				assert.Equal(t, hopTokens[j], metaStringTokens[index])
				index++
			}
		}

		assert.Equal(t, metaData.SDK, metaStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%d", metaData.Connection), metaStringTokens[index])
		index++

		assert.Equal(t, fmt.Sprintf("%d", len(metaData.NearbyRelays)), metaStringTokens[index])
		index++
		for i := 0; i < len(metaData.NearbyRelays); i++ {
			relayTokens := strings.Split(metaData.NearbyRelays[i].RedisString(), "|")
			for j := 0; j < len(relayTokens); j++ {
				assert.Equal(t, relayTokens[j], metaStringTokens[index])
				index++
			}
		}

		assert.Equal(t, fmt.Sprintf("%d", metaData.Platform), metaStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%016x", metaData.BuyerID), metaStringTokens[index])
		index++

		assert.Equal(t, index, len(metaStringTokens))
	})

	t.Run("test parse redis string", func(t *testing.T) {
		metaData := testSessionMeta()
		expectedMetaData := transport.SessionMeta{}

		metaRedisString := metaData.RedisString()

		metaStringTokens := strings.Split(metaRedisString, "|")
		err := expectedMetaData.ParseRedisString(metaStringTokens)
		assert.NoError(t, err)

		assert.Equal(t, expectedMetaData.Version, metaData.Version)
		assert.Equal(t, expectedMetaData.ID, metaData.ID)
		assert.Equal(t, expectedMetaData.BuyerID, metaData.BuyerID)
		assert.Equal(t, expectedMetaData.DatacenterName, metaData.DatacenterName)
		assert.Equal(t, expectedMetaData.DatacenterAlias, metaData.DatacenterAlias)
		assert.Equal(t, expectedMetaData.OnNetworkNext, metaData.OnNetworkNext)
		assert.Equal(t, expectedMetaData.NextRTT, math.Round(metaData.NextRTT*100)/100)
		assert.Equal(t, expectedMetaData.DirectRTT, math.Round(metaData.DirectRTT*100)/100)
		assert.Equal(t, expectedMetaData.DeltaRTT, math.Round(metaData.DeltaRTT*100)/100)
		assert.Equal(t, fmt.Sprintf("%.2f", expectedMetaData.Location.Latitude), fmt.Sprintf("%.2f", metaData.Location.Latitude))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedMetaData.Location.Longitude), fmt.Sprintf("%.2f", metaData.Location.Longitude))
		assert.Equal(t, expectedMetaData.Location.City, metaData.Location.City)
		assert.Equal(t, expectedMetaData.Location.Country, metaData.Location.Country)
		assert.Equal(t, expectedMetaData.Location.CountryCode, metaData.Location.CountryCode)
		assert.Equal(t, expectedMetaData.Location.Region, metaData.Location.Region)
		assert.Equal(t, expectedMetaData.Location.ISP, metaData.Location.ISP)
		assert.Equal(t, expectedMetaData.ClientAddr, metaData.ClientAddr)
		assert.Equal(t, expectedMetaData.ServerAddr, metaData.ServerAddr)
		assert.Equal(t, expectedMetaData.SDK, metaData.SDK)
		assert.Equal(t, expectedMetaData.Connection, metaData.Connection)
		assert.Equal(t, expectedMetaData.Platform, metaData.Platform)

		for i, hop := range metaData.Hops {
			assert.Equal(t, expectedMetaData.Hops[i].ID, hop.ID)
			assert.Equal(t, expectedMetaData.Hops[i].Name, hop.Name)
			assert.Equal(t, expectedMetaData.Hops[i].Version, hop.Version)
		}

		for i, relay := range metaData.NearbyRelays {
			assert.Equal(t, expectedMetaData.NearbyRelays[i].ID, relay.ID)
			assert.Equal(t, expectedMetaData.NearbyRelays[i].Name, relay.Name)
			assert.Equal(t, expectedMetaData.NearbyRelays[i].Version, relay.Version)
			assert.Equal(t, fmt.Sprintf("%.2f", expectedMetaData.NearbyRelays[i].ClientStats.RTT), fmt.Sprintf("%.2f", relay.ClientStats.RTT))
			assert.Equal(t, fmt.Sprintf("%.2f", expectedMetaData.NearbyRelays[i].ClientStats.Jitter), fmt.Sprintf("%.2f", relay.ClientStats.Jitter))
			assert.Equal(t, fmt.Sprintf("%.2f", expectedMetaData.NearbyRelays[i].ClientStats.PacketLoss), fmt.Sprintf("%.2f", relay.ClientStats.PacketLoss))
		}
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

func TestSessionSlice_RedisString(t *testing.T) {
	t.Parallel()

	t.Run("test to redis string", func(t *testing.T) {
		sliceData := testSessionSlice()
		sliceRedisString := sliceData.RedisString()
		assert.NotEqual(t, "", sliceRedisString)

		nextStatTokens := strings.Split(sliceData.Next.RedisString(), "|")
		directStatTokens := strings.Split(sliceData.Direct.RedisString(), "|")
		predictedStatTokens := strings.Split(sliceData.Predicted.RedisString(), "|")
		clientServerStatTokens := strings.Split(sliceData.ClientToServerStats.RedisString(), "|")
		serverClientStatTokens := strings.Split(sliceData.ServerToClientStats.RedisString(), "|")
		envelopeTokens := strings.Split(sliceData.Envelope.RedisString(), "|")

		sliceStringTokens := strings.Split(sliceRedisString, "|")

		index := 0
		assert.Equal(t, fmt.Sprintf("%d", sliceData.Version), sliceStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%d", sliceData.Timestamp.Unix()), sliceStringTokens[index])
		index++

		for i := 0; i < len(nextStatTokens); i++ {
			assert.Equal(t, nextStatTokens[i], sliceStringTokens[index])
			index++
		}

		for i := 0; i < len(directStatTokens); i++ {
			assert.Equal(t, directStatTokens[i], sliceStringTokens[index])
			index++
		}

		for i := 0; i < len(predictedStatTokens); i++ {
			assert.Equal(t, predictedStatTokens[i], sliceStringTokens[index])
			index++
		}

		for i := 0; i < len(clientServerStatTokens); i++ {
			assert.Equal(t, clientServerStatTokens[i], sliceStringTokens[index])
			index++
		}

		for i := 0; i < len(serverClientStatTokens); i++ {
			assert.Equal(t, serverClientStatTokens[i], sliceStringTokens[index])
			index++
		}

		assert.Equal(t, fmt.Sprintf("%d", sliceData.RouteDiversity), sliceStringTokens[index])
		index++

		for i := 0; i < len(envelopeTokens); i++ {
			assert.Equal(t, envelopeTokens[i], sliceStringTokens[index])
			index++
		}

		onNetworkNextString := "0"
		if sliceData.OnNetworkNext {
			onNetworkNextString = "1"
		}

		isMultipathString := "0"
		if sliceData.IsMultiPath {
			isMultipathString = "1"
		}

		isTryBeforeYouBuyString := "0"
		if sliceData.IsTryBeforeYouBuy {
			isTryBeforeYouBuyString = "1"
		}

		assert.Equal(t, onNetworkNextString, sliceStringTokens[index])
		index++
		assert.Equal(t, isMultipathString, sliceStringTokens[index])
		index++
		assert.Equal(t, isTryBeforeYouBuyString, sliceStringTokens[index])
		index++

		assert.Equal(t, index, len(sliceStringTokens))
	})

	t.Run("test parse redis string", func(t *testing.T) {
		sliceData := testSessionSlice()
		expectedSliceData := transport.SessionSlice{}

		sliceRedisString := sliceData.RedisString()

		sliceStringTokens := strings.Split(sliceRedisString, "|")
		err := expectedSliceData.ParseRedisString(sliceStringTokens)
		assert.NoError(t, err)

		assert.Equal(t, expectedSliceData.Version, sliceData.Version)
		assert.Equal(t, expectedSliceData.Timestamp.Unix(), sliceData.Timestamp.Unix())

		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.Next.RTT), fmt.Sprintf("%.2f", sliceData.Next.RTT))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.Next.Jitter), fmt.Sprintf("%.2f", sliceData.Next.Jitter))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.Next.PacketLoss), fmt.Sprintf("%.2f", sliceData.Next.PacketLoss))

		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.Direct.RTT), fmt.Sprintf("%.2f", sliceData.Direct.RTT))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.Direct.Jitter), fmt.Sprintf("%.2f", sliceData.Direct.Jitter))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.Direct.PacketLoss), fmt.Sprintf("%.2f", sliceData.Direct.PacketLoss))

		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.Predicted.RTT), fmt.Sprintf("%.2f", sliceData.Predicted.RTT))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.Predicted.Jitter), fmt.Sprintf("%.2f", sliceData.Predicted.Jitter))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.Predicted.PacketLoss), fmt.Sprintf("%.2f", sliceData.Predicted.PacketLoss))

		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.ClientToServerStats.RTT), fmt.Sprintf("%.2f", sliceData.ClientToServerStats.RTT))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.ClientToServerStats.Jitter), fmt.Sprintf("%.2f", sliceData.ClientToServerStats.Jitter))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.ClientToServerStats.PacketLoss), fmt.Sprintf("%.2f", sliceData.ClientToServerStats.PacketLoss))

		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.ServerToClientStats.RTT), fmt.Sprintf("%.2f", sliceData.ServerToClientStats.RTT))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.ServerToClientStats.Jitter), fmt.Sprintf("%.2f", sliceData.ServerToClientStats.Jitter))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedSliceData.ServerToClientStats.PacketLoss), fmt.Sprintf("%.2f", sliceData.ServerToClientStats.PacketLoss))
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

func TestMapPoint_RedisString(t *testing.T) {
	t.Parallel()

	t.Run("test to redis string", func(t *testing.T) {
		mapPointData := testSessionMapPoint()
		mapPointRedisString := mapPointData.RedisString()
		assert.NotEqual(t, "", mapPointRedisString)

		mapPointStringTokens := strings.Split(mapPointRedisString, "|")

		index := 0
		assert.Equal(t, fmt.Sprintf("%d", mapPointData.Version), mapPointStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%.2f", mapPointData.Latitude), mapPointStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%.2f", mapPointData.Longitude), mapPointStringTokens[index])
		index++
		assert.Equal(t, fmt.Sprintf("%016x", mapPointData.SessionID), mapPointStringTokens[index])
		index++

		assert.Equal(t, index, len(mapPointStringTokens))
	})

	t.Run("test parse redis string", func(t *testing.T) {
		mapPointData := testSessionMapPoint()
		expectedMapPointData := transport.SessionMapPoint{}

		mapPointRedisString := mapPointData.RedisString()

		mapPointStringTokens := strings.Split(mapPointRedisString, "|")
		err := expectedMapPointData.ParseRedisString(mapPointStringTokens)
		assert.NoError(t, err)

		assert.Equal(t, expectedMapPointData.Version, mapPointData.Version)

		assert.Equal(t, fmt.Sprintf("%.2f", expectedMapPointData.Latitude), fmt.Sprintf("%.2f", mapPointData.Latitude))
		assert.Equal(t, fmt.Sprintf("%.2f", expectedMapPointData.Longitude), fmt.Sprintf("%.2f", mapPointData.Longitude))
		assert.Equal(t, expectedMapPointData.SessionID, mapPointData.SessionID)
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
