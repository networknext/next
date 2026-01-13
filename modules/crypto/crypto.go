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

	Auth_SignatureSize = 32
	Auth_KeySize       = 32

	SecretKey_PublicKeySize  = 32
	SecretKey_PrivateKeySize = 32
	SecretKey_KeySize        = 32
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

func Auth_Key() []byte {
	key := make([]byte, Auth_KeySize)
	C.crypto_auth_keygen((*C.uchar)(&key[0]))
	return key
}

func Auth_Sign(data []byte, key []byte, signature []byte) {
	length := len(data)
	C.crypto_auth((*C.uchar)(&signature[0]), (*C.uchar)(&data[0]), C.ulonglong(length), (*C.uchar)(&key[0]))
}

func Auth_Verify(data []byte, key []byte, signature []byte) bool {
	length := len(data)
	return C.crypto_auth_verify((*C.uchar)(&signature[0]), (*C.uchar)(&data[0]), C.ulonglong(length), (*C.uchar)(&key[0])) == 0
}

// ----------------------------------------------------

func SecretKey_KeyPair() ([]byte, []byte) {
	publicKey := make([]byte, SecretKey_PublicKeySize)
	privateKey := make([]byte, SecretKey_PrivateKeySize)
	C.crypto_kx_keypair((*C.uchar)(&publicKey[0]), (*C.uchar)(&privateKey[0]))
	return publicKey, privateKey
}

func SecretKey_GenerateLocal(localPublicKey []byte, localPrivateKey []byte, remotePublicKey []byte) ([]byte, error) {
	secretKey := make([]byte, SecretKey_KeySize)
	result := C.crypto_kx_client_session_keys((*C.uchar)(&secretKey[0]), nil, (*C.uchar)(&localPublicKey[0]), (*C.uchar)(&localPrivateKey[0]), (*C.uchar)(&remotePublicKey[0]))
	if result != 0 {
		return nil, fmt.Errorf("could not generate local secret key")
	}
	return secretKey, nil
}

func SecretKey_GenerateRemote(remotePublicKey []byte, remotePrivateKey []byte, localPublicKey []byte) ([]byte, error) {
	secretKey := make([]byte, SecretKey_KeySize)
	result := C.crypto_kx_server_session_keys(nil, (*C.uchar)(&secretKey[0]), (*C.uchar)(&remotePublicKey[0]), (*C.uchar)(&remotePrivateKey[0]), (*C.uchar)(&localPublicKey[0]))
	if result != 0 {
		return nil, fmt.Errorf("could not generate remote secret key")
	}
	return secretKey, nil
}

// ----------------------------------------------------

const 
(
	SDK_CRYPTO_SIGN_BYTES             = 64
	SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES  = 32
	SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES = 64
)

func SDK_SignKeypair(publicKey []byte, privateKey []byte) int {
	result := C.crypto_sign_keypair((*C.uchar)(&publicKey[0]), (*C.uchar)(&privateKey[0]))
	return int(result)
}

func SDK_SignPacket(packetData []byte, privateKey []byte) bool {
	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[18]), C.ulonglong(len(packetData)-18-SDK_CRYPTO_SIGN_BYTES))
	result := C.crypto_sign_final_create(&state, (*C.uchar)(&packetData[len(packetData)-SDK_CRYPTO_SIGN_BYTES]), nil, (*C.uchar)(&privateKey[0]))
	if result != 0 {
		return false
	}
	return true
}

func SDK_CheckPacketSignature(packetData []byte, publicKey []byte) bool {
	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[18]), C.ulonglong(len(packetData)-18-SDK_CRYPTO_SIGN_BYTES))
	result := C.crypto_sign_final_verify(&state, (*C.uchar)(&packetData[len(packetData)-SDK_CRYPTO_SIGN_BYTES]), (*C.uchar)(&publicKey[0]))
	if result != 0 {
		return false
	}
	return true
}

// ----------------------------------------------------
