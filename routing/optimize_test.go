package routing_test

import "testing"

import "encoding/binary"

import "github.com/networknext/backend/routing"

import "github.com/stretchr/testify/assert"

import "github.com/networknext/backend/core"

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

func putDatacenter(buff []byte, offset *int, datacenterIDs []uint64, relayIDs [][]uint64) {
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

				numRelays := len(relayAddrs)

				publicKeys := [][]byte{core.RandomBytes(routing.LengthOfRelayToken), core.RandomBytes(routing.LengthOfRelayToken)}

				datacenters := []uint64{0, 1}

				numDatacenters := len(datacenters)

				datacenterRelays := [][]uint64{{core.GetRelayID("127.0.0.1")}, {core.GetRelayID("127.0.0.2")}}

				rtts := []int32{-1, 1}

				buff := make([]byte, 4+4+8*numRelays+routing.MaxRelayAddressLength*numRelays+routing.LengthOfRelayToken*numRelays+4+8*numDatacenters+4*numRelays+8*numRelays+4*len(rtts))

				offset := 0
				putVersionNumber(buff, &offset, 0)
				putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
				putRelayAddresses(buff, &offset, relayAddrs)
				putRelayPublicKeys(buff, &offset, publicKeys)
				putDatacenter(buff, &offset, datacenters, datacenterRelays)
				putRtts(buff, &offset, rtts)

				var matrix routing.CostMatrix

				err := matrix.UnmarshalBinary(buff)

				assert.Nil(t, err)
			})

			t.Run("version number == 1", func(t *testing.T) {

			})

			t.Run("version number >= 2", func(t *testing.T) {

			})
		})
	})
}
