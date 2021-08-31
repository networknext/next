package billing_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"

	"github.com/stretchr/testify/assert"
)

// Helper function to create a random string of a specified length
// Useful for testing constant string lengths
// Adapted from: https://stackoverflow.com/a/22892986
func generateRandomStringSequence(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

// Returns a BillingEntry2 struct with all the data filled out and each condition flag disabled
func getTestBillingEntry2() *billing.BillingEntry2 {

	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	numTags := rand.Intn(billing.BillingEntryMaxTags)
	var tags [billing.BillingEntryMaxTags]uint64
	for i := 0; i < numTags; i++ {
		tags[i] = rand.Uint64()
	}

	numNearRelays := rand.Intn(billing.BillingEntryMaxNearRelays)
	var nearRelayIDs [billing.BillingEntryMaxNearRelays]uint64
	var nearRelayRTTs [billing.BillingEntryMaxNearRelays]int32
	var nearRelayJitters [billing.BillingEntryMaxNearRelays]int32
	var nearRelayPacketLosses [billing.BillingEntryMaxNearRelays]int32
	for i := 0; i < numNearRelays; i++ {
		nearRelayIDs[i] = rand.Uint64()
		nearRelayRTTs[i] = int32(rand.Intn(255))
		nearRelayJitters[i] = int32(rand.Intn(255))
		nearRelayPacketLosses[i] = int32(rand.Intn(100))
	}

	numNextRelays := rand.Intn(billing.BillingEntryMaxRelays)
	var nextRelays [billing.BillingEntryMaxRelays]uint64
	var nextRelayPrice [billing.BillingEntryMaxRelays]uint64
	for i := 0; i < numNextRelays; i++ {
		nextRelays[i] = rand.Uint64()
		nextRelayPrice[i] = rand.Uint64()
	}

	return &billing.BillingEntry2{
		Version:                         uint32(billing.BillingEntryVersion2),
		Timestamp:                       uint32(time.Now().Unix()),
		SessionID:                       crypto.GenerateSessionID(),
		SliceNumber:                     5,
		DirectRTT:                       int32(rand.Intn(1024)),
		DirectJitter:                    int32(rand.Intn(255)),
		DirectPacketLoss:                int32(rand.Intn(100)),
		RealPacketLoss:                  int32(rand.Intn(100)),
		RealPacketLoss_Frac:             uint32(rand.Intn(255)),
		RealJitter:                      uint32(rand.Intn(255)),
		Next:                            false,
		Flagged:                         false,
		Summary:                         false,
		UseDebug:                        false,
		Debug:                           generateRandomStringSequence(billing.BillingEntryMaxDebugLength - 1),
		RouteDiversity:                  int32(rand.Intn(32)),
		DatacenterID:                    rand.Uint64(),
		BuyerID:                         rand.Uint64(),
		UserHash:                        rand.Uint64(),
		EnvelopeBytesDown:               rand.Uint64(),
		EnvelopeBytesUp:                 rand.Uint64(),
		Latitude:                        rand.Float32(),
		Longitude:                       rand.Float32(),
		ISP:                             generateRandomStringSequence(billing.BillingEntryMaxISPLength - 1),
		ConnectionType:                  int32(rand.Intn(3)),
		PlatformType:                    int32(rand.Intn(10)),
		SDKVersion:                      generateRandomStringSequence(billing.BillingEntryMaxSDKVersionLength - 1),
		NumTags:                         int32(numTags),
		Tags:                            tags,
		ABTest:                          false,
		Pro:                             false,
		ClientToServerPacketsSent:       rand.Uint64(),
		ServerToClientPacketsSent:       rand.Uint64(),
		ClientToServerPacketsLost:       rand.Uint64(),
		ServerToClientPacketsLost:       rand.Uint64(),
		ClientToServerPacketsOutOfOrder: rand.Uint64(),
		ServerToClientPacketsOutOfOrder: rand.Uint64(),
		NumNearRelays:                   int32(numNearRelays),
		NearRelayIDs:                    nearRelayIDs,
		NearRelayRTTs:                   nearRelayRTTs,
		NearRelayJitters:                nearRelayJitters,
		NearRelayPacketLosses:           nearRelayPacketLosses,
		TotalPriceSum:                   rand.Uint64(),
		EnvelopeBytesUpSum:              rand.Uint64(),
		EnvelopeBytesDownSum:            rand.Uint64(),
		SessionDuration:                 5 * billing.BillingSliceSeconds,
		EverOnNext:                      false,
		DurationOnNext:                  4 * billing.BillingSliceSeconds,
		NextRTT:                         int32(rand.Intn(255)),
		NextJitter:                      int32(rand.Intn(255)),
		NextPacketLoss:                  int32(rand.Intn(100)),
		PredictedNextRTT:                int32(rand.Intn(255)),
		NearRelayRTT:                    int32(rand.Intn(255)),
		NumNextRelays:                   int32(numNextRelays),
		NextRelays:                      nextRelays,
		NextRelayPrice:                  nextRelayPrice,
		TotalPrice:                      rand.Uint64(),
		Uncommitted:                     true,
		Multipath:                       false,
		RTTReduction:                    false,
		PacketLossReduction:             false,
		RouteChanged:                    false,
		FallbackToDirect:                false,
		MultipathVetoed:                 false,
		Mispredicted:                    false,
		Vetoed:                          false,
		LatencyWorse:                    false,
		NoRoute:                         false,
		NextLatencyTooHigh:              false,
		CommitVeto:                      false,
		UnknownDatacenter:               false,
		DatacenterNotEnabled:            false,
		BuyerNotLive:                    false,
		StaleRouteMatrix:                false,
		NextBytesUp:                     rand.Uint64(),
		NextBytesDown:                   rand.Uint64(),
	}
}

// Helper function to check if write and read serialization provides the same entry
func writeReadEqualBillingEntry2(entry *billing.BillingEntry2) ([]byte, error) {

	data, err := billing.WriteBillingEntry2(entry)
	if len(data) == 0 || err != nil {
		return data, err
	}

	entryRead := &billing.BillingEntry2{}

	err = billing.ReadBillingEntry2(entryRead, data)

	return data, err
}

// Helper function to check if write and read serialization work even if entry is clamped
func writeReadClampBillingEntry2(entry *billing.BillingEntry2) ([]byte, *billing.BillingEntry2, error) {

	entry.ClampEntry()

	data, err := billing.WriteBillingEntry2(entry)
	if len(data) == 0 || err != nil {
		return data, &billing.BillingEntry2{}, err
	}

	readEntry := &billing.BillingEntry2{}
	err = billing.ReadBillingEntry2(readEntry, data)

	return data, readEntry, err
}

func TestSerializeBillingEntry2_Empty(t *testing.T) {

	t.Parallel()

	const BufferSize = 256

	buffer := [BufferSize]byte{}

	writeStream, err := encoding.CreateWriteStream(buffer[:])
	assert.NoError(t, err)

	writeObject := &billing.BillingEntry2{Version: billing.BillingEntryVersion2}
	err = writeObject.Serialize(writeStream)
	assert.NoError(t, err)

	writeStream.Flush()

	readStream := encoding.CreateReadStream(buffer[:])
	readObject := &billing.BillingEntry2{Version: billing.BillingEntryVersion2}
	err = readObject.Serialize(readStream)
	assert.NoError(t, err)

	assert.Equal(t, writeObject, readObject)
}

func TestWriteBillingEntry2_Empty(t *testing.T) {

	t.Parallel()

	entry := &billing.BillingEntry2{Version: billing.BillingEntryVersion2}
	data, err := billing.WriteBillingEntry2(entry)

	assert.NotEmpty(t, data)
	assert.NoError(t, err)
}

func TestReadBillingEntry2_Empty(t *testing.T) {

	t.Parallel()

	entry := &billing.BillingEntry2{Version: billing.BillingEntryVersion2}
	data, err := billing.WriteBillingEntry2(entry)

	assert.NotEmpty(t, data)
	assert.NoError(t, err)

	entryRead := &billing.BillingEntry2{Version: billing.BillingEntryVersion2}

	err = billing.ReadBillingEntry2(entryRead, data)
	assert.NoError(t, err)
	assert.Equal(t, entry, entryRead)
}

func TestSerializeBillingEntry2_EveryCondition(t *testing.T) {

	t.Parallel()

	// We would never actually run into the case where
	// the slice is 0, on next, summary is true, and has an error state,
	// but it should still serialize properly

	entry := getTestBillingEntry2()
	entry.SliceNumber = 0
	entry.Next = true
	entry.EverOnNext = true
	entry.Summary = true
	entry.Uncommitted = false
	entry.FallbackToDirect = true

	data, err := writeReadEqualBillingEntry2(entry)
	assert.NotEmpty(t, data)
	assert.NoError(t, err)
}

func TestSerializeBillingEntry2_InitialSlice(t *testing.T) {

	t.Parallel()

	entry := getTestBillingEntry2()
	entry.SliceNumber = 0

	data, err := writeReadEqualBillingEntry2(entry)
	assert.NotEmpty(t, data)
	assert.NoError(t, err)
}

func TestSerializeBillingEntry2_DecidingToTakeNext(t *testing.T) {

	t.Parallel()

	entry := getTestBillingEntry2()
	entry.SliceNumber = 1
	entry.RouteDiversity = 8
	entry.UseDebug = true
	entry.Debug = "decided to take network next"
	entry.Uncommitted = false

	data, err := writeReadEqualBillingEntry2(entry)
	assert.NotEmpty(t, data)
	assert.NoError(t, err)
}

func TestSerializeBillingEntry2_OnNext(t *testing.T) {

	t.Parallel()

	entry := getTestBillingEntry2()
	entry.Next = true
	entry.EverOnNext = true
	entry.UseDebug = true
	entry.Debug = "on network next"
	entry.Uncommitted = false

	data, err := writeReadEqualBillingEntry2(entry)
	assert.NotEmpty(t, data)
	assert.NoError(t, err)
}

func TestSerializeBillingEntry2_OnNext_RouteChanged(t *testing.T) {

	t.Parallel()

	entry := getTestBillingEntry2()
	entry.Next = true
	entry.EverOnNext = true
	entry.UseDebug = true
	entry.Debug = "on network next. route changed."
	entry.Uncommitted = false
	entry.RouteChanged = true

	data, err := writeReadEqualBillingEntry2(entry)
	assert.NotEmpty(t, data)
	assert.NoError(t, err)
}

func TestSerializeBillingEntry2_ErrorState_Next(t *testing.T) {

	t.Parallel()

	t.Run("error state - fallback to direct", func(t *testing.T) {
		entry := getTestBillingEntry2()
		entry.Next = true
		entry.EverOnNext = true
		entry.UseDebug = true
		entry.Debug = "leaving network next. fallback to direct."
		entry.Uncommitted = true
		entry.FallbackToDirect = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - multipath veto", func(t *testing.T) {
		entry := getTestBillingEntry2()
		entry.Next = true
		entry.EverOnNext = true
		entry.UseDebug = true
		entry.Debug = "leaving network next. multipath veto."
		entry.Uncommitted = true
		entry.MultipathVetoed = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - Mispredicted", func(t *testing.T) {
		entry := getTestBillingEntry2()
		entry.Next = true
		entry.EverOnNext = true
		entry.UseDebug = true
		entry.Debug = "leaving network next. Mispredicted."
		entry.Uncommitted = true
		entry.Mispredicted = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - vetoed", func(t *testing.T) {
		entry := getTestBillingEntry2()
		entry.Next = true
		entry.EverOnNext = true
		entry.UseDebug = true
		entry.Debug = "leaving network next. vetoed."
		entry.Uncommitted = true
		entry.Vetoed = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - latency worse", func(t *testing.T) {
		entry := getTestBillingEntry2()
		entry.Next = true
		entry.EverOnNext = true
		entry.UseDebug = true
		entry.Debug = "leaving network next. latency worse."
		entry.Uncommitted = true
		entry.LatencyWorse = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - no route", func(t *testing.T) {
		entry := getTestBillingEntry2()
		entry.Next = true
		entry.EverOnNext = true
		entry.UseDebug = true
		entry.Debug = "leaving network next. no route."
		entry.Uncommitted = true
		entry.NoRoute = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - next latency too high", func(t *testing.T) {
		entry := getTestBillingEntry2()
		entry.Next = true
		entry.EverOnNext = true
		entry.UseDebug = true
		entry.Debug = "leaving network next. next latency too high."
		entry.Uncommitted = true
		entry.NextLatencyTooHigh = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - commit veto", func(t *testing.T) {
		entry := getTestBillingEntry2()
		entry.Next = true
		entry.EverOnNext = true
		entry.UseDebug = true
		entry.Debug = "leaving network next. commit veto."
		entry.Uncommitted = true
		entry.CommitVeto = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - unknown datacenter", func(t *testing.T) {
		// We wouldn't be on next for this case, but still need to test

		entry := getTestBillingEntry2()
		entry.Next = true
		entry.EverOnNext = true
		entry.UseDebug = true
		entry.Debug = "leaving network next. unknown datacenter."
		entry.Uncommitted = true
		entry.UnknownDatacenter = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - datacenter not enabled", func(t *testing.T) {
		// We wouldn't be on next for this case, but still need to test

		entry := getTestBillingEntry2()
		entry.Next = true
		entry.EverOnNext = true
		entry.UseDebug = true
		entry.Debug = "leaving network next. datacenter not enabled."
		entry.Uncommitted = true
		entry.DatacenterNotEnabled = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - buyer not live", func(t *testing.T) {
		// We wouldn't be on next for this case, but still need to test

		entry := getTestBillingEntry2()
		entry.Next = true
		entry.EverOnNext = true
		entry.UseDebug = true
		entry.Debug = "leaving network next. buyer not live."
		entry.Uncommitted = true
		entry.BuyerNotLive = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - stale route matrix", func(t *testing.T) {
		entry := getTestBillingEntry2()
		entry.Next = true
		entry.EverOnNext = true
		entry.UseDebug = true
		entry.Debug = "leaving network next. stale route matrix."
		entry.Uncommitted = true
		entry.StaleRouteMatrix = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})
}

func TestSerializeBillingEntry2_ErrorState_Direct(t *testing.T) {

	t.Parallel()

	t.Run("error state - fallback to direct", func(t *testing.T) {
		entry := getTestBillingEntry2()

		entry.UseDebug = true
		entry.Debug = "on direct. fallback to direct."
		entry.Uncommitted = true
		entry.FallbackToDirect = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - multipath veto", func(t *testing.T) {
		entry := getTestBillingEntry2()

		entry.UseDebug = true
		entry.Debug = "on direct. multipath veto."
		entry.Uncommitted = true
		entry.MultipathVetoed = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - Mispredicted", func(t *testing.T) {
		entry := getTestBillingEntry2()

		entry.UseDebug = true
		entry.Debug = "on direct. Mispredicted."
		entry.Uncommitted = true
		entry.Mispredicted = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - vetoed", func(t *testing.T) {
		entry := getTestBillingEntry2()

		entry.UseDebug = true
		entry.Debug = "on direct. vetoed."
		entry.Uncommitted = true
		entry.Vetoed = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - latency worse", func(t *testing.T) {
		entry := getTestBillingEntry2()

		entry.UseDebug = true
		entry.Debug = "on direct. latency worse."
		entry.Uncommitted = true
		entry.LatencyWorse = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - no route", func(t *testing.T) {
		entry := getTestBillingEntry2()

		entry.UseDebug = true
		entry.Debug = "on direct. no route."
		entry.Uncommitted = true
		entry.NoRoute = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - next latency too high", func(t *testing.T) {
		entry := getTestBillingEntry2()

		entry.UseDebug = true
		entry.Debug = "on direct. next latency too high."
		entry.Uncommitted = true
		entry.NextLatencyTooHigh = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - commit veto", func(t *testing.T) {
		entry := getTestBillingEntry2()

		entry.UseDebug = true
		entry.Debug = "on direct. commit veto."
		entry.Uncommitted = true
		entry.CommitVeto = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - unknown datacenter", func(t *testing.T) {
		entry := getTestBillingEntry2()

		entry.UseDebug = true
		entry.Debug = "on direct. unknown datacenter."
		entry.Uncommitted = true
		entry.UnknownDatacenter = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - datacenter not enabled", func(t *testing.T) {
		entry := getTestBillingEntry2()

		entry.UseDebug = true
		entry.Debug = "on direct. datacenter not enabled."
		entry.Uncommitted = true
		entry.DatacenterNotEnabled = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - buyer not live", func(t *testing.T) {
		entry := getTestBillingEntry2()

		entry.UseDebug = true
		entry.Debug = "on direct. buyer not live."
		entry.Uncommitted = true
		entry.BuyerNotLive = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})

	t.Run("error state - stale route matrix", func(t *testing.T) {
		entry := getTestBillingEntry2()

		entry.UseDebug = true
		entry.Debug = "on direct. stale route matrix."
		entry.Uncommitted = true
		entry.StaleRouteMatrix = true

		data, err := writeReadEqualBillingEntry2(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
	})
}

func TestSerializeBillingEntry2_Summary_Direct(t *testing.T) {

	t.Parallel()

	entry := getTestBillingEntry2()
	entry.Summary = true

	data, err := writeReadEqualBillingEntry2(entry)
	assert.NotEmpty(t, data)
	assert.NoError(t, err)
}

func TestSerializeBillingEntry2_Summary_Next(t *testing.T) {

	t.Parallel()

	entry := getTestBillingEntry2()
	entry.Summary = true
	entry.Next = true
	entry.EverOnNext = true
	entry.Uncommitted = false

	data, err := writeReadEqualBillingEntry2(entry)
	assert.NotEmpty(t, data)
	assert.NoError(t, err)
}

func TestSerializeBillingEntry2_Clamp(t *testing.T) {

	t.Parallel()

	var data []byte
	var entry *billing.BillingEntry2
	var readEntry *billing.BillingEntry2
	var err error

	t.Run("test always clamp", func(t *testing.T) {

		t.Run("direct RTT", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.DirectRTT = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.DirectRTT)

			entry = getTestBillingEntry2()
			entry.DirectRTT = 1024
			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(1023), readEntry.DirectRTT)
		})

		t.Run("direct jitter", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.DirectJitter = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.DirectJitter)

			entry = getTestBillingEntry2()
			entry.DirectJitter = 256
			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(255), readEntry.DirectJitter)
		})

		t.Run("direct packet loss", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.DirectPacketLoss = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.DirectPacketLoss)

			entry = getTestBillingEntry2()
			entry.DirectPacketLoss = 101
			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(100), readEntry.DirectPacketLoss)
		})

		t.Run("real packet loss", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.RealPacketLoss = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.RealPacketLoss)

			entry = getTestBillingEntry2()
			entry.RealPacketLoss = 101
			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(100), readEntry.RealPacketLoss)
		})

		t.Run("real jitter", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.RealJitter = 1001
			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, uint32(1000), readEntry.RealJitter)
		})

		t.Run("debug length", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.UseDebug = true
			debugStr := generateRandomStringSequence(billing.BillingEntryMaxDebugLength + 1)
			assert.Equal(t, billing.BillingEntryMaxDebugLength+1, len(debugStr))
			entry.Debug = debugStr

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, debugStr[:billing.BillingEntryMaxDebugLength-1], readEntry.Debug)
		})

		t.Run("route diversity", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.RouteDiversity = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.RouteDiversity)

			entry = getTestBillingEntry2()
			entry.RouteDiversity = 33
			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(32), readEntry.RouteDiversity)
		})

	})

	t.Run("test first slice", func(t *testing.T) {

		t.Run("isp length", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.SliceNumber = 0
			ispStr := generateRandomStringSequence(billing.BillingEntryMaxISPLength + 1)
			assert.Equal(t, billing.BillingEntryMaxISPLength+1, len(ispStr))
			entry.ISP = ispStr

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, ispStr[:billing.BillingEntryMaxISPLength-1], readEntry.ISP)
		})

		t.Run("connection type", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.SliceNumber = 0
			entry.ConnectionType = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.ConnectionType)

			entry = getTestBillingEntry2()
			entry.SliceNumber = 0
			entry.ConnectionType = 4

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.ConnectionType)
		})

		t.Run("platform type", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.SliceNumber = 0
			entry.PlatformType = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.PlatformType)

			entry = getTestBillingEntry2()
			entry.SliceNumber = 0
			entry.PlatformType = 11

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.PlatformType)
		})

		t.Run("sdk version", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.SliceNumber = 0
			sdkStr := generateRandomStringSequence(billing.BillingEntryMaxSDKVersionLength + 1)
			assert.Equal(t, billing.BillingEntryMaxSDKVersionLength+1, len(sdkStr))
			entry.SDKVersion = sdkStr

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, sdkStr[:billing.BillingEntryMaxSDKVersionLength-1], readEntry.SDKVersion)
		})

		t.Run("num tags", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.SliceNumber = 0
			entry.NumTags = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.NumTags)

			entry = getTestBillingEntry2()
			entry.SliceNumber = 0
			entry.NumTags = billing.BillingEntryMaxTags + 1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(billing.BillingEntryMaxTags), readEntry.NumTags)
		})
	})

	t.Run("test summary slice", func(t *testing.T) {

		t.Run("num near relays", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.Summary = true
			entry.NumNearRelays = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.NumNearRelays)

			entry = getTestBillingEntry2()
			entry.Summary = true
			entry.NumNearRelays = billing.BillingEntryMaxNearRelays + 1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(billing.BillingEntryMaxNearRelays), readEntry.NumNearRelays)
		})

		t.Run("near relay RTTs", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.Summary = true
			entry.NumNearRelays = 1
			entry.NearRelayRTTs = [billing.BillingEntryMaxNearRelays]int32{}
			entry.NearRelayRTTs[0] = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.NearRelayRTTs[0])

			entry = getTestBillingEntry2()
			entry.Summary = true
			entry.NumNearRelays = 1
			entry.NearRelayRTTs = [billing.BillingEntryMaxNearRelays]int32{}
			entry.NearRelayRTTs[0] = 256

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(255), readEntry.NearRelayRTTs[0])
		})

		t.Run("near relay jitters", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.Summary = true
			entry.NumNearRelays = 1
			entry.NearRelayJitters = [billing.BillingEntryMaxNearRelays]int32{}
			entry.NearRelayJitters[0] = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.NearRelayJitters[0])

			entry = getTestBillingEntry2()
			entry.Summary = true
			entry.NumNearRelays = 1
			entry.NearRelayJitters = [billing.BillingEntryMaxNearRelays]int32{}
			entry.NearRelayJitters[0] = 256

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(255), readEntry.NearRelayJitters[0])
		})

		t.Run("near relay packet losses", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.Summary = true
			entry.NumNearRelays = 1
			entry.NearRelayPacketLosses = [billing.BillingEntryMaxNearRelays]int32{}
			entry.NearRelayPacketLosses[0] = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.NearRelayPacketLosses[0])

			entry = getTestBillingEntry2()
			entry.Summary = true
			entry.NumNearRelays = 1
			entry.NearRelayPacketLosses = [billing.BillingEntryMaxNearRelays]int32{}
			entry.NearRelayPacketLosses[0] = 101

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(100), readEntry.NearRelayPacketLosses[0])
		})
	})

	t.Run("test on network next", func(t *testing.T) {

		t.Run("next rtt", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.Next = true
			entry.EverOnNext = true
			entry.NextRTT = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.NextRTT)

			entry = getTestBillingEntry2()
			entry.Next = true
			entry.EverOnNext = true
			entry.NextRTT = 256

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(255), readEntry.NextRTT)
		})

		t.Run("next jitter", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.Next = true
			entry.EverOnNext = true
			entry.NextJitter = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.NextJitter)

			entry = getTestBillingEntry2()
			entry.Next = true
			entry.EverOnNext = true
			entry.NextJitter = 256

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(255), readEntry.NextJitter)
		})

		t.Run("next packet loss", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.Next = true
			entry.EverOnNext = true
			entry.NextPacketLoss = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.NextPacketLoss)

			entry = getTestBillingEntry2()
			entry.Next = true
			entry.EverOnNext = true
			entry.NextPacketLoss = 101

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(100), readEntry.NextPacketLoss)
		})

		t.Run("predicted next rtt", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.Next = true
			entry.EverOnNext = true
			entry.PredictedNextRTT = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.PredictedNextRTT)

			entry = getTestBillingEntry2()
			entry.Next = true
			entry.EverOnNext = true
			entry.PredictedNextRTT = 256

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(255), readEntry.PredictedNextRTT)
		})

		t.Run("near relay rtt", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.Next = true
			entry.EverOnNext = true
			entry.NearRelayRTT = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.NearRelayRTT)

			entry = getTestBillingEntry2()
			entry.Next = true
			entry.EverOnNext = true
			entry.NearRelayRTT = 256

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(255), readEntry.NearRelayRTT)
		})

		t.Run("num next relays", func(t *testing.T) {
			entry = getTestBillingEntry2()
			entry.Next = true
			entry.EverOnNext = true
			entry.NumNextRelays = -1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(0), readEntry.NumNextRelays)

			entry = getTestBillingEntry2()
			entry.Next = true
			entry.EverOnNext = true
			entry.NumNextRelays = billing.BillingEntryMaxRelays + 1

			data, readEntry, err = writeReadClampBillingEntry2(entry)
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
			assert.NotEqual(t, entry, readEntry)
			assert.Equal(t, int32(billing.BillingEntryMaxRelays), readEntry.NumNextRelays)
		})
	})
}
