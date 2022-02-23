package match_data

import (
	"math"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
)

const (
	MatchDataEntryVersion = uint32(0)

	MaxMatchDataEntryBytes = 2048

	MatchDataMaxAddressLength = 256
	MatchDataMaxMatchValues   = 64
)

type MatchDataEntry struct {
	Version        uint32
	Timestamp      uint32
	BuyerID        uint64
	ServerAddress  string
	DatacenterID   uint64
	UserHash       uint64
	SessionID      uint64
	MatchID        uint64
	NumMatchValues int32
	MatchValues    [MatchDataMaxMatchValues]float64
}

func (entry *MatchDataEntry) Serialize(stream encoding.Stream) error {
	stream.SerializeBits(&entry.Version, 32)
	stream.SerializeBits(&entry.Timestamp, 32)

	stream.SerializeUint64(&entry.BuyerID)
	stream.SerializeString(&entry.ServerAddress, MatchDataMaxAddressLength)
	stream.SerializeUint64(&entry.DatacenterID)
	stream.SerializeUint64(&entry.UserHash)
	stream.SerializeUint64(&entry.SessionID)
	stream.SerializeUint64(&entry.MatchID)
	stream.SerializeInteger(&entry.NumMatchValues, 0, MatchDataMaxMatchValues)
	for i := 0; i < int(entry.NumMatchValues); i++ {
		stream.SerializeFloat64(&entry.MatchValues[i])
	}

	return stream.Error()
}

func WriteMatchDataEntry(entry *MatchDataEntry) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			core.Error("recovered from panic during MatchDataEntry packet entry write: %v\n", r)
		}
	}()

	buffer := [MaxMatchDataEntryBytes]byte{}

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := entry.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadMatchDataEntry(entry *MatchDataEntry, data []byte) error {
	if err := entry.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}

	return nil
}

// Validate a match data entry. Returns true if the match data entry is valid, false if invalid.
func (entry *MatchDataEntry) Validate() bool {
	if entry.Version < 0 {
		core.Error("invalid version")
		return false
	}

	if entry.Timestamp < 0 {
		core.Error("invalid timestamp")
		return false
	}

	if entry.BuyerID == 0 {
		core.Error("invalid buyer id")
		return false
	}

	if entry.ServerAddress == "" {
		core.Error("invalid server address")
		return false
	}

	// NOTE: we don't validate the DatacenterID and UserHash since that can come in as 0 from the SDK

	if entry.SessionID == 0 {
		core.Error("invalid session id")
		return false
	}

	if entry.MatchID == 0 {
		core.Error("invalid match id")
		return false
	}

	if entry.NumMatchValues < 0 || entry.NumMatchValues > MatchDataMaxMatchValues {
		core.Error("invalid num match values (%d)", entry.NumMatchValues)
		return false
	}

	return true
}

// Checks all floating point numbers in the MatchDataEntry for NaN and +-Inf and forces them to 0
func (entry *MatchDataEntry) CheckNaNOrInf() (bool, []string) {
	var nanOrInfExists bool
	var nanOrInfFields []string

	for i := 0; i < int(entry.NumMatchValues); i++ {
		if math.IsNaN(entry.MatchValues[i]) || math.IsInf(entry.MatchValues[i], 0) {
			nanOrInfExists = true
			nanOrInfFields = []string{"MatchValues"}
			entry.MatchValues[i] = float64(0)
		}
	}

	return nanOrInfExists, nanOrInfFields
}

// To save bits during serialization, clamp integer and string fields if they go beyond the min
// or max values as defined in MatchDataEntry.Serialize()
func (entry *MatchDataEntry) ClampEntry() {

	if len(entry.ServerAddress) >= MatchDataMaxAddressLength {
		core.Debug("MatchDataEntry Server IP Address length (%d) >= MatchDataMaxAddressLength (%d). Clamping to MatchDataMaxAddressLength - 1 (%d)", len(entry.ServerAddress), MatchDataMaxAddressLength, MatchDataMaxAddressLength-1)
		entry.ServerAddress = entry.ServerAddress[:MatchDataMaxAddressLength-1]
	}

	if entry.NumMatchValues < 0 {
		core.Error("MatchDataEntry NumMatchValues (%d) < 0. Clamping to 0.", entry.NumMatchValues)
		entry.NumMatchValues = 0
	}

	if entry.NumMatchValues > MatchDataMaxMatchValues {
		core.Debug("MatchDataEntry NumMatchValues (%d) > MatchDataMaxMatchValues (%d). Clamping to MatchDataMaxMatchValues.", entry.NumMatchValues, MatchDataMaxMatchValues)
		entry.NumMatchValues = MatchDataMaxMatchValues
	}
}

func (entry *MatchDataEntry) Save() (map[string]bigquery.Value, string, error) {

	e := make(map[string]bigquery.Value)

	e["timestamp"] = int(entry.Timestamp)
	e["buyerID"] = int(entry.BuyerID)
	e["serverAddress"] = entry.ServerAddress
	e["datacenterID"] = int(entry.DatacenterID)
	e["userHash"] = int(entry.UserHash)
	e["sessionID"] = int(entry.SessionID)
	e["matchID"] = int(entry.MatchID)

	if entry.NumMatchValues > 0 {
		matchValues := make([]bigquery.Value, entry.NumMatchValues)
		for i := 0; i < int(entry.NumMatchValues); i++ {
			matchValues[i] = int(entry.MatchValues[i])
		}
		e["matchValues"] = matchValues
	}

	return e, "", nil
}
