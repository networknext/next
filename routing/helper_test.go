package routing_test

import (
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
