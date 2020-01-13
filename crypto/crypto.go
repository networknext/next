package crypto

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"crypto/ed25519"
	"crypto/rand"
	"log"
)

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

func GenerateRelayKeyPair() ([]byte, []byte) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalln(err)
	}
	return publicKey, privateKey
}

func GenerateCustomerKeyPair() ([]byte, []byte) {
	customerId := make([]byte, 8)
	rand.Read(customerId)
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalln(err)
	}
	customerPublicKey := make([]byte, 0)
	customerPublicKey = append(customerPublicKey, customerId...)
	customerPublicKey = append(customerPublicKey, publicKey...)
	customerPrivateKey := make([]byte, 0)
	customerPrivateKey = append(customerPrivateKey, customerId...)
	customerPrivateKey = append(customerPrivateKey, privateKey...)
	return customerPublicKey, customerPrivateKey
}
