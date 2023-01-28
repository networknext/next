package common

import (
	"fmt"
	"math"
	"net"
	"sort"
	"sync"
)

const RelayTimeout = 10

const HistorySize = 300 // 5 minutes @ one relay update per-second

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
	lastUpdateTime    int64
	rtt               float32
	jitter            float32
	packetLoss        float32
	historyIndex      int32
	historyRTT        [HistorySize]float32
	historyJitter     [HistorySize]float32
	historyPacketLoss [HistorySize]float32
}

type RelayManagerSourceEntry struct {
	lastUpdateTime int64
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

func (relayManager *RelayManager) ProcessRelayUpdate(currentTime int64, relayId uint64, relayName string, relayAddress net.UDPAddr, sessions int, relayVersion string, shuttingDown bool, numSamples int, sampleRelayId []uint64, sampleRTT []float32, sampleJitter []float32, samplePacketLoss []float32) {

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
	*/

	// look up the entry corresponding to the source relay, or create it if it doesn't exist

	relayManager.mutex.Lock()

	// todo
	/*
		iowa:   357632dc51bc4049
		oregon: b79d96976666501a
	*/

	debug := relayId == 0x357632dc51bc4049 || relayId == 0xb79d96976666501a

	if debug {
		fmt.Printf("=============================================================\n")
	}

	sourceEntry, exists := relayManager.sourceEntries[relayId]
	if !exists {
		if debug {
			fmt.Printf("source entry %x does not exist. adding new one\n", relayId)
		}
		sourceEntry = &RelayManagerSourceEntry{}
		sourceEntry.destEntries = make(map[uint64]*RelayManagerDestEntry)
		relayManager.sourceEntries[relayId] = sourceEntry
	}

	// update stats for the source relay, then...
	// iterate across all samples and insert them into the history buffer
	// in the dest entry corresponding to their relay pair (source,dest)

	sourceEntry.lastUpdateTime = currentTime
	sourceEntry.relayId = relayId
	sourceEntry.relayName = relayName
	sourceEntry.relayAddress = relayAddress
	sourceEntry.sessions = sessions
	sourceEntry.relayVersion = relayVersion
	sourceEntry.shuttingDown = shuttingDown

	for i := 0; i < numSamples; i++ {

		destRelayId := sampleRelayId[i]

		// todo
		if destRelayId != 0x357632dc51bc4049 && destRelayId != 0xb79d96976666501a {
			continue
		}

		destEntry, exists := sourceEntry.destEntries[destRelayId]
		if !exists {
			if debug {
				fmt.Printf("dest entry %x does not exist. adding new one\n", destRelayId)
			}
			destEntry = &RelayManagerDestEntry{}
			sourceEntry.destEntries[destRelayId] = destEntry
		}

		// IMPORTANT: clear newly created AND timed out dest entry history buffers
		// this is important so that newly created relays, and relays that are stopped and restarted
		// don't get routed across, until at least 5 minutes has passed!

		if currentTime-destEntry.lastUpdateTime > RelayTimeout {
			if debug {
				fmt.Printf("clearing history for (%x,%x)\n", relayId, destRelayId)
			}
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

		if debug {
			fmt.Printf("clearing history for (%x,%x)\n", relayId, destRelayId)
		}

		destEntry.historyIndex = (destEntry.historyIndex + 1) % HistorySize

		destEntry.lastUpdateTime = currentTime

		if debug {
			fmt.Printf("update entry (%x,%x) -> rtt = %.1f, jitter = %.1f, pl = %.2f\n", relayId, destRelayId, destEntry.rtt, destEntry.jitter, destEntry.packetLoss)
		}
	}

	if debug {
		fmt.Printf("=============================================================\n")
	}

	relayManager.mutex.Unlock()
}

func (relayManager *RelayManager) getSample(currentTime int64, sourceRelayId uint64, destRelayId uint64) (float32, float32, float32) {

	sourceRTT := float32(InvalidRouteValue)
	sourceJitter := float32(InvalidRouteValue)
	sourcePacketLoss := float32(InvalidRouteValue)

	destRTT := float32(InvalidRouteValue)
	destJitter := float32(InvalidRouteValue)
	destPacketLoss := float32(InvalidRouteValue)

	// todo
	/*
		iowa:   357632dc51bc4049
		oregon: b79d96976666501a
	*/

	debug := (sourceRelayId == 0x357632dc51bc4049 && destRelayId == 0xb79d96976666501a) || (sourceRelayId == 0xb79d96976666501a) || (destRelayId == 0x357632dc51bc4049)

	if debug {
		fmt.Printf("=============================================================\n")
	}

	// get source ping values
	{
		sourceEntry, exists := relayManager.sourceEntries[sourceRelayId]

		if debug && !exists {
			fmt.Printf("(1) source entry for relay %x does not exist?!\n", sourceRelayId)
		}

		if exists {

			if debug {
				fmt.Printf("(1) source relay last update time = %d\n", sourceEntry.lastUpdateTime)
				if sourceEntry.shuttingDown {
					fmt.Printf("(1) source relay is shutting down?!\n")
				}
			}

			if currentTime-sourceEntry.lastUpdateTime < RelayTimeout && !sourceEntry.shuttingDown {
				destEntry, exists := sourceEntry.destEntries[destRelayId]
				if exists {
					if currentTime-destEntry.lastUpdateTime < RelayTimeout {
						sourceRTT = destEntry.rtt
						sourceJitter = destEntry.jitter
						sourcePacketLoss = destEntry.packetLoss
					} else {
						if debug {
							fmt.Printf("(1) dest entry has timed out?!\n")
						}
					}
				}
			} else {
				if debug {
					fmt.Printf("(1) source entry has timed out?!\n")
				}
			}
		}
	}

	// get dest ping values
	{
		sourceEntry, exists := relayManager.sourceEntries[destRelayId]

		if debug && !exists {
			fmt.Printf("(2) source entry for relay %x does not exist?!\n", destRelayId)
		}

		if exists {

			if debug {
				fmt.Printf("(2) source relay last update time = %d\n", sourceEntry.lastUpdateTime)
				if sourceEntry.shuttingDown {
					fmt.Printf("(2) source relay is shutting down?!\n")
				}
			}

			if currentTime-sourceEntry.lastUpdateTime < RelayTimeout && !sourceEntry.shuttingDown {
				destEntry, exists := sourceEntry.destEntries[sourceRelayId]

				if exists {
					if currentTime-destEntry.lastUpdateTime < RelayTimeout {
						destRTT = destEntry.rtt
						destJitter = destEntry.jitter
						destPacketLoss = destEntry.packetLoss
					} else {
						if debug {
							fmt.Printf("(2) dest entry has timed out?!\n")
						}
					}
				}
			} else {
				if debug {
					fmt.Printf("(2) source entry has timed out?!\n")
				}
			}
		}
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

	if debug {
		fmt.Printf("[debug] rtt = %.1f, jitter = %.1f, pl = %.1f\n", rtt, jitter, packetLoss)
		fmt.Printf("=============================================================\n")
	}

	return rtt, jitter, packetLoss
}

func (relayManager *RelayManager) GetCosts(currentTime int64, relayIds []uint64, maxRTT float32, maxJitter float32, maxPacketLoss float32, local bool) []int32 {

	numRelays := len(relayIds)

	costs := make([]int32, TriMatrixLength(numRelays))

	// special permissive cost matrix for local

	// todo: really we should unify here. i think lack of unification caused problems in prod that didn't repro in local

	if local {
		for i := range costs {
			costs[i] = -1
		}
		activeRelayHash := relayManager.GetActiveRelayHash(currentTime)
		for i := 0; i < numRelays; i++ {
			sourceRelayId := uint64(relayIds[i])
			_, sourceActive := activeRelayHash[sourceRelayId]
			if !sourceActive {
				continue
			}
			for j := 0; j < i; j++ {
				index := TriMatrixIndex(i, j)
				destRelayId := uint64(relayIds[j])
				_, destActive := activeRelayHash[destRelayId]
				if destActive {
					costs[index] = 0
				}
			}
		}
		return costs
	}	

	// production code

	relayManager.mutex.RLock()

	for i := range costs {
		costs[i] = -1
	}
	activeRelayHash := relayManager.GetActiveRelayHash(currentTime)
	for i := 0; i < numRelays; i++ {
		sourceRelayId := uint64(relayIds[i])
		_, sourceActive := activeRelayHash[sourceRelayId]
		if !sourceActive {
			continue
		}
		for j := 0; j < i; j++ {
			index := TriMatrixIndex(i, j)
			destRelayId := uint64(relayIds[j])
			_, destActive := activeRelayHash[destRelayId]
			if destActive {
				rtt, jitter, packetLoss := relayManager.getSample(currentTime, sourceRelayId, destRelayId)
				if rtt < maxRTT && jitter < maxJitter && packetLoss < maxPacketLoss {
					costs[index] = int32(math.Ceil(float64(rtt)))
				} else {
					costs[index] = -1
				}
			}
		}
	}

	/*
	for i := 0; i < numRelays; i++ {
		sourceRelayId := uint64(relayIds[i])
		for j := 0; j < i; j++ {
			index := TriMatrixIndex(i, j)
			destRelayId := uint64(relayIds[j])
			rtt, jitter, packetLoss := relayManager.getSample(currentTime, sourceRelayId, destRelayId)
			if rtt < maxRTT && jitter < maxJitter && packetLoss < maxPacketLoss {
				costs[index] = int32(math.Ceil(float64(rtt)))
			} else {
				costs[index] = -1
			}
		}
	}
	*/

	relayManager.mutex.RUnlock()

	return costs
}

const RELAY_STATUS_OFFLINE = 0
const RELAY_STATUS_ONLINE = 1
const RELAY_STATUS_SHUTTING_DOWN = 2

var RelayStatusStrings = [3]string{"offline", "online", "shutting down"}

type Relay struct {
	Id       uint64
	Name     string
	Address  net.UDPAddr
	Status   int
	Sessions int
	Version  string
}

func (relayManager *RelayManager) GetRelays(currentTime int64, relayIds []uint64, relayNames []string, relayAddresses []net.UDPAddr) []Relay {

	relayManager.mutex.RLock()

	keys := make([]uint64, len(relayManager.sourceEntries))
	index := 0
	for k := range relayManager.sourceEntries {
		keys[index] = k
		index++
	}

	relays := make([]Relay, 0, len(keys))

	for i := range keys {

		sourceEntry, ok := relayManager.sourceEntries[keys[i]]

		if !ok {
			continue
		}

		relay := Relay{}

		relay.Id = sourceEntry.relayId
		relay.Name = sourceEntry.relayName
		relay.Address = sourceEntry.relayAddress
		relay.Sessions = sourceEntry.sessions

		relay.Status = RELAY_STATUS_ONLINE

		if sourceEntry.shuttingDown {
			relay.Status = RELAY_STATUS_SHUTTING_DOWN
		}

		expired := currentTime-sourceEntry.lastUpdateTime > RelayTimeout

		if expired {
			relay.Status = RELAY_STATUS_OFFLINE
		}

		if relay.Status == RELAY_STATUS_ONLINE {
			relay.Version = sourceEntry.relayVersion
		}

		if relay.Status != RELAY_STATUS_ONLINE {
			relay.Sessions = 0
		}

		relays = append(relays, relay)
	}

	// pick up any relays that the relay manager doesn't know about as offline

	for i := 0; i < len(relayIds); i++ {

		_, exists := relayManager.sourceEntries[relayIds[i]]
		if exists {
			continue
		}

		relay := Relay{}

		relay.Id = relayIds[i]
		relay.Name = relayNames[i]
		relay.Address = relayAddresses[i]
		relay.Sessions = 0
		relay.Version = ""
		relay.Status = RELAY_STATUS_OFFLINE

		relays = append(relays, relay)
	}

	relayManager.mutex.RUnlock()

	// sort to make sure the set of relays is stable order over time

	sort.SliceStable(relays, func(i, j int) bool { return relays[i].Name < relays[j].Name })

	return relays
}

func (relayManager *RelayManager) GetActiveRelays(currentTime int64) []Relay {

	relayManager.mutex.RLock()

	keys := make([]uint64, len(relayManager.sourceEntries))
	index := 0
	for k := range relayManager.sourceEntries {
		keys[index] = k
		index++
	}

	activeRelays := make([]Relay, 0, len(keys))

	for i := range keys {

		sourceEntry, ok := relayManager.sourceEntries[keys[i]]

		if !ok {
			continue
		}

		activeRelay := Relay{}
		activeRelay.Status = RELAY_STATUS_ONLINE
		activeRelay.Name = sourceEntry.relayName
		activeRelay.Address = sourceEntry.relayAddress
		activeRelay.Id = sourceEntry.relayId
		activeRelay.Sessions = sourceEntry.sessions
		activeRelay.Version = sourceEntry.relayVersion

		expired := currentTime-sourceEntry.lastUpdateTime > RelayTimeout

		shuttingDown := sourceEntry.shuttingDown

		if expired || shuttingDown {
			continue
		}

		activeRelays = append(activeRelays, activeRelay)
	}

	relayManager.mutex.RUnlock()

	sort.SliceStable(activeRelays, func(i, j int) bool { return activeRelays[i].Name < activeRelays[j].Name })

	return activeRelays
}

func (relayManager *RelayManager) GetActiveRelayHash(currentTime int64) map[uint64]Relay {
	activeRelays := relayManager.GetActiveRelays(currentTime)
	activeRelayHash := make(map[uint64]Relay)
	for i := range activeRelays {
		activeRelayHash[activeRelays[i].Id] = activeRelays[i]
	}
	return activeRelayHash
}

func (relayManager *RelayManager) GetRelaysCSV(currentTime int64, relayIds []uint64, relayNames []string, relayAddresses []net.UDPAddr) []byte {

	relaysCSV := "name,address,id,status,sessions,version\n"

	relays := relayManager.GetRelays(currentTime, relayIds, relayNames, relayAddresses)

	for i := range relays {
		relay := relays[i]
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
