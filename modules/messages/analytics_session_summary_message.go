package messages

import (
	"fmt"
	"net"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/next/modules/encoding"
)

const (
	AnalyticsSessionSummaryMessageVersion_Min   = 0
	AnalyticsSessionSummaryMessageVersion_Max   = 0
	AnalyticsSessionSummaryMessageVersion_Write = 0
)

type AnalyticsSessionSummaryMessage struct {
	Version                         uint8
	Timestamp                       uint64
	SessionId                       uint64
	DatacenterId                    uint64
	BuyerId                         uint64
	UserHash                        uint64
	Latitude                        float32
	Longitude                       float32
	ClientAddress                   net.UDPAddr
	ServerAddress                   net.UDPAddr
	ConnectionType                  byte
	PlatformType                    byte
	SDKVersion_Major                byte
	SDKVersion_Minor                byte
	SDKVersion_Patch                byte
	ClientToServerPacketsSent       uint64
	ServerToClientPacketsSent       uint64
	ClientToServerPacketsLost       uint64
	ServerToClientPacketsLost       uint64
	ClientToServerPacketsOutOfOrder uint64
	ServerToClientPacketsOutOfOrder uint64
	TotalNextEnvelopeBytesUp        uint64
	TotalNextEnvelopeBytesDown      uint64
	DurationOnNext                  uint32
	SessionDuration                 uint32
	StartTimestamp                  uint64

	// flags

	Error                        uint64
	Reported                     bool
	LatencyReduction             bool
	PacketLossReduction          bool
	ForceNext                    bool
	LongSessionUpdate            bool
	ClientNextBandwidthOverLimit bool
	ServerNextBandwidthOverLimit bool
	Veto                         bool
	Disabled                     bool
	NotSelected                  bool
	A                            bool
	B                            bool
	LatencyWorse                 bool
	LocationVeto                 bool
	Mispredict                   bool
	LackOfDiversity              bool
}

func (message *AnalyticsSessionSummaryMessage) GetMaxSize() int {
	return 512
}

func (message *AnalyticsSessionSummaryMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < AnalyticsSessionSummaryMessageVersion_Min || message.Version > AnalyticsSessionSummaryMessageVersion_Max {
		panic(fmt.Sprintf("invalid analytics session summary message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint64(buffer, &index, message.SessionId)
	encoding.WriteUint64(buffer, &index, message.DatacenterId)
	encoding.WriteUint64(buffer, &index, message.BuyerId)
	encoding.WriteUint64(buffer, &index, message.UserHash)
	encoding.WriteFloat32(buffer, &index, message.Latitude)
	encoding.WriteFloat32(buffer, &index, message.Longitude)
	encoding.WriteAddress(buffer, &index, &message.ClientAddress)
	encoding.WriteAddress(buffer, &index, &message.ServerAddress)
	encoding.WriteUint8(buffer, &index, message.ConnectionType)
	encoding.WriteUint8(buffer, &index, message.PlatformType)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Major)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Minor)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Patch)
	encoding.WriteUint64(buffer, &index, message.ClientToServerPacketsSent)
	encoding.WriteUint64(buffer, &index, message.ServerToClientPacketsSent)
	encoding.WriteUint64(buffer, &index, message.ClientToServerPacketsLost)
	encoding.WriteUint64(buffer, &index, message.ServerToClientPacketsLost)
	encoding.WriteUint64(buffer, &index, message.ClientToServerPacketsOutOfOrder)
	encoding.WriteUint64(buffer, &index, message.ServerToClientPacketsOutOfOrder)
	encoding.WriteUint32(buffer, &index, message.SessionDuration)
	encoding.WriteUint64(buffer, &index, message.TotalNextEnvelopeBytesUp)
	encoding.WriteUint64(buffer, &index, message.TotalNextEnvelopeBytesDown)
	encoding.WriteUint32(buffer, &index, message.DurationOnNext)
	encoding.WriteUint64(buffer, &index, message.StartTimestamp)

	// flags

	encoding.WriteUint64(buffer, &index, message.Error)
	encoding.WriteBool(buffer, &index, message.Reported)
	encoding.WriteBool(buffer, &index, message.LatencyReduction)
	encoding.WriteBool(buffer, &index, message.PacketLossReduction)
	encoding.WriteBool(buffer, &index, message.ForceNext)
	encoding.WriteBool(buffer, &index, message.LongSessionUpdate)
	encoding.WriteBool(buffer, &index, message.ClientNextBandwidthOverLimit)
	encoding.WriteBool(buffer, &index, message.ServerNextBandwidthOverLimit)
	encoding.WriteBool(buffer, &index, message.Veto)
	encoding.WriteBool(buffer, &index, message.Disabled)
	encoding.WriteBool(buffer, &index, message.NotSelected)
	encoding.WriteBool(buffer, &index, message.A)
	encoding.WriteBool(buffer, &index, message.B)
	encoding.WriteBool(buffer, &index, message.LatencyWorse)
	encoding.WriteBool(buffer, &index, message.LocationVeto)
	encoding.WriteBool(buffer, &index, message.Mispredict)
	encoding.WriteBool(buffer, &index, message.LackOfDiversity)

	return buffer[:index]
}

func (message *AnalyticsSessionSummaryMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read analytics session summary message version")
	}

	if message.Version < AnalyticsSessionSummaryMessageVersion_Min || message.Version > AnalyticsSessionSummaryMessageVersion_Max {
		return fmt.Errorf("invalid session analytics summary message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadUint64(buffer, &index, &message.SessionId) {
		return fmt.Errorf("failed to read session id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.DatacenterId) {
		return fmt.Errorf("failed to read datacenter id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.BuyerId) {
		return fmt.Errorf("failed to read buyer id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.UserHash) {
		return fmt.Errorf("failed to read user hash")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.Latitude) {
		return fmt.Errorf("failed to read latitude")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.Longitude) {
		return fmt.Errorf("failed to read longitude")
	}

	if !encoding.ReadAddress(buffer, &index, &message.ClientAddress) {
		return fmt.Errorf("failed to read client address")
	}

	if !encoding.ReadAddress(buffer, &index, &message.ServerAddress) {
		return fmt.Errorf("failed to read server address")
	}

	if !encoding.ReadUint8(buffer, &index, &message.ConnectionType) {
		return fmt.Errorf("failed to read connection type")
	}

	if !encoding.ReadUint8(buffer, &index, &message.PlatformType) {
		return fmt.Errorf("failed to read platform type")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Major) {
		return fmt.Errorf("failed to read sdk version major")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Minor) {
		return fmt.Errorf("failed to read sdk version minor")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Patch) {
		return fmt.Errorf("failed to read sdk version patch")
	}

	if !encoding.ReadUint64(buffer, &index, &message.ClientToServerPacketsSent) {
		return fmt.Errorf("failed to read client to server packets sent")
	}

	if !encoding.ReadUint64(buffer, &index, &message.ServerToClientPacketsSent) {
		return fmt.Errorf("failed to read server to client packets sent")
	}

	if !encoding.ReadUint64(buffer, &index, &message.ClientToServerPacketsLost) {
		return fmt.Errorf("failed to read client to server packets lost")
	}

	if !encoding.ReadUint64(buffer, &index, &message.ServerToClientPacketsLost) {
		return fmt.Errorf("failed to read server to client packets lost")
	}

	if !encoding.ReadUint64(buffer, &index, &message.ClientToServerPacketsOutOfOrder) {
		return fmt.Errorf("failed to read client to server packets out of order")
	}

	if !encoding.ReadUint64(buffer, &index, &message.ServerToClientPacketsOutOfOrder) {
		return fmt.Errorf("failed to read server to client packets out of order")
	}

	if !encoding.ReadUint32(buffer, &index, &message.SessionDuration) {
		return fmt.Errorf("failed to read session duration")
	}

	if !encoding.ReadUint64(buffer, &index, &message.TotalNextEnvelopeBytesUp) {
		return fmt.Errorf("failed to read total next envelope bytes up sum")
	}

	if !encoding.ReadUint64(buffer, &index, &message.TotalNextEnvelopeBytesDown) {
		return fmt.Errorf("failed to read total next envelope bytes down sum")
	}

	if !encoding.ReadUint32(buffer, &index, &message.DurationOnNext) {
		return fmt.Errorf("failed to read duration on next")
	}

	if !encoding.ReadUint64(buffer, &index, &message.StartTimestamp) {
		return fmt.Errorf("failed to read start timestamp")
	}

	if !encoding.ReadUint64(buffer, &index, &message.Error) {
		return fmt.Errorf("failed to read error")
	}

	if !encoding.ReadBool(buffer, &index, &message.Reported) {
		return fmt.Errorf("failed to read reported flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.LatencyReduction) {
		return fmt.Errorf("failed to read latency reduction flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.PacketLossReduction) {
		return fmt.Errorf("failed to read latency packet loss reduction flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.ForceNext) {
		return fmt.Errorf("failed to read force next flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.LongSessionUpdate) {
		return fmt.Errorf("failed to read long session update flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.ClientNextBandwidthOverLimit) {
		return fmt.Errorf("failed to read client next bandwidth over limit flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.ServerNextBandwidthOverLimit) {
		return fmt.Errorf("failed to read server next bandwidth over limit flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.Veto) {
		return fmt.Errorf("failed to read veto flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.Disabled) {
		return fmt.Errorf("failed to read disabled flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.NotSelected) {
		return fmt.Errorf("failed to read not selected flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.A) {
		return fmt.Errorf("failed to read A flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.B) {
		return fmt.Errorf("failed to read B flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.LatencyWorse) {
		return fmt.Errorf("failed to read latency worse flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.LocationVeto) {
		return fmt.Errorf("failed to read location veto flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.Mispredict) {
		return fmt.Errorf("failed to read mispredict flag")
	}

	if !encoding.ReadBool(buffer, &index, &message.LackOfDiversity) {
		return fmt.Errorf("failed to read lack of diversity flag")
	}

	return nil
}

func (message *AnalyticsSessionSummaryMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["session_id"] = int(message.SessionId)
	bigquery_message["datacenter_id"] = int(message.DatacenterId)
	bigquery_message["buyer_id"] = int(message.BuyerId)
	bigquery_message["user_hash"] = int(message.UserHash)
	bigquery_message["latitude"] = float64(message.Latitude)
	bigquery_message["longitude"] = float64(message.Longitude)
	bigquery_message["client_address"] = message.ClientAddress.String()
	bigquery_message["server_address"] = message.ServerAddress.String()
	bigquery_message["connection_type"] = int(message.ConnectionType)
	bigquery_message["platform_type"] = int(message.PlatformType)
	bigquery_message["sdk_version_major"] = int(message.SDKVersion_Major)
	bigquery_message["sdk_version_minor"] = int(message.SDKVersion_Minor)
	bigquery_message["sdk_version_patch"] = int(message.SDKVersion_Patch)
	bigquery_message["client_to_server_packets_sent"] = int(message.ClientToServerPacketsSent)
	bigquery_message["server_to_client_packets_sent"] = int(message.ServerToClientPacketsSent)
	bigquery_message["client_to_server_packets_lost"] = int(message.ClientToServerPacketsLost)
	bigquery_message["server_to_client_packets_lost"] = int(message.ServerToClientPacketsLost)
	bigquery_message["client_to_server_packets_out_of_order"] = int(message.ClientToServerPacketsOutOfOrder)
	bigquery_message["server_to_client_packets_out_of_order"] = int(message.ServerToClientPacketsOutOfOrder)
	bigquery_message["total_next_envelope_bytes_up"] = int(message.TotalNextEnvelopeBytesUp)
	bigquery_message["total_next_envelope_bytes_down"] = int(message.TotalNextEnvelopeBytesDown)
	bigquery_message["duration_on_next"] = int(message.DurationOnNext)
	bigquery_message["session_duration"] = int(message.SessionDuration)
	bigquery_message["start_timestamp"] = int(message.StartTimestamp)

	// flags

	bigquery_message["error"] = message.Error
	bigquery_message["reported"] = bool(message.Reported)
	bigquery_message["latency_reduction"] = bool(message.LatencyReduction)
	bigquery_message["packet_loss_reduction"] = bool(message.PacketLossReduction)
	bigquery_message["force_next"] = bool(message.ForceNext)
	bigquery_message["long_session_update"] = bool(message.LongSessionUpdate)
	bigquery_message["client_next_bandwidth_over_limit"] = bool(message.ClientNextBandwidthOverLimit)
	bigquery_message["server_next_bandwidth_over_limit"] = bool(message.ClientNextBandwidthOverLimit)
	bigquery_message["veto"] = bool(message.Veto)
	bigquery_message["disabled"] = bool(message.Disabled)
	bigquery_message["not_selected"] = bool(message.NotSelected)
	bigquery_message["a"] = bool(message.A)
	bigquery_message["b"] = bool(message.B)
	bigquery_message["latency_worse"] = bool(message.LatencyWorse)
	bigquery_message["location_veto"] = bool(message.LocationVeto)
	bigquery_message["mispredict"] = bool(message.Mispredict)
	bigquery_message["lack_of_diversity"] = bool(message.LackOfDiversity)

	return bigquery_message, "", nil
}
