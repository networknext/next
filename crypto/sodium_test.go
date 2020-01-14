package crypto_test

import (
	"crypto/rand"
	"golang.org/x/crypto/nacl/box"
	"testing"

	"github.com/networknext/backend/crypto"
)

func BenchmarkSodiumEncrypt(b *testing.B) {
	randread := rand.Reader

	nonce := make([]byte, 24)
	randread.Read(nonce)

	pubkey, privkey, _ := box.GenerateKey(randread)

	msg := []byte("this is a message to be encrypted")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crypto.Encrypt(privkey[:], pubkey[:], nonce, msg, len(msg))
	}
}

func BenchmarkNaClEncrypt(b *testing.B) {
	randread := rand.Reader

	var nonce [24]byte
	randread.Read(nonce[:])

	pubkey, privkey, _ := box.GenerateKey(randread)

	msg := []byte("this is a message to be encrypted")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crypto.Encrypt_NaCl(privkey, pubkey, &nonce, msg)
	}
}
