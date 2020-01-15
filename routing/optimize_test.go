package routing_test

import (
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func addrsToIDs(addrs []string) []uint64 {
	retval := make([]uint64, len(addrs))
	for i, addr := range addrs {
		retval[i] = core.GetRelayID(addr)
	}
	return retval
}

func putVersionNumber(buff []byte, offset *int, version uint32) {
	binary.LittleEndian.PutUint32(buff, version)
	*offset += 4
}

func putRelayIDs(buff []byte, offset *int, ids []uint64) {
	count := len(ids)

	binary.LittleEndian.PutUint32(buff[*offset:], uint32(count))
	*offset += 4

	for _, id := range ids {
		binary.LittleEndian.PutUint64(buff[*offset:], id)
		*offset += 8
	}
}

func putRelayNames(buff []byte, offset *int, names []string) {
	for _, name := range names {
		binary.LittleEndian.PutUint32(buff[*offset:], uint32(len(name)))
		*offset += 4
		copy(buff[*offset:], name)
		*offset += len(name)
	}
}

func putDatacenterStuff(buff []byte, offset *int, ids []uint64, names []string) {
	binary.LittleEndian.PutUint32(buff[*offset:], uint32(len(ids)))
	*offset += 4

	for i := 0; i < len(ids); i++ {
		id, name := ids[i], names[i]
		binary.LittleEndian.PutUint64(buff[*offset:], id)
		*offset += 8
		binary.LittleEndian.PutUint32(buff[*offset:], uint32(len(name)))
		*offset += 4
		copy(buff[*offset:], name)
		*offset += len(name)
	}
}

func putRelayAddresses(buff []byte, offset *int, addrs []string) {
	for _, addr := range addrs {
		tmp := make([]byte, routing.MaxRelayAddressLength)
		copy(tmp, addr)
		copy(buff[*offset:], tmp)
		*offset += len(tmp)
	}
}

func putRelayPublicKeys(buff []byte, offset *int, pks [][]byte) {
	for _, pk := range pks {
		copy(buff[*offset:], pk)
		*offset += len(pk)
	}
}

func putDatacenters(buff []byte, offset *int, datacenterIDs []uint64, relayIDs [][]uint64) {
	binary.LittleEndian.PutUint32(buff[*offset:], uint32(len(datacenterIDs)))
	*offset += 4

	for i, dcID := range datacenterIDs {
		relays := relayIDs[i]
		binary.LittleEndian.PutUint64(buff[*offset:], dcID)
		*offset += 8

		binary.LittleEndian.PutUint32(buff[*offset:], uint32(len(relays)))
		*offset += 4

		for _, rID := range relays {
			binary.LittleEndian.PutUint64(buff[*offset:], rID)
			*offset += 8
		}
	}
}

func putRtts(buff []byte, offset *int, rtts []int32) {
	for _, rtt := range rtts {
		binary.LittleEndian.PutUint32(buff[*offset:], uint32(rtt))
		*offset += 4
	}
}

func TestOptimize(t *testing.T) {
	t.Run("CostMatrix", func(t *testing.T) {
		t.Run("UnmarshalBinary()", func(t *testing.T) {
			unmarshalAssertionsVer0 := func(matrix *routing.CostMatrix, numRelays, numDatacenters int, relayIDs, datacenters []uint64, relayAddrs []string, datacenterRelays [][]uint64, publicKeys [][]byte, rtts []int32) {
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

				for _, rtt := range rtts {
					assert.Contains(t, matrix.RTT, rtt)
				}
			}

			unmarshalAssertionsVer1 := func(matrix *routing.CostMatrix, relayNames []string) {
				assert.Len(t, matrix.RelayNames, len(relayNames))
				for _, name := range relayNames {
					assert.Contains(t, matrix.RelayNames, name)
				}
			}

			unmarshalAssertionsVer2 := func(matrix *routing.CostMatrix, datacenterIDs []uint64, datacenterNames []string) {
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
				putVersionNumber(buff, &offset, 3)
				var matrix routing.CostMatrix

				err := matrix.UnmarshalBinary(buff)

				assert.EqualError(t, err, "unknown cost matrix version 3")
			})

			t.Run("version number == 0", func(t *testing.T) {
				relayAddrs := []string{"127.0.0.1", "127.0.0.2"}

				relayIDs := addrsToIDs(relayAddrs)

				numRelays := len(relayAddrs)

				publicKeys := [][]byte{core.RandomBytes(routing.LengthOfRelayToken), core.RandomBytes(routing.LengthOfRelayToken)}

				datacenters := []uint64{0, 1}

				numDatacenters := len(datacenters)

				datacenterRelays := [][]uint64{{core.GetRelayID("127.0.0.1")}, {core.GetRelayID("127.0.0.2")}}

				rtts := make([]int32, core.TriMatrixLength(numRelays))

				for i, _ := range rtts {
					rtts[i] = int32(rand.Int())
				}

				buff := make([]byte, 4+4+8*numRelays+routing.MaxRelayAddressLength*numRelays+routing.LengthOfRelayToken*numRelays+4+8*numDatacenters+4*numRelays+8*numRelays+4*len(rtts))

				offset := 0
				putVersionNumber(buff, &offset, 0)
				putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
				putRelayAddresses(buff, &offset, relayAddrs)
				putRelayPublicKeys(buff, &offset, publicKeys)
				putDatacenters(buff, &offset, datacenters, datacenterRelays)
				putRtts(buff, &offset, rtts)

				var matrix routing.CostMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(&matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, rtts)
			})

			t.Run("version number == 1", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{core.RandomBytes(routing.LengthOfRelayToken), core.RandomBytes(routing.LengthOfRelayToken)}
				datacenters := []uint64{0, 1}
				numDatacenters := len(datacenters)
				datacenterRelays := [][]uint64{{core.GetRelayID("127.0.0.1")}, {core.GetRelayID("127.0.0.2")}}
				rtts := make([]int32, core.TriMatrixLength(numRelays))

				for i, _ := range rtts {
					rtts[i] = int32(rand.Int())
				}

				// version 1 stuff
				relayNames := []string{"a name", "another name"}

				buff := make([]byte, 4+4+8*numRelays+routing.MaxRelayAddressLength*numRelays+routing.LengthOfRelayToken*numRelays+4+8*numDatacenters+4*numRelays+8*numRelays+4*len(rtts)+4+len(relayNames[0])+4+len(relayNames[1]))

				offset := 0
				putVersionNumber(buff, &offset, 1)
				putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames) //version >= 1
				putRelayAddresses(buff, &offset, relayAddrs)
				putRelayPublicKeys(buff, &offset, publicKeys)
				putDatacenters(buff, &offset, datacenters, datacenterRelays)
				putRtts(buff, &offset, rtts)

				var matrix routing.CostMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(&matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, rtts)
				unmarshalAssertionsVer1(&matrix, relayNames)
			})

			t.Run("version number >= 2", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{core.RandomBytes(routing.LengthOfRelayToken), core.RandomBytes(routing.LengthOfRelayToken)}
				datacenters := []uint64{0, 1}
				numDatacenters := len(datacenters)
				datacenterRelays := [][]uint64{{core.GetRelayID("127.0.0.1")}, {core.GetRelayID("127.0.0.2")}}
				rtts := make([]int32, core.TriMatrixLength(numRelays))

				for i, _ := range rtts {
					rtts[i] = int32(rand.Int())
				}

				// version 1 stuff
				relayNames := []string{"a name", "another name"}

				// version 2 stuff
				// resusing datacenters for the ID array
				datacenterNames := []string{"a datacenter", "another datacenter"}

				buff := make([]byte, 4+4+8*numRelays+routing.MaxRelayAddressLength*numRelays+routing.LengthOfRelayToken*numRelays+4+8*numDatacenters+4*numRelays+8*numRelays+4*len(rtts)+4+len(relayNames[0])+4+len(relayNames[1])+4+8+4+len(datacenterNames[0])+8+4+len(datacenterNames[1]))

				offset := 0
				putVersionNumber(buff, &offset, 2)
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
				unmarshalAssertionsVer0(&matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, rtts)
				unmarshalAssertionsVer1(&matrix, relayNames)
				unmarshalAssertionsVer2(&matrix, datacenters, datacenterNames)
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

				matrix.RelayAddresses = make([][]byte, 2)
				matrix.RelayAddresses[0] = core.RandomBytes(routing.MaxRelayAddressLength)
				matrix.RelayAddresses[1] = core.RandomBytes(routing.MaxRelayAddressLength)

				matrix.RelayPublicKeys = make([][]byte, 2)
				matrix.RelayPublicKeys[0] = core.RandomBytes(routing.LengthOfRelayToken)
				matrix.RelayPublicKeys[1] = core.RandomBytes(routing.LengthOfRelayToken)

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

				// essentialy this asserts the result of MarshalBinary(),
				// if the Unmarshal test passes then the result of Marshal
				// should work
				other.UnmarshalBinary(bin)

				assert.Nil(t, err)
				assert.Equal(t, matrix, other)
			})
		})
	})
}
