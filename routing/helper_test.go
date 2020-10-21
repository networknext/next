package routing_test

import (
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
	"net"
	"time"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
)

func randomPublicKey() []byte {
	arr := make([]byte, crypto.KeySize)
	rand.Read(arr)
	return arr
}

func randomString(length int) string {
	arr := make([]byte, length)
	for i := 0; i < length; i++ {
		arr[i] = byte(rand.Int()%26 + 65)
	}
	return string(arr)
}

func fillRelayDatabase(relayMap *routing.RelayMap) {
	fillData := func(addr string, updateTime time.Time) {
		id := crypto.HashID(addr)
		udp, _ := net.ResolveUDPAddr("udp", addr)
		data := &routing.RelayData{
			ID:   id,
			Name: addr,
			Addr: *udp,
			Datacenter: routing.Datacenter{
				ID:   uint64(rand.Uint64()%(math.MaxUint64-1) + 1), // non-zero random number
				Name: randomString(5),
			},
			PublicKey:      randomPublicKey(),
			LastUpdateTime: updateTime,
		}
		relayMap.AddRelayDataEntry(data.Addr.String(), data)
	}

	fillData("127.0.0.1:40000", time.Now().Add(time.Second*-1))
	fillData("127.0.0.2:40000", time.Now().Add(time.Second*-5))
	fillData("127.0.0.3:40000", time.Now().Add(time.Second*-10))
	fillData("127.0.0.4:40000", time.Now().Add(time.Second*-100))
	fillData("127.0.0.5:40000", time.Now().Add(time.Second*-25))
	fillData("127.0.0.6:40000", time.Now().Add(time.Second*-1000))
}

func fillStatsDatabase(statsdb *routing.StatsDatabase) {
	makeEntry := func(statsdb *routing.StatsDatabase, addr string, conns ...string) {
		entry := routing.NewStatsEntry()
		makeStats := func(entry *routing.StatsEntry, addr string) {
			stats := routing.NewStatsEntryRelay()
			entry.Relays[crypto.HashID(addr)] = stats
		}

		for _, c := range conns {
			makeStats(entry, c)
		}

		statsdb.Entries[crypto.HashID(addr)] = *entry
	}

	makeEntry(statsdb, "127.0.0.1:40000", "127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000", "127.0.0.6:40000")
	makeEntry(statsdb, "127.0.0.2:40000", "127.0.0.1:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000", "127.0.0.6:40000")
}

// A buffer type that implements io.Write and io.Read but always returns an error for testing
type ErrorBuffer struct {
}

func (*ErrorBuffer) Write(p []byte) (int, error) {
	return 0, errors.New("descriptive error")
}

func (*ErrorBuffer) Read(p []byte) (int, error) {
	return 0, errors.New("descriptive error")
}

func addrsToIDs(addrs []string) []uint64 {
	retval := make([]uint64, len(addrs))
	for i, addr := range addrs {
		retval[i] = uint64(crypto.HashID(addr))
	}
	return retval
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

func putFloat64s(buff []byte, offset *int, nums ...float64) {
	for _, num := range nums {
		uintNum := math.Float64bits(num)
		binary.LittleEndian.PutUint64(buff[*offset:], uintNum)
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

func putRelayLatitudes(buff []byte, offset *int, latitudes []float64) {
	putFloat64s(buff, offset, latitudes...)
}

func putRelayLongitudes(buff []byte, offset *int, latitudes []float64) {
	putFloat64s(buff, offset, latitudes...)
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
		putUint64s(buff, offset, uint64(seller.IngressPriceNibblinsPerGB))
		putUint64s(buff, offset, uint64(seller.EgressPriceNibblinsPerGB))
	}
}

func putSessionCounts(buf []byte, offset *int, sessionCounts []uint32) {
	putUint32s(buf, offset, sessionCounts...)
}

func putMaxSessionCounts(buf []byte, offset *int, maxSessionCounts []uint32) {
	putUint32s(buf, offset, maxSessionCounts...)
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

func sizeofRelayLatitudes(latitudes []float64) int {
	return len(latitudes) * 8
}

func sizeofRelayLongitudes(longitudes []float64) int {
	return len(longitudes) * 8
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

func sizeofSessionCounts(sessionCounts []uint32) int {
	return 4 * len(sessionCounts)
}

func sizeofMaxSessionCounts(maxSessionCounts []uint32) int {
	return 4 * len(maxSessionCounts)
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
