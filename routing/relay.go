package routing

import (
	"errors"
	"hash/fnv"
	"math"
	"net"
	"strconv"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
)

const (
	EncryptedTokenSize = crypto.KeySize + crypto.MACSize

	RedisHashKeyStart = "RELAY-"
)

type Relay struct {
	ID   uint64
	Name string

	Addr      net.UDPAddr
	PublicKey []byte

	Datacenter     uint64
	DatacenterName string

	Latitude  float64
	Longitude float64

	RTT        float64
	Jitter     float64
	PacketLoss float64

	LastUpdateTime uint64

	_CachedKey string
}

func NewRelay() *Relay {
	relay := new(Relay)
	relay.PublicKey = make([]byte, crypto.KeySize)
	return relay
}

// TODO add other things to this
func (r *Relay) UnmarshalBinary(data []byte) error {
	index := 0

	var addr string
	if !(encoding.ReadUint64(data, &index, &r.ID) &&
		encoding.ReadString(data, &index, &r.Name, math.MaxInt32) && // TODO define a actual limit on this
		encoding.ReadString(data, &index, &addr, MaxRelayAddressLength) &&
		encoding.ReadUint64(data, &index, &r.Datacenter) &&
		encoding.ReadString(data, &index, &r.DatacenterName, math.MaxInt32) &&
		encoding.ReadBytes(data, &index, &r.PublicKey, crypto.KeySize) &&
		encoding.ReadUint64(data, &index, &r.LastUpdateTime)) {
		return errors.New("Invalid RelayData")
	}
	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Addr = *udp
	} else {
		return errors.New("Invalid relay address")
	}
	return nil
}

// TODO add other things to this
func (r Relay) MarshalBinary() (data []byte, err error) {
	strAddr := r.Addr.String()
	length := 8 + 4 + len(r.Name) + 4 + len(strAddr) + 8 + 4 + len(r.DatacenterName) + len(r.PublicKey) + 8

	data = make([]byte, length)

	index := 0
	encoding.WriteUint64(data, &index, r.ID)
	encoding.WriteString(data, &index, r.Name, uint32(len(r.Name)))
	encoding.WriteString(data, &index, strAddr, uint32(len(strAddr)))
	encoding.WriteUint64(data, &index, r.Datacenter)
	encoding.WriteString(data, &index, r.DatacenterName, uint32(len(r.DatacenterName)))
	encoding.WriteBytes(data, &index, r.PublicKey, crypto.KeySize)
	encoding.WriteUint64(data, &index, r.LastUpdateTime)

	return data, err
}

func (r Relay) Key() string {
	if len(r._CachedKey) == 0 {
		r._CachedKey = RedisHashKeyStart + strconv.FormatUint(r.ID, 10)
	}

	return r._CachedKey
}

type RelayUpdate struct {
	ID             uint64
	Name           string
	Address        net.UDPAddr
	Datacenter     uint64
	DatacenterName string
	PublicKey      []byte
	Shutdown       bool
}

// GetRelayID hashes the name of the relay and returns the result. Typically name is the address of the relay
func GetRelayID(addr string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(addr))
	return hash.Sum64()
}
