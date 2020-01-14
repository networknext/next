package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"log"
)

var RouterPrivateKey = [...]byte{0x96, 0xce, 0x57, 0x8b, 0x00, 0x19, 0x44, 0x27, 0xf2, 0xb9, 0x90, 0x1b, 0x43, 0x56, 0xfd, 0x4f, 0x56, 0xe1, 0xd9, 0x56, 0x58, 0xf2, 0xf4, 0x3b, 0x86, 0x9f, 0x12, 0x75, 0x24, 0xd2, 0x47, 0xb3}

var BackendPrivateKey = []byte{21, 124, 5, 171, 56, 198, 148, 140, 20, 15, 8, 170, 212, 222, 84, 155, 149, 84, 122, 199, 107, 225, 243, 246, 133, 85, 118, 114, 114, 126, 200, 4, 76, 97, 202, 140, 71, 135, 62, 212, 160, 181, 151, 195, 202, 224, 207, 113, 8, 45, 37, 60, 145, 14, 212, 111, 25, 34, 175, 186, 37, 150, 163, 64}

// Check checks encyption of the packet
func Check(data []byte, nonce []byte, publicKey []byte, privateKey []byte) bool {
	return false
	// return C.crypto_box_open((*C.uchar)(&data[0]), (*C.uchar)(&data[0]), C.ulonglong(len(data)), (*C.uchar)(&nonce[0]), (*C.uchar)(&publicKey[0]), (*C.uchar)(&privateKey[0])) != 0
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
