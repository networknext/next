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
	// EncryptedTokenSize ...
	EncryptedTokenSize = crypto.KeySize + crypto.MACSize

	// HashKeyPrefixRelay ...
	HashKeyPrefixRelay = "RELAY-"

	// HashKeyAllRelays ...
	HashKeyAllRelays = "ALL_RELAYS"

	// How frequently we need to recieve updates from relays to keep them in redis
	RelayTimeout = 10 * time.Second
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

	Stats Stats

	LastUpdateTime uint64
}

// NewRelay ...
func NewRelay() Relay {
	relay := Relay{
		PublicKey: make([]byte, crypto.KeySize),
	}

	return relay
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

	if !encoding.ReadBytes(data, &index, &r.PublicKey, crypto.KeySize) {
		return errors.New("failed to unmarshal relay public key")
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
	strAddr := r.Addr.String()
	length := 8 + 4 + len(r.Name) + 4 + len(strAddr) + 4 + len(r.Seller.Name) + 8 + 8 + 8 + 4 + len(r.Datacenter.Name) + len(r.PublicKey) + 8 + 8 + 8

	data = make([]byte, length)

	index := 0
	encoding.WriteUint64(data, &index, r.ID)
	encoding.WriteString(data, &index, r.Name, uint32(len(r.Name)))
	encoding.WriteString(data, &index, strAddr, uint32(len(strAddr)))
	encoding.WriteString(data, &index, r.Seller.Name, uint32(len(r.Seller.Name)))
	encoding.WriteUint64(data, &index, r.Seller.IngressPriceCents)
	encoding.WriteUint64(data, &index, r.Seller.EgressPriceCents)
	encoding.WriteUint64(data, &index, r.Datacenter.ID)
	encoding.WriteString(data, &index, r.Datacenter.Name, uint32(len(r.Datacenter.Name)))
	encoding.WriteBytes(data, &index, r.PublicKey, crypto.KeySize)
	encoding.WriteFloat64(data, &index, r.Latitude)
	encoding.WriteFloat64(data, &index, r.Longitude)
	encoding.WriteUint64(data, &index, r.LastUpdateTime)

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

// RelayUpdate ...
type RelayUpdate struct {
	ID             uint64
	Name           string
	Address        net.UDPAddr
	Datacenter     uint64
	DatacenterName string
	PublicKey      []byte
	Shutdown       bool
}

type RelayPingData struct {
	ID      uint64
	Address string
}
