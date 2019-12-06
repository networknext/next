package core

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"encoding/binary"
	"fmt"
	"net"
	"unsafe"
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

func Encrypt_ChaCha20(buffer []byte, additional []byte, privateKey []byte) ([]byte, []byte, error) {
	nonce := RandomBytes(C.crypto_aead_xchacha20poly1305_ietf_NPUBBYTES)
	encrypted := make([]byte, len(buffer)+C.crypto_aead_xchacha20poly1305_ietf_ABYTES)
	var encryptedLength = C.ulonglong(0)
	result := C.crypto_aead_xchacha20poly1305_ietf_encrypt((*C.uchar)(&encrypted[0]), &encryptedLength,
		(*C.uchar)(&buffer[0]), C.ulonglong(len(buffer)),
		(*C.uchar)(&additional[0]), C.ulonglong(len(additional)),
		nil, (*C.uchar)(&nonce[0]), (*C.uchar)(&privateKey[0]))
	if result != 0 {
		return nil, nil, fmt.Errorf("failed to encrypt chacha20: result = %d", result)
	} else {
		return encrypted, nonce, nil
	}
}

func Decrypt_ChaCha20(encrypted []byte, additional []byte, nonce []byte, privateKey []byte) ([]byte, error) {
	if len(encrypted) <= C.crypto_aead_xchacha20poly1305_ietf_ABYTES {
		return nil, fmt.Errorf("failed to decrypt chacha20: encrypted data is too small")
	}
	decrypted := make([]byte, len(encrypted)-C.crypto_aead_xchacha20poly1305_ietf_ABYTES)
	var decryptedLength = C.ulonglong(0)
	result := C.crypto_aead_xchacha20poly1305_ietf_decrypt((*C.uchar)(&decrypted[0]), &decryptedLength, nil,
		(*C.uchar)(&encrypted[0]), C.ulonglong(len(encrypted)),
		(*C.uchar)(&additional[0]), C.ulonglong(len(additional)),
		(*C.uchar)(&nonce[0]), (*C.uchar)(&privateKey[0]))
	if result != 0 {
		return nil, fmt.Errorf("failed to decrypt chacha20: result = %d", result)
	} else {
		return decrypted, nil
	}
}

func GenerateSessionId() uint64 {
	var sessionId uint64
	C.randombytes_buf(unsafe.Pointer(&sessionId), 8)
	sessionId &= ^((uint64(1)) << 63)
	return sessionId
}

func RandomBytes(bytes int) []byte {
	buffer := make([]byte, bytes)
	C.randombytes_buf(unsafe.Pointer(&buffer[0]), C.size_t(bytes))
	return buffer
}

func CryptoSignVerify(sign_data []byte, signature []byte, public_key []byte) bool {
	if len(public_key) != C.crypto_sign_PUBLICKEYBYTES || len(signature) != C.crypto_sign_BYTES {
		return false
	}
	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&sign_data[0]), C.ulonglong(len(sign_data)))
	return C.crypto_sign_final_verify(&state, (*C.uchar)(&signature[0]), (*C.uchar)(&public_key[0])) == 0
}

func CryptoSignCreate(sign_data []byte, private_key []byte) []byte {
	if len(private_key) != C.crypto_sign_BYTES {
		return nil
	}
	signature := make([]byte, C.crypto_sign_BYTES)
	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&sign_data[0]), C.ulonglong(len(sign_data)))
	C.crypto_sign_final_create(&state, (*C.uchar)(&signature[0]), nil, (*C.uchar)(&private_key[0]))
	return signature
}

func WriteSessionToken(token *SessionToken, buffer []byte) {
	binary.LittleEndian.PutUint64(buffer[0:], token.expireTimestamp)
	binary.LittleEndian.PutUint64(buffer[8:], token.sessionId)
	buffer[8+8] = token.sessionVersion
	buffer[8+8+1] = token.sessionFlags
	binary.LittleEndian.PutUint32(buffer[8+8+2:], token.kbpsUp)
	binary.LittleEndian.PutUint32(buffer[8+8+2+4:], token.kbpsDown)
	WriteAddress(buffer[8+8+2+4+4:], token.nextAddress)
	copy(buffer[8+8+2+4+4+AddressBytes:], token.privateKey)
}

func WriteContinueToken(token *ContinueToken, buffer []byte) {
	binary.LittleEndian.PutUint64(buffer[0:], token.expireTimestamp)
	binary.LittleEndian.PutUint64(buffer[8:], token.sessionId)
	buffer[8+8] = token.sessionVersion
	buffer[8+8+1] = token.sessionFlags
}

func ReadContinueToken(buffer []byte) (*ContinueToken, error) {
	if len(buffer) < ContinueTokenBytes {
		return nil, fmt.Errorf("buffer too small to read continue token")
	}
	token := &ContinueToken{}
	token.expireTimestamp = binary.LittleEndian.Uint64(buffer[0:])
	token.sessionId = binary.LittleEndian.Uint64(buffer[8:])
	token.sessionVersion = buffer[8+8]
	token.sessionFlags = buffer[8+8+1]
	return token, nil
}

func WriteEncryptedContinueToken(buffer []byte, token *ContinueToken, senderPrivateKey []byte, receiverPublicKey []byte) error {
	nonce := RandomBytes(NonceBytes)
	copy(buffer, nonce)
	WriteContinueToken(token, buffer[NonceBytes:])
	result := Encrypt(senderPrivateKey, receiverPublicKey, nonce, buffer[NonceBytes:], ContinueTokenBytes)
	return result
}

func WriteRouteTokens(expireTimestamp uint64, sessionId uint64, sessionVersion uint8, sessionFlags uint8, kbpsUp uint32, kbpsDown uint32, numNodes int, addresses []*net.UDPAddr, publicKeys [][]byte, masterPrivateKey [KeyBytes]byte) ([]byte, error) {
	if numNodes < 1 || numNodes > MaxNodes {
		return nil, fmt.Errorf("invalid numNodes %d. expected value in range [1,%d]", numNodes, MaxNodes)
	}
	privateKey := RandomBytes(KeyBytes)
	tokenData := make([]byte, numNodes*EncryptedSessionTokenBytes)
	for i := 0; i < numNodes; i++ {
		nonce := RandomBytes(NonceBytes)
		token := &SessionToken{}
		token.expireTimestamp = expireTimestamp
		token.sessionId = sessionId
		token.sessionVersion = sessionVersion
		token.sessionFlags = sessionFlags
		// todo: bandwidth limits are temporarily disabled
		token.kbpsUp = kbpsUp * 100
		token.kbpsDown = kbpsDown * 100
		if i != numNodes-1 {
			token.nextAddress = addresses[i+1]
		}
		token.privateKey = privateKey
		err := WriteEncryptedSessionToken(tokenData[i*EncryptedSessionTokenBytes:], token, masterPrivateKey[:], publicKeys[i], nonce)
		if err != nil {
			return nil, err
		}
	}
	return tokenData, nil
}

func WriteContinueTokens(expireTimestamp uint64, sessionId uint64, sessionVersion uint8, sessionFlags uint8, numNodes int, publicKeys [][]byte, masterPrivateKey [KeyBytes]byte) ([]byte, error) {
	if numNodes < 1 || numNodes > MaxNodes {
		return nil, fmt.Errorf("invalid numNodes %d. expected value in range [1,%d]", numNodes, MaxNodes)
	}
	tokenData := make([]byte, numNodes*EncryptedContinueTokenBytes)
	for i := 0; i < numNodes; i++ {
		token := &ContinueToken{}
		token.expireTimestamp = expireTimestamp
		token.sessionId = sessionId
		token.sessionVersion = sessionVersion
		token.sessionFlags = sessionFlags
		err := WriteEncryptedContinueToken(tokenData[i*EncryptedContinueTokenBytes:], token, masterPrivateKey[:], publicKeys[i])
		if err != nil {
			return nil, err
		}
	}
	return tokenData, nil
}

func WriteEncryptedSessionToken(buffer []byte, token *SessionToken, senderPrivateKey []byte, receiverPublicKey []byte, nonce []byte) error {
	copy(buffer, nonce)
	WriteSessionToken(token, buffer[NonceBytes:])
	result := Encrypt(senderPrivateKey, receiverPublicKey, nonce, buffer[NonceBytes:], SessionTokenBytes)
	return result
}

func ReadEncryptedContinueToken(tokenData []byte, senderPublicKey []byte, receiverPrivateKey []byte) (*ContinueToken, error) {
	if len(tokenData) < EncryptedContinueTokenBytes {
		return nil, fmt.Errorf("not enough bytes for encrypted continue token")
	}
	nonce := tokenData[0 : C.crypto_box_NONCEBYTES-1]
	tokenData = tokenData[C.crypto_box_NONCEBYTES:]
	if err := Decrypt(senderPublicKey, receiverPrivateKey, nonce, tokenData, ContinueTokenBytes+C.crypto_box_MACBYTES); err != nil {
		return nil, err
	}
	return ReadContinueToken(tokenData)
}
