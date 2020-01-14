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
		t.Run("MarshalBinary -> UnmarshalBinary equality", func(t *testing.T) {
			t.Skip()
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

			matrix.RTT = make([]int32, 2)
			matrix.RTT[0] = 7
			matrix.RTT[1] = 13

			var other routing.CostMatrix

			bin, _ := matrix.MarshalBinary()
			other.UnmarshalBinary(bin)

			//assert.Equal(t, matrix, other)
		})

		t.Run("UnmarshalBinary()", func(t *testing.T) {
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

				assert.Len(t, matrix.RelayIds, numRelays)
				assert.Len(t, matrix.RelayNames, 0)
				assert.Len(t, matrix.RelayAddresses, numRelays)
				assert.Len(t, matrix.RelayPublicKeys, numRelays)
				assert.Len(t, matrix.DatacenterIds, 0)
				assert.Len(t, matrix.DatacenterNames, 0)
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
			})

			t.Run("version number == 1", func(t *testing.T) {

			})

			t.Run("version number >= 2", func(t *testing.T) {

			})
		})
	})
}
