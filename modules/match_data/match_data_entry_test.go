package match_data_test

// todo: tests disabled. there is a heisenbug in one of these tests
/*
import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	md "github.com/networknext/backend/modules/match_data"

	"github.com/stretchr/testify/assert"
)

func getTestMatchDataEntry() *md.MatchDataEntry {

	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	numMatchValues := rand.Intn(md.MatchDataMaxMatchValues)
	var matchValues [md.MatchDataMaxMatchValues]float64
	for i := 0; i < numMatchValues; i++ {
		matchValues[i] = rand.ExpFloat64()
	}

	return &md.MatchDataEntry{
		Version:        uint32(md.MatchDataEntryVersion),
		Timestamp:      uint32(time.Now().Unix()),
		BuyerID:        rand.Uint64(),
		ServerAddress:  backend.GenerateRandomStringSequence(md.MatchDataMaxAddressLength - 1),
		DatacenterID:   rand.Uint64(),
		UserHash:       rand.Uint64(),
		SessionID:      crypto.GenerateSessionID(),
		MatchID:        rand.Uint64(),
		NumMatchValues: int32(numMatchValues),
		MatchValues:    matchValues,
	}
}

// Helper function to check if write and read serialization provides the same entry
func writeReadEqualMatchDataEntry(entry *md.MatchDataEntry) ([]byte, error) {

	data, err := md.WriteMatchDataEntry(entry)
	if len(data) == 0 || err != nil {
		return data, err
	}

	entryRead := &md.MatchDataEntry{}

	err = md.ReadMatchDataEntry(entryRead, data)

	return data, err
}

// Helper function to check if write and read serialization work even if entry is clamped
func writeReadClampMatchDataEntry(entry *md.MatchDataEntry) ([]byte, *md.MatchDataEntry, error) {

	entry.ClampEntry()

	data, err := md.WriteMatchDataEntry(entry)
	if len(data) == 0 || err != nil {
		return data, &md.MatchDataEntry{}, err
	}

	readEntry := &md.MatchDataEntry{}
	err = md.ReadMatchDataEntry(readEntry, data)

	return data, readEntry, err
}

func TestSerializeMatchDataEntry_Empty(t *testing.T) {
	t.Parallel()

	const BufferSize = 256

	buffer := [BufferSize]byte{}

	writeStream, err := encoding.CreateWriteStream(buffer[:])
	assert.NoError(t, err)

	writeObject := &md.MatchDataEntry{Version: md.MatchDataEntryVersion}
	err = writeObject.Serialize(writeStream)
	assert.NoError(t, err)

	writeStream.Flush()

	readStream := encoding.CreateReadStream(buffer[:])
	readObject := &md.MatchDataEntry{Version: md.MatchDataEntryVersion}
	err = readObject.Serialize(readStream)
	assert.NoError(t, err)

	assert.Equal(t, writeObject, readObject)
}

func TestWriteMatchDataEntry_Empty(t *testing.T) {

	t.Parallel()

	entry := &md.MatchDataEntry{Version: md.MatchDataEntryVersion}
	data, err := md.WriteMatchDataEntry(entry)

	assert.NotEmpty(t, data)
	assert.NoError(t, err)
}

func TestReadMatchDataEntry_Empty(t *testing.T) {

	t.Parallel()

	entry := &md.MatchDataEntry{Version: md.MatchDataEntryVersion}
	data, err := md.WriteMatchDataEntry(entry)

	assert.NotEmpty(t, data)
	assert.NoError(t, err)

	entryRead := &md.MatchDataEntry{Version: md.MatchDataEntryVersion}

	err = md.ReadMatchDataEntry(entryRead, data)
	assert.NoError(t, err)
	assert.Equal(t, entry, entryRead)
}

func TestSerializeMatchDataEntry_FullMatchValues(t *testing.T) {
	t.Parallel()

	entry := getTestMatchDataEntry()
	entry.NumMatchValues = md.MatchDataMaxMatchValues
	for i := 0; i < int(entry.NumMatchValues); i++ {
		entry.MatchValues[i] = rand.ExpFloat64()
	}

	_, err := writeReadEqualMatchDataEntry(entry)
	assert.NoError(t, err)
}

func TestValidateMatchDataEntry(t *testing.T) {
	t.Parallel()

	t.Run("buyer id", func(t *testing.T) {
		entry := getTestMatchDataEntry()
		entry.BuyerID = 0

		valid := entry.Validate()
		assert.False(t, valid)
	})

	t.Run("server ip address", func(t *testing.T) {
		entry := getTestMatchDataEntry()
		entry.ServerAddress = ""

		valid := entry.Validate()
		assert.False(t, valid)
	})

	t.Run("session id", func(t *testing.T) {
		entry := getTestMatchDataEntry()
		entry.SessionID = 0

		valid := entry.Validate()
		assert.False(t, valid)
	})

	t.Run("num match values", func(t *testing.T) {
		entry := getTestMatchDataEntry()
		entry.NumMatchValues = -1

		valid := entry.Validate()
		assert.False(t, valid)
	})

	t.Run("num match values", func(t *testing.T) {
		entry := getTestMatchDataEntry()
		entry.NumMatchValues = md.MatchDataMaxMatchValues + 1

		valid := entry.Validate()
		assert.False(t, valid)
	})

	t.Run("success", func(t *testing.T) {
		entry := getTestMatchDataEntry()

		valid := entry.Validate()
		assert.True(t, valid)
	})
}

func TestCheckNaNOrInf(t *testing.T) {
	t.Parallel()

	t.Run("nan values", func(t *testing.T) {
		entry := getTestMatchDataEntry()
		if entry.NumMatchValues == 0 {
			entry.NumMatchValues = 1
		}

		for i := 0; i < int(entry.NumMatchValues); i++ {
			entry.MatchValues[i] = math.NaN()
		}

		nanOrInfExists, nanOrInfFields := entry.CheckNaNOrInf()
		assert.True(t, nanOrInfExists)
		assert.Equal(t, []string{"MatchValues"}, nanOrInfFields)
	})

	t.Run("pos inf values", func(t *testing.T) {
		entry := getTestMatchDataEntry()
		if entry.NumMatchValues == 0 {
			entry.NumMatchValues = 1
		}

		for i := 0; i < int(entry.NumMatchValues); i++ {
			entry.MatchValues[i] = math.Inf(1)
		}

		nanOrInfExists, nanOrInfFields := entry.CheckNaNOrInf()
		assert.True(t, nanOrInfExists)
		assert.Equal(t, []string{"MatchValues"}, nanOrInfFields)
	})

	t.Run("neg inf values", func(t *testing.T) {
		entry := getTestMatchDataEntry()
		if entry.NumMatchValues == 0 {
			entry.NumMatchValues = 1
		}

		for i := 0; i < int(entry.NumMatchValues); i++ {
			entry.MatchValues[i] = math.Inf(-1)
		}

		nanOrInfExists, nanOrInfFields := entry.CheckNaNOrInf()
		assert.True(t, nanOrInfExists)
		assert.Equal(t, []string{"MatchValues"}, nanOrInfFields)
	})

	t.Run("no nan or inf values", func(t *testing.T) {
		entry := getTestMatchDataEntry()
		if entry.NumMatchValues == 0 {
			entry.NumMatchValues = 1
		}

		for i := 0; i < int(entry.NumMatchValues); i++ {
			entry.MatchValues[i] = rand.ExpFloat64()
		}

		nanOrInfExists, nanOrInfFields := entry.CheckNaNOrInf()
		assert.False(t, nanOrInfExists)
		assert.Zero(t, len(nanOrInfFields))
	})
}

func TestSerializeMatchDataEntry_Clamp(t *testing.T) {
	t.Parallel()

	t.Run("server IP address length", func(t *testing.T) {
		entry := getTestMatchDataEntry()
		serverAddrStr := backend.GenerateRandomStringSequence(md.MatchDataMaxAddressLength + 1)
		assert.Equal(t, md.MatchDataMaxAddressLength+1, len(serverAddrStr))
		entry.ServerAddress = serverAddrStr

		data, readEntry, err := writeReadClampMatchDataEntry(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
		assert.Equal(t, serverAddrStr[:md.MatchDataMaxAddressLength-1], readEntry.ServerAddress)
	})

	t.Run("num match values", func(t *testing.T) {
		entry := getTestMatchDataEntry()
		entry.NumMatchValues = -1
		assert.Equal(t, int32(-1), entry.NumMatchValues)

		data, readEntry, err := writeReadClampMatchDataEntry(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
		assert.NotEqual(t, entry, readEntry)
		assert.Equal(t, int32(0), readEntry.NumMatchValues)

		entry = getTestMatchDataEntry()
		entry.NumMatchValues = md.MatchDataMaxMatchValues + 1
		assert.Equal(t, int32(md.MatchDataMaxMatchValues+1), entry.NumMatchValues)

		data, readEntry, err = writeReadClampMatchDataEntry(entry)
		assert.NotEmpty(t, data)
		assert.NoError(t, err)
		assert.Equal(t, int32(md.MatchDataMaxMatchValues), readEntry.NumMatchValues)
	})
}
*/
