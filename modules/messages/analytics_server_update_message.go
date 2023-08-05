package messages

import (
	"fmt"
	"net"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/next/modules/encoding"
)

const (
	AnalyticsServerUpdateMessageVersion_Min   = 1
	AnalyticsServerUpdateMessageVersion_Max   = 1
	AnalyticsServerUpdateMessageVersion_Write = 1

	MaxAnalyticsServerUpdateMessageSize = 128
)

type AnalyticsServerUpdateMessage struct {
	Version          byte
	Timestamp        uint64
	SDKVersion_Major byte
	SDKVersion_Minor byte
	SDKVersion_Patch byte
	BuyerId          uint64
	DatacenterId     uint64
	MatchId          uint64
	NumSessions      uint32
	ServerAddress    net.UDPAddr
}

func (message *AnalyticsServerUpdateMessage) GetMaxSize() int {
	return 64
}

func (message *AnalyticsServerUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read analytics server update message version")
	}

	if message.Version < AnalyticsServerUpdateMessageVersion_Min || message.Version > AnalyticsServerUpdateMessageVersion_Max {
		return fmt.Errorf("invalid server update message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Major) {
		return fmt.Errorf("failed to read sdk version major")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Minor) {
		return fmt.Errorf("failed to read sdk version major")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Patch) {
		return fmt.Errorf("failed to read sdk version major")
	}

	if !encoding.ReadUint64(buffer, &index, &message.BuyerId) {
		return fmt.Errorf("failed to read buyer id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.DatacenterId) {
		return fmt.Errorf("failed to read datacenter id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.MatchId) {
		return fmt.Errorf("failed to read match id")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumSessions) {
		return fmt.Errorf("failed to read num sessions")
	}

	if !encoding.ReadAddress(buffer, &index, &message.ServerAddress) {
		return fmt.Errorf("failed to read server address")
	}

	return nil
}

func (message *AnalyticsServerUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < AnalyticsServerUpdateMessageVersion_Min || message.Version > AnalyticsServerUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid analytics server update message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Major)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Minor)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Patch)
	encoding.WriteUint64(buffer, &index, message.BuyerId)
	encoding.WriteUint64(buffer, &index, message.DatacenterId)
	encoding.WriteUint64(buffer, &index, message.MatchId)
	encoding.WriteUint32(buffer, &index, message.NumSessions)
	encoding.WriteAddress(buffer, &index, &message.ServerAddress)

	return buffer[:index]
}

func (message *AnalyticsServerUpdateMessage) Save() (map[string]bigquery.Value, string, error) {
	bigquery_entry := make(map[string]bigquery.Value)
	bigquery_entry["timestamp"] = int(message.Timestamp)
	bigquery_entry["sdk_version_major"] = int(message.SDKVersion_Major)
	bigquery_entry["sdk_version_minor"] = int(message.SDKVersion_Minor)
	bigquery_entry["sdk_version_patch"] = int(message.SDKVersion_Patch)
	bigquery_entry["buyer_id"] = int(message.BuyerId)
	bigquery_entry["datacenter_id"] = int(message.DatacenterId)
	if message.MatchId != 0 {
		bigquery_entry["match_id"] = int(message.MatchId)
	}
	bigquery_entry["num_sessions"] = int(message.NumSessions)
	bigquery_entry["server_address"] = message.ServerAddress.String()
	return bigquery_entry, "", nil
}
