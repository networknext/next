package messages

import (
	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/modules/encoding"
)

const (
	MatchDataMessageVersion = uint32(0)

	MaxMatchDataMessageBytes = 2048

	MatchDataMaxAddressLength = 256
	MatchDataMaxMatchValues   = 64
)

type MatchDataMessage struct {
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

func (message *MatchDataMessage) Serialize(stream encoding.Stream) error {
	stream.SerializeBits(&message.Version, 32)
	stream.SerializeBits(&message.Timestamp, 32)

	stream.SerializeUint64(&message.BuyerID)
	stream.SerializeString(&message.ServerAddress, MatchDataMaxAddressLength)
	stream.SerializeUint64(&message.DatacenterID)
	stream.SerializeUint64(&message.UserHash)
	stream.SerializeUint64(&message.SessionID)
	stream.SerializeUint64(&message.MatchID)
	stream.SerializeInteger(&message.NumMatchValues, 0, MatchDataMaxMatchValues)
	for i := 0; i < int(message.NumMatchValues); i++ {
		stream.SerializeFloat64(&message.MatchValues[i])
	}

	return stream.Error()
}

/*
func WriteMatchDataMessage(message *MatchDataMessage) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			core.Error("recovered from panic during MatchDataMessage packet entry write: %v\n", r)
		}
	}()

	buffer := [MaxMatchDataMessageBytes]byte{}

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := message.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadMatchDataMessage(message *MatchDataMessage, data []byte) error {
	if err := message.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}

	return nil
}
*/

/*
// Validate a match data message. Returns true if the match data entry is valid, false if invalid.
func (message *MatchDataMessage) Validate() bool {
	if message.Version < 0 {
		core.Error("invalid version")
		return false
	}

	if message.Timestamp < 0 {
		core.Error("invalid timestamp")
		return false
	}

	if message.BuyerID == 0 {
		core.Error("invalid buyer id")
		return false
	}

	if message.ServerAddress == "" {
		core.Error("invalid server address")
		return false
	}

	// NOTE: we don't validate the DatacenterID, UserHash, and MatchID since that can come in as 0 from the SDK

	if message.SessionID == 0 {
		core.Error("invalid session id")
		return false
	}

	if message.NumMatchValues < 0 || message.NumMatchValues > MatchDataMaxMatchValues {
		core.Error("invalid num match values (%d)", message.NumMatchValues)
		return false
	}

	return true
}

// Checks all floating point numbers in the MatchDataMessage for NaN and +-Inf and forces them to 0
func (message *MatchDataMessage) CheckNaNOrInf() (bool, []string) {
	var nanOrInfExists bool
	var nanOrInfFields []string

	for i := 0; i < int(message.NumMatchValues); i++ {
		if math.IsNaN(message.MatchValues[i]) || math.IsInf(message.MatchValues[i], 0) {
			nanOrInfExists = true
			nanOrInfFields = []string{"MatchValues"}
			message.MatchValues[i] = float64(0)
		}
	}

	return nanOrInfExists, nanOrInfFields
}

// To save bits during serialization, clamp integer and string fields if they go beyond the min
// or max values as defined in MatchDataMessage.Serialize()
func (message *MatchDataMessage) ClampMessage() {

	if len(message.ServerAddress) >= MatchDataMaxAddressLength {
		core.Debug("MatchDataMessage Server IP Address length (%d) >= MatchDataMaxAddressLength (%d). Clamping to MatchDataMaxAddressLength - 1 (%d)", len(message.ServerAddress), MatchDataMaxAddressLength, MatchDataMaxAddressLength-1)
		message.ServerAddress = message.ServerAddress[:MatchDataMaxAddressLength-1]
	}

	if message.NumMatchValues < 0 {
		core.Error("MatchDataMessage NumMatchValues (%d) < 0. Clamping to 0.", message.NumMatchValues)
		message.NumMatchValues = 0
	}

	if message.NumMatchValues > MatchDataMaxMatchValues {
		core.Debug("MatchDataMessage NumMatchValues (%d) > MatchDataMaxMatchValues (%d). Clamping to MatchDataMaxMatchValues.", message.NumMatchValues, MatchDataMaxMatchValues)
		message.NumMatchValues = MatchDataMaxMatchValues
	}
}
*/

func (message *MatchDataMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["buyerID"] = int(message.BuyerID)
	bigquery_message["serverAddress"] = message.ServerAddress
	bigquery_message["datacenterID"] = int(message.DatacenterID)
	bigquery_message["userHash"] = int(message.UserHash)
	bigquery_message["sessionID"] = int(message.SessionID)
	bigquery_message["matchID"] = int(message.MatchID)

	if message.NumMatchValues > 0 {
		matchValues := make([]bigquery.Value, message.NumMatchValues)
		for i := 0; i < int(message.NumMatchValues); i++ {
			matchValues[i] = float64(message.MatchValues[i])
		}
		bigquery_message["matchValues"] = matchValues
	}

	return bigquery_message, "", nil
}
