package transport

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"errors"

	"github.com/networknext/backend/rw"
)

const gMaxRelayIDLength = 256
const gMaxRelayAddressLength = 256

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
	if !(rw.ReadUint32(buf, &index, &r.magic) &&
		rw.ReadUint32(buf, &index, &r.version) &&
		rw.ReadBytes(buf, &index, &r.nonce, C.crypto_box_NONCEBYTES) &&
		rw.ReadString(buf, &index, &r.address, gMaxRelayAddressLength) &&
		rw.ReadBytes(buf, &index, &r.encryptedToken, gRelayTokenBytes+C.crypto_box_MACBYTES)) {
		return errors.New("Invalid Packet")
	}

	return nil
}
