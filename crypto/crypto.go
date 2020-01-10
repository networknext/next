package crypto

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

// Check checks encyption of the packet
func Check(data []byte, nonce []byte, publicKey []byte, privateKey []byte) bool {
	return C.crypto_box_open((*C.uchar)(&data[0]), (*C.uchar)(&data[0]), C.ulonglong(len(data)), (*C.uchar)(&nonce[0]), (*C.uchar)(&publicKey[0]), (*C.uchar)(&privateKey[0])) != 0
}

// CompareTokens compares two byte arrays (doesn't necessarily belong here)
func CompareTokens(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
