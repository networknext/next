package routing

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
)

const (
	// EncryptedRelayTokenSize ...
	EncryptedRelayTokenSize = crypto.KeySize + crypto.MACSize

	RelayTimeout = 30 * time.Second

	// MaxRelayAddressLength ...
	MaxRelayAddressLength = 256
)

type RelayState uint32

func (state RelayState) String() string {
	switch state {
	case RelayStateEnabled:
		return "enabled"
	case RelayStateMaintenance:
		return "maintenance"
	case RelayStateDisabled:
		return "disabled"
	case RelayStateQuarantine:
		return "quarantine"
	case RelayStateDecommissioned:
		return "decommissioned"
	case RelayStateOffline:
		return "offline"
	default:
		return "unknown"
	}
}

// ParseRelayState returns a relay state type given the state in string form
func ParseRelayState(str string) (RelayState, error) {
	switch str {
	case "enabled":
		return RelayStateEnabled, nil
	case "maintenance":
		return RelayStateMaintenance, nil
	case "disabled":
		return RelayStateDisabled, nil
	case "quarantine":
		return RelayStateQuarantine, nil
	case "decommissioned":
		return RelayStateDecommissioned, nil
	case "offline":
		return RelayStateOffline, nil
	default:
		return RelayStateEnabled, fmt.Errorf("invalid relay state '%s'", str)
	}
}

const (
	// RelayStateEnabled if running and communicating with backend
	RelayStateEnabled RelayState = 0
	// RelayStateMaintenance System does a clean shutdown of the relay process
	RelayStateMaintenance RelayState = 1
	// RelayStateDisabled System does a clean shutdown and is shown as disabled as a temporary state before decommissioning in case it needs to be re-enabled
	RelayStateDisabled RelayState = 2
	// RelayStateQuarantine Backend has detected an unexpected disruption in service from this relay and has removed it from getting packets until it is manually added back into the system
	RelayStateQuarantine RelayState = 3
	// RelayStateDecommissioned System is removed from the UI and lists. Reusable fields like IP and name are cleared in firestore. Most data is retained for historical purposes
	RelayStateDecommissioned RelayState = 4
	// RelayStateOffline Relay has stopped communicating with the backend for some unknown reason
	RelayStateOffline RelayState = 5
)

// BandWidthRule Flat / Burst / Pool: Relates to bandwidth
type BandWidthRule uint32

const (
	BWRuleNone  BandWidthRule = iota
	BWRuleFlat  BandWidthRule = iota // can not go over allocated amount
	BWRuleBurst BandWidthRule = iota // can go over amount
	BWRulePool  BandWidthRule = iota // supplier gives X amount of bandwidth for all relays in the pool
)

// MachineType is the type of server the relay is running on
type MachineType uint32

const (
	NoneSpecified  MachineType = iota
	BareMetal      MachineType = iota
	VirtualMachine MachineType = iota
)

type Relay struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`

	Addr      net.UDPAddr `json:"addr"`
	PublicKey []byte      `json:"public_key"`

	Seller     Seller     `json:"seller"`
	Datacenter Datacenter `json:"datacenter"`

	NICSpeedMbps        int32 `json:"nicSpeedMbps"`
	IncludedBandwidthGB int32 `json:"includedBandwidthGB"`

	LastUpdateTime time.Time `json:"last_udpate_time"`

	State RelayState `json:"state"`

	ManagementAddr string `json:"management_addr"`
	SSHUser        string `json:"ssh_user"`
	SSHPort        int64  `json:"ssh_port"`

	TrafficStats TrafficStats `json:"traffic_stats"`

	MaxSessions uint32 `json:"max_sessions"`

	CPUUsage float32 `json:"cpu_usage"`
	MemUsage float32 `json:"mem_usage"`

	UpdateKey   []byte `json:"update_key"`
	FirestoreID string `json:"firestore_id"`

	// MRC is the monthly recurring cost for the relay
	MRC Nibblin `json:"monthlyRecurringChargeNibblins"`
	// Overage is the charge/penalty if we exceed the bandwidth alloted for the relay
	Overage Nibblin       `json:"overage"`
	BWRule  BandWidthRule `json:"bandwidthRule"`
	//ContractTerm is the term in months
	ContractTerm int32 `json:"contractTerm"`
	// StartDate is the date the contract term starts
	StartDate time.Time `json:"startDate"`
	// EndDate is the date the contract term ends
	EndDate time.Time   `json:"endDate"`
	Type    MachineType `json:"machineType"`

	// Useful in data science analysis
	SignedID int64 `json:"signed_id"`
}

func (r *Relay) EncodedPublicKey() string {
	return base64.StdEncoding.EncodeToString(r.PublicKey)
}

// TrafficStats describes the measured relay traffic statistics reported from the relay
type TrafficStats struct {
	SessionCount uint64
	EnvelopeUp   uint64
	EnvelopeDown uint64

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

		EnvelopeUp:   rts.EnvelopeUp + other.EnvelopeUp,
		EnvelopeDown: rts.EnvelopeDown + other.EnvelopeDown,

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
	encoding.WriteUint64(data, index, rts.EnvelopeUp)
	encoding.WriteUint64(data, index, rts.EnvelopeDown)
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

	if !encoding.ReadUint64(data, index, &rts.EnvelopeUp) {
		return errors.New("invalid data, could not read envelope up")
	}

	if !encoding.ReadUint64(data, index, &rts.EnvelopeDown) {
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

type PeakTrafficStats struct {
	SessionCount uint64
	EnvelopeUp   uint64
	EnvelopeDown uint64
}

func (pts *PeakTrafficStats) MaxValues(other PeakTrafficStats) PeakTrafficStats {
	if pts.SessionCount > other.SessionCount {
		other.SessionCount = pts.SessionCount
	}

	if pts.EnvelopeUp > other.EnvelopeDown {
		other.EnvelopeUp = pts.EnvelopeUp
	}

	if pts.EnvelopeDown > other.EnvelopeDown {
		other.EnvelopeDown = pts.EnvelopeDown
	}

	return other
}

type Stats struct {
	RTT        float64 `json:"rtt"`
	Jitter     float64 `json:"jitter"`
	PacketLoss float64 `json:"packet_loss"`
}

func (s Stats) String() string {
	return fmt.Sprintf("RTT(%f) J(%f) PL(%f)", s.RTT, s.Jitter, s.PacketLoss)
}

func (s Stats) RedisString() string {
	return fmt.Sprintf("%.2f|%.2f|%.2f", s.RTT, s.Jitter, s.PacketLoss)
}

func (s *Stats) ParseRedisString(values []string) error {
	var index int
	var err error

	if s.RTT, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[Stats] failed to read RTT from redis data: %v", err)
	}
	index++

	if s.Jitter, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[Stats] failed to read jitter from redis data: %v", err)
	}
	index++

	if s.PacketLoss, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[Stats] failed to read packet loss from redis data: %v", err)
	}
	index++

	return nil
}

type RelayPingData struct {
	ID      uint64 `json:"relay_id"`
	Address string `json:"relay_address"`
}

func RelayAddrs(relays []Relay) string {
	var b strings.Builder
	for _, relay := range relays {
		b.WriteString("{")
		b.WriteString(relay.Addr.String())
		b.WriteString("}")
	}
	return b.String()
}
