package messages

import (
	"fmt"
	"net"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/accelerate/modules/constants"
	"github.com/networknext/accelerate/modules/encoding"
)

const (
	AnalyticsMatchDataMessageVersion_Min   = 1
	AnalyticsMatchDataMessageVersion_Max   = 1
	AnalyticsMatchDataMessageVersion_Write = 1

	MaxAnalyticsMatchDataMessageBytes = 2048
)

type AnalyticsMatchDataMessage struct {
	Version        uint8
	Timestamp      uint64
	Type           uint64
	BuyerId        uint64
	ServerAddress  net.UDPAddr
	DatacenterId   uint64
	SessionId      uint64
	MatchId        uint64
	NumMatchValues uint32
	MatchValues    [constants.MaxMatchValues]float64
}

func (message *AnalyticsMatchDataMessage) GetMaxSize() int {
	return 64 + 8*constants.MaxMatchValues
}

func (message *AnalyticsMatchDataMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read analytics match data version")
	}

	if message.Version < AnalyticsMatchDataMessageVersion_Min || message.Version > AnalyticsMatchDataMessageVersion_Max {
		return fmt.Errorf("invalid analytics match data version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadUint64(buffer, &index, &message.Type) {
		return fmt.Errorf("failed to read type")
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

func (message *AnalyticsMatchDataMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < AnalyticsMatchDataMessageVersion_Min || message.Version > AnalyticsMatchDataMessageVersion_Max {
		panic(fmt.Sprintf("invalid analytics match data version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint64(buffer, &index, message.Type)
	encoding.WriteUint64(buffer, &index, message.BuyerId)
	encoding.WriteAddress(buffer, &index, &message.ServerAddress)
	encoding.WriteUint64(buffer, &index, message.DatacenterId)
	encoding.WriteUint64(buffer, &index, message.SessionId)
	encoding.WriteUint64(buffer, &index, message.MatchId)
	encoding.WriteUint32(buffer, &index, message.NumMatchValues)

	for i := 0; i < int(message.NumMatchValues); i++ {
		encoding.WriteFloat64(buffer, &index, message.MatchValues[i])
	}

	return buffer[:index]
}

func (message *AnalyticsMatchDataMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["type"] = int(message.Type)
	bigquery_message["buyer_id"] = int(message.BuyerId)
	bigquery_message["server_address"] = message.ServerAddress.String()
	bigquery_message["datacenter_id"] = int(message.DatacenterId)
	bigquery_message["session_id"] = int(message.SessionId)
	bigquery_message["match_id"] = int(message.MatchId)

	matchValues := make([]bigquery.Value, message.NumMatchValues)
	for i := 0; i < int(message.NumMatchValues); i++ {
		matchValues[i] = float64(message.MatchValues[i])
	}
	bigquery_message["match_values"] = matchValues

	return bigquery_message, "", nil
}
