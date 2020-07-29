package routing

import (
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
)

const (
	// EncryptedRelayTokenSize ...
	EncryptedRelayTokenSize = crypto.KeySize + crypto.MACSize

	// HashKeyPrefixRelay ...
	HashKeyPrefixRelay = "RELAY-"

	// HashKeyAllRelays ...
	HashKeyAllRelays = "ALL_RELAYS"

	// Relax
	RelayTimeout = 60 * time.Second

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

	NICSpeedMbps        int32 `json:"nic_speed_mbps"`
	IncludedBandwidthGB int32 `json:"included_bandwidth_GB"`

	LastUpdateTime time.Time `json:"last_udpate_time"`

	State RelayState `json:"state"`

	ManagementAddr string `json:"management_addr"`
	SSHUser        string `json:"ssh_user"`
	SSHPort        int64  `json:"ssh_port"`

	TrafficStats RelayTrafficStats `json:"traffic_stats"`
	ClientStats  Stats             `json:"client_stats"`

	MaxSessions uint32 `json:"max_sessions"`

	CPUUsage float32 `json:"cpu_usage"`
	MemUsage float32 `json:"mem_usage"`

	UpdateKey   []byte `json:"update_key"`
	FirestoreID string `json:"firestore_id"`

	// MRC is the monthly recurring cost for the relay
	MRC Nibblin `json:"mrc"`
	// Overage is the charge/penalty if we exceed the bandwidth alloted for the relay
	Overage Nibblin       `json:"overage"`
	BWRule  BandWidthRule `json:"bw_rule"`
	//ContractTerm is the term in months
	ContractTerm int32 `json:"contract_term"`
	// StartDate is the date the contract term starts
	StartDate time.Time `json:"start_date"`
	// EndDate is the date the contract term ends
	EndDate time.Time   `json:"end_date"`
	Type    MachineType `json:"machine_type"`
}

func (r *Relay) EncodedPublicKey() string {
	return base64.StdEncoding.EncodeToString(r.PublicKey)
}

func (r *Relay) Size() uint64 {
	return uint64(8 + // ID
		4 + len(r.Name) + // Name
		4 + len(r.Addr.String()) + // Address
		len(r.PublicKey) + // Public Key
		4 + len(r.Seller.ID) + // Seller ID
		4 + len(r.Seller.Name) + // Seller Name
		8 + // Seller Ingress Price
		8 + // Seller Egress Price
		8 + // Datacenter ID
		4 + len(r.Datacenter.Name) + // Datacenter Name
		1 + // Datacenter Enabled
		8 + // Datacenter Location Latitude
		8 + // Datacenter Location Longitude
		8 + // NIC Speed Mbps
		8 + // Included Bandwidth GB
		8 + // Last Update Time
		4 + // Relay State
		4 + len(r.ManagementAddr) + // Management Address
		4 + len(r.SSHUser) + // SSH Username
		8 + // SSH Port
		8 + // Traffic Stats Session Count
		8 + // Traffic Stats Bytes Sent
		8 + // Traffic Stats Bytes Received
		4 + // BWRule
		4 + // Contract Term
		8 + // contract EndDate
		8 + // contract StartDate
		8 + // Overage (Nibblin, uint64)
		8 + // MRC (Nibblin, uint64)
		4 + // MachineType
		4 + // Max Sessions
		4 + // CPU usage
		4, // Mem usage
	)
}

type RelayCacheEntry struct {
	ID             uint64
	Name           string
	Addr           net.UDPAddr
	PublicKey      []byte
	Seller         Seller
	Datacenter     Datacenter
	LastUpdateTime time.Time
	TrafficStats   RelayTrafficStats
	MaxSessions    uint32
	CPUUsage       float32
	MemUsage       float32
	Version        string
}

func (e *RelayCacheEntry) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, e)
}

func (e RelayCacheEntry) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(e)
}

// Key returns the key used for Redis
func (r *RelayCacheEntry) Key() string {
	return HashKeyPrefixRelay + strconv.FormatUint(r.ID, 10)
}

// RelayTrafficStats describes the measured relay traffic statistics reported from the relay
type RelayTrafficStats struct {
	SessionCount  uint64
	BytesSent     uint64
	BytesReceived uint64
}

type Stats struct {
	RTT        float64 `json:"rtt"`
	Jitter     float64 `json:"jitter"`
	PacketLoss float64 `json:"packet_loss"`
}

func (s Stats) String() string {
	return fmt.Sprintf("RTT(%f) J(%f) PL(%f)", s.RTT, s.Jitter, s.PacketLoss)
}

type RelayPingData struct {
	ID      uint64 `json:"relay_id"`
	Address string `json:"relay_address"`
}

type LegacyPingToken struct {
	Timeout uint64
	RelayID uint64
	HMac    [32]byte
}

func (l LegacyPingToken) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 48) // ping token binary is 57 bytes
	index := 0
	encoding.WriteUint64(data, &index, l.Timeout)
	encoding.WriteUint64(data, &index, l.RelayID)
	encoding.WriteBytes(data, &index, l.HMac[:], len(l.HMac))
	return data, nil
}

type LegacyPingData struct {
	RelayPingData
	PingToken string `json:"ping_info"` // base64 of LegacyPingToken binary form
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
