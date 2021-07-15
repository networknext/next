package routing

import (
	"errors"
	"fmt"

	"github.com/networknext/backend/modules/encoding"
)

// TrafficStats describes the measured relay traffic statistics reported from the relay
type TrafficStats struct {
	SessionCount     uint64
	EnvelopeUpKbps   uint64
	EnvelopeDownKbps uint64

	OutboundPingTx uint64

	RouteRequestRx uint64
	RouteRequestTx uint64

	RouteResponseRx uint64
	RouteResponseTx uint64

	ClientToServerRx uint64
	ClientToServerTx uint64

	ServerToClientRx uint64
	ServerToClientTx uint64

	InboundPingRx uint64
	InboundPingTx uint64

	PongRx uint64

	SessionPingRx uint64
	SessionPingTx uint64

	SessionPongRx uint64
	SessionPongTx uint64

	ContinueRequestRx uint64
	ContinueRequestTx uint64

	ContinueResponseRx uint64
	ContinueResponseTx uint64

	NearPingRx uint64
	NearPingTx uint64

	UnknownRx uint64

	BytesSent     uint64
	BytesReceived uint64
}

func (rts *TrafficStats) Add(other *TrafficStats) TrafficStats {
	return TrafficStats{
		SessionCount: rts.SessionCount + other.SessionCount,

		EnvelopeUpKbps:   rts.EnvelopeUpKbps + other.EnvelopeUpKbps,
		EnvelopeDownKbps: rts.EnvelopeDownKbps + other.EnvelopeDownKbps,

		OutboundPingTx: rts.OutboundPingTx + other.OutboundPingTx,

		RouteRequestRx: rts.RouteRequestRx + other.RouteRequestRx,
		RouteRequestTx: rts.RouteRequestTx + other.RouteRequestTx,

		RouteResponseRx: rts.RouteResponseRx + other.RouteResponseRx,
		RouteResponseTx: rts.RouteResponseTx + other.RouteResponseTx,

		ClientToServerRx: rts.ClientToServerRx + other.ClientToServerRx,
		ClientToServerTx: rts.ClientToServerTx + other.ClientToServerTx,

		ServerToClientRx: rts.ServerToClientRx + other.ServerToClientRx,
		ServerToClientTx: rts.ServerToClientTx + other.ServerToClientTx,

		InboundPingRx: rts.InboundPingRx + other.InboundPingRx,
		InboundPingTx: rts.InboundPingTx + other.InboundPingTx,

		PongRx: rts.PongRx + other.PongRx,

		SessionPingRx: rts.SessionPingRx + other.SessionPingRx,
		SessionPingTx: rts.SessionPingTx + other.SessionPingTx,

		SessionPongRx: rts.SessionPongRx + other.SessionPongRx,
		SessionPongTx: rts.SessionPongTx + other.SessionPongTx,

		ContinueRequestRx: rts.ContinueRequestRx + other.ContinueRequestRx,
		ContinueRequestTx: rts.ContinueRequestTx + other.ContinueRequestTx,

		ContinueResponseRx: rts.ContinueResponseRx + other.ContinueResponseRx,
		ContinueResponseTx: rts.ContinueResponseTx + other.ContinueResponseTx,

		NearPingRx: rts.NearPingRx + other.NearPingRx,
		NearPingTx: rts.NearPingTx + other.NearPingTx,

		UnknownRx: rts.UnknownRx + other.UnknownRx,

		BytesSent:     rts.BytesSent + other.BytesSent,
		BytesReceived: rts.BytesReceived + other.BytesReceived,
	}
}

// OtherStatsRx returns the relay to relay rx stats
func (rts *TrafficStats) OtherStatsRx() uint64 {
	return rts.PongRx + rts.InboundPingRx
}

// OtherStatsTx returns the relay to relay tx stats
func (rts *TrafficStats) OtherStatsTx() uint64 {
	return rts.OutboundPingTx + rts.InboundPingTx
}

// GameStatsRx returns the game <-> relay rx stats
func (rts *TrafficStats) GameStatsRx() uint64 {
	return rts.RouteRequestRx + rts.RouteResponseRx + rts.ClientToServerRx + rts.ServerToClientRx + rts.SessionPingRx + rts.SessionPongRx + rts.ContinueRequestRx + rts.ContinueResponseRx + rts.NearPingRx
}

// GameStatsTx returns the game <-> relay tx stats
func (rts *TrafficStats) GameStatsTx() uint64 {
	return rts.RouteRequestTx + rts.RouteResponseTx + rts.ClientToServerTx + rts.ServerToClientTx + rts.SessionPingTx + rts.SessionPongTx + rts.ContinueRequestTx + rts.ContinueResponseTx + rts.NearPingTx
}

func (rts *TrafficStats) AllRx() uint64 {
	return rts.OtherStatsRx() + rts.GameStatsRx() + rts.UnknownRx
}

func (rts *TrafficStats) AllTx() uint64 {
	return rts.OtherStatsTx() + rts.GameStatsTx()
}

func (rts *TrafficStats) WriteTo(data []byte, index *int, version uint8) error {
	switch version {
	case 0:
		rts.writeToV0(data, index)
	case 1:
		rts.writeToV1(data, index)
	case 2:
		rts.writeToV2(data, index)
	default:
		return fmt.Errorf("invalid traffic stats version: %d", version)
	}

	return nil
}

func (rts *TrafficStats) writeToV0(data []byte, index *int) {
	encoding.WriteUint64(data, index, rts.SessionCount)
	encoding.WriteUint64(data, index, rts.BytesSent)
	encoding.WriteUint64(data, index, rts.BytesReceived)
}

func (rts *TrafficStats) writeToV1(data []byte, index *int) {
	encoding.WriteUint64(data, index, rts.SessionCount)
	encoding.WriteUint64(data, index, rts.OutboundPingTx)
	encoding.WriteUint64(data, index, rts.RouteRequestRx)
	encoding.WriteUint64(data, index, rts.RouteRequestTx)
	encoding.WriteUint64(data, index, rts.RouteResponseRx)
	encoding.WriteUint64(data, index, rts.RouteResponseTx)
	encoding.WriteUint64(data, index, rts.ClientToServerRx)
	encoding.WriteUint64(data, index, rts.ClientToServerTx)
	encoding.WriteUint64(data, index, rts.ServerToClientRx)
	encoding.WriteUint64(data, index, rts.ServerToClientTx)
	encoding.WriteUint64(data, index, rts.InboundPingRx)
	encoding.WriteUint64(data, index, rts.InboundPingTx)
	encoding.WriteUint64(data, index, rts.PongRx)
	encoding.WriteUint64(data, index, rts.SessionPingRx)
	encoding.WriteUint64(data, index, rts.SessionPingTx)
	encoding.WriteUint64(data, index, rts.SessionPongRx)
	encoding.WriteUint64(data, index, rts.SessionPongTx)
	encoding.WriteUint64(data, index, rts.ContinueRequestRx)
	encoding.WriteUint64(data, index, rts.ContinueRequestTx)
	encoding.WriteUint64(data, index, rts.ContinueResponseRx)
	encoding.WriteUint64(data, index, rts.ContinueResponseTx)
	encoding.WriteUint64(data, index, rts.NearPingRx)
	encoding.WriteUint64(data, index, rts.NearPingTx)
	encoding.WriteUint64(data, index, rts.UnknownRx)
}

func (rts *TrafficStats) writeToV2(data []byte, index *int) {
	encoding.WriteUint64(data, index, rts.SessionCount)
	encoding.WriteUint64(data, index, rts.EnvelopeUpKbps)
	encoding.WriteUint64(data, index, rts.EnvelopeDownKbps)
	encoding.WriteUint64(data, index, rts.OutboundPingTx)
	encoding.WriteUint64(data, index, rts.RouteRequestRx)
	encoding.WriteUint64(data, index, rts.RouteRequestTx)
	encoding.WriteUint64(data, index, rts.RouteResponseRx)
	encoding.WriteUint64(data, index, rts.RouteResponseTx)
	encoding.WriteUint64(data, index, rts.ClientToServerRx)
	encoding.WriteUint64(data, index, rts.ClientToServerTx)
	encoding.WriteUint64(data, index, rts.ServerToClientRx)
	encoding.WriteUint64(data, index, rts.ServerToClientTx)
	encoding.WriteUint64(data, index, rts.InboundPingRx)
	encoding.WriteUint64(data, index, rts.InboundPingTx)
	encoding.WriteUint64(data, index, rts.PongRx)
	encoding.WriteUint64(data, index, rts.SessionPingRx)
	encoding.WriteUint64(data, index, rts.SessionPingTx)
	encoding.WriteUint64(data, index, rts.SessionPongRx)
	encoding.WriteUint64(data, index, rts.SessionPongTx)
	encoding.WriteUint64(data, index, rts.ContinueRequestRx)
	encoding.WriteUint64(data, index, rts.ContinueRequestTx)
	encoding.WriteUint64(data, index, rts.ContinueResponseRx)
	encoding.WriteUint64(data, index, rts.ContinueResponseTx)
	encoding.WriteUint64(data, index, rts.NearPingRx)
	encoding.WriteUint64(data, index, rts.NearPingTx)
	encoding.WriteUint64(data, index, rts.UnknownRx)
}

func (rts *TrafficStats) ReadFrom(data []byte, index *int, version uint8) error {
	switch version {
	case 0:
		return rts.readFromV0(data, index)
	case 1:
		return rts.readFromV1(data, index)
	case 2:
		return rts.readFromV2(data, index)
	default:
		return fmt.Errorf("invalid traffic stats version: %d", version)
	}
}

func (rts *TrafficStats) readFromV0(data []byte, index *int) error {
	if !encoding.ReadUint64(data, index, &rts.SessionCount) {
		return errors.New("invalid data, could not read session count")
	}

	if !encoding.ReadUint64(data, index, &rts.BytesSent) {
		return errors.New("invalid data, could not read bytes sent")
	}

	if !encoding.ReadUint64(data, index, &rts.BytesReceived) {
		return errors.New("invalid data, could not read bytes received")
	}

	return nil
}
func (rts *TrafficStats) readFromV1(data []byte, index *int) error {
	if !encoding.ReadUint64(data, index, &rts.SessionCount) {
		return errors.New("invalid data, could not read session count")
	}

	if !encoding.ReadUint64(data, index, &rts.OutboundPingTx) {
		return errors.New("invalid data, could not read outbound ping tx")
	}

	if !encoding.ReadUint64(data, index, &rts.RouteRequestRx) {
		return errors.New("invalid data, could not read route request rx")
	}
	if !encoding.ReadUint64(data, index, &rts.RouteRequestTx) {
		return errors.New("invalid data, could not read route request tx")
	}

	if !encoding.ReadUint64(data, index, &rts.RouteResponseRx) {
		return errors.New("invalid data, could not read route response rx")
	}
	if !encoding.ReadUint64(data, index, &rts.RouteResponseTx) {
		return errors.New("invalid data, could not read route response tx")
	}

	if !encoding.ReadUint64(data, index, &rts.ClientToServerRx) {
		return errors.New("invalid data, could not read client to server rx")
	}
	if !encoding.ReadUint64(data, index, &rts.ClientToServerTx) {
		return errors.New("invalid data, could not read client to server tx")
	}

	if !encoding.ReadUint64(data, index, &rts.ServerToClientRx) {
		return errors.New("invalid data, could not read server to client rx")
	}
	if !encoding.ReadUint64(data, index, &rts.ServerToClientTx) {
		return errors.New("invalid data, could not read server to client tx")
	}

	if !encoding.ReadUint64(data, index, &rts.InboundPingRx) {
		return errors.New("invalid data, could not read inbound ping rx")
	}
	if !encoding.ReadUint64(data, index, &rts.InboundPingTx) {
		return errors.New("invalid data, could not read inbound ping tx")
	}

	if !encoding.ReadUint64(data, index, &rts.PongRx) {
		return errors.New("invalid data, could not read pong rx")
	}

	if !encoding.ReadUint64(data, index, &rts.SessionPingRx) {
		return errors.New("invalid data, could not read session ping rx")
	}
	if !encoding.ReadUint64(data, index, &rts.SessionPingTx) {
		return errors.New("invalid data, could not read session ping tx")
	}

	if !encoding.ReadUint64(data, index, &rts.SessionPongRx) {
		return errors.New("invalid data, could not read session pong rx")
	}
	if !encoding.ReadUint64(data, index, &rts.SessionPongTx) {
		return errors.New("invalid data, could not read session pong tx")
	}

	if !encoding.ReadUint64(data, index, &rts.ContinueRequestRx) {
		return errors.New("invalid data, could not read continue request rx")
	}
	if !encoding.ReadUint64(data, index, &rts.ContinueRequestTx) {
		return errors.New("invalid data, could not read continue request tx")
	}

	if !encoding.ReadUint64(data, index, &rts.ContinueResponseRx) {
		return errors.New("invalid data, could not read continue response rx")
	}
	if !encoding.ReadUint64(data, index, &rts.ContinueResponseTx) {
		return errors.New("invalid data, could not read continue response tx")
	}

	if !encoding.ReadUint64(data, index, &rts.NearPingRx) {
		return errors.New("invalid data, could not read near ping rx")
	}
	if !encoding.ReadUint64(data, index, &rts.NearPingTx) {
		return errors.New("invalid data, could not read near ping tx")
	}

	if !encoding.ReadUint64(data, index, &rts.UnknownRx) {
		return errors.New("invalid data, could not read unknown rx")
	}

	return nil
}

func (rts *TrafficStats) readFromV2(data []byte, index *int) error {
	if !encoding.ReadUint64(data, index, &rts.SessionCount) {
		return errors.New("invalid data, could not read session count")
	}

	if !encoding.ReadUint64(data, index, &rts.EnvelopeUpKbps) {
		return errors.New("invalid data, could not read envelope up")
	}

	if !encoding.ReadUint64(data, index, &rts.EnvelopeDownKbps) {
		return errors.New("invalid data, could not read envelope down")
	}

	if !encoding.ReadUint64(data, index, &rts.OutboundPingTx) {
		return errors.New("invalid data, could not read outbound ping tx")
	}

	if !encoding.ReadUint64(data, index, &rts.RouteRequestRx) {
		return errors.New("invalid data, could not read route request rx")
	}
	if !encoding.ReadUint64(data, index, &rts.RouteRequestTx) {
		return errors.New("invalid data, could not read route request tx")
	}

	if !encoding.ReadUint64(data, index, &rts.RouteResponseRx) {
		return errors.New("invalid data, could not read route response rx")
	}
	if !encoding.ReadUint64(data, index, &rts.RouteResponseTx) {
		return errors.New("invalid data, could not read route response tx")
	}

	if !encoding.ReadUint64(data, index, &rts.ClientToServerRx) {
		return errors.New("invalid data, could not read client to server rx")
	}
	if !encoding.ReadUint64(data, index, &rts.ClientToServerTx) {
		return errors.New("invalid data, could not read client to server tx")
	}

	if !encoding.ReadUint64(data, index, &rts.ServerToClientRx) {
		return errors.New("invalid data, could not read server to client rx")
	}
	if !encoding.ReadUint64(data, index, &rts.ServerToClientTx) {
		return errors.New("invalid data, could not read server to client tx")
	}

	if !encoding.ReadUint64(data, index, &rts.InboundPingRx) {
		return errors.New("invalid data, could not read inbound ping rx")
	}
	if !encoding.ReadUint64(data, index, &rts.InboundPingTx) {
		return errors.New("invalid data, could not read inbound ping tx")
	}

	if !encoding.ReadUint64(data, index, &rts.PongRx) {
		return errors.New("invalid data, could not read pong rx")
	}

	if !encoding.ReadUint64(data, index, &rts.SessionPingRx) {
		return errors.New("invalid data, could not read session ping rx")
	}
	if !encoding.ReadUint64(data, index, &rts.SessionPingTx) {
		return errors.New("invalid data, could not read session ping tx")
	}

	if !encoding.ReadUint64(data, index, &rts.SessionPongRx) {
		return errors.New("invalid data, could not read session pong rx")
	}
	if !encoding.ReadUint64(data, index, &rts.SessionPongTx) {
		return errors.New("invalid data, could not read session pong tx")
	}

	if !encoding.ReadUint64(data, index, &rts.ContinueRequestRx) {
		return errors.New("invalid data, could not read continue request rx")
	}
	if !encoding.ReadUint64(data, index, &rts.ContinueRequestTx) {
		return errors.New("invalid data, could not read continue request tx")
	}

	if !encoding.ReadUint64(data, index, &rts.ContinueResponseRx) {
		return errors.New("invalid data, could not read continue response rx")
	}
	if !encoding.ReadUint64(data, index, &rts.ContinueResponseTx) {
		return errors.New("invalid data, could not read continue response tx")
	}

	if !encoding.ReadUint64(data, index, &rts.NearPingRx) {
		return errors.New("invalid data, could not read near ping rx")
	}
	if !encoding.ReadUint64(data, index, &rts.NearPingTx) {
		return errors.New("invalid data, could not read near ping tx")
	}

	if !encoding.ReadUint64(data, index, &rts.UnknownRx) {
		return errors.New("invalid data, could not read unknown rx")
	}

	return nil
}
