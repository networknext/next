package transport_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/stats"
	"github.com/networknext/backend/storage"

	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func seedRedis(t *testing.T, redisServer *miniredis.Miniredis, addressesToAdd []string) {
	addEntry := func(addr string) {
		relay := routing.NewRelay()
		udpAddr, _ := net.ResolveUDPAddr("udp", addr)
		relay.Addr = *udpAddr
		relay.ID = crypto.HashID(addr)
		bin, _ := relay.MarshalBinary()
		redisServer.HSet(routing.HashKeyAllRelays, relay.Key(), string(bin))
	}

	for _, addr := range addressesToAdd {
		addEntry(addr)
	}

}

func relayUpdateAssertions(t *testing.T, endpoint string, body []byte, expectedCode int, redisClient *redis.Client, statsdb *routing.StatsDatabase) *httptest.ResponseRecorder {
	if redisClient == nil {
		redisServer, _ := miniredis.Run()
		redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	}

	if statsdb == nil {
		statsdb = routing.NewStatsDatabase()
	}

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))

	var handler func(writer http.ResponseWriter, request *http.Request)
	if endpoint == "/relay_update" {
		handler = transport.RelayUpdateHandlerFunc(log.NewNopLogger(), redisClient, statsdb, &metrics.EmptyGauge{}, &metrics.EmptyCounter{})
	} else if endpoint == "/relay_update_json" {
		handler = transport.RelayUpdateJSONHandlerFunc(log.NewNopLogger(), redisClient, statsdb, &metrics.EmptyGauge{}, &metrics.EmptyCounter{}, &stats.NoOpTrafficStatsPublisher{}, &storage.InMemory{})
	}

	handler(recorder, request)

	assert.Equal(t, expectedCode, recorder.Code)

	return recorder
}

func TestRelayUpdateHandler(t *testing.T) {
	const (
		endpointPacketUpdate = "/relay_update"
		endpointJSONUpdate   = "/relay_update_json"
	)

	t.Run("relay data is invalid", func(t *testing.T) {
		buff := make([]byte, 10) // invalid relay packet size
		relayUpdateAssertions(t, endpointPacketUpdate, buff, http.StatusUnprocessableEntity, nil, nil)
	})

	t.Run("relay public token bytes not equal", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
		addr := "127.0.0.1:40000"
		udp, _ := net.ResolveUDPAddr("udp", addr)

		token1 := make([]byte, routing.EncryptedTokenSize)
		rand.Read(token1)
		token2 := make([]byte, routing.EncryptedTokenSize)
		rand.Read(token2)
		packet := transport.RelayUpdatePacket{
			Address:   *udp,
			Token:     token1,
			PingStats: make([]routing.RelayStatsPing, 0),
		}

		entry := routing.Relay{
			ID:   crypto.HashID(addr),
			Addr: *udp,
			Datacenter: routing.Datacenter{
				ID:   1,
				Name: "some name",
			},
			PublicKey:      token2,
			LastUpdateTime: uint64(time.Now().Unix() - 1),
		}

		raw, _ := entry.MarshalBinary()
		redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), string(raw))

		buff, _ := packet.MarshalBinary()
		relayUpdateAssertions(t, endpointPacketUpdate, buff, http.StatusBadRequest, redisClient, nil)
	})

	t.Run("address is invalid", func(t *testing.T) {
		udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		packet := transport.RelayUpdatePacket{
			Address: *udp,
			Token:   make([]byte, routing.EncryptedTokenSize),
		}
		buff, _ := packet.MarshalBinary()
		buff[10] = 'x' // assign this index (which should be the first item in the address) as the letter 'x' making it invalid
		relayUpdateAssertions(t, endpointPacketUpdate, buff, http.StatusUnprocessableEntity, nil, nil)
	})

	t.Run("number of relays exceeds max", func(t *testing.T) {
		udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		packet := transport.RelayUpdatePacket{
			Address:   *udp,
			NumRelays: 1025,
			Token:     make([]byte, routing.EncryptedTokenSize),
			PingStats: make([]routing.RelayStatsPing, 1025),
		}
		buff, _ := packet.MarshalBinary()
		relayUpdateAssertions(t, endpointPacketUpdate, buff, http.StatusBadRequest, nil, nil)
	})

	t.Run("relay not found", func(t *testing.T) {
		udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		packet := transport.RelayUpdatePacket{
			NumRelays: 3,
			Address:   *udp,
			Token:     make([]byte, crypto.KeySize),
			PingStats: make([]routing.RelayStatsPing, 3),
		}
		buff, _ := packet.MarshalBinary()
		relayUpdateAssertions(t, endpointPacketUpdate, buff, http.StatusNotFound, nil, nil)
	})

	t.Run("valid", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
		addr := "127.0.0.1:40000"
		udp, _ := net.ResolveUDPAddr("udp", addr)
		statsdb := routing.NewStatsDatabase()
		statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
		packet := transport.RelayUpdatePacket{
			NumRelays: uint32(len(statIps)),
			Address:   *udp,
			Token:     make([]byte, crypto.KeySize),
		}

		packet.PingStats = make([]routing.RelayStatsPing, packet.NumRelays)
		for i, addr := range statIps {
			stats := &packet.PingStats[i]
			stats.RelayID = crypto.HashID(addr)
			stats.RTT = rand.Float32()
			stats.Jitter = rand.Float32()
			stats.PacketLoss = rand.Float32()
		}

		seedRedis(t, redisServer, statIps)

		entry := routing.Relay{
			ID:   crypto.HashID(addr),
			Addr: *udp,
			Datacenter: routing.Datacenter{
				ID:   1,
				Name: "some name",
			},
			PublicKey:      make([]byte, crypto.KeySize),
			LastUpdateTime: uint64(time.Now().Unix() - 1),
		}

		raw, _ := entry.MarshalBinary()
		redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), string(raw))

		buff, _ := packet.MarshalBinary()

		recorder := relayUpdateAssertions(t, endpointPacketUpdate, buff, http.StatusOK, redisClient, statsdb)

		res := redisClient.HGet(routing.HashKeyAllRelays, entry.Key())
		var actual routing.Relay
		raw, _ = res.Bytes()
		actual.UnmarshalBinary(raw)

		assert.Equal(t, entry.ID, actual.ID)
		assert.Equal(t, entry.Name, actual.Name)
		assert.Equal(t, entry.Addr, actual.Addr)
		assert.Equal(t, entry.Datacenter.ID, actual.Datacenter.ID)
		assert.Equal(t, entry.Datacenter.Name, actual.Datacenter.Name)
		assert.Equal(t, entry.PublicKey, actual.PublicKey)
		assert.NotEqual(t, entry.LastUpdateTime, actual.LastUpdateTime)

		// response assertions
		header := recorder.Header()
		contentType, _ := header["Content-Type"]
		if recorder.Code == 200 {
			assert.Equal(t, "application/octet-stream", contentType[0])
		}

		indx := 0
		body := recorder.Body.Bytes()

		var version uint32
		encoding.ReadUint32(body, &indx, &version)

		var numRelaysToPing uint32
		encoding.ReadUint32(body, &indx, &numRelaysToPing)

		assert.Equal(t, uint32(len(statIps)), numRelaysToPing)

		relaysToPingIDs := make([]uint64, 0)
		relaysToPingAddrs := make([]string, 0)

		for i := 0; uint32(i) < numRelaysToPing; i++ {
			var id uint64
			var addr string
			encoding.ReadUint64(body, &indx, &id)
			encoding.ReadString(body, &indx, &addr, transport.MaxRelayAddressLength)
			relaysToPingIDs = append(relaysToPingIDs, id)
			relaysToPingAddrs = append(relaysToPingAddrs, addr)
		}

		assert.Contains(t, statsdb.Entries, entry.ID)
		relations := statsdb.Entries[entry.ID]
		for _, addr := range statIps {
			id := crypto.HashID(addr)
			assert.Contains(t, relaysToPingIDs, id)
			assert.Contains(t, relaysToPingAddrs, addr)
			assert.Contains(t, relations.Relays, id)
		}

		assert.NotContains(t, relaysToPingIDs, entry.ID)
		assert.NotContains(t, relaysToPingAddrs, packet.Address)
	})

	t.Run("json version", func(t *testing.T) {
		t.Run("unparsable json", func(t *testing.T) {
			JSONData := "{" // basic but gets the job done
			buff := []byte(JSONData)
			relayUpdateAssertions(t, endpointJSONUpdate, buff, http.StatusUnprocessableEntity, nil, nil)
		})

		t.Run("invalid address", func(t *testing.T) {
			statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
			token := make([]byte, crypto.KeySize)
			b64Token := base64.StdEncoding.EncodeToString(token)
			var request transport.RelayUpdateRequestJSON
			request.StringAddr = "invalid address"
			request.Metadata.TokenBase64 = b64Token
			request.PingStats = make([]routing.RelayStatsPing, uint32(len(statIps)))

			for i, addr := range statIps {
				stats := &request.PingStats[i]
				stats.RelayID = crypto.HashID(addr)
				stats.RTT = rand.Float32()
				stats.Jitter = rand.Float32()
				stats.PacketLoss = rand.Float32()
			}

			buff, err := json.Marshal(request)
			assert.Nil(t, err)

			relayUpdateAssertions(t, endpointJSONUpdate, buff, http.StatusUnprocessableEntity, nil, nil)
		})

		t.Run("token invalid base64", func(t *testing.T) {
			statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
			var request transport.RelayUpdateRequestJSON
			request.StringAddr = "127.0.0.1:40000"
			request.Metadata.TokenBase64 = "invalid\n\t\nbase64"
			request.PingStats = make([]routing.RelayStatsPing, uint32(len(statIps)))

			for i, addr := range statIps {
				stats := &request.PingStats[i]
				stats.RelayID = crypto.HashID(addr)
				stats.RTT = rand.Float32()
				stats.Jitter = rand.Float32()
				stats.PacketLoss = rand.Float32()
			}

			buff, err := json.Marshal(request)
			assert.Nil(t, err)

			relayUpdateAssertions(t, endpointJSONUpdate, buff, http.StatusUnprocessableEntity, nil, nil)
		})

		t.Run("valid", func(t *testing.T) {
			redisServer, _ := miniredis.Run()
			redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
			addr := "127.0.0.1:40000"
			udp, _ := net.ResolveUDPAddr("udp", addr)
			statsdb := routing.NewStatsDatabase()
			statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
			token := make([]byte, crypto.KeySize)
			b64Token := base64.StdEncoding.EncodeToString(token)
			var request transport.RelayUpdateRequestJSON
			request.StringAddr = "127.0.0.1:40000"
			request.Metadata.TokenBase64 = b64Token
			request.PingStats = make([]routing.RelayStatsPing, uint32(len(statIps)))

			for i, addr := range statIps {
				stats := &request.PingStats[i]
				stats.RelayID = crypto.HashID(addr)
				stats.RTT = rand.Float32()
				stats.Jitter = rand.Float32()
				stats.PacketLoss = rand.Float32()
			}

			seedRedis(t, redisServer, statIps)

			entry := routing.Relay{
				ID:   crypto.HashID(addr),
				Addr: *udp,
				Datacenter: routing.Datacenter{
					ID:   1,
					Name: "some name",
				},
				PublicKey:      make([]byte, crypto.KeySize),
				LastUpdateTime: uint64(time.Now().Unix() - 1),
			}

			raw, _ := entry.MarshalBinary()
			redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), string(raw))

			buff, err := json.Marshal(request)
			assert.Nil(t, err)

			recorder := relayUpdateAssertions(t, endpointJSONUpdate, buff, http.StatusOK, redisClient, statsdb)

			res := redisClient.HGet(routing.HashKeyAllRelays, entry.Key())
			var actual routing.Relay
			raw, _ = res.Bytes()
			actual.UnmarshalBinary(raw)

			assert.Equal(t, entry.ID, actual.ID)
			assert.Equal(t, entry.Name, actual.Name)
			assert.Equal(t, entry.Addr, actual.Addr)
			assert.Equal(t, entry.Datacenter.ID, actual.Datacenter.ID)
			assert.Equal(t, entry.Datacenter.Name, actual.Datacenter.Name)
			assert.Equal(t, entry.PublicKey, actual.PublicKey)
			assert.NotEqual(t, entry.LastUpdateTime, actual.LastUpdateTime)

			// response assertions
			header := recorder.Header()
			contentType, _ := header["Content-Type"]
			if recorder.Code == http.StatusOK {
				assert.Equal(t, "application/json", contentType[0])
			}

			body := recorder.Body.Bytes()

			var response transport.RelayUpdateResponseJSON
			json.Unmarshal(body, &response)

			assert.Equal(t, len(statIps), len(response.RelaysToPing))

			relaysToPingIDs := make([]uint64, 0)
			relaysToPingAddrs := make([]string, 0)

			for _, data := range response.RelaysToPing {
				relaysToPingIDs = append(relaysToPingIDs, data.ID)
				relaysToPingAddrs = append(relaysToPingAddrs, data.Address)
			}

			assert.Contains(t, statsdb.Entries, entry.ID)
			relations := statsdb.Entries[entry.ID]
			for _, addr := range statIps {
				id := crypto.HashID(addr)
				assert.Contains(t, relaysToPingIDs, id)
				assert.Contains(t, relaysToPingAddrs, addr)
				assert.Contains(t, relations.Relays, id)
			}

			assert.NotContains(t, relaysToPingIDs, entry.ID)
			assert.NotContains(t, relaysToPingAddrs, request.StringAddr)
		})
	})
}
