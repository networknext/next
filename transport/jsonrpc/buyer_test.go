package jsonrpc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

func TestSessionsMap(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessions := []transport.SessionCacheEntry{
		{CustomerID: 12345, SessionID: 1, RouteDecision: routing.Decision{OnNetworkNext: false}, Location: routing.Location{Latitude: 0, Longitude: 0}},
		{CustomerID: 12345, SessionID: 2, RouteDecision: routing.Decision{OnNetworkNext: true}, Location: routing.Location{Latitude: 13, Longitude: 14}},
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
		var reply jsonrpc.MapReply
		err := svc.SessionsMap(nil, &jsonrpc.MapArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.MapReply
		err := svc.SessionsMap(nil, &jsonrpc.MapArgs{BuyerID: "12345"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(reply.SessionPoints))

		assert.Equal(t, reply.SessionPoints[0].OnNetworkNext, false)
		assert.NotZero(t, reply.SessionPoints[0].Coordinates[0])
		assert.NotZero(t, reply.SessionPoints[0].Coordinates[1])

		assert.Equal(t, reply.SessionPoints[1].OnNetworkNext, true)
		assert.Equal(t, reply.SessionPoints[1].Coordinates[0], float64(13))
		assert.Equal(t, reply.SessionPoints[1].Coordinates[1], float64(14))
	})
}

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

func TestTopSessions(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	buyerID1 := fmt.Sprintf("%x", 111)
	buyerID2 := fmt.Sprintf("%x", 222)

	sessionID1 := fmt.Sprintf("%x", 111)
	sessionID2 := fmt.Sprintf("%x", 222)
	sessionID3 := fmt.Sprintf("%x", 333)
	sessionID4 := "missing"

	redisServer.ZAdd("top-global", 50, sessionID1)
	redisServer.ZAdd("top-global", 100, sessionID2)
	redisServer.ZAdd("top-global", 150, sessionID3)
	redisServer.ZAdd("top-global", 150, sessionID4)

	redisServer.ZAdd(fmt.Sprintf("top-buyer-%s", buyerID2), 50, sessionID1)
	redisServer.ZAdd(fmt.Sprintf("top-buyer-%s", buyerID1), 100, sessionID2)
	redisServer.ZAdd(fmt.Sprintf("top-buyer-%s", buyerID1), 150, sessionID3)
	redisServer.ZAdd(fmt.Sprintf("top-buyer-%s", buyerID1), 150, sessionID4)

	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID1), routing.SessionMeta{ID: sessionID1}, time.Hour)
	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID2), routing.SessionMeta{ID: sessionID2}, time.Hour)
	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID3), routing.SessionMeta{ID: sessionID3}, time.Hour)

	svc := jsonrpc.BuyersService{
		RedisClient: redisClient,
	}

	t.Run("top global", func(t *testing.T) {
		var reply jsonrpc.TopSessionsReply
		err := svc.TopSessions(nil, &jsonrpc.TopSessionsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(reply.Sessions))
		assert.Equal(t, sessionID1, reply.Sessions[0].ID)
		assert.Equal(t, sessionID2, reply.Sessions[1].ID)
		assert.Equal(t, sessionID3, reply.Sessions[2].ID)
	})

	t.Run("top buyer", func(t *testing.T) {
		var reply jsonrpc.TopSessionsReply
		err := svc.TopSessions(nil, &jsonrpc.TopSessionsArgs{BuyerID: buyerID1}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(reply.Sessions))
		assert.Equal(t, sessionID2, reply.Sessions[0].ID)
		assert.Equal(t, sessionID3, reply.Sessions[1].ID)
	})
}

func TestSessionDetails(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionID := fmt.Sprintf("%x", 999)

	meta := routing.SessionMeta{
		Location:   routing.Location{Latitude: 10, Longitude: 20},
		ClientAddr: "127.0.0.1:1313",
		ServerAddr: "10.0.0.1:50000",
		Hops:       3,
		SDK:        "3.4.4",
	}
	slice1 := routing.SessionSlice{
		Timestamp: time.Now(),
		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
	}
	slice2 := routing.SessionSlice{
		Timestamp: time.Now().Add(10 * time.Second),
		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
	}

	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID), meta, 720*time.Hour)
	redisClient.SAdd(fmt.Sprintf("session-%s-slices", sessionID), slice1, slice2)

	svc := jsonrpc.BuyersService{
		RedisClient: redisClient,
	}

	t.Run("session_id not found", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(nil, &jsonrpc.SessionDetailsArgs{SessionID: "nope"}, &reply)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(nil, &jsonrpc.SessionDetailsArgs{SessionID: sessionID}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, meta, reply.Meta)
		assert.Equal(t, slice1.Timestamp.Hour(), reply.Slices[0].Timestamp.Hour())
		assert.Equal(t, slice1.Next, reply.Slices[0].Next)
		assert.Equal(t, slice1.Direct, reply.Slices[0].Direct)
		assert.Equal(t, slice1.Envelope, reply.Slices[0].Envelope)
		assert.Equal(t, slice2.Timestamp.Hour(), reply.Slices[1].Timestamp.Hour())
		assert.Equal(t, slice2.Next, reply.Slices[1].Next)
		assert.Equal(t, slice2.Direct, reply.Slices[1].Direct)
		assert.Equal(t, slice2.Envelope, reply.Slices[1].Envelope)
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
