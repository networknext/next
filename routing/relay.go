package routing

import (
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"

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

	/* Duplicated in package: transport */
	// MaxRelayAddressLength ...
	MaxRelayAddressLength = 256

	RelayStateOffline      = 0
	RelayStateInitalized   = 1
	RelayStateOnline       = 2
	RelayStateShuttingDown = 3
)

// Relay ...
type Relay struct {
	ID   uint64
	Name string

	Addr      net.UDPAddr
	PublicKey []byte

	Seller     Seller
	Datacenter Datacenter

	Latitude  float64
	Longitude float64

	LastUpdateTime uint64

	State uint32
}

func (r *Relay) Size() uint64 {
	return uint64(8 + 4 + len(r.Name) + 4 + len(r.Addr.String()) + len(r.PublicKey) + 4 + len(r.Seller.ID) + 4 + len(r.Seller.Name) + 8 + 8 + 8 + 4 + len(r.Datacenter.Name) + 1 + 8 + 8 + 8 + 4)
}

// UnmarshalBinary ...
// TODO add other fields to this
func (r *Relay) UnmarshalBinary(data []byte) error {
	index := 0

	var addr string
	if !encoding.ReadUint64(data, &index, &r.ID) {
		return errors.New("failed to unmarshal relay ID")
	}

	// TODO define an actual limit on this
	if !encoding.ReadString(data, &index, &r.Name, math.MaxInt32) {
		return errors.New("failed to unmarshal relay name")
	}

	if !encoding.ReadString(data, &index, &addr, MaxRelayAddressLength) {
		return errors.New("failed to unmarshal relay address")
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

	if !encoding.ReadFloat64(data, &index, &r.Latitude) {
		return errors.New("failed to unmarshal relay latitude")
	}

	if !encoding.ReadFloat64(data, &index, &r.Longitude) {
		return errors.New("failed to unmarshal relay longitude")
	}

	if !encoding.ReadUint64(data, &index, &r.LastUpdateTime) {
		return errors.New("failed to unmarshal relay last update time")
	}

	if !encoding.ReadUint32(data, &index, &r.State) {
		return errors.New("failed to unmarshal relay state")
	}

	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Addr = *udp
	} else {
		return errors.New("invalid relay address")
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
	encoding.WriteFloat64(data, &index, r.Latitude)
	encoding.WriteFloat64(data, &index, r.Longitude)
	encoding.WriteUint64(data, &index, r.LastUpdateTime)
	encoding.WriteUint32(data, &index, r.State)

	return data, err
}

// Key returns the key used for Redis
func (r *Relay) Key() string {
	return HashKeyPrefixRelay + strconv.FormatUint(r.ID, 10)
}

type Stats struct {
	RTT        float64
	Jitter     float64
	PacketLoss float64
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
