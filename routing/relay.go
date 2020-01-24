package routing

import (
	"errors"
	"math"
	"net"
	"strconv"

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
)

// Relay ...
type Relay struct {
	ID   uint64
	Name string

	Addr      net.UDPAddr
	PublicKey []byte

	Datacenter     uint64
	DatacenterName string

	Latitude  float64
	Longitude float64

	Stats Stats

	LastUpdateTime uint64

	cachedKey string
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
	if !(encoding.ReadUint64(data, &index, &r.ID) &&
		encoding.ReadString(data, &index, &r.Name, math.MaxInt32) && // TODO define an actual limit on this
		encoding.ReadString(data, &index, &addr, MaxRelayAddressLength) &&
		encoding.ReadUint64(data, &index, &r.Datacenter) &&
		encoding.ReadString(data, &index, &r.DatacenterName, math.MaxInt32) &&
		encoding.ReadBytes(data, &index, &r.PublicKey, crypto.KeySize) &&
		encoding.ReadFloat64(data, &index, &r.Latitude) &&
		encoding.ReadFloat64(data, &index, &r.Longitude) &&
		encoding.ReadUint64(data, &index, &r.LastUpdateTime)) {
		return errors.New("Invalid Relay")
	}
	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Addr = *udp
	} else {
		return errors.New("Invalid relay address")
	}
	return nil
}

// MarshalBinary ...
// TODO add other fields to this
func (r Relay) MarshalBinary() (data []byte, err error) {
	strAddr := r.Addr.String()
	length := 8 + 4 + len(r.Name) + 4 + len(strAddr) + 8 + 4 + len(r.DatacenterName) + len(r.PublicKey) + 8 + 8 + 8

	data = make([]byte, length)

	index := 0
	encoding.WriteUint64(data, &index, r.ID)
	encoding.WriteString(data, &index, r.Name, uint32(len(r.Name)))
	encoding.WriteString(data, &index, strAddr, uint32(len(strAddr)))
	encoding.WriteUint64(data, &index, r.Datacenter)
	encoding.WriteString(data, &index, r.DatacenterName, uint32(len(r.DatacenterName)))
	encoding.WriteBytes(data, &index, r.PublicKey, crypto.KeySize)
	encoding.WriteFloat64(data, &index, r.Latitude)
	encoding.WriteFloat64(data, &index, r.Longitude)
	encoding.WriteUint64(data, &index, r.LastUpdateTime)

	return data, err
}

// Key returns the key used for Redis
func (r *Relay) Key() string {
	if len(r.cachedKey) == 0 {
		r.cachedKey = HashKeyPrefixRelay + strconv.FormatUint(r.ID, 10)
	}

	return r.cachedKey
}

type Stats struct {
	RTT        float64
	Jitter     float64
	PacketLoss float64
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
