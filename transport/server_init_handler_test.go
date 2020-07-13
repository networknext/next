package transport_test

// todo: disabled

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"net"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestServerInitHandlerFunc(t *testing.T) {
	t.Parallel()

	t.Run("failed to read packet", func(t *testing.T) {
		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		initMetrics := metrics.EmptyServerInitMetrics
		localMetrics := metrics.LocalHandler{}

		metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
		assert.NoError(t, err)

		initMetrics.ErrorMetrics.ReadPacketFailure = metric
		serverInitCounters := transport.ServerInitCounters{}

		serverInitParms := transport.ServerInitParams{
			Logger:   log.NewNopLogger(),
			Metrics:  &initMetrics,
			Counters: &serverInitCounters,
		}

		handler := transport.ServerInitHandlerFunc(&serverInitParms)
		handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: []byte("this is not a proper packet")})

		assert.Equal(t, 1.0, initMetrics.ErrorMetrics.ReadPacketFailure.Value())
	})

	t.Run("datacenter not found", func(t *testing.T) {
		t.Skip()
		db := storage.InMemory{}

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		initMetrics := metrics.EmptyServerInitMetrics
		localMetrics := metrics.LocalHandler{}

		metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
		assert.NoError(t, err)

		initMetrics.ErrorMetrics.DatacenterNotFound = metric
		serverInitCounters := transport.ServerInitCounters{}

		packet := transport.ServerInitRequestPacket{
			RequestID:    1,
			CustomerID:   2,
			DatacenterID: 13,

			Version: transport.SDKVersionMin,

			Signature: make([]byte, ed25519.SignatureSize),
		}

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		serverInitParms := transport.ServerInitParams{
			Logger:   log.NewNopLogger(),
			Storer:   &db,
			Metrics:  &initMetrics,
			Counters: &serverInitCounters,
		}

		handler := transport.ServerInitHandlerFunc(&serverInitParms)
		handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

		assert.Equal(t, 1.0, initMetrics.ErrorMetrics.DatacenterNotFound.Value())
	})

	/*
		t.Run("SDK version too old", func(t *testing.T) {
			t.Skip()
			addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
			assert.NoError(t, err)

			initMetrics := metrics.EmptyServerInitMetrics
			localMetrics := metrics.LocalHandler{}

			metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
			assert.NoError(t, err)

			initMetrics.ErrorMetrics.SDKTooOld = metric

			packet := transport.ServerInitRequestPacket{
				RequestID:    1,
				CustomerID:   2,
				DatacenterID: 13,

				Version: transport.SDKVersion{1, 2, 3},

				Signature: make([]byte, ed25519.SignatureSize),
			}

			data, err := packet.MarshalBinary()
			assert.NoError(t, err)

			handler := transport.ServerInitHandlerFunc(log.NewNopLogger(), nil, nil, &initMetrics, nil)
			handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

			assert.Equal(t, 1.0, initMetrics.ErrorMetrics.SDKTooOld.Value())
		})


		t.Run("customer not found", func(t *testing.T) {
			t.Skip()
			db := storage.InMemory{}
			db.AddDatacenter(context.Background(), routing.Datacenter{ID: 13})

			addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
			assert.NoError(t, err)

			initMetrics := metrics.EmptyServerInitMetrics
			localMetrics := metrics.LocalHandler{}

			metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
			assert.NoError(t, err)

			initMetrics.ErrorMetrics.BuyerNotFound = metric

			packet := transport.ServerInitRequestPacket{
				RequestID:    1,
				CustomerID:   2,
				DatacenterID: 13,

				Version: transport.SDKVersionMin,

				Signature: make([]byte, ed25519.SignatureSize),
			}

			data, err := packet.MarshalBinary()
			assert.NoError(t, err)

			handler := transport.ServerInitHandlerFunc(log.NewNopLogger(), nil, &db, &initMetrics, nil)
			handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

			assert.Equal(t, 1.0, initMetrics.ErrorMetrics.BuyerNotFound.Value())
		})

		t.Run("signature verification failed", func(t *testing.T) {
			t.Skip()
			db := storage.InMemory{}
			db.AddDatacenter(context.Background(), routing.Datacenter{ID: 13})
			db.AddBuyer(context.Background(), routing.Buyer{ID: 2})

			addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
			assert.NoError(t, err)

			initMetrics := metrics.EmptyServerInitMetrics
			localMetrics := metrics.LocalHandler{}

			metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
			assert.NoError(t, err)

			initMetrics.ErrorMetrics.VerificationFailure = metric

			packet := transport.ServerInitRequestPacket{
				RequestID:    1,
				CustomerID:   2,
				DatacenterID: 13,

				Version: transport.SDKVersionMin,

				Signature: make([]byte, ed25519.SignatureSize),
			}

			data, err := packet.MarshalBinary()
			assert.NoError(t, err)

			handler := transport.ServerInitHandlerFunc(log.NewNopLogger(), nil, &db, &initMetrics, nil)
			handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

			assert.Equal(t, 1.0, initMetrics.ErrorMetrics.VerificationFailure.Value())
		})

		t.Run("success", func(t *testing.T) {
			t.Skip()
			buyersServerPubKey, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
			assert.NoError(t, err)

			redisServer, err := miniredis.Run()
			assert.NoError(t, err)
			redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

			redisClient.Set("SERVER-2-0.0.0.0:13", 0, 10*time.Second)

			db := storage.InMemory{}
			db.AddDatacenter(context.Background(), routing.Datacenter{ID: 13})
			db.AddBuyer(context.Background(), routing.Buyer{
				ID:        2,
				PublicKey: buyersServerPubKey,
			})

			initMetrics := metrics.EmptyServerInitMetrics
			localMetrics := metrics.LocalHandler{}

			metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
			assert.NoError(t, err)

			initMetrics.ErrorMetrics.SDKTooOld = metric
			initMetrics.ErrorMetrics.BuyerNotFound = metric
			initMetrics.ErrorMetrics.DatacenterNotFound = metric
			initMetrics.ErrorMetrics.VerificationFailure = metric

			// Create a ServerUpdatePacket and marshal it to binary so sent it into the UDP handler
			packet := transport.ServerInitRequestPacket{
				RequestID:    1,
				CustomerID:   2,
				DatacenterID: 13,

				Version: transport.SDKVersionMin,

				Signature: make([]byte, ed25519.SignatureSize),
			}
			packet.Signature = crypto.Sign(buyersServerPrivKey, packet.GetSignData())

			data, err := packet.MarshalBinary()
			assert.NoError(t, err)

			addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
			assert.NoError(t, err)

			// Initialize the UDP handler with the required redis client
			handler := transport.ServerInitHandlerFunc(log.NewNopLogger(), redisClient, &db, &initMetrics, buyersServerPrivKey)
			handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

			assert.Equal(t, 0.0, initMetrics.ErrorMetrics.SDKTooOld.Value())
			assert.Equal(t, 0.0, initMetrics.ErrorMetrics.BuyerNotFound.Value())
			assert.Equal(t, 0.0, initMetrics.ErrorMetrics.DatacenterNotFound.Value())
			assert.Equal(t, 0.0, initMetrics.ErrorMetrics.VerificationFailure.Value())

			cmd := redisClient.Get("SERVER-2-0.0.0.0:13")
			assert.EqualError(t, cmd.Err(), "redis: nil")
			assert.Equal(t, "", cmd.Val())
		})
	*/
}
