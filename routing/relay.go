package routing

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
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

	// How frequently we need to recieve updates from relays to keep them in redis
	// 10 seconds + a 1 second grace period
	RelayTimeout = 11 * time.Second

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

type Relay struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`

	Addr      net.UDPAddr `json:"addr"`
	PublicKey []byte      `json:"public_key"`

	Seller     Seller     `json:"seller"`
	Datacenter Datacenter `json:"datacenter"`

	NICSpeedMbps        uint64 `json:"nic_speed_mbps"`
	IncludedBandwidthGB uint64 `json:"included_bandwidth_GB"`

	LastUpdateTime time.Time `json:"last_udpate_time"`

	State RelayState `json:"state"`

	ManagementAddr string `json:"management_addr"`
	SSHUser        string `json:"ssh_user"`
	SSHPort        int64  `json:"ssh_port"`

	TrafficStats RelayTrafficStats `json:"traffic_stats"`
	ClientStats  Stats             `json:"client_stats"`

	MaxSessions uint32 `json:"max_sessions"`

	UpdateKey   []byte `json:"update_key"`
	FirestoreID string `json:"firestore_id"`

	// MRC is the monthly recurring cost for the relay
	MRC Nibblin `json:"mrc"`
	// Overage is the charge/penalty if we exceed the bandwidth alloted for the relay
	Overage Nibblin `json:"overage"`
	// BWRule: Flat / Burst / Pool: Relates to bandwidth
	//	Flat : can not go over allocated amount
	// 	Burst: can go over amount
	//	Pool : supplier gives X amount of bandwidth for all relays in the pool
	BWRule string `json:"bw_rule"`
	//ContractTerm is the term in months
	ContractTerm uint32 `json:"contract_term"`
	// StartDate is the date the contract term starts
	StartDate time.Time `json:"start_date"`
	// EndDate is the date the contract term ends
	EndDate time.Time `json:"end_date"`
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
		4 + len(r.BWRule) + // bandwidth rule
		4 + // Contract Term
		4 + len(r.StartDate.String()) + // contract start date converted to json date string
		4 + len(r.EndDate.String()) + // contract start date converted to json date string
		8 + // Overage (Nibblin, uint64)
		8 + // MRC (Nibblin, uint64)
		4, // Max Sessions
	)
}

// UnmarshalBinary ...
// TODO add other fields to this
func (r *Relay) UnmarshalBinary(data []byte) error {
	index := 0

	if !encoding.ReadUint64(data, &index, &r.ID) {
		return errors.New("failed to unmarshal relay ID")
	}

	// TODO define an actual limit on this
	if !encoding.ReadString(data, &index, &r.Name, math.MaxInt32) {
		return errors.New("failed to unmarshal relay name")
	}

	var addr string
	if !encoding.ReadString(data, &index, &addr, MaxRelayAddressLength) {
		return errors.New("failed to unmarshal relay address")
	}

	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Addr = *udp
	} else {
		return errors.New("invalid relay address")
	}

	if !encoding.ReadBytes(data, &index, &r.PublicKey, crypto.KeySize) {
		return errors.New("failed to unmarshal relay public key")
	}

	if !encoding.ReadString(data, &index, &r.Seller.ID, math.MaxInt32) {
		return errors.New("failed to unmarshal relay seller ID")
	}

	if !encoding.ReadString(data, &index, &r.Seller.Name, math.MaxInt32) {
		return errors.New("failed to unmarshal relay seller name")
	}

	if !encoding.ReadUint64(data, &index, &r.Seller.IngressPriceCents) {
		return errors.New("failed to unmarshal relay seller ingress price")
	}

	if !encoding.ReadUint64(data, &index, &r.Seller.EgressPriceCents) {
		return errors.New("failed to unmarshal relay seller egress price")
	}

	if !encoding.ReadUint64(data, &index, &r.Datacenter.ID) {
		return errors.New("failed to unmarshal relay datacenter id")
	}

	if !encoding.ReadString(data, &index, &r.Datacenter.Name, math.MaxInt32) {
		return errors.New("failed to unmarshal relay datacenter name")
	}

	if !encoding.ReadBool(data, &index, &r.Datacenter.Enabled) {
		return errors.New("failed to unmarshal relay datacenter enabled")
	}

	if !encoding.ReadFloat64(data, &index, &r.Datacenter.Location.Latitude) {
		return errors.New("failed to unmarshal relay latitude")
	}

	if !encoding.ReadFloat64(data, &index, &r.Datacenter.Location.Longitude) {
		return errors.New("failed to unmarshal relay longitude")
	}

	if !encoding.ReadUint64(data, &index, &r.NICSpeedMbps) {
		return errors.New("failed to unmarshal relay NIC speed")
	}

	if !encoding.ReadUint64(data, &index, &r.IncludedBandwidthGB) {
		return errors.New("failed to unmarshal relay included bandwidth")
	}

	var lastUpdateTime uint64
	if !encoding.ReadUint64(data, &index, &lastUpdateTime) {
		return errors.New("failed to unmarshal relay last update time")
	}
	r.LastUpdateTime = time.Unix(0, int64(lastUpdateTime))

	var state uint32
	if !encoding.ReadUint32(data, &index, &state) {
		return errors.New("failed to unmarshal relay state")
	}
	r.State = RelayState(state)

	if !encoding.ReadString(data, &index, &r.ManagementAddr, math.MaxInt32) {
		return errors.New("failed to unmarshal relay management address")
	}

	if !encoding.ReadString(data, &index, &r.SSHUser, math.MaxInt32) {
		return errors.New("failed to unmarshal relay SSH username")
	}

	var sshPort uint64
	if !encoding.ReadUint64(data, &index, &sshPort) {
		return errors.New("failed to unmarshal relay SSH port")
	}
	r.SSHPort = int64(sshPort)

	if !encoding.ReadUint64(data, &index, &r.TrafficStats.SessionCount) {
		return errors.New("failed to unmarshal relay session count")
	}

	if !encoding.ReadUint64(data, &index, &r.TrafficStats.BytesSent) {
		return errors.New("failed to unmarshal relay bytes sent")
	}

	if !encoding.ReadUint64(data, &index, &r.TrafficStats.BytesReceived) {
		return errors.New("failed to unmarshal relay bytes received")
	}

	if !encoding.ReadUint32(data, &index, &r.MaxSessions) {
		return errors.New("failed to unmarshal relay max sessions")
	}

	return nil
}

// MarshalBinary ...
// TODO add other fields to this
func (r Relay) MarshalBinary() (data []byte, err error) {
	data = make([]byte, r.Size())
	strAddr := r.Addr.String()
	index := 0

	encoding.WriteUint64(data, &index, r.ID)
	encoding.WriteString(data, &index, r.Name, uint32(len(r.Name)))
	encoding.WriteString(data, &index, strAddr, uint32(len(strAddr)))
	encoding.WriteBytes(data, &index, r.PublicKey, crypto.KeySize)
	encoding.WriteString(data, &index, r.Seller.ID, uint32(len(r.Seller.ID)))
	encoding.WriteString(data, &index, r.Seller.Name, uint32(len(r.Seller.Name)))
	encoding.WriteUint64(data, &index, r.Seller.IngressPriceCents)
	encoding.WriteUint64(data, &index, r.Seller.EgressPriceCents)
	encoding.WriteUint64(data, &index, r.Datacenter.ID)
	encoding.WriteString(data, &index, r.Datacenter.Name, uint32(len(r.Datacenter.Name)))
	encoding.WriteBool(data, &index, r.Datacenter.Enabled)
	encoding.WriteFloat64(data, &index, r.Datacenter.Location.Latitude)
	encoding.WriteFloat64(data, &index, r.Datacenter.Location.Longitude)
	encoding.WriteUint64(data, &index, r.NICSpeedMbps)
	encoding.WriteUint64(data, &index, r.IncludedBandwidthGB)
	encoding.WriteUint64(data, &index, uint64(r.LastUpdateTime.UnixNano()))
	encoding.WriteUint32(data, &index, uint32(r.State))
	encoding.WriteString(data, &index, r.ManagementAddr, uint32(len(r.ManagementAddr)))
	encoding.WriteString(data, &index, r.SSHUser, uint32(len(r.SSHUser)))
	encoding.WriteUint64(data, &index, uint64(r.SSHPort))
	encoding.WriteUint64(data, &index, r.TrafficStats.SessionCount)
	encoding.WriteUint64(data, &index, r.TrafficStats.BytesSent)
	encoding.WriteUint64(data, &index, r.TrafficStats.BytesReceived)
	encoding.WriteUint32(data, &index, r.MaxSessions)

	return data, err
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
