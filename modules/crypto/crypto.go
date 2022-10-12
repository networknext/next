package crypto

import (
	"crypto/ed25519"
	"crypto/rand"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/poly1305"
)

const (
	Box_MACSize   = poly1305.TagSize
	Box_NonceSize = chacha20poly1305.NonceSizeX
	Box_KeySize   = chacha20poly1305.KeySize
)

// todo: we should roll new keypairs for SDK5, and these should *not* be included in the source code

var RouterPrivateKey = []byte{
	0x96, 0xce, 0x57, 0x8b, 0x00, 0x19, 0x44, 0x27, 0xf2, 0xb9, 0x90, 0x1b, 0x43, 0x56, 0xfd, 0x4f,
	0x56, 0xe1, 0xd9, 0x56, 0x58, 0xf2, 0xf4, 0x3b, 0x86, 0x9f, 0x12, 0x75, 0x24, 0xd2, 0x47, 0xb3,
}

var BackendPrivateKey = []byte{
	0x15, 0x7c, 0x05, 0xab, 0x38, 0xc6, 0x94, 0x8c, 0x14, 0x0f, 0x08, 0xaa, 0xd4, 0xde, 0x54, 0x9b,
	0x95, 0x54, 0x7a, 0xc7, 0x6b, 0xe1, 0xf3, 0xf6, 0x85, 0x55, 0x76, 0x72, 0x72, 0x7e, 0xc8, 0x04,
	0x4c, 0x61, 0xca, 0x8c, 0x47, 0x87, 0x3e, 0xd4, 0xa0, 0xb5, 0x97, 0xc3, 0xca, 0xe0, 0xcf, 0x71,
	0x08, 0x2d, 0x25, 0x3c, 0x91, 0x0e, 0xd4, 0x6f, 0x19, 0x22, 0xaf, 0xba, 0x25, 0x96, 0xa3, 0x40,
}

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

func Box_Open(data []byte, nonce []byte, publicKey []byte, privateKey []byte) ([]byte, bool) {

	var n [Box_NonceSize]byte
	var pub [Box_KeySize]byte
	var priv [Box_KeySize]byte

	copy(n[:], nonce)
	copy(pub[:], publicKey)
	copy(priv[:], privateKey)

	return box.Open(nil, data, &n, &pub, &priv)
}

func Box_Seal(data []byte, nonce []byte, publicKey []byte, privateKey []byte) []byte {
	var n [Box_NonceSize]byte
	var pub [Box_KeySize]byte
	var priv [Box_KeySize]byte

	copy(n[:], nonce)
	copy(pub[:], publicKey)
	copy(priv[:], privateKey)

	return box.Seal(nil, data, &n, &pub, &priv)
}
