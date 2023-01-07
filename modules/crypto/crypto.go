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

func Sign_Keypair() ([]byte, []byte) {
	pub, priv, err := ed25519.GenerateKey(crypto_rand.Reader)
	if err != nil {
		panic(err)
	}
	publicKey, privateKey := new([32]byte), new([64]byte)
	copy((*publicKey)[:], pub)
	copy((*privateKey)[:], priv)
	return publicKey[:], privateKey[:]
}

func Sign(data []byte, privateKey []byte) []byte {
	return ed25519.Sign(ed25519.PrivateKey(privateKey), data)
}

func Verify(data []byte, publicKey []byte, signature []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(publicKey), data, signature)
}

/*
func SDK5_CheckPacketSignature(packetData []byte, publicKey []byte) bool {

	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[16]), C.ulonglong(len(packetData)-16-2-SDK5_CRYPTO_SIGN_BYTES))
	result := C.crypto_sign_final_verify(&state, (*C.uchar)(&packetData[len(packetData)-2-SDK5_CRYPTO_SIGN_BYTES]), (*C.uchar)(&publicKey[0]))

	if result != 0 {
		core.Error("signed packet did not verify")
		return false
	}

	return true
}

func SDK5_SignPacket(packetData []byte, privateKey []byte) {
	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[16]), C.ulonglong(len(packetData)-16-2-SDK5_CRYPTO_SIGN_BYTES))
	C.crypto_sign_final_create(&state, (*C.uchar)(&packetData[len(packetData)-2-SDK5_CRYPTO_SIGN_BYTES]), nil, (*C.uchar)(&privateKey[0]))
}
*/

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
