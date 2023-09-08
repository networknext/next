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
	MatchId                         uint64
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
	Error                           uint64
	Reported                        bool
	LatencyReduction                bool
	PacketLossReduction             bool
	ForceNext                       bool
	LongSessionUpdate               bool
	ClientNextBandwidthOverLimit    bool
	ServerNextBandwidthOverLimit    bool
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
	encoding.WriteUint64(buffer, &index, message.MatchId)
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
	encoding.WriteUint64(buffer, &index, message.Error)

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

	if !encoding.ReadUint64(buffer, &index, &message.MatchId) {
		return fmt.Errorf("failed to read match id")
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

	return nil
}

func (message *AnalyticsSessionSummaryMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["session_id"] = int(message.SessionId)
	if message.MatchId != 0 {
		bigquery_message["match_id"] = int(message.MatchId)
	}
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
	bigquery_message["server_to_client_packets_sent"] = int(message.ServerToClientPacketsLost)
	bigquery_message["client_to_server_packets_lost"] = int(message.ClientToServerPacketsLost)
	bigquery_message["server_to_client_packets_lost"] = int(message.ServerToClientPacketsLost)
	bigquery_message["client_to_server_packets_out_of_order"] = int(message.ClientToServerPacketsOutOfOrder)
	bigquery_message["server_to_client_packets_out_of_order"] = int(message.ServerToClientPacketsOutOfOrder)
	bigquery_message["total_next_envelope_bytes_up"] = int(message.TotalNextEnvelopeBytesUp)
	bigquery_message["total_next_envelope_bytes_down"] = int(message.TotalNextEnvelopeBytesDown)
	bigquery_message["duration_on_next"] = int(message.DurationOnNext)
	bigquery_message["session_duration"] = int(message.SessionDuration)
	bigquery_message["start_timestamp"] = int(message.StartTimestamp)
	bigquery_message["error"] = message.Error

	return bigquery_message, "", nil
}
