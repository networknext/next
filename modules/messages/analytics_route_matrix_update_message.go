package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/accelerate/modules/encoding"
)

const (
	AnalyticsRouteMatrixUpdateMessageVersion_Min   = 1
	AnalyticsRouteMatrixUpdateMessageVersion_Max   = 1
	AnalyticsRouteMatrixUpdateMessageVersion_Write = 1
)

type AnalyticsRouteMatrixUpdateMessage struct {
	Version                 uint8
	Timestamp               uint64
	RouteMatrixSize         int
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

func (message *AnalyticsRouteMatrixUpdateMessage) GetMaxSize() int {
	return 256
}

func (message *AnalyticsRouteMatrixUpdateMessage) Write(buffer []byte) []byte {
	index := 0
	if message.Version < AnalyticsRouteMatrixUpdateMessageVersion_Min || message.Version > AnalyticsRouteMatrixUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid analytics route matrix update message version %d", message.Version))
	}
	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteInt(buffer, &index, message.RouteMatrixSize)
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

func (message *AnalyticsRouteMatrixUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read analytics route matrix update version")
	}

	if message.Version < AnalyticsRouteMatrixUpdateMessageVersion_Min || message.Version > AnalyticsRouteMatrixUpdateMessageVersion_Max {
		return fmt.Errorf("invalid analytics route matrix update message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadInt(buffer, &index, &message.RouteMatrixSize) {
		return fmt.Errorf("failed to read route matrix size")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumRelays) {
		return fmt.Errorf("failed to read num relays")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumDestRelays) {
		return fmt.Errorf("failed to read num dest relays")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumFullRelays) {
		return fmt.Errorf("failed to read num full relays")
	}

	if !encoding.ReadInt(buffer, &index, &message.NumDatacenters) {
		return fmt.Errorf("failed to read num datacenters")
	}

	if !encoding.ReadInt(buffer, &index, &message.TotalRoutes) {
		return fmt.Errorf("failed to read total routes")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.AverageNumRoutes) {
		return fmt.Errorf("failed to read average num routes")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.AverageRouteLength) {
		return fmt.Errorf("failed to read average route length")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.NoRoutePercent) {
		return fmt.Errorf("failed to read no route percent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.OneRoutePercent) {
		return fmt.Errorf("failed to read one route percent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.NoDirectRoutePercent) {
		return fmt.Errorf("failed to read no direct route percent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_NoImprovement) {
		return fmt.Errorf("failed to read rtt bucket no improvement")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_0_5ms) {
		return fmt.Errorf("failed to read rtt bucket 0-5ms improvement")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_5_10ms) {
		return fmt.Errorf("failed to read rtt bucket 5-10ms improvement")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_10_15ms) {
		return fmt.Errorf("failed to read rtt bucket 10-15ms improvement")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_15_20ms) {
		return fmt.Errorf("failed to read rtt bucket 15-20ms improvement")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_20_25ms) {
		return fmt.Errorf("failed to read rtt bucket 20-25ms improvement")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_25_30ms) {
		return fmt.Errorf("failed to read rtt bucket 25-30ms improvement")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_30_35ms) {
		return fmt.Errorf("failed to read rtt bucket 30-35ms improvement")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_35_40ms) {
		return fmt.Errorf("failed to read rtt bucket 35-40ms improvement")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_40_45ms) {
		return fmt.Errorf("failed to read rtt bucket 40-45ms improvement")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_45_50ms) {
		return fmt.Errorf("failed to read rtt bucket 45-50ms improvement")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTTBucket_50ms_Plus) {
		return fmt.Errorf("failed to read rtt bucket 50ms+ improvement")
	}

	return nil
}

func (message *AnalyticsRouteMatrixUpdateMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)
	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["route_matrix_size"] = int(message.RouteMatrixSize)
	bigquery_message["num_relays"] = int(message.NumRelays)
	bigquery_message["num_dest_relays"] = int(message.NumDestRelays)
	bigquery_message["num_full_relays"] = int(message.NumFullRelays)
	bigquery_message["num_datacenters"] = int(message.NumDatacenters)
	bigquery_message["total_routes"] = int(message.TotalRoutes)
	bigquery_message["average_num_routes"] = float64(message.AverageNumRoutes)
	bigquery_message["average_route_length"] = float64(message.AverageRouteLength)
	bigquery_message["no_route_percent"] = float64(message.NoRoutePercent)
	bigquery_message["one_route_percent"] = float64(message.OneRoutePercent)
	bigquery_message["no_direct_route_percent"] = float64(message.NoDirectRoutePercent)
	bigquery_message["rtt_bucket_no_improvement"] = float64(message.RTTBucket_NoImprovement)
	bigquery_message["rtt_bucket_0_5ms"] = float64(message.RTTBucket_0_5ms)
	bigquery_message["rtt_bucket_5_10ms"] = float64(message.RTTBucket_5_10ms)
	bigquery_message["rtt_bucket_10_15ms"] = float64(message.RTTBucket_10_15ms)
	bigquery_message["rtt_bucket_15_20ms"] = float64(message.RTTBucket_15_20ms)
	bigquery_message["rtt_bucket_20_25ms"] = float64(message.RTTBucket_20_25ms)
	bigquery_message["rtt_bucket_25_30ms"] = float64(message.RTTBucket_25_30ms)
	bigquery_message["rtt_bucket_30_35ms"] = float64(message.RTTBucket_30_35ms)
	bigquery_message["rtt_bucket_35_40ms"] = float64(message.RTTBucket_35_40ms)
	bigquery_message["rtt_bucket_40_45ms"] = float64(message.RTTBucket_40_45ms)
	bigquery_message["rtt_bucket_45_50ms"] = float64(message.RTTBucket_45_50ms)
	bigquery_message["rtt_bucket_50ms_plus"] = float64(message.RTTBucket_50ms_Plus)
	return bigquery_message, "", nil
}
