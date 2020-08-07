package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"hash/fnv"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/poly1305"

	"github.com/spaolacci/murmur3"
)

const (
	MACSize        = poly1305.TagSize
	NonceSize      = chacha20poly1305.NonceSizeX
	KeySize        = chacha20poly1305.KeySize
	PacketHashSize = 8
)

var (
	RouterPrivateKey = []byte{
		0x96, 0xce, 0x57, 0x8b, 0x00, 0x19, 0x44, 0x27, 0xf2, 0xb9, 0x90, 0x1b, 0x43, 0x56, 0xfd, 0x4f,
		0x56, 0xe1, 0xd9, 0x56, 0x58, 0xf2, 0xf4, 0x3b, 0x86, 0x9f, 0x12, 0x75, 0x24, 0xd2, 0x47, 0xb3,
	}

	BackendPrivateKey = []byte{
		0x15, 0x7c, 0x05, 0xab, 0x38, 0xc6, 0x94, 0x8c, 0x14, 0x0f, 0x08, 0xaa, 0xd4, 0xde, 0x54, 0x9b,
		0x95, 0x54, 0x7a, 0xc7, 0x6b, 0xe1, 0xf3, 0xf6, 0x85, 0x55, 0x76, 0x72, 0x72, 0x7e, 0xc8, 0x04,
		0x4c, 0x61, 0xca, 0x8c, 0x47, 0x87, 0x3e, 0xd4, 0xa0, 0xb5, 0x97, 0xc3, 0xca, 0xe0, 0xcf, 0x71,
		0x08, 0x2d, 0x25, 0x3c, 0x91, 0x0e, 0xd4, 0x6f, 0x19, 0x22, 0xaf, 0xba, 0x25, 0x96, 0xa3, 0x40,
	}

	// This is a simulated key and it not used in production!
	RelayPublicKey = [...]byte{
		0xf5, 0x22, 0xad, 0xc1, 0xee, 0x04, 0x6a, 0xbe, 0x7d, 0x89, 0x0c, 0x81, 0x3a, 0x08, 0x31, 0xba,
		0xdc, 0xdd, 0xb5, 0x52, 0xcb, 0x73, 0x56, 0x10, 0xda, 0xa9, 0xc0, 0xae, 0x08, 0xa2, 0xcf, 0x5e,
	}

	// PacketHashKey for the Blake2b hashing of each packet
	PacketHashKey = []byte{
		0xe3, 0x18, 0x61, 0x72, 0xee, 0x70, 0x62, 0x37, 0x40, 0xf6, 0x0a, 0xea, 0xe0, 0xb5, 0x1a, 0x2c,
		0x2a, 0x47, 0x98, 0x8f, 0x27, 0xec, 0x63, 0x2c, 0x25, 0x04, 0x74, 0x89, 0xaf, 0x5a, 0xeb, 0x24,
	}
)

// HashID hashes a string to a uint64 so it can be used as IDs for Relays, Datacenters, etc.
func HashID(s string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(s))
	return hash.Sum64()
}

// For hashing when speed is the only thing desired
func FastHash(s string) uint64 {
	hash := murmur3.New64()
	hash.Write([]byte(s))
	return hash.Sum64()
}

// GenerateSessionID creates a uint64 from random bytes
func GenerateSessionID() uint64 {
	buf := make([]byte, 8)
	rand.Read(buf)

	id := binary.LittleEndian.Uint64(buf)
	id &= ^((uint64(1)) << 63)

	return id
}

// GenerateRelayKeyPair creates a public and private keypair using crypto/ed25519 and prepends a random 8 byte customer ID
func GenerateCustomerKeyPair() ([]byte, []byte, error) {
	customerID := make([]byte, 8)
	rand.Read(customerID)

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, err
	}

	customerPublicKey := make([]byte, 0)
	customerPublicKey = append(customerPublicKey, customerID...)
	customerPublicKey = append(customerPublicKey, publicKey...)
	customerPrivateKey := make([]byte, 0)
	customerPrivateKey = append(customerPrivateKey, customerID...)
	customerPrivateKey = append(customerPrivateKey, privateKey...)

	return customerPublicKey, customerPrivateKey, nil
}

// Open wraps (x/crypto/nacl/box).Open to make working with byte slices easier
func Open(data []byte, nonce []byte, publicKey []byte, privateKey []byte) ([]byte, bool) {
	var n [NonceSize]byte
	var pub [KeySize]byte
	var priv [KeySize]byte

	copy(n[:], nonce)
	copy(pub[:], publicKey)
	copy(priv[:], privateKey)

	return box.Open(nil, data, &n, &pub, &priv)
}

// Seal wraps (x/crypto/nacl/box).Seal to make working with byte slices easier
func Seal(data []byte, nonce []byte, publicKey []byte, privateKey []byte) []byte {
	var n [NonceSize]byte
	var pub [KeySize]byte
	var priv [KeySize]byte

	copy(n[:], nonce)
	copy(pub[:], publicKey)
	copy(priv[:], privateKey)

	return box.Seal(nil, data, &n, &pub, &priv)
}

// Sign wraps sodiumSign with is a wrapper around libsodium
// We wrap this to avoid inclding C in other libs breaking
// code linting
func Sign(privateKey []byte, data []byte) []byte {
	return sodiumSign(data, privateKey)
}

// Verify wraps sodiumVerify with is a wrapper around libsodium
// We wrap this to avoid inclding C in other libs breaking
// code linting
func Verify(publicKey []byte, data []byte, sig []byte) bool {
	return sodiumVerify(data, sig, publicKey)
}

// Hash wraps sodiumHash with is a wrapper around libsodium
// We wrap this to avoid inclding C in other libs breaking
// code linting
func Hash(key []byte, data []byte) []byte {
	return sodiumHash(data, key)
}

// Check wraps sodiumCheck with is a wrapper around libsodium
// We wrap this to avoid inclding C in other libs breaking
// code linting
func Check(key []byte, data []byte) bool {
	return sodiumCheck(data, key)
}
