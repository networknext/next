package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/encoding"
)

const (
	AnalyticsSessionUpdateMessageVersion_Min   = 1
	AnalyticsSessionUpdateMessageVersion_Max   = 3
	AnalyticsSessionUpdateMessageVersion_Write = 2
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
	SessionFlags     uint64
	SessionEvents    uint64
	InternalEvents   uint64
	DirectRTT        float32
	DirectJitter     float32
	DirectPacketLoss float32
	DirectKbpsUp     uint32
	DirectKbpsDown   uint32

	// next only

	Next               bool
	NextRTT            float32
	NextJitter         float32
	NextPacketLoss     float32
	NextKbpsUp         uint32
	NextKbpsDown       uint32
	NextPredictedRTT   uint32
	NextNumRouteRelays uint32
	NextRouteRelayId   [constants.MaxRouteRelays]uint64

	// flags

	FallbackToDirect   bool
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
	encoding.WriteUint64(buffer, &index, message.SessionFlags)
	encoding.WriteUint64(buffer, &index, message.SessionEvents)
	encoding.WriteUint64(buffer, &index, message.InternalEvents)
	encoding.WriteFloat32(buffer, &index, message.DirectRTT)
	encoding.WriteFloat32(buffer, &index, message.DirectJitter)
	encoding.WriteFloat32(buffer, &index, message.DirectPacketLoss)
	encoding.WriteUint32(buffer, &index, message.DirectKbpsUp)
	encoding.WriteUint32(buffer, &index, message.DirectKbpsDown)

	// next only

	if message.Version >= 2 {

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

	} else {
	
		if (message.SessionFlags & constants.SessionFlags_Next) != 0 {
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

	}

	// flags

	if message.Version >= 3 {

		encoding.WriteBool(buffer, &index, message.FallbackToDirect)

	}

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

	if !encoding.ReadUint64(buffer, &index, &message.SessionFlags) {
		return fmt.Errorf("failed to read session flags")
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

	if message.Version >= 2 {

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

	} else {

		// next only

		if (message.SessionFlags & constants.SessionFlags_Next) != 0 {

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
	}

	if message.Version >= 3 {
		if !encoding.ReadBool(buffer, &index, &message.FallbackToDirect) {
			return fmt.Errorf("failed to read fallback to direct")
		}
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
	bigquery_message["session_flags"] = int(message.SessionFlags)
	bigquery_message["session_events"] = int(message.SessionEvents)
	bigquery_message["internal_events"] = int(message.InternalEvents)
	bigquery_message["direct_rtt"] = float64(message.DirectRTT)
	bigquery_message["direct_jitter"] = float64(message.DirectJitter)
	bigquery_message["direct_packet_loss"] = float64(message.DirectPacketLoss)
	bigquery_message["direct_kbps_up"] = int(message.DirectKbpsUp)
	bigquery_message["direct_kbps_down"] = int(message.DirectKbpsDown)

	bigquery_message["next"] = bool(message.Next)

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

	bigquery_message["fallback_to_direct"] = message.FallbackToDirect

	return bigquery_message, "", nil
}
