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
	magic          uint32
	version        uint32
	nonce          []byte
	address        string
	encryptedToken []byte
}

// UnmarshalBinary decodes binary data into a RelayInitPacket struct
func (r *RelayInitPacket) UnmarshalBinary(buf []byte) error {
	index := 0
	if !(encoding.ReadUint32(buf, &index, &r.magic) &&
		encoding.ReadUint32(buf, &index, &r.version) &&
		encoding.ReadBytes(buf, &index, &r.nonce, C.crypto_box_NONCEBYTES) &&
		encoding.ReadString(buf, &index, &r.address, MaxRelayAddressLength) &&
		encoding.ReadBytes(buf, &index, &r.encryptedToken, LengthOfRelayToken+C.crypto_box_MACBYTES)) {
		return errors.New("Invalid Packet")
	}

	return nil
}
