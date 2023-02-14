package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/encoding"
)

const (
	DatabaseUpdateMessageVersion_Min   = 3
	DatabaseUpdateMessageVersion_Max   = 3
	DatabaseUpdateMessageVersion_Write = 3

	DatabaseUpdateMessageSize = 256
)

type DatabaseUpdateMessage struct {
	Version        uint8
	Timestamp      uint64
	DatabaseSize   uint32
	NumRelays      uint32
	NumDatacenters uint32
	NumSellers     uint32
	NumBuyers      uint32
}

func (message *DatabaseUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read database update version")
	}

	if message.Version < DatabaseUpdateMessageVersion_Min || message.Version > DatabaseUpdateMessageVersion_Max {
		return fmt.Errorf("invalid database update message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read database update timestamp")
	}

	if !encoding.ReadUint32(buffer, &index, &message.DatabaseSize) {
		return fmt.Errorf("failed to read database update database size")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumRelays) {
		return fmt.Errorf("failed to read database update num relays")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumDatacenters) {
		return fmt.Errorf("failed to read database update num datacenters")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumSellers) {
		return fmt.Errorf("failed to read database update num sellers")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumBuyers) {
		return fmt.Errorf("failed to read database update num buyers")
	}

	return nil
}

func (message *DatabaseUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < DatabaseUpdateMessageVersion_Min || message.Version > DatabaseUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid database update message version %d", message.Version))
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

func (message *DatabaseUpdateMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["database_size"] = int(message.DatabaseSize)
	bigquery_message["num_relays"] = int(message.NumRelays)
	bigquery_message["num_datacenters"] = int(message.NumDatacenters)
	bigquery_message["num_sellers"] = int(message.NumSellers)
	bigquery_message["num_buyers"] = int(message.NumBuyers)

	return bigquery_message, "", nil
}
