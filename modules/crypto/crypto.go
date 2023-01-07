package crypto

import (
	"crypto/ed25519"
	crypto_rand "crypto/rand"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/poly1305"
)

const (
	Box_KeySize   = chacha20poly1305.KeySize
	Box_NonceSize = chacha20poly1305.NonceSizeX
	Box_MacSize   = poly1305.TagSize

	Sign_SignatureSize  = 64
	Sign_PublicKeySize  = 32
	Sign_PrivateKeySize = 64
)

// ----------------------------------------------------

func Box_KeyPair() ([]byte, []byte) {
	publicKey, privateKey, err := box.GenerateKey(crypto_rand.Reader)
	if err != nil {
		panic(err)
	}
	return publicKey[:], privateKey[:]
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

// ----------------------------------------------------

func Sign_KeyPair() ([]byte, []byte) {
	pub, priv, err := ed25519.GenerateKey(crypto_rand.Reader)
	if err != nil {
		panic(err)
	}
	return pub, priv
}

func Sign(data []byte, privateKey []byte) []byte {
	return ed25519.Sign(ed25519.PrivateKey(privateKey), data)
}

func Verify(data []byte, publicKey []byte, signature []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(publicKey), data, signature)
}

// ----------------------------------------------------

func GenerateCustomerKeyPair() ([]byte, []byte, error) {

	customerID := make([]byte, 8)

	crypto_rand.Read(customerID)

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
