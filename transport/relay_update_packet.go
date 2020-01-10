package transport

import (
	"errors"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/encoding"
)

// RelayUpdatePacket is the struct wrapping a update packet
type RelayUpdatePacket struct {
	Version   uint32
	Address   string
	Token     []byte
	NumRelays uint32

	PingStats []core.RelayStatsPing
}

// UnmarshalBinary decodes the binary data into a RelayUpdatePacket struct
func (r *RelayUpdatePacket) UnmarshalBinary(buff []byte) error {
	index := 0
	if !(encoding.ReadUint32(buff, &index, &r.Version) &&
		encoding.ReadString(buff, &index, &r.Address, MaxRelayAddressLength) &&
		encoding.ReadBytes(buff, &index, &r.Token, LengthOfRelayToken) &&
		encoding.ReadUint32(buff, &index, &r.NumRelays)) {
		return errors.New("Invalid Packet")
	}

	for i := 0; i < int(r.NumRelays); i++ {
		var id uint64

		pingStats := core.RelayStatsPing{}

		if !(encoding.ReadUint64(buff, &index, &id) &&
			encoding.ReadFloat32(buff, &index, &pingStats.RTT) &&
			encoding.ReadFloat32(buff, &index, &pingStats.Jitter) &&
			encoding.ReadFloat32(buff, &index, &pingStats.PacketLoss)) {
			return errors.New("Invalid Packet")
		}

		pingStats.RelayID = id

		r.PingStats = append(r.PingStats, pingStats)
	}

	return nil
}
