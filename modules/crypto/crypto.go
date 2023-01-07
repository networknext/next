package crypto

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"crypto/ed25519"
	crypto_rand "crypto/rand"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/poly1305"
)

const (
	Box_PublicKeySize  = chacha20poly1305.KeySize
	Box_PrivateKeySize = chacha20poly1305.KeySize
	Box_NonceSize      = chacha20poly1305.NonceSizeX
	Box_MacSize        = poly1305.TagSize

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

func Box_Encrypt(senderPrivateKey []byte, receiverPublicKey []byte, nonce []byte, buffer []byte, bytes int) int {
	C.crypto_box_easy((*C.uchar)(&buffer[0]),
		(*C.uchar)(&buffer[0]),
		C.ulonglong(bytes),
		(*C.uchar)(&nonce[0]),
		(*C.uchar)(&receiverPublicKey[0]),
		(*C.uchar)(&senderPrivateKey[0]))
	return bytes + Box_MacSize
}

func Box_Decrypt(senderPublicKey []byte, receiverPrivateKey []byte, nonce []byte, buffer []byte, bytes int) error {
	result := C.crypto_box_open_easy(
		(*C.uchar)(&buffer[0]),
		(*C.uchar)(&buffer[0]),
		C.ulonglong(bytes),
		(*C.uchar)(&nonce[0]),
		(*C.uchar)(&senderPublicKey[0]),
		(*C.uchar)(&receiverPrivateKey[0]))
	if result != 0 {
		return fmt.Errorf("failed to decrypt: result = %d", result)
	} else {
		return nil
	}
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
