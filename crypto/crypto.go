package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"

	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/poly1305"
)

const (
	MACSize   = poly1305.TagSize
	NonceSize = 24
	KeySize   = 32

	SignatureSize  = ed25519.SignatureSize
	PublicKeySize  = ed25519.PublicKeySize
	PrivateKeySize = ed25519.PrivateKeySize
)

var (
	RouterPrivateKey = [...]byte{
		0x96, 0xce, 0x57, 0x8b, 0x00, 0x19, 0x44, 0x27, 0xf2, 0xb9, 0x90, 0x1b, 0x43, 0x56, 0xfd, 0x4f,
		0x56, 0xe1, 0xd9, 0x56, 0x58, 0xf2, 0xf4, 0x3b, 0x86, 0x9f, 0x12, 0x75, 0x24, 0xd2, 0x47, 0xb3,
	}

	BackendPrivateKey = [...]byte{
		0x15, 0x7c, 0x05, 0xab, 0x38, 0xc6, 0x94, 0x8c, 0x14, 0x0f, 0x08, 0xaa, 0xd4, 0xde, 0x54, 0x9b,
		0x95, 0x54, 0x7a, 0xc7, 0x6b, 0xe1, 0xf3, 0xf6, 0x85, 0x55, 0x76, 0x72, 0x72, 0x7e, 0xc8, 0x04,
		0x4c, 0x61, 0xca, 0x8c, 0x47, 0x87, 0x3e, 0xd4, 0xa0, 0xb5, 0x97, 0xc3, 0xca, 0xe0, 0xcf, 0x71,
		0x08, 0x2d, 0x25, 0x3c, 0x91, 0x0e, 0xd4, 0x6f, 0x19, 0x22, 0xaf, 0xba, 0x25, 0x96, 0xa3, 0x40,
	}

	RelayPublicKey = [...]byte{
		0xf5, 0x22, 0xad, 0xc1, 0xee, 0x04, 0x6a, 0xbe, 0x7d, 0x89, 0x0c, 0x81, 0x3a, 0x08, 0x31, 0xba,
		0xdc, 0xdd, 0xb5, 0x52, 0xcb, 0x73, 0x56, 0x10, 0xda, 0xa9, 0xc0, 0xae, 0x08, 0xa2, 0xcf, 0x5e,
	}
)

func GenerateSessionID() uint64 {
	buf := make([]byte, 8)
	rand.Read(buf)

	id := binary.LittleEndian.Uint64(buf)
	id &= ^((uint64(1)) << 63)

	return id
}

func GenerateRelayKeyPair() ([]byte, []byte, error) {
	return ed25519.GenerateKey(nil)
}

func GenerateCustomerKeyPair() ([]byte, []byte, error) {
	customerId := make([]byte, 8)
	rand.Read(customerId)
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, err
	}

	customerPublicKey := make([]byte, 0)
	customerPublicKey = append(customerPublicKey, customerId...)
	customerPublicKey = append(customerPublicKey, publicKey...)
	customerPrivateKey := make([]byte, 0)
	customerPrivateKey = append(customerPrivateKey, customerId...)
	customerPrivateKey = append(customerPrivateKey, privateKey...)

	return customerPublicKey, customerPrivateKey, nil
}

func Check(data []byte, nonce []byte, publicKey []byte, privateKey []byte) bool {
	var n [NonceSize]byte
	var pub [KeySize]byte
	var priv [KeySize]byte

	copy(n[:], nonce)
	copy(pub[:], publicKey)
	copy(priv[:], privateKey)

	_, ok := box.Open(nil, data, &n, &pub, &priv)
	return ok
}

func Encrypt(buffer []byte, nonce *[24]byte, receiverPublicKey *[32]byte, senderPrivateKey *[32]byte) []byte {
	return box.Seal(nil, buffer, nonce, receiverPublicKey, senderPrivateKey)
}

func Decrypt(buffer []byte, nonce *[24]byte, senderPublicKey *[32]byte, receiverPrivateKey *[32]byte) ([]byte, bool) {
	return box.Open(nil, buffer, nonce, senderPublicKey, receiverPrivateKey)
}

func Sign(privateKey []byte, buffer []byte) []byte {
	return ed25519.Sign(privateKey, buffer)
}

func Verify(publicKey []byte, buffer []byte, sig []byte) bool {
	return ed25519.Verify(publicKey, buffer, sig)
}
