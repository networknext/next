package transport

import (
	"errors"
	"math"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
)

// RelayInitPacket is the struct that describes the packets comming into the relay_init endpoint
type RelayInitPacket struct {
	Magic          uint32
	Version        uint32
	Nonce          []byte
	Address        string
	EncryptedToken []byte
}

// UnmarshalBinary decodes binary data into a RelayInitPacket struct
func (r *RelayInitPacket) UnmarshalBinary(buf []byte) error {
	index := 0
	if !(encoding.ReadUint32(buf, &index, &r.Magic) &&
		encoding.ReadUint32(buf, &index, &r.Version) &&
		encoding.ReadBytes(buf, &index, &r.Nonce, crypto.NonceSize) &&
		encoding.ReadString(buf, &index, &r.Address, MaxRelayAddressLength) &&
		encoding.ReadBytes(buf, &index, &r.EncryptedToken, routing.EncryptedTokenSize)) {
		return errors.New("invalid packet")
	}

	return nil
}

func (r RelayInitPacket) MarshalBinary() ([]byte, error) {
	data := make([]byte, 4+4+crypto.NonceSize+4+len(r.Address)+routing.EncryptedTokenSize)
	index := 0
	encoding.WriteUint32(data, &index, r.Magic)
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteBytes(data, &index, r.Nonce, crypto.NonceSize)
	encoding.WriteString(data, &index, r.Address, uint32(len(r.Address)))
	encoding.WriteBytes(data, &index, r.EncryptedToken, routing.EncryptedTokenSize)

	return data, nil
}

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
		encoding.ReadBytes(buff, &index, &r.Token, crypto.KeySize) &&
		encoding.ReadUint32(buff, &index, &r.NumRelays)) {
		return errors.New("Invalid Packet")
	}

	r.PingStats = make([]core.RelayStatsPing, r.NumRelays)
	for i := 0; i < int(r.NumRelays); i++ {
		stats := &r.PingStats[i]

		if !(encoding.ReadUint64(buff, &index, &stats.RelayID) &&
			encoding.ReadFloat32(buff, &index, &stats.RTT) &&
			encoding.ReadFloat32(buff, &index, &stats.Jitter) &&
			encoding.ReadFloat32(buff, &index, &stats.PacketLoss)) {
			return errors.New("Invalid Packet")
		}
	}

	return nil
}

// MarshalBinary ...
func (r RelayUpdatePacket) MarshalBinary() ([]byte, error) {
	data := make([]byte, 4+4+len(r.Address)+routing.EncryptedTokenSize+4+20*len(r.PingStats))

	index := 0
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteString(data, &index, r.Address, math.MaxInt32)
	encoding.WriteBytes(data, &index, r.Token, crypto.KeySize)
	encoding.WriteUint32(data, &index, r.NumRelays)

	for i := 0; i < int(r.NumRelays); i++ {
		stats := &r.PingStats[i]

		encoding.WriteUint64(data, &index, stats.RelayID)
		encoding.WriteUint32(data, &index, math.Float32bits(stats.RTT))
		encoding.WriteUint32(data, &index, math.Float32bits(stats.Jitter))
		encoding.WriteUint32(data, &index, math.Float32bits(stats.PacketLoss))
	}

	return data, nil
}
