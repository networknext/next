package jsonrpc_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
	"gopkg.in/auth0.v4/management"
)

func TestBuyersList(t *testing.T) {
	storer := storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "local.local.1"})

	svc := jsonrpc.BuyersService{
		Storage: &storer,
	}
	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6ImpvaG4iLCJuYW1lIjoiam9obkBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMGIzZTgwMDFjYTJkN2NlM2I2ZmZlMTU2ZTczODRlZTU_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZqby5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNS0xOVQxOTo1MDoyMC44NjNaIiwiZW1haWwiOiJqb2huQG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1ZWJhYzhiMjA3ZWU4YjFjMTliNGMwZTIiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU4OTkxNzgyMiwiZXhwIjoxNzQ3NTk3ODIyLCJub25jZSI6IlJuRjFaVzlYYW1aS2VUTkVPRFJMTFhreVVVNVRielJSWVdjdVRGWjFUVlpFZDFVellYNDBXR05sTUE9PSJ9.Va2WRHDUj7XoXzvSkUDfx819RDpewyHMxyv0CIBfsWfVOCB80jRPBvQo7oImRM0FPMYyCl5r4i8-rU5jyg8fZUC3vSABVPALqxX4ViNy3qB4Zgn1RidXoUGKuAUTfi40fS_xDSDBoErRjkxzZuMby_9xNhBw5WwL6sKDGzGL-nayBWHf7LTf0wPwrhZPI4YtHdrJEzYUkwdMCJnMsuSZsgpwvfzvpLgg9NJ4me-VhTQAKJjxXIAsHD_QiI7EEPK1tcd58T11J_xsTktSmDVxuG0-QIs2ioWs0DJSepjcld4tLTlDDZObHIjo_edXd5Wk9zalxfAE7sPWUexFZPQMDA"
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

		assert.Equal(t, "1", reply.Buyers[0].ID)
		assert.Equal(t, "local.local.1", reply.Buyers[0].Name)
	})
}

func TestUserSessions(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	userHash1 := fmt.Sprintf("%x", 111)
	userHash2 := fmt.Sprintf("%x", 222)

	sessionID1 := fmt.Sprintf("%x", 111)
	sessionID2 := fmt.Sprintf("%x", 222)
	sessionID3 := fmt.Sprintf("%x", 333)
	sessionID4 := "missing"

	redisServer.SetAdd(fmt.Sprintf("user-%s-sessions", userHash2), sessionID1)
	redisServer.SetAdd(fmt.Sprintf("user-%s-sessions", userHash1), sessionID2)
	redisServer.SetAdd(fmt.Sprintf("user-%s-sessions", userHash1), sessionID3)
	redisServer.SetAdd(fmt.Sprintf("user-%s-sessions", userHash1), sessionID4)

	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID1), routing.SessionMeta{ID: sessionID1}, time.Hour)
	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID2), routing.SessionMeta{ID: sessionID2}, time.Hour)
	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID3), routing.SessionMeta{ID: sessionID3}, time.Hour)

	svc := jsonrpc.BuyersService{
		RedisClient: redisClient,
	}
	logger := log.NewNopLogger()

	manager, err := management.New(
		"networknext.auth0.com",
		"0Hn8oZfUwy5UPo6bUk0hYCQ2hMJnwQYg",
		"l2namTU5jKVAkuCwV3votIPcP87jcOuJREtscx07aLgo8EykReX69StUVBfJOzx5",
	)
	assert.NoError(t, err)

	auth0Client := storage.Auth0{
		Manager: manager,
		Logger:  logger,
	}

	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6ImpvaG4iLCJuYW1lIjoiam9obkBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMGIzZTgwMDFjYTJkN2NlM2I2ZmZlMTU2ZTczODRlZTU_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZqby5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNS0xOVQxOTo1MDoyMC44NjNaIiwiZW1haWwiOiJqb2huQG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1ZWJhYzhiMjA3ZWU4YjFjMTliNGMwZTIiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU4OTkxNzgyMiwiZXhwIjoxNzQ3NTk3ODIyLCJub25jZSI6IlJuRjFaVzlYYW1aS2VUTkVPRFJMTFhreVVVNVRielJSWVdjdVRGWjFUVlpFZDFVellYNDBXR05sTUE9PSJ9.Va2WRHDUj7XoXzvSkUDfx819RDpewyHMxyv0CIBfsWfVOCB80jRPBvQo7oImRM0FPMYyCl5r4i8-rU5jyg8fZUC3vSABVPALqxX4ViNy3qB4Zgn1RidXoUGKuAUTfi40fS_xDSDBoErRjkxzZuMby_9xNhBw5WwL6sKDGzGL-nayBWHf7LTf0wPwrhZPI4YtHdrJEzYUkwdMCJnMsuSZsgpwvfzvpLgg9NJ4me-VhTQAKJjxXIAsHD_QiI7EEPK1tcd58T11J_xsTktSmDVxuG0-QIs2ioWs0DJSepjcld4tLTlDDZObHIjo_edXd5Wk9zalxfAE7sPWUexFZPQMDA"
	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Bearer "+jwtSideload)
	res := httptest.NewRecorder()

	authMiddleware.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	user := req.Context().Value("user")
	assert.NotEqual(t, user, nil)
	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

	requestID, ok := claims["sub"]

	assert.True(t, ok)
	assert.Equal(t, "auth0|5ebac8b207ee8b1c19b4c0e2", requestID)

	roles, err := auth0Client.Manager.User.Roles(requestID.(string))

	assert.NoError(t, err)
	req = jsonrpc.SetRoles(req, *roles)

	t.Run("missing user_hash", func(t *testing.T) {
		var reply jsonrpc.UserSessionsReply
		err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(reply.Sessions))
	})

	t.Run("user_hash not found", func(t *testing.T) {
		var reply jsonrpc.UserSessionsReply
		err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserHash: "12345"}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(reply.Sessions))
	})

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.UserSessionsReply
		err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserHash: userHash1}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Sessions), 2)

		assert.Equal(t, reply.Sessions[0].ID, sessionID3)
		assert.Equal(t, reply.Sessions[1].ID, sessionID2)
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

	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6ImpvaG4iLCJuYW1lIjoiam9obkBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMGIzZTgwMDFjYTJkN2NlM2I2ZmZlMTU2ZTczODRlZTU_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZqby5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNS0xOVQxOTo1MDoyMC44NjNaIiwiZW1haWwiOiJqb2huQG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1ZWJhYzhiMjA3ZWU4YjFjMTliNGMwZTIiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU4OTkxNzgyMiwiZXhwIjoxNzQ3NTk3ODIyLCJub25jZSI6IlJuRjFaVzlYYW1aS2VUTkVPRFJMTFhreVVVNVRielJSWVdjdVRGWjFUVlpFZDFVellYNDBXR05sTUE9PSJ9.Va2WRHDUj7XoXzvSkUDfx819RDpewyHMxyv0CIBfsWfVOCB80jRPBvQo7oImRM0FPMYyCl5r4i8-rU5jyg8fZUC3vSABVPALqxX4ViNy3qB4Zgn1RidXoUGKuAUTfi40fS_xDSDBoErRjkxzZuMby_9xNhBw5WwL6sKDGzGL-nayBWHf7LTf0wPwrhZPI4YtHdrJEzYUkwdMCJnMsuSZsgpwvfzvpLgg9NJ4me-VhTQAKJjxXIAsHD_QiI7EEPK1tcd58T11J_xsTktSmDVxuG0-QIs2ioWs0DJSepjcld4tLTlDDZObHIjo_edXd5Wk9zalxfAE7sPWUexFZPQMDA"
	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Bearer "+jwtSideload)
	res := httptest.NewRecorder()

	authMiddleware.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	t.Run("top global", func(t *testing.T) {
		var reply jsonrpc.TopSessionsReply
		err := svc.TopSessions(req, &jsonrpc.TopSessionsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(reply.Sessions))
		assert.Equal(t, sessionID1, reply.Sessions[0].ID)
		assert.Equal(t, sessionID2, reply.Sessions[1].ID)
		assert.Equal(t, sessionID3, reply.Sessions[2].ID)

		assert.Greater(t, int(redisClient.TTL("session-missing-meta").Val()), 0)
	})

	t.Run("top buyer", func(t *testing.T) {
		var reply jsonrpc.TopSessionsReply
		err := svc.TopSessions(req, &jsonrpc.TopSessionsArgs{BuyerID: buyerID1}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(reply.Sessions))
		assert.Equal(t, sessionID2, reply.Sessions[0].ID)
		assert.Equal(t, sessionID3, reply.Sessions[1].ID)

		assert.Greater(t, int(redisClient.TTL("session-missing-meta").Val()), 0)
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
		Hops: []routing.Relay{
			{ID: 1234},
			{ID: 1234},
			{ID: 1234},
		},
		SDK: "3.4.4",
		NearbyRelays: []routing.Relay{
			{ID: 1, ClientStats: routing.Stats{RTT: 1, Jitter: 2, PacketLoss: 3}},
		},
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

	redisClient.Set(fmt.Sprintf("session-%s-meta", sessionID), meta, 30*time.Second)
	redisClient.SAdd(fmt.Sprintf("session-%s-slices", sessionID), slice1, slice2)

	// After setting the cache without the name, set the name to the expected output we need
	meta.NearbyRelays[0].Name = "local"

	inMemory := storage.InMemory{}
	inMemory.AddSeller(context.Background(), routing.Seller{ID: "local"})
	inMemory.AddDatacenter(context.Background(), routing.Datacenter{ID: 1})
	inMemory.AddRelay(context.Background(), routing.Relay{ID: 1, Name: "local", Seller: routing.Seller{ID: "local"}, Datacenter: routing.Datacenter{ID: 1}})

	svc := jsonrpc.BuyersService{
		RedisClient: redisClient,
		Storage:     &inMemory,
	}
	logger := log.NewNopLogger()

	manager, err := management.New(
		"networknext.auth0.com",
		"0Hn8oZfUwy5UPo6bUk0hYCQ2hMJnwQYg",
		"l2namTU5jKVAkuCwV3votIPcP87jcOuJREtscx07aLgo8EykReX69StUVBfJOzx5",
	)
	assert.NoError(t, err)

	auth0Client := storage.Auth0{
		Manager: manager,
		Logger:  logger,
	}

	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6ImpvaG4iLCJuYW1lIjoiam9obkBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMGIzZTgwMDFjYTJkN2NlM2I2ZmZlMTU2ZTczODRlZTU_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZqby5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNS0xOVQxOTo1MDoyMC44NjNaIiwiZW1haWwiOiJqb2huQG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1ZWJhYzhiMjA3ZWU4YjFjMTliNGMwZTIiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU4OTkxNzgyMiwiZXhwIjoxNzQ3NTk3ODIyLCJub25jZSI6IlJuRjFaVzlYYW1aS2VUTkVPRFJMTFhreVVVNVRielJSWVdjdVRGWjFUVlpFZDFVellYNDBXR05sTUE9PSJ9.Va2WRHDUj7XoXzvSkUDfx819RDpewyHMxyv0CIBfsWfVOCB80jRPBvQo7oImRM0FPMYyCl5r4i8-rU5jyg8fZUC3vSABVPALqxX4ViNy3qB4Zgn1RidXoUGKuAUTfi40fS_xDSDBoErRjkxzZuMby_9xNhBw5WwL6sKDGzGL-nayBWHf7LTf0wPwrhZPI4YtHdrJEzYUkwdMCJnMsuSZsgpwvfzvpLgg9NJ4me-VhTQAKJjxXIAsHD_QiI7EEPK1tcd58T11J_xsTktSmDVxuG0-QIs2ioWs0DJSepjcld4tLTlDDZObHIjo_edXd5Wk9zalxfAE7sPWUexFZPQMDA"
	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Bearer "+jwtSideload)
	res := httptest.NewRecorder()

	authMiddleware.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	user := req.Context().Value("user")
	assert.NotEqual(t, user, nil)
	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

	requestID, ok := claims["sub"]

	assert.True(t, ok)
	assert.Equal(t, "auth0|5ebac8b207ee8b1c19b4c0e2", requestID)

	roles, err := auth0Client.Manager.User.Roles(requestID.(string))

	assert.NoError(t, err)
	req = jsonrpc.SetRoles(req, *roles)

	t.Run("session_id not found", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: "nope"}, &reply)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
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
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	buyerID1 := fmt.Sprintf("%x", 111)
	buyerID2 := fmt.Sprintf("%x", 222)

	sessionID1 := fmt.Sprintf("%x", 111)
	sessionID2 := fmt.Sprintf("%x", 222)
	sessionID3 := fmt.Sprintf("%x", 333)
	sessionID4 := "missing"

	redisServer.SetAdd("map-points-global", sessionID1)
	redisServer.SetAdd("map-points-global", sessionID2)
	redisServer.SetAdd("map-points-global", sessionID3)
	redisServer.SetAdd("map-points-global", sessionID4)

	redisServer.SetAdd(fmt.Sprintf("map-points-buyer-%s", buyerID2), sessionID1)
	redisServer.SetAdd(fmt.Sprintf("map-points-buyer-%s", buyerID1), sessionID2)
	redisServer.SetAdd(fmt.Sprintf("map-points-buyer-%s", buyerID1), sessionID3)
	redisServer.SetAdd(fmt.Sprintf("map-points-buyer-%s", buyerID1), sessionID4)

	points := []routing.SessionMapPoint{
		{Latitude: 10, Longitude: 40, OnNetworkNext: true},
		{Latitude: 20, Longitude: 50, OnNetworkNext: false},
		{Latitude: 30, Longitude: 60, OnNetworkNext: true},
	}

	redisClient.Set(fmt.Sprintf("session-%s-point", sessionID1), points[0], time.Hour)
	redisClient.Set(fmt.Sprintf("session-%s-point", sessionID2), points[1], time.Hour)
	redisClient.Set(fmt.Sprintf("session-%s-point", sessionID3), points[2], time.Hour)

	svc := jsonrpc.BuyersService{
		RedisClient: redisClient,
	}

	logger := log.NewNopLogger()

	manager, err := management.New(
		"networknext.auth0.com",
		"0Hn8oZfUwy5UPo6bUk0hYCQ2hMJnwQYg",
		"l2namTU5jKVAkuCwV3votIPcP87jcOuJREtscx07aLgo8EykReX69StUVBfJOzx5",
	)
	assert.NoError(t, err)

	auth0Client := storage.Auth0{
		Manager: manager,
		Logger:  logger,
	}

	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6ImpvaG4iLCJuYW1lIjoiam9obkBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMGIzZTgwMDFjYTJkN2NlM2I2ZmZlMTU2ZTczODRlZTU_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZqby5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNS0xOVQxOTo1MDoyMC44NjNaIiwiZW1haWwiOiJqb2huQG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1ZWJhYzhiMjA3ZWU4YjFjMTliNGMwZTIiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU4OTkxNzgyMiwiZXhwIjoxNzQ3NTk3ODIyLCJub25jZSI6IlJuRjFaVzlYYW1aS2VUTkVPRFJMTFhreVVVNVRielJSWVdjdVRGWjFUVlpFZDFVellYNDBXR05sTUE9PSJ9.Va2WRHDUj7XoXzvSkUDfx819RDpewyHMxyv0CIBfsWfVOCB80jRPBvQo7oImRM0FPMYyCl5r4i8-rU5jyg8fZUC3vSABVPALqxX4ViNy3qB4Zgn1RidXoUGKuAUTfi40fS_xDSDBoErRjkxzZuMby_9xNhBw5WwL6sKDGzGL-nayBWHf7LTf0wPwrhZPI4YtHdrJEzYUkwdMCJnMsuSZsgpwvfzvpLgg9NJ4me-VhTQAKJjxXIAsHD_QiI7EEPK1tcd58T11J_xsTktSmDVxuG0-QIs2ioWs0DJSepjcld4tLTlDDZObHIjo_edXd5Wk9zalxfAE7sPWUexFZPQMDA"
	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Bearer "+jwtSideload)
	res := httptest.NewRecorder()

	authMiddleware.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	user := req.Context().Value("user")
	assert.NotEqual(t, user, nil)
	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

	requestID, ok := claims["sub"]

	assert.True(t, ok)
	assert.Equal(t, "auth0|5ebac8b207ee8b1c19b4c0e2", requestID)

	roles, err := auth0Client.Manager.User.Roles(requestID.(string))

	assert.NoError(t, err)
	req = jsonrpc.SetRoles(req, *roles)

	t.Run("points global", func(t *testing.T) {
		var reply jsonrpc.MapPointsReply
		err := svc.SessionMapPoints(req, &jsonrpc.MapPointsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(reply.Points))
		assert.Contains(t, reply.Points, points[0])
		assert.Contains(t, reply.Points, points[1])
		assert.Contains(t, reply.Points, points[2])

		assert.Greater(t, int(redisClient.TTL("session-missing-point").Val()), 0)
	})

	t.Run("points by buyer", func(t *testing.T) {
		var reply jsonrpc.MapPointsReply
		err := svc.SessionMapPoints(req, &jsonrpc.MapPointsArgs{BuyerID: buyerID1}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(reply.Points))
		assert.NotContains(t, reply.Points, points[0])
		assert.Contains(t, reply.Points, points[1])
		assert.Contains(t, reply.Points, points[2])

		assert.Greater(t, int(redisClient.TTL("session-missing-point").Val()), 0)
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

	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6ImpvaG4iLCJuYW1lIjoiam9obkBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMGIzZTgwMDFjYTJkN2NlM2I2ZmZlMTU2ZTczODRlZTU_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZqby5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNS0xOVQxOTo1MDoyMC44NjNaIiwiZW1haWwiOiJqb2huQG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1ZWJhYzhiMjA3ZWU4YjFjMTliNGMwZTIiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU4OTkxNzgyMiwiZXhwIjoxNzQ3NTk3ODIyLCJub25jZSI6IlJuRjFaVzlYYW1aS2VUTkVPRFJMTFhreVVVNVRielJSWVdjdVRGWjFUVlpFZDFVellYNDBXR05sTUE9PSJ9.Va2WRHDUj7XoXzvSkUDfx819RDpewyHMxyv0CIBfsWfVOCB80jRPBvQo7oImRM0FPMYyCl5r4i8-rU5jyg8fZUC3vSABVPALqxX4ViNy3qB4Zgn1RidXoUGKuAUTfi40fS_xDSDBoErRjkxzZuMby_9xNhBw5WwL6sKDGzGL-nayBWHf7LTf0wPwrhZPI4YtHdrJEzYUkwdMCJnMsuSZsgpwvfzvpLgg9NJ4me-VhTQAKJjxXIAsHD_QiI7EEPK1tcd58T11J_xsTktSmDVxuG0-QIs2ioWs0DJSepjcld4tLTlDDZObHIjo_edXd5Wk9zalxfAE7sPWUexFZPQMDA"
	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Bearer "+jwtSideload)
	res := httptest.NewRecorder()

	authMiddleware.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	t.Run("missing buyer_id", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.GameConfiguration(req, &jsonrpc.GameConfigurationArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failed to convert buyer_id to uint64", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.GameConfiguration(req, &jsonrpc.GameConfigurationArgs{BuyerID: "asdgagasdgfa"}, &reply)
		assert.Error(t, err)
	})

	t.Run("single", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.GameConfiguration(req, &jsonrpc.GameConfigurationArgs{BuyerID: "1"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.GameConfiguration.PublicKey, "AAAAAA==")
	})

	t.Run("update public key", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{BuyerID: "1", NewPublicKey: "iVBORw0KGgoAAAANSUhEUgAAAGQAAABkCAYA"}, &reply)

		assert.NoError(t, err)

		assert.Equal(t, reply.GameConfiguration.PublicKey, "iVBORw0KGgoAAAANSUhEUgAAAGQAAABkCAYA")
	})

	t.Run("failed to update public key", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{BuyerID: "1", NewPublicKey: "askjfgbdalksjdf balkjsdbf lkja flfakjs bdlkafs"}, &reply)

		assert.Error(t, err)
	})
}
