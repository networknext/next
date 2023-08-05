package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/next/modules/encoding"
)

const (
	AnalyticsDatabaseUpdateMessageVersion_Min   = 3
	AnalyticsDatabaseUpdateMessageVersion_Max   = 3
	AnalyticsDatabaseUpdateMessageVersion_Write = 3
)

type AnalyticsDatabaseUpdateMessage struct {
	Version        uint8
	Timestamp      uint64
	DatabaseSize   uint32
	NumRelays      uint32
	NumDatacenters uint32
	NumSellers     uint32
	NumBuyers      uint32
}

func (message *AnalyticsDatabaseUpdateMessage) GetMaxSize() int {
	return 32
}

func (message *AnalyticsDatabaseUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read analytics database update version")
	}

	if message.Version < AnalyticsDatabaseUpdateMessageVersion_Min || message.Version > AnalyticsDatabaseUpdateMessageVersion_Max {
		return fmt.Errorf("invalid analytics database update message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadUint32(buffer, &index, &message.DatabaseSize) {
		return fmt.Errorf("failed to read database size")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumRelays) {
		return fmt.Errorf("failed to read num relays")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumDatacenters) {
		return fmt.Errorf("failed to read num datacenters")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumSellers) {
		return fmt.Errorf("failed to read num sellers")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumBuyers) {
		return fmt.Errorf("failed to read num buyers")
	}

	return nil
}

func (message *AnalyticsDatabaseUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < AnalyticsDatabaseUpdateMessageVersion_Min || message.Version > AnalyticsDatabaseUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid analytics database update message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint32(buffer, &index, message.DatabaseSize)
	encoding.WriteUint32(buffer, &index, message.NumRelays)
	encoding.WriteUint32(buffer, &index, message.NumDatacenters)
	encoding.WriteUint32(buffer, &index, message.NumSellers)
	encoding.WriteUint32(buffer, &index, message.NumBuyers)

	return buffer[:index]
}

func (message *AnalyticsDatabaseUpdateMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["database_size"] = int(message.DatabaseSize)
	bigquery_message["num_relays"] = int(message.NumRelays)
	bigquery_message["num_datacenters"] = int(message.NumDatacenters)
	bigquery_message["num_sellers"] = int(message.NumSellers)
	bigquery_message["num_buyers"] = int(message.NumBuyers)

	return bigquery_message, "", nil
}
