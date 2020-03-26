package routing_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

// A buffer type that implements io.Write and io.Read but always returns an error for testing
type ErrorBuffer struct {
}

func (*ErrorBuffer) Write(p []byte) (int, error) {
	return 0, errors.New("descriptive error")
}

func (*ErrorBuffer) Read(p []byte) (int, error) {
	return 0, errors.New("descriptive error")
}

func getPopulatedCostMatrix(malformed bool) *routing.CostMatrix {
	var matrix routing.CostMatrix

	matrix.RelayIndicies = make(map[uint64]int)
	matrix.RelayIndicies[123] = 0
	matrix.RelayIndicies[456] = 1

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
	matrix.RelayPublicKeys[0] = RandomPublicKey()
	matrix.RelayPublicKeys[1] = RandomPublicKey()

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
		{ID: "1234", Name: "Seller One", IngressPriceCents: 10, EgressPriceCents: 20},
		{ID: "5678", Name: "Seller Two", IngressPriceCents: 30, EgressPriceCents: 40},
	}

	return &matrix
}

func getPopulatedRouteMatrix(malformed bool) *routing.RouteMatrix {
	var matrix routing.RouteMatrix

	matrix.RelayIndicies = make(map[uint64]int)
	matrix.RelayIndicies[123] = 0
	matrix.RelayIndicies[456] = 1

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
	matrix.RelayPublicKeys[0] = RandomPublicKey()
	matrix.RelayPublicKeys[1] = RandomPublicKey()

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
		routing.RouteMatrixEntry{
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

	return &matrix
}

func addrsToIDs(addrs []string) []uint64 {
	retval := make([]uint64, len(addrs))
	for i, addr := range addrs {
		retval[i] = uint64(crypto.HashID(addr))
	}
	return retval
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

func putInt32s(buff []byte, offset *int, nums ...int32) {
	for _, num := range nums {
		putUint32s(buff, offset, uint32(num))
	}
}

func putUint32s(buff []byte, offset *int, nums ...uint32) {
	for _, num := range nums {
		binary.LittleEndian.PutUint32(buff[*offset:], num)
		*offset += 4
	}
}

func putUint64s(buff []byte, offset *int, nums ...uint64) {
	for _, num := range nums {
		binary.LittleEndian.PutUint64(buff[*offset:], num)
		*offset += 8
	}
}

func putStrings(buff []byte, offset *int, strings ...string) {
	for _, str := range strings {
		putUint32s(buff, offset, uint32(len(str)))
		copy(buff[*offset:], str)
		*offset += len(str)
	}
}

func putBytes(buff []byte, offset *int, bytes ...[]byte) {
	for _, arr := range bytes {
		copy(buff[*offset:], arr)
		*offset += len(arr)
	}
}

func putBytesOld(buff []byte, offset *int, bytes ...[]byte) {
	for _, arr := range bytes {
		putUint32s(buff, offset, uint32(len(arr)))
		putBytes(buff, offset, arr)
	}
}

func putVersionNumber(buff []byte, offset *int, version uint32) {
	putUint32s(buff, offset, version)
}

func putRelayIDs(buff []byte, offset *int, ids []uint64) {
	putUint32s(buff, offset, uint32(len(ids)))
	putUint64s(buff, offset, ids...)
}

func putRelayIDsOld(buff []byte, offset *int, ids []uint64) {
	putUint32s(buff, offset, uint32(len(ids)))

	for _, id := range ids {
		putUint32s(buff, offset, uint32(id))
	}
}

func putRelayNames(buff []byte, offset *int, names []string) {
	putStrings(buff, offset, names...)
}

func putDatacenterStuff(buff []byte, offset *int, ids []uint64, names []string) {
	putUint32s(buff, offset, uint32(len(ids)))

	for i := 0; i < len(ids); i++ {
		id, name := ids[i], names[i]
		putUint64s(buff, offset, id)
		putStrings(buff, offset, name)
	}
}

func putDatacenterStuffOld(buff []byte, offset *int, ids []uint64, names []string) {
	putUint32s(buff, offset, uint32(len(ids)))

	for i := 0; i < len(ids); i++ {
		id, name := ids[i], names[i]
		putUint32s(buff, offset, uint32(id))
		putStrings(buff, offset, name)
	}
}

func putRelayAddresses(buff []byte, offset *int, addrs []string) {
	for _, addr := range addrs {
		tmp := make([]byte, routing.MaxRelayAddressLength)
		copy(tmp, addr)
		putBytes(buff, offset, tmp)
	}
}

func putRelayAddressesOld(buff []byte, offset *int, addrs []string) {
	for _, addr := range addrs {
		putBytesOld(buff, offset, []byte(addr))
	}
}

func putRelayPublicKeys(buff []byte, offset *int, pks [][]byte) {
	putBytes(buff, offset, pks...)
}

func putRelayPublicKeysOld(buff []byte, offset *int, pks [][]byte) {
	putBytesOld(buff, offset, pks...)
}

func putDatacenters(buff []byte, offset *int, datacenterIDs []uint64, relayIDs [][]uint64) {
	putUint32s(buff, offset, uint32(len(datacenterIDs)))

	for i, dcID := range datacenterIDs {
		putUint64s(buff, offset, dcID)
		putUint32s(buff, offset, uint32(len(relayIDs[i])))
		putUint64s(buff, offset, relayIDs[i]...)
	}
}

func putDatacentersOld(buff []byte, offset *int, datacenterIDs []uint64, relayIDs [][]uint64) {
	putUint32s(buff, offset, uint32(len(datacenterIDs)))

	for i, dcID := range datacenterIDs {
		putUint32s(buff, offset, uint32(dcID))
		putUint32s(buff, offset, uint32(len(relayIDs[i])))
		for _, rID := range relayIDs[i] {
			putUint32s(buff, offset, uint32(rID))
		}
	}
}

func putRTTs(buff []byte, offset *int, rtts []int32) {
	putInt32s(buff, offset, rtts...)
}

func putSellers(buff []byte, offset *int, sellers []routing.Seller) {
	for _, seller := range sellers {
		putStrings(buff, offset, seller.ID, seller.Name)
		putUint64s(buff, offset, seller.IngressPriceCents)
		putUint64s(buff, offset, seller.EgressPriceCents)
	}
}

func putEntries(buff []byte, offset *int, entries []routing.RouteMatrixEntry) {
	for _, entry := range entries {
		putInt32s(buff, offset, entry.DirectRTT)
		putInt32s(buff, offset, entry.NumRoutes)

		for i := 0; i < int(entry.NumRoutes); i++ {
			putInt32s(buff, offset, entry.RouteRTT[i])
			putInt32s(buff, offset, entry.RouteNumRelays[i])
			putUint64s(buff, offset, entry.RouteRelays[i][:]...)
		}
	}
}

func putEntriesOld(buff []byte, offset *int, entries []routing.RouteMatrixEntry) {
	for _, entry := range entries {
		putInt32s(buff, offset, entry.DirectRTT)
		putInt32s(buff, offset, entry.NumRoutes)

		for i := 0; i < int(entry.NumRoutes); i++ {
			putInt32s(buff, offset, entry.RouteRTT[i])
			putInt32s(buff, offset, entry.RouteNumRelays[i])
			for _, id := range entry.RouteRelays[i] {
				putUint32s(buff, offset, uint32(id))
			}
		}
	}
}

func sizeofVersionNumber() int {
	return 4
}

func sizeofRelayCount() int {
	return 4
}

func sizeofRelayIDs32(ids []uint64) int {
	return 4 * len(ids)
}

func sizeofRelayIDs64(ids []uint64) int {
	return 8 * len(ids)
}

func sizeofRelayNames(names []string) int {
	length := 0
	for _, name := range names {
		length += 4 + len(name)
	}
	return length

}

func sizeofDatacenterCount() int {
	return 4
}

func sizeofDatacenterIDs32(datacenterIDs []uint64) int {
	return 4 * len(datacenterIDs)
}

func sizeofDatacenterIDs64(datacenterIDs []uint64) int {
	return 8 * len(datacenterIDs)
}

func sizeofDatacenterNames(names []string) int {
	length := 0
	for _, name := range names {
		length += 4 + len(name)
	}
	return length
}

func sizeofRelayAddressOld(addrs []string) int {
	length := 0
	for _, addr := range addrs {
		length += 4 + len(addr)
	}
	return length
}

func sizeofRelayAddress(addrs []string) int {
	return len(addrs) * routing.MaxRelayAddressLength
}

func sizeofRelayPublicKeysOld(keys [][]byte) int {
	length := 0
	for _, key := range keys {
		length += 4 + len(key)
	}
	return length
}

func sizeofRelayPublicKeys(keys [][]byte) int {
	return len(keys) * crypto.KeySize
}

// the second area the datacenter count is stored, as of right now identical to the other func
func sizeofDataCenterCount2() int {
	return 4
}

func sizeofRelaysInDatacenterCount(datacenterIDs []uint64) int {
	return 4 * len(datacenterIDs)
}

func sizeofRTTs(rtts []int32) int {
	return 4 * len(rtts)
}

func sizeofSellers(sellers []routing.Seller) int {
	length := 0
	for _, seller := range sellers {
		length += 4 + len(seller.ID) + 4 + len(seller.Name) + 8 + 8
	}

	return length
}

func sizeofRouteMatrixEntry(entries []routing.RouteMatrixEntry) int {
	length := 0
	for range entries {
		length += 4 + 4 + 32 + 32 + 320
	}
	return length
}

func sizeofRouteMatrixEntryOld(entries []routing.RouteMatrixEntry) int {
	length := 0
	for range entries {
		length += 4 + 4 + 32 + 32 + 160
	}
	return length
}

func TestOptimize(t *testing.T) {
	t.Run("CostMatrix", func(t *testing.T) {
		t.Run("UnmarshalBinary()", func(t *testing.T) {
			unmarshalAssertionsVer0 := func(t *testing.T, matrix *routing.CostMatrix, numRelays, numDatacenters int, relayIDs, datacenters []uint64, relayAddrs []string, datacenterRelays [][]uint64, publicKeys [][]byte, rtts []int32) {
				assert.Len(t, matrix.RelayIDs, numRelays)
				assert.Len(t, matrix.RelayAddresses, numRelays)
				assert.Len(t, matrix.RelayPublicKeys, numRelays)
				assert.Len(t, matrix.DatacenterRelays, numDatacenters)
				assert.Len(t, matrix.RTT, len(rtts))

				for _, id := range relayIDs {
					assert.Contains(t, matrix.RelayIDs, id&0xFFFFFFFF)
				}

				for i, addr := range relayAddrs {
					tmp := make([]byte, len(addr))
					copy(tmp, addr)
					assert.Equal(t, matrix.RelayAddresses[i], tmp)
				}

				for i, pk := range publicKeys {
					assert.Equal(t, matrix.RelayPublicKeys[i], pk)
				}

				for i := 0; i < numDatacenters; i++ {
					assert.Contains(t, matrix.DatacenterRelays, datacenters[i]&0xFFFFFFFF)

					relays := matrix.DatacenterRelays[datacenters[i]]
					for j := 0; j < len(datacenterRelays[i]); j++ {
						assert.Equal(t, relays[j], datacenterRelays[i][j]&0xFFFFFFFF)
					}
				}

				for i, rtt := range rtts {
					assert.Equal(t, matrix.RTT[i], rtt)
				}
			}

			unmarshalAssertionsVer1 := func(t *testing.T, matrix *routing.CostMatrix, relayNames []string) {
				assert.Len(t, matrix.RelayNames, len(relayNames))
				for _, name := range relayNames {
					assert.Contains(t, matrix.RelayNames, name)
				}
			}

			unmarshalAssertionsVer2 := func(t *testing.T, matrix *routing.CostMatrix, datacenterIDs []uint64, datacenterNames []string) {
				assert.Len(t, matrix.DatacenterIDs, len(datacenterIDs))
				assert.Len(t, matrix.DatacenterNames, len(datacenterNames))

				for _, id := range datacenterIDs {
					assert.Contains(t, matrix.DatacenterIDs, id&0xFFFFFFFF)
				}

				for _, name := range datacenterNames {
					assert.Contains(t, matrix.DatacenterNames, name)
				}
			}

			unmarshalAssertionsVer3 := func(t *testing.T, matrix *routing.CostMatrix, numRelays, numDatacenters int, relayIDs, datacenters []uint64, relayAddrs []string, datacenterRelays [][]uint64, publicKeys [][]byte, rtts []int32, relayNames []string, datacenterIDs []uint64, datacenterNames []string) {
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

				unmarshalAssertionsVer1(t, matrix, relayNames)

				assert.Len(t, matrix.DatacenterIDs, len(datacenterIDs))
				assert.Len(t, matrix.DatacenterNames, len(datacenterNames))

				for _, id := range datacenterIDs {
					assert.Contains(t, matrix.DatacenterIDs, id)
				}

				for _, name := range datacenterNames {
					assert.Contains(t, matrix.DatacenterNames, name)
				}
			}

			unmarshalAssertionsVer4 := func(t *testing.T, matrix *routing.CostMatrix, sellers []routing.Seller) {
				assert.Len(t, matrix.RelaySellers, len(sellers))
				for i, seller := range sellers {
					assert.Equal(t, seller.ID, matrix.RelaySellers[i].ID)
					assert.Equal(t, seller.Name, matrix.RelaySellers[i].Name)
					assert.Equal(t, seller.IngressPriceCents, matrix.RelaySellers[i].IngressPriceCents)
					assert.Equal(t, seller.EgressPriceCents, matrix.RelaySellers[i].EgressPriceCents)
				}
			}

			t.Run("version number == 0", func(t *testing.T) {
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}

				relayIDs := addrsToIDs(relayAddrs)

				numRelays := len(relayAddrs)

				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}

				datacenters := []uint64{0, 1, 2, 3, 4}

				numDatacenters := len(datacenters)

				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}

				rtts := make([]int32, routing.TriMatrixLength(numRelays))

				for i := range rtts {
					rtts[i] = int32(rand.Int())
				}

				buffSize := 0
				// ...
				buffSize += sizeofVersionNumber()
				// relay count number
				buffSize += sizeofRelayCount()
				// all relay ids
				buffSize += sizeofRelayIDs32(relayIDs)
				// relay addresses in the old format
				buffSize += sizeofRelayAddressOld(relayAddrs)
				// relay public keys in the old format
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				// the second time the datacenter count is read
				buffSize += sizeofDataCenterCount2()
				// the datacenter ids that will be put into the map
				buffSize += sizeofDatacenterIDs32(datacenters)
				// the number of relays in the datacenter
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				// the relays in the datacenters
				buffSize += sizeofRelayIDs32(relayIDs)
				// the rtt entries
				buffSize += sizeofRTTs(rtts)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 0)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putRTTs(buff, &offset, rtts)

				var matrix routing.CostMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, rtts)
			})

			t.Run("version number == 1", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
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

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs32(relayIDs)
				// the relay names
				buffSize += sizeofRelayNames(relayNames)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRTTs(rtts)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 1)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames) //version >= 1
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putRTTs(buff, &offset, rtts)

				var matrix routing.CostMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, rtts)
				unmarshalAssertionsVer1(t, &matrix, relayNames)
			})

			t.Run("version number == 2", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
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

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRelayNames(relayNames)
				// the first time the datacenter count is read
				buffSize += sizeofDatacenterCount()
				// all the individual datacenter ids
				buffSize += sizeofDatacenterIDs32(datacenters)
				// all the individual datacenter names
				buffSize += sizeofDatacenterNames(datacenterNames)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRTTs(rtts)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 2)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                           // version 1
				putDatacenterStuffOld(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putRTTs(buff, &offset, rtts)

				var matrix routing.CostMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, rtts)
				unmarshalAssertionsVer1(t, &matrix, relayNames)
				unmarshalAssertionsVer2(t, &matrix, datacenters, datacenterNames)
			})

			t.Run("version number == 3", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
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

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 3)
				putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                        // version 1
				putDatacenterStuff(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddresses(buff, &offset, relayAddrs)
				putRelayPublicKeys(buff, &offset, publicKeys)
				putDatacenters(buff, &offset, datacenters, datacenterRelays)
				putRTTs(buff, &offset, rtts)

				var matrix routing.CostMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer3(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, rtts, relayNames, datacenters, datacenterNames)
			})

			t.Run("version number >= 4", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
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
					routing.Seller{ID: "id0", Name: "name0", IngressPriceCents: 1, EgressPriceCents: 2},
					routing.Seller{ID: "id1", Name: "name1", IngressPriceCents: 3, EgressPriceCents: 4},
					routing.Seller{ID: "id2", Name: "name2", IngressPriceCents: 5, EgressPriceCents: 6},
					routing.Seller{ID: "id3", Name: "name3", IngressPriceCents: 7, EgressPriceCents: 8},
					routing.Seller{ID: "id4", Name: "name4", IngressPriceCents: 9, EgressPriceCents: 10},
				}

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

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 4)
				putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                        // version 1
				putDatacenterStuff(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddresses(buff, &offset, relayAddrs)
				putRelayPublicKeys(buff, &offset, publicKeys)
				putDatacenters(buff, &offset, datacenters, datacenterRelays)
				putRTTs(buff, &offset, rtts)
				putSellers(buff, &offset, sellers)

				var matrix routing.CostMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer3(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, rtts, relayNames, datacenters, datacenterNames)
				unmarshalAssertionsVer4(t, &matrix, sellers)
			})

			t.Run("Error cases - v0", func(t *testing.T) {
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}

				relayIDs := addrsToIDs(relayAddrs)

				numRelays := len(relayAddrs)

				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}

				datacenters := []uint64{0, 1, 2, 3, 4}

				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}

				rtts := make([]int32, routing.TriMatrixLength(numRelays))

				for i := range rtts {
					rtts[i] = int32(rand.Int())
				}

				buffSize := 0
				// ...
				buffSize += sizeofVersionNumber()
				// relay count number
				buffSize += sizeofRelayCount()
				// all relay ids
				buffSize += sizeofRelayIDs32(relayIDs)
				// relay addresses in the old format
				buffSize += sizeofRelayAddressOld(relayAddrs)
				// relay public keys in the old format
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				// the second time the datacenter count is read
				buffSize += sizeofDataCenterCount2()
				// the datacenter ids that will be put into the map
				buffSize += sizeofDatacenterIDs32(datacenters)
				// the number of relays in the datacenter
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				// the relays in the datacenters
				buffSize += sizeofRelayIDs32(relayIDs)
				// the rtt entries
				buffSize += sizeofRTTs(rtts)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 0)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putRTTs(buff, &offset, rtts)

				t.Run("version of incoming bin data too high", func(t *testing.T) {
					buff := make([]byte, 4)
					offset := 0
					putVersionNumber(buff, &offset, 5)
					var matrix routing.CostMatrix

					err := matrix.UnmarshalBinary(buff)

					assert.EqualError(t, err, "unknown cost matrix version 5")
				})

				t.Run("Invalid version read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at version number")
				})

				t.Run("Invalid relay count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays")
				})

				t.Run("Invalid relay id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids - ver < 3")
				})

				t.Run("Invalid relay address read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay addresses - ver < 3")
				})

				t.Run("Invalid relay public key read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay public keys - ver < 3")
				})

				t.Run("Invalid datacenter count read second time", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of datacenters (second time)")
				})

				t.Run("Invalid datacenter id read second time", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter id - ver < 3")
				})

				t.Run("Invalid datacenter relay count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays in datacenter")
				})

				t.Run("Invalid datacenter relay id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + 4 + 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids for datacenter - ver < 3")
				})

				t.Run("Invalid rtt read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRTTs(rtts) + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at rtt")
				})
			})

			t.Run("Error cases - v1", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}
				datacenters := []uint64{0, 1, 2, 3, 4}
				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}
				rtts := make([]int32, routing.TriMatrixLength(numRelays))
				for i := range rtts {
					rtts[i] = int32(rand.Int())
				}

				// version 1 stuff
				relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs32(relayIDs)
				// the relay names
				buffSize += sizeofRelayNames(relayNames)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRTTs(rtts)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 1)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames) //version >= 1
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putRTTs(buff, &offset, rtts)

				t.Run("version of incoming bin data too high", func(t *testing.T) {
					buff := make([]byte, 4)
					offset := 0
					putVersionNumber(buff, &offset, 5)
					var matrix routing.CostMatrix

					err := matrix.UnmarshalBinary(buff)

					assert.EqualError(t, err, "unknown cost matrix version 5")
				})

				t.Run("Invalid version read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at version number")
				})

				t.Run("Invalid relay count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays")
				})

				t.Run("Invalid relay id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids - ver < 3")
				})

				t.Run("Invalid relay name read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay names")
				})

				t.Run("Invalid relay address read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay addresses - ver < 3")
				})

				t.Run("Invalid relay public key read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay public keys - ver < 3")
				})

				t.Run("Invalid datacenter count read second time", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of datacenters (second time)")
				})

				t.Run("Invalid datacenter id read second time", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter id - ver < 3")
				})

				t.Run("Invalid datacenter relay count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays in datacenter")
				})

				t.Run("Invalid datacenter relay id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + 4 + 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids for datacenter - ver < 3")
				})

				t.Run("Invalid rtt read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRTTs(rtts) + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at rtt")
				})
			})

			t.Run("Error cases - v2", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}
				datacenters := []uint64{0, 1, 2, 3, 4}
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

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRelayNames(relayNames)
				// the first time the datacenter count is read
				buffSize += sizeofDatacenterCount()
				// all the individual datacenter ids
				buffSize += sizeofDatacenterIDs32(datacenters)
				// all the individual datacenter names
				buffSize += sizeofDatacenterNames(datacenterNames)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRTTs(rtts)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 2)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                           // version 1
				putDatacenterStuffOld(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putRTTs(buff, &offset, rtts)

				t.Run("version of incoming bin data too high", func(t *testing.T) {
					buff := make([]byte, 4)
					offset := 0
					putVersionNumber(buff, &offset, 5)
					var matrix routing.CostMatrix

					err := matrix.UnmarshalBinary(buff)

					assert.EqualError(t, err, "unknown cost matrix version 5")
				})

				t.Run("Invalid version read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at version number")
				})

				t.Run("Invalid relay count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays")
				})

				t.Run("Invalid relay id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids - ver < 3")
				})

				t.Run("Invalid relay name read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay names")
				})

				t.Run("Invalid datacenter count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter count")
				})

				t.Run("Invalid datacenter id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter ids - ver < 3")
				})

				t.Run("Invalid datacenter name read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofDatacenterNames(datacenterNames) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter names")
				})

				t.Run("Invalid relay address read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay addresses - ver < 3")
				})

				t.Run("Invalid relay public key read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay public keys - ver < 3")
				})

				t.Run("Invalid datacenter count read second time", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of datacenters (second time)")
				})

				t.Run("Invalid datacenter id read second time", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter id - ver < 3")
				})

				t.Run("Invalid datacenter relay count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays in datacenter")
				})

				t.Run("Invalid datacenter relay id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + 4 + 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids for datacenter - ver < 3")
				})

				t.Run("Invalid rtt read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRTTs(rtts) + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at rtt")
				})
			})

			t.Run("Error cases - v3", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}
				datacenters := []uint64{0, 1, 2, 3, 4}
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

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 3)
				putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                        // version 1
				putDatacenterStuff(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddresses(buff, &offset, relayAddrs)
				putRelayPublicKeys(buff, &offset, publicKeys)
				putDatacenters(buff, &offset, datacenters, datacenterRelays)
				putRTTs(buff, &offset, rtts)

				t.Run("version of incoming bin data too high", func(t *testing.T) {
					buff := make([]byte, 4)
					offset := 0
					putVersionNumber(buff, &offset, 5)
					var matrix routing.CostMatrix

					err := matrix.UnmarshalBinary(buff)

					assert.EqualError(t, err, "unknown cost matrix version 5")
				})

				t.Run("Invalid version read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at version number")
				})

				t.Run("Invalid relay count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays")
				})

				t.Run("Invalid relay id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids - ver >= v3")
				})

				t.Run("Invalid relay name read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay names")
				})

				t.Run("Invalid datacenter count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter count")
				})

				t.Run("Invalid datacenter id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 8 + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter ids - ver >= v3")
				})

				t.Run("Invalid datacenter name read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofDatacenterNames(datacenterNames) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter names")
				})

				t.Run("Invalid relay address read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay addresses - ver >= v3")
				})

				t.Run("Invalid relay public key read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay public keys - ver >= v3")
				})

				t.Run("Invalid datacenter count read second time", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of datacenters (second time)")
				})

				t.Run("Invalid datacenter id read second time", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter id - ver >= v3")
				})

				t.Run("Invalid datacenter relay count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays in datacenter")
				})

				t.Run("Invalid datacenter relay id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 8 + 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids for datacenter - ver >= v3")
				})

				t.Run("Invalid rtt read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRTTs(rtts) + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at rtt")
				})
			})

			t.Run("Error cases - v4", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}
				datacenters := []uint64{0, 1, 2, 3, 4}
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
					routing.Seller{ID: "id0", Name: "name0", IngressPriceCents: 1, EgressPriceCents: 2},
					routing.Seller{ID: "id1", Name: "name1", IngressPriceCents: 3, EgressPriceCents: 4},
					routing.Seller{ID: "id2", Name: "name2", IngressPriceCents: 5, EgressPriceCents: 6},
					routing.Seller{ID: "id3", Name: "name3", IngressPriceCents: 7, EgressPriceCents: 8},
					routing.Seller{ID: "id4", Name: "name4", IngressPriceCents: 9, EgressPriceCents: 10},
				}

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

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 4)
				putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                        // version 1
				putDatacenterStuff(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddresses(buff, &offset, relayAddrs)
				putRelayPublicKeys(buff, &offset, publicKeys)
				putDatacenters(buff, &offset, datacenters, datacenterRelays)
				putRTTs(buff, &offset, rtts)
				putSellers(buff, &offset, sellers)

				t.Run("version of incoming bin data too high", func(t *testing.T) {
					buff := make([]byte, 4)
					offset := 0
					putVersionNumber(buff, &offset, 5)
					var matrix routing.CostMatrix

					err := matrix.UnmarshalBinary(buff)

					assert.EqualError(t, err, "unknown cost matrix version 5")
				})

				t.Run("Invalid version read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at version number")
				})

				t.Run("Invalid relay count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays")
				})

				t.Run("Invalid relay id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids - ver >= v3")
				})

				t.Run("Invalid relay name read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay names")
				})

				t.Run("Invalid datacenter count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter count")
				})

				t.Run("Invalid datacenter id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 8 + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter ids - ver >= v3")
				})

				t.Run("Invalid datacenter name read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofDatacenterNames(datacenterNames) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter names")
				})

				t.Run("Invalid relay address read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay addresses - ver >= v3")
				})

				t.Run("Invalid relay public key read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay public keys - ver >= v3")
				})

				t.Run("Invalid datacenter count read second time", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of datacenters (second time)")
				})

				t.Run("Invalid datacenter id read second time", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at datacenter id - ver >= v3")
				})

				t.Run("Invalid datacenter relay count read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at number of relays in datacenter")
				})

				t.Run("Invalid datacenter relay id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 8 + 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at relay ids for datacenter - ver >= v3")
				})

				t.Run("Invalid rtt read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := sizeofRTTs(rtts) + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read at rtt")
				})

				t.Run("Invalid seller id read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + len(sellers[0].ID) + sizeofRTTs(rtts) + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read on relay seller ID")
				})

				t.Run("Invalid seller name read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 4 + len(sellers[0].Name) + 4 + len(sellers[0].ID) + sizeofRTTs(rtts) + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read on relay seller name")
				})

				t.Run("Invalid seller ingress price read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 8 + 4 + len(sellers[0].Name) + 4 + len(sellers[0].ID) + sizeofRTTs(rtts) + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read on relay seller ingress price")
				})

				t.Run("Invalid seller egress price read", func(t *testing.T) {
					var matrix routing.CostMatrix
					offset := 8 + 8 + 4 + len(sellers[0].Name) + 4 + len(sellers[0].ID) + sizeofRTTs(rtts) + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[CostMatrix] invalid read on relay seller egress price")
				})
			})
		})

		t.Run("MarshalBinary()", func(t *testing.T) {
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
		})

		t.Run("ServeHTTP()", func(t *testing.T) {
			t.Run("Failure to serve HTTP", func(t *testing.T) {
				// Create and populate a malformed cost matrix
				matrix := getPopulatedCostMatrix(true)

				// Create a dummy http request to test ServeHTTP
				recorder := httptest.NewRecorder()
				request, err := http.NewRequest("GET", "/", nil)
				assert.NoError(t, err)

				matrix.ServeHTTP(recorder, request)

				// Get the response
				response := recorder.Result()

				assert.Equal(t, 500, response.StatusCode)
			})

			t.Run("Successful Serve", func(t *testing.T) {
				// Create and populate a cost matrix
				matrix := getPopulatedCostMatrix(false)

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

				// Validate the response
				assert.Equal(t, "application/octet-stream", response.Header.Get("Content-Type"))
				assert.Equal(t, matrix, &receivedMatrix)
			})
		})

		t.Run("WriteTo()", func(t *testing.T) {
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
		})

		t.Run("ReadFrom()", func(t *testing.T) {
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
		})

		t.Run("Optimize()", func(t *testing.T) {
			t.Skip("using old Optimize() tests from core/core_test.go for now")
		})
	})

	t.Run("Old tests from core/core_test.go", func(t *testing.T) {
		t.Run("TestCostMatrix() - cost matrix assertions with version 0 data", func(t *testing.T) {
			raw, err := ioutil.ReadFile("test_data/cost.bin")
			assert.Nil(t, err)
			assert.Equal(t, len(raw), 355188, "cost.bin should be 355188 bytes")

			var costMatrix routing.CostMatrix
			err = costMatrix.UnmarshalBinary(raw)
			assert.Nil(t, err)

			costMatrixData, err := costMatrix.MarshalBinary()
			assert.NoError(t, err)

			var readCostMatrix routing.CostMatrix
			err = readCostMatrix.UnmarshalBinary(costMatrixData)
			assert.Nil(t, err)

			assert.Equal(t, costMatrix.RelayIDs, readCostMatrix.RelayIDs, "relay id mismatch")

			// this was the old line however because relay addresses are written with extra 0's this is how they must be checked
			// assert.Equal(t, costMatrix.RelayAddresses, readCostMatrix.RelayAddresses, "relay address mismatch")

			assert.Len(t, readCostMatrix.RelayAddresses, len(costMatrix.RelayAddresses))
			for i, addr := range costMatrix.RelayAddresses {
				assert.Equal(t, string(addr), strings.Trim(string(readCostMatrix.RelayAddresses[i]), string([]byte{0x0})))
			}

			assert.Equal(t, costMatrix.RelayPublicKeys, readCostMatrix.RelayPublicKeys, "relay public key mismatch")
			assert.Equal(t, costMatrix.DatacenterRelays, readCostMatrix.DatacenterRelays, "datacenter relays mismatch")
			assert.Equal(t, costMatrix.RTT, readCostMatrix.RTT, "relay rtt mismatch")
		})
	})
}

func TestRouting(t *testing.T) {
	t.Run("ResolveRelay", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			costfile, err := os.Open("./test_data/cost.bin")
			assert.NoError(t, err)

			var costMatrix routing.CostMatrix
			_, err = costMatrix.ReadFrom(costfile)
			assert.NoError(t, err)

			var routeMatrix routing.RouteMatrix
			err = costMatrix.Optimize(&routeMatrix, 1)
			assert.NoError(t, err)

			expected := routing.Relay{
				ID: 2836356269,
				Addr: net.UDPAddr{
					IP:   net.ParseIP("13.238.77.175"),
					Port: 40000,
				},
				PublicKey: []byte{0x58, 0xaf, 0x19, 0x5, 0xf7, 0xa8, 0xae, 0x73, 0xc6, 0xd3, 0xec, 0x85, 0x2f, 0xd8, 0x9b, 0x5a, 0xce, 0x0, 0x38, 0xca, 0x26, 0x39, 0xa4, 0x5d, 0x82, 0x3c, 0x71, 0xa8, 0x4, 0x11, 0xfb, 0x32},
			}

			actual, err := routeMatrix.ResolveRelay(2836356269)
			assert.NoError(t, err)
			assert.Equal(t, expected, actual)
		})

		t.Run("Relay ID not found", func(t *testing.T) {
			routeMatrix := routing.RouteMatrix{
				RelayIndicies: map[uint64]int{},
			}
			_, err := routeMatrix.ResolveRelay(0)
			assert.EqualError(t, err, "relay 0 not in matrix")
		})

		t.Run("Invalid relay index", func(t *testing.T) {
			routeMatrix := routing.RouteMatrix{
				RelayIndicies:  map[uint64]int{0: 10},
				RelayAddresses: [][]byte{},
			}
			_, err := routeMatrix.ResolveRelay(0)
			assert.EqualError(t, err, "relay 0 has an invalid index 10")
		})

		t.Run("Invalid relay address", func(t *testing.T) {
			routeMatrix := routing.RouteMatrix{
				RelayIndicies:   map[uint64]int{0: 0},
				RelayAddresses:  [][]byte{[]byte("Invalid")},
				RelayPublicKeys: [][]byte{{0x58, 0xaf, 0x19, 0x5, 0xf7, 0xa8, 0xae, 0x73, 0xc6, 0xd3, 0xec, 0x85, 0x2f, 0xd8, 0x9b, 0x5a, 0xce, 0x0, 0x38, 0xca, 0x26, 0x39, 0xa4, 0x5d, 0x82, 0x3c, 0x71, 0xa8, 0x4, 0x11, 0xfb, 0x32}},
			}
			_, err := routeMatrix.ResolveRelay(0)
			assert.Error(t, err)
		})

		t.Run("Failed to parse port", func(t *testing.T) {
			routeMatrix := routing.RouteMatrix{
				RelayIndicies:   map[uint64]int{0: 0},
				RelayAddresses:  [][]byte{[]byte("127.0.0.1:abcde")},
				RelayPublicKeys: [][]byte{{0x58, 0xaf, 0x19, 0x5, 0xf7, 0xa8, 0xae, 0x73, 0xc6, 0xd3, 0xec, 0x85, 0x2f, 0xd8, 0x9b, 0x5a, 0xce, 0x0, 0x38, 0xca, 0x26, 0x39, 0xa4, 0x5d, 0x82, 0x3c, 0x71, 0xa8, 0x4, 0x11, 0xfb, 0x32}},
			}
			_, err := routeMatrix.ResolveRelay(0)
			assert.Error(t, err)
		})
	})

	t.Run("RelaysIn", func(t *testing.T) {
		costfile, err := os.Open("./test_data/cost.bin")
		assert.NoError(t, err)

		var costMatrix routing.CostMatrix
		_, err = costMatrix.ReadFrom(costfile)
		assert.NoError(t, err)

		var routeMatrix routing.RouteMatrix
		err = costMatrix.Optimize(&routeMatrix, 1)
		assert.NoError(t, err)

		tests := []struct {
			name     string
			input    routing.Datacenter
			expected []routing.Relay
		}{
			{"datacenter not found", routing.Datacenter{ID: 0}, nil},
			{
				"datacenter with relays",
				routing.Datacenter{ID: 69517923},
				[]routing.Relay{
					{ID: 3407334631, Addr: net.UDPAddr{IP: net.ParseIP("162.253.71.170"), Port: 40000}, PublicKey: []byte{0x87, 0xde, 0x7, 0x9, 0x35, 0xee, 0xdd, 0xb0, 0xf0, 0xfe, 0xfe, 0xa7, 0xa5, 0x4e, 0x14, 0xd1, 0x2d, 0x3b, 0xd9, 0x8c, 0x0, 0x49, 0xcd, 0xf0, 0x14, 0x7e, 0xa5, 0xe0, 0x52, 0xb4, 0xe6, 0x76}},
					{ID: 1447163127, Addr: net.UDPAddr{IP: net.ParseIP("172.98.66.170"), Port: 40000}, PublicKey: []byte{0x1e, 0x80, 0x89, 0x6a, 0x46, 0xa9, 0xb4, 0x6d, 0x27, 0x54, 0x28, 0x16, 0x56, 0xe, 0x1f, 0x6f, 0xee, 0xee, 0x6a, 0x98, 0x5a, 0xbb, 0x8b, 0x83, 0x96, 0xcb, 0x13, 0xc5, 0x66, 0x8, 0x92, 0x31}},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				actual := routeMatrix.RelaysIn(test.input)
				assert.Equal(t, test.expected, actual)
			})
		}

		// relay length is 0
		routeMatrix.DatacenterRelays[0] = []uint64{}
		relays := routeMatrix.RelaysIn(routing.Datacenter{ID: 0})
		assert.Nil(t, relays)

		// error while resolving at least one relay
		routeMatrix = routing.RouteMatrix{
			RelayIndicies:    map[uint64]int{0: 0},
			RelayAddresses:   [][]byte{[]byte("127.0.0.1:abcde")},
			RelayPublicKeys:  [][]byte{{0x58, 0xaf, 0x19, 0x5, 0xf7, 0xa8, 0xae, 0x73, 0xc6, 0xd3, 0xec, 0x85, 0x2f, 0xd8, 0x9b, 0x5a, 0xce, 0x0, 0x38, 0xca, 0x26, 0x39, 0xa4, 0x5d, 0x82, 0x3c, 0x71, 0xa8, 0x4, 0x11, 0xfb, 0x32}},
			DatacenterRelays: map[uint64][]uint64{0: []uint64{0, 1}},
		}
		relays = routeMatrix.RelaysIn(routing.Datacenter{ID: 0})
		assert.NotNil(t, relays)
	})

	t.Run("Routes", func(t *testing.T) {
		costfile, err := os.Open("./test_data/cost.bin")
		assert.NoError(t, err)

		var costMatrix routing.CostMatrix
		_, err = costMatrix.ReadFrom(costfile)
		assert.NoError(t, err)

		var routeMatrix routing.RouteMatrix
		err = costMatrix.Optimize(&routeMatrix, 1)
		assert.NoError(t, err)

		tests := []struct {
			name        string
			from        []routing.Relay
			to          []routing.Relay
			expected    []routing.Route
			expectedErr error
			selectors   []routing.SelectorFunc
		}{
			{"empty from/to sets", []routing.Relay{}, []routing.Relay{}, nil, errors.New("no routes found"), nil},
			{"relays not found", []routing.Relay{{ID: 1}}, []routing.Relay{{ID: 2}}, nil, errors.New("no routes found"), nil},
			{"one relay found", []routing.Relay{{ID: 1}}, []routing.Relay{{ID: 1500948990}}, nil, errors.New("no routes found"), nil},
			{
				"no selectors",
				[]routing.Relay{{ID: 2836356269}},
				[]routing.Relay{{ID: 3263834878}, {ID: 1500948990}},
				[]routing.Route{
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2923051732}, {ID: 1884974764}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 182},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2641807504}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 182},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 1348914502}, {ID: 1884974764}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 182},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2576485547}, {ID: 1835585494}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 183},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1348914502}, {ID: 1884974764}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 183},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2663193268}, {ID: 2504465311}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 184},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 427962386}, {ID: 2504465311}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 184},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 4058587524}, {ID: 1350942731}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 184},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1500948990}},
						Stats:  routing.Stats{RTT: 311},
					},
				},
				nil,
				nil,
			},
			{
				"best RTT",
				[]routing.Relay{{ID: 2836356269}},
				[]routing.Relay{{ID: 3263834878}, {ID: 1500948990}},
				[]routing.Route{
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2923051732}, {ID: 1884974764}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 182},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2641807504}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 182},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 1348914502}, {ID: 1884974764}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 182},
					},
				},
				nil,
				[]routing.SelectorFunc{
					routing.SelectBestRTT(),
				},
			},
			{
				"acceptable routes",
				[]routing.Relay{{ID: 2836356269}},
				[]routing.Relay{{ID: 3263834878}, {ID: 1500948990}},
				[]routing.Route{
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2923051732}, {ID: 1884974764}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 182},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2641807504}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 182},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 1348914502}, {ID: 1884974764}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 182},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2576485547}, {ID: 1835585494}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 183},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1348914502}, {ID: 1884974764}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 183},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2663193268}, {ID: 2504465311}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 184},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 427962386}, {ID: 2504465311}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 184},
					},
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 4058587524}, {ID: 1350942731}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 184},
					},
				},
				nil,
				[]routing.SelectorFunc{
					routing.SelectAcceptableRoutesFromBestRTT(10),
				},
			},
			{
				"contains route",
				[]routing.Relay{{ID: 2836356269}},
				[]routing.Relay{{ID: 3263834878}, {ID: 1500948990}},
				[]routing.Route{
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2923051732}, {ID: 1884974764}, {ID: 3263834878}},
						Stats:  routing.Stats{RTT: 182},
					},
				},
				nil,
				[]routing.SelectorFunc{
					routing.SelectContainsRouteHash(14287039991941962633),
				},
			},
			{
				"routes by random dest relay",
				[]routing.Relay{{ID: 2836356269}},
				[]routing.Relay{{ID: 3263834878}, {ID: 1500948990}},
				[]routing.Route{
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1500948990}},
						Stats:  routing.Stats{RTT: 311},
					},
				},
				nil,
				[]routing.SelectorFunc{
					routing.SelectRoutesByRandomDestRelay(rand.NewSource(0)),
				},
			},
			{
				"random route",
				[]routing.Relay{{ID: 2836356269}},
				[]routing.Relay{{ID: 3263834878}, {ID: 1500948990}},
				[]routing.Route{
					routing.Route{
						Relays: []routing.Relay{{ID: 2836356269}, {ID: 1370686037}, {ID: 2923051732}, {ID: 1884974764}},
						Stats:  routing.Stats{RTT: 182},
					},
				},
				nil,
				[]routing.SelectorFunc{
					routing.SelectRandomRoute(rand.NewSource(0)),
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				actual, err := routeMatrix.Routes(test.from, test.to, test.selectors...)
				assert.Equal(t, test.expectedErr, err)
				assert.Equal(t, len(test.expected), len(actual))

				for routeidx, route := range test.expected {
					assert.Equal(t, len(test.expected[routeidx].Relays), len(route.Relays))

					for relayidx := range route.Relays {
						assert.Equal(t, test.expected[routeidx].Relays[relayidx].ID, actual[routeidx].Relays[relayidx].ID)
						assert.NotNil(t, actual[routeidx].Relays[relayidx].Addr.IP)
						assert.False(t, actual[routeidx].Relays[relayidx].Addr.IP.IsLoopback())
						assert.Greater(t, actual[routeidx].Relays[relayidx].Addr.Port, 0)
						assert.NotNil(t, actual[routeidx].Relays[relayidx].PublicKey)
						assert.Equal(t, crypto.KeySize, len(actual[routeidx].Relays[relayidx].PublicKey))
					}

					assert.Equal(t, test.expected[routeidx].Stats, actual[routeidx].Stats)
				}
			})
		}
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

func BenchmarkRouting(b *testing.B) {
	costfile, _ := os.Open("./test_data/cost.bin")

	var costMatrix routing.CostMatrix
	costMatrix.ReadFrom(costfile)

	var routeMatrix routing.RouteMatrix
	costMatrix.Optimize(&routeMatrix, 1)

	from := []routing.Relay{{ID: 2836356269}}
	to := []routing.Relay{{ID: 3263834878}, {ID: 1500948990}}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		routeMatrix.Routes(from, to)
	}
}

func BenchmarkResolveRelay(b *testing.B) {
	costfile, _ := os.Open("./test_data/cost.bin")

	var costMatrix routing.CostMatrix
	costMatrix.ReadFrom(costfile)

	var routeMatrix routing.RouteMatrix
	costMatrix.Optimize(&routeMatrix, 1)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		routeMatrix.ResolveRelay(2836356269)
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
		routeMatrix.RelaysIn(routing.Datacenter{ID: routeMatrix.DatacenterIDs[0], Name: routeMatrix.DatacenterNames[0]})
	}
}
