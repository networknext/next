package crypto

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"fmt"
	"golang.org/x/crypto/nacl/box"
)

const Crypto_kx_PUBLICKEYBYTES = C.crypto_kx_PUBLICKEYBYTES
const Crypto_box_PUBLICKEYBYTES = C.crypto_box_PUBLICKEYBYTES

// todo: these are bad. we should use the libsodium directly, otherwise they get mixed up and used with the wrong libsodium functions
const KeyBytes = 32
const NonceBytes = 24
const SignatureBytes = C.crypto_sign_BYTES
const PublicKeyBytes = C.crypto_sign_PUBLICKEYBYTES

func Encrypt(senderPrivateKey []byte, receiverPublicKey []byte, nonce []byte, buffer []byte, bytes int) error {
	result := C.crypto_box_easy((*C.uchar)(&buffer[0]),
		(*C.uchar)(&buffer[0]),
		C.ulonglong(bytes),
		(*C.uchar)(&nonce[0]),
		(*C.uchar)(&receiverPublicKey[0]),
		(*C.uchar)(&senderPrivateKey[0]))
	if result != 0 {
		return fmt.Errorf("failed to encrypt: result = %d", result)
	} else {
		return nil
	}
}

func Encrypt_NaCl(senderPrivateKey *[32]byte, receiverPublicKey *[32]byte, nonce *[24]byte, buffer []byte) []byte {
	return box.Seal(nil, buffer, nonce, receiverPublicKey, senderPrivateKey)
}

func Decrypt(senderPublicKey []byte, receiverPrivateKey []byte, nonce []byte, buffer []byte, bytes int) error {
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