package crypto_test

import (
	"testing"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/crypto"

	"github.com/stretchr/testify/assert"
)

func Test_Sign(t *testing.T) {

	publicKey, privateKey := crypto.Sign_KeyPair()

	data := make([]byte, 256)
	common.RandomBytes(data)

	signature := crypto.Sign(data, privateKey)

	assert.True(t, crypto.Verify(data, publicKey, signature))
}

func Test_Auth(t *testing.T) {

	senderPublicKey, senderPrivateKey := crypto.Box_KeyPair()

	receiverPublicKey, receiverPrivateKey := crypto.Box_KeyPair()

	// encrypt random data and verify we can decrypt it

	nonce := make([]byte, crypto.Box_NonceSize)
	common.RandomBytes(nonce)

	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
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

func Test_Encrypt(t *testing.T) {

	// verify that we can sign data and verify

	key := crypto.Auth_Key()

	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}

	signature := make([]byte, crypto.Auth_SignatureSize)

	crypto.Auth_Sign(data, key, signature)

	assert.True(t, crypto.Auth_Verify(data, key, signature))

	// modify the data, and the verify should fail

	for i := range data {
		data[i] = 0
	}

	assert.False(t, crypto.Auth_Verify(data, key, signature))
}

func Test_SecretKey(t *testing.T) {
	localPublicKey, localPrivateKey := crypto.SecretKey_KeyPair()
	remotePublicKey, remotePrivateKey := crypto.SecretKey_KeyPair()
	localSecretKey, err := crypto.SecretKey_GenerateLocal(localPublicKey, localPrivateKey, remotePublicKey)
	assert.Nil(t, err)
	remoteSecretKey, err := crypto.SecretKey_GenerateRemote(remotePublicKey, remotePrivateKey, localPublicKey)
	assert.Nil(t, err)
	assert.Equal(t, localSecretKey, remoteSecretKey)
}
