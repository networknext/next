package jsonrpc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

func TestBuyersList(t *testing.T) {
	t.Parallel()
	var storer = storage.InMemory{}

	storer.AddCustomer(context.Background(), routing.Customer{Name: "Local", Code: "local"})
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local"})
	storer.AddCustomer(context.Background(), routing.Customer{Name: "Local Local", Code: "local-local"})
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 2, CompanyCode: "local-local"})
	storer.AddCustomer(context.Background(), routing.Customer{Name: "Local Local Local", Code: "local-local-local"})
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 3, CompanyCode: "local-local-local"})

	logger := log.NewNopLogger()

	svc := jsonrpc.BuyersService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("list - empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.BuyerListReply
		err := svc.Buyers(req, &jsonrpc.BuyerListArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(reply.Buyers))
	})

	t.Run("list - !admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local")
		req = req.WithContext(reqContext)
		var reply jsonrpc.BuyerListReply
		err := svc.Buyers(req, &jsonrpc.BuyerListArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(reply.Buyers))
		assert.Equal(t, "0000000000000001", reply.Buyers[0].ID)
		assert.Equal(t, "local", reply.Buyers[0].CompanyCode)
	})

	t.Run("list - admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		assert.True(t, jsonrpc.VerifyAllRoles(req, jsonrpc.AdminRole))
		var reply jsonrpc.BuyerListReply
		err := svc.Buyers(req, &jsonrpc.BuyerListArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(reply.Buyers))
		assert.Equal(t, "0000000000000001", reply.Buyers[0].ID)
		assert.Equal(t, "local", reply.Buyers[0].CompanyCode)
		assert.Equal(t, "0000000000000002", reply.Buyers[1].ID)
		assert.Equal(t, "local-local", reply.Buyers[1].CompanyCode)
		assert.Equal(t, "0000000000000003", reply.Buyers[2].ID)
		assert.Equal(t, "local-local-local", reply.Buyers[2].CompanyCode)
	})
}

// User Sessions is currently disabled
func TestUserSessions(t *testing.T) {
	t.Parallel()

	var storer = storage.InMemory{}

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), 5, 5)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	userHash1 := fmt.Sprintf("%016x", 111)
	userHash2 := fmt.Sprintf("%016x", 222)

	sessionID1 := fmt.Sprintf("%016x", 111)
	sessionID2 := fmt.Sprintf("%016x", 222)
	sessionID3 := fmt.Sprintf("%016x", 333)
	sessionID4 := "missing"

	now := time.Now()
	secs := now.Unix()
	minutes := secs / 60

	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 50, sessionID1)
	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 100, sessionID2)
	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 150, sessionID3)
	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 150, sessionID4)

	redisServer.ZAdd(fmt.Sprintf("sc-%s-%d", userHash2, minutes), 50, sessionID1)
	redisServer.ZAdd(fmt.Sprintf("sc-%s-%d", userHash1, minutes), 100, sessionID2)
	redisServer.ZAdd(fmt.Sprintf("sc-%s-%d", userHash1, minutes), 150, sessionID3)
	redisServer.ZAdd(fmt.Sprintf("sc-%s-%d", userHash1, minutes), 150, sessionID4)

	redisClient.Set(fmt.Sprintf("sm-%s", sessionID1), transport.SessionMeta{ID: 111, DeltaRTT: 50}.RedisString(), time.Hour)
	redisClient.Set(fmt.Sprintf("sm-%s", sessionID2), transport.SessionMeta{ID: 222, DeltaRTT: 100}.RedisString(), time.Hour)
	redisClient.Set(fmt.Sprintf("sm-%s", sessionID3), transport.SessionMeta{ID: 333, DeltaRTT: 150}.RedisString(), time.Hour)

	logger := log.NewNopLogger()

	svc := jsonrpc.BuyersService{
		Storage:                &storer,
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		RedisPoolUserSessions:  redisPool,
		Logger:                 logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("missing user_hash", func(t *testing.T) {
		var reply jsonrpc.UserSessionsReply
		err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(reply.Sessions))
	})

	t.Run("user_hash not found", func(t *testing.T) {
		var reply jsonrpc.UserSessionsReply
		err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: "12345"}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(reply.Sessions))
	})

	t.Run("list - ID", func(t *testing.T) {
		var reply jsonrpc.UserSessionsReply
		err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: "111"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Sessions), 2)

		assert.Equal(t, reply.Sessions[0].ID, sessionID3)
		assert.Equal(t, reply.Sessions[1].ID, sessionID2)
	})

	t.Run("list - hash", func(t *testing.T) {
		var reply jsonrpc.UserSessionsReply
		err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: userHash1}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Sessions), 2)

		assert.Equal(t, reply.Sessions[0].ID, sessionID3)
		assert.Equal(t, reply.Sessions[1].ID, sessionID2)
	})
}

func TestDatacenterMaps(t *testing.T) {
	var storer = storage.InMemory{}
	dcMap := routing.DatacenterMap{
		Alias:      "some.server.alias",
		BuyerID:    0xbdbebdbf0f7be395,
		Datacenter: 0x7edb88d7b6fc0713,
	}

	buyer := routing.Buyer{
		ID:          0xbdbebdbf0f7be395,
		CompanyCode: "local",
	}

	customer := routing.Customer{
		Code: "local",
		Name: "Local",
	}

	datacenter := routing.Datacenter{
		ID:   0x7edb88d7b6fc0713,
		Name: "local.datacenter",
	}

	ctx := context.Background()
	storer.AddBuyer(ctx, buyer)
	storer.AddCustomer(ctx, customer)
	storer.AddDatacenter(ctx, datacenter)

	logger := log.NewNopLogger()

	svc := jsonrpc.BuyersService{
		Storage: &storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("add", func(t *testing.T) {
		var reply jsonrpc.AddDatacenterMapReply
		var args = jsonrpc.AddDatacenterMapArgs{
			DatacenterMap: dcMap,
		}
		err := svc.AddDatacenterMap(req, &args, &reply)
		assert.NoError(t, err)
	})

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.DatacenterMapsReply
		var args = jsonrpc.DatacenterMapsArgs{
			ID: 0xbdbebdbf0f7be395,
		}
		err := svc.DatacenterMapsForBuyer(req, &args, &reply)
		assert.NoError(t, err)

		assert.Equal(t, "7edb88d7b6fc0713", reply.DatacenterMaps[0].DatacenterID)
		assert.Equal(t, "some.server.alias", reply.DatacenterMaps[0].Alias)
		assert.Equal(t, "bdbebdbf0f7be395", reply.DatacenterMaps[0].BuyerID)
	})

	t.Run("list empty", func(t *testing.T) {
		var reply jsonrpc.DatacenterMapsReply
		var args = jsonrpc.DatacenterMapsArgs{
			ID: 0xbdbebdbf0f7be390,
		}
		err := svc.DatacenterMapsForBuyer(req, &args, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(reply.DatacenterMaps))
	})

	t.Run("remove", func(t *testing.T) {
		var reply jsonrpc.RemoveDatacenterMapReply
		var args = jsonrpc.RemoveDatacenterMapArgs{
			DatacenterMap: dcMap,
		}
		err := svc.RemoveDatacenterMap(req, &args, &reply)
		assert.NoError(t, err)
	})

	// entry has been removed
	t.Run("remove w/ error", func(t *testing.T) {
		var reply jsonrpc.RemoveDatacenterMapReply
		var args = jsonrpc.RemoveDatacenterMapArgs{
			DatacenterMap: dcMap,
		}
		err := svc.RemoveDatacenterMap(req, &args, &reply)
		assert.Error(t, err)
	})

}

func TestTotalSessions(t *testing.T) {
	t.Parallel()
	var storer = storage.InMemory{}

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), 1, 1)

	minutes := time.Now().Unix() / 60

	redisServer.HSet(fmt.Sprintf("n-0000000000000001-%d", minutes), "123", "")
	redisServer.HSet(fmt.Sprintf("n-0000000000000002-%d", minutes), "456", "")
	redisServer.HSet(fmt.Sprintf("n-0000000000000003-%d", minutes), "789", "")

	redisServer.HSet(fmt.Sprintf("d-0000000000000001-%d", minutes), "012", "")

	redisServer.HSet(fmt.Sprintf("c-0000000000000001-%d", minutes), "101", "2")
	redisServer.HSet(fmt.Sprintf("c-0000000000000002-%d", minutes), "102", "1")
	redisServer.HSet(fmt.Sprintf("c-0000000000000003-%d", minutes), "103", "1")

	pubkey := make([]byte, 4)

	ctx := context.Background()

	storer.AddCustomer(ctx, routing.Customer{Code: "local", Name: "Local"})
	storer.AddCustomer(ctx, routing.Customer{Code: "local-local", Name: "Local Local"})
	storer.AddCustomer(ctx, routing.Customer{Code: "local-local-local", Name: "Local Local Local"})

	storer.AddBuyer(ctx, routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	storer.AddBuyer(ctx, routing.Buyer{ID: 2, CompanyCode: "local-local", PublicKey: pubkey})
	storer.AddBuyer(ctx, routing.Buyer{ID: 3, CompanyCode: "local-local-local", PublicKey: pubkey})

	logger := log.NewNopLogger()
	svc := jsonrpc.BuyersService{
		RedisPoolTopSessions:   redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolSessionMap:    redisPool,
		Storage:                &storer,
		Logger:                 logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local")
	req = req.WithContext(reqContext)

	t.Run("all", func(t *testing.T) {
		var reply jsonrpc.TotalSessionsReply
		err := svc.TotalSessions(req, &jsonrpc.TotalSessionsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 3, reply.Next)
		assert.Equal(t, 1, reply.Direct)
	})

	t.Run("filtered - sameBuyer - !admin", func(t *testing.T) {
		var reply jsonrpc.TotalSessionsReply
		// test per buyer counts
		err := svc.TotalSessions(req, &jsonrpc.TotalSessionsArgs{CompanyCode: "local"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, reply.Next)
		assert.Equal(t, 1, reply.Direct)
	})

	t.Run("filtered - !samebuyer - !admin", func(t *testing.T) {
		var reply jsonrpc.TotalSessionsReply
		// test per buyer counts
		err := svc.TotalSessions(req, &jsonrpc.TotalSessionsArgs{CompanyCode: "local-local"}, &reply)
		assert.Error(t, err)
	})

	t.Run("filtered - admin", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.TotalSessionsReply
		// test per buyer counts
		err := svc.TotalSessions(req, &jsonrpc.TotalSessionsArgs{CompanyCode: "local-local"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 0, reply.Direct)
		assert.Equal(t, 1, reply.Next)
	})
}

func TestTotalSessionsWithGhostArmy(t *testing.T) {
	t.Parallel()
	var storer = storage.InMemory{}

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), 1, 1)

	minutes := time.Now().Unix() / 60

	redisServer.HSet(fmt.Sprintf("n-0000000000000000-%d", minutes), "123", "") // ghost army
	redisServer.HSet(fmt.Sprintf("n-0000000000000001-%d", minutes), "456", "")
	redisServer.HSet(fmt.Sprintf("n-0000000000000002-%d", minutes), "789", "")

	redisServer.HSet(fmt.Sprintf("d-0000000000000001-%d", minutes), "012", "")

	redisServer.HSet(fmt.Sprintf("c-0000000000000001-%d", minutes), "102", "2")
	redisServer.HSet(fmt.Sprintf("c-0000000000000002-%d", minutes), "103", "1")

	pubkey := make([]byte, 4)

	ctx := context.Background()

	storer.AddCustomer(ctx, routing.Customer{Code: "local", Name: "Local"})
	storer.AddCustomer(ctx, routing.Customer{Code: "local-local", Name: "Local Local"})
	storer.AddCustomer(ctx, routing.Customer{Code: "local-local-local", Name: "Local Local Local"})

	storer.AddBuyer(ctx, routing.Buyer{ID: 0, CompanyCode: "local", PublicKey: pubkey})
	storer.AddBuyer(ctx, routing.Buyer{ID: 1, CompanyCode: "local-local", PublicKey: pubkey})
	storer.AddBuyer(ctx, routing.Buyer{ID: 2, CompanyCode: "local-local-local", PublicKey: pubkey})

	logger := log.NewNopLogger()
	svc := jsonrpc.BuyersService{
		RedisPoolTopSessions:   redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolSessionMap:    redisPool,
		Storage:                &storer,
		Logger:                 logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local")
	req = req.WithContext(reqContext)

	t.Run("all", func(t *testing.T) {
		var reply jsonrpc.TotalSessionsReply
		err := svc.TotalSessions(req, &jsonrpc.TotalSessionsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 3, reply.Next)
		assert.Equal(t, 51, reply.Direct)
	})

	t.Run("filtered - sameBuyer - !admin", func(t *testing.T) {
		var reply jsonrpc.TotalSessionsReply
		// test per buyer counts
		err := svc.TotalSessions(req, &jsonrpc.TotalSessionsArgs{CompanyCode: "local"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, reply.Next)
		assert.Equal(t, 50, reply.Direct)
	})
}

func TestTopSessions(t *testing.T) {
	t.Parallel()
	var storer = storage.InMemory{}

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), 5, 5)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	buyerID1 := fmt.Sprintf("%016x", 111)
	buyerID2 := fmt.Sprintf("%016x", 222)

	sessionID1 := fmt.Sprintf("%016x", 111)
	sessionID2 := fmt.Sprintf("%016x", 222)
	sessionID3 := fmt.Sprintf("%016x", 333)
	sessionID4 := "missing"

	now := time.Now()
	secs := now.Unix()
	minutes := secs / 60

	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 50, sessionID1)
	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 100, sessionID2)
	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 150, sessionID3)
	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 150, sessionID4)

	redisServer.ZAdd(fmt.Sprintf("sc-%s-%d", buyerID2, minutes), 50, sessionID1)
	redisServer.ZAdd(fmt.Sprintf("sc-%s-%d", buyerID1, minutes), 100, sessionID2)
	redisServer.ZAdd(fmt.Sprintf("sc-%s-%d", buyerID1, minutes), 150, sessionID3)
	redisServer.ZAdd(fmt.Sprintf("sc-%s-%d", buyerID1, minutes), 150, sessionID4)

	redisClient.Set(fmt.Sprintf("sm-%s", sessionID1), transport.SessionMeta{ID: 111, DeltaRTT: 50}.RedisString(), time.Hour)
	redisClient.Set(fmt.Sprintf("sm-%s", sessionID2), transport.SessionMeta{ID: 222, DeltaRTT: 100}.RedisString(), time.Hour)
	redisClient.Set(fmt.Sprintf("sm-%s", sessionID3), transport.SessionMeta{ID: 333, DeltaRTT: 150}.RedisString(), time.Hour)

	pubkey := make([]byte, 4)
	storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 111, CompanyCode: "local", PublicKey: pubkey})
	storer.AddCustomer(context.Background(), routing.Customer{Code: "local-local", Name: "Local Local"})
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 222, CompanyCode: "local-local", PublicKey: pubkey})

	logger := log.NewNopLogger()
	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
		Logger:                 logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local")
	req = req.WithContext(reqContext)

	t.Run("all", func(t *testing.T) {
		var reply jsonrpc.TopSessionsReply
		err := svc.TopSessions(req, &jsonrpc.TopSessionsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(reply.Sessions))
	})

	t.Run("filtered - !admin - sameBuyer", func(t *testing.T) {
		var reply jsonrpc.TopSessionsReply
		err := svc.TopSessions(req, &jsonrpc.TopSessionsArgs{CompanyCode: "local"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(reply.Sessions))
		assert.Equal(t, sessionID3, fmt.Sprintf("%016x", reply.Sessions[0].ID))
		assert.Equal(t, sessionID2, fmt.Sprintf("%016x", reply.Sessions[1].ID))
	})

	t.Run("filtered - !admin - !sameBuyer", func(t *testing.T) {
		var reply jsonrpc.TopSessionsReply
		err := svc.TopSessions(req, &jsonrpc.TopSessionsArgs{CompanyCode: "local-local"}, &reply)
		assert.Error(t, err)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("filtered - admin", func(t *testing.T) {
		var reply jsonrpc.TopSessionsReply
		err := svc.TopSessions(req, &jsonrpc.TopSessionsArgs{CompanyCode: "local-local"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(reply.Sessions))
		assert.Equal(t, sessionID1, fmt.Sprintf("%016x", reply.Sessions[0].ID))
	})
}

func TestSessionDetails(t *testing.T) {
	t.Parallel()

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), 5, 5)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionID := fmt.Sprintf("%016x", 999)

	meta := transport.SessionMeta{
		BuyerID:    111,
		Location:   routing.Location{Latitude: 10, Longitude: 20},
		ClientAddr: "127.0.0.1:1313",
		ServerAddr: "10.0.0.1:50000",
		Hops: []transport.RelayHop{
			{ID: 1234},
			{ID: 1234},
			{ID: 1234},
		},
		SDK: "3.4.4",
		NearbyRelays: []transport.NearRelayPortalData{
			{ID: 1, Name: "local", ClientStats: routing.Stats{RTT: 1, Jitter: 2, PacketLoss: 3}},
		},
	}

	anonMeta := transport.SessionMeta{
		BuyerID:    111,
		Location:   routing.Location{Latitude: 10, Longitude: 20},
		ClientAddr: "127.0.0.1:1313",
		ServerAddr: "10.0.0.1:50000",
		Hops: []transport.RelayHop{
			{ID: 1234},
			{ID: 1234},
			{ID: 1234},
		},
		SDK: "3.4.4",
		NearbyRelays: []transport.NearRelayPortalData{
			{ID: 1, Name: "local", ClientStats: routing.Stats{RTT: 1, Jitter: 2, PacketLoss: 3}},
		},
	}
	anonMeta.Anonymise()

	slice1 := transport.SessionSlice{
		Timestamp: time.Now(),
		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
	}
	slice2 := transport.SessionSlice{
		Timestamp: time.Now().Add(10 * time.Second),
		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
	}

	redisClient.Set(fmt.Sprintf("sm-%s", sessionID), meta.RedisString(), 30*time.Second)
	redisClient.RPush(fmt.Sprintf("ss-%s", sessionID), slice1.RedisString(), slice2.RedisString())

	inMemory := storage.InMemory{}
	inMemory.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	inMemory.AddBuyer(context.Background(), routing.Buyer{ID: 111, CompanyCode: "local"})
	inMemory.AddSeller(context.Background(), routing.Seller{ID: "local"})
	inMemory.AddDatacenter(context.Background(), routing.Datacenter{ID: 1})
	inMemory.AddRelay(context.Background(), routing.Relay{ID: 1, Name: "local", Seller: routing.Seller{ID: "local"}, Datacenter: routing.Datacenter{ID: 1}})

	logger := log.NewNopLogger()
	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &inMemory,
		Logger:                 logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("session_id not found", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: "nope"}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - !admin", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: sessionID}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, anonMeta, reply.Meta)
		assert.Equal(t, slice1.Timestamp.Hour(), reply.Slices[0].Timestamp.Hour())
		assert.Equal(t, slice1.Next, reply.Slices[0].Next)
		assert.Equal(t, slice1.Direct, reply.Slices[0].Direct)
		assert.Equal(t, slice1.Envelope, reply.Slices[0].Envelope)
		assert.Equal(t, slice2.Timestamp.Hour(), reply.Slices[1].Timestamp.Hour())
		assert.Equal(t, slice2.Next, reply.Slices[1].Next)
		assert.Equal(t, slice2.Direct, reply.Slices[1].Direct)
		assert.Equal(t, slice2.Envelope, reply.Slices[1].Envelope)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local")
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("success - admin", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: sessionID}, &reply)
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

func TestSessionMapPoints(t *testing.T) {
	t.Parallel()
	var storer = storage.InMemory{}

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), 5, 5)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	buyerID1 := fmt.Sprintf("%016x", 111)
	buyerID2 := fmt.Sprintf("%016x", 222)

	sessionID1 := fmt.Sprintf("%016x", 111)
	sessionID2 := fmt.Sprintf("%016x", 222)
	sessionID3 := fmt.Sprintf("%016x", 333)

	points := []transport.SessionMapPoint{
		{Latitude: 10, Longitude: 40},
		{Latitude: 20, Longitude: 50},
		{Latitude: 30, Longitude: 60},
	}

	now := time.Now()
	secs := now.Unix()
	minutes := secs / 60

	redisClient.HSet(fmt.Sprintf("n-%s-%d", buyerID1, minutes), sessionID1, points[0].RedisString())
	redisClient.HSet(fmt.Sprintf("n-%s-%d", buyerID2, minutes), sessionID2, points[1].RedisString())
	redisClient.HSet(fmt.Sprintf("d-%s-%d", buyerID1, minutes), sessionID3, points[2].RedisString())

	pubkey := make([]byte, 4)
	storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	storer.AddCustomer(context.Background(), routing.Customer{Code: "local-local", Name: "Local Local"})
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 111, CompanyCode: "local", PublicKey: pubkey})
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 222, CompanyCode: "local-local", PublicKey: pubkey})

	logger := log.NewNopLogger()
	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
		Logger:                 logger,
	}

	err := svc.GenerateMapPointsPerBuyer()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("all", func(t *testing.T) {
		var reply jsonrpc.MapPointsReply
		err := svc.SessionMapPoints(req, &jsonrpc.MapPointsArgs{}, &reply)
		assert.NoError(t, err)

		var mappoints []transport.SessionMapPoint
		err = json.Unmarshal(reply.Points, &mappoints)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(mappoints))
		assert.Contains(t, mappoints, points[0])
		assert.Contains(t, mappoints, points[1])
		assert.Contains(t, mappoints, points[2])
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local")
	req = req.WithContext(reqContext)

	t.Run("filtered - !admin - sameBuyer", func(t *testing.T) {
		var reply jsonrpc.MapPointsReply
		err := svc.SessionMapPoints(req, &jsonrpc.MapPointsArgs{CompanyCode: "local"}, &reply)
		assert.NoError(t, err)

		var mappoints []transport.SessionMapPoint
		err = json.Unmarshal(reply.Points, &mappoints)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(mappoints))
		assert.Equal(t, mappoints[0], points[2])
		assert.Equal(t, mappoints[1], points[0])
	})

	t.Run("filtered - !admin - !sameBuyer", func(t *testing.T) {
		var reply jsonrpc.MapPointsReply
		err := svc.SessionMapPoints(req, &jsonrpc.MapPointsArgs{CompanyCode: "local-local"}, &reply)
		assert.Error(t, err)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("filtered - admin", func(t *testing.T) {
		var reply jsonrpc.MapPointsReply
		err := svc.SessionMapPoints(req, &jsonrpc.MapPointsArgs{CompanyCode: "local-local"}, &reply)
		assert.NoError(t, err)

		var mappoints []transport.SessionMapPoint
		err = json.Unmarshal(reply.Points, &mappoints)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(mappoints))
		assert.Equal(t, mappoints[0], points[1])
	})
}

func TestSessionMap(t *testing.T) {
	t.Parallel()
	var storer = storage.InMemory{}

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), 5, 5)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	buyerID1 := fmt.Sprintf("%016x", 111)
	buyerID2 := fmt.Sprintf("%016x", 222)

	sessionID1 := fmt.Sprintf("%016x", 111)
	sessionID2 := fmt.Sprintf("%016x", 222)
	sessionID3 := fmt.Sprintf("%016x", 333)

	points := []transport.SessionMapPoint{
		{Latitude: 10, Longitude: 40},
		{Latitude: 20, Longitude: 50},
		{Latitude: 30, Longitude: 60},
	}

	now := time.Now()
	secs := now.Unix()
	minutes := secs / 60

	redisClient.HSet(fmt.Sprintf("n-%s-%d", buyerID1, minutes), sessionID1, points[0].RedisString())
	redisClient.HSet(fmt.Sprintf("n-%s-%d", buyerID2, minutes), sessionID2, points[1].RedisString())
	redisClient.HSet(fmt.Sprintf("d-%s-%d", buyerID1, minutes), sessionID3, points[2].RedisString())

	pubkey := make([]byte, 4)
	storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	storer.AddCustomer(context.Background(), routing.Customer{Code: "local-local", Name: "Local Local"})
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 111, CompanyCode: "local", PublicKey: pubkey})
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 222, CompanyCode: "local-local", PublicKey: pubkey})

	logger := log.NewNopLogger()
	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
		Logger:                 logger,
	}

	err := svc.GenerateMapPointsPerBuyer()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("all", func(t *testing.T) {
		var reply jsonrpc.MapPointsReply
		err := svc.SessionMap(req, &jsonrpc.MapPointsArgs{}, &reply)
		assert.NoError(t, err)

		var mappoints [][]interface{}
		err = json.Unmarshal(reply.Points, &mappoints)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(mappoints))
		assert.Equal(t, []interface{}{float64(60), float64(30), false}, mappoints[0])
		assert.Equal(t, []interface{}{float64(40), float64(10), true}, mappoints[1])
		assert.Equal(t, []interface{}{float64(50), float64(20), true}, mappoints[2])
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local")
	req = req.WithContext(reqContext)

	t.Run("filtered - !admin - sameBuyer", func(t *testing.T) {
		var reply jsonrpc.MapPointsReply
		err := svc.SessionMap(req, &jsonrpc.MapPointsArgs{CompanyCode: "local"}, &reply)
		assert.NoError(t, err)

		var mappoints [][]interface{}
		err = json.Unmarshal(reply.Points, &mappoints)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(mappoints))
		assert.Equal(t, []interface{}{float64(60), float64(30), false}, mappoints[0])
		assert.Equal(t, []interface{}{float64(40), float64(10), true}, mappoints[1])
	})

	t.Run("filtered - !admin - !sameBuyer", func(t *testing.T) {
		var reply jsonrpc.MapPointsReply
		err := svc.SessionMap(req, &jsonrpc.MapPointsArgs{CompanyCode: "local-local"}, &reply)
		assert.Error(t, err)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("filtered - admin", func(t *testing.T) {
		var reply jsonrpc.MapPointsReply
		err := svc.SessionMap(req, &jsonrpc.MapPointsArgs{CompanyCode: "local-local"}, &reply)
		assert.NoError(t, err)

		var mappoints [][]interface{}
		err = json.Unmarshal(reply.Points, &mappoints)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(mappoints))
		assert.Equal(t, []interface{}{float64(50), float64(20), true}, mappoints[0])
	})
}

func TestGameConfiguration(t *testing.T) {
	t.Parallel()
	var storer = storage.InMemory{}

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), 1, 1)
	pubkey := make([]byte, 4)
	storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})

	logger := log.NewNopLogger()
	svc := jsonrpc.BuyersService{
		RedisPoolTopSessions:   redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolSessionMap:    redisPool,
		Storage:                &storer,
		Logger:                 logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	jsonrpc.SetIsAnonymous(req, true)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.GameConfiguration(req, &jsonrpc.GameConfigurationArgs{}, &reply)
		assert.Error(t, err)
	})

	jsonrpc.SetIsAnonymous(req, false)

	t.Run("no company", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.GameConfiguration(req, &jsonrpc.GameConfigurationArgs{}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local")
	req = req.WithContext(reqContext)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.GameConfiguration(req, &jsonrpc.GameConfigurationArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.GameConfiguration.PublicKey, "AQAAAAAAAAAAAAAA")
	})
}

func TestUpdateGameConfiguration(t *testing.T) {
	t.Parallel()
	var storer = storage.InMemory{}

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), 1, 1)
	pubkey := make([]byte, 4)
	storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	storer.AddCustomer(context.Background(), routing.Customer{Code: "local-local", Name: "Local Local"})
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey, Live: true})

	logger := log.NewNopLogger()
	svc := jsonrpc.BuyersService{
		RedisPoolTopSessions:   redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolSessionMap:    redisPool,
		Storage:                &storer,
		Logger:                 logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	jsonrpc.SetIsAnonymous(req, true)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{}, &reply)
		assert.Error(t, err)
	})

	jsonrpc.SetIsAnonymous(req, false)

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
		"Owner",
	})
	req = req.WithContext(reqContext)

	t.Run("no company", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{}, &reply)
		assert.Error(t, err)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local-local")
	req = req.WithContext(reqContext)

	t.Run("no public key", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - new buyer", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{NewPublicKey: "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A=="}, &reply)
		assert.NoError(t, err)

		newBuyer, err := storer.BuyerWithCompanyCode("local-local")
		assert.NoError(t, err)

		assert.Equal(t, "local-local", newBuyer.CompanyCode)
		assert.False(t, newBuyer.Live)
		assert.Equal(t, "12939405032490452521", fmt.Sprintf("%d", newBuyer.ID))
		assert.Equal(t, "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==", reply.GameConfiguration.PublicKey)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local")
	req = req.WithContext(reqContext)

	t.Run("success - existing buyer", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{NewPublicKey: "45Q+5CKzGkcf3mh8cD43UM8L6Wn81tVwmmlT3Xvs9HWSJp5Zyh5xZg=="}, &reply)
		assert.NoError(t, err)

		oldBuyer, err := storer.BuyerWithCompanyCode("local")
		assert.NoError(t, err)

		assert.Equal(t, "local", oldBuyer.CompanyCode)
		assert.Equal(t, "5123604488526927075", fmt.Sprintf("%d", oldBuyer.ID))
		assert.True(t, oldBuyer.Live)
		assert.Equal(t, "45Q+5CKzGkcf3mh8cD43UM8L6Wn81tVwmmlT3Xvs9HWSJp5Zyh5xZg==", reply.GameConfiguration.PublicKey)
	})
}

func TestSameBuyerRoleFunction(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})

	logger := log.NewNopLogger()
	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
		Logger:                 logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("fail - no company", func(t *testing.T) {
		sameBuyerRoleFunc := svc.SameBuyerRole("local-local")
		verified, err := sameBuyerRoleFunc(req)
		assert.Error(t, err)
		assert.False(t, verified)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local")
	req = req.WithContext(reqContext)

	t.Run("fail - not same buyer", func(t *testing.T) {
		sameBuyerRoleFunc := svc.SameBuyerRole("local-local")
		verified, err := sameBuyerRoleFunc(req)
		assert.NoError(t, err)
		assert.False(t, verified)
	})

	t.Run("success - !admin", func(t *testing.T) {
		sameBuyerRoleFunc := svc.SameBuyerRole("local")
		verified, err := sameBuyerRoleFunc(req)
		assert.NoError(t, err)
		assert.True(t, verified)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local-local")
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("success - admin", func(t *testing.T) {
		sameBuyerRoleFunc := svc.SameBuyerRole("local")
		verified, err := sameBuyerRoleFunc(req)
		assert.NoError(t, err)
		assert.True(t, verified)
	})
}
