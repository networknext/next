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
