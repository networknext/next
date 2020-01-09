package core_test

import (
	"math/rand"
	"time"

	"github.com/networknext/backend/core"
)

func RandomPublicKey() []byte {
	arr := make([]byte, 64)
	for i := 0; i < 64; i++ {
		arr[i] = byte(rand.Int())
	}
	return arr
}

func FillRelayDatabase(relaydb *core.RelayDatabase) {
	fillData := func(relaydb *core.RelayDatabase, addr string, updateTime int64) {
		id := core.GetRelayID(addr)
		data := core.RelayData{
			ID:             id,
			Name:           addr,
			Address:        addr,
			Datacenter:     core.DatacenterId(0),
			DatacenterName: "n/a",
			PublicKey:      RandomPublicKey(),
			LastUpdateTime: uint64(updateTime),
		}
		relaydb.Relays[id] = data
	}

	fillData(relaydb, "127.0.0.1", time.Now().Unix()-1)
	fillData(relaydb, "123.4.5.6", time.Now().Unix()-10)
	fillData(relaydb, "654.3.2.1", time.Now().Unix()-100)
	fillData(relaydb, "000.0.0.0", time.Now().Unix()-25)
	fillData(relaydb, "999.9.9.9", time.Now().Unix()-1000)
}

func FillStatsDatabase(statsdb *core.StatsDatabase) {
	makeEntry := func(statsdb *core.StatsDatabase, addr string, conns ...string) {
		entry := core.NewStatsEntry()
		makeStats := func(entry *core.StatsEntry, addr string) {
			stats := core.NewStatsEntryRelay()
			entry.Relays[core.GetRelayID(addr)] = stats
		}

		for _, c := range conns {
			makeStats(entry, c)
		}

		statsdb.Entries[core.GetRelayID(addr)] = *entry
	}

	makeEntry(statsdb, "127.0.0.1", "127.0.0.2", "123.4.5.6", "654.3.2.1", "000.0.0.0", "999.9.9.9")
	makeEntry(statsdb, "127.0.0.2", "127.0.0.1", "123.4.5.6", "654.3.2.1", "000.0.0.0", "999.9.9.9")
}
