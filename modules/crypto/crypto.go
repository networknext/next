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

	Routing_PublicKeySize = ed25519.PublicKeySize
	Routing_PrivateKeySize = ed25519.PrivateKeySize
	Routing_SignatureSize = ed25519.SignatureSize
)

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

func GenerateRoutingKeyPair() ([]byte, []byte) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	return publicKey, privateKey
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
