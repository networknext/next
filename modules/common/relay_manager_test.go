package common_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/common"

	"github.com/stretchr/testify/assert"
)

func TestRelayManager_Local(t *testing.T) {

	t.Parallel()

	relayManager := common.CreateRelayManager(true)

	// 10 database relays. A B C D E F G H I J

	databaseRelayNames := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}

	databaseNumRelays := len(databaseRelayNames)

	databaseRelayIds := make([]uint64, databaseNumRelays)

	databaseRelayAddresses := make([]net.UDPAddr, databaseNumRelays)

	for i := range databaseRelayIds {
		databaseRelayIds[i] = common.RelayId(databaseRelayNames[i])
		databaseRelayAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
	}

	// 5 active relays. A B C D E

	relayNames := []string{"A", "B", "C", "D", "E"}

	numRelays := len(relayNames)

	relayIds := make([]uint64, numRelays)

	relayAddresses := make([]net.UDPAddr, numRelays)

	for i := range relayIds {
		relayIds[i] = common.RelayId(relayNames[i])
		relayAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
	}

	// initially, the relay manager should have an empty cost matrix and no relays

	const MaxJitter = 100
	const MaxPacketLoss = 1

	currentTime := time.Now().Unix()

	costs := relayManager.GetCosts(currentTime, relayIds, MaxJitter, MaxPacketLoss)

	assert.Equal(t, len(costs), int(common.TriMatrixLength(numRelays)))

	for i := range costs {
		assert.Equal(t, costs[i], uint8(0xFF))
	}

	activeRelays := relayManager.GetActiveRelays(currentTime)

	assert.Equal(t, len(activeRelays), 0)

	relays := relayManager.GetRelays(currentTime, databaseRelayIds, databaseRelayNames, databaseRelayAddresses)

	assert.Equal(t, len(relays), databaseNumRelays)

	// now apply some relay updates, only for relays "A" and "B", and verify we see only see the two relays as active

	counters := [constants.NumRelayCounters]uint64{}

	relayManager.ProcessRelayUpdate(currentTime, relayIds[0], relayNames[0], relayAddresses[0], 0, "test", 0, 0, nil, nil, nil, nil, counters[:])

	relayManager.ProcessRelayUpdate(currentTime, relayIds[1], relayNames[1], relayAddresses[1], 0, "test", 0, 0, nil, nil, nil, nil, counters[:])

	// we should see both relay A and B in the active relays

	activeRelays = relayManager.GetActiveRelays(currentTime)

	assert.Equal(t, len(activeRelays), 2)

	assert.Equal(t, activeRelays[0].Id, relayIds[0])
	assert.Equal(t, activeRelays[1].Id, relayIds[1])

	relays = relayManager.GetRelays(currentTime, databaseRelayIds, databaseRelayNames, databaseRelayAddresses)

	assert.Equal(t, len(relays), databaseNumRelays)

	numActive := 0
	for i := 0; i < databaseNumRelays; i++ {
		if relays[i].Id == activeRelays[0].Id || relays[i].Id == activeRelays[1].Id {
			assert.Equal(t, relays[i].Status, constants.RelayStatus_Online)
			numActive++
		} else {
			assert.Equal(t, relays[i].Status, constants.RelayStatus_Offline)
		}
	}

	assert.Equal(t, numActive, 2)

	// since we provided no samples in the relay updates, the route matrix should still be empty.

	costs = relayManager.GetCosts(currentTime, relayIds, MaxJitter, MaxPacketLoss)

	assert.Equal(t, len(costs), int(common.TriMatrixLength(numRelays)))

	for i := range costs {
		assert.Equal(t, uint8(0xFF), costs[i])
	}

	// now get active relays, but with a timestamp in the future enough that they should be timed out

	activeRelays = relayManager.GetActiveRelays(currentTime + 30)

	assert.Equal(t, 0, len(activeRelays))

	// now get relays in the future. all relays should be in the offline status

	relays = relayManager.GetRelays(currentTime+30, databaseRelayIds, databaseRelayNames, databaseRelayAddresses)

	assert.Equal(t, len(relays), databaseNumRelays)

	for i := 0; i < databaseNumRelays; i++ {
		assert.Equal(t, relays[i].Status, constants.RelayStatus_Offline)
	}
}

func TestRelayManager_Real(t *testing.T) {

	t.Parallel()

	relayManager := common.CreateRelayManager(false)

	// 5 active relays. A B C D E

	relayNames := []string{"A", "B", "C", "D", "E"}

	numRelays := len(relayNames)

	relayIds := make([]uint64, numRelays)

	relayAddresses := make([]net.UDPAddr, numRelays)

	for i := range relayIds {
		relayIds[i] = common.RelayId(relayNames[i])
		relayAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
	}

	// iterate adding samples from A <-> B. initially, they should remain unroutable until they have both been
	// alive for at least HistorySize relay updates. this avoids sending traffic to relays when they first start
	// and we don't necessarily know their routes are stable yet.

	for i := 0; i < constants.RelayHistorySize*2; i++ {

		// add some samples from relay A -> B
		{
			sampleRelayId := [1]uint64{relayIds[1]}
			sampleRTT := [1]float32{10.0}
			sampleJitter := [1]float32{0.0}
			samplePacketLoss := [1]float32{0.0}
			relayManager.ProcessRelayUpdate(currentTime, relayIds[0], relayNames[0], relayAddresses[0], 0, "test", 0, 1, sampleRelayId[:], sampleRTT[:], sampleJitter[:], samplePacketLoss[:], counters[:])
		}

		// add some samples from relay B -> A
		{
			sampleRelayId := [1]uint64{relayIds[0]}
			sampleRTT := [1]float32{10.0}
			sampleJitter := [1]float32{0.0}
			samplePacketLoss := [1]float32{0.0}
			relayManager.ProcessRelayUpdate(currentTime, relayIds[1], relayNames[1], relayAddresses[1], 0, "test", false, 1, sampleRelayId[:], sampleRTT[:], sampleJitter[:], samplePacketLoss[:])
		}

		if i < constants.RelayHistorySize {

			// we should see no routes between A and B until HistorySize relay updates

			for i := range costs {
				assert.Equal(t, int32(-1), costs[i])
			}

		} else {

			// we should see entries between A and B as routable in the cost matrix, but all other entries should be -1

			costs = relayManager.GetCosts(currentTime, relayIds, MaxJitter, MaxPacketLoss)

			assert.Equal(t, len(costs), int(common.TriMatrixLength(numRelays)))

			for i := 0; i < 5; i++ {
				for j := 0; j < 5; j++ {
					index := common.TriMatrixIndex(i, j)
					if i == j {
						continue
					}
					if i < 2 && j < 2 {
						// expect valid entry
						assert.Equal(t, int32(10), costs[index])
					} else {
						// expect -1 entry
						assert.Equal(t, int32(-1), costs[index])
					}
				}
			}
		}
	}

/*
	// getting costs 30 seconds in the future should result in routes between A and B being timed out

	costs = relayManager.GetCosts(currentTime+30, relayIds, MaxRTT, MaxJitter, MaxPacketLoss, false)

	assert.Equal(t, len(costs), int(common.TriMatrixLength(numRelays)))

	for i := range costs {
		assert.Equal(t, int32(-1), costs[i])
	}

	// getting active relays 30 seconds in the future should result in no active relays (A and B timed out)

	activeRelays = relayManager.GetActiveRelays(currentTime + 30)

	assert.Equal(t, 0, len(activeRelays))

	// relays should be in the offline state 30 seconds in the future

	relays = relayManager.GetRelays(currentTime+30, databaseRelayIds, databaseRelayNames, databaseRelayAddresses)

	assert.Equal(t, len(relays), databaseNumRelays)

	for i := 0; i < databaseNumRelays; i++ {
		assert.Equal(t, relays[i].Status, common.RELAY_STATUS_OFFLINE)
	}

	// apply a relay update that says relay A is shutting down. routes between relay A and B should instantly go away.

	relayManager.ProcessRelayUpdate(currentTime, relayIds[0], relayNames[0], relayAddresses[0], 0, "test", true, 0, nil, nil, nil, nil)

	costs = relayManager.GetCosts(currentTime, relayIds, MaxRTT, MaxJitter, MaxPacketLoss, false)

	assert.Equal(t, len(costs), int(common.TriMatrixLength(numRelays)))

	for i := range costs {
		assert.Equal(t, int32(-1), costs[i])
	}

	// active relays should not include relays that are shutting down

	activeRelays = relayManager.GetActiveRelays(currentTime)

	assert.Equal(t, 1, len(activeRelays))

	assert.Equal(t, activeRelays[0].Id, relayIds[1]) // only relay "B" is still online

	// we should see the shutting down relays in the relays array

	relays = relayManager.GetRelays(currentTime, databaseRelayIds, databaseRelayNames, databaseRelayAddresses)

	assert.Equal(t, len(relays), databaseNumRelays)

	for i := 0; i < databaseNumRelays; i++ {
		if relays[i].Id == relayIds[0] {
			assert.Equal(t, relays[i].Status, common.RELAY_STATUS_SHUTTING_DOWN)
		} else if relays[i].Id == relayIds[1] {
			assert.Equal(t, relays[i].Status, common.RELAY_STATUS_ONLINE)
		} else {
			assert.Equal(t, relays[i].Status, common.RELAY_STATUS_OFFLINE)
		}
	}

	// 30 seconds in the future, shutting down should become offline

	activeRelays = relayManager.GetActiveRelays(currentTime + 30)

	assert.Equal(t, 0, len(activeRelays))

	relays = relayManager.GetRelays(currentTime+30, databaseRelayIds, databaseRelayNames, databaseRelayAddresses)

	assert.Equal(t, len(relays), databaseNumRelays)

	for i := 0; i < databaseNumRelays; i++ {
		assert.Equal(t, relays[i].Status, common.RELAY_STATUS_OFFLINE)
	}

	// now restart relay A. we should need to wait another HistorySize number of updates before we see routes between A and B

	for i := 0; i < common.HistorySize*2; i++ {

		// add some samples from relay A -> B
		{
			sampleRelayId := [1]uint64{relayIds[1]}
			sampleRTT := [1]float32{10.0}
			sampleJitter := [1]float32{0.0}
			samplePacketLoss := [1]float32{0.0}
			relayManager.ProcessRelayUpdate(currentTime, relayIds[0], relayNames[0], relayAddresses[0], 0, "test", false, 1, sampleRelayId[:], sampleRTT[:], sampleJitter[:], samplePacketLoss[:])
		}

		// add some samples from relay B -> A
		{
			sampleRelayId := [1]uint64{relayIds[0]}
			sampleRTT := [1]float32{10.0}
			sampleJitter := [1]float32{0.0}
			samplePacketLoss := [1]float32{0.0}
			relayManager.ProcessRelayUpdate(currentTime, relayIds[1], relayNames[1], relayAddresses[1], 0, "test", false, 1, sampleRelayId[:], sampleRTT[:], sampleJitter[:], samplePacketLoss[:])
		}

		if i < common.HistorySize {

			// we should see no routes between A and B until HistorySize relay updates

			for i := range costs {
				assert.Equal(t, int32(-1), costs[i])
			}

		} else {

			// we should see entries between A and B as routable in the cost matrix, but all other entries should be -1

			costs = relayManager.GetCosts(currentTime, relayIds, MaxRTT, MaxJitter, MaxPacketLoss, false)

			assert.Equal(t, len(costs), int(common.TriMatrixLength(numRelays)))

			for i := 0; i < 5; i++ {
				for j := 0; j < 5; j++ {
					index := common.TriMatrixIndex(i, j)
					if i == j {
						continue
					}
					if i < 2 && j < 2 {
						// expect valid entry
						assert.Equal(t, int32(10), costs[index])
					} else {
						// expect -1 entry
						assert.Equal(t, int32(-1), costs[index])
					}
				}
			}
		}

		// relays "A" and "B" should be active throughout

		activeRelays = relayManager.GetActiveRelays(currentTime)

		assert.Equal(t, len(activeRelays), 2)

		assert.Equal(t, activeRelays[0].Id, relayIds[0])
		assert.Equal(t, activeRelays[1].Id, relayIds[1])

		relays = relayManager.GetRelays(currentTime, databaseRelayIds, databaseRelayNames, databaseRelayAddresses)

		assert.Equal(t, len(relays), databaseNumRelays)

		numActive := 0
		for i := 0; i < databaseNumRelays; i++ {
			if relays[i].Id == activeRelays[0].Id || relays[i].Id == activeRelays[1].Id {
				assert.Equal(t, relays[i].Status, common.RELAY_STATUS_ONLINE)
				numActive++
			} else {
				assert.Equal(t, relays[i].Status, common.RELAY_STATUS_OFFLINE)
			}
		}

		assert.Equal(t, numActive, 2)
	}
*/
}
