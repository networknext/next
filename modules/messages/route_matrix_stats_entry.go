package messages

import (
	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/modules/encoding"
)

const RouteMatrixStatsVersion = byte(0)		// IMPORTANT: increase this each time you change the data structure

type RouteMatrixStatsEntry struct {
	Version uint8
	Timestamp  uint64
	Bytes int
	NumRelays int
	NumDestRelays int
	NumFullRelays int
	NumDatacenters int
	TotalRoutes int
	AverageNumRoutes float32
	AverageRouteLength float32
	NoRoutePercent float32
	OneRoutePercent float32
	NoDirectRoutePercent float32
	RTTBucket_NoImprovement float32
	RTTBucket_0_5ms float32
	RTTBucket_5_10ms float32
	RTTBucket_10_15ms float32
	RTTBucket_15_20ms float32
	RTTBucket_20_25ms float32
	RTTBucket_25_30ms float32
	RTTBucket_30_35ms float32
	RTTBucket_35_40ms float32
	RTTBucket_40_45ms float32
	RTTBucket_45_50ms float32
	RTTBucket_50ms_Plus float32
}

func (entry *RouteMatrixStatsEntry) Write(buffer []byte) []byte {
	index := 0
	encoding.WriteUint8(buffer, &index, entry.Version)
	encoding.WriteUint64(buffer, &index, entry.Timestamp)
	encoding.WriteInt(buffer, &index, entry.Bytes)
	encoding.WriteInt(buffer, &index, entry.NumRelays)
	encoding.WriteInt(buffer, &index, entry.NumDestRelays)
	encoding.WriteInt(buffer, &index, entry.NumFullRelays)
	encoding.WriteInt(buffer, &index, entry.NumDatacenters)
	encoding.WriteInt(buffer, &index, entry.TotalRoutes)
	encoding.WriteFloat32(buffer, &index, entry.AverageNumRoutes)
	encoding.WriteFloat32(buffer, &index, entry.AverageRouteLength)
	encoding.WriteFloat32(buffer, &index, entry.NoRoutePercent)
	encoding.WriteFloat32(buffer, &index, entry.OneRoutePercent)
	encoding.WriteFloat32(buffer, &index, entry.NoDirectRoutePercent)
	encoding.WriteFloat32(buffer, &index, entry.RTTBucket_NoImprovement)
	encoding.WriteFloat32(buffer, &index, entry.RTTBucket_0_5ms)
	encoding.WriteFloat32(buffer, &index, entry.RTTBucket_5_10ms)
	encoding.WriteFloat32(buffer, &index, entry.RTTBucket_10_15ms)
	encoding.WriteFloat32(buffer, &index, entry.RTTBucket_15_20ms)
	encoding.WriteFloat32(buffer, &index, entry.RTTBucket_20_25ms)
	encoding.WriteFloat32(buffer, &index, entry.RTTBucket_25_30ms)
	encoding.WriteFloat32(buffer, &index, entry.RTTBucket_30_35ms)
	encoding.WriteFloat32(buffer, &index, entry.RTTBucket_35_40ms)
	encoding.WriteFloat32(buffer, &index, entry.RTTBucket_40_45ms)
	encoding.WriteFloat32(buffer, &index, entry.RTTBucket_45_50ms)
	encoding.WriteFloat32(buffer, &index, entry.RTTBucket_50ms_Plus)
	return buffer[:index]
}

func (entry *RouteMatrixStatsEntry) Read(buffer []byte) error {
	index := 0
	encoding.ReadUint8(buffer, &index, &entry.Version)
	encoding.ReadUint64(buffer, &index, &entry.Timestamp)
	encoding.ReadInt(buffer, &index, &entry.Bytes)
	encoding.ReadInt(buffer, &index, &entry.NumRelays)
	encoding.ReadInt(buffer, &index, &entry.NumDestRelays)
	encoding.ReadInt(buffer, &index, &entry.NumFullRelays)
	encoding.ReadInt(buffer, &index, &entry.NumDatacenters)
	encoding.ReadInt(buffer, &index, &entry.TotalRoutes)
	encoding.ReadFloat32(buffer, &index, &entry.AverageNumRoutes)
	encoding.ReadFloat32(buffer, &index, &entry.AverageRouteLength)
	encoding.ReadFloat32(buffer, &index, &entry.NoRoutePercent)
	encoding.ReadFloat32(buffer, &index, &entry.OneRoutePercent)
	encoding.ReadFloat32(buffer, &index, &entry.NoDirectRoutePercent)
	encoding.ReadFloat32(buffer, &index, &entry.RTTBucket_NoImprovement)
	encoding.ReadFloat32(buffer, &index, &entry.RTTBucket_0_5ms)
	encoding.ReadFloat32(buffer, &index, &entry.RTTBucket_5_10ms)
	encoding.ReadFloat32(buffer, &index, &entry.RTTBucket_10_15ms)
	encoding.ReadFloat32(buffer, &index, &entry.RTTBucket_15_20ms)
	encoding.ReadFloat32(buffer, &index, &entry.RTTBucket_20_25ms)
	encoding.ReadFloat32(buffer, &index, &entry.RTTBucket_25_30ms)
	encoding.ReadFloat32(buffer, &index, &entry.RTTBucket_30_35ms)
	encoding.ReadFloat32(buffer, &index, &entry.RTTBucket_35_40ms)
	encoding.ReadFloat32(buffer, &index, &entry.RTTBucket_40_45ms)
	encoding.ReadFloat32(buffer, &index, &entry.RTTBucket_45_50ms)
	encoding.ReadFloat32(buffer, &index, &entry.RTTBucket_50ms_Plus)
	return nil
}

func (entry *RouteMatrixStatsEntry) Save() (map[string]bigquery.Value, string, error) {

	bigquery_entry := make(map[string]bigquery.Value)
	bigquery_entry["timestamp"] = int(entry.Timestamp)
	bigquery_entry["bytes"] = int(entry.Bytes)
	bigquery_entry["numRelays"] = int(entry.NumRelays)
	bigquery_entry["numDestRelays"] = int(entry.NumDestRelays)
	bigquery_entry["numFullRelays"] = int(entry.NumFullRelays)
	bigquery_entry["numDatacenters"] = int(entry.NumDatacenters)
	bigquery_entry["totalRoutes"] = int(entry.TotalRoutes)
	bigquery_entry["averageNumRoutes"] = int(entry.AverageNumRoutes)
	bigquery_entry["averageRouteLength"] = int(entry.AverageRouteLength)
	bigquery_entry["noRoutePercent"] = int(entry.NoRoutePercent)
	bigquery_entry["oneRoutePercent"] = int(entry.OneRoutePercent)
	bigquery_entry["noDirectRoutePercent"] = int(entry.NoDirectRoutePercent)
	bigquery_entry["rttBucket_NoImprovement"] = int(entry.RTTBucket_NoImprovement)
	bigquery_entry["rttBucket_0_5ms"] = int(entry.RTTBucket_0_5ms)
	bigquery_entry["rttBucket_5_10ms"] = int(entry.RTTBucket_5_10ms)
	bigquery_entry["rttBucket_10_15ms"] = int(entry.RTTBucket_10_15ms)
	bigquery_entry["rttBucket_15_20ms"] = int(entry.RTTBucket_15_20ms)
	bigquery_entry["rttBucket_20_25ms"] = int(entry.RTTBucket_20_25ms)
	bigquery_entry["rttBucket_25_30ms"] = int(entry.RTTBucket_25_30ms)
	bigquery_entry["rttBucket_30_35ms"] = int(entry.RTTBucket_30_35ms)
	bigquery_entry["rttBucket_35_40ms"] = int(entry.RTTBucket_35_40ms)
	bigquery_entry["rttBucket_40_45ms"] = int(entry.RTTBucket_40_45ms)
	bigquery_entry["rttBucket_45_50ms"] = int(entry.RTTBucket_45_50ms)
	bigquery_entry["rttBucket_50ms_Plus"] = int(entry.RTTBucket_50ms_Plus)
	return bigquery_entry, "", nil
}
