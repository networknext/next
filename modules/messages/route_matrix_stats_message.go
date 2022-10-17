package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/encoding"
)

const (
	RouteMatrixStatsMessageVersion_Min   = 1
	RouteMatrixStatsMessageVersion_Max   = 1
	RouteMatrixStatsMessageVersion_Write = 1
)

type RouteMatrixStatsMessage struct {
	Version                 uint8
	Timestamp               uint64
	Bytes                   int
	NumRelays               int
	NumDestRelays           int
	NumFullRelays           int
	NumDatacenters          int
	TotalRoutes             int
	AverageNumRoutes        float32
	AverageRouteLength      float32
	NoRoutePercent          float32
	OneRoutePercent         float32
	NoDirectRoutePercent    float32
	RTTBucket_NoImprovement float32
	RTTBucket_0_5ms         float32
	RTTBucket_5_10ms        float32
	RTTBucket_10_15ms       float32
	RTTBucket_15_20ms       float32
	RTTBucket_20_25ms       float32
	RTTBucket_25_30ms       float32
	RTTBucket_30_35ms       float32
	RTTBucket_35_40ms       float32
	RTTBucket_40_45ms       float32
	RTTBucket_45_50ms       float32
	RTTBucket_50ms_Plus     float32
}

func (message *RouteMatrixStatsMessage) Write(buffer []byte) []byte {
	index := 0
	if message.Version < RouteMatrixStatsMessageVersion_Min || message.Version > RouteMatrixStatsMessageVersion_Max {
		panic(fmt.Sprintf("invalid route matrix stats version %d", message.Version))
	}
	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteInt(buffer, &index, message.Bytes)
	encoding.WriteInt(buffer, &index, message.NumRelays)
	encoding.WriteInt(buffer, &index, message.NumDestRelays)
	encoding.WriteInt(buffer, &index, message.NumFullRelays)
	encoding.WriteInt(buffer, &index, message.NumDatacenters)
	encoding.WriteInt(buffer, &index, message.TotalRoutes)
	encoding.WriteFloat32(buffer, &index, message.AverageNumRoutes)
	encoding.WriteFloat32(buffer, &index, message.AverageRouteLength)
	encoding.WriteFloat32(buffer, &index, message.NoRoutePercent)
	encoding.WriteFloat32(buffer, &index, message.OneRoutePercent)
	encoding.WriteFloat32(buffer, &index, message.NoDirectRoutePercent)
	encoding.WriteFloat32(buffer, &index, message.RTTBucket_NoImprovement)
	encoding.WriteFloat32(buffer, &index, message.RTTBucket_0_5ms)
	encoding.WriteFloat32(buffer, &index, message.RTTBucket_5_10ms)
	encoding.WriteFloat32(buffer, &index, message.RTTBucket_10_15ms)
	encoding.WriteFloat32(buffer, &index, message.RTTBucket_15_20ms)
	encoding.WriteFloat32(buffer, &index, message.RTTBucket_20_25ms)
	encoding.WriteFloat32(buffer, &index, message.RTTBucket_25_30ms)
	encoding.WriteFloat32(buffer, &index, message.RTTBucket_30_35ms)
	encoding.WriteFloat32(buffer, &index, message.RTTBucket_35_40ms)
	encoding.WriteFloat32(buffer, &index, message.RTTBucket_40_45ms)
	encoding.WriteFloat32(buffer, &index, message.RTTBucket_45_50ms)
	encoding.WriteFloat32(buffer, &index, message.RTTBucket_50ms_Plus)
	return buffer[:index]
}

func (message *RouteMatrixStatsMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read route matrix stats version")
	}

	if message.Version < RouteMatrixStatsMessageVersion_Min || message.Version > RouteMatrixStatsMessageVersion_Max {
		return fmt.Errorf("invalid route matrix stats version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read route matrix stats timestamp")
	}

	if !encoding.ReadInt(buffer, &index, &message.Bytes) {
		return fmt.Errorf("failed to read route matrix stats bytes")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumRelays) {
		return fmt.Errorf("failed to read route matrix stats num relays")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumDestRelays) {
		return fmt.Errorf("failed to read route matrix stats num dest relays")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumFullRelays) {
		return fmt.Errorf("failed to read route matrix stats num full relays")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumDatacenters) {
		return fmt.Errorf("failed to read route matrix stats num datacenters")
	}

	if !encoding.ReadInt(buffer, &index, &message.TotalRoutes) {
		return fmt.Errorf("failed to read route matrix stats total routes")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.AverageNumRoutes) {
		return fmt.Errorf("failed to read route matrix stats average num routes")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.AverageRouteLength) {
		return fmt.Errorf("failed to read route matrix stats average route length")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.NoRoutePercent) {
		return fmt.Errorf("failed to read route matrix stats no route percent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.OneRoutePercent) {
		return fmt.Errorf("failed to read route matrix stats one route percent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.NoDirectRoutePercent) {
		return fmt.Errorf("failed to read route matrix stats no direct route percent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_NoImprovement) {
		return fmt.Errorf("failed to read route matrix stats RTTBucket_NoImprovement")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_0_5ms) {
		return fmt.Errorf("failed to read route matrix stats RTTBucket_0_5ms")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_5_10ms) {
		return fmt.Errorf("failed to read route matrix stats RTTBucket_5_10ms")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_10_15ms) {
		return fmt.Errorf("failed to read route matrix stats RTTBucket_10_15ms")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_15_20ms) {
		return fmt.Errorf("failed to read route matrix stats RTTBucket_15_20ms")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_20_25ms) {
		return fmt.Errorf("failed to read route matrix stats RTTBucket_20_25ms")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_25_30ms) {
		return fmt.Errorf("failed to read route matrix stats RTTBucket_25_30ms")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_30_35ms) {
		return fmt.Errorf("failed to read route matrix stats RTTBucket_30_35ms")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_35_40ms) {
		return fmt.Errorf("failed to read route matrix stats RTTBucket_35_40ms")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_40_45ms) {
		return fmt.Errorf("failed to read route matrix stats RTTBucket_40_45ms")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_45_50ms) {
		return fmt.Errorf("failed to read route matrix stats RTTBucket_45_50ms")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_50ms_Plus) {
		return fmt.Errorf("failed to read route matrix stats RTTBucket_50ms_Plus")
	}

	return nil
}

func (message *RouteMatrixStatsMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)
	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["bytes"] = int(message.Bytes)
	bigquery_message["numRelays"] = int(message.NumRelays)
	bigquery_message["numDestRelays"] = int(message.NumDestRelays)
	bigquery_message["numFullRelays"] = int(message.NumFullRelays)
	bigquery_message["numDatacenters"] = int(message.NumDatacenters)
	bigquery_message["totalRoutes"] = int(message.TotalRoutes)
	bigquery_message["averageNumRoutes"] = int(message.AverageNumRoutes)
	bigquery_message["averageRouteLength"] = int(message.AverageRouteLength)
	bigquery_message["noRoutePercent"] = int(message.NoRoutePercent)
	bigquery_message["oneRoutePercent"] = int(message.OneRoutePercent)
	bigquery_message["noDirectRoutePercent"] = int(message.NoDirectRoutePercent)
	bigquery_message["rttBucket_NoImprovement"] = int(message.RTTBucket_NoImprovement)
	bigquery_message["rttBucket_0_5ms"] = int(message.RTTBucket_0_5ms)
	bigquery_message["rttBucket_5_10ms"] = int(message.RTTBucket_5_10ms)
	bigquery_message["rttBucket_10_15ms"] = int(message.RTTBucket_10_15ms)
	bigquery_message["rttBucket_15_20ms"] = int(message.RTTBucket_15_20ms)
	bigquery_message["rttBucket_20_25ms"] = int(message.RTTBucket_20_25ms)
	bigquery_message["rttBucket_25_30ms"] = int(message.RTTBucket_25_30ms)
	bigquery_message["rttBucket_30_35ms"] = int(message.RTTBucket_30_35ms)
	bigquery_message["rttBucket_35_40ms"] = int(message.RTTBucket_35_40ms)
	bigquery_message["rttBucket_40_45ms"] = int(message.RTTBucket_40_45ms)
	bigquery_message["rttBucket_45_50ms"] = int(message.RTTBucket_45_50ms)
	bigquery_message["rttBucket_50ms_Plus"] = int(message.RTTBucket_50ms_Plus)
	return bigquery_message, "", nil
}
