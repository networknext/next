package routing_test

import (
	"encoding/binary"
	"io/ioutil"
	"math/rand"
	"strings"
	"testing"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func addrsToIDs(addrs []string) []uint64 {
	retval := make([]uint64, len(addrs))
	for i, addr := range addrs {
		retval[i] = uint64(routing.GetRelayID(addr))
	}
	return retval
}

func generateRouteMatrixEntries(entries []routing.RouteMatrixEntry) {
	for i := 0; i < len(entries); i++ {
		entry := routing.RouteMatrixEntry{
			DirectRTT: rand.Int31(),
			NumRoutes: 8,
		}

		var routeRtt [8]int32
		for j := 0; j < 8; j++ {
			routeRtt[j] = rand.Int31()
		}
		entry.RouteRTT = routeRtt

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

func putRtts(buff []byte, offset *int, rtts []int32) {
	putInt32s(buff, offset, rtts...)
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
				assert.Len(t, matrix.RelayIds, numRelays)
				assert.Len(t, matrix.RelayAddresses, numRelays)
				assert.Len(t, matrix.RelayPublicKeys, numRelays)
				assert.Len(t, matrix.DatacenterRelays, numDatacenters)
				assert.Len(t, matrix.RTT, len(rtts))

				for _, id := range relayIDs {
					assert.Contains(t, matrix.RelayIds, id&0xFFFFFFFF)
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
				assert.Len(t, matrix.DatacenterIds, len(datacenterIDs))
				assert.Len(t, matrix.DatacenterNames, len(datacenterNames))

				for _, id := range datacenterIDs {
					assert.Contains(t, matrix.DatacenterIds, id&0xFFFFFFFF)
				}

				for _, name := range datacenterNames {
					assert.Contains(t, matrix.DatacenterNames, name)
				}
			}

			unmarshalAssertionsVer3 := func(t *testing.T, matrix *routing.CostMatrix, numRelays, numDatacenters int, relayIDs, datacenters []uint64, relayAddrs []string, datacenterRelays [][]uint64, publicKeys [][]byte, rtts []int32, relayNames []string, datacenterIDs []uint64, datacenterNames []string) {
				assert.Len(t, matrix.RelayIds, numRelays)
				assert.Len(t, matrix.RelayAddresses, numRelays)
				assert.Len(t, matrix.RelayPublicKeys, numRelays)
				assert.Len(t, matrix.DatacenterRelays, numDatacenters)
				assert.Len(t, matrix.RTT, len(rtts))

				for _, id := range relayIDs {
					assert.Contains(t, matrix.RelayIds, id)
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

				assert.Len(t, matrix.DatacenterIds, len(datacenterIDs))
				assert.Len(t, matrix.DatacenterNames, len(datacenterNames))

				for _, id := range datacenterIDs {
					assert.Contains(t, matrix.DatacenterIds, id)
				}

				for _, name := range datacenterNames {
					assert.Contains(t, matrix.DatacenterNames, name)
				}
			}

			t.Run("version of incoming bin data too high", func(t *testing.T) {
				buff := make([]byte, 4)
				offset := 0
				putVersionNumber(buff, &offset, 4)
				var matrix routing.CostMatrix

				err := matrix.UnmarshalBinary(buff)

				assert.EqualError(t, err, "unknown cost matrix version 4")
			})

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

				for i, _ := range rtts {
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
				putRtts(buff, &offset, rtts)

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
				for i, _ := range rtts {
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
				putRtts(buff, &offset, rtts)

				var matrix routing.CostMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, rtts)
				unmarshalAssertionsVer1(t, &matrix, relayNames)
			})

			t.Run("version number >= 2", func(t *testing.T) {
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
				for i, _ := range rtts {
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
				putRtts(buff, &offset, rtts)

				var matrix routing.CostMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, rtts)
				unmarshalAssertionsVer1(t, &matrix, relayNames)
				unmarshalAssertionsVer2(t, &matrix, datacenters, datacenterNames)
			})

			t.Run("version number >= 3", func(t *testing.T) {
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
				for i, _ := range rtts {
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
				putRtts(buff, &offset, rtts)

				var matrix routing.CostMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer3(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, rtts, relayNames, datacenters, datacenterNames)
			})
		})

		t.Run("MarshalBinary()", func(t *testing.T) {
			t.Run("MarshalBinary -> UnmarshalBinary equality", func(t *testing.T) {
				var matrix routing.CostMatrix
				matrix.RelayIds = make([]uint64, 2)
				matrix.RelayIds[0] = 123
				matrix.RelayIds[1] = 456

				matrix.RelayNames = make([]string, 2)
				matrix.RelayNames[0] = "first"
				matrix.RelayNames[1] = "second"

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

				matrix.DatacenterIds = make([]uint64, 2)
				matrix.DatacenterIds[0] = 999
				matrix.DatacenterIds[1] = 111

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

				var other routing.CostMatrix

				bin, err := matrix.MarshalBinary()
				assert.NoError(t, err)

				// essentialy this asserts the result of MarshalBinary(),
				// if Unmarshal tests pass then the binary data from Marshal
				// is valid if unmarshaling equals the original
				err = other.UnmarshalBinary(bin)
				assert.NoError(t, err)

				assert.Equal(t, matrix, other)
			})
		})

		t.Run("Optimize()", func(t *testing.T) {
			t.Skip("using old Optimize() tests from core/core_test.go for now")
		})
	})

	t.Run("RouteMatrix", func(t *testing.T) {
		t.Run("UnmarshalBinary()", func(t *testing.T) {
			unmarshalAssertionsVer0 := func(t *testing.T, matrix *routing.RouteMatrix, numRelays, numDatacenters int, relayIDs, datacenters []uint64, relayAddrs []string, datacenterRelays [][]uint64, publicKeys [][]byte, entries []routing.RouteMatrixEntry) {
				assert.Len(t, matrix.RelayIds, numRelays)
				assert.Len(t, matrix.RelayAddresses, numRelays)
				assert.Len(t, matrix.RelayPublicKeys, numRelays)
				assert.Len(t, matrix.DatacenterRelays, numDatacenters)
				assert.Len(t, matrix.Entries, len(entries))

				for _, id := range relayIDs {
					assert.Contains(t, matrix.RelayIds, id&0xFFFFFFFF)
				}

				for _, addr := range relayAddrs {
					tmp := make([]byte, len(addr))
					copy(tmp, addr)
					assert.Contains(t, matrix.RelayAddresses, tmp)
				}

				for _, pk := range publicKeys {
					assert.Contains(t, matrix.RelayPublicKeys, pk)
				}

				for i := 0; i < numDatacenters; i++ {
					assert.Contains(t, matrix.DatacenterRelays, datacenters[i]&0xFFFFFFFF)

					relays := matrix.DatacenterRelays[datacenters[i]]
					for j := 0; j < len(datacenterRelays[i]); j++ {
						assert.Contains(t, relays, datacenterRelays[i][j]&0xFFFFFFFF)
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
							assert.Equal(t, id&0xFFFFFFFF, actual.RouteRelays[i][j])
						}
					}
				}
			}

			unmarshalAssertionsVer1 := func(t *testing.T, matrix *routing.RouteMatrix, relayNames []string) {
				assert.Len(t, matrix.RelayNames, len(relayNames))
				for _, name := range relayNames {
					assert.Contains(t, matrix.RelayNames, name)
				}
			}

			unmarshalAssertionsVer2 := func(t *testing.T, matrix *routing.RouteMatrix, datacenterIDs []uint64, datacenterNames []string) {
				assert.Len(t, matrix.DatacenterIds, len(datacenterIDs))
				assert.Len(t, matrix.DatacenterNames, len(datacenterNames))

				for _, id := range datacenterIDs {
					assert.Contains(t, matrix.DatacenterIds, id&0xFFFFFFFF)
				}

				for _, name := range datacenterNames {
					assert.Contains(t, matrix.DatacenterNames, name)
				}
			}

			unmarshalAssertionsVer3 := func(t *testing.T, matrix *routing.RouteMatrix, numRelays, numDatacenters int, relayIDs, datacenters []uint64, relayAddrs []string, datacenterRelays [][]uint64, publicKeys [][]byte, entries []routing.RouteMatrixEntry, relayNames []string, datacenterIDs []uint64, datacenterNames []string) {
				assert.Len(t, matrix.RelayIds, numRelays)
				assert.Len(t, matrix.RelayAddresses, numRelays)
				assert.Len(t, matrix.RelayPublicKeys, numRelays)
				assert.Len(t, matrix.DatacenterRelays, numDatacenters)
				assert.Len(t, matrix.Entries, len(entries))

				for _, id := range relayIDs {
					assert.Contains(t, matrix.RelayIds, id)
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

				for i, entry := range entries {
					assert.Equal(t, matrix.Entries[i], entry)
				}

				unmarshalAssertionsVer1(t, matrix, relayNames)

				assert.Len(t, matrix.DatacenterIds, len(datacenterIDs))
				assert.Len(t, matrix.DatacenterNames, len(datacenterNames))

				for _, id := range datacenterIDs {
					assert.Contains(t, matrix.DatacenterIds, id)
				}

				for _, name := range datacenterNames {
					assert.Contains(t, matrix.DatacenterNames, name)
				}
			}

			t.Run("version of incoming bin data too high", func(t *testing.T) {
				buff := make([]byte, 4)
				offset := 0
				putVersionNumber(buff, &offset, 4)
				var matrix routing.RouteMatrix

				err := matrix.UnmarshalBinary(buff)

				assert.EqualError(t, err, "unknown route matrix version: 4")
			})

			t.Run("version number 0", func(t *testing.T) {
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

				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				// the size of each route entry
				buffSize += sizeofRouteMatrixEntryOld(entries)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 0)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putEntriesOld(buff, &offset, entries)

				var matrix routing.RouteMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, entries)
			})

			t.Run("version number 1", func(t *testing.T) {
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
				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

				// version 1 stuff
				relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRelayNames(relayNames)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRouteMatrixEntryOld(entries)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 1)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames) //version >= 1
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putEntriesOld(buff, &offset, entries)

				var matrix routing.RouteMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, entries)
				unmarshalAssertionsVer1(t, &matrix, relayNames)
			})

			t.Run("version number 2", func(t *testing.T) {
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
				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

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
				buffSize += sizeofDatacenterCount()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofDatacenterNames(datacenterNames)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRouteMatrixEntryOld(entries)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 2)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                           // version 1
				putDatacenterStuffOld(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putEntriesOld(buff, &offset, entries)

				var matrix routing.RouteMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, entries)
				unmarshalAssertionsVer1(t, &matrix, relayNames)
				unmarshalAssertionsVer2(t, &matrix, datacenters, datacenterNames)
			})

			t.Run("version number 3", func(t *testing.T) {
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
				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

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
				buffSize += sizeofRouteMatrixEntry(entries)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 3)
				putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                        // version 1
				putDatacenterStuff(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddresses(buff, &offset, relayAddrs)
				putRelayPublicKeys(buff, &offset, publicKeys)
				putDatacenters(buff, &offset, datacenters, datacenterRelays)
				putEntries(buff, &offset, entries)

				var matrix routing.RouteMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer3(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, entries, relayNames, datacenters, datacenterNames)
			})
		})

		t.Run("MarshalBinary()", func(t *testing.T) {
			t.Run("MarshalBinary -> UnmarshalBinary equality", func(t *testing.T) {
				var matrix routing.RouteMatrix
				matrix.RelayIds = make([]uint64, 2)
				matrix.RelayIds[0] = 123
				matrix.RelayIds[1] = 456

				matrix.RelayNames = make([]string, 2)
				matrix.RelayNames[0] = "first"
				matrix.RelayNames[1] = "second"

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

				matrix.DatacenterIds = make([]uint64, 2)
				matrix.DatacenterIds[0] = 999
				matrix.DatacenterIds[1] = 111

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

				var other routing.RouteMatrix

				bin, err := matrix.MarshalBinary()

				// essentialy this asserts the result of MarshalBinary(),
				// if Unmarshal tests pass then the binary data from Marshal
				// is valid if unmarshaling equals the original
				other.UnmarshalBinary(bin)

				assert.Nil(t, err)
				assert.Equal(t, matrix, other)
			})
		})
	})

	t.Run("Old tests from core/core_test.go", func(t *testing.T) {
		analyze := func(t *testing.T, route_matrix *routing.RouteMatrix) {
			src := route_matrix.RelayIds
			dest := route_matrix.RelayIds

			entries := make([]int32, 0, len(src)*len(dest))

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
								entries = append(entries, improvement)
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

		t.Run("TestCostMatrix() - cost matrix assertions with version 0 data", func(t *testing.T) {
			raw, err := ioutil.ReadFile("test_data/cost.bin")
			assert.Nil(t, err)
			assert.Equal(t, len(raw), 355188, "cost.bin should be 355188 bytes")

			var costMatrix routing.CostMatrix
			err = costMatrix.UnmarshalBinary(raw)
			assert.Nil(t, err)

			costMatrixData, err := costMatrix.MarshalBinary()

			var readCostMatrix routing.CostMatrix
			err = readCostMatrix.UnmarshalBinary(costMatrixData)
			assert.Nil(t, err)

			assert.Equal(t, costMatrix.RelayIds, readCostMatrix.RelayIds, "relay id mismatch")

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

		t.Run("TestRouteMatrixSanity() - test using version 2 example data", func(t *testing.T) {
			var cmatrix routing.CostMatrix
			var rmatrix routing.RouteMatrix

			raw, err := ioutil.ReadFile("test_data/cost-for-sanity-check.bin")
			assert.Nil(t, err)

			err = cmatrix.UnmarshalBinary(raw)
			assert.Nil(t, err)

			err = cmatrix.Optimize(&rmatrix, 1.0)
			assert.Nil(t, err)

			src := rmatrix.RelayIds
			dest := rmatrix.RelayIds

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
			assert.NotNil(t, routeMatrix)
			assert.Equal(t, costMatrix.RelayIds, routeMatrix.RelayIds, "relay id mismatch")
			assert.Equal(t, costMatrix.RelayAddresses, routeMatrix.RelayAddresses, "relay address mismatch")
			assert.Equal(t, costMatrix.RelayPublicKeys, routeMatrix.RelayPublicKeys, "relay public key mismatch")

			routeMatrixData, err := routeMatrix.MarshalBinary()
			assert.Nil(t, err)

			var readRouteMatrix routing.RouteMatrix
			err = readRouteMatrix.UnmarshalBinary(routeMatrixData)
			assert.Nil(t, err)

			assert.Equal(t, routeMatrix.RelayIds, readRouteMatrix.RelayIds, "relay id mismatch")
			// todo: relay names soon
			// this was the old line however because relay addresses are written with extra 0's this is how they must be checked
			// assert.Equal(t, routeMatrix.RelayAddresses, readRouteMatrix.RelayAddresses, "relay address mismatch")

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
							t.Errorf("RouteRelayId mismatch\n")
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
}
