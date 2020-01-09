package transport

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"errors"

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
		encoding.ReadBytes(buf, &index, &r.Nonce, C.crypto_box_NONCEBYTES) &&
		encoding.ReadString(buf, &index, &r.Address, MaxRelayAddressLength) &&
		encoding.ReadBytes(buf, &index, &r.EncryptedToken, LengthOfRelayToken+C.crypto_box_MACBYTES)) {
		return errors.New("Invalid Packet")
	}

	return nil
}
