package crypto

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

func sodiumSign(sign_data []byte, private_key []byte) []byte {
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

func sodiumVerify(sign_data []byte, signature []byte, public_key []byte) bool {
	if len(public_key) != C.crypto_sign_PUBLICKEYBYTES || len(signature) != C.crypto_sign_BYTES {
		return false
	}
	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&sign_data[0]), C.ulonglong(len(sign_data)))
	return C.crypto_sign_final_verify(&state, (*C.uchar)(&signature[0]), (*C.uchar)(&public_key[0])) == 0
}

func sodiumHash(data []byte, key []byte) []byte {
	signedPacketData := make([]byte, len(data)+PacketHashSize)
	C.crypto_generichash(
		(*C.uchar)(&signedPacketData[0]),
		C.ulong(PacketHashSize),
		(*C.uchar)(&data[0]),
		C.ulonglong(len(data)),
		(*C.uchar)(&key[0]),
		C.ulong(C.crypto_generichash_KEYBYTES),
	)
	for i := 0; i < len(data); i++ {
		signedPacketData[PacketHashSize+i] = data[i]
	}
	return signedPacketData
}

func sodiumCheck(data []byte, key []byte) bool {
	if len(data) <= PacketHashSize {
		return false
	}

	hash := make([]byte, PacketHashSize)
	C.crypto_generichash(
		(*C.uchar)(&hash[0]),
		C.ulong(PacketHashSize),
		(*C.uchar)(&data[PacketHashSize]),
		C.ulonglong(len(data)-PacketHashSize),
		(*C.uchar)(&key[0]),
		C.ulong(C.crypto_generichash_KEYBYTES),
	)
	for i := 0; i < PacketHashSize; i++ {
		if hash[i] != data[i] {
			return false
		}
	}
	return true
}
