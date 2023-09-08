package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/encoding"
)

const (
	AnalyticsSessionUpdateMessageVersion_Min   = 3
	AnalyticsSessionUpdateMessageVersion_Max   = 3
	AnalyticsSessionUpdateMessageVersion_Write = 3
)

type AnalyticsSessionUpdateMessage struct {

	// always

	Version          byte
	Timestamp        uint64
	SessionId        uint64
	SliceNumber      uint32
	RealPacketLoss   float32
	RealJitter       float32
	RealOutOfOrder   float32
	SessionEvents    uint64
	InternalEvents   uint64
	DirectRTT        float32
	DirectJitter     float32
	DirectPacketLoss float32
	DirectKbpsUp     uint32
	DirectKbpsDown   uint32

	// next only

	NextRTT            float32
	NextJitter         float32
	NextPacketLoss     float32
	NextKbpsUp         uint32
	NextKbpsDown       uint32
	NextPredictedRTT   uint32
	NextNumRouteRelays uint32
	NextRouteRelayId   [constants.MaxRouteRelays]uint64

	// flags

	Next                         bool
	Reported                     bool
	LatencyReduction             bool
	PacketLossReduction          bool
	ForceNext                    bool
	LongSessionUpdate            bool
	ClientNextBandwidthOverLimit bool
	ServerNextBandwidthOverLimit bool
}

func (message *AnalyticsSessionUpdateMessage) GetMaxSize() int {
	return 256 + 8*constants.MaxRouteRelays
}

func (message *AnalyticsSessionUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < AnalyticsSessionUpdateMessageVersion_Min || message.Version > AnalyticsSessionUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid session update message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)

	// always

	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint64(buffer, &index, message.SessionId)
	encoding.WriteUint32(buffer, &index, message.SliceNumber)
	encoding.WriteFloat32(buffer, &index, message.RealPacketLoss)
	encoding.WriteFloat32(buffer, &index, message.RealJitter)
	encoding.WriteFloat32(buffer, &index, message.RealOutOfOrder)
	encoding.WriteUint64(buffer, &index, message.SessionEvents)
	encoding.WriteUint64(buffer, &index, message.InternalEvents)
	encoding.WriteFloat32(buffer, &index, message.DirectRTT)
	encoding.WriteFloat32(buffer, &index, message.DirectJitter)
	encoding.WriteFloat32(buffer, &index, message.DirectPacketLoss)
	encoding.WriteUint32(buffer, &index, message.DirectKbpsUp)
	encoding.WriteUint32(buffer, &index, message.DirectKbpsDown)

	// next only

	encoding.WriteBool(buffer, &index, message.Next)
	if message.Next {
		encoding.WriteFloat32(buffer, &index, message.NextRTT)
		encoding.WriteFloat32(buffer, &index, message.NextJitter)
		encoding.WriteFloat32(buffer, &index, message.NextPacketLoss)
		encoding.WriteUint32(buffer, &index, message.NextKbpsUp)
		encoding.WriteUint32(buffer, &index, message.NextKbpsDown)
		encoding.WriteUint32(buffer, &index, message.NextPredictedRTT)
		encoding.WriteUint32(buffer, &index, message.NextNumRouteRelays)
		for i := 0; i < int(message.NextNumRouteRelays); i++ {
			encoding.WriteUint64(buffer, &index, message.NextRouteRelayId[i])
		}
	}

	// flags

	encoding.WriteBool(buffer, &index, message.Reported)
	encoding.WriteBool(buffer, &index, message.LatencyReduction)
	encoding.WriteBool(buffer, &index, message.PacketLossReduction)
	encoding.WriteBool(buffer, &index, message.ForceNext)
	encoding.WriteBool(buffer, &index, message.LongSessionUpdate)
	encoding.WriteBool(buffer, &index, message.ClientNextBandwidthOverLimit)
	encoding.WriteBool(buffer, &index, message.ServerNextBandwidthOverLimit)

	return buffer[:index]
}

func (message *AnalyticsSessionUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read session update message version")
	}

	if message.Version < AnalyticsSessionUpdateMessageVersion_Min || message.Version > AnalyticsSessionUpdateMessageVersion_Max {
		return fmt.Errorf("invalid session update message version %d", message.Version)
	}

	// always

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadUint64(buffer, &index, &message.SessionId) {
		return fmt.Errorf("failed to read session id")
	}

	if !encoding.ReadUint32(buffer, &index, &message.SliceNumber) {
		return fmt.Errorf("failed to read slice number")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RealPacketLoss) {
		return fmt.Errorf("failed to read real packet loss")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RealJitter) {
		return fmt.Errorf("failed to read real jitter")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RealOutOfOrder) {
		return fmt.Errorf("failed to read real out of order")
	}

	if !encoding.ReadUint64(buffer, &index, &message.SessionEvents) {
		return fmt.Errorf("failed to read session events")
	}

	if !encoding.ReadUint64(buffer, &index, &message.InternalEvents) {
		return fmt.Errorf("failed to read internal events")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.DirectRTT) {
		return fmt.Errorf("failed to read direct rtt")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.DirectJitter) {
		return fmt.Errorf("failed to read direct jitter")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.DirectPacketLoss) {
		return fmt.Errorf("failed to read direct packet loss")
	}

	if !encoding.ReadUint32(buffer, &index, &message.DirectKbpsUp) {
		return fmt.Errorf("failed to read direct kbps up")
	}

	if !encoding.ReadUint32(buffer, &index, &message.DirectKbpsDown) {
		return fmt.Errorf("failed to read direct kbps down")
	}

	// next only

	if !encoding.ReadBool(buffer, &index, &message.Next) {
		return fmt.Errorf("failed to read next flag")
	}

	if message.Next {

		if !encoding.ReadFloat32(buffer, &index, &message.NextRTT) {
			return fmt.Errorf("failed to read next rtt")
		}

		if !encoding.ReadFloat32(buffer, &index, &message.NextJitter) {
			return fmt.Errorf("failed to read next jitter")
		}

		if !encoding.ReadFloat32(buffer, &index, &message.NextPacketLoss) {
			return fmt.Errorf("failed to read next packet loss")
		}

		if !encoding.ReadUint32(buffer, &index, &message.NextKbpsUp) {
			return fmt.Errorf("failed to read next kbps up")
		}

		if !encoding.ReadUint32(buffer, &index, &message.NextKbpsDown) {
			return fmt.Errorf("failed to read next kbps down")
		}

		if !encoding.ReadUint32(buffer, &index, &message.NextPredictedRTT) {
			return fmt.Errorf("failed to read next predicted rtt")
		}

		if !encoding.ReadUint32(buffer, &index, &message.NextNumRouteRelays) {
			return fmt.Errorf("failed to read next num route relays")
		}

		for i := 0; i < int(message.NextNumRouteRelays); i++ {
			if !encoding.ReadUint64(buffer, &index, &message.NextRouteRelayId[i]) {
				return fmt.Errorf("failed to read next route relay id")
			}
		}
	}

	// flags

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
	
	return nil
}

func (message *AnalyticsSessionUpdateMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["session_id"] = int(message.SessionId)
	bigquery_message["slice_number"] = int(message.SliceNumber)
	bigquery_message["real_packet_loss"] = float64(message.RealPacketLoss)
	bigquery_message["real_jitter"] = float64(message.RealJitter)
	bigquery_message["real_out_of_order"] = float64(message.RealOutOfOrder)
	bigquery_message["session_events"] = int(message.SessionEvents)
	bigquery_message["internal_events"] = int(message.InternalEvents)
	bigquery_message["direct_rtt"] = float64(message.DirectRTT)
	bigquery_message["direct_jitter"] = float64(message.DirectJitter)
	bigquery_message["direct_packet_loss"] = float64(message.DirectPacketLoss)
	bigquery_message["direct_kbps_up"] = int(message.DirectKbpsUp)
	bigquery_message["direct_kbps_down"] = int(message.DirectKbpsDown)

	if message.Next {

		bigquery_message["next_rtt"] = float64(message.NextRTT)
		bigquery_message["next_jitter"] = float64(message.NextJitter)
		bigquery_message["next_packet_loss"] = float64(message.NextPacketLoss)
		bigquery_message["next_kbps_up"] = int(message.NextKbpsUp)
		bigquery_message["next_kbps_down"] = int(message.NextKbpsDown)
		bigquery_message["next_predicted_rtt"] = int(message.NextPredictedRTT)

		next_route_relays := make([]bigquery.Value, message.NextNumRouteRelays)
		for i := 0; i < int(message.NextNumRouteRelays); i++ {
			next_route_relays[i] = int(message.NextRouteRelayId[i])
		}
		bigquery_message["next_route_relays"] = next_route_relays
	}

	// flags

	bigquery_message["next"] = bool(message.Next)
	bigquery_message["reported"] = bool(message.Reported)
	bigquery_message["latency_reduction"] = bool(message.LatencyReduction)
	bigquery_message["packet_loss_reduction"] = bool(message.PacketLossReduction)
	bigquery_message["force_next"] = bool(message.ForceNext)
	bigquery_message["long_session_update"] = bool(message.LongSessionUpdate)
	bigquery_message["client_next_bandwidth_over_limit"] = bool(message.ClientNextBandwidthOverLimit)
	bigquery_message["server_next_bandwidth_over_limit"] = bool(message.ClientNextBandwidthOverLimit)

	return bigquery_message, "", nil
}
