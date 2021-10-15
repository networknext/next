package jsonrpc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/networknext/backend/modules/transport/middleware"
	"github.com/stretchr/testify/assert"
)

func checkBigtableEmulation(t *testing.T) {
	bigtableEmulatorHost := os.Getenv("BIGTABLE_EMULATOR_HOST")
	if bigtableEmulatorHost == "" {
		t.Skip("Bigtable emulator not set up, skipping bigtable test")
	}
}

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
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.BuyerListReply
		err := svc.Buyers(req, &jsonrpc.BuyerListArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(reply.Buyers))
	})

	t.Run("list - !admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
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
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		assert.True(t, middleware.VerifyAllRoles(req, middleware.AdminRole))
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

func TestUserSessions_Binary(t *testing.T) {
	checkBigtableEmulation(t)

	var storer = storage.InMemory{}

	buyer0 := routing.Buyer{
		ID:          0,
		CompanyCode: "local0",
	}
	buyer7 := routing.Buyer{
		ID:          777,
		CompanyCode: "local7",
	}
	buyer8 := routing.Buyer{
		ID:          888,
		CompanyCode: "local8",
	}
	buyer9 := routing.Buyer{
		ID:          999,
		CompanyCode: "local9",
	}

	customer0 := routing.Customer{
		Code: "local0",
		Name: "Local0",
	}
	customer7 := routing.Customer{
		Code: "local7",
		Name: "Local7",
	}
	customer8 := routing.Customer{
		Code: "local8",
		Name: "Local8",
	}
	customer9 := routing.Customer{
		Code: "local9",
		Name: "Local9",
	}

	ctx := context.Background()
	storer.AddBuyer(ctx, buyer0)
	storer.AddBuyer(ctx, buyer7)
	storer.AddBuyer(ctx, buyer8)
	storer.AddBuyer(ctx, buyer9)
	storer.AddCustomer(ctx, customer0)
	storer.AddCustomer(ctx, customer7)
	storer.AddCustomer(ctx, customer8)
	storer.AddCustomer(ctx, customer9)

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 5, 5)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	rawUserID1 := 111
	rawUserID2 := 222
	userID1 := fmt.Sprintf("%d", rawUserID1)
	userID2 := fmt.Sprintf("%d", rawUserID2)

	hash1 := fnv.New64a()
	_, err := hash1.Write([]byte(userID1))
	assert.Nil(t, err)
	userHash1 := hash1.Sum64()

	hash2 := fnv.New64a()
	_, err = hash2.Write([]byte(userID2))
	assert.Nil(t, err)
	userHash2 := hash2.Sum64()
	userHash3 := uint64(8353685330869585599) // Signed decimal hash

	sessionID1 := fmt.Sprintf("%016x", 111)
	sessionID2 := fmt.Sprintf("%016x", 222)
	sessionID3 := fmt.Sprintf("%016x", 333)
	sessionID4 := fmt.Sprintf("%016x", 444)
	sessionID5 := "missing"

	now := time.Now()
	secs := now.Unix()
	minutes := secs / 60

	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 50, sessionID1)
	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 100, sessionID2)
	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 150, sessionID3)
	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 150, sessionID4)

	redisServer.ZAdd(fmt.Sprintf("sc-%016x-%d", userHash2, minutes), 50, sessionID1)
	redisServer.ZAdd(fmt.Sprintf("sc-%016x-%d", userHash1, minutes), 100, sessionID2)
	redisServer.ZAdd(fmt.Sprintf("sc-%016x-%d", userHash1, minutes), 150, sessionID3)
	redisServer.ZAdd(fmt.Sprintf("sc-%016x-%d", userHash1, minutes), 150, sessionID5)
	redisServer.ZAdd(fmt.Sprintf("sc-%016x-%d", userHash3, minutes), 150, sessionID4)

	redisClient.Set(fmt.Sprintf("sm-%s", sessionID1), transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 111, UserHash: userHash2, DeltaRTT: 50}.RedisString(), time.Hour)
	redisClient.Set(fmt.Sprintf("sm-%s", sessionID2), transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 222, UserHash: userHash1, DeltaRTT: 100}.RedisString(), time.Hour)
	redisClient.Set(fmt.Sprintf("sm-%s", sessionID3), transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 333, UserHash: userHash1, DeltaRTT: 150}.RedisString(), time.Hour)
	redisClient.Set(fmt.Sprintf("sm-%s", sessionID4), transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 444, UserHash: userHash3, DeltaRTT: 150}.RedisString(), time.Hour)

	logger := log.NewNopLogger()

	btTableName, btTableEnvVarOK := os.LookupEnv("BIGTABLE_TABLE_NAME")
	if !btTableEnvVarOK {
		btTableName = "Test"
		os.Setenv("BIGTABLE_TABLE_NAME", btTableName)
		defer os.Unsetenv("BIGTABLE_TABLE_NAME")
	}

	// Get the column family name
	btCfName, btCfNameEnvVarOK := os.LookupEnv("BIGTABLE_CF_NAME")
	if !btCfNameEnvVarOK {
		btCfName = "TestCfName"
		os.Setenv("BIGTABLE_CF_NAME", btCfName)
		defer os.Unsetenv("BIGTABLE_CF_NAME")
	}

	// Check if table exists and create it if needed
	btAdmin, err := storage.NewBigTableAdmin(ctx, "local", "localhost:8086", logger)
	assert.NoError(t, err)
	assert.NotNil(t, btAdmin)
	btTableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	assert.NoError(t, err)
	if !btTableExists {
		// Create a table
		btAdmin.CreateTable(ctx, btTableName, []string{btCfName})
	}

	defer func() {
		err := btAdmin.DeleteTable(ctx, btTableName)
		assert.NoError(t, err)

		err = btAdmin.Close()
		assert.NoError(t, err)
	}()

	btClient, err := storage.NewBigTable(ctx, "local", "localhost:8086", btTableName, logger)
	assert.Nil(t, err)
	assert.NotNil(t, btClient)

	defer func() {
		err := btClient.Close()
		assert.NoError(t, err)
	}()

	// Add user sessions to bigtable
	metaBin1, err := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 111, UserHash: userHash2, BuyerID: 999}.MarshalBinary()
	assert.NoError(t, err)
	slice1 := transport.SessionSlice{Version: transport.SessionSliceVersion}
	sliceBin1, err := slice1.MarshalBinary()
	assert.NoError(t, err)
	sessionRowKey1 := sessionID1
	sliceRowKey1 := fmt.Sprintf("%s#%v", sessionID1, slice1.Timestamp)
	userRowKey1 := fmt.Sprintf("%016x#%s", userHash2, sessionID1)
	metaRowKeys1 := []string{sessionRowKey1, userRowKey1}
	sliceRowKeys1 := []string{sliceRowKey1}

	metaBin2, err := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 222, UserHash: userHash1, BuyerID: 888}.MarshalBinary()
	assert.NoError(t, err)
	slice2 := transport.SessionSlice{Version: transport.SessionSliceVersion}
	sliceBin2, err := slice2.MarshalBinary()
	assert.NoError(t, err)
	sessionRowKey2 := sessionID2
	sliceRowKey2 := fmt.Sprintf("%s#%v", sessionID2, slice2.Timestamp)
	userRowKey2 := fmt.Sprintf("%016x#%s", userHash1, sessionID2)
	metaRowKeys2 := []string{sessionRowKey2, userRowKey2}
	sliceRowKeys2 := []string{sliceRowKey2}

	metaBin3, err := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 333, UserHash: userHash1, BuyerID: 888}.MarshalBinary()
	assert.NoError(t, err)
	slice3 := transport.SessionSlice{Version: transport.SessionSliceVersion}
	sliceBin3, err := slice3.MarshalBinary()
	assert.NoError(t, err)
	sessionRowKey3 := sessionID3
	sliceRowKey3 := fmt.Sprintf("%s#%v", sessionID3, slice3.Timestamp)
	userRowKey3 := fmt.Sprintf("%016x#%s", userHash1, sessionID3)
	metaRowKeys3 := []string{sessionRowKey3, userRowKey3}
	sliceRowKeys3 := []string{sliceRowKey3}

	metaBin4, err := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 444, UserHash: userHash3, BuyerID: 777}.MarshalBinary()
	assert.NoError(t, err)
	slice4 := transport.SessionSlice{Version: transport.SessionSliceVersion}
	sliceBin4, err := slice4.MarshalBinary()
	assert.NoError(t, err)
	sessionRowKey4 := sessionID4
	sliceRowKey4 := fmt.Sprintf("%s#%v", sessionID4, slice4.Timestamp)
	userRowKey4 := fmt.Sprintf("%016x#%s", userHash3, sessionID4)
	metaRowKeys4 := []string{sessionRowKey4, userRowKey4}
	sliceRowKeys4 := []string{sliceRowKey4}

	err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin1, metaRowKeys1)
	assert.NoError(t, err)
	err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin1, sliceRowKeys1)
	assert.NoError(t, err)
	err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin2, metaRowKeys2)
	assert.NoError(t, err)
	err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin2, sliceRowKeys2)
	assert.NoError(t, err)
	err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin3, metaRowKeys3)
	assert.NoError(t, err)
	err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin3, sliceRowKeys3)
	assert.NoError(t, err)
	err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin4, metaRowKeys4)
	assert.NoError(t, err)
	err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin4, sliceRowKeys4)
	assert.NoError(t, err)

	redisClient.RPush(fmt.Sprintf("ss-%s", sessionID1), slice1.RedisString())
	redisClient.RPush(fmt.Sprintf("ss-%s", sessionID2), slice2.RedisString())
	redisClient.RPush(fmt.Sprintf("ss-%s", sessionID3), slice3.RedisString())
	redisClient.RPush(fmt.Sprintf("ss-%s", sessionID4), slice4.RedisString())

	svc := jsonrpc.BuyersService{
		Storage:                &storer,
		UseBigtable:            true,
		BigTableCfName:         btCfName,
		BigTable:               btClient,
		BigTableMetrics:        &metrics.EmptyBigTableMetrics,
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
		assert.EqualError(t, err, "UserSessions() user id is required")
		assert.Equal(t, 0, len(reply.Sessions))
	})

	t.Run("user_hash not found", func(t *testing.T) {
		var reply jsonrpc.UserSessionsReply
		err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: "12345"}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(reply.Sessions))
	})

	t.Run("Live Sessions", func(t *testing.T) {

		t.Run("list live - ID", func(t *testing.T) {
			var reply jsonrpc.UserSessionsReply
			err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: userID1}, &reply)
			assert.NoError(t, err)

			assert.Equal(t, 2, len(reply.Sessions))

			for _, session := range reply.Sessions {
				idString := fmt.Sprintf("%016x", session.Meta.ID)
				if idString != sessionID3 && idString != sessionID2 {
					t.Fail()
				}
			}
		})

		t.Run("list live - hash", func(t *testing.T) {
			var reply jsonrpc.UserSessionsReply
			err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: fmt.Sprintf("%016x", userHash1)}, &reply)
			assert.NoError(t, err)

			assert.Equal(t, 2, len(reply.Sessions))

			for _, session := range reply.Sessions {
				idString := fmt.Sprintf("%016x", session.Meta.ID)
				if idString != sessionID3 && idString != sessionID2 {
					t.Fail()
				}
			}
		})

		t.Run("list live - signed decimal hash", func(t *testing.T) {
			var reply jsonrpc.UserSessionsReply
			err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: fmt.Sprintf("%016x", userHash3)}, &reply)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(reply.Sessions))

			assert.Equal(t, fmt.Sprintf("%016x", reply.Sessions[0].Meta.ID), sessionID4)
		})
	})

	t.Run("Historic and Live Sessions", func(t *testing.T) {
		// Insert additional historic sessions
		sessionID6 := fmt.Sprintf("%016x", 666)
		sessionID7 := fmt.Sprintf("%016x", 777)
		sessionID8 := fmt.Sprintf("%016x", 888)

		metaBin6, err := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 666, UserHash: userHash2, BuyerID: 999}.MarshalBinary()
		assert.NoError(t, err)
		slice6 := transport.SessionSlice{Version: transport.SessionSliceVersion}
		sliceBin6, err := slice6.MarshalBinary()
		assert.NoError(t, err)
		sessionRowKey6 := sessionID6
		sliceRowKey6 := fmt.Sprintf("%s#%v", sessionID6, slice6.Timestamp)
		userRowKey6 := fmt.Sprintf("%016x#%s", userHash2, sessionID6)
		metaRowKeys6 := []string{sessionRowKey6, userRowKey6}
		sliceRowKeys6 := []string{sliceRowKey6}

		metaBin7, err := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 777, UserHash: userHash1, BuyerID: 888}.MarshalBinary()
		assert.NoError(t, err)
		slice7 := transport.SessionSlice{Version: transport.SessionSliceVersion}
		sliceBin7, err := slice7.MarshalBinary()
		assert.NoError(t, err)
		sessionRowKey7 := sessionID7
		sliceRowKey7 := fmt.Sprintf("%s#%v", sessionID7, slice7.Timestamp)
		userRowKey7 := fmt.Sprintf("%016x#%s", userHash1, sessionID7)
		metaRowKeys7 := []string{sessionRowKey7, userRowKey7}
		sliceRowKeys7 := []string{sliceRowKey7}

		metaBin8, err := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 888, UserHash: userHash1, BuyerID: 888}.MarshalBinary()
		assert.NoError(t, err)
		slice8 := transport.SessionSlice{Version: transport.SessionSliceVersion}
		sliceBin8, err := slice8.MarshalBinary()
		assert.NoError(t, err)
		sessionRowKey8 := sessionID8
		sliceRowKey8 := fmt.Sprintf("%s#%v", sessionID8, slice8.Timestamp)
		userRowKey8 := fmt.Sprintf("%016x#%s", userHash3, sessionID8)
		metaRowKeys8 := []string{sessionRowKey8, userRowKey8}
		sliceRowKeys8 := []string{sliceRowKey8}

		err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin6, metaRowKeys6)
		assert.NoError(t, err)
		err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin6, sliceRowKeys6)
		assert.NoError(t, err)
		err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin7, metaRowKeys7)
		assert.NoError(t, err)
		err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin7, sliceRowKeys7)
		assert.NoError(t, err)
		err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin8, metaRowKeys8)
		assert.NoError(t, err)
		err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin8, sliceRowKeys8)
		assert.NoError(t, err)

		t.Run("list live and historic - ID", func(t *testing.T) {
			var reply jsonrpc.UserSessionsReply
			err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: userID1}, &reply)
			assert.NoError(t, err)

			assert.Equal(t, 3, len(reply.Sessions))

			for _, session := range reply.Sessions {
				idString := fmt.Sprintf("%016x", session.Meta.ID)
				if idString != sessionID3 && idString != sessionID2 && idString != sessionID7 {
					t.Fail()
				}
			}
		})

		t.Run("list live and historic - hash", func(t *testing.T) {
			var reply jsonrpc.UserSessionsReply
			err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: fmt.Sprintf("%016x", userHash1)}, &reply)
			assert.NoError(t, err)

			assert.Equal(t, 3, len(reply.Sessions))
			for _, session := range reply.Sessions {
				idString := fmt.Sprintf("%016x", session.Meta.ID)
				if idString != sessionID3 && idString != sessionID2 && idString != sessionID7 {
					t.Fail()
				}
			}
		})

		t.Run("list live and historic - signed decimal hash", func(t *testing.T) {
			var reply jsonrpc.UserSessionsReply
			err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: fmt.Sprintf("%016x", userHash3)}, &reply)
			assert.NoError(t, err)

			assert.Equal(t, 2, len(reply.Sessions))
			for _, session := range reply.Sessions {
				idString := fmt.Sprintf("%016x", session.Meta.ID)
				if idString != sessionID4 && idString != sessionID8 {
					t.Fail()
				}
			}
		})
	})
}

func TestUserSessions_Serialize(t *testing.T) {
	checkBigtableEmulation(t)

	var storer = storage.InMemory{}

	buyer0 := routing.Buyer{
		ID:          0,
		CompanyCode: "local0",
	}
	buyer7 := routing.Buyer{
		ID:          777,
		CompanyCode: "local7",
	}
	buyer8 := routing.Buyer{
		ID:          888,
		CompanyCode: "local8",
	}
	buyer9 := routing.Buyer{
		ID:          999,
		CompanyCode: "local9",
	}

	customer0 := routing.Customer{
		Code: "local0",
		Name: "Local0",
	}
	customer7 := routing.Customer{
		Code: "local7",
		Name: "Local7",
	}
	customer8 := routing.Customer{
		Code: "local8",
		Name: "Local8",
	}
	customer9 := routing.Customer{
		Code: "local9",
		Name: "Local9",
	}

	ctx := context.Background()
	storer.AddBuyer(ctx, buyer0)
	storer.AddBuyer(ctx, buyer7)
	storer.AddBuyer(ctx, buyer8)
	storer.AddBuyer(ctx, buyer9)
	storer.AddCustomer(ctx, customer0)
	storer.AddCustomer(ctx, customer7)
	storer.AddCustomer(ctx, customer8)
	storer.AddCustomer(ctx, customer9)

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 5, 5)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	rawUserID1 := 111
	rawUserID2 := 222
	userID1 := fmt.Sprintf("%d", rawUserID1)
	userID2 := fmt.Sprintf("%d", rawUserID2)

	hash1 := fnv.New64a()
	_, err := hash1.Write([]byte(userID1))
	assert.Nil(t, err)
	userHash1 := hash1.Sum64()

	hash2 := fnv.New64a()
	_, err = hash2.Write([]byte(userID2))
	assert.Nil(t, err)
	userHash2 := hash2.Sum64()
	userHash3 := uint64(8353685330869585599) // Signed decimal hash

	sessionID1 := fmt.Sprintf("%016x", 111)
	sessionID2 := fmt.Sprintf("%016x", 222)
	sessionID3 := fmt.Sprintf("%016x", 333)
	sessionID4 := fmt.Sprintf("%016x", 444)
	sessionID5 := "missing"

	now := time.Now()
	secs := now.Unix()
	minutes := secs / 60

	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 50, sessionID1)
	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 100, sessionID2)
	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 150, sessionID3)
	redisServer.ZAdd(fmt.Sprintf("s-%d", minutes), 150, sessionID4)

	redisServer.ZAdd(fmt.Sprintf("sc-%016x-%d", userHash2, minutes), 50, sessionID1)
	redisServer.ZAdd(fmt.Sprintf("sc-%016x-%d", userHash1, minutes), 100, sessionID2)
	redisServer.ZAdd(fmt.Sprintf("sc-%016x-%d", userHash1, minutes), 150, sessionID3)
	redisServer.ZAdd(fmt.Sprintf("sc-%016x-%d", userHash1, minutes), 150, sessionID5)
	redisServer.ZAdd(fmt.Sprintf("sc-%016x-%d", userHash3, minutes), 150, sessionID4)

	redisClient.Set(fmt.Sprintf("sm-%s", sessionID1), transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 111, UserHash: userHash2, DeltaRTT: 50}.RedisString(), time.Hour)
	redisClient.Set(fmt.Sprintf("sm-%s", sessionID2), transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 222, UserHash: userHash1, DeltaRTT: 100}.RedisString(), time.Hour)
	redisClient.Set(fmt.Sprintf("sm-%s", sessionID3), transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 333, UserHash: userHash1, DeltaRTT: 150}.RedisString(), time.Hour)
	redisClient.Set(fmt.Sprintf("sm-%s", sessionID4), transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 444, UserHash: userHash3, DeltaRTT: 150}.RedisString(), time.Hour)

	logger := log.NewNopLogger()

	btTableName, btTableEnvVarOK := os.LookupEnv("BIGTABLE_TABLE_NAME")
	if !btTableEnvVarOK {
		btTableName = "Test"
		os.Setenv("BIGTABLE_TABLE_NAME", btTableName)
		defer os.Unsetenv("BIGTABLE_TABLE_NAME")
	}

	// Get the column family name
	btCfName, btCfNameEnvVarOK := os.LookupEnv("BIGTABLE_CF_NAME")
	if !btCfNameEnvVarOK {
		btCfName = "TestCfName"
		os.Setenv("BIGTABLE_CF_NAME", btCfName)
		defer os.Unsetenv("BIGTABLE_CF_NAME")
	}

	// Check if table exists and create it if needed
	btAdmin, err := storage.NewBigTableAdmin(ctx, "local", "localhost:8086", logger)
	assert.NoError(t, err)
	assert.NotNil(t, btAdmin)
	btTableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	assert.NoError(t, err)
	if !btTableExists {
		// Create a table
		btAdmin.CreateTable(ctx, btTableName, []string{btCfName})
	}

	defer func() {
		err := btAdmin.DeleteTable(ctx, btTableName)
		assert.NoError(t, err)

		err = btAdmin.Close()
		assert.NoError(t, err)
	}()

	btClient, err := storage.NewBigTable(ctx, "local", "localhost:8086", btTableName, logger)
	assert.Nil(t, err)
	assert.NotNil(t, btClient)

	defer func() {
		err := btClient.Close()
		assert.NoError(t, err)
	}()

	// Add user sessions to bigtable
	meta1 := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 111, UserHash: userHash2, BuyerID: 999}
	metaBin1, err := transport.WriteSessionMeta(&meta1)
	assert.NoError(t, err)
	slice1 := transport.SessionSlice{Version: transport.SessionSliceVersion}
	sliceBin1, err := transport.WriteSessionSlice(&slice1)
	assert.NoError(t, err)
	sessionRowKey1 := sessionID1
	sliceRowKey1 := fmt.Sprintf("%s#%v", sessionID1, slice1.Timestamp)
	userRowKey1 := fmt.Sprintf("%016x#%s", userHash2, sessionID1)
	metaRowKeys1 := []string{sessionRowKey1, userRowKey1}
	sliceRowKeys1 := []string{sliceRowKey1}

	meta2 := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 222, UserHash: userHash1, BuyerID: 888}
	metaBin2, err := transport.WriteSessionMeta(&meta2)
	assert.NoError(t, err)
	slice2 := transport.SessionSlice{Version: transport.SessionSliceVersion}
	sliceBin2, err := transport.WriteSessionSlice(&slice2)
	assert.NoError(t, err)
	sessionRowKey2 := sessionID2
	sliceRowKey2 := fmt.Sprintf("%s#%v", sessionID2, slice2.Timestamp)
	userRowKey2 := fmt.Sprintf("%016x#%s", userHash1, sessionID2)
	metaRowKeys2 := []string{sessionRowKey2, userRowKey2}
	sliceRowKeys2 := []string{sliceRowKey2}

	meta3 := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 333, UserHash: userHash1, BuyerID: 888}
	metaBin3, err := transport.WriteSessionMeta(&meta3)
	assert.NoError(t, err)
	slice3 := transport.SessionSlice{Version: transport.SessionSliceVersion}
	sliceBin3, err := transport.WriteSessionSlice(&slice3)
	assert.NoError(t, err)
	sessionRowKey3 := sessionID3
	sliceRowKey3 := fmt.Sprintf("%s#%v", sessionID3, slice3.Timestamp)
	userRowKey3 := fmt.Sprintf("%016x#%s", userHash1, sessionID3)
	metaRowKeys3 := []string{sessionRowKey3, userRowKey3}
	sliceRowKeys3 := []string{sliceRowKey3}

	meta4 := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 444, UserHash: userHash3, BuyerID: 777}
	metaBin4, err := transport.WriteSessionMeta(&meta4)
	assert.NoError(t, err)
	slice4 := transport.SessionSlice{Version: transport.SessionSliceVersion}
	sliceBin4, err := transport.WriteSessionSlice(&slice4)
	assert.NoError(t, err)
	sessionRowKey4 := sessionID4
	sliceRowKey4 := fmt.Sprintf("%s#%v", sessionID4, slice4.Timestamp)
	userRowKey4 := fmt.Sprintf("%016x#%s", userHash3, sessionID4)
	metaRowKeys4 := []string{sessionRowKey4, userRowKey4}
	sliceRowKeys4 := []string{sliceRowKey4}

	err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin1, metaRowKeys1)
	assert.NoError(t, err)
	err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin1, sliceRowKeys1)
	assert.NoError(t, err)
	err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin2, metaRowKeys2)
	assert.NoError(t, err)
	err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin2, sliceRowKeys2)
	assert.NoError(t, err)
	err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin3, metaRowKeys3)
	assert.NoError(t, err)
	err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin3, sliceRowKeys3)
	assert.NoError(t, err)
	err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin4, metaRowKeys4)
	assert.NoError(t, err)
	err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin4, sliceRowKeys4)
	assert.NoError(t, err)

	redisClient.RPush(fmt.Sprintf("ss-%s", sessionID1), slice1.RedisString())
	redisClient.RPush(fmt.Sprintf("ss-%s", sessionID2), slice2.RedisString())
	redisClient.RPush(fmt.Sprintf("ss-%s", sessionID3), slice3.RedisString())
	redisClient.RPush(fmt.Sprintf("ss-%s", sessionID4), slice4.RedisString())

	svc := jsonrpc.BuyersService{
		Storage:                &storer,
		UseBigtable:            true,
		BigTableCfName:         btCfName,
		BigTable:               btClient,
		BigTableMetrics:        &metrics.EmptyBigTableMetrics,
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
		assert.EqualError(t, err, "UserSessions() user id is required")
		assert.Equal(t, 0, len(reply.Sessions))
	})

	t.Run("user_hash not found", func(t *testing.T) {
		var reply jsonrpc.UserSessionsReply
		err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: "12345"}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(reply.Sessions))
	})

	t.Run("Live Sessions", func(t *testing.T) {

		t.Run("list live - ID", func(t *testing.T) {
			var reply jsonrpc.UserSessionsReply
			err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: userID1}, &reply)
			assert.NoError(t, err)

			assert.Equal(t, 2, len(reply.Sessions))

			for _, session := range reply.Sessions {
				idString := fmt.Sprintf("%016x", session.Meta.ID)
				if idString != sessionID3 && idString != sessionID2 {
					t.Fail()
				}
			}
		})

		t.Run("list live - hash", func(t *testing.T) {
			var reply jsonrpc.UserSessionsReply
			err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: fmt.Sprintf("%016x", userHash1)}, &reply)
			assert.NoError(t, err)

			assert.Equal(t, 2, len(reply.Sessions))

			for _, session := range reply.Sessions {
				idString := fmt.Sprintf("%016x", session.Meta.ID)
				if idString != sessionID3 && idString != sessionID2 {
					t.Fail()
				}
			}
		})

		t.Run("list live - signed decimal hash", func(t *testing.T) {
			var reply jsonrpc.UserSessionsReply
			err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: fmt.Sprintf("%016x", userHash3)}, &reply)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(reply.Sessions))

			assert.Equal(t, fmt.Sprintf("%016x", reply.Sessions[0].Meta.ID), sessionID4)
		})
	})

	t.Run("Historic and Live Sessions", func(t *testing.T) {
		// Insert additional historic sessions
		sessionID6 := fmt.Sprintf("%016x", 666)
		sessionID7 := fmt.Sprintf("%016x", 777)
		sessionID8 := fmt.Sprintf("%016x", 888)

		meta6 := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 666, UserHash: userHash2, BuyerID: 999}
		metaBin6, err := transport.WriteSessionMeta(&meta6)
		assert.NoError(t, err)
		slice6 := transport.SessionSlice{Version: transport.SessionSliceVersion}
		sliceBin6, err := transport.WriteSessionSlice(&slice6)
		assert.NoError(t, err)
		sessionRowKey6 := sessionID6
		sliceRowKey6 := fmt.Sprintf("%s#%v", sessionID6, slice6.Timestamp)
		userRowKey6 := fmt.Sprintf("%016x#%s", userHash2, sessionID6)
		metaRowKeys6 := []string{sessionRowKey6, userRowKey6}
		sliceRowKeys6 := []string{sliceRowKey6}

		meta7 := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 777, UserHash: userHash1, BuyerID: 888}
		metaBin7, err := transport.WriteSessionMeta(&meta7)
		assert.NoError(t, err)
		slice7 := transport.SessionSlice{Version: transport.SessionSliceVersion}
		sliceBin7, err := transport.WriteSessionSlice(&slice7)
		assert.NoError(t, err)
		sessionRowKey7 := sessionID7
		sliceRowKey7 := fmt.Sprintf("%s#%v", sessionID7, slice7.Timestamp)
		userRowKey7 := fmt.Sprintf("%016x#%s", userHash1, sessionID7)
		metaRowKeys7 := []string{sessionRowKey7, userRowKey7}
		sliceRowKeys7 := []string{sliceRowKey7}

		meta8 := transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 888, UserHash: userHash1, BuyerID: 888}
		metaBin8, err := transport.WriteSessionMeta(&meta8)
		assert.NoError(t, err)
		slice8 := transport.SessionSlice{Version: transport.SessionSliceVersion}
		sliceBin8, err := transport.WriteSessionSlice(&slice8)
		assert.NoError(t, err)
		sessionRowKey8 := sessionID8
		sliceRowKey8 := fmt.Sprintf("%s#%v", sessionID8, slice8.Timestamp)
		userRowKey8 := fmt.Sprintf("%016x#%s", userHash3, sessionID8)
		metaRowKeys8 := []string{sessionRowKey8, userRowKey8}
		sliceRowKeys8 := []string{sliceRowKey8}

		err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin6, metaRowKeys6)
		assert.NoError(t, err)
		err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin6, sliceRowKeys6)
		assert.NoError(t, err)
		err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin7, metaRowKeys7)
		assert.NoError(t, err)
		err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin7, sliceRowKeys7)
		assert.NoError(t, err)
		err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin8, metaRowKeys8)
		assert.NoError(t, err)
		err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin8, sliceRowKeys8)
		assert.NoError(t, err)

		t.Run("list live and historic - ID", func(t *testing.T) {
			var reply jsonrpc.UserSessionsReply
			err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: userID1}, &reply)
			assert.NoError(t, err)

			assert.Equal(t, 3, len(reply.Sessions))

			for _, session := range reply.Sessions {
				idString := fmt.Sprintf("%016x", session.Meta.ID)
				if idString != sessionID3 && idString != sessionID2 && idString != sessionID7 {
					t.Fail()
				}
			}
		})

		t.Run("list live and historic - hash", func(t *testing.T) {
			var reply jsonrpc.UserSessionsReply
			err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: fmt.Sprintf("%016x", userHash1)}, &reply)
			assert.NoError(t, err)

			assert.Equal(t, 3, len(reply.Sessions))
			for _, session := range reply.Sessions {
				idString := fmt.Sprintf("%016x", session.Meta.ID)
				if idString != sessionID3 && idString != sessionID2 && idString != sessionID7 {
					t.Fail()
				}
			}
		})

		t.Run("list live and historic - signed decimal hash", func(t *testing.T) {
			var reply jsonrpc.UserSessionsReply
			err := svc.UserSessions(req, &jsonrpc.UserSessionsArgs{UserID: fmt.Sprintf("%016x", userHash3)}, &reply)
			assert.NoError(t, err)

			assert.Equal(t, 2, len(reply.Sessions))
			for _, session := range reply.Sessions {
				idString := fmt.Sprintf("%016x", session.Meta.ID)
				if idString != sessionID4 && idString != sessionID8 {
					t.Fail()
				}
			}
		})
	})
}

func TestDatacenterMaps(t *testing.T) {
	var storer = storage.InMemory{}
	dcMap := routing.DatacenterMap{
		BuyerID:      0xbdbebdbf0f7be395,
		DatacenterID: 0x7edb88d7b6fc0713,
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
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)

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
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
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
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
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
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)

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
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
	req = req.WithContext(reqContext)

	t.Run("all", func(t *testing.T) {
		var reply jsonrpc.TotalSessionsReply
		err := svc.TotalSessions(req, &jsonrpc.TotalSessionsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 7, reply.Next)
		assert.Equal(t, 250, reply.Direct)
	})

	t.Run("filtered - sameBuyer - !admin", func(t *testing.T) {
		var reply jsonrpc.TotalSessionsReply
		// test per buyer counts
		err := svc.TotalSessions(req, &jsonrpc.TotalSessionsArgs{CompanyCode: "local"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 5, reply.Next)
		assert.Equal(t, 250, reply.Direct)
	})
}

func TestTopSessions(t *testing.T) {
	t.Parallel()
	var storer = storage.InMemory{}

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 5, 5)
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

	redisClient.Set(fmt.Sprintf("sm-%s", sessionID1), transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 111, DeltaRTT: 50, BuyerID: 222}.RedisString(), time.Hour)
	redisClient.Set(fmt.Sprintf("sm-%s", sessionID2), transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 222, DeltaRTT: 100, BuyerID: 111}.RedisString(), time.Hour)
	redisClient.Set(fmt.Sprintf("sm-%s", sessionID3), transport.SessionMeta{Version: transport.SessionMetaVersion, ID: 333, DeltaRTT: 150, BuyerID: 111}.RedisString(), time.Hour)

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
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
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
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
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

func TestSessionDetails_Binary(t *testing.T) {
	checkBigtableEmulation(t)
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 5, 5)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionID := fmt.Sprintf("%016x", 999)

	meta := transport.SessionMeta{
		Version:    transport.SessionMetaVersion,
		BuyerID:    111,
		Location:   routing.Location{Latitude: 10, Longitude: 20},
		ClientAddr: "127.0.0.1:1313",
		ServerAddr: "10.0.0.1:50000",
		Hops: []transport.RelayHop{
			{Version: transport.RelayHopVersion, ID: 1234},
			{Version: transport.RelayHopVersion, ID: 1234},
			{Version: transport.RelayHopVersion, ID: 1234},
		},
		SDK: "3.4.4",
		NearbyRelays: []transport.NearRelayPortalData{
			{Version: transport.NearRelayPortalDataVersion, ID: 1, Name: "local", ClientStats: routing.Stats{RTT: 1, Jitter: 2, PacketLoss: 3}},
		},
	}

	anonMeta := transport.SessionMeta{
		Version:    transport.SessionMetaVersion,
		BuyerID:    111,
		Location:   routing.Location{Latitude: 10, Longitude: 20},
		ClientAddr: "127.0.0.1:1313",
		ServerAddr: "10.0.0.1:50000",
		Hops: []transport.RelayHop{
			{Version: transport.RelayHopVersion, ID: 1234},
			{Version: transport.RelayHopVersion, ID: 1234},
			{Version: transport.RelayHopVersion, ID: 1234},
		},
		SDK: "3.4.4",
		NearbyRelays: []transport.NearRelayPortalData{
			{Version: transport.NearRelayPortalDataVersion, ID: 1, Name: "local", ClientStats: routing.Stats{RTT: 1, Jitter: 2, PacketLoss: 3}},
		},
	}
	anonMeta.Anonymise()

	slice1 := transport.SessionSlice{
		Version:   transport.SessionSliceVersion,
		Timestamp: time.Now(),
		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
	}
	slice2 := transport.SessionSlice{
		Version:   transport.SessionSliceVersion,
		Timestamp: time.Now().Add(10 * time.Second),
		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
	}
	slice3 := transport.SessionSlice{
		Version:   transport.SessionSliceVersion,
		Timestamp: time.Now().Add(20 * time.Second),
		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
	}
	slice4 := transport.SessionSlice{
		Version:   transport.SessionSliceVersion,
		Timestamp: time.Now().Add(30 * time.Second),
		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
	}

	ctx := context.Background()
	logger := log.NewNopLogger()

	// Setup Bigtable
	btTableName, btTableEnvVarOK := os.LookupEnv("BIGTABLE_TABLE_NAME")
	if !btTableEnvVarOK {
		btTableName = "Test"
		os.Setenv("BIGTABLE_TABLE_NAME", btTableName)
		defer os.Unsetenv("BIGTABLE_TABLE_NAME")
	}

	// Get the column family name
	btCfName, btCfNameEnvVarOK := os.LookupEnv("BIGTABLE_CF_NAME")
	if !btCfNameEnvVarOK {
		btCfName = "TestCfName"
		os.Setenv("BIGTABLE_CF_NAME", btCfName)
		defer os.Unsetenv("BIGTABLE_CF_NAME")
	}

	// Check if table exists and create it if needed
	btAdmin, err := storage.NewBigTableAdmin(ctx, "local", "localhost:8086", logger)
	assert.NoError(t, err)
	assert.NotNil(t, btAdmin)
	btTableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	assert.NoError(t, err)
	if !btTableExists {
		// Create a table
		btAdmin.CreateTable(ctx, btTableName, []string{btCfName})
	}

	defer func() {
		err := btAdmin.DeleteTable(ctx, btTableName)
		assert.NoError(t, err)

		err = btAdmin.Close()
		assert.NoError(t, err)
	}()

	btClient, err := storage.NewBigTable(ctx, "local", "localhost:8086", btTableName, logger)
	assert.Nil(t, err)
	assert.NotNil(t, btClient)

	defer func() {
		err := btClient.Close()
		assert.NoError(t, err)
	}()

	inMemory := storage.InMemory{}
	inMemory.AddCustomer(ctx, routing.Customer{Code: "local", Name: "Local"})
	inMemory.AddBuyer(ctx, routing.Buyer{ID: 111, CompanyCode: "local"})
	inMemory.AddSeller(ctx, routing.Seller{ID: "local"})
	inMemory.AddDatacenter(ctx, routing.Datacenter{ID: 1})
	inMemory.AddRelay(ctx, routing.Relay{ID: 1, Name: "local", Seller: routing.Seller{ID: "local"}, Datacenter: routing.Datacenter{ID: 1}})

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		UseBigtable:            true,
		BigTableCfName:         btCfName,
		BigTable:               btClient,
		BigTableMetrics:        &metrics.EmptyBigTableMetrics,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &inMemory,
		Logger:                 logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("session_id not found", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: ""}, &reply)
		assert.Error(t, err)
	})

	// Add user sessions to bigtable
	metaBin, err := meta.MarshalBinary()
	assert.NoError(t, err)
	sliceBin3, err := slice3.MarshalBinary()
	assert.NoError(t, err)
	sliceBin4, err := slice4.MarshalBinary()
	assert.NoError(t, err)
	sessionRowKey := sessionID
	sliceRowKey3 := fmt.Sprintf("%s#%v", sessionID, slice3.Timestamp)
	sliceRowKey4 := fmt.Sprintf("%s#%v", sessionID, slice4.Timestamp)
	metaRowKeys := []string{sessionRowKey}
	sliceRowKeys3 := []string{sliceRowKey3}
	sliceRowKeys4 := []string{sliceRowKey4}

	err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin, metaRowKeys)
	assert.NoError(t, err)
	err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin3, sliceRowKeys3)
	assert.NoError(t, err)
	err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin4, sliceRowKeys4)
	assert.NoError(t, err)

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local-local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{})
	req = req.WithContext(reqContext)

	t.Run("success - bigtable - !admin - !sameBuyer", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: sessionID}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, anonMeta, reply.Meta)
		assert.Equal(t, slice3.Timestamp.Hour(), reply.Slices[0].Timestamp.Hour())
		assert.Equal(t, slice3.Next, reply.Slices[0].Next)
		assert.Equal(t, slice3.Direct, reply.Slices[0].Direct)
		assert.Equal(t, slice3.Envelope, reply.Slices[0].Envelope)
		assert.Equal(t, slice4.Timestamp.Hour(), reply.Slices[1].Timestamp.Hour())
		assert.Equal(t, slice4.Next, reply.Slices[1].Next)
		assert.Equal(t, slice4.Direct, reply.Slices[1].Direct)
		assert.Equal(t, slice4.Envelope, reply.Slices[1].Envelope)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{})
	req = req.WithContext(reqContext)

	t.Run("success - bigtable - !admin - sameBuyer", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: sessionID}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, meta, reply.Meta)
		assert.Equal(t, slice3.Timestamp.Hour(), reply.Slices[0].Timestamp.Hour())
		assert.Equal(t, slice3.Next, reply.Slices[0].Next)
		assert.Equal(t, slice3.Direct, reply.Slices[0].Direct)
		assert.Equal(t, slice3.Envelope, reply.Slices[0].Envelope)
		assert.Equal(t, slice4.Timestamp.Hour(), reply.Slices[1].Timestamp.Hour())
		assert.Equal(t, slice4.Next, reply.Slices[1].Next)
		assert.Equal(t, slice4.Direct, reply.Slices[1].Direct)
		assert.Equal(t, slice4.Envelope, reply.Slices[1].Envelope)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local-local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("success - bigtable - admin", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: sessionID}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, meta, reply.Meta)
		assert.Equal(t, slice3.Timestamp.Hour(), reply.Slices[0].Timestamp.Hour())
		assert.Equal(t, slice3.Next, reply.Slices[0].Next)
		assert.Equal(t, slice3.Direct, reply.Slices[0].Direct)
		assert.Equal(t, slice3.Envelope, reply.Slices[0].Envelope)
		assert.Equal(t, slice4.Timestamp.Hour(), reply.Slices[1].Timestamp.Hour())
		assert.Equal(t, slice4.Next, reply.Slices[1].Next)
		assert.Equal(t, slice4.Direct, reply.Slices[1].Direct)
		assert.Equal(t, slice4.Envelope, reply.Slices[1].Envelope)
	})

	redisClient.Set(fmt.Sprintf("sm-%s", sessionID), meta.RedisString(), 30*time.Second)
	redisClient.RPush(fmt.Sprintf("ss-%s", sessionID), slice1.RedisString(), slice2.RedisString())

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local-local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{})
	req = req.WithContext(reqContext)

	t.Run("success - !admin - !sameBuyer", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: sessionID}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, anonMeta, reply.Meta)
		assert.Equal(t, 2, len(reply.Slices))
		assert.Equal(t, slice1.Timestamp.Hour(), reply.Slices[0].Timestamp.Hour())
		assert.Equal(t, slice1.Next, reply.Slices[0].Next)
		assert.Equal(t, slice1.Direct, reply.Slices[0].Direct)
		assert.Equal(t, slice1.Envelope, reply.Slices[0].Envelope)
		assert.Equal(t, slice2.Timestamp.Hour(), reply.Slices[1].Timestamp.Hour())
		assert.Equal(t, slice2.Next, reply.Slices[1].Next)
		assert.Equal(t, slice2.Direct, reply.Slices[1].Direct)
		assert.Equal(t, slice2.Envelope, reply.Slices[1].Envelope)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{})
	req = req.WithContext(reqContext)

	t.Run("success - !admin - sameBuyer", func(t *testing.T) {
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

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local-local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
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

func TestSessionDetails_Serialize(t *testing.T) {
	checkBigtableEmulation(t)
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 5, 5)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionID := fmt.Sprintf("%016x", 999)

	meta := transport.SessionMeta{
		Version:    transport.SessionMetaVersion,
		BuyerID:    111,
		Location:   routing.Location{Latitude: 10, Longitude: 20},
		ClientAddr: "127.0.0.1:1313",
		ServerAddr: "10.0.0.1:50000",
		Hops: []transport.RelayHop{
			{Version: transport.RelayHopVersion, ID: 1234},
			{Version: transport.RelayHopVersion, ID: 1234},
			{Version: transport.RelayHopVersion, ID: 1234},
		},
		SDK: "3.4.4",
		NearbyRelays: []transport.NearRelayPortalData{
			{Version: transport.NearRelayPortalDataVersion, ID: 1, Name: "local", ClientStats: routing.Stats{RTT: 1, Jitter: 2, PacketLoss: 3}},
		},
	}

	anonMeta := transport.SessionMeta{
		Version:    transport.SessionMetaVersion,
		BuyerID:    111,
		Location:   routing.Location{Latitude: 10, Longitude: 20},
		ClientAddr: "127.0.0.1:1313",
		ServerAddr: "10.0.0.1:50000",
		Hops: []transport.RelayHop{
			{Version: transport.RelayHopVersion, ID: 1234},
			{Version: transport.RelayHopVersion, ID: 1234},
			{Version: transport.RelayHopVersion, ID: 1234},
		},
		SDK: "3.4.4",
		NearbyRelays: []transport.NearRelayPortalData{
			{Version: transport.NearRelayPortalDataVersion, ID: 1, Name: "local", ClientStats: routing.Stats{RTT: 1, Jitter: 2, PacketLoss: 3}},
		},
	}
	anonMeta.Anonymise()

	slice1 := transport.SessionSlice{
		Version:   transport.SessionSliceVersion,
		Timestamp: time.Now(),
		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
	}
	slice2 := transport.SessionSlice{
		Version:   transport.SessionSliceVersion,
		Timestamp: time.Now().Add(10 * time.Second),
		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
	}
	slice3 := transport.SessionSlice{
		Version:   transport.SessionSliceVersion,
		Timestamp: time.Now().Add(20 * time.Second),
		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
	}
	slice4 := transport.SessionSlice{
		Version:   transport.SessionSliceVersion,
		Timestamp: time.Now().Add(30 * time.Second),
		Next:      routing.Stats{RTT: 5, Jitter: 10, PacketLoss: 15},
		Direct:    routing.Stats{RTT: 15, Jitter: 20, PacketLoss: 25},
		Envelope:  routing.Envelope{Up: 1500, Down: 1500},
	}

	ctx := context.Background()
	logger := log.NewNopLogger()

	// Setup Bigtable
	btTableName, btTableEnvVarOK := os.LookupEnv("BIGTABLE_TABLE_NAME")
	if !btTableEnvVarOK {
		btTableName = "Test"
		os.Setenv("BIGTABLE_TABLE_NAME", btTableName)
		defer os.Unsetenv("BIGTABLE_TABLE_NAME")
	}

	// Get the column family name
	btCfName, btCfNameEnvVarOK := os.LookupEnv("BIGTABLE_CF_NAME")
	if !btCfNameEnvVarOK {
		btCfName = "TestCfName"
		os.Setenv("BIGTABLE_CF_NAME", btCfName)
		defer os.Unsetenv("BIGTABLE_CF_NAME")
	}

	// Check if table exists and create it if needed
	btAdmin, err := storage.NewBigTableAdmin(ctx, "local", "localhost:8086", logger)
	assert.NoError(t, err)
	assert.NotNil(t, btAdmin)
	btTableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	assert.NoError(t, err)
	if !btTableExists {
		// Create a table
		btAdmin.CreateTable(ctx, btTableName, []string{btCfName})
	}

	defer func() {
		err := btAdmin.DeleteTable(ctx, btTableName)
		assert.NoError(t, err)

		err = btAdmin.Close()
		assert.NoError(t, err)
	}()

	btClient, err := storage.NewBigTable(ctx, "local", "localhost:8086", btTableName, logger)
	assert.Nil(t, err)
	assert.NotNil(t, btClient)

	defer func() {
		err := btClient.Close()
		assert.NoError(t, err)
	}()

	inMemory := storage.InMemory{}
	inMemory.AddCustomer(ctx, routing.Customer{Code: "local", Name: "Local"})
	inMemory.AddBuyer(ctx, routing.Buyer{ID: 111, CompanyCode: "local"})
	inMemory.AddSeller(ctx, routing.Seller{ID: "local"})
	inMemory.AddDatacenter(ctx, routing.Datacenter{ID: 1})
	inMemory.AddRelay(ctx, routing.Relay{ID: 1, Name: "local", Seller: routing.Seller{ID: "local"}, Datacenter: routing.Datacenter{ID: 1}})

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		UseBigtable:            true,
		BigTableCfName:         btCfName,
		BigTable:               btClient,
		BigTableMetrics:        &metrics.EmptyBigTableMetrics,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &inMemory,
		Logger:                 logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("session_id not found", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: ""}, &reply)
		assert.Error(t, err)
	})

	// Add user sessions to bigtable
	metaBin, err := transport.WriteSessionMeta(&meta)
	assert.NoError(t, err)
	sliceBin3, err := transport.WriteSessionSlice(&slice3)
	assert.NoError(t, err)
	sliceBin4, err := transport.WriteSessionSlice(&slice4)
	assert.NoError(t, err)
	sessionRowKey := sessionID
	sliceRowKey3 := fmt.Sprintf("%s#%v", sessionID, slice3.Timestamp)
	sliceRowKey4 := fmt.Sprintf("%s#%v", sessionID, slice4.Timestamp)
	metaRowKeys := []string{sessionRowKey}
	sliceRowKeys3 := []string{sliceRowKey3}
	sliceRowKeys4 := []string{sliceRowKey4}

	err = btClient.InsertSessionMetaData(ctx, []string{btCfName}, metaBin, metaRowKeys)
	assert.NoError(t, err)
	err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin3, sliceRowKeys3)
	assert.NoError(t, err)
	err = btClient.InsertSessionSliceData(ctx, []string{btCfName}, sliceBin4, sliceRowKeys4)
	assert.NoError(t, err)

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local-local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{})
	req = req.WithContext(reqContext)

	t.Run("success - bigtable - !admin - !sameBuyer", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: sessionID}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, anonMeta, reply.Meta)
		assert.Equal(t, slice3.Timestamp.Hour(), reply.Slices[0].Timestamp.Hour())
		assert.Equal(t, slice3.Next, reply.Slices[0].Next)
		assert.Equal(t, slice3.Direct, reply.Slices[0].Direct)
		assert.Equal(t, slice3.Envelope, reply.Slices[0].Envelope)
		assert.Equal(t, slice4.Timestamp.Hour(), reply.Slices[1].Timestamp.Hour())
		assert.Equal(t, slice4.Next, reply.Slices[1].Next)
		assert.Equal(t, slice4.Direct, reply.Slices[1].Direct)
		assert.Equal(t, slice4.Envelope, reply.Slices[1].Envelope)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{})
	req = req.WithContext(reqContext)

	t.Run("success - bigtable - !admin - sameBuyer", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: sessionID}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, meta, reply.Meta)
		assert.Equal(t, slice3.Timestamp.Hour(), reply.Slices[0].Timestamp.Hour())
		assert.Equal(t, slice3.Next, reply.Slices[0].Next)
		assert.Equal(t, slice3.Direct, reply.Slices[0].Direct)
		assert.Equal(t, slice3.Envelope, reply.Slices[0].Envelope)
		assert.Equal(t, slice4.Timestamp.Hour(), reply.Slices[1].Timestamp.Hour())
		assert.Equal(t, slice4.Next, reply.Slices[1].Next)
		assert.Equal(t, slice4.Direct, reply.Slices[1].Direct)
		assert.Equal(t, slice4.Envelope, reply.Slices[1].Envelope)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local-local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("success - bigtable - admin", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: sessionID}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, meta, reply.Meta)
		assert.Equal(t, slice3.Timestamp.Hour(), reply.Slices[0].Timestamp.Hour())
		assert.Equal(t, slice3.Next, reply.Slices[0].Next)
		assert.Equal(t, slice3.Direct, reply.Slices[0].Direct)
		assert.Equal(t, slice3.Envelope, reply.Slices[0].Envelope)
		assert.Equal(t, slice4.Timestamp.Hour(), reply.Slices[1].Timestamp.Hour())
		assert.Equal(t, slice4.Next, reply.Slices[1].Next)
		assert.Equal(t, slice4.Direct, reply.Slices[1].Direct)
		assert.Equal(t, slice4.Envelope, reply.Slices[1].Envelope)
	})

	redisClient.Set(fmt.Sprintf("sm-%s", sessionID), meta.RedisString(), 30*time.Second)
	redisClient.RPush(fmt.Sprintf("ss-%s", sessionID), slice1.RedisString(), slice2.RedisString())

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local-local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{})
	req = req.WithContext(reqContext)

	t.Run("success - !admin - !sameBuyer", func(t *testing.T) {
		var reply jsonrpc.SessionDetailsReply
		err := svc.SessionDetails(req, &jsonrpc.SessionDetailsArgs{SessionID: sessionID}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, anonMeta, reply.Meta)
		assert.Equal(t, 2, len(reply.Slices))
		assert.Equal(t, slice1.Timestamp.Hour(), reply.Slices[0].Timestamp.Hour())
		assert.Equal(t, slice1.Next, reply.Slices[0].Next)
		assert.Equal(t, slice1.Direct, reply.Slices[0].Direct)
		assert.Equal(t, slice1.Envelope, reply.Slices[0].Envelope)
		assert.Equal(t, slice2.Timestamp.Hour(), reply.Slices[1].Timestamp.Hour())
		assert.Equal(t, slice2.Next, reply.Slices[1].Next)
		assert.Equal(t, slice2.Direct, reply.Slices[1].Direct)
		assert.Equal(t, slice2.Envelope, reply.Slices[1].Envelope)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{})
	req = req.WithContext(reqContext)

	t.Run("success - !admin - sameBuyer", func(t *testing.T) {
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

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local-local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
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
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 5, 5)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	buyerID1 := fmt.Sprintf("%016x", 111)
	buyerID2 := fmt.Sprintf("%016x", 222)

	sessionID1 := fmt.Sprintf("%016x", 111)
	sessionID2 := fmt.Sprintf("%016x", 222)
	sessionID3 := fmt.Sprintf("%016x", 333)

	points := []transport.SessionMapPoint{
		{Version: transport.SessionMapPointVersion, Latitude: 10, Longitude: 40},
		{Version: transport.SessionMapPointVersion, Latitude: 20, Longitude: 50},
		{Version: transport.SessionMapPointVersion, Latitude: 30, Longitude: 60},
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

	err := svc.GenerateMapPointsPerBuyer(context.Background())
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
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
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
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
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
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 5, 5)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	buyerID1 := fmt.Sprintf("%016x", 111)
	buyerID2 := fmt.Sprintf("%016x", 222)

	sessionID1 := fmt.Sprintf("%016x", 111)
	sessionID2 := fmt.Sprintf("%016x", 222)
	sessionID3 := fmt.Sprintf("%016x", 333)

	points := []transport.SessionMapPoint{
		{Version: transport.SessionMapPointVersion, Latitude: 10, Longitude: 40, SessionID: uint64(123456789)},
		{Version: transport.SessionMapPointVersion, Latitude: 20, Longitude: 50, SessionID: uint64(123123123)},
		{Version: transport.SessionMapPointVersion, Latitude: 30, Longitude: 60, SessionID: uint64(456456456)},
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

	err := svc.GenerateMapPointsPerBuyer(context.Background())
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
		assert.Equal(t, []interface{}{float64(60), float64(30), false, "000000001b34f908"}, mappoints[0])
		assert.Equal(t, []interface{}{float64(40), float64(10), true, "00000000075bcd15"}, mappoints[1])
		assert.Equal(t, []interface{}{float64(50), float64(20), true, "000000000756b5b3"}, mappoints[2])
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
	req = req.WithContext(reqContext)

	t.Run("filtered - !admin - sameBuyer", func(t *testing.T) {
		var reply jsonrpc.MapPointsReply
		err := svc.SessionMap(req, &jsonrpc.MapPointsArgs{CompanyCode: "local"}, &reply)
		assert.NoError(t, err)

		var mappoints [][]interface{}
		err = json.Unmarshal(reply.Points, &mappoints)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(mappoints))
		assert.Equal(t, []interface{}{float64(60), float64(30), false, "000000001b34f908"}, mappoints[0])
		assert.Equal(t, []interface{}{float64(40), float64(10), true, "00000000075bcd15"}, mappoints[1])
	})

	t.Run("filtered - !admin - !sameBuyer", func(t *testing.T) {
		var reply jsonrpc.MapPointsReply
		err := svc.SessionMap(req, &jsonrpc.MapPointsArgs{CompanyCode: "local-local"}, &reply)
		assert.Error(t, err)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
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
		assert.Equal(t, []interface{}{float64(50), float64(20), true, "000000000756b5b3"}, mappoints[0])
	})
}

func TestGameConfiguration(t *testing.T) {
	t.Parallel()
	var storer = storage.InMemory{}

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
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

	middleware.SetIsAnonymous(req, true)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.GameConfiguration(req, &jsonrpc.GameConfigurationArgs{}, &reply)
		assert.Error(t, err)
	})

	middleware.SetIsAnonymous(req, false)

	t.Run("no company", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.GameConfiguration(req, &jsonrpc.GameConfigurationArgs{}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
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
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
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

	middleware.SetIsAnonymous(req, true)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{}, &reply)
		assert.Error(t, err)
	})

	middleware.SetIsAnonymous(req, false)

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Owner",
	})
	req = req.WithContext(reqContext)

	t.Run("no company", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{}, &reply)
		assert.Error(t, err)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local-local")
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

		newBuyer, err := storer.BuyerWithCompanyCode(reqContext, "local-local")
		assert.NoError(t, err)

		assert.Equal(t, "local-local", newBuyer.CompanyCode)
		assert.False(t, newBuyer.Live)
		assert.Equal(t, "12939405032490452521", fmt.Sprintf("%d", newBuyer.ID))
		assert.Equal(t, "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==", reply.GameConfiguration.PublicKey)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
	req = req.WithContext(reqContext)

	t.Run("success - existing buyer", func(t *testing.T) {
		var reply jsonrpc.GameConfigurationReply
		err := svc.UpdateGameConfiguration(req, &jsonrpc.GameConfigurationArgs{NewPublicKey: "45Q+5CKzGkcf3mh8cD43UM8L6Wn81tVwmmlT3Xvs9HWSJp5Zyh5xZg=="}, &reply)
		assert.NoError(t, err)

		oldBuyer, err := storer.BuyerWithCompanyCode(reqContext, "local")
		assert.NoError(t, err)

		assert.Equal(t, "local", oldBuyer.CompanyCode)
		assert.Equal(t, "5123604488526927075", fmt.Sprintf("%d", oldBuyer.ID))
		assert.True(t, oldBuyer.Live)
		assert.Equal(t, "45Q+5CKzGkcf3mh8cD43UM8L6Wn81tVwmmlT3Xvs9HWSJp5Zyh5xZg==", reply.GameConfiguration.PublicKey)
	})
}

func TestSameBuyerRoleFunction(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
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
		assert.NoError(t, err)
		assert.False(t, verified)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
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
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local-local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
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

// moved to /backend/cmd/next - will have to port the test
// func TestGetAllSessionBillingInfo(t *testing.T) {
// 	// this test follows all the add& rules for the SQL storer

// 	redisServer, _ := miniredis.Run()
// 	redisPool := storage.NewRedisPool(redisServer.Addr(), 1, 1)
// 	var storer = storage.InMemory{}

// 	ctx := context.Background()

// 	customerShortName := "shortName"
// 	customer := routing.Customer{
// 		Code:                   customerShortName,
// 		Name:                   "Company, Ltd.",
// 		AutomaticSignInDomains: "fredscuttle.com",
// 	}

// 	err := storer.AddCustomer(ctx, customer)
// 	assert.NoError(t, err)

// 	outerCustomer, err := storer.Customer(customerShortName)
// 	assert.NoError(t, err)

// 	seller := routing.Seller{
// 		ID:                        customerShortName,
// 		ShortName:                 customerShortName,
// 		CompanyCode:               customerShortName,
// 		IngressPriceNibblinsPerGB: 10,
// 		EgressPriceNibblinsPerGB:  20,
// 		CustomerID:                outerCustomer.DatabaseID,
// 	}

// 	err = storer.AddSeller(ctx, seller)
// 	assert.NoError(t, err)

// 	outerSeller, err := storer.Seller(customerShortName)
// 	assert.NoError(t, err)

// 	publicKey := []byte("01234567890123456789012345678901")
// 	internalBuyerID := binary.LittleEndian.Uint64(publicKey[:8])

// 	buyer := routing.Buyer{
// 		ShortName:   outerCustomer.Code,
// 		CompanyCode: outerCustomer.Code,
// 		Live:        true,
// 		Debug:       true,
// 		PublicKey:   publicKey,
// 		ID:          internalBuyerID, // sql store ignores this field, fs and in_memory require it
// 	}

// 	err = storer.AddBuyer(ctx, buyer)
// 	assert.NoError(t, err)

// 	outerBuyer, err := storer.Buyer(internalBuyerID)
// 	assert.NoError(t, err)

// 	// we need to add the buyer ID to bq_billing_row.json
// 	// fmt.Printf("BuyerID: %d\n", outerBuyer.ID)

// 	datacenter := routing.Datacenter{
// 		ID:   crypto.HashID("some.locale.name"),
// 		Name: "some.locale.name",
// 		Location: routing.Location{
// 			Latitude:  70.5,
// 			Longitude: 120.5,
// 		},
// 		SellerID: outerSeller.DatabaseID,
// 	}

// 	err = storer.AddDatacenter(ctx, datacenter)
// 	assert.NoError(t, err)

// 	outerDatacenter, err := storer.Datacenter(datacenter.ID)
// 	assert.NoError(t, err)

// 	// we need to add the datacenter ID to bq_billing_row.json
// 	// fmt.Printf("DatacenterID: %d\n", int64(outerDatacenter.ID))

// 	addr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
// 	assert.NoError(t, err)

// 	internalAddr1, err := net.ResolveUDPAddr("udp", "192.168.0.1:40000")
// 	assert.NoError(t, err)

// 	rid1 := crypto.HashID(addr1.String())

// 	relay1PublicKey := []byte("12345678901234567890123456789012")

// 	relay1 := routing.Relay{
// 		ID:                  rid1,
// 		Name:                "local.1",
// 		Addr:                *addr1,
// 		InternalAddr:        *internalAddr1,
// 		ManagementAddr:      "1.2.3.4",
// 		SSHPort:             22,
// 		SSHUser:             "fred",
// 		MaxSessions:         1000,
// 		PublicKey:           relay1PublicKey,
// 		Datacenter:          outerDatacenter,
// 		MRC:                 19700000000000,
// 		Overage:             26000000000000,
// 		BWRule:              routing.BWRuleBurst,
// 		ContractTerm:        12,
// 		StartDate:           time.Now(),
// 		EndDate:             time.Now(),
// 		Type:                routing.BareMetal,
// 		State:               routing.RelayStateMaintenance,
// 		IncludedBandwidthGB: 10000,
// 		NICSpeedMbps:        1000,
// 		Seller:              outerSeller, // TODO: remove - in_memory requires this, it is meaningless for the SQL storer
// 	}

// 	err = storer.AddRelay(ctx, relay1)
// 	assert.NoError(t, err)

// 	// checkRelay1, err := storer.Relay(rid1)
// 	_, err = storer.Relay(rid1)
// 	assert.NoError(t, err)

// 	// we need to add the relay ID to bq_billing_row.json
// 	// fmt.Printf("Relay1 ID: %d\n", int64(checkRelay1.ID))

// 	addr2, err := net.ResolveUDPAddr("udp", "127.0.0.2:40000")
// 	assert.NoError(t, err)

// 	internalAddr2, err := net.ResolveUDPAddr("udp", "192.168.0.2:40000")
// 	assert.NoError(t, err)

// 	rid2 := crypto.HashID(addr2.String())

// 	relay2PublicKey := []byte("12345678901234567890123456789012")

// 	relay2 := routing.Relay{
// 		ID:                  rid2,
// 		Name:                "local.2",
// 		Addr:                *addr2,
// 		InternalAddr:        *internalAddr2,
// 		ManagementAddr:      "1.2.3.4",
// 		SSHPort:             22,
// 		SSHUser:             "fred",
// 		MaxSessions:         1000,
// 		PublicKey:           relay2PublicKey,
// 		Datacenter:          outerDatacenter,
// 		MRC:                 19700000000000,
// 		Overage:             26000000000000,
// 		BWRule:              routing.BWRuleBurst,
// 		ContractTerm:        12,
// 		StartDate:           time.Now(),
// 		EndDate:             time.Now(),
// 		Type:                routing.BareMetal,
// 		State:               routing.RelayStateMaintenance,
// 		IncludedBandwidthGB: 10000,
// 		NICSpeedMbps:        1000,
// 		Seller:              outerSeller, // TODO: remove - in_memory requires this, it is meaningless for the SQL storer
// 	}

// 	err = storer.AddRelay(ctx, relay2)
// 	assert.NoError(t, err)

// 	// checkRelay2, err := storer.Relay(rid2)
// 	_, err = storer.Relay(rid2)
// 	assert.NoError(t, err)

// 	// we need to add the relay ID to bq_billing_row.json
// 	// fmt.Printf("Relay2 ID: %d\n", int64(checkRelay2.ID))

// 	logger := log.NewNopLogger()
// 	svc := jsonrpc.BuyersService{
// 		RedisPoolSessionMap:    redisPool,
// 		RedisPoolSessionMeta:   redisPool,
// 		RedisPoolSessionSlices: redisPool,
// 		RedisPoolTopSessions:   redisPool,
// 		Storage:                &storer,
// 		Logger:                 logger,
// 	}
// 	req := httptest.NewRequest(http.MethodGet, "/", nil)

// 	t.Run("get local json", func(t *testing.T) {

// 		// this test relies on the values in place in testdata/bq_billing_row.json

// 		arg := &jsonrpc.GetAllSessionBillingInfoArg{
// 			SessionID: 5782814167830682977,
// 		}

// 		reply := &jsonrpc.GetAllSessionBillingInfoReply{
// 			SessionBillingInfo: []transport.BigQueryBillingEntry{},
// 		}

// 		err := svc.GetAllSessionBillingInfo(req, arg, reply)
// 		assert.NoError(t, err)

// 		assert.Equal(t, int64(outerBuyer.ID), reply.SessionBillingInfo[0].BuyerID)
// 		assert.Equal(t, outerBuyer.ShortName, reply.SessionBillingInfo[0].BuyerString)
// 		assert.Equal(t, int64(8000587274513071088), reply.SessionBillingInfo[0].SessionID)
// 		assert.Equal(t, int64(11), reply.SessionBillingInfo[0].SliceNumber)
// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].Next)
// 		assert.Equal(t, 21.8623, reply.SessionBillingInfo[0].DirectRTT)
// 		assert.Equal(t, 0.41678265, reply.SessionBillingInfo[0].DirectJitter)
// 		assert.Equal(t, 0.0, reply.SessionBillingInfo[0].DirectPacketLoss)
// 		assert.Equal(t, bigquery.NullFloat64{Float64: 23.9552, Valid: true}, reply.SessionBillingInfo[0].NextRTT)
// 		assert.Equal(t, bigquery.NullFloat64{Float64: 3.2498908, Valid: true}, reply.SessionBillingInfo[0].NextJitter)
// 		assert.Equal(t, bigquery.NullFloat64{Float64: 0.0, Valid: true}, reply.SessionBillingInfo[0].NextPacketLoss)

// 		assert.Equal(t, int64(relay1.ID), reply.SessionBillingInfo[0].NextRelays[0])
// 		assert.Equal(t, relay1.Name, reply.SessionBillingInfo[0].NextRelaysStrings[0])
// 		assert.Equal(t, int64(relay2.ID), reply.SessionBillingInfo[0].NextRelays[1])
// 		assert.Equal(t, relay2.Name, reply.SessionBillingInfo[0].NextRelaysStrings[1])

// 		assert.Equal(t, int64(12800000), reply.SessionBillingInfo[0].TotalPrice)

// 		// 2 empty/null columns
// 		assert.Equal(t, bigquery.NullInt64{Int64: 0, Valid: false}, reply.SessionBillingInfo[0].ClientToServerPacketsLost)
// 		assert.Equal(t, bigquery.NullInt64{Int64: 0, Valid: false}, reply.SessionBillingInfo[0].ServerToClientPacketsLost)

// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].Committed)
// 		assert.Equal(t, bigquery.NullBool{Bool: false, Valid: true}, reply.SessionBillingInfo[0].Flagged)
// 		assert.Equal(t, bigquery.NullBool{Bool: false, Valid: true}, reply.SessionBillingInfo[0].Multipath)
// 		assert.Equal(t, bigquery.NullInt64{Int64: 127500, Valid: true}, reply.SessionBillingInfo[0].NextBytesUp)
// 		assert.Equal(t, bigquery.NullInt64{Int64: 153750, Valid: true}, reply.SessionBillingInfo[0].NextBytesDown)
// 		assert.Equal(t, bigquery.NullInt64{Int64: int64(outerDatacenter.ID), Valid: true}, reply.SessionBillingInfo[0].DatacenterID)
// 		assert.Equal(t, bigquery.NullString{StringVal: outerDatacenter.Name, Valid: true}, reply.SessionBillingInfo[0].DatacenterString)
// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].RttReduction)
// 		assert.Equal(t, bigquery.NullBool{Bool: false, Valid: true}, reply.SessionBillingInfo[0].PacketLossReduction)

// 		assert.Equal(t, int64(5120000), reply.SessionBillingInfo[0].NextRelaysPrice[0])
// 		assert.Equal(t, int64(5120000), reply.SessionBillingInfo[0].NextRelaysPrice[1])

// 		assert.Equal(t, bigquery.NullInt64{Int64: 3161472066075933729, Valid: true}, reply.SessionBillingInfo[0].UserHash)
// 		assert.Equal(t, bigquery.NullFloat64{Float64: 42.7273, Valid: true}, reply.SessionBillingInfo[0].Latitude)
// 		assert.Equal(t, bigquery.NullFloat64{Float64: -73.6696, Valid: true}, reply.SessionBillingInfo[0].Longitude)
// 		assert.Equal(t, bigquery.NullString{StringVal: "CSTC", Valid: true}, reply.SessionBillingInfo[0].ISP)
// 		assert.Equal(t, bigquery.NullBool{Bool: false, Valid: true}, reply.SessionBillingInfo[0].ABTest)
// 		assert.Equal(t, bigquery.NullInt64{Int64: 1, Valid: true}, reply.SessionBillingInfo[0].ConnectionType)
// 		assert.Equal(t, bigquery.NullInt64{Int64: 1, Valid: true}, reply.SessionBillingInfo[0].PlatformType)
// 		assert.Equal(t, bigquery.NullString{StringVal: "4.0.1", Valid: true}, reply.SessionBillingInfo[0].SdkVersion)

// 		assert.Equal(t, bigquery.NullInt64{Int64: 1280000, Valid: true}, reply.SessionBillingInfo[0].EnvelopeBytesUp)
// 		assert.Equal(t, bigquery.NullInt64{Int64: 1280000, Valid: true}, reply.SessionBillingInfo[0].EnvelopeBytesDown)
// 		assert.Equal(t, bigquery.NullFloat64{Float64: 6.0, Valid: true}, reply.SessionBillingInfo[0].PredictedNextRTT)
// 		assert.Equal(t, bigquery.NullBool{Bool: false, Valid: true}, reply.SessionBillingInfo[0].MultipathVetoed)
// 		assert.Equal(t, bigquery.NullString{StringVal: "", Valid: false}, reply.SessionBillingInfo[0].Debug)
// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].FallbackToDirect)
// 		assert.Equal(t, bigquery.NullInt64{Int64: 8, Valid: true}, reply.SessionBillingInfo[0].ClientFlags)
// 		assert.Equal(t, bigquery.NullInt64{Int64: 6, Valid: true}, reply.SessionBillingInfo[0].UserFlags)

// 		assert.Equal(t, bigquery.NullFloat64{Float64: 31.8623, Valid: true}, reply.SessionBillingInfo[0].NearRelayRTT)
// 		assert.Equal(t, bigquery.NullInt64{Int64: 5, Valid: true}, reply.SessionBillingInfo[0].PacketsOutOfOrderClientToServer)
// 		assert.Equal(t, bigquery.NullInt64{Int64: 15, Valid: true}, reply.SessionBillingInfo[0].PacketsOutOfOrderServerToClient)
// 		assert.Equal(t, bigquery.NullFloat64{Float64: 64.328, Valid: true}, reply.SessionBillingInfo[0].JitterClientToServer)
// 		assert.Equal(t, bigquery.NullFloat64{Float64: 75.764, Valid: true}, reply.SessionBillingInfo[0].JitterServerToClient)
// 		assert.Equal(t, bigquery.NullInt64{Int64: 5, Valid: true}, reply.SessionBillingInfo[0].NumNearRelays)

// 		assert.Equal(t, int64(1141867895387079451), reply.SessionBillingInfo[0].NearRelayIDs[0])
// 		assert.Equal(t, int64(-7664475006302134894), reply.SessionBillingInfo[0].NearRelayIDs[1])
// 		assert.Equal(t, int64(-6848787315892866519), reply.SessionBillingInfo[0].NearRelayIDs[2])

// 		assert.Equal(t, float64(12.345), reply.SessionBillingInfo[0].NearRelayRTTs[0])
// 		assert.Equal(t, float64(23.456), reply.SessionBillingInfo[0].NearRelayRTTs[1])
// 		assert.Equal(t, float64(34.567), reply.SessionBillingInfo[0].NearRelayRTTs[2])

// 		assert.Equal(t, float64(1.23), reply.SessionBillingInfo[0].NearRelayJitters[0])
// 		assert.Equal(t, float64(2.34), reply.SessionBillingInfo[0].NearRelayJitters[1])
// 		assert.Equal(t, float64(3.45), reply.SessionBillingInfo[0].NearRelayJitters[2])

// 		assert.Equal(t, float64(5.43), reply.SessionBillingInfo[0].NearRelayPacketLosses[0])
// 		assert.Equal(t, float64(4.32), reply.SessionBillingInfo[0].NearRelayPacketLosses[1])
// 		assert.Equal(t, float64(3.21), reply.SessionBillingInfo[0].NearRelayPacketLosses[2])

// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].RelayWentAway)
// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].RouteLost)

// 		assert.Equal(t, int64(1141867895387079451), reply.SessionBillingInfo[0].Tags[0])
// 		assert.Equal(t, int64(-7664475006302134894), reply.SessionBillingInfo[0].Tags[1])
// 		assert.Equal(t, int64(-6848787315892866519), reply.SessionBillingInfo[0].Tags[2])

// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].Mispredicted)
// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].Vetoed)
// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].LatencyWorse)
// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].NoRoute)
// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].NextLatencyTooHigh)
// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].RouteChanged)
// 		assert.Equal(t, bigquery.NullBool{Bool: true, Valid: true}, reply.SessionBillingInfo[0].CommitVeto)

// 	})
// }
