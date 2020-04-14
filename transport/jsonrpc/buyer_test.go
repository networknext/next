package jsonrpc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

func TestSessions(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessions := []transport.SessionCacheEntry{
		{CustomerID: 12345, SessionID: 111, UserHash: 999, DirectRTT: 5, NextRTT: 1},
		{CustomerID: 12345, SessionID: 222, UserHash: 888, DirectRTT: 10, NextRTT: 5},
		{CustomerID: 12345, SessionID: 333, UserHash: 777, DirectRTT: 20, NextRTT: 10},
		{CustomerID: 12345, SessionID: 444, UserHash: 666, DirectRTT: 20, NextRTT: 40},

		{CustomerID: 54321, SessionID: 555, UserHash: 555, DirectRTT: 20, NextRTT: 10},
	}
	for _, session := range sessions {
		buf, err := session.MarshalBinary()
		assert.NoError(t, err)

		err = redisServer.Set(fmt.Sprintf("SESSION-%d-%d", session.CustomerID, session.SessionID), string(buf))
		assert.NoError(t, err)
	}

	svc := jsonrpc.BuyersService{
		RedisClient: redisClient,
	}

	t.Run("missing buyer_id", func(t *testing.T) {
		var reply jsonrpc.SessionsReply
		err := svc.Sessions(nil, &jsonrpc.SessionsArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("session_id not found", func(t *testing.T) {
		var reply jsonrpc.SessionsReply
		err := svc.Sessions(nil, &jsonrpc.SessionsArgs{BuyerID: "12345", SessionID: "3434"}, &reply)
		assert.Error(t, err)
	})

	t.Run("failed to convert buyer_id to uint64", func(t *testing.T) {
		var reply jsonrpc.SessionsReply
		err := svc.Sessions(nil, &jsonrpc.SessionsArgs{BuyerID: "asdgagasdgfa"}, &reply)
		assert.Error(t, err)
	})

	t.Run("failed to convert session_id to uint64", func(t *testing.T) {
		var reply jsonrpc.SessionsReply
		err := svc.Sessions(nil, &jsonrpc.SessionsArgs{SessionID: "asdgagasdgfa"}, &reply)
		assert.Error(t, err)
	})

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.SessionsReply
		err := svc.Sessions(nil, &jsonrpc.SessionsArgs{BuyerID: "12345"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Sessions), 4)

		assert.Equal(t, reply.Sessions[0].SessionID, string("333"))
		assert.Equal(t, reply.Sessions[0].UserHash, string("777"))
		assert.Equal(t, reply.Sessions[0].DirectRTT, float64(20))
		assert.Equal(t, reply.Sessions[0].NextRTT, float64(10))
		assert.Equal(t, reply.Sessions[0].ChangeRTT, float64(-10))

		assert.Equal(t, reply.Sessions[1].SessionID, string("222"))
		assert.Equal(t, reply.Sessions[1].UserHash, string("888"))
		assert.Equal(t, reply.Sessions[1].DirectRTT, float64(10))
		assert.Equal(t, reply.Sessions[1].NextRTT, float64(5))
		assert.Equal(t, reply.Sessions[1].ChangeRTT, float64(-5))

		assert.Equal(t, reply.Sessions[2].SessionID, string("111"))
		assert.Equal(t, reply.Sessions[2].UserHash, string("999"))
		assert.Equal(t, reply.Sessions[2].DirectRTT, float64(5))
		assert.Equal(t, reply.Sessions[2].NextRTT, float64(1))
		assert.Equal(t, reply.Sessions[2].ChangeRTT, float64(-4))

		assert.Equal(t, reply.Sessions[3].SessionID, string("444"))
		assert.Equal(t, reply.Sessions[3].UserHash, string("666"))
		assert.Equal(t, reply.Sessions[3].DirectRTT, float64(20))
		assert.Equal(t, reply.Sessions[3].NextRTT, float64(40))
		assert.Equal(t, reply.Sessions[3].ChangeRTT, float64(20))
	})

	t.Run("single", func(t *testing.T) {
		var reply jsonrpc.SessionsReply
		err := svc.Sessions(nil, &jsonrpc.SessionsArgs{BuyerID: "54321", SessionID: "555"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Sessions), 1)

		assert.Equal(t, reply.Sessions[0].SessionID, string("555"))
		assert.Equal(t, reply.Sessions[0].UserHash, string("555"))
		assert.Equal(t, reply.Sessions[0].DirectRTT, float64(20))
		assert.Equal(t, reply.Sessions[0].NextRTT, float64(10))
		assert.Equal(t, reply.Sessions[0].ChangeRTT, float64(-10))
	})
}

func TestGameConfiguration(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	storer := storage.InMemory{}
	pubkey := make([]byte, 4)
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "local.local.1", PublicKey: pubkey})

	svc := jsonrpc.BuyersService{
		RedisClient: redisClient,
		Storage:     &storer,
	}

	t.Run("missing buyer_id", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.GameConfiguration(nil, &jsonrpc.GameConfigurationArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failed to convert buyer_id to uint64", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.GameConfiguration(nil, &jsonrpc.GameConfigurationArgs{BuyerID: "asdgagasdgfa"}, &reply)
		assert.Error(t, err)
	})

	t.Run("single", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.GameConfiguration(nil, &jsonrpc.GameConfigurationArgs{BuyerID: "1"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.GameConfiguration.PublicKey, "AAAAAA==")
	})

	t.Run("update public key", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(nil, &jsonrpc.GameConfigurationArgs{BuyerID: "1", NewPublicKey: "iVBORw0KGgoAAAANSUhEUgAAAGQAAABkCAYA"}, &reply)

		assert.NoError(t, err)

		assert.Equal(t, reply.GameConfiguration.PublicKey, "iVBORw0KGgoAAAANSUhEUgAAAGQAAABkCAYA")
	})

	t.Run("failed to update public key", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(nil, &jsonrpc.GameConfigurationArgs{BuyerID: "1", NewPublicKey: "askjfgbdalksjdf balkjsdbf lkja flfakjs bdlkafs"}, &reply)

		assert.Error(t, err)
	})
}
