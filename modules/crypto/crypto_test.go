package crypto_test

import (
	"testing"

	"github.com/networknext/accelerate/modules/common"
	"github.com/networknext/accelerate/modules/crypto"

	"github.com/stretchr/testify/assert"
)

func Test_Sign(t *testing.T) {

	publicKey, privateKey := crypto.Sign_KeyPair()

	data := make([]byte, 256)
	common.RandomBytes(data)

	signature := crypto.Sign(data, privateKey)

	assert.True(t, crypto.Verify(data, publicKey, signature))
}

func Test_Encrypt(t *testing.T) {

	senderPublicKey, senderPrivateKey := crypto.Box_KeyPair()

	receiverPublicKey, receiverPrivateKey := crypto.Box_KeyPair()

	// encrypt random data and verify we can decrypt it

	nonce := make([]byte, crypto.Box_NonceSize)
	common.RandomBytes(nonce)

	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(data[i])
	}

	encryptedData := make([]byte, 256+crypto.Box_MacSize)

	encryptedBytes := crypto.Box_Encrypt(senderPrivateKey[:], receiverPublicKey[:], nonce, encryptedData, len(data))

	assert.Equal(t, 256+crypto.Box_MacSize, encryptedBytes)

	err := crypto.Box_Decrypt(senderPublicKey[:], receiverPrivateKey[:], nonce, encryptedData, encryptedBytes)

	assert.NoError(t, err)

	// decryption should fail with garbage data

	garbageData := make([]byte, 256+crypto.Box_MacSize)
	common.RandomBytes(garbageData[:])

	err = crypto.Box_Decrypt(senderPublicKey[:], receiverPrivateKey[:], nonce, garbageData, encryptedBytes)

	assert.Error(t, err)

	// decryption should fail with the wrong receiver private key

	common.RandomBytes(receiverPrivateKey[:])

	err = crypto.Box_Decrypt(senderPublicKey[:], receiverPrivateKey[:], nonce, encryptedData, encryptedBytes)

	assert.Error(t, err)
}
