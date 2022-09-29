package messages

import (
	"net"
	"fmt"

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
	Timestamp      uint64
	BuyerId        uint64
	ServerAddress  net.UDPAddr
	DatacenterId   uint64
	UserHash       uint64
	SessionId      uint64
	MatchId        uint64
	NumMatchValues uint32
	MatchValues    [MatchDataMaxMatchValues]float64
}

func (message *MatchDataMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint32(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read match data version")
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadUint64(buffer, &index, &message.BuyerId) {
		return fmt.Errorf("failed to read buyer id")
	}

	if !encoding.ReadAddress(buffer, &index, &message.ServerAddress) {
		return fmt.Errorf("failed to read server address")
	}

	if !encoding.ReadUint64(buffer, &index, &message.DatacenterId) {
		return fmt.Errorf("failed to read datacenter id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.UserHash) {
		return fmt.Errorf("failed to read user hash")
	}

	if !encoding.ReadUint64(buffer, &index, &message.SessionId) {
		return fmt.Errorf("failed to read session id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.MatchId) {
		return fmt.Errorf("failed to read match id")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumMatchValues) {
		return fmt.Errorf("failed to read num match values")
	}

	for i := 0; i < int(message.NumMatchValues); i++ {
		if !encoding.ReadFloat64(buffer, &index, &message.MatchValues[i]) {
			return fmt.Errorf("failed to read match value %d", i)
		}
	}

	return nil
}

func (message *MatchDataMessage) Write(buffer []byte) []byte {
	
	index := 0

	// todo: implement write

	return buffer[:index]
}

func (message *MatchDataMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["buyerID"] = int(message.BuyerId)
	bigquery_message["serverAddress"] = message.ServerAddress
	bigquery_message["datacenterID"] = int(message.DatacenterId)
	bigquery_message["userHash"] = int(message.UserHash)
	bigquery_message["sessionID"] = int(message.SessionId)
	bigquery_message["matchID"] = int(message.MatchId)

	if message.NumMatchValues > 0 {
		matchValues := make([]bigquery.Value, message.NumMatchValues)
		for i := 0; i < int(message.NumMatchValues); i++ {
			matchValues[i] = float64(message.MatchValues[i])
		}
		bigquery_message["matchValues"] = matchValues
	}

	return bigquery_message, "", nil
}
