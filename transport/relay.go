package transport

import (
	"errors"
	"math"

	"github.com/networknext/backend/encoding"
)

// RelayData is the data to be stored on the redis server
type RelayData struct {
	ID             uint64
	Name           string
	Address        string
	Datacenter     uint64
	DatacenterName string
	PublicKey      []byte
	LastUpdateTime uint64
}

func (r *RelayData) UnmarshalBinary(data []byte) error {
	index := 0
	if !(encoding.ReadUint64(data, &index, &r.ID) &&
		encoding.ReadString(data, &index, &r.Name, math.MaxInt32) && // TODO define a actual limit on this
		encoding.ReadString(data, &index, &r.Address, math.MaxInt32) && // and this
		encoding.ReadUint64(data, &index, &r.Datacenter) &&
		encoding.ReadString(data, &index, &r.DatacenterName, math.MaxInt32) &&
		encoding.ReadBytes(data, &index, &r.PublicKey, LengthOfRelayToken) &&
		encoding.ReadUint64(data, &index, &r.LastUpdateTime)) {
		return errors.New("Invalid RelayData")
	}

	return nil
}

func (r *RelayData) MarshalBinary(data []byte) error {
	length := 8 + 4 + len(r.Name) + 4 + len(r.Address) + 8 + 4 + len(r.DatacenterName) + len(r.PublicKey) + 8
	if len(data) < length {
		return errors.New("Length of buffer is too small")
	}

	index := 0
	encoding.WriteUint64(data, &index, r.ID)
	encoding.WriteString(data, &index, r.Name, uint32(len(r.Name)))
	encoding.WriteString(data, &index, r.Address, uint32(len(r.Address)))
	encoding.WriteUint64(data, &index, r.Datacenter)
	encoding.WriteString(data, &index, r.DatacenterName, uint32(len(r.DatacenterName)))
	encoding.WriteBytes(data, &index, r.PublicKey, LengthOfRelayToken)
	encoding.WriteUint64(data, &index, r.LastUpdateTime)

	return nil
}
