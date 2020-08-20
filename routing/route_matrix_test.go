package routing_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func getPopulatedRouteMatrix(malformed bool) *routing.RouteMatrix {
	var matrix routing.RouteMatrix

	matrix.RelayIndices = make(map[uint64]int)
	matrix.RelayIndices[123] = 0
	matrix.RelayIndices[456] = 1

	matrix.RelayIDs = make([]uint64, 2)
	matrix.RelayIDs[0] = 123
	matrix.RelayIDs[1] = 456

	if !malformed {
		matrix.RelayNames = make([]string, 2)
		matrix.RelayNames[0] = "first"
		matrix.RelayNames[1] = "second"
	} else {
		matrix.RelayNames = make([]string, 1)
		matrix.RelayNames[0] = "first"
	}

	tmpAddr1 := make([]byte, routing.MaxRelayAddressLength)
	tmpAddr2 := make([]byte, routing.MaxRelayAddressLength)

	matrix.RelayAddresses = make([][]byte, 2)
	rand.Read(tmpAddr1)
	matrix.RelayAddresses[0] = tmpAddr1
	rand.Read(tmpAddr2)
	matrix.RelayAddresses[1] = tmpAddr2

	matrix.RelayPublicKeys = make([][]byte, 2)
	matrix.RelayPublicKeys[0] = randomPublicKey()
	matrix.RelayPublicKeys[1] = randomPublicKey()

	matrix.DatacenterIDs = make([]uint64, 2)
	matrix.DatacenterIDs[0] = 999
	matrix.DatacenterIDs[1] = 111

	matrix.DatacenterNames = make([]string, 2)
	matrix.DatacenterNames[0] = "a name"
	matrix.DatacenterNames[1] = "another name"

	matrix.DatacenterRelays = make(map[uint64][]uint64)
	matrix.DatacenterRelays[999] = make([]uint64, 1)
	matrix.DatacenterRelays[999][0] = 123
	matrix.DatacenterRelays[111] = make([]uint64, 1)
	matrix.DatacenterRelays[111][0] = 456

	matrix.Entries = []routing.RouteMatrixEntry{
		{
			DirectRTT:      123,
			NumRoutes:      1,
			RouteRTT:       [8]int32{1},
			RouteNumRelays: [8]int32{2},
			RouteRelays:    [8][5]uint64{{123, 456}},
		},
	}

	matrix.RelaySellers = []routing.Seller{
		{Name: "Seller One"}, {Name: "Seller Two"},
	}

	matrix.RelaySessionCounts = []uint32{100, 200}
	matrix.RelayMaxSessionCounts = []uint32{100, 200}

	matrix.UpdateRelayAddressCache()

	matrix.RelayLatitude = []float64{1.0, 2.0}
	matrix.RelayLongitude = []float64{3.0, 4.0}

	return &matrix
}

func generateRouteMatrixEntries(entries []routing.RouteMatrixEntry) {
	for i := 0; i < len(entries); i++ {
		entry := routing.RouteMatrixEntry{
			DirectRTT: rand.Int31(),
			NumRoutes: 8,
		}

		var routeRTT [8]int32
		for j := 0; j < 8; j++ {
			routeRTT[j] = rand.Int31()
		}
		entry.RouteRTT = routeRTT

		var routeNumRelays [8]int32
		for j := 0; j < 8; j++ {
			routeNumRelays[j] = 5
		}
		entry.RouteNumRelays = routeNumRelays

		var routeRelays [8][5]uint64
		for j := 0; j < 8; j++ {
			for k := 0; k < 5; k++ {
				// doesn't have to be accurrate
				routeRelays[j][k] = rand.Uint64()
			}
		}
		entry.RouteRelays = routeRelays

		entries[i] = entry
	}
}

func routeMatrixUnmarshalAssertionsVer5(t *testing.T, matrix *routing.RouteMatrix, numRelays, numDatacenters int, relayIDs, datacenters []uint64, relayAddrs []string, datacenterRelays [][]uint64, publicKeys [][]byte, entries []routing.RouteMatrixEntry, relayNames []string, datacenterIDs []uint64, datacenterNames []string, sellers []routing.Seller, sessionCounts []uint32, maxSessionCounts []uint32) {
	assert.Len(t, matrix.RelayIDs, numRelays)
	assert.Len(t, matrix.RelayAddresses, numRelays)
	assert.Len(t, matrix.RelayPublicKeys, numRelays)
	assert.Len(t, matrix.DatacenterRelays, numDatacenters)
	assert.Len(t, matrix.Entries, len(entries))

	for _, id := range relayIDs {
		assert.Contains(t, matrix.RelayIDs, id)
	}

	for _, addr := range relayAddrs {
		tmp := make([]byte, routing.MaxRelayAddressLength)
		copy(tmp, addr)
		assert.Contains(t, matrix.RelayAddresses, tmp)
	}

	for _, pk := range publicKeys {
		assert.Contains(t, matrix.RelayPublicKeys, pk)
	}

	for i := 0; i < numDatacenters; i++ {
		assert.Contains(t, matrix.DatacenterRelays, datacenters[i])

		relays := matrix.DatacenterRelays[datacenters[i]]
		for j := 0; j < len(datacenterRelays[i]); j++ {
			assert.Contains(t, relays, datacenterRelays[i][j])
		}
	}

	for i, expected := range entries {
		actual := matrix.Entries[i]

		assert.Equal(t, expected.DirectRTT, actual.DirectRTT)
		assert.Equal(t, expected.NumRoutes, actual.NumRoutes)
		assert.Equal(t, expected.RouteRTT, actual.RouteRTT)
		assert.Equal(t, expected.RouteNumRelays, actual.RouteNumRelays)

		for i, ids := range expected.RouteRelays {
			for j, id := range ids {
				assert.Equal(t, id, actual.RouteRelays[i][j])
			}
		}
	}

	assert.Len(t, matrix.RelayNames, len(relayNames))
	assert.Len(t, matrix.RelayIDs, len(relayNames))
	for _, name := range relayNames {
		assert.Contains(t, matrix.RelayNames, name)
	}

	assert.Len(t, matrix.DatacenterIDs, len(datacenterIDs))
	assert.Len(t, matrix.DatacenterNames, len(datacenterNames))

	for _, id := range datacenterIDs {
		assert.Contains(t, matrix.DatacenterIDs, id)
	}

	for _, name := range datacenterNames {
		assert.Contains(t, matrix.DatacenterNames, name)
	}

	assert.Len(t, matrix.RelaySellers, len(sellers))
	for i, seller := range sellers {
		assert.Equal(t, matrix.RelaySellers[i].ID, seller.ID)
		assert.Equal(t, matrix.RelaySellers[i].Name, seller.Name)
		assert.Equal(t, matrix.RelaySellers[i].IngressPriceNibblinsPerGB, seller.IngressPriceNibblinsPerGB)
		assert.Equal(t, matrix.RelaySellers[i].EgressPriceNibblinsPerGB, seller.EgressPriceNibblinsPerGB)
	}

	assert.Equal(t, matrix.RelaySessionCounts, sessionCounts)
	assert.Equal(t, matrix.RelayMaxSessionCounts, maxSessionCounts)
}

func routeMatrixUnmarshalAssertionsVer6(t *testing.T, matrix *routing.RouteMatrix, latitudes []float64, longitudes []float64) {
	assert.Equal(t, matrix.RelayLatitude, latitudes)
	assert.Equal(t, matrix.RelayLongitude, longitudes)
}

type routeMatrixData struct {
	buff             []byte
	numRelays        int
	relayIDs         []uint64
	relayNames       []string
	numDatacenters   int
	datacenterIDs    []uint64
	datacenterNames  []string
	relayAddrs       []string
	relayLatitudes   []float64
	relayLongitudes  []float64
	publicKeys       [][]byte
	datacenterRelays [][]uint64
	entries          []routing.RouteMatrixEntry
	sellers          []routing.Seller
	sessionCounts    []uint32
	maxSessionCounts []uint32
}

func getRouteMatrixDataV5() routeMatrixData {
	relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
	relayIDs := addrsToIDs(relayAddrs)
	numRelays := len(relayAddrs)
	publicKeys := [][]byte{
		randomPublicKey(),
		randomPublicKey(),
		randomPublicKey(),
		randomPublicKey(),
		randomPublicKey(),
	}
	datacenters := []uint64{0, 1, 2, 3, 4}
	numDatacenters := len(datacenters)
	datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}
	numEntries := routing.TriMatrixLength(numRelays)
	entries := make([]routing.RouteMatrixEntry, numEntries)
	generateRouteMatrixEntries(entries)

	relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}

	datacenterNames := []string{"a datacenter", "another datacenter", "third", "fourth", "fifth"}

	sellers := []routing.Seller{
		{ID: "id0", Name: "name0", IngressPriceNibblinsPerGB: 1, EgressPriceNibblinsPerGB: 2},
		{ID: "id1", Name: "name1", IngressPriceNibblinsPerGB: 3, EgressPriceNibblinsPerGB: 4},
		{ID: "id2", Name: "name2", IngressPriceNibblinsPerGB: 5, EgressPriceNibblinsPerGB: 6},
		{ID: "id3", Name: "name3", IngressPriceNibblinsPerGB: 7, EgressPriceNibblinsPerGB: 8},
		{ID: "id4", Name: "name4", IngressPriceNibblinsPerGB: 9, EgressPriceNibblinsPerGB: 10},
	}

	sessionCounts := []uint32{100, 200, 300, 400, 500}
	maxSessionCounts := []uint32{3000, 3000, 3000, 3000, 6000}

	buffSize := 0
	buffSize += sizeofVersionNumber()
	buffSize += sizeofRelayCount()
	buffSize += sizeofRelayIDs64(relayIDs)
	buffSize += sizeofRelayNames(relayNames)
	buffSize += sizeofDatacenterCount()
	buffSize += sizeofDatacenterIDs64(datacenters)
	buffSize += sizeofDatacenterNames(datacenterNames)
	buffSize += sizeofRelayAddress(relayAddrs)
	buffSize += sizeofRelayPublicKeys(publicKeys)
	buffSize += sizeofDataCenterCount2()
	buffSize += sizeofDatacenterIDs64(datacenters)
	buffSize += sizeofRelaysInDatacenterCount(datacenters)
	buffSize += sizeofRelayIDs64(relayIDs)
	buffSize += sizeofRouteMatrixEntry(entries)
	buffSize += sizeofSellers(sellers)
	buffSize += sizeofSessionCounts(sessionCounts)
	buffSize += sizeofMaxSessionCounts(maxSessionCounts)

	buff := make([]byte, buffSize)

	offset := 0
	putVersionNumber(buff, &offset, 5)
	putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
	putRelayNames(buff, &offset, relayNames)
	putDatacenterStuff(buff, &offset, datacenters, datacenterNames)
	putRelayAddresses(buff, &offset, relayAddrs)
	putRelayPublicKeys(buff, &offset, publicKeys)
	putDatacenters(buff, &offset, datacenters, datacenterRelays)
	putEntries(buff, &offset, entries)
	putSellers(buff, &offset, sellers)
	putSessionCounts(buff, &offset, sessionCounts)
	putSessionCounts(buff, &offset, maxSessionCounts)

	return routeMatrixData{
		buff:             buff,
		numRelays:        numRelays,
		relayIDs:         relayIDs,
		relayNames:       relayNames,
		numDatacenters:   numDatacenters,
		datacenterIDs:    datacenters,
		datacenterNames:  datacenterNames,
		relayAddrs:       relayAddrs,
		publicKeys:       publicKeys,
		datacenterRelays: datacenterRelays,
		entries:          entries,
		sellers:          sellers,
		sessionCounts:    sessionCounts,
		maxSessionCounts: maxSessionCounts,
	}
}

func getRouteMatrixDataV6() routeMatrixData {
	// version 5 stuff
	relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
	relayIDs := addrsToIDs(relayAddrs)
	numRelays := len(relayAddrs)
	publicKeys := [][]byte{
		randomPublicKey(),
		randomPublicKey(),
		randomPublicKey(),
		randomPublicKey(),
		randomPublicKey(),
	}
	datacenters := []uint64{0, 1, 2, 3, 4}
	numDatacenters := len(datacenters)
	datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}
	numEntries := routing.TriMatrixLength(numRelays)
	entries := make([]routing.RouteMatrixEntry, numEntries)
	generateRouteMatrixEntries(entries)

	relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}

	datacenterNames := []string{"a datacenter", "another datacenter", "third", "fourth", "fifth"}

	sellers := []routing.Seller{
		{ID: "id0", Name: "name0", IngressPriceNibblinsPerGB: 1, EgressPriceNibblinsPerGB: 2},
		{ID: "id1", Name: "name1", IngressPriceNibblinsPerGB: 3, EgressPriceNibblinsPerGB: 4},
		{ID: "id2", Name: "name2", IngressPriceNibblinsPerGB: 5, EgressPriceNibblinsPerGB: 6},
		{ID: "id3", Name: "name3", IngressPriceNibblinsPerGB: 7, EgressPriceNibblinsPerGB: 8},
		{ID: "id4", Name: "name4", IngressPriceNibblinsPerGB: 9, EgressPriceNibblinsPerGB: 10},
	}

	sessionCounts := []uint32{100, 200, 300, 400, 500}
	maxSessionCounts := []uint32{3000, 3000, 3000, 3000, 6000}

	// version 6 stuff
	relayLatitudes := []float64{90, 80, 70, 60, 50}
	relayLongitudes := []float64{180, 170, 160, 150, 140}

	buffSize := 0
	buffSize += sizeofVersionNumber()
	buffSize += sizeofRelayCount()
	buffSize += sizeofRelayIDs64(relayIDs)
	buffSize += sizeofRelayNames(relayNames)
	buffSize += sizeofDatacenterCount()
	buffSize += sizeofDatacenterIDs64(datacenters)
	buffSize += sizeofDatacenterNames(datacenterNames)
	buffSize += sizeofRelayAddress(relayAddrs)
	buffSize += sizeofRelayLatitudes(relayLatitudes)
	buffSize += sizeofRelayLongitudes(relayLongitudes)
	buffSize += sizeofRelayPublicKeys(publicKeys)
	buffSize += sizeofDataCenterCount2()
	buffSize += sizeofDatacenterIDs64(datacenters)
	buffSize += sizeofRelaysInDatacenterCount(datacenters)
	buffSize += sizeofRelayIDs64(relayIDs)
	buffSize += sizeofRouteMatrixEntry(entries)
	buffSize += sizeofSellers(sellers)
	buffSize += sizeofSessionCounts(sessionCounts)
	buffSize += sizeofMaxSessionCounts(maxSessionCounts)

	buff := make([]byte, buffSize)

	offset := 0
	putVersionNumber(buff, &offset, 6)
	putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
	putRelayNames(buff, &offset, relayNames)
	putDatacenterStuff(buff, &offset, datacenters, datacenterNames)
	putRelayAddresses(buff, &offset, relayAddrs)
	putRelayLatitudes(buff, &offset, relayLatitudes)
	putRelayLongitudes(buff, &offset, relayLongitudes)
	putRelayPublicKeys(buff, &offset, publicKeys)
	putDatacenters(buff, &offset, datacenters, datacenterRelays)
	putEntries(buff, &offset, entries)
	putSellers(buff, &offset, sellers)
	putSessionCounts(buff, &offset, sessionCounts)
	putSessionCounts(buff, &offset, maxSessionCounts)

	return routeMatrixData{
		buff:             buff,
		numRelays:        numRelays,
		relayIDs:         relayIDs,
		relayNames:       relayNames,
		numDatacenters:   numDatacenters,
		datacenterIDs:    datacenters,
		datacenterNames:  datacenterNames,
		relayAddrs:       relayAddrs,
		relayLatitudes:   relayLatitudes,
		relayLongitudes:  relayLongitudes,
		publicKeys:       publicKeys,
		datacenterRelays: datacenterRelays,
		entries:          entries,
		sellers:          sellers,
		sessionCounts:    sessionCounts,
		maxSessionCounts: maxSessionCounts,
	}
}

func TestRouteMatrixUnmarshalBinaryV5(t *testing.T) {
	data := getRouteMatrixDataV5()

	t.Run("version of incoming bin data too high", func(t *testing.T) {
		buff := make([]byte, 4)
		offset := 0
		putVersionNumber(buff, &offset, 100)
		var matrix routing.RouteMatrix

		err := matrix.UnmarshalBinary(buff)

		assert.EqualError(t, err, "unknown route matrix version: 100")
	})

	t.Run("Invalid version read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at version number")
	})

	t.Run("Invalid relay count read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays")
	})

	t.Run("Invalid relay id read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids - ver >= v3")
	})

	t.Run("Invalid relay name read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relay names")
	})

	t.Run("Invalid datacenter count read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter count")
	})

	t.Run("Invalid datacenter id read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 8 + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter ids - ver >= v3")
	})

	t.Run("Invalid datacenter name read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter names")
	})

	t.Run("Invalid relay address read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relay addresses - ver >= v3")
	})

	t.Run("Invalid relay public key read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relay public keys - ver >= v3")
	})

	t.Run("Invalid datacenter count read second time", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at number of datacenters (second time)")
	})

	t.Run("Invalid datacenter id read second time", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter id - ver >= v3")
	})

	t.Run("Invalid datacenter relay count read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in datacenter")
	})

	t.Run("Invalid datacenter relay id read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 8 + 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids for datacenter - ver >= v3")
	})

	t.Run("Invalid direct route RTT read in matrix entry", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at direct rtt")
	})

	t.Run("Invalid route count read in matrix entry", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + 4 + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at number of routes")
	})

	t.Run("Invalid route RTT read in matrix entry", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + 4 + 4 + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at route rtt")
	})

	t.Run("Invalid relay count read in matrix entry", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + 4 + 4 + 4 + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in route")
	})

	t.Run("Invalid relay read in matrix entry", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relays in route - ver >= v3")
	})

	t.Run("Invalid seller ID read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + len(data.sellers[0].ID) + sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read on relay seller ID")
	})

	t.Run("Invalid seller name read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + len(data.sellers[0].Name) + 4 + len(data.sellers[0].ID) + sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read on relay seller name")
	})

	t.Run("Invalid seller ingress price read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 8 + 4 + len(data.sellers[0].Name) + 4 + len(data.sellers[0].ID) + sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read on relay seller ingress price")
	})

	t.Run("Invalid seller egress price read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 8 + 8 + 4 + len(data.sellers[0].Name) + 4 + len(data.sellers[0].ID) + sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read on relay seller egress price")
	})

	t.Run("Invalid session count read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + sizeofSellers(data.sellers) + sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read on relay session count")
	})

	t.Run("Invalid max session count read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + sizeofSessionCounts(data.sessionCounts) + sizeofSellers(data.sellers) + sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read on relay max session count")
	})

	t.Run("Success", func(t *testing.T) {
		var matrix routing.RouteMatrix
		err := matrix.UnmarshalBinary(data.buff)
		assert.Nil(t, err)
		routeMatrixUnmarshalAssertionsVer5(t, &matrix, data.numRelays, data.numDatacenters, data.relayIDs, data.datacenterIDs, data.relayAddrs, data.datacenterRelays, data.publicKeys, data.entries, data.relayNames, data.datacenterIDs, data.datacenterNames, data.sellers, data.sessionCounts, data.maxSessionCounts)
	})
}

func TestRouteMatrixUnmarshalBinaryV6(t *testing.T) {
	data := getRouteMatrixDataV6()

	t.Run("version of incoming bin data too high", func(t *testing.T) {
		buff := make([]byte, 4)
		offset := 0
		putVersionNumber(buff, &offset, 100)
		var matrix routing.RouteMatrix

		err := matrix.UnmarshalBinary(buff)

		assert.EqualError(t, err, "unknown route matrix version: 100")
	})

	t.Run("Invalid version read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at version number")
	})

	t.Run("Invalid relay count read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays")
	})

	t.Run("Invalid relay id read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids - ver >= v3")
	})

	t.Run("Invalid relay name read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relay names")
	})

	t.Run("Invalid datacenter count read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter count")
	})

	t.Run("Invalid datacenter id read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 8 + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter ids - ver >= v3")
	})

	t.Run("Invalid datacenter name read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter names")
	})

	t.Run("Invalid relay address read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relay addresses - ver >= v3")
	})

	t.Run("Invalid relay latitude read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relay latitude")
	})

	t.Run("Invalid relay longitude read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relay longitude")
	})

	t.Run("Invalid relay public key read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relay public keys - ver >= v3")
	})

	t.Run("Invalid datacenter count read second time", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at number of datacenters (second time)")
	})

	t.Run("Invalid datacenter id read second time", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter id - ver >= v3")
	})

	t.Run("Invalid datacenter relay count read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in datacenter")
	})

	t.Run("Invalid datacenter relay id read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 8 + 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids for datacenter - ver >= v3")
	})

	t.Run("Invalid direct route RTT read in matrix entry", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at direct rtt")
	})

	t.Run("Invalid route count read in matrix entry", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + 4 + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at number of routes")
	})

	t.Run("Invalid route RTT read in matrix entry", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + 4 + 4 + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at route rtt")
	})

	t.Run("Invalid relay count read in matrix entry", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + 4 + 4 + 4 + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in route")
	})

	t.Run("Invalid relay read in matrix entry", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read at relays in route - ver >= v3")
	})

	t.Run("Invalid seller ID read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + len(data.sellers[0].ID) + sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read on relay seller ID")
	})

	t.Run("Invalid seller name read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + len(data.sellers[0].Name) + 4 + len(data.sellers[0].ID) + sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read on relay seller name")
	})

	t.Run("Invalid seller ingress price read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 8 + 4 + len(data.sellers[0].Name) + 4 + len(data.sellers[0].ID) + sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read on relay seller ingress price")
	})

	t.Run("Invalid seller egress price read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 8 + 8 + 4 + len(data.sellers[0].Name) + 4 + len(data.sellers[0].ID) + sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read on relay seller egress price")
	})

	t.Run("Invalid session count read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + sizeofSellers(data.sellers) + sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read on relay session count")
	})

	t.Run("Invalid max session count read", func(t *testing.T) {
		var matrix routing.RouteMatrix
		offset := 4 + sizeofSessionCounts(data.sessionCounts) + sizeofSellers(data.sellers) + sizeofRouteMatrixEntry(data.entries) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) +
			sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[RouteMatrix] invalid read on relay max session count")
	})

	t.Run("Success", func(t *testing.T) {
		var matrix routing.RouteMatrix
		err := matrix.UnmarshalBinary(data.buff)
		assert.Nil(t, err)
		routeMatrixUnmarshalAssertionsVer5(t, &matrix, data.numRelays, data.numDatacenters, data.relayIDs, data.datacenterIDs, data.relayAddrs, data.datacenterRelays, data.publicKeys, data.entries, data.relayNames, data.datacenterIDs, data.datacenterNames, data.sellers, data.sessionCounts, data.maxSessionCounts)
		routeMatrixUnmarshalAssertionsVer6(t, &matrix, data.relayLatitudes, data.relayLongitudes)
	})
}

func TestRouteMatrixMarshalBinary(t *testing.T) {
	t.Run("MarshalBinary -> UnmarshalBinary equality", func(t *testing.T) {
		matrix := getPopulatedRouteMatrix(false)

		var other routing.RouteMatrix

		bin, err := matrix.MarshalBinary()

		// essentialy this asserts the result of MarshalBinary(),
		// if Unmarshal tests pass then the binary data from Marshal
		// is valid if unmarshaling equals the original
		other.UnmarshalBinary(bin)

		assert.Nil(t, err)
		assert.Equal(t, matrix, &other)
	})

	t.Run("Relay ID and name buffers different sizes", func(t *testing.T) {
		var matrix routing.RouteMatrix

		matrix.RelayIDs = make([]uint64, 2)
		matrix.RelayIDs[0] = 123
		matrix.RelayIDs[1] = 456

		matrix.RelayNames = make([]string, 1) // Only 1 name but 2 IDs
		matrix.RelayNames[0] = "first"

		_, err := matrix.MarshalBinary()
		errorString := fmt.Errorf("length of Relay IDs not equal to length of Relay Names: %d != %d", len(matrix.RelayIDs), len(matrix.RelayNames))
		assert.EqualError(t, err, errorString.Error())
	})

	t.Run("Datacenter ID and name buffers different sizes", func(t *testing.T) {
		var matrix routing.RouteMatrix

		matrix.DatacenterIDs = make([]uint64, 2)
		matrix.DatacenterIDs[0] = 999
		matrix.DatacenterIDs[1] = 111

		matrix.DatacenterNames = make([]string, 1) // Only 1 name but 2 IDs
		matrix.DatacenterNames[0] = "a name"

		_, err := matrix.MarshalBinary()
		errorString := fmt.Errorf("length of Datacenter IDs not equal to length of Datacenter Names: %d != %d", len(matrix.DatacenterIDs), len(matrix.DatacenterNames))
		assert.EqualError(t, err, errorString.Error())
	})
}

func TestRouteMatrixWriteTo(t *testing.T) {
	t.Run("Error during MarshalBinary()", func(t *testing.T) {
		// Create and populate a malformed route matrix
		matrix := getPopulatedRouteMatrix(true)

		var buff bytes.Buffer
		_, err := matrix.WriteTo(&buff)
		assert.EqualError(t, err, fmt.Sprintf("length of Relay IDs not equal to length of Relay Names: %v != %v", len(matrix.RelayIDs), len(matrix.RelayNames)))
	})

	t.Run("Error during write", func(t *testing.T) {
		// Create and populate a route matrix
		matrix := getPopulatedRouteMatrix(false)

		var buff ErrorBuffer
		_, err := matrix.WriteTo(&buff)
		assert.Error(t, err)
	})

	t.Run("Success", func(t *testing.T) {
		// Create and populate a route matrix
		matrix := getPopulatedRouteMatrix(false)

		var buff bytes.Buffer
		_, err := matrix.WriteTo(&buff)
		assert.NoError(t, err)
	})
}

func TestRouteMatrixReadFrom(t *testing.T) {
	t.Run("ReadFrom()", func(t *testing.T) {
		t.Run("Nil reader", func(t *testing.T) {
			// Create and populate a route matrix
			matrix := getPopulatedRouteMatrix(false)

			// Try to read from nil reader
			_, err := matrix.ReadFrom(nil)
			assert.EqualError(t, err, "reader is nil")
		})

		t.Run("Error during read", func(t *testing.T) {
			// Create and populate a route matrix
			matrix := getPopulatedRouteMatrix(false)

			// Try to read into the ErrorBuffer
			var buff ErrorBuffer
			_, err := matrix.ReadFrom(&buff)
			assert.Error(t, err)
		})

		t.Run("Error during UnmarshalBinary()", func(t *testing.T) {
			// Create and populate a route matrix
			matrix := getPopulatedRouteMatrix(false)

			// Marshal the route matrix, modify it, then attempt to unmarshal it
			buff, err := matrix.MarshalBinary()
			assert.NoError(t, err)

			buffSlice := buff[:3] // Only send the first 3 bytes so that the version read fails and throws an error

			_, err = matrix.ReadFrom(bytes.NewBuffer(buffSlice))
			assert.Error(t, err)
		})

		t.Run("Success", func(t *testing.T) {
			// Create and populate a route matrix
			matrix := getPopulatedRouteMatrix(false)

			// Marshal the route matrix so we can read it in
			buff, err := matrix.MarshalBinary()
			assert.NoError(t, err)

			// Read into a byte buffer
			_, err = matrix.ReadFrom(bytes.NewBuffer(buff))
			assert.NoError(t, err)
		})
	})
}

func TestRouteMatrix(t *testing.T) {
	t.Run("old optimize tests from core/core_test.go", func(t *testing.T) {
		analyze := func(t *testing.T, route_matrix *routing.RouteMatrix) {
			src := route_matrix.RelayIDs
			dest := route_matrix.RelayIDs

			numRelayPairs := 0
			numValidRelayPairs := 0
			numValidRelayPairsWithoutImprovement := 0

			buckets := make([]int, 11)

			for i := range src {
				for j := range dest {
					if j < i {
						numRelayPairs++
						abFlatIndex := routing.TriMatrixIndex(i, j)
						if len(route_matrix.Entries[abFlatIndex].RouteRTT) > 0 {
							numValidRelayPairs++
							improvement := route_matrix.Entries[abFlatIndex].DirectRTT - route_matrix.Entries[abFlatIndex].RouteRTT[0]
							if improvement > 0.0 {
								if improvement <= 5 {
									buckets[0]++
								} else if improvement <= 10 {
									buckets[1]++
								} else if improvement <= 15 {
									buckets[2]++
								} else if improvement <= 20 {
									buckets[3]++
								} else if improvement <= 25 {
									buckets[4]++
								} else if improvement <= 30 {
									buckets[5]++
								} else if improvement <= 35 {
									buckets[6]++
								} else if improvement <= 40 {
									buckets[7]++
								} else if improvement <= 45 {
									buckets[8]++
								} else if improvement <= 50 {
									buckets[9]++
								} else {
									buckets[10]++
								}
							} else {
								numValidRelayPairsWithoutImprovement++
							}
						}
					}
				}
			}

			assert.Equal(t, 43916, numValidRelayPairsWithoutImprovement, "optimizer is broken")

			expected := []int{2561, 8443, 6531, 4690, 3208, 2336, 1775, 1364, 1078, 749, 5159}

			assert.Equal(t, expected, buckets, "optimizer is broken")
		}

		t.Run("TestRouteMatrixSanity() - test using version 2 example data", func(t *testing.T) {
			var cmatrix routing.CostMatrix
			var rmatrix routing.RouteMatrix

			raw, err := ioutil.ReadFile("test_data/cost-for-sanity-check.bin")
			assert.Nil(t, err)

			err = cmatrix.UnmarshalBinary(raw)
			assert.Nil(t, err)

			err = cmatrix.Optimize(&rmatrix, 1.0)
			assert.Nil(t, err)

			src := rmatrix.RelayIDs
			dest := rmatrix.RelayIDs

			for i := range src {
				for j := range dest {
					if j < i {
						ijFlatIndex := routing.TriMatrixIndex(i, j)

						entries := rmatrix.Entries[ijFlatIndex]
						for k := 0; k < int(entries.NumRoutes); k++ {
							numRelays := entries.RouteNumRelays[k]
							firstRelay := entries.RouteRelays[k][0]
							lastRelay := entries.RouteRelays[k][numRelays-1]

							assert.Equal(t, src[firstRelay], dest[i], "invalid route entry #%d at (%d,%d), near relay %d (idx %d) != %d (idx %d)\n", k, i, j, src[firstRelay], firstRelay, dest[i], i)
							assert.Equal(t, src[lastRelay], dest[j], "invalid route entry #%d at (%d,%d), dest relay %d (idx %d) != %d (idx %d)\n", k, i, j, src[lastRelay], lastRelay, dest[j], j)
						}
					}
				}
			}
		})

		t.Run("TestRouteMatrix() - another test with different version 0 sample data", func(t *testing.T) {
			raw, err := ioutil.ReadFile("test_data/cost.bin")
			assert.Nil(t, err)
			assert.Equal(t, len(raw), 355188, "cost.bin should be 355188 bytes")

			var costMatrix routing.CostMatrix
			err = costMatrix.UnmarshalBinary(raw)
			assert.Nil(t, err)

			costMatrixData, err := costMatrix.MarshalBinary()
			assert.Nil(t, err)

			var readCostMatrix routing.CostMatrix
			err = readCostMatrix.UnmarshalBinary(costMatrixData)
			assert.Nil(t, err)

			var routeMatrix routing.RouteMatrix

			costMatrix.Optimize(&routeMatrix, 5)
			assert.NotNil(t, &routeMatrix)
			assert.Equal(t, costMatrix.RelayIDs, routeMatrix.RelayIDs, "relay id mismatch")
			assert.Equal(t, costMatrix.RelayAddresses, routeMatrix.RelayAddresses, "relay address mismatch")
			assert.Equal(t, costMatrix.RelayPublicKeys, routeMatrix.RelayPublicKeys, "relay public key mismatch")

			routeMatrixData, err := routeMatrix.MarshalBinary()
			assert.Nil(t, err)

			var readRouteMatrix routing.RouteMatrix
			err = readRouteMatrix.UnmarshalBinary(routeMatrixData)
			assert.Nil(t, err)

			assert.Equal(t, routeMatrix.RelayIDs, readRouteMatrix.RelayIDs, "relay id mismatch")

			assert.Len(t, readCostMatrix.RelayAddresses, len(costMatrix.RelayAddresses))
			for i, addr := range costMatrix.RelayAddresses {
				assert.Equal(t, string(addr), strings.Trim(string(readCostMatrix.RelayAddresses[i]), string([]byte{0x0})))
			}
			assert.Equal(t, routeMatrix.RelayPublicKeys, readRouteMatrix.RelayPublicKeys, "relay public key mismatch")
			assert.Equal(t, routeMatrix.DatacenterRelays, readRouteMatrix.DatacenterRelays, "datacenter relays mismatch")

			equal := true

			assert.Len(t, readRouteMatrix.Entries, len(routeMatrix.Entries))
			for i := 0; i < len(routeMatrix.Entries); i++ {

				if routeMatrix.Entries[i].DirectRTT != readRouteMatrix.Entries[i].DirectRTT {
					t.Errorf("DirectRTT mismatch: %d != %d\n", routeMatrix.Entries[i].DirectRTT, readRouteMatrix.Entries[i].DirectRTT)
					equal = false
					break
				}

				if routeMatrix.Entries[i].NumRoutes != readRouteMatrix.Entries[i].NumRoutes {
					t.Errorf("NumRoutes mismatch\n")
					equal = false
					break
				}

				for j := 0; j < int(routeMatrix.Entries[i].NumRoutes); j++ {

					if routeMatrix.Entries[i].RouteRTT[j] != readRouteMatrix.Entries[i].RouteRTT[j] {
						t.Errorf("RouteRTT mismatch\n")
						equal = false
						break
					}

					if routeMatrix.Entries[i].RouteNumRelays[j] != readRouteMatrix.Entries[i].RouteNumRelays[j] {
						t.Errorf("RouteNumRelays mismatch\n")
						equal = false
						break
					}

					for k := 0; k < int(routeMatrix.Entries[i].RouteNumRelays[j]); k++ {
						if routeMatrix.Entries[i].RouteRelays[j][k] != readRouteMatrix.Entries[i].RouteRelays[j][k] {
							t.Errorf("RouteRelayID mismatch\n")
							equal = false
							break
						}
					}
				}
			}

			assert.True(t, equal, "route matrix entries mismatch")
			analyze(t, &readRouteMatrix)
		})
	})

	costfile, err := os.Open("./test_data/cost.bin")
	assert.NoError(t, err)

	var costMatrix routing.CostMatrix
	_, err = costMatrix.ReadFrom(costfile)
	assert.NoError(t, err)

	var routeMatrix routing.RouteMatrix
	err = costMatrix.Optimize(&routeMatrix, 1)
	assert.NoError(t, err)

	t.Run("RelaysIn", func(t *testing.T) {
		t.Run("datacenter not found", func(t *testing.T) {
			actual := routeMatrix.GetDatacenterRelayIDs(routing.Datacenter{ID: 0})
			assert.Nil(t, actual)
		})

		t.Run("relay length is 0", func(t *testing.T) {
			routeMatrixCopy := routeMatrix
			routeMatrixCopy.DatacenterRelays[0] = []uint64{}

			actual := routeMatrixCopy.GetDatacenterRelayIDs(routing.Datacenter{ID: 0})
			assert.Empty(t, actual)
		})

		t.Run("error resolving at least one relay", func(t *testing.T) {
			routeMatrix := routing.RouteMatrix{
				RelayIndices:     map[uint64]int{0: 0},
				RelayAddresses:   [][]byte{[]byte("127.0.0.1:abcde")},
				RelayPublicKeys:  [][]byte{{0x58, 0xaf, 0x19, 0x5, 0xf7, 0xa8, 0xae, 0x73, 0xc6, 0xd3, 0xec, 0x85, 0x2f, 0xd8, 0x9b, 0x5a, 0xce, 0x0, 0x38, 0xca, 0x26, 0x39, 0xa4, 0x5d, 0x82, 0x3c, 0x71, 0xa8, 0x4, 0x11, 0xfb, 0x32}},
				DatacenterRelays: map[uint64][]uint64{0: {0, 1}},
			}

			actual := routeMatrix.GetDatacenterRelayIDs(routing.Datacenter{ID: 0})
			assert.NotNil(t, actual)
		})

		t.Run("datacenter with relays", func(t *testing.T) {
			expected := []uint64{
				3407334631,
				1447163127,
			}

			actual := routeMatrix.GetDatacenterRelayIDs(routing.Datacenter{ID: 69517923})
			assert.Equal(t, expected, actual)
		})
	})

	t.Run("GetRoutes", func(t *testing.T) {
		routeMatrixCopy := routeMatrix

		// Hack to insert relay session counts without regenerating a new route matrix
		numRelays := len(routeMatrixCopy.RelayIDs)
		for i := 0; i < numRelays; i++ {
			routeMatrixCopy.RelaySessionCounts[i] = uint32(i)
			routeMatrixCopy.RelayMaxSessionCounts[i] = 3000
		}

		// Have a relay be encumbered
		routeMatrixCopy.RelaySessionCounts[3] = 3000

		t.Run("empty near/dest sets", func(t *testing.T) {
			near := []routing.NearRelayData{}
			dest := []uint64{}

			actual, err := routeMatrixCopy.GetAcceptableRoutes(near, dest, 0, 0)
			assert.EqualError(t, err, "no routes in route matrix")
			assert.Equal(t, 0, len(actual))
		})

		t.Run("relays not found", func(t *testing.T) {
			near := []routing.NearRelayData{{ID: 1}}
			dest := []uint64{2}

			actual, err := routeMatrixCopy.GetAcceptableRoutes(near, dest, 0, 0)
			assert.EqualError(t, err, "no routes in route matrix")
			assert.Equal(t, 0, len(actual))
		})

		t.Run("one relay found", func(t *testing.T) {
			near := []routing.NearRelayData{{ID: 1}}
			dest := []uint64{1500948990}

			actual, err := routeMatrixCopy.GetAcceptableRoutes(near, dest, 0, 0)
			assert.EqualError(t, err, "no routes in route matrix")
			assert.Equal(t, 0, len(actual))
		})

		t.Run("success", func(t *testing.T) {
			near := []routing.NearRelayData{{ID: 2836356269}}
			dest := []uint64{3263834878, 1500948990}

			expected := []routing.Route{
				{
					RelayIDs: []uint64{2836356269, 1370686037, 2923051732, 1884974764, 3263834878},
					Stats:    routing.Stats{RTT: 182},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 2641807504, 3263834878},
					Stats:    routing.Stats{RTT: 182},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 1348914502, 1884974764, 3263834878},
					Stats:    routing.Stats{RTT: 182},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 2576485547, 1835585494, 3263834878},
					Stats:    routing.Stats{RTT: 183},
				},
				{
					RelayIDs: []uint64{2836356269, 1348914502, 1884974764, 3263834878},
					Stats:    routing.Stats{RTT: 183},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 2663193268, 2504465311, 3263834878},
					Stats:    routing.Stats{RTT: 184},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 427962386, 2504465311, 3263834878},
					Stats:    routing.Stats{RTT: 184},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 4058587524, 1350942731, 3263834878},
					Stats:    routing.Stats{RTT: 184},
				},
				{
					RelayIDs: []uint64{2836356269, 1500948990},
					Stats:    routing.Stats{RTT: 311},
				},
			}

			actual, err := routeMatrixCopy.GetAcceptableRoutes(near, dest, 0, 500)
			assert.NoError(t, err)
			assert.Equal(t, len(expected), len(actual))

			for routeidx, route := range expected {
				assert.Equal(t, len(expected[routeidx].RelayIDs), len(route.RelayIDs))

				for relayidx := range route.RelayIDs {
					assert.Equal(t, expected[routeidx].RelayIDs[relayidx], actual[routeidx].RelayIDs[relayidx])
				}

				assert.Equal(t, expected[routeidx].Stats, actual[routeidx].Stats)
			}
		})

		t.Run("only best RTT routes", func(t *testing.T) {
			near := []routing.NearRelayData{{ID: 2836356269}}
			dest := []uint64{3263834878, 1500948990}
			expected := []routing.Route{
				{
					RelayIDs: []uint64{2836356269, 1370686037, 2923051732, 1884974764, 3263834878},
					Stats:    routing.Stats{RTT: 182},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 2641807504, 3263834878},
					Stats:    routing.Stats{RTT: 182},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 1348914502, 1884974764, 3263834878},
					Stats:    routing.Stats{RTT: 182},
				},
			}

			actual, err := routeMatrixCopy.GetAcceptableRoutes(near, dest, 0, 0)
			assert.NoError(t, err)
			assert.Equal(t, len(expected), len(actual))

			for routeidx, route := range expected {
				assert.Equal(t, len(expected[routeidx].RelayIDs), len(route.RelayIDs))

				for relayidx := range route.RelayIDs {
					assert.Equal(t, expected[routeidx].RelayIDs[relayidx], actual[routeidx].RelayIDs[relayidx])
				}

				assert.Equal(t, expected[routeidx].Stats, actual[routeidx].Stats)
			}
		})

		t.Run("acceptable routes", func(t *testing.T) {
			near := []routing.NearRelayData{{ID: 2836356269}}
			dest := []uint64{3263834878, 1500948990}
			expected := []routing.Route{
				{
					RelayIDs: []uint64{2836356269, 1370686037, 2923051732, 1884974764, 3263834878},
					Stats:    routing.Stats{RTT: 182},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 2641807504, 3263834878},
					Stats:    routing.Stats{RTT: 182},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 1348914502, 1884974764, 3263834878},
					Stats:    routing.Stats{RTT: 182},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 2576485547, 1835585494, 3263834878},
					Stats:    routing.Stats{RTT: 183},
				},
				{
					RelayIDs: []uint64{2836356269, 1348914502, 1884974764, 3263834878},
					Stats:    routing.Stats{RTT: 183},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 2663193268, 2504465311, 3263834878},
					Stats:    routing.Stats{RTT: 184},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 427962386, 2504465311, 3263834878},
					Stats:    routing.Stats{RTT: 184},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 4058587524, 1350942731, 3263834878},
					Stats:    routing.Stats{RTT: 184},
				},
			}

			actual, err := routeMatrixCopy.GetAcceptableRoutes(near, dest, 0, 5)
			assert.NoError(t, err)
			assert.Equal(t, len(expected), len(actual))

			for routeidx, route := range expected {
				assert.Equal(t, len(expected[routeidx].RelayIDs), len(route.RelayIDs))

				for relayidx := range route.RelayIDs {
					assert.Equal(t, expected[routeidx].RelayIDs[relayidx], actual[routeidx].RelayIDs[relayidx])
				}

				assert.Equal(t, expected[routeidx].Stats, actual[routeidx].Stats)
			}
		})

		t.Run("contains route and route is still acceptable", func(t *testing.T) {
			near := []routing.NearRelayData{{ID: 2836356269}}
			dest := []uint64{3263834878, 1500948990}

			route := routing.Route{
				RelayIDs: []uint64{2836356269, 1370686037, 2663193268, 2504465311, 3263834878},
				Stats:    routing.Stats{RTT: 184},
			}
			routeHash := route.Hash64()

			expected := []routing.Route{route}

			actual, err := routeMatrixCopy.GetAcceptableRoutes(near, dest, routeHash, 5)
			assert.NoError(t, err)
			assert.Equal(t, len(expected), len(actual))

			for routeidx, route := range expected {
				assert.Equal(t, len(expected[routeidx].RelayIDs), len(route.RelayIDs))

				for relayidx := range route.RelayIDs {
					assert.Equal(t, expected[routeidx].RelayIDs[relayidx], actual[routeidx].RelayIDs[relayidx])
				}

				assert.Equal(t, expected[routeidx].Stats, actual[routeidx].Stats)
			}
		})

		t.Run("contains route but route is no longer acceptable", func(t *testing.T) {
			near := []routing.NearRelayData{{ID: 2836356269}}
			dest := []uint64{3263834878, 1500948990}

			route := routing.Route{
				RelayIDs: []uint64{2836356269, 1370686037, 2663193268, 2504465311, 3263834878},
				Stats:    routing.Stats{RTT: 184},
			}
			routeHash := route.Hash64()

			expected := []routing.Route{
				{
					RelayIDs: []uint64{2836356269, 1370686037, 2923051732, 1884974764, 3263834878},
					Stats:    routing.Stats{RTT: 182},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 2641807504, 3263834878},
					Stats:    routing.Stats{RTT: 182},
				},
				{
					RelayIDs: []uint64{2836356269, 1370686037, 1348914502, 1884974764, 3263834878},
					Stats:    routing.Stats{RTT: 182},
				},
			}

			actual, err := routeMatrixCopy.GetAcceptableRoutes(near, dest, routeHash, 0)
			assert.NoError(t, err)
			assert.Equal(t, len(expected), len(actual))

			for routeidx, route := range expected {
				assert.Equal(t, len(expected[routeidx].RelayIDs), len(route.RelayIDs))

				for relayidx := range route.RelayIDs {
					assert.Equal(t, expected[routeidx].RelayIDs[relayidx], actual[routeidx].RelayIDs[relayidx])
				}

				assert.Equal(t, expected[routeidx].Stats, actual[routeidx].Stats)
			}
		})
		// {
		// 	"unencumbered routes",
		// 	[]routing.Relay{{ID: 2836356269}},
		// 	[]routing.Relay{{ID: 3263834878}, {ID: 1500948990}},
		// 	[]routing.Route{
		// 		{
		// 			Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2923051732}, {ID: 1884974764}, {ID: 3263834878}},
		// 			Stats:  routing.Stats{RTT: 182},
		// 		},
		// 		{
		// 			Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2641807504}, {ID: 3263834878}},
		// 			Stats:  routing.Stats{RTT: 182},
		// 		},
		// 		{
		// 			Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 1348914502}, {ID: 1884974764}, {ID: 3263834878}},
		// 			Stats:  routing.Stats{RTT: 182},
		// 		},
		// 		{
		// 			Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2576485547}, {ID: 1835585494}, {ID: 3263834878}},
		// 			Stats:  routing.Stats{RTT: 183},
		// 		},
		// 		{
		// 			Relays: []routing.Relay{{ID: 2836356269}, {ID: 1348914502}, {ID: 1884974764}, {ID: 3263834878}},
		// 			Stats:  routing.Stats{RTT: 183},
		// 		},
		// 		{
		// 			Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2663193268}, {ID: 2504465311}, {ID: 3263834878}},
		// 			Stats:  routing.Stats{RTT: 184},
		// 		},
		// 		{
		// 			Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 427962386}, {ID: 2504465311}, {ID: 3263834878}},
		// 			Stats:  routing.Stats{RTT: 184},
		// 		},
		// 		{
		// 			Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 4058587524}, {ID: 1350942731}, {ID: 3263834878}},
		// 			Stats:  routing.Stats{RTT: 184},
		// 		},
		// 	},
		// 	nil,
		// 	[]routing.SelectorFunc{
		// 		routing.SelectUnencumberedRoutes(0.8),
		// 	},
		// },
		// {
		// 	"routes by random dest relay",
		// 	[]routing.Relay{{ID: 2836356269}},
		// 	[]routing.Relay{{ID: 3263834878}, {ID: 1500948990}},
		// 	[]routing.Route{
		// 		{
		// 			Relays: []routing.Relay{{ID: 2836356269}, {ID: 1500948990}},
		// 			Stats:  routing.Stats{RTT: 311},
		// 		},
		// 	},
		// 	nil,
		// 	[]routing.SelectorFunc{
		// 		routing.SelectRoutesByRandomDestRelay(rand.NewSource(0)),
		// 	},
		// },
		// {
		// 	"random route",
		// 	[]routing.Relay{{ID: 2836356269}},
		// 	[]routing.Relay{{ID: 3263834878}, {ID: 1500948990}},
		// 	[]routing.Route{
		// 		{
		// 			Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2923051732}, {ID: 1884974764}},
		// 			Stats:  routing.Stats{RTT: 182},
		// 		},
		// 	},
		// 	nil,
		// 	[]routing.SelectorFunc{
		// 		routing.SelectRandomRoute(rand.NewSource(0)),
		// 	},
		// },
	})
}

func BenchmarkGetRoutes(b *testing.B) {
	costfile, _ := os.Open("./test_data/cost.bin")

	var costMatrix routing.CostMatrix
	costMatrix.ReadFrom(costfile)

	var routeMatrix routing.RouteMatrix
	costMatrix.Optimize(&routeMatrix, 1)

	from := []routing.NearRelayData{{ID: 2836356269}}
	to := []uint64{3263834878, 1500948990}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		routeMatrix.GetAcceptableRoutes(from, to, 0, 500)
	}
}

// Benchmarks fetching all relays in the given datacenter for the first data center in the file
func BenchmarkRelaysIn(b *testing.B) {
	costfile, _ := os.Open("./test_data/cost-for-sanity-check.bin") // This file actually has datacenters in it

	var costMatrix routing.CostMatrix
	costMatrix.ReadFrom(costfile)

	var routeMatrix routing.RouteMatrix
	costMatrix.Optimize(&routeMatrix, 1)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		routeMatrix.GetDatacenterRelayIDs(routing.Datacenter{ID: routeMatrix.DatacenterIDs[0], Name: routeMatrix.DatacenterNames[0]})
	}
}
