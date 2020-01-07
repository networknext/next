package crypto

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"hash/fnv"
)

func CryptoCheck(data []byte, nonce []byte, publicKey []byte, privateKey []byte) bool {
	return C.crypto_box_open((*C.uchar)(&data[0]), (*C.uchar)(&data[0]), C.ulonglong(len(data)), (*C.uchar)(&nonce[0]), (*C.uchar)(&publicKey[0]), (*C.uchar)(&privateKey[0])) != 0
}
