package routing_test

import (
	"math"
	"math/rand"
	"net"
	"time"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
)

func RandomPublicKey() []byte {
	arr := make([]byte, crypto.KeySize)
	rand.Read(arr)
	return arr
}

func RandomString(length int) string {
	arr := make([]byte, length)
	for i := 0; i < length; i++ {
		arr[i] = byte(rand.Int()%26 + 65)
	}
	return string(arr)
}

func FillRelayDatabase(relaydb *routing.RelayDatabase) {
	fillData := func(relaydb *routing.RelayDatabase, addr string, updateTime int64) {
		id := routing.GetRelayID(addr)
		udp, _ := net.ResolveUDPAddr("udp", addr)
		data := routing.Relay{
			ID:             id,
			Name:           addr,
			Addr:           *udp,
			Datacenter:     uint64(rand.Uint64()%(math.MaxUint64-1) + 1), // non-zero random number
			DatacenterName: RandomString(5),
			PublicKey:      RandomPublicKey(),
			LastUpdateTime: uint64(updateTime),
		}
		relaydb.Relays[id] = data
	}

	fillData(relaydb, "127.0.0.1:40000", time.Now().Unix()-1)
	fillData(relaydb, "127.0.0.2:40000", time.Now().Unix()-5)
	fillData(relaydb, "127.0.0.3:40000", time.Now().Unix()-10)
	fillData(relaydb, "127.0.0.4:40000", time.Now().Unix()-100)
	fillData(relaydb, "127.0.0.5:40000", time.Now().Unix()-25)
	fillData(relaydb, "127.0.0.6:40000", time.Now().Unix()-1000)
}

func FillStatsDatabase(statsdb *routing.StatsDatabase) {
	makeEntry := func(statsdb *routing.StatsDatabase, addr string, conns ...string) {
		entry := routing.NewStatsEntry()
		makeStats := func(entry *routing.StatsEntry, addr string) {
			stats := routing.NewStatsEntryRelay()
			entry.Relays[routing.GetRelayID(addr)] = stats
		}

		for _, c := range conns {
			makeStats(entry, c)
		}

		statsdb.Entries[routing.GetRelayID(addr)] = *entry
	}

	makeEntry(statsdb, "127.0.0.1:40000", "127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000", "127.0.0.6:40000")
	makeEntry(statsdb, "127.0.0.2:40000", "127.0.0.1:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000", "127.0.0.6:40000")
}
