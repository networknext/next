package common

import (
	"math"
	"sync"
	"time"
)

// todo: might want to put a lastUpdateTime in the sourceEntry as well, then we can extract set of active relays easily
// (relays that have posted an update in the last 10 seconds)

// todo: might want to look at RTT variation across the 5 minutes, in addition to jitter. jitter is only across 1 second of pings (10 samples)

const HistorySize = 300 // 5 minutes @ one relay update per-second

const InvalidRouteValue = float32(1000000000.0)

func TriMatrixLength(size int) int {
	return (size * (size - 1)) / 2
}

func TriMatrixIndex(i, j int) int {
	if i <= j {
		i, j = j, i
	}
	return i*(i+1)/2 - i + j
}

func historyMean(history []float32) float32 {
	var sum float64
	for i := 0; i < len(history); i++ {
		sum += float64(history[i])
	}
	return float32(sum / float64(HistorySize))
}

type RelayStatsDestEntry struct {
	lastUpdateTime    time.Time
	rtt               float32
	jitter            float32
	packetLoss        float32
	historyIndex      int32
	historyRTT        [HistorySize]float32
	historyJitter     [HistorySize]float32
	historyPacketLoss [HistorySize]float32
}

type RelayStatsSourceEntry struct {
	mutex       sync.RWMutex
	destEntries map[uint64]*RelayStatsDestEntry
}

type RelayStats struct {
	mutex         sync.RWMutex
	sourceEntries map[uint64]*RelayStatsSourceEntry
}

func CreateRelayStats() *RelayStats {
	relayStats := &RelayStats{}
	relayStats.sourceEntries = make(map[uint64]*RelayStatsSourceEntry)
	return relayStats
}

func (relayStats *RelayStats) ProcessRelayUpdate(sourceRelayId uint64, numSamples int, sampleRelayId []uint64, sampleRTT []float32, sampleJitter []float32, samplePacketLoss []float32) {

	/*
		Process Relay Update
		--------------------

		Our goal is to get stable RTT, Jitter and PL values to feed into our route optimization algorithm,
		so we can send traffic across stable routes that aren't subject to change.

		To achieve this, this function processes a "relay update" sent from a relay and stores it
		in a data structure use to generate RTT, Jitter and Packet Loss values per-relay pair (source,dest)
		to feed into our route optimization algorithm.

		The relay update contains samples the relay has derived by pinging n other relays for roughly one second.

		Each sample contains:

			1. The id of the relay being pinged.

			1. RTT (minimum RTT seen over the last second)

			2. Jitter (one standard deviation of jitter, relative to min RTT over one second)

			3. Packet Loss (%)

		To achieve this we have a ~5 minute history buffer per (source,dest) relay pair, and we take the
		average RTT, Jitter and PL values across this history.

		In addition, we "poison" the history buffer for any newly seen relay pair by setting RTT, Jitter
		and Packet Loss values very high, so we won't route any traffic across it, until it has proven
		itself to be stable for at least 5 minutes.

		ps. The data structure used to store relay stats is designed primarily to minimize lock contention.
	*/

	// look up the entry corresponding to the source relay, or create it if it doesn't exist

	relayStats.mutex.Lock()

	sourceEntry, exists := relayStats.sourceEntries[sourceRelayId]
	if !exists {
		sourceEntry = &RelayStatsSourceEntry{}
		sourceEntry.destEntries = make(map[uint64]*RelayStatsDestEntry)
		relayStats.sourceEntries[sourceRelayId] = sourceEntry
	}

	relayStats.mutex.Unlock()

	// iterate across all samples and insert them into the history buffer
	// in the dest entry corresponding to their relay pair (source,dest)

	currentTime := time.Now()

	sourceEntry.mutex.Lock()

	for i := 0; i < numSamples; i++ {

		destRelayId := sampleRelayId[i]

		destEntry, exists := sourceEntry.destEntries[destRelayId]
		if !exists {
			destEntry = &RelayStatsDestEntry{}
			sourceEntry.destEntries[destRelayId] = destEntry
		}

		// IMPORTANT: clear newly created AND timed out dest entry history buffers
		// this is important so that newly created relays, and relays that are stopped and restarted
		// don't get routed across, until at least 5 minutes has passed!

		if currentTime.Sub(destEntry.lastUpdateTime) > 10*time.Second {
			for j := 0; j < HistorySize; j++ {
				destEntry.historyIndex = 0
				destEntry.historyRTT[j] = InvalidRouteValue
				destEntry.historyJitter[j] = InvalidRouteValue
				destEntry.historyPacketLoss[j] = InvalidRouteValue
			}
		}

		destEntry.historyRTT[destEntry.historyIndex] = sampleRTT[i]
		destEntry.historyJitter[destEntry.historyIndex] = sampleJitter[i]
		destEntry.historyPacketLoss[destEntry.historyIndex] = samplePacketLoss[i]

		destEntry.rtt = historyMean(destEntry.historyRTT[:])
		destEntry.jitter = historyMean(destEntry.historyJitter[:])
		destEntry.packetLoss = historyMean(destEntry.historyPacketLoss[:])

		destEntry.historyIndex = (destEntry.historyIndex + 1) % HistorySize
	}

	sourceEntry.mutex.Unlock()
}

func (relayStats *RelayStats) GetSample(currentTime time.Time, sourceRelayId uint64, destRelayId uint64) (float32, float32, float32) {

	sourceRTT := InvalidRouteValue
	sourceJitter := InvalidRouteValue
	sourcePacketLoss := InvalidRouteValue

	destRTT := InvalidRouteValue
	destJitter := InvalidRouteValue
	destPacketLoss := InvalidRouteValue

	// get source ping values
	{
		relayStats.mutex.RLock()
		sourceEntry, exists := relayStats.sourceEntries[sourceRelayId]
		if exists {
			sourceEntry.mutex.RLock()
			destEntry, exists := sourceEntry.destEntries[destRelayId]
			if exists {
				if currentTime.Sub(destEntry.lastUpdateTime) < 10*time.Second {
					sourceRTT = destEntry.rtt
					sourceJitter = destEntry.jitter
					sourcePacketLoss = destEntry.packetLoss
				}
			}
			sourceEntry.mutex.RUnlock()
		}
		relayStats.mutex.RUnlock()
	}

	// get dest ping values
	{
		relayStats.mutex.RLock()
		sourceEntry, exists := relayStats.sourceEntries[destRelayId]
		if exists {
			sourceEntry.mutex.RLock()
			destEntry, exists := sourceEntry.destEntries[sourceRelayId]
			if exists {
				if currentTime.Sub(destEntry.lastUpdateTime) < 10*time.Second {
					destRTT = destEntry.rtt
					destJitter = destEntry.jitter
					destPacketLoss = destEntry.packetLoss
				}
			}
			sourceEntry.mutex.RUnlock()
		}
		relayStats.mutex.RUnlock()
	}

	// take maximum values in each direction

	rtt := sourceRTT
	jitter := sourceJitter
	packetLoss := sourcePacketLoss

	if destRTT > rtt {
		rtt = destRTT
	}

	if destJitter > jitter {
		jitter = destJitter
	}

	if destPacketLoss > packetLoss {
		packetLoss = destPacketLoss
	}

	return rtt, jitter, packetLoss
}

func (relayStats *RelayStats) GetCosts(relayIds []uint64, maxRTT float32, maxJitter float32, maxPacketLoss float32, local bool) []int32 {

	numRelays := len(relayIds)

	costs := make([]int32, TriMatrixLength(numRelays))

	// IMPORTANT: special permissive route matrix for local env only
	if local {
		return costs
	}

	currentTime := time.Now()

	for i := 0; i < numRelays; i++ {
		sourceRelayId := uint64(relayIds[i])
		for j := 0; j < i; j++ {
			index := TriMatrixIndex(i, j)
			destRelayId := uint64(relayIds[j])
			rtt, jitter, packetLoss := relayStats.GetSample(currentTime, sourceRelayId, destRelayId)
			if rtt < maxRTT && jitter < maxJitter && packetLoss < maxPacketLoss {
				costs[index] = int32(math.Ceil(float64(rtt)))
			} else {
				costs[index] = -1
			}
		}
	}

	return costs
}
