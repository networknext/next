package transport_test

import (
	"bytes"
	"crypto/rand"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func relayInitAssertions(t *testing.T, body []byte, expectedCode int, redisClient *redis.Client, relayPublicKey []byte, routerPrivateKey []byte) *httptest.ResponseRecorder {
	if redisClient == nil {
		redisServer, _ := miniredis.Run()
		redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	}

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/relay_init", bytes.NewBuffer(body))

	handler := transport.RelayInitHandlerFunc(redisClient, relayPublicKey, routerPrivateKey)

	handler(recorder, request)

	assert.Equal(t, expectedCode, recorder.Code)

	return recorder
}

func TestRelayInitHandler(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	t.Run("address is invalid", func(t *testing.T) {
		// generate keys
		relayPublicKey, relayPrivateKey, _ := crypto.GenerateRelayKeyPair()
		routerPublicKey, routerPrivateKey, _ := crypto.GenerateRelayKeyPair()

		// generate nonce
		nonce := make([]byte, crypto.NonceSize)
		rand.Read(nonce)

		// generate token
		token := make([]byte, routing.TokenSize)
		rand.Read(token)

		// encrypt token
		encryptedToken := crypto.Seal(token, nonce, routerPublicKey, relayPrivateKey)

		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Version:        0,
			Nonce:          nonce,
			Address:        "invalid",
			EncryptedToken: encryptedToken,
		}
		buff, _ := packet.MarshalBinary()
		relayInitAssertions(t, buff, http.StatusBadRequest, nil, relayPublicKey, routerPrivateKey)
	})

	t.Run("encryption token is 0'ed", func(t *testing.T) {
		// generate keys
		relayPublicKey, _, _ := crypto.GenerateRelayKeyPair()
		_, routerPrivateKey, _ := crypto.GenerateRelayKeyPair()

		// generate nonce
		nonce := make([]byte, crypto.NonceSize)
		rand.Read(nonce)

		// generate token but leave it as 0's
		token := make([]byte, routing.TokenSize)

		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Version:        0,
			Nonce:          nonce,
			Address:        "127.0.0.1:40000",
			EncryptedToken: token,
		}
		buff, _ := packet.MarshalBinary()
		relayInitAssertions(t, buff, http.StatusUnauthorized, nil, relayPublicKey, routerPrivateKey)
	})

	t.Run("nonce bytes are 0'ed", func(t *testing.T) {
		// generate keys
		relayPublicKey, relayPrivateKey, _ := crypto.GenerateRelayKeyPair()
		routerPublicKey, routerPrivateKey, _ := crypto.GenerateRelayKeyPair()

		// generate nonce but leave it as 0's
		nonce := make([]byte, crypto.NonceSize)

		// generate random token
		token := core.RandomBytes(routing.TokenSize)

		// seal it with the bad nonce
		encryptedToken := crypto.Seal(token, nonce, routerPublicKey, relayPrivateKey)

		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Version:        0,
			Nonce:          nonce,
			Address:        "127.0.0.1:40000",
			EncryptedToken: encryptedToken,
		}

		buff, _ := packet.MarshalBinary()

		relayInitAssertions(t, buff, http.StatusUnauthorized, nil, relayPublicKey, routerPrivateKey)
	})

	t.Run("relay already exists", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		// generate keys
		relayPublicKey, relayPrivateKey, _ := crypto.GenerateRelayKeyPair()
		routerPublicKey, routerPrivateKey, _ := crypto.GenerateRelayKeyPair()

		// generate nonce
		nonce := make([]byte, crypto.NonceSize)
		rand.Read(nonce)

		// generate token
		token := make([]byte, routing.TokenSize)
		rand.Read(token)

		// encrypt token
		encryptedToken := crypto.Seal(token, nonce, routerPublicKey, relayPrivateKey)

		name := "some name"
		addr := "127.0.0.1:40000"
		udpAddr, _ := net.ResolveUDPAddr("udp", addr)
		dcname := "another name"

		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Version:        0,
			Nonce:          nonce,
			Address:        addr,
			EncryptedToken: encryptedToken,
		}

		buff, _ := packet.MarshalBinary()

		entry := routing.Relay{
			ID:             core.GetRelayID(addr),
			Name:           name,
			Addr:           *udpAddr,
			Datacenter:     32,
			DatacenterName: dcname,
			PublicKey:      token,
			LastUpdateTime: 1234,
		}

		// get the binary data from the entry
		data, _ := entry.MarshalBinary()

		// set it in the redis instance
		redisServer.HSet(transport.RedisHashName, entry.Key(), string(data))

		relayInitAssertions(t, buff, http.StatusNotFound, redisClient, routerPrivateKey, relayPublicKey)
	})

	t.Run("valid", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		relayPublicKey, relayPrivateKey, _ := crypto.GenerateRelayKeyPair()
		routerPublicKey, routerPrivateKey, _ := crypto.GenerateRelayKeyPair()

		nonce := make([]byte, crypto.NonceSize)
		rand.Read(nonce)

		addr := "127.0.0.1:40000"
		udpAddr, _ := net.ResolveUDPAddr("udp", addr)

		token := make([]byte, routing.TokenSize)
		rand.Read(token)

		encryptedToken := crypto.Seal(token, nonce, routerPublicKey, relayPrivateKey)

		before := uint64(time.Now().Unix())

		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Nonce:          nonce,
			Address:        addr,
			EncryptedToken: encryptedToken,
		}
		buff, _ := packet.MarshalBinary()

		recorder := relayInitAssertions(t, buff, http.StatusOK, redisClient, relayPublicKey, routerPrivateKey)

		header := recorder.Header()
		contentType, _ := header["Content-Type"]
		expected := routing.Relay{
			ID:   core.GetRelayID(addr),
			Name: addr,
			Addr: *udpAddr,
		}

		resp := redisClient.HGet(transport.RedisHashName, expected.Key())

		var actual routing.Relay
		bin, _ := resp.Bytes()
		actual.UnmarshalBinary(bin)

		indx := 0
		body := recorder.Body.Bytes()

		var version uint32
		encoding.ReadUint32(body, &indx, &version)

		var timestamp uint64
		encoding.ReadUint64(body, &indx, &timestamp)

		var publicKey []byte
		encoding.ReadBytes(body, &indx, &publicKey, routing.TokenSize)

		if recorder.Code == 200 {
			assert.Equal(t, "application/octet-stream", contentType[0])
		}
		assert.Equal(t, transport.VersionNumberInitResponse, int(version))
		assert.LessOrEqual(t, before, timestamp)
		assert.GreaterOrEqual(t, uint64(time.Now().Unix()), timestamp)
		assert.Equal(t, actual.PublicKey, publicKey) // entry gets a public key assigned at init which is returned in the response

		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Name, actual.Name)
		assert.Equal(t, expected.Addr, actual.Addr)
		assert.NotZero(t, actual.LastUpdateTime)
		assert.Len(t, actual.PublicKey, 32)
	})
}
