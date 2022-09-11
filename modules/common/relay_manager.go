package common

import (
	"fmt"
	"math"
	"net"
	"sort"
	"sync"
	"time"
)

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

type RelayManagerDestEntry struct {
	lastUpdateTime    time.Time
	rtt               float32
	jitter            float32
	packetLoss        float32
	historyIndex      int32
	historyRTT        [HistorySize]float32
	historyJitter     [HistorySize]float32
	historyPacketLoss [HistorySize]float32
}

type RelayManagerSourceEntry struct {
	mutex          sync.RWMutex
	lastUpdateTime time.Time
	relayId        uint64
	relayName      string
	relayAddress   net.UDPAddr
	sessions       int
	relayVersion   string
	shuttingDown   bool
	destEntries    map[uint64]*RelayManagerDestEntry
}

type RelayManager struct {
	mutex         sync.RWMutex
	sourceEntries map[uint64]*RelayManagerSourceEntry
}

func CreateRelayManager() *RelayManager {
	relayManager := &RelayManager{}
	relayManager.sourceEntries = make(map[uint64]*RelayManagerSourceEntry)
	return relayManager
}

func (relayManager *RelayManager) ProcessRelayUpdate(relayId uint64, relayName string, relayAddress net.UDPAddr, sessions int, relayVersion string, shuttingDown bool, numSamples int, sampleRelayId []uint64, sampleRTT []float32, sampleJitter []float32, samplePacketLoss []float32) {

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

	relayManager.mutex.Lock()

	sourceEntry, exists := relayManager.sourceEntries[relayId]
	if !exists {
		sourceEntry = &RelayManagerSourceEntry{}
		sourceEntry.destEntries = make(map[uint64]*RelayManagerDestEntry)
		relayManager.sourceEntries[relayId] = sourceEntry
	}

	relayManager.mutex.Unlock()

	// update stats for the source relay, then...
	// iterate across all samples and insert them into the history buffer
	// in the dest entry corresponding to their relay pair (source,dest)

	currentTime := time.Now()

	sourceEntry.mutex.Lock()

	sourceEntry.lastUpdateTime = currentTime
	sourceEntry.relayId = relayId
	sourceEntry.relayName = relayName
	sourceEntry.relayAddress = relayAddress
	sourceEntry.sessions = sessions
	sourceEntry.relayVersion = relayVersion
	sourceEntry.shuttingDown = shuttingDown

	for i := 0; i < numSamples; i++ {

		destRelayId := sampleRelayId[i]

		destEntry, exists := sourceEntry.destEntries[destRelayId]
		if !exists {
			destEntry = &RelayManagerDestEntry{}
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

		destEntry.lastUpdateTime = currentTime
	}

	sourceEntry.mutex.Unlock()
}

func (relayManager *RelayManager) GetSample(currentTime time.Time, sourceRelayId uint64, destRelayId uint64) (float32, float32, float32) {

	sourceRTT := InvalidRouteValue
	sourceJitter := InvalidRouteValue
	sourcePacketLoss := InvalidRouteValue

	destRTT := InvalidRouteValue
	destJitter := InvalidRouteValue
	destPacketLoss := InvalidRouteValue

	// get source ping values
	{
		relayManager.mutex.RLock()
		sourceEntry, exists := relayManager.sourceEntries[sourceRelayId]
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
		relayManager.mutex.RUnlock()
	}

	// get dest ping values
	{
		relayManager.mutex.RLock()
		sourceEntry, exists := relayManager.sourceEntries[destRelayId]
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
		relayManager.mutex.RUnlock()
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

func (relayManager *RelayManager) GetCosts(relayIds []uint64, maxRTT float32, maxJitter float32, maxPacketLoss float32, local bool) []int32 {

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
			rtt, jitter, packetLoss := relayManager.GetSample(currentTime, sourceRelayId, destRelayId)
			if rtt < maxRTT && jitter < maxJitter && packetLoss < maxPacketLoss {
				costs[index] = int32(math.Ceil(float64(rtt)))
			} else {
				costs[index] = -1
			}
		}
	}

	return costs
}

const RELAY_STATUS_OFFLINE = 0
const RELAY_STATUS_ONLINE = 1
const RELAY_STATUS_SHUTTING_DOWN = 2

var RelayStatusStrings = [3]string{"offline", "online", "shutting down"}

type ActiveRelay struct {
	Name     string
	Id       uint64
	Address  net.UDPAddr
	Status   int
	Sessions int
	Version  string
}

func (relayManager *RelayManager) GetActiveRelays() []ActiveRelay {

	relayManager.mutex.RLock()
	keys := make([]uint64, len(relayManager.sourceEntries))
	index := 0
	for k := range relayManager.sourceEntries {
		keys[index] = k
		index++
	}
	relayManager.mutex.RUnlock()

	activeRelays := make([]ActiveRelay, 0, len(keys))

	currentTime := time.Now()

	for i := range keys {

		relayManager.mutex.RLock()
		sourceEntry, ok := relayManager.sourceEntries[keys[i]]
		relayManager.mutex.RUnlock()

		if !ok {
			continue
		}

		sourceEntry.mutex.RLock()

		activeRelay := ActiveRelay{}

		activeRelay.Name = sourceEntry.relayName
		activeRelay.Address = sourceEntry.relayAddress
		activeRelay.Id = sourceEntry.relayId
		activeRelay.Sessions = sourceEntry.sessions

		activeRelay.Status = RELAY_STATUS_ONLINE
		if currentTime.Sub(sourceEntry.lastUpdateTime) > 10*time.Second {
			activeRelay.Status = RELAY_STATUS_ONLINE
		}
		if sourceEntry.shuttingDown {
			activeRelay.Status = RELAY_STATUS_SHUTTING_DOWN
		}

		activeRelay.Version = sourceEntry.relayVersion

		sourceEntry.mutex.RUnlock()

		activeRelays = append(activeRelays, activeRelay)
	}

	sort.SliceStable(activeRelays, func(i, j int) bool { return activeRelays[i].Name < activeRelays[j].Name })

	return activeRelays
}

func (relayManager *RelayManager) GetRelaysCSV() []byte {

	relaysCSV := "name,address,id,status,sessions,version\n"

	activeRelays := relayManager.GetActiveRelays()

	for i := range activeRelays {
		relay := activeRelays[i]
		relaysCSV += fmt.Sprintf("%s,%s,%016x,%s,%d,%s\n",
			relay.Name,
			relay.Address.String(),
			relay.Id,
			RelayStatusStrings[relay.Status],
			relay.Sessions,
			relay.Version)
	}

	return []byte(relaysCSV)
}
