package jsonrpc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

// This test depends on auth0 and the JWT doesn't have the right permissions.
/* func TestBuyersList(t *testing.T) {

	storer := storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "local.local.1"})

	logger := log.NewNopLogger()

	svc := jsonrpc.BuyersService{
		Storage: &storer,
		Logger:  logger,
	}
	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"
	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Bearer "+jwtSideload)
	res := httptest.NewRecorder()

	authMiddleware.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.BuyerListReply
		err := svc.Buyers(req, &jsonrpc.BuyerListArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, "0000000000000001", reply.Buyers[0].ID)
		assert.Equal(t, "local.local.1", reply.Buyers[0].Name)
	})
} */

// todo: this test is failing with "context deadline exceeded". I believe it's reaching out to Auth0, in which case
// it should be rewritten to not do that.

// func TestUserSessions(t *testing.T) {
// 	t.Parallel()

// 	redisServer, _ := miniredis.Run()
// 	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

// 	userHash1 := fmt.Sprintf("%x", 111)
// 	userHash2 := fmt.Sprintf("%x", 222)

// 	sessionID1 := fmt.Sprintf("%x", 111)
// 	sessionID2 := fmt.Sprintf("%x", 222)
// 	sessionID3 := fmt.Sprintf("%x", 333)
// 	sessionID4 := "missing"

// 	redisServer.SetAdd(fmt.Sprintf("user-%s-sessions", userHash2), sessionID1)
// 	redisServer.SetAdd(fmt.Sprintf("user-%s-sessions", userHash1), sessionID2)
// 	redisServer.SetAdd(fmt.Sprintf("user-%s-sessions", userHash1), sessionID3)
// 	redisServer.SetAdd(fmt.Sprintf("user-%s-sessions", userHash1), sessionID4)

// 	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID1), routing.SessionMeta{ID: sessionID1}, time.Hour)
// 	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID2), routing.SessionMeta{ID: sessionID2}, time.Hour)
// 	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID3), routing.SessionMeta{ID: sessionID3}, time.Hour)

// 	logger := log.NewNopLogger()

// 	svc := jsonrpc.BuyersService{
// 		RedisClient: redisClient,
// 		Logger:      logger,
// 	}

// 	manager, err := management.New(
// 		"networknext.auth0.com",
// 		"0Hn8oZfUwy5UPo6bUk0hYCQ2hMJnwQYg",
// 		"l2namTU5jKVAkuCwV3votIPcP87jcOuJREtscx07aLgo8EykReX69StUVBfJOzx5",
// 	)
// 	assert.NoError(t, err)

// 	auth0Client := storage.Auth0{
// 		Manager: manager,
// 		Logger:  logger,
// 	}

// 	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"
// 	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	}
// 	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
// 	req.Header.Add("Authorization", "Bearer "+jwtSideload)
// 	res := httptest.NewRecorder()

// 	authMiddleware.ServeHTTP(res, req)
// 	assert.Equal(t, http.StatusOK, res.Code)

// 	user := req.Context().Value("user")
// 	assert.NotEqual(t, user, nil)
// 	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

// 	requestID, ok := claims["sub"]

// 	assert.True(t, ok)
// 	assert.Equal(t, "auth0|5b96f61cf1642721ad84eeb6", requestID)

// 	roles, err := auth0Client.Manager.User.Roles(requestID.(string))

// 	assert.NoError(t, err)
// 	req = jsonrpc.SetRoles(req, *roles)

// 	t.Run("missing user_hash", func(t *testing.T) {
// 		var reply jsonrpc.UserSessionsReply
// 		err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{}, &reply)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 0, len(reply.Sessions))
// 	})

// 	t.Run("user_hash not found", func(t *testing.T) {
// 		var reply jsonrpc.UserSessionsReply
// 		err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserHash: "12345"}, &reply)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 0, len(reply.Sessions))
// 	})

// 	t.Run("list", func(t *testing.T) {
// 		var reply jsonrpc.UserSessionsReply
// 		err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserHash: userHash1}, &reply)
// 		assert.NoError(t, err)

// 		assert.Equal(t, len(reply.Sessions), 2)

// 		assert.Equal(t, reply.Sessions[0].ID, sessionID3)
// 		assert.Equal(t, reply.Sessions[1].ID, sessionID2)
// 	})
// }

func TestDatacenterMaps(t *testing.T) {
	dcMap := routing.DatacenterMap{
		Alias:      "some.server.alias",
		BuyerID:    0xbdbebdbf0f7be395,
		Datacenter: 0x7edb88d7b6fc0713,
	}

	storer := storage.InMemory{}

	buyer := routing.Buyer{
		ID:   0xbdbebdbf0f7be395,
		Name: "local.buyer",
	}

	datacenter := routing.Datacenter{
		ID:   0x7edb88d7b6fc0713,
		Name: "local.datacenter",
	}

	storer.AddBuyer(context.Background(), buyer)
	storer.AddDatacenter(context.Background(), datacenter)

	logger := log.NewNopLogger()

	svc := jsonrpc.BuyersService{
		Storage: &storer,
		Logger:  logger,
	}
	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"
	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Bearer "+jwtSideload)
	res := httptest.NewRecorder()

	authMiddleware.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

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

	// belongs in ops
	// t.Run("list w/o buyer ID", func(t *testing.T) {
	// 	var reply jsonrpc.DatacenterMapsReply
	// 	var args = jsonrpc.DatacenterMapsArgs{
	// 		ID: "",
	// 	}
	// 	err := svc.ListDatacenterMaps(req, &args, &reply)
	// 	assert.NoError(t, err)

	// 	assert.Equal(t, "7edb88d7b6fc0713", reply.DatacenterMaps[id].Datacenter)
	// 	assert.Equal(t, "some.server.alias", reply.DatacenterMaps[id].Alias)
	// 	assert.Equal(t, "bdbebdbf0f7be395", reply.DatacenterMaps[id].BuyerID)

	// })

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

	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	redisServer.ZAdd("total-next", 10, "session-1")
	redisServer.ZAdd("total-next", 20, "session-2")
	redisServer.ZAdd("total-next", 30, "session-5")
	redisServer.ZAdd("total-direct", 5, "session-2")

	logger := log.NewNopLogger()
	svc := jsonrpc.BuyersService{
		RedisClient: redisClient,
		Logger:      logger,
	}

	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"
	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Bearer "+jwtSideload)
	res := httptest.NewRecorder()

	authMiddleware.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	var reply jsonrpc.TotalSessionsReply
	err := svc.TotalSessions(req, &jsonrpc.TotalSessionsArgs{}, &reply)
	assert.NoError(t, err)

	assert.Equal(t, 3, reply.Next)
	assert.Equal(t, 1, reply.Direct)
}

// todo: this test is failing with "context deadline exceeded". I believe it's reaching out to Auth0, in which case
// it should be rewritten to not do that.

// func TestTopSessions(t *testing.T) {
// 	t.Parallel()

// 	redisServer, _ := miniredis.Run()
// 	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

// 	buyerID1 := fmt.Sprintf("%x", 111)
// 	buyerID2 := fmt.Sprintf("%x", 222)

// 	sessionID1 := fmt.Sprintf("%x", 111)
// 	sessionID2 := fmt.Sprintf("%x", 222)
// 	sessionID3 := fmt.Sprintf("%x", 333)
// 	sessionID4 := "missing"

// 	redisServer.ZAdd("total-next", 50, sessionID1)
// 	redisServer.ZAdd("total-next", 100, sessionID2)
// 	redisServer.ZAdd("total-next", 150, sessionID3)
// 	redisServer.ZAdd("total-next", 150, sessionID4)

// 	redisServer.ZAdd(fmt.Sprintf("total-next-buyer-%s", buyerID2), 50, sessionID1)
// 	redisServer.ZAdd(fmt.Sprintf("total-next-buyer-%s", buyerID1), 100, sessionID2)
// 	redisServer.ZAdd(fmt.Sprintf("total-next-buyer-%s", buyerID1), 150, sessionID3)
// 	redisServer.ZAdd(fmt.Sprintf("total-next-buyer-%s", buyerID1), 150, sessionID4)

// 	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID1), routing.SessionMeta{ID: sessionID1, DeltaRTT: 50}, time.Hour)
// 	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID2), routing.SessionMeta{ID: sessionID2, DeltaRTT: 100}, time.Hour)
// 	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID3), routing.SessionMeta{ID: sessionID3, DeltaRTT: 150}, time.Hour)

// 	storer := storage.InMemory{}
// 	pubkey := make([]byte, 4)
// 	storer.AddBuyer(context.Background(), routing.Buyer{ID: 111, Name: "local.local.1", PublicKey: pubkey, Domain: "networknext.com"})

// 	logger := log.NewNopLogger()
// 	svc := jsonrpc.BuyersService{
// 		RedisClient: redisClient,
// 		Storage:     &storer,
// 		Logger:      logger,
// 	}

// 	manager, err := management.New(
// 		"networknext.auth0.com",
// 		"0Hn8oZfUwy5UPo6bUk0hYCQ2hMJnwQYg",
// 		"l2namTU5jKVAkuCwV3votIPcP87jcOuJREtscx07aLgo8EykReX69StUVBfJOzx5",
// 	)
// 	assert.NoError(t, err)

// 	auth0Client := storage.Auth0{
// 		Manager: manager,
// 		Logger:  logger,
// 	}

// 	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"
// 	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	}
// 	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
// 	req.Header.Add("Authorization", "Bearer "+jwtSideload)
// 	res := httptest.NewRecorder()

// 	authMiddleware.ServeHTTP(res, req)
// 	assert.Equal(t, http.StatusOK, res.Code)

// 	user := req.Context().Value("user")
// 	assert.NotEqual(t, user, nil)
// 	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

// 	requestID, ok := claims["sub"]

// 	assert.True(t, ok)
// 	assert.Equal(t, "auth0|5b96f61cf1642721ad84eeb6", requestID)

// 	roles, err := auth0Client.Manager.User.Roles(requestID.(string))

// 	assert.NoError(t, err)
// 	req = jsonrpc.SetRoles(req, *roles)

// 	t.Run("top global", func(t *testing.T) {
// 		var reply jsonrpc.TopSessionsReply
// 		err := svc.TopSessions(req, &jsonrpc.TopSessionsArgs{}, &reply)
// 		assert.NoError(t, err)

// 		assert.Equal(t, 3, len(reply.Sessions))
// 		assert.Equal(t, sessionID3, reply.Sessions[0].ID)
// 		assert.Equal(t, sessionID2, reply.Sessions[1].ID)
// 		assert.Equal(t, sessionID1, reply.Sessions[2].ID)
// 	})

// 	t.Run("top buyer", func(t *testing.T) {
// 		var reply jsonrpc.TopSessionsReply
// 		err := svc.TopSessions(req, &jsonrpc.TopSessionsArgs{BuyerID: buyerID1}, &reply)
// 		assert.NoError(t, err)

// 		assert.Equal(t, 2, len(reply.Sessions))
// 		assert.Equal(t, sessionID3, reply.Sessions[0].ID)
// 		assert.Equal(t, sessionID2, reply.Sessions[1].ID)
// 	})
// }

// todo: this test is failing with "context deadline exceeded". I believe it's reaching out to Auth0, in which case
// it should be rewritten to not do that.

// func TestSessionDetails(t *testing.T) {
// 	t.Parallel()

// 	redisServer, _ := miniredis.Run()
// 	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

// 	sessionID := fmt.Sprintf("%x", 999)

// 	meta := routing.SessionMeta{
// 		Location:   routing.Location{Latitude: 10, Longitude: 20},
// 		ClientAddr: "127.0.0.1:1313",
// 		ServerAddr: "10.0.0.1:50000",
// 		Hops: []routing.Relay{
// 			{ID: 1234},
// 			{ID: 1234},
// 			{ID: 1234},
// 		},
// 		SDK: "3.4.4",
// 		NearbyRelays: []routing.Relay{
// 			{ID: 1, Name: "local", ClientStats: routing.Stats{RTT: 1, Jitter: 2, PacketLoss: 3}},
// 		},
// 	}
// 	slice1 := routing.SessionSlice{
// 		Timestamp: time.Now(),
// 		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
// 		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
// 		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
// 	}
// 	slice2 := routing.SessionSlice{
// 		Timestamp: time.Now().Add(10 * time.Second),
// 		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
// 		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
// 		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
// 	}

// 	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID), meta, 30*time.Second)
// 	redisClient.SAdd(fmt.Sprintf("session-%s-slices", sessionID), slice1, slice2)

// 	inMemory := storage.InMemory{}
// 	inMemory.AddSeller(context.Background(), routing.Seller{ID: "local"})
// 	inMemory.AddDatacenter(context.Background(), routing.Datacenter{ID: 1})
// 	inMemory.AddRelay(context.Background(), routing.Relay{ID: 1, Name: "local", Seller: routing.Seller{ID: "local"}, Datacenter: routing.Datacenter{ID: 1}})

// 	logger := log.NewNopLogger()
// 	svc := jsonrpc.BuyersService{
// 		RedisClient: redisClient,
// 		Storage:     &inMemory,
// 		Logger:      logger,
// 	}

// 	manager, err := management.New(
// 		"networknext.auth0.com",
// 		"0Hn8oZfUwy5UPo6bUk0hYCQ2hMJnwQYg",
// 		"l2namTU5jKVAkuCwV3votIPcP87jcOuJREtscx07aLgo8EykReX69StUVBfJOzx5",
// 	)
// 	assert.NoError(t, err)

// 	auth0Client := storage.Auth0{
// 		Manager: manager,
// 		Logger:  logger,
// 	}

// 	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"
// 	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	}
// 	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
// 	req.Header.Add("Authorization", "Bearer "+jwtSideload)
// 	res := httptest.NewRecorder()

// 	authMiddleware.ServeHTTP(res, req)
// 	assert.Equal(t, http.StatusOK, res.Code)

// 	user := req.Context().Value("user")
// 	assert.NotEqual(t, user, nil)
// 	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

// 	requestID, ok := claims["sub"]

// 	assert.True(t, ok)
// 	assert.Equal(t, "auth0|5b96f61cf1642721ad84eeb6", requestID)

// 	roles, err := auth0Client.Manager.User.Roles(requestID.(string))

// 	assert.NoError(t, err)
// 	req = jsonrpc.SetRoles(req, *roles)

// 	t.Run("session_id not found", func(t *testing.T) {
// 		var reply jsonrpc.SessionDetailsReply
// 		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: "nope"}, &reply)
// 		assert.Error(t, err)
// 	})

// 	t.Run("success", func(t *testing.T) {
// 		var reply jsonrpc.SessionDetailsReply
// 		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: sessionID}, &reply)
// 		assert.NoError(t, err)
// 		assert.Equal(t, meta, reply.Meta)
// 		assert.Equal(t, slice1.Timestamp.Hour(), reply.Slices[0].Timestamp.Hour())
// 		assert.Equal(t, slice1.Next, reply.Slices[0].Next)
// 		assert.Equal(t, slice1.Direct, reply.Slices[0].Direct)
// 		assert.Equal(t, slice1.Envelope, reply.Slices[0].Envelope)
// 		assert.Equal(t, slice2.Timestamp.Hour(), reply.Slices[1].Timestamp.Hour())
// 		assert.Equal(t, slice2.Next, reply.Slices[1].Next)
// 		assert.Equal(t, slice2.Direct, reply.Slices[1].Direct)
// 		assert.Equal(t, slice2.Envelope, reply.Slices[1].Envelope)
// 	})
// }

// todo: this test is failing with "context deadline exceeded". I believe it's reaching out to Auth0, in which case
// it should be rewritten to not do that.

// func TestSessionMapPoints(t *testing.T) {
// 	t.Parallel()

// 	redisServer, _ := miniredis.Run()
// 	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

// 	buyerID1 := fmt.Sprintf("%016x", 111)
// 	buyerID2 := fmt.Sprintf("%016x", 222)

// 	sessionID1 := fmt.Sprintf("%016x", 111)
// 	sessionID2 := fmt.Sprintf("%016x", 222)
// 	sessionID3 := fmt.Sprintf("%016x", 333)
// 	sessionID4 := "missing"

// 	redisServer.SetAdd(fmt.Sprintf("map-points-%s-buyer", buyerID2), sessionID1)
// 	redisServer.SetAdd(fmt.Sprintf("map-points-%s-buyer", buyerID1), sessionID2)
// 	redisServer.SetAdd(fmt.Sprintf("map-points-%s-buyer", buyerID1), sessionID3)
// 	redisServer.SetAdd(fmt.Sprintf("map-points-%s-buyer", buyerID1), sessionID4)

// 	points := []routing.SessionMapPoint{
// 		{Latitude: 10, Longitude: 40, OnNetworkNext: true},
// 		{Latitude: 20, Longitude: 50, OnNetworkNext: false},
// 		{Latitude: 30, Longitude: 60, OnNetworkNext: true},
// 	}

// 	redisClient.Set(fmt.Sprintf("session-%s-point", sessionID1), points[0], time.Hour)
// 	redisClient.Set(fmt.Sprintf("session-%s-point", sessionID2), points[1], time.Hour)
// 	redisClient.Set(fmt.Sprintf("session-%s-point", sessionID3), points[2], time.Hour)

// 	storer := storage.InMemory{}
// 	pubkey := make([]byte, 4)
// 	storer.AddBuyer(context.Background(), routing.Buyer{ID: 111, Name: "local.local.1", PublicKey: pubkey})
// 	storer.AddBuyer(context.Background(), routing.Buyer{ID: 222, Name: "local.local.2", PublicKey: pubkey})

// 	logger := log.NewNopLogger()
// 	svc := jsonrpc.BuyersService{
// 		RedisClient: redisClient,
// 		Storage:     &storer,
// 		Logger:      logger,
// 	}

// 	err := svc.GenerateMapPointsPerBuyer()
// 	assert.NoError(t, err)

// 	manager, err := management.New(
// 		"networknext.auth0.com",
// 		"0Hn8oZfUwy5UPo6bUk0hYCQ2hMJnwQYg",
// 		"l2namTU5jKVAkuCwV3votIPcP87jcOuJREtscx07aLgo8EykReX69StUVBfJOzx5",
// 	)
// 	assert.NoError(t, err)

// 	auth0Client := storage.Auth0{
// 		Manager: manager,
// 		Logger:  logger,
// 	}

// 	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"
// 	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	}
// 	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
// 	req.Header.Add("Authorization", "Bearer "+jwtSideload)
// 	res := httptest.NewRecorder()

// 	authMiddleware.ServeHTTP(res, req)
// 	assert.Equal(t, http.StatusOK, res.Code)

// 	user := req.Context().Value("user")
// 	assert.NotEqual(t, user, nil)
// 	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

// 	requestID, ok := claims["sub"]

// 	assert.True(t, ok)
// 	assert.Equal(t, "auth0|5b96f61cf1642721ad84eeb6", requestID)

// 	roles, err := auth0Client.Manager.User.Roles(requestID.(string))

// 	assert.NoError(t, err)
// 	req = jsonrpc.SetRoles(req, *roles)

// 	t.Run("points global", func(t *testing.T) {
// 		var reply jsonrpc.MapPointsReply
// 		err := svc.SessionMapPoints(req, &jsonrpc.MapPointsArgs{}, &reply)
// 		assert.NoError(t, err)

// 		var mappoints []routing.SessionMapPoint
// 		err = json.Unmarshal(reply.Points, &mappoints)
// 		assert.NoError(t, err)

// 		assert.Equal(t, 3, len(mappoints))
// 		assert.Contains(t, mappoints, points[0])
// 		assert.Contains(t, mappoints, points[1])
// 		assert.Contains(t, mappoints, points[2])
// 	})

// 	t.Run("points by buyer", func(t *testing.T) {
// 		var reply jsonrpc.MapPointsReply
// 		err := svc.SessionMapPoints(req, &jsonrpc.MapPointsArgs{BuyerID: buyerID2}, &reply)
// 		assert.NoError(t, err)

// 		var mappoints []routing.SessionMapPoint
// 		err = json.Unmarshal(reply.Points, &mappoints)
// 		assert.NoError(t, err)

// 		assert.Equal(t, 1, len(mappoints))
// 		assert.Contains(t, mappoints, points[0])
// 	})
// }

// todo: this test is failing with "context deadline exceeded". I believe it's reaching out to Auth0, in which case
// it should be rewritten to not do that.

// func TestSessionMap(t *testing.T) {
// 	t.Parallel()

// 	redisServer, _ := miniredis.Run()
// 	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

// 	buyerID1 := fmt.Sprintf("%016x", 111)
// 	buyerID2 := fmt.Sprintf("%016x", 222)

// 	sessionID1 := fmt.Sprintf("%016x", 111)
// 	sessionID2 := fmt.Sprintf("%016x", 222)
// 	sessionID3 := fmt.Sprintf("%016x", 333)
// 	sessionID4 := "missing"

// 	redisServer.SetAdd(fmt.Sprintf("map-points-%s-buyer", buyerID2), sessionID1)
// 	redisServer.SetAdd(fmt.Sprintf("map-points-%s-buyer", buyerID1), sessionID2)
// 	redisServer.SetAdd(fmt.Sprintf("map-points-%s-buyer", buyerID1), sessionID3)
// 	redisServer.SetAdd(fmt.Sprintf("map-points-%s-buyer", buyerID1), sessionID4)

// 	points := []routing.SessionMapPoint{
// 		{Latitude: 10, Longitude: 40, OnNetworkNext: true},
// 		{Latitude: 20, Longitude: 50, OnNetworkNext: false},
// 		{Latitude: 30, Longitude: 60, OnNetworkNext: true},
// 	}

// 	redisClient.Set(fmt.Sprintf("session-%s-point", sessionID1), points[0], time.Hour)
// 	redisClient.Set(fmt.Sprintf("session-%s-point", sessionID2), points[1], time.Hour)
// 	redisClient.Set(fmt.Sprintf("session-%s-point", sessionID3), points[2], time.Hour)

// 	storer := storage.InMemory{}
// 	pubkey := make([]byte, 4)
// 	storer.AddBuyer(context.Background(), routing.Buyer{ID: 111, Name: "local.local.1", PublicKey: pubkey})
// 	storer.AddBuyer(context.Background(), routing.Buyer{ID: 222, Name: "local.local.2", PublicKey: pubkey})

// 	logger := log.NewNopLogger()
// 	svc := jsonrpc.BuyersService{
// 		RedisClient: redisClient,
// 		Storage:     &storer,
// 		Logger:      logger,
// 	}

// 	err := svc.GenerateMapPointsPerBuyer()
// 	assert.NoError(t, err)

// 	manager, err := management.New(
// 		"networknext.auth0.com",
// 		"0Hn8oZfUwy5UPo6bUk0hYCQ2hMJnwQYg",
// 		"l2namTU5jKVAkuCwV3votIPcP87jcOuJREtscx07aLgo8EykReX69StUVBfJOzx5",
// 	)
// 	assert.NoError(t, err)

// 	auth0Client := storage.Auth0{
// 		Manager: manager,
// 		Logger:  logger,
// 	}

// 	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"
// 	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	}
// 	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
// 	req.Header.Add("Authorization", "Bearer "+jwtSideload)
// 	res := httptest.NewRecorder()

// 	authMiddleware.ServeHTTP(res, req)
// 	assert.Equal(t, http.StatusOK, res.Code)

// 	user := req.Context().Value("user")
// 	assert.NotEqual(t, user, nil)
// 	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

// 	requestID, ok := claims["sub"]

// 	assert.True(t, ok)
// 	assert.Equal(t, "auth0|5b96f61cf1642721ad84eeb6", requestID)

// 	roles, err := auth0Client.Manager.User.Roles(requestID.(string))

// 	assert.NoError(t, err)
// 	req = jsonrpc.SetRoles(req, *roles)

// 	t.Run("points global", func(t *testing.T) {
// 		var reply jsonrpc.MapPointsReply
// 		err := svc.SessionMap(req, &jsonrpc.MapPointsArgs{}, &reply)
// 		assert.NoError(t, err)

// 		var mappoints [][]interface{}
// 		err = json.Unmarshal(reply.Points, &mappoints)
// 		assert.NoError(t, err)

// 		assert.Equal(t, 3, len(mappoints))
// 		assert.Equal(t, []interface{}{float64(50), float64(20), float64(0)}, mappoints[0])
// 		assert.Equal(t, []interface{}{float64(60), float64(30), float64(1)}, mappoints[1])
// 		assert.Equal(t, []interface{}{float64(40), float64(10), float64(1)}, mappoints[2])
// 	})

// 	t.Run("points by buyer", func(t *testing.T) {
// 		var reply jsonrpc.MapPointsReply
// 		err := svc.SessionMap(req, &jsonrpc.MapPointsArgs{BuyerID: buyerID1}, &reply)
// 		assert.NoError(t, err)

// 		var mappoints [][]interface{}
// 		err = json.Unmarshal(reply.Points, &mappoints)
// 		assert.NoError(t, err)

// 		assert.Equal(t, 2, len(mappoints))
// 		assert.Equal(t, []interface{}{float64(50), float64(20), float64(0)}, mappoints[0])
// 		assert.Equal(t, []interface{}{float64(60), float64(30), float64(1)}, mappoints[1])
// 	})
// }

func TestGameConfiguration(t *testing.T) {
	t.Parallel()

	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	storer := storage.InMemory{}
	pubkey := make([]byte, 4)
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "local.local.1", Domain: "local.com", PublicKey: pubkey})

	logger := log.NewNopLogger()
	svc := jsonrpc.BuyersService{
		RedisClient: redisClient,
		Storage:     &storer,
		Logger:      logger,
	}

	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"
	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Bearer "+jwtSideload)
	res := httptest.NewRecorder()

	authMiddleware.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	t.Run("missing name", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{Domain: "local.com"}, &reply)
		assert.Error(t, err)
	})

	t.Run("missing domain", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{Name: "local.local.1"}, &reply)
		assert.Error(t, err)
	})

	t.Run("single", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.GameConfiguration(req, &jsonrpc.GameConfigurationArgs{Domain: "local.com"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.GameConfiguration.PublicKey, "AQAAAAAAAAAAAAAA")
	})

	t.Run("failed to update public key", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{Domain: "local.com", Name: "local.local.1", NewPublicKey: "askjfgbdalksjdf balkjsdbf lkja flfakjs bdlkafs"}, &reply)

		assert.Error(t, err)

		assert.Equal(t, "", reply.GameConfiguration.PublicKey)
	})
}

/*
func TestSameBuyerRoleFunction(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	storer := storage.InMemory{}
	pubkey := make([]byte, 4)
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "local.local.1", Domain: "networknext.com", PublicKey: pubkey})

	logger := log.NewNopLogger()
	svc := jsonrpc.BuyersService{
		RedisClient: redisClient,
		Storage:     &storer,
		Logger:      logger,
	}

	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"
	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Bearer "+jwtSideload)
	res := httptest.NewRecorder()

	authMiddleware.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	t.Run("samebuyer role function", func(t *testing.T) {
		sameBuyerRoleFunc := svc.SameBuyerRole("1")
		verified, err := sameBuyerRoleFunc(req)
		assert.NoError(t, err)
		assert.True(t, verified)
	})
} */
