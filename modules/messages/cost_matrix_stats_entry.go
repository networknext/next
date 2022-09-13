package messages

import (
	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/modules/encoding"
)

const CostMatrixStatsVersion = byte(0)		// IMPORTANT: increase this each time you change the data structure

type CostMatrixStatsEntry struct {
	Version         byte
	Timestamp  		uint64
	Bytes 			int
	NumRelays       int
	NumDestRelays   int
	NumDatacenters  int
}

func (entry *CostMatrixStatsEntry) Write(buffer []byte) []byte {
	index := 0
	encoding.WriteUint8(buffer, &index, entry.Version)
	encoding.WriteUint64(buffer, &index, entry.Timestamp)
	encoding.WriteInt(buffer, &index, entry.Bytes)
	encoding.WriteInt(buffer, &index, entry.NumRelays)
	encoding.WriteInt(buffer, &index, entry.NumDatacenters)
	return buffer[:index]
}

func (entry *CostMatrixStatsEntry) Read(buffer []byte) error {
	index := 0
	encoding.ReadUint8(buffer, &index, &entry.Version)
	encoding.ReadUint64(buffer, &index, &entry.Timestamp)
	encoding.ReadInt(buffer, &index, &entry.Bytes)
	encoding.ReadInt(buffer, &index, &entry.NumRelays)
	encoding.ReadInt(buffer, &index, &entry.NumDestRelays)
	encoding.ReadInt(buffer, &index, &entry.NumDatacenters)
	return nil
}

func (entry *CostMatrixStatsEntry) Save() (map[string]bigquery.Value, string, error) {
	bigquery_entry := make(map[string]bigquery.Value)
	bigquery_entry["timestamp"] = int(entry.Timestamp)
	bigquery_entry["bytes"] = int(entry.Bytes)
	bigquery_entry["numRelays"] = int(entry.NumRelays)
	bigquery_entry["numDestRelays"] = int(entry.NumDestRelays)
	bigquery_entry["numDatacenters"] = int(entry.NumDatacenters)
	return bigquery_entry, "", nil
}
