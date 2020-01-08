package transport

import (
	"errors"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/rw"
)

// RelayUpdatePacket is the struct wrapping a update packet
type RelayUpdatePacket struct {
	version   uint32
	address   string
	token     []byte
	numRelays uint32

	pingStats []core.RelayStatsPing
}

// UnmarshalBinary decodes the binary data into a RelayUpdatePacket struct
func (r *RelayUpdatePacket) UnmarshalBinary(buff []byte) error {
	index := 0
	if !(rw.ReadUint32(buff, &index, &r.version) &&
		rw.ReadString(buff, &index, &r.address, MaxRelayAddressLength) &&
		rw.ReadBytes(buff, &index, &r.token, RelayTokenBytes) &&
		rw.ReadUint32(buff, &index, &r.numRelays)) {
		return errors.New("Invalid Packet")
	}

	for i := 0; i < int(r.numRelays); i++ {
		var id uint64

		pingStats := core.RelayStatsPing{}

		if !(rw.ReadUint64(buff, &index, &id) &&
			rw.ReadFloat32(buff, &index, &pingStats.RTT) &&
			rw.ReadFloat32(buff, &index, &pingStats.Jitter) &&
			rw.ReadFloat32(buff, &index, &pingStats.PacketLoss)) {
			return errors.New("Invalid Packet")
		}

		pingStats.RelayId = core.RelayId(id)

		r.pingStats = append(r.pingStats, pingStats)
	}

	return nil
}
