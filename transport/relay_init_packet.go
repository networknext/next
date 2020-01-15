package transport

import (
	"errors"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
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
		encoding.ReadBytes(buf, &index, &r.EncryptedToken, crypto.PublicKeySize+crypto.MACSize)) {
		return errors.New("invalid packet")
	}

	return nil
}
