package routing_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestHistory(t *testing.T) {
	t.Run("HistoryMax()", func(t *testing.T) {
		t.Run("returns the max value in the array", func(t *testing.T) {
			history := []float32{1, 2, 3, 4, 5, 4, 3, 2, 1}
			assert.Equal(t, float32(5), routing.HistoryMax(history))
		})
	})

	t.Run("HistoryNotSet()", func(t *testing.T) {
		t.Run("returns a []float32 the size of HistorySize containing nothing but InvalidHistoryValue", func(t *testing.T) {
			history := routing.HistoryNotSet()
			assert.Len(t, history, routing.HistorySize)
			for i := 0; i < routing.HistorySize; i++ {
				assert.Equal(t, routing.HistoryInvalidValue, int(history[i]))
			}
		})
	})

	t.Run("HistoryMean()", func(t *testing.T) {
		t.Run("returns a float32 that is the average of the input", func(t *testing.T) {
			history := []float32{1, 2, 3, 4, 5, 4, 3, 2, 1}
			assert.Equal(t, float32(25.0/9.0), routing.HistoryMean(history))
		})
	})

	t.Run("Old tests", func(t *testing.T) {
		t.Run("TestHistoryMax()", func(t *testing.T) {

			t.Parallel()

			history := routing.HistoryNotSet()

			assert.Equal(t, float32(0.0), routing.HistoryMax(history[:]))

			history[0] = 5.0
			history[1] = 3.0
			history[2] = 100.0

			assert.Equal(t, float32(100.0), routing.HistoryMax(history[:]))
		})

		t.Run("TestHistoryMean", func(t *testing.T) {

			t.Parallel()

			history := routing.HistoryNotSet()

			assert.Equal(t, float32(0.0), routing.HistoryMean(history[:]))

			history[0] = 5.0
			history[1] = 3.0
			history[2] = 100.0

			assert.Equal(t, float32(36.0), routing.HistoryMean(history[:]))
		})
	})
}

func TestStatsDatabase(t *testing.T) {
	sourceId := routing.GetRelayID("127.0.0.1")
	relay1Id := routing.GetRelayID("127.9.9.9")
	relay2Id := routing.GetRelayID("999.999.9.9")

	makeBasicStats := func() *routing.StatsEntryRelay {
		entry := routing.NewStatsEntryRelay()
		entry.Index = 1
		entry.Rtt = 1
		entry.Jitter = 1
		entry.PacketLoss = 1
		entry.RttHistory[0] = 1
		entry.JitterHistory[0] = 1
		entry.PacketLossHistory[0] = 1
		return entry
	}

	t.Run("ProcessStats()", func(t *testing.T) {
		update := routing.RelayStatsUpdate{
			ID: sourceId,
			PingStats: []routing.RelayStatsPing{
				routing.RelayStatsPing{
					RelayID:    relay1Id,
					RTT:        0.5,
					Jitter:     0.7,
					PacketLoss: 0.9,
				},
				routing.RelayStatsPing{
					RelayID:    relay2Id,
					RTT:        0,
					Jitter:     0,
					PacketLoss: 0,
				},
			},
		}

		t.Run("if the entry for the source relay does not exist, it is created properly", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			statsdb.ProcessStats(&update)
			entry, ok := statsdb.Entries[update.ID]
			assert.True(t, ok)
			assert.NotNil(t, entry.Relays)
		})

		t.Run("if the entry for the destination relay does not exist in the source relay's collection, then it is created properly", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			statsdb.ProcessStats(&update)
			entry, _ := statsdb.Entries[update.ID]
			for _, r := range entry.Relays {
				assert.NotNil(t, r.RttHistory)
				assert.NotNil(t, r.JitterHistory)
				assert.NotNil(t, r.PacketLossHistory)
			}
		})

		t.Run("source and destination both don't exist but entries are updated properly", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			statsdb.ProcessStats(&update)
			entry, _ := statsdb.Entries[update.ID]
			assert.Equal(t, len(entry.Relays), len(update.PingStats))
			for _, stats := range update.PingStats {
				destRelay, _ := entry.Relays[stats.RelayID]
				assert.Equal(t, stats.RTT, destRelay.Rtt)
				assert.Equal(t, stats.Jitter, destRelay.Jitter)
				assert.Equal(t, stats.PacketLoss, destRelay.PacketLoss)
				assert.Equal(t, stats.RTT, destRelay.RttHistory[0])
				assert.Equal(t, stats.Jitter, destRelay.JitterHistory[0])
				assert.Equal(t, stats.PacketLoss, destRelay.PacketLossHistory[0])
			}
		})

		t.Run("source and destination do exist and their entries are updated", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			entry := routing.NewStatsEntry()
			statsEntry1 := makeBasicStats()
			entry.Relays[relay1Id] = statsEntry1
			statsdb.Entries[update.ID] = *entry

			statsdb.ProcessStats(&update)

			assert.Equal(t, 2, statsEntry1.Index)
			assert.Equal(t, float32(0.75), statsEntry1.Rtt)
			assert.Equal(t, float32(0.85), statsEntry1.Jitter)
			assert.Equal(t, float32(0.95), statsEntry1.PacketLoss)
		})
	})

	t.Run("MakeCopy()", func(t *testing.T) {
		t.Run("makes a copy", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			entry := routing.NewStatsEntry()
			statsEntry1 := makeBasicStats()
			entry.Relays[relay1Id] = statsEntry1
			statsdb.Entries[sourceId] = *entry

			cpy := statsdb.MakeCopy()

			assert.Equal(t, statsdb, cpy)
		})
	})

	t.Run("GetEntry()", func(t *testing.T) {
		t.Run("entry does not exist", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()

			stats := statsdb.GetEntry(sourceId, relay1Id)

			assert.Nil(t, stats)
		})

		t.Run("entry exists but stats for the internal entry does not", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			entry := routing.NewStatsEntry()
			statsdb.Entries[sourceId] = *entry

			stats := statsdb.GetEntry(sourceId, relay1Id)

			assert.Nil(t, stats)
		})

		t.Run("both the entry and the internal entry exist", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			entry := routing.NewStatsEntry()
			// this entry makes no sense, test puposes only
			statsEntry1 := makeBasicStats()
			entry.Relays[relay1Id] = statsEntry1
			statsdb.Entries[sourceId] = *entry

			stats := statsdb.GetEntry(sourceId, relay1Id)

			assert.Equal(t, statsEntry1, stats)
		})
	})

	t.Run("GetSample()", func(t *testing.T) {
		id1 := sourceId
		id2 := relay1Id

		makeBasicConnection := func(statsdb *routing.StatsDatabase, id1, id2 uint64) {
			stats := routing.NewStatsEntryRelay()
			entry := routing.NewStatsEntry()
			entry.Relays[id2] = stats
			statsdb.Entries[id1] = *entry
		}

		t.Run("relay 1 -> relay 2 does not exist", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			statsdb.Entries[id1] = *routing.NewStatsEntry()
			makeBasicConnection(statsdb, id2, id1)

			rtt, jitter, packetLoss := statsdb.GetSample(id1, id2)

			assert.Equal(t, float32(routing.InvalidRouteValue), rtt)
			assert.Equal(t, float32(routing.InvalidRouteValue), jitter)
			assert.Equal(t, float32(routing.InvalidRouteValue), packetLoss)
		})

		t.Run("relay 2 -> relay 1 does not exist", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			statsdb.Entries[id2] = *routing.NewStatsEntry()
			makeBasicConnection(statsdb, id1, id2)

			rtt, jitter, packetLoss := statsdb.GetSample(id1, id2)

			assert.Equal(t, float32(routing.InvalidRouteValue), rtt)
			assert.Equal(t, float32(routing.InvalidRouteValue), jitter)
			assert.Equal(t, float32(routing.InvalidRouteValue), packetLoss)
		})

		t.Run("both relays are valid", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			entryForId1 := routing.NewStatsEntry()
			entryForId2 := routing.NewStatsEntry()
			statsForId1 := routing.NewStatsEntryRelay()
			statsForId2 := routing.NewStatsEntryRelay()

			statsForId1.Rtt = 1000.0
			statsForId1.Jitter = 12.345
			statsForId1.PacketLoss = 987.654

			statsForId2.Rtt = 999.99
			statsForId2.Jitter = 13.0
			statsForId2.PacketLoss = 989.0

			entryForId1.Relays[id2] = statsForId2
			entryForId2.Relays[id1] = statsForId1

			statsdb.Entries[id1] = *entryForId1
			statsdb.Entries[id2] = *entryForId2

			rtt, jitter, packetLoss := statsdb.GetSample(id1, id2)

			assert.Equal(t, float32(1000.0), rtt)
			assert.Equal(t, float32(13.0), jitter)
			assert.Equal(t, float32(989.0), packetLoss)
		})
	})

	t.Run("GetCostMatrix()", func(t *testing.T) {
		t.Run("returns the cost matrix", func(t *testing.T) {
			redisServer, _ := miniredis.Run()
			redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

			statsdb := routing.NewStatsDatabase()

			// Setup

			FillRelayDatabase(redisClient)
			FillStatsDatabase(statsdb)

			// make the datacenter of the first relay 0
			// otherwise push the rest into the validDcIDs array
			// same with Dc names
			validDcIDs := make([]uint64, 0)
			validDcNames := make([]string, 0)

			hgetallResult := redisClient.HGetAll(routing.RedisHashName)

			i := 0
			for _, raw := range hgetallResult.Val() {
				var r routing.Relay
				r.UnmarshalBinary([]byte(raw))
				if i == 0 {
					r.Datacenter = 0
					redisClient.HSet(routing.RedisHashName, r.Key(), r)
				} else {
					r.Datacenter = uint64(i)
					redisClient.HSet(routing.RedisHashName, r.Key(), r)
					validDcIDs = append(validDcIDs, r.Datacenter)
					validDcNames = append(validDcNames, r.DatacenterName)
				}
				i++
			}

			modifyEntry := func(addr1, addr2 string, rtt, jitter, packetloss float32) {
				entry := statsdb.GetEntry(routing.GetRelayID(addr1), routing.GetRelayID(addr2))
				entry.Rtt = rtt
				entry.Jitter = jitter
				entry.PacketLoss = packetloss
			}

			// valid
			modifyEntry("127.0.0.1:40000", "127.0.0.2:40000", 123.00, 1.3, 0.0)

			// invalid - rtt = invalid route value
			modifyEntry("127.0.0.1:40000", "127.0.0.3:40000", routing.InvalidRouteValue, 0.3, 0.0)

			// invalid - jitter > MaxJitter
			modifyEntry("127.0.0.1:40000", "127.0.0.4:40000", 1.0, routing.MaxJitter+1, 0.0)

			// invalid - packet loss > MaxPacketLoss
			modifyEntry("127.0.0.1:40000", "127.0.0.5:40000", 1.0, 0.3, routing.MaxPacketLoss+1)

			var costMatrix routing.CostMatrix
			assert.Nil(t, statsdb.GetCostMatrix(&costMatrix, redisClient))

			// Testing
			hgetallResult = redisClient.HGetAll(routing.RedisHashName)

			// assert each entry in the relay db is present in the cost matrix
			for _, raw := range hgetallResult.Val() {
				var relay routing.Relay
				relay.UnmarshalBinary([]byte(raw))
				assert.Contains(t, costMatrix.RelayIds, relay.ID)
				assert.Contains(t, costMatrix.RelayNames, relay.Name)
				assert.Contains(t, costMatrix.RelayPublicKeys, relay.PublicKey)
			}

			// assert the length of the valid ids equals the length of all the datacenter ids in the matrix
			assert.Equal(t, len(validDcIDs), len(costMatrix.DatacenterIds))
			assert.Equal(t, len(validDcNames), len(costMatrix.DatacenterNames))
			for i, id := range validDcIDs {
				// assert all valid ids are present in the matrix
				assert.Contains(t, costMatrix.DatacenterIds, id)

				// find the relays whose datacenter id matches this one
				validRelayIDs := make([]uint64, 0)
				for _, raw := range hgetallResult.Val() {
					var relay routing.Relay
					relay.UnmarshalBinary([]byte(raw))
					if relay.Datacenter == id {
						validRelayIDs = append(validRelayIDs, relay.ID)
						break
					}
				}

				// assert the datacenter id -> relay ids mapping contains the actual ids
				// i + 1 because in the first for loop each datacenter id is reset as i which is > 0
				for _, relayID := range validRelayIDs {
					assert.Contains(t, costMatrix.DatacenterRelays[uint64(i+1)], relayID)
				}
			}

			// assert all names valid Dc names are within the cost matrix
			for _, name := range validDcNames {
				assert.Contains(t, costMatrix.DatacenterNames, name)
			}

			// assert that all non-invalid rtt's are within the cost matrix
			getAddressIndex := func(addr1, addr2 string) int {
				addr1ID := routing.GetRelayID(addr1)
				addr2ID := routing.GetRelayID(addr2)
				indxOfI := -1
				indxOfJ := -1

				for i, id := range costMatrix.RelayIds {
					if uint64(id) == addr1ID {
						indxOfI = i
					} else if uint64(id) == addr2ID {
						indxOfJ = i
					}

					if indxOfI != -1 && indxOfJ != -1 {
						break
					}
				}

				return routing.TriMatrixIndex(indxOfI, indxOfJ)
			}

			assert.Equal(t, int32(124), costMatrix.RTT[getAddressIndex("127.0.0.1:40000", "127.0.0.2:40000")])
			assert.Equal(t, int32(-1), costMatrix.RTT[getAddressIndex("127.0.0.1:40000", "127.0.0.3:40000")])
			assert.Equal(t, int32(-1), costMatrix.RTT[getAddressIndex("127.0.0.1:40000", "654.0.0.4:40000")])
			assert.Equal(t, int32(-1), costMatrix.RTT[getAddressIndex("127.0.0.1:40000", "000.0.0.5:40000")])
		})
	})
}
