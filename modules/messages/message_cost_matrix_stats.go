package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/modules-old/encoding"
)

const CostMatrixStatsMessageVersion = byte(0) // IMPORTANT: increase this each time you change the data structure

type CostMatrixStatsMessage struct {
	Version        byte
	Timestamp      uint64
	Bytes          int
	NumRelays      int
	NumDestRelays  int
	NumDatacenters int
}

func (message *CostMatrixStatsMessage) Write(buffer []byte) []byte {
	index := 0
	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteInt(buffer, &index, message.Bytes)
	encoding.WriteInt(buffer, &index, message.NumRelays)
	encoding.WriteInt(buffer, &index, message.NumDestRelays)
	encoding.WriteInt(buffer, &index, message.NumDatacenters)
	return buffer[:index]
}

func (message *CostMatrixStatsMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read cost matrix stats Version")
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read cost matrix stats Timestamp")
	}

	if !encoding.ReadInt(buffer, &index, &message.Bytes) {
		return fmt.Errorf("failed to read cost matrix stats Bytes")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumRelays) {
		return fmt.Errorf("failed to read cost matrix stats NumRelays")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumDestRelays) {
		return fmt.Errorf("failed to read cost matrix stats NumDestRelays")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumDatacenters) {
		return fmt.Errorf("failed to read cost matrix stats NumDatacenters")
	}

	return nil
}

func (message *CostMatrixStatsMessage) Save() (map[string]bigquery.Value, string, error) {
	bigquery_entry := make(map[string]bigquery.Value)
	bigquery_entry["timestamp"] = int(message.Timestamp)
	bigquery_entry["bytes"] = int(message.Bytes)
	bigquery_entry["numRelays"] = int(message.NumRelays)
	bigquery_entry["numDestRelays"] = int(message.NumDestRelays)
	bigquery_entry["numDatacenters"] = int(message.NumDatacenters)
	return bigquery_entry, "", nil
}
