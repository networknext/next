package routing_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func getPopulatedCostMatrix(malformed bool) *routing.CostMatrix {
	var matrix routing.CostMatrix

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

	matrix.RTT = make([]int32, 1)
	matrix.RTT[0] = 7

	matrix.RelaySellers = []routing.Seller{
		{ID: "1234", Name: "Seller One", IngressPriceNibblinsPerGB: 10, EgressPriceNibblinsPerGB: 20},
		{ID: "5678", Name: "Seller Two", IngressPriceNibblinsPerGB: 30, EgressPriceNibblinsPerGB: 40},
	}

	matrix.RelaySessionCounts = []uint32{100, 200}
	matrix.RelayMaxSessionCounts = []uint32{3000, 6000}

	matrix.RelayLatitude = []float64{1.0, 2.0}
	matrix.RelayLongitude = []float64{3.0, 4.0}

	return &matrix
}

func costMatrixUnmarshalAssertionsVer5(t *testing.T, matrix *routing.CostMatrix, numRelays, numDatacenters int, relayIDs, datacenters []uint64, relayAddrs []string, datacenterRelays [][]uint64, publicKeys [][]byte, rtts []int32, relayNames []string, datacenterIDs []uint64, datacenterNames []string, sellers []routing.Seller, sessionCounts []uint32, maxSessionCounts []uint32) {
	assert.Len(t, matrix.RelayIDs, numRelays)
	assert.Len(t, matrix.RelayAddresses, numRelays)
	assert.Len(t, matrix.RelayPublicKeys, numRelays)
	assert.Len(t, matrix.DatacenterRelays, numDatacenters)
	assert.Len(t, matrix.RTT, len(rtts))

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

	for i, rtt := range rtts {
		assert.Equal(t, matrix.RTT[i], rtt)
	}

	assert.Len(t, matrix.RelayNames, len(relayNames))
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
		assert.Equal(t, seller.ID, matrix.RelaySellers[i].ID)
		assert.Equal(t, seller.Name, matrix.RelaySellers[i].Name)
		assert.Equal(t, seller.IngressPriceNibblinsPerGB, matrix.RelaySellers[i].IngressPriceNibblinsPerGB)
		assert.Equal(t, seller.EgressPriceNibblinsPerGB, matrix.RelaySellers[i].EgressPriceNibblinsPerGB)
	}

	assert.Equal(t, sessionCounts, matrix.RelaySessionCounts)
	assert.Equal(t, maxSessionCounts, matrix.RelayMaxSessionCounts)
}

func costMatrixUnmarshalAssertionsVer6(t *testing.T, matrix *routing.CostMatrix, latitudes []float64, longitudes []float64) {
	assert.Equal(t, latitudes, matrix.RelayLatitude)
	assert.Equal(t, longitudes, matrix.RelayLongitude)
}

type costMatrixData struct {
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
	rtts             []int32
	sellers          []routing.Seller
	sessionCounts    []uint32
	maxSessionCounts []uint32
}

func getCostMatrixDataV5() costMatrixData {
	// version 0 stuff
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
	rtts := make([]int32, routing.TriMatrixLength(numRelays))
	for i := range rtts {
		rtts[i] = int32(rand.Int())
	}

	// version 1 stuff
	relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}
	// version 2 stuff
	// resusing datacenters for the ID array
	datacenterNames := []string{"a datacenter", "another datacenter", "third", "fourth", "fifth"}

	// version 4 stuff
	sellers := []routing.Seller{
		{ID: "id0", Name: "name0", IngressPriceNibblinsPerGB: 1, EgressPriceNibblinsPerGB: 2},
		{ID: "id1", Name: "name1", IngressPriceNibblinsPerGB: 3, EgressPriceNibblinsPerGB: 4},
		{ID: "id2", Name: "name2", IngressPriceNibblinsPerGB: 5, EgressPriceNibblinsPerGB: 6},
		{ID: "id3", Name: "name3", IngressPriceNibblinsPerGB: 7, EgressPriceNibblinsPerGB: 8},
		{ID: "id4", Name: "name4", IngressPriceNibblinsPerGB: 9, EgressPriceNibblinsPerGB: 10},
	}

	// version 5 stuff
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
	buffSize += sizeofRTTs(rtts)
	buffSize += sizeofSellers(sellers)
	buffSize += sizeofSessionCounts(sessionCounts)
	buffSize += sizeofMaxSessionCounts(maxSessionCounts)

	buff := make([]byte, buffSize)

	offset := 0
	putVersionNumber(buff, &offset, 5)
	putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
	putRelayNames(buff, &offset, relayNames)                        // version 1
	putDatacenterStuff(buff, &offset, datacenters, datacenterNames) // version 2
	putRelayAddresses(buff, &offset, relayAddrs)
	putRelayPublicKeys(buff, &offset, publicKeys)
	putDatacenters(buff, &offset, datacenters, datacenterRelays)
	putRTTs(buff, &offset, rtts)
	putSellers(buff, &offset, sellers)
	putSessionCounts(buff, &offset, sessionCounts)
	putMaxSessionCounts(buff, &offset, maxSessionCounts)

	return costMatrixData{
		buff:             buff,
		numRelays:        numRelays,
		relayIDs:         relayIDs,
		relayNames:       relayNames,
		numDatacenters:   numDatacenters,
		datacenterIDs:    datacenters,
		datacenterNames:  datacenterNames,
		relayAddrs:       relayAddrs,
		datacenterRelays: datacenterRelays,
		publicKeys:       publicKeys,
		rtts:             rtts,
		sellers:          sellers,
		sessionCounts:    sessionCounts,
		maxSessionCounts: maxSessionCounts,
	}
}

func getCostMatrixDataV6() costMatrixData {
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
	rtts := make([]int32, routing.TriMatrixLength(numRelays))
	for i := range rtts {
		rtts[i] = int32(rand.Int())
	}

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
	buffSize += sizeofRTTs(rtts)
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
	putRTTs(buff, &offset, rtts)
	putSellers(buff, &offset, sellers)
	putSessionCounts(buff, &offset, sessionCounts)
	putMaxSessionCounts(buff, &offset, maxSessionCounts)

	return costMatrixData{
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
		rtts:             rtts,
		sellers:          sellers,
		sessionCounts:    sessionCounts,
		maxSessionCounts: maxSessionCounts,
	}
}

func TestCostMatrixUnmarshalBinaryV5(t *testing.T) {
	data := getCostMatrixDataV5()

	t.Run("version of incoming bin data too high", func(t *testing.T) {
		buff := make([]byte, 4)
		offset := 0
		putVersionNumber(buff, &offset, 100)
		var matrix routing.CostMatrix

		err := matrix.UnmarshalBinary(buff)

		assert.EqualError(t, err, "unknown cost matrix version 100")
	})

	t.Run("Invalid version read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at version number")
	})

	t.Run("Invalid relay count read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays")
	})

	t.Run("Invalid relay id read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids - ver >= v3")
	})

	t.Run("Invalid relay name read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at relay names")
	})

	t.Run("Invalid datacenter count read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter count")
	})

	t.Run("Invalid datacenter id read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 8 + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter ids - ver >= v3")
	})

	t.Run("Invalid datacenter name read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter names")
	})

	t.Run("Invalid relay address read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at relay addresses - ver >= v3")
	})

	t.Run("Invalid relay public key read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at relay public keys - ver >= v3")
	})

	t.Run("Invalid datacenter count read second time", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at number of datacenters (second time)")
	})

	t.Run("Invalid datacenter id read second time", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter id - ver >= v3")
	})

	t.Run("Invalid datacenter relay count read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays in datacenter")
	})

	t.Run("Invalid datacenter relay id read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 8 + 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids for datacenter - ver >= v3")
	})

	t.Run("Invalid rtt read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at rtt")
	})

	t.Run("Invalid seller id read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 4 + len(data.sellers[0].ID) + sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read on relay seller ID")
	})

	t.Run("Invalid seller name read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 4 + len(data.sellers[0].Name) + 4 + len(data.sellers[0].ID) + sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read on relay seller name")
	})

	t.Run("Invalid seller ingress price read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 8 + 4 + len(data.sellers[0].Name) + 4 + len(data.sellers[0].ID) + sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read on relay seller ingress price")
	})

	t.Run("Invalid seller egress price read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 8 + 8 + 4 + len(data.sellers[0].Name) + 4 + len(data.sellers[0].ID) + sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read on relay seller egress price")
	})

	t.Run("Invalid session count read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 4 + sizeofSellers(data.sellers) + sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read on relay session count")
	})

	t.Run("Invalid max session count read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 4 + sizeofSessionCounts(data.sessionCounts) + sizeofSellers(data.sellers) + sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read on relay max session count")
	})

	t.Run("Success", func(t *testing.T) {
		var matrix routing.CostMatrix
		err := matrix.UnmarshalBinary(data.buff)
		assert.Nil(t, err)
		costMatrixUnmarshalAssertionsVer5(t, &matrix, data.numRelays, data.numDatacenters, data.relayIDs, data.datacenterIDs, data.relayAddrs, data.datacenterRelays, data.publicKeys, data.rtts, data.relayNames, data.datacenterIDs, data.datacenterNames, data.sellers, data.sessionCounts, data.maxSessionCounts)
	})
}

func TestCostMatrixUnmarshalBinaryV6(t *testing.T) {
	data := getCostMatrixDataV6()

	t.Run("version of incoming bin data too high", func(t *testing.T) {
		buff := make([]byte, 4)
		offset := 0
		putVersionNumber(buff, &offset, 100)
		var matrix routing.CostMatrix

		err := matrix.UnmarshalBinary(buff)

		assert.EqualError(t, err, "unknown cost matrix version 100")
	})

	t.Run("Invalid version read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at version number")
	})

	t.Run("Invalid relay count read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays")
	})

	t.Run("Invalid relay id read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids - ver >= v3")
	})

	t.Run("Invalid relay name read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at relay names")
	})

	t.Run("Invalid datacenter count read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter count")
	})

	t.Run("Invalid datacenter id read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 8 + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter ids - ver >= v3")
	})

	t.Run("Invalid datacenter name read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterCount() + sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter names")
	})

	t.Run("Invalid relay address read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at relay addresses - ver >= v3")
	})

	t.Run("Invalid relay latitude read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at relay latitude")
	})

	t.Run("Invalid relay longitude read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at relay longitude")
	})

	t.Run("Invalid relay public key read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at relay public keys - ver >= v3")
	})

	t.Run("Invalid datacenter count read second time", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at number of datacenters (second time)")
	})

	t.Run("Invalid datacenter id read second time", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter id - ver >= v3")
	})

	t.Run("Invalid datacenter relay count read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays in datacenter")
	})

	t.Run("Invalid datacenter relay id read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 8 + 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids for datacenter - ver >= v3")
	})

	t.Run("Invalid rtt read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read at rtt")
	})

	t.Run("Invalid seller id read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 4 + len(data.sellers[0].ID) + sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read on relay seller ID")
	})

	t.Run("Invalid seller name read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 4 + len(data.sellers[0].Name) + 4 + len(data.sellers[0].ID) + sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read on relay seller name")
	})

	t.Run("Invalid seller ingress price read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 8 + 4 + len(data.sellers[0].Name) + 4 + len(data.sellers[0].ID) + sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read on relay seller ingress price")
	})

	t.Run("Invalid seller egress price read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 8 + 8 + 4 + len(data.sellers[0].Name) + 4 + len(data.sellers[0].ID) + sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read on relay seller egress price")
	})

	t.Run("Invalid session count read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 4 + sizeofSellers(data.sellers) + sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read on relay session count")
	})

	t.Run("Invalid max session count read", func(t *testing.T) {
		var matrix routing.CostMatrix
		offset := 4 + sizeofSessionCounts(data.sessionCounts) + sizeofSellers(data.sellers) + sizeofRTTs(data.rtts) + sizeofRelayIDs64(data.relayIDs) + sizeofRelaysInDatacenterCount(data.datacenterIDs) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDataCenterCount2() +
			sizeofRelayPublicKeys(data.publicKeys) + sizeofRelayLongitudes(data.relayLongitudes) + sizeofRelayLatitudes(data.relayLatitudes) + sizeofRelayAddress(data.relayAddrs) + sizeofDatacenterNames(data.datacenterNames) + sizeofDatacenterIDs64(data.datacenterIDs) + sizeofDatacenterCount() +
			sizeofRelayNames(data.relayNames) + sizeofRelayIDs64(data.relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
		err := matrix.UnmarshalBinary(data.buff[:offset])
		assert.EqualError(t, err, "[CostMatrix] invalid read on relay max session count")
	})

	t.Run("Success", func(t *testing.T) {
		var matrix routing.CostMatrix
		err := matrix.UnmarshalBinary(data.buff)
		assert.Nil(t, err)
		costMatrixUnmarshalAssertionsVer5(t, &matrix, data.numRelays, data.numDatacenters, data.relayIDs, data.datacenterIDs, data.relayAddrs, data.datacenterRelays, data.publicKeys, data.rtts, data.relayNames, data.datacenterIDs, data.datacenterNames, data.sellers, data.sessionCounts, data.maxSessionCounts)
		costMatrixUnmarshalAssertionsVer6(t, &matrix, data.relayLatitudes, data.relayLongitudes)
	})
}

func TestCostMatrixMarshalBinary(t *testing.T) {
	t.Run("MarshalBinary -> UnmarshalBinary equality", func(t *testing.T) {
		matrix := getPopulatedCostMatrix(false)
		var other routing.CostMatrix

		bin, err := matrix.MarshalBinary()
		assert.NoError(t, err)

		// essentialy this asserts the result of MarshalBinary(),
		// if Unmarshal tests pass then the binary data from Marshal
		// is valid if unmarshaling equals the original
		err = other.UnmarshalBinary(bin)
		assert.NoError(t, err)

		assert.Equal(t, matrix, &other)
	})

	t.Run("Relay ID and name buffers different sizes", func(t *testing.T) {
		var matrix routing.CostMatrix

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
		var matrix routing.CostMatrix

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

func TestCostMatrixServeHTTP(t *testing.T) {
	t.Run("Successful Serve", func(t *testing.T) {
		// Create and populate a cost matrix
		matrix := getPopulatedCostMatrix(false)
		err := matrix.WriteResponseData()
		assert.NoError(t, err)

		// Create a dummy http request to test ServeHTTP
		recorder := httptest.NewRecorder()
		request, err := http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)

		matrix.ServeHTTP(recorder, request)

		// Get the response
		response := recorder.Result()

		// Read the response body
		body, err := ioutil.ReadAll(response.Body)
		assert.NoError(t, err)
		response.Body.Close()

		// Create a new matrix to store the response
		var receivedMatrix routing.CostMatrix
		err = receivedMatrix.UnmarshalBinary(body)
		assert.NoError(t, err)

		// Create a new expected matrix so that the response buffer is empty
		var expected routing.CostMatrix
		err = expected.UnmarshalBinary(matrix.GetResponseData())
		assert.NoError(t, err)

		// Validate the response
		assert.Equal(t, "application/octet-stream", response.Header.Get("Content-Type"))
		assert.Equal(t, &expected, &receivedMatrix)
	})
}

func TestCostMatrixWriteTo(t *testing.T) {
	t.Run("Error during MarshalBinary()", func(t *testing.T) {
		// Create and populate a malformed cost matrix
		matrix := getPopulatedCostMatrix(true)

		var buff bytes.Buffer
		_, err := matrix.WriteTo(&buff)
		assert.EqualError(t, err, fmt.Sprintf("length of Relay IDs not equal to length of Relay Names: %v != %v", len(matrix.RelayIDs), len(matrix.RelayNames)))
	})

	t.Run("Error during write", func(t *testing.T) {
		// Create and populate a cost matrix
		matrix := getPopulatedCostMatrix(false)

		var buff ErrorBuffer
		_, err := matrix.WriteTo(&buff)
		assert.Error(t, err)
	})

	t.Run("Success", func(t *testing.T) {
		// Create and populate a cost matrix
		matrix := getPopulatedCostMatrix(false)

		var buff bytes.Buffer
		_, err := matrix.WriteTo(&buff)
		assert.NoError(t, err)
	})
}

func TestCostMatrixReadFrom(t *testing.T) {
	t.Run("Error during read", func(t *testing.T) {
		// Create and populate a cost matrix
		matrix := getPopulatedCostMatrix(false)

		// Try to read into the ErrorBuffer
		var buff ErrorBuffer
		_, err := matrix.ReadFrom(&buff)
		assert.Error(t, err)
	})

	t.Run("Error during UnmarshalBinary()", func(t *testing.T) {
		// Create and populate a cost matrix
		matrix := getPopulatedCostMatrix(false)

		// Marshal the cost matrix, modify it, then attempt to unmarshal it
		buff, err := matrix.MarshalBinary()
		assert.NoError(t, err)

		buffSlice := buff[:3] // Only send the first 3 bytes so that the version read fails and throws an error

		_, err = matrix.ReadFrom(bytes.NewBuffer(buffSlice))
		assert.Error(t, err)
	})

	t.Run("Success", func(t *testing.T) {
		// Create and populate a cost matrix
		matrix := getPopulatedCostMatrix(false)

		// Marshal the cost matrix so we can read it in
		buff, err := matrix.MarshalBinary()
		assert.NoError(t, err)

		// Read into a byte buffer
		_, err = matrix.ReadFrom(bytes.NewBuffer(buff))
		assert.NoError(t, err)
	})
}

func BenchmarkOptimize(b *testing.B) {
	costfile, _ := os.Open("./test_data/cost.bin")

	var costMatrix routing.CostMatrix
	costMatrix.ReadFrom(costfile)

	var routeMatrix routing.RouteMatrix

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		costMatrix.Optimize(&routeMatrix, 1)
	}
}
