package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/next/modules/encoding"
)

const (
	AnalyticsCostMatrixUpdateMessageVersion_Min   = 1
	AnalyticsCostMatrixUpdateMessageVersion_Max   = 1
	AnalyticsCostMatrixUpdateMessageVersion_Write = 1
)

type AnalyticsCostMatrixUpdateMessage struct {
	Version        byte
	Timestamp      uint64
	CostMatrixSize int
	NumRelays      int
	NumDestRelays  int
	NumDatacenters int
}

func (message *AnalyticsCostMatrixUpdateMessage) GetMaxSize() int {
	return 64
}

func (message *AnalyticsCostMatrixUpdateMessage) Write(buffer []byte) []byte {
	index := 0
	if message.Version < AnalyticsCostMatrixUpdateMessageVersion_Min || message.Version > AnalyticsCostMatrixUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid analytics cost matrix update version %d", message.Version))
	}
	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteInt(buffer, &index, message.CostMatrixSize)
	encoding.WriteInt(buffer, &index, message.NumRelays)
	encoding.WriteInt(buffer, &index, message.NumDestRelays)
	encoding.WriteInt(buffer, &index, message.NumDatacenters)
	return buffer[:index]
}

func (message *AnalyticsCostMatrixUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read analytics cost matrix update version")
	}

	if message.Version < AnalyticsCostMatrixUpdateMessageVersion_Min || message.Version > AnalyticsCostMatrixUpdateMessageVersion_Max {
		return fmt.Errorf("invalid analytics cost matrix update version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadInt(buffer, &index, &message.CostMatrixSize) {
		return fmt.Errorf("failed to read cost matrix size")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumRelays) {
		return fmt.Errorf("failed to read num relays")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumDestRelays) {
		return fmt.Errorf("failed to read num dest relays")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumDatacenters) {
		return fmt.Errorf("failed to read num datacenters")
	}

	return nil
}

func (message *AnalyticsCostMatrixUpdateMessage) Save() (map[string]bigquery.Value, string, error) {
	bigquery_entry := make(map[string]bigquery.Value)
	bigquery_entry["timestamp"] = int(message.Timestamp)
	bigquery_entry["cost_matrix_size"] = int(message.CostMatrixSize)
	bigquery_entry["num_relays"] = int(message.NumRelays)
	bigquery_entry["num_dest_relays"] = int(message.NumDestRelays)
	bigquery_entry["num_datacenters"] = int(message.NumDatacenters)
	return bigquery_entry, "", nil
}
