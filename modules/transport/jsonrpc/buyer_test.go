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
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/modules/core"
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

	err := storer.AddCustomer(context.Background(), routing.Customer{Name: "Local", Code: "local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local"})
	assert.NoError(t, err)
	err = storer.AddCustomer(context.Background(), routing.Customer{Name: "Local Local", Code: "local-local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 2, CompanyCode: "local-local"})
	assert.NoError(t, err)
	err = storer.AddCustomer(context.Background(), routing.Customer{Name: "Local Local Local", Code: "local-local-local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 3, CompanyCode: "local-local-local"})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		Storage: &storer,
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
	err := storer.AddBuyer(ctx, buyer0)
	assert.NoError(t, err)
	err = storer.AddBuyer(ctx, buyer7)
	assert.NoError(t, err)
	err = storer.AddBuyer(ctx, buyer8)
	assert.NoError(t, err)
	err = storer.AddBuyer(ctx, buyer9)
	assert.NoError(t, err)
	err = storer.AddCustomer(ctx, customer0)
	assert.NoError(t, err)
	err = storer.AddCustomer(ctx, customer7)
	assert.NoError(t, err)
	err = storer.AddCustomer(ctx, customer8)
	assert.NoError(t, err)
	err = storer.AddCustomer(ctx, customer9)
	assert.NoError(t, err)

	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 5, 5)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	rawUserID1 := 111
	rawUserID2 := 222
	userID1 := fmt.Sprintf("%d", rawUserID1)
	userID2 := fmt.Sprintf("%d", rawUserID2)

	hash1 := fnv.New64a()
	_, err = hash1.Write([]byte(userID1))
	assert.NoError(t, err)
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
	btAdmin, err := storage.NewBigTableAdmin(ctx, "local", "localhost:8086")
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

	btClient, err := storage.NewBigTable(ctx, "local", "localhost:8086", btTableName)
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
	err := storer.AddBuyer(ctx, buyer)
	assert.NoError(t, err)
	err = storer.AddCustomer(ctx, customer)
	assert.NoError(t, err)
	err = storer.AddDatacenter(ctx, datacenter)
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		Storage: &storer,
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

	err := storer.AddCustomer(ctx, routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddCustomer(ctx, routing.Customer{Code: "local-local", Name: "Local Local"})
	assert.NoError(t, err)
	err = storer.AddCustomer(ctx, routing.Customer{Code: "local-local-local", Name: "Local Local Local"})
	assert.NoError(t, err)

	err = storer.AddBuyer(ctx, routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)
	err = storer.AddBuyer(ctx, routing.Buyer{ID: 2, CompanyCode: "local-local", PublicKey: pubkey})
	assert.NoError(t, err)
	err = storer.AddBuyer(ctx, routing.Buyer{ID: 3, CompanyCode: "local-local-local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolTopSessions:   redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolSessionMap:    redisPool,
		Storage:                &storer,
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

	err := storer.AddCustomer(ctx, routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddCustomer(ctx, routing.Customer{Code: "local-local", Name: "Local Local"})
	assert.NoError(t, err)
	err = storer.AddCustomer(ctx, routing.Customer{Code: "local-local-local", Name: "Local Local Local"})
	assert.NoError(t, err)

	err = storer.AddBuyer(ctx, routing.Buyer{ID: 0, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)
	err = storer.AddBuyer(ctx, routing.Buyer{ID: 1, CompanyCode: "local-local", PublicKey: pubkey})
	assert.NoError(t, err)
	err = storer.AddBuyer(ctx, routing.Buyer{ID: 2, CompanyCode: "local-local-local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolTopSessions:   redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolSessionMap:    redisPool,
		Storage:                &storer,
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
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 111, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)
	err = storer.AddCustomer(context.Background(), routing.Customer{Code: "local-local", Name: "Local Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 222, CompanyCode: "local-local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
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
	btAdmin, err := storage.NewBigTableAdmin(ctx, "local", "localhost:8086")
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

	btClient, err := storage.NewBigTable(ctx, "local", "localhost:8086", btTableName)
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
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddCustomer(context.Background(), routing.Customer{Code: "local-local", Name: "Local Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 111, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 222, CompanyCode: "local-local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	err = svc.GenerateMapPointsPerBuyer(context.Background())
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
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddCustomer(context.Background(), routing.Customer{Code: "local-local", Name: "Local Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 111, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 222, CompanyCode: "local-local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	err = svc.GenerateMapPointsPerBuyer(context.Background())
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
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolTopSessions:   redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolSessionMap:    redisPool,
		Storage:                &storer,
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
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddCustomer(context.Background(), routing.Customer{Code: "local-local", Name: "Local Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey, Live: true})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolTopSessions:   redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolSessionMap:    redisPool,
		Storage:                &storer,
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
		assert.True(t, newBuyer.Live)
		assert.True(t, newBuyer.RouteShader.AnalysisOnly)
		assert.Equal(t, "12939405032490452521", fmt.Sprintf("%d", newBuyer.ID))
		assert.Equal(t, "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==", reply.GameConfiguration.PublicKey)

		// TODO: Figure out how to test bin generation and verification
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
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
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

func TestInternalConfig(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("invalid buyer id", func(t *testing.T) {
		var reply jsonrpc.InternalConfigReply
		err := svc.InternalConfig(req, &jsonrpc.InternalConfigArg{BuyerID: "badBuyerID"}, &reply)
		assert.Error(t, err)
	})

	t.Run("no internal config", func(t *testing.T) {
		var reply jsonrpc.InternalConfigReply
		err := svc.InternalConfig(req, &jsonrpc.InternalConfigArg{BuyerID: "1"}, &reply)
		assert.Error(t, err)
	})

	ic := core.InternalConfig{
		RouteSelectThreshold:           2,
		RouteSwitchThreshold:           5,
		MaxLatencyTradeOff:             20,
		RTTVeto_Default:                -10,
		RTTVeto_Multipath:              -20,
		RTTVeto_PacketLoss:             -30,
		MultipathOverloadThreshold:     500,
		TryBeforeYouBuy:                false,
		ForceNext:                      false,
		LargeCustomer:                  false,
		Uncommitted:                    false,
		MaxRTT:                         300,
		HighFrequencyPings:             true,
		RouteDiversity:                 0,
		MultipathThreshold:             25,
		EnableVanityMetrics:            false,
		ReducePacketLossMinSliceNumber: 0,
	}

	err = storer.AddCustomer(context.Background(), routing.Customer{Code: "local-local", Name: "Local-Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 2, CompanyCode: "local-local", PublicKey: pubkey, InternalConfig: ic})
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.InternalConfigReply
		err := svc.InternalConfig(req, &jsonrpc.InternalConfigArg{BuyerID: "2"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, ic.RouteSelectThreshold, int32(reply.InternalConfig.RouteSelectThreshold))
		assert.Equal(t, ic.RouteSwitchThreshold, int32(reply.InternalConfig.RouteSwitchThreshold))
		assert.Equal(t, ic.MaxLatencyTradeOff, int32(reply.InternalConfig.MaxLatencyTradeOff))
		assert.Equal(t, ic.RTTVeto_Default, int32(reply.InternalConfig.RTTVeto_Default))
		assert.Equal(t, ic.RTTVeto_Multipath, int32(reply.InternalConfig.RTTVeto_Multipath))
		assert.Equal(t, ic.RTTVeto_PacketLoss, int32(reply.InternalConfig.RTTVeto_PacketLoss))
		assert.Equal(t, ic.MultipathOverloadThreshold, int32(reply.InternalConfig.MultipathOverloadThreshold))
		assert.Equal(t, ic.TryBeforeYouBuy, reply.InternalConfig.TryBeforeYouBuy)
		assert.Equal(t, ic.ForceNext, reply.InternalConfig.ForceNext)
		assert.Equal(t, ic.LargeCustomer, reply.InternalConfig.LargeCustomer)
		assert.Equal(t, ic.Uncommitted, reply.InternalConfig.Uncommitted)
		assert.Equal(t, ic.MaxRTT, int32(reply.InternalConfig.MaxRTT))
		assert.Equal(t, ic.HighFrequencyPings, reply.InternalConfig.HighFrequencyPings)
		assert.Equal(t, ic.RouteDiversity, int32(reply.InternalConfig.RouteDiversity))
		assert.Equal(t, ic.MultipathThreshold, int32(reply.InternalConfig.MultipathThreshold))
		assert.Equal(t, ic.EnableVanityMetrics, reply.InternalConfig.EnableVanityMetrics)
		assert.Equal(t, ic.ReducePacketLossMinSliceNumber, int32(reply.InternalConfig.ReducePacketLossMinSliceNumber))
	})
}

func TestJSAddInternalConfig(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("invalid buyer id", func(t *testing.T) {
		var reply jsonrpc.JSAddInternalConfigReply
		err := svc.JSAddInternalConfig(req, &jsonrpc.JSAddInternalConfigArgs{BuyerID: "badBuyerID"}, &reply)
		assert.Error(t, err)
	})

	t.Run("buyer does not exist", func(t *testing.T) {
		var reply jsonrpc.JSAddInternalConfigReply
		err := svc.JSAddInternalConfig(req, &jsonrpc.JSAddInternalConfigArgs{BuyerID: "0"}, &reply)
		assert.Error(t, err)
	})

	ic := core.InternalConfig{
		RouteSelectThreshold:           2,
		RouteSwitchThreshold:           5,
		MaxLatencyTradeOff:             20,
		RTTVeto_Default:                -10,
		RTTVeto_Multipath:              -20,
		RTTVeto_PacketLoss:             -30,
		MultipathOverloadThreshold:     500,
		TryBeforeYouBuy:                false,
		ForceNext:                      false,
		LargeCustomer:                  false,
		Uncommitted:                    false,
		MaxRTT:                         300,
		HighFrequencyPings:             true,
		RouteDiversity:                 0,
		MultipathThreshold:             25,
		EnableVanityMetrics:            false,
		ReducePacketLossMinSliceNumber: 0,
	}

	jsIC := jsonrpc.JSInternalConfig{
		RouteSelectThreshold:           int64(ic.RouteSelectThreshold + 1),
		RouteSwitchThreshold:           int64(ic.RouteSwitchThreshold),
		MaxLatencyTradeOff:             int64(ic.MaxLatencyTradeOff),
		RTTVeto_Default:                int64(ic.RTTVeto_Default),
		RTTVeto_Multipath:              int64(ic.RTTVeto_Multipath),
		RTTVeto_PacketLoss:             int64(ic.RTTVeto_PacketLoss),
		MultipathOverloadThreshold:     int64(ic.MultipathOverloadThreshold),
		TryBeforeYouBuy:                ic.TryBeforeYouBuy,
		ForceNext:                      ic.ForceNext,
		LargeCustomer:                  ic.LargeCustomer,
		Uncommitted:                    ic.Uncommitted,
		MaxRTT:                         int64(ic.MaxRTT),
		HighFrequencyPings:             ic.HighFrequencyPings,
		RouteDiversity:                 int64(ic.RouteDiversity),
		MultipathThreshold:             int64(ic.MultipathThreshold),
		EnableVanityMetrics:            ic.EnableVanityMetrics,
		ReducePacketLossMinSliceNumber: int64(ic.ReducePacketLossMinSliceNumber),
	}

	err = storer.AddCustomer(context.Background(), routing.Customer{Code: "local-local", Name: "Local-Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 2, CompanyCode: "local-local", PublicKey: pubkey, InternalConfig: ic})
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var jsReply jsonrpc.JSAddInternalConfigReply
		err := svc.JSAddInternalConfig(req, &jsonrpc.JSAddInternalConfigArgs{BuyerID: "2", InternalConfig: jsIC}, &jsReply)
		assert.NoError(t, err)

		var reply jsonrpc.InternalConfigReply
		err = svc.InternalConfig(req, &jsonrpc.InternalConfigArg{BuyerID: "2"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, ic.RouteSelectThreshold+1, int32(reply.InternalConfig.RouteSelectThreshold))
		assert.Equal(t, ic.RouteSwitchThreshold, int32(reply.InternalConfig.RouteSwitchThreshold))
		assert.Equal(t, ic.MaxLatencyTradeOff, int32(reply.InternalConfig.MaxLatencyTradeOff))
		assert.Equal(t, ic.RTTVeto_Default, int32(reply.InternalConfig.RTTVeto_Default))
		assert.Equal(t, ic.RTTVeto_Multipath, int32(reply.InternalConfig.RTTVeto_Multipath))
		assert.Equal(t, ic.RTTVeto_PacketLoss, int32(reply.InternalConfig.RTTVeto_PacketLoss))
		assert.Equal(t, ic.MultipathOverloadThreshold, int32(reply.InternalConfig.MultipathOverloadThreshold))
		assert.Equal(t, ic.TryBeforeYouBuy, reply.InternalConfig.TryBeforeYouBuy)
		assert.Equal(t, ic.ForceNext, reply.InternalConfig.ForceNext)
		assert.Equal(t, ic.LargeCustomer, reply.InternalConfig.LargeCustomer)
		assert.Equal(t, ic.Uncommitted, reply.InternalConfig.Uncommitted)
		assert.Equal(t, ic.MaxRTT, int32(reply.InternalConfig.MaxRTT))
		assert.Equal(t, ic.HighFrequencyPings, reply.InternalConfig.HighFrequencyPings)
		assert.Equal(t, ic.RouteDiversity, int32(reply.InternalConfig.RouteDiversity))
		assert.Equal(t, ic.MultipathThreshold, int32(reply.InternalConfig.MultipathThreshold))
		assert.Equal(t, ic.EnableVanityMetrics, reply.InternalConfig.EnableVanityMetrics)
		assert.Equal(t, ic.ReducePacketLossMinSliceNumber, int32(reply.InternalConfig.ReducePacketLossMinSliceNumber))
	})
}

func TestUpdateInternalConfig(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("invalid hex buyer id", func(t *testing.T) {
		var reply jsonrpc.UpdateInternalConfigReply
		err := svc.UpdateInternalConfig(req, &jsonrpc.UpdateInternalConfigArgs{HexBuyerID: "badHexBuyerID"}, &reply)
		assert.Error(t, err)
	})

	t.Run("invalid buyer id", func(t *testing.T) {
		var reply jsonrpc.UpdateInternalConfigReply
		err := svc.UpdateInternalConfig(req, &jsonrpc.UpdateInternalConfigArgs{BuyerID: 0, Field: "RouteSelectThreshold", Value: "1"}, &reply)
		assert.Error(t, err)
	})

	err = storer.AddInternalConfig(context.Background(), core.NewInternalConfig(), 1)
	assert.NoError(t, err)

	int32Fields := []string{"RouteSelectThreshold", "RouteSwitchThreshold", "MaxLatencyTradeOff",
		"RTTVeto_Default", "RTTVeto_PacketLoss", "RTTVeto_Multipath",
		"MultipathOverloadThreshold", "MaxRTT", "RouteDiversity", "MultipathThreshold",
		"ReducePacketLossMinSliceNumber"}

	boolFields := []string{"TryBeforeYouBuy", "ForceNext", "LargeCustomer", "Uncommitted",
		"HighFrequencyPings", "EnableVanityMetrics"}

	t.Run("invalid int32 fields", func(t *testing.T) {
		for _, field := range int32Fields {
			var reply jsonrpc.UpdateInternalConfigReply
			err := svc.UpdateInternalConfig(req, &jsonrpc.UpdateInternalConfigArgs{BuyerID: 1, Field: field, Value: "a"}, &reply)
			assert.EqualError(t, fmt.Errorf("Value: %v is not a valid integer type", "a"), err.Error())
		}
	})

	t.Run("invalid bool fields", func(t *testing.T) {
		for _, field := range boolFields {
			var reply jsonrpc.UpdateInternalConfigReply
			err := svc.UpdateInternalConfig(req, &jsonrpc.UpdateInternalConfigArgs{BuyerID: 1, Field: field, Value: "a"}, &reply)
			assert.EqualError(t, fmt.Errorf("Value: %v is not a valid boolean type", "a"), err.Error())
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		var reply jsonrpc.UpdateInternalConfigReply
		err := svc.UpdateInternalConfig(req, &jsonrpc.UpdateInternalConfigArgs{BuyerID: 1, Field: "unknown", Value: "a"}, &reply)
		assert.EqualError(t, fmt.Errorf("Field '%v' does not exist on the InternalConfig type", "unknown"), err.Error())
	})

	t.Run("success int32 fields", func(t *testing.T) {
		for _, field := range int32Fields {
			var reply jsonrpc.UpdateInternalConfigReply
			err := svc.UpdateInternalConfig(req, &jsonrpc.UpdateInternalConfigArgs{BuyerID: 1, Field: field, Value: "1"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success bool fields", func(t *testing.T) {
		for _, field := range boolFields {
			var reply jsonrpc.UpdateInternalConfigReply
			err := svc.UpdateInternalConfig(req, &jsonrpc.UpdateInternalConfigArgs{BuyerID: 1, Field: field, Value: "true"}, &reply)
			assert.NoError(t, err)
		}
	})
}

func TestRemoveInternalConfig(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("invalid buyer id", func(t *testing.T) {
		var reply jsonrpc.RemoveInternalConfigReply
		err := svc.RemoveInternalConfig(req, &jsonrpc.RemoveInternalConfigArg{BuyerID: "badBuyerID"}, &reply)
		assert.Error(t, err)
	})

	t.Run("unknown buyer id", func(t *testing.T) {
		var reply jsonrpc.RemoveInternalConfigReply
		err := svc.RemoveInternalConfig(req, &jsonrpc.RemoveInternalConfigArg{BuyerID: "0"}, &reply)
		assert.Error(t, err)
	})

	t.Run("remove from buyer without internal config", func(t *testing.T) {
		var reply jsonrpc.RemoveInternalConfigReply
		err := svc.RemoveInternalConfig(req, &jsonrpc.RemoveInternalConfigArg{BuyerID: "1"}, &reply)
		assert.NoError(t, err)
	})

	err = storer.AddInternalConfig(context.Background(), core.NewInternalConfig(), 1)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RemoveInternalConfigReply
		err := svc.RemoveInternalConfig(req, &jsonrpc.RemoveInternalConfigArg{BuyerID: "1"}, &reply)
		assert.NoError(t, err)
	})
}

func TestRouteShader(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("invalid buyer id", func(t *testing.T) {
		var reply jsonrpc.RouteShaderReply
		err := svc.RouteShader(req, &jsonrpc.RouteShaderArg{BuyerID: "badBuyerID"}, &reply)
		assert.Error(t, err)
	})

	t.Run("no route shader", func(t *testing.T) {
		var reply jsonrpc.RouteShaderReply
		err := svc.RouteShader(req, &jsonrpc.RouteShaderArg{BuyerID: "1"}, &reply)
		assert.Error(t, err)
	})

	rs := core.NewRouteShader()

	err = storer.AddCustomer(context.Background(), routing.Customer{Code: "local-local", Name: "Local-Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 2, CompanyCode: "local-local", PublicKey: pubkey, RouteShader: rs})
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RouteShaderReply
		err := svc.RouteShader(req, &jsonrpc.RouteShaderArg{BuyerID: "2"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, rs.DisableNetworkNext, reply.RouteShader.DisableNetworkNext)
		assert.Equal(t, rs.AnalysisOnly, reply.RouteShader.AnalysisOnly)
		assert.Equal(t, rs.SelectionPercent, int(reply.RouteShader.SelectionPercent))
		assert.Equal(t, rs.ABTest, reply.RouteShader.ABTest)
		assert.Equal(t, rs.ProMode, reply.RouteShader.ProMode)
		assert.Equal(t, rs.ReduceLatency, reply.RouteShader.ReduceLatency)
		assert.Equal(t, rs.ReduceJitter, reply.RouteShader.ReduceJitter)
		assert.Equal(t, rs.ReducePacketLoss, reply.RouteShader.ReducePacketLoss)
		assert.Equal(t, rs.Multipath, reply.RouteShader.Multipath)
		assert.Equal(t, rs.AcceptableLatency, int32(reply.RouteShader.AcceptableLatency))
		assert.Equal(t, rs.LatencyThreshold, int32(reply.RouteShader.LatencyThreshold))
		assert.Equal(t, rs.AcceptablePacketLoss, float32(reply.RouteShader.AcceptablePacketLoss))
		assert.Equal(t, rs.BandwidthEnvelopeUpKbps, int32(reply.RouteShader.BandwidthEnvelopeUpKbps))
		assert.Equal(t, rs.BandwidthEnvelopeDownKbps, int32(reply.RouteShader.BandwidthEnvelopeDownKbps))
		assert.Equal(t, len(rs.BannedUsers), len(reply.RouteShader.BannedUsers))
		assert.Equal(t, rs.PacketLossSustained, float32(reply.RouteShader.PacketLossSustained))
	})
}

func TestJSAddRouteShader(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("invalid buyer id", func(t *testing.T) {
		var reply jsonrpc.JSAddRouteShaderReply
		err := svc.JSAddRouteShader(req, &jsonrpc.JSAddRouteShaderArgs{BuyerID: "badBuyerID"}, &reply)
		assert.Error(t, err)
	})

	t.Run("buyer does not exist", func(t *testing.T) {
		var reply jsonrpc.JSAddRouteShaderReply
		err := svc.JSAddRouteShader(req, &jsonrpc.JSAddRouteShaderArgs{BuyerID: "0"}, &reply)
		assert.Error(t, err)
	})

	rs := core.NewRouteShader()

	jsRS := jsonrpc.JSRouteShader{
		DisableNetworkNext:        true,
		AnalysisOnly:              true,
		SelectionPercent:          int64(rs.SelectionPercent),
		ABTest:                    rs.ABTest,
		ProMode:                   rs.ProMode,
		ReduceLatency:             rs.ReduceLatency,
		ReduceJitter:              rs.ReduceJitter,
		ReducePacketLoss:          rs.ReducePacketLoss,
		Multipath:                 rs.Multipath,
		AcceptableLatency:         int64(rs.AcceptableLatency),
		LatencyThreshold:          int64(rs.LatencyThreshold),
		AcceptablePacketLoss:      float64(rs.AcceptablePacketLoss),
		BandwidthEnvelopeUpKbps:   int64(rs.BandwidthEnvelopeUpKbps),
		BandwidthEnvelopeDownKbps: int64(rs.BandwidthEnvelopeDownKbps),
		BannedUsers:               make(map[string]bool),
		PacketLossSustained:       float64(rs.PacketLossSustained),
	}

	err = storer.AddCustomer(context.Background(), routing.Customer{Code: "local-local", Name: "Local-Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 2, CompanyCode: "local-local", PublicKey: pubkey, RouteShader: rs})
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var jsReply jsonrpc.JSAddRouteShaderReply
		err := svc.JSAddRouteShader(req, &jsonrpc.JSAddRouteShaderArgs{BuyerID: "2", RouteShader: jsRS}, &jsReply)
		assert.NoError(t, err)

		var reply jsonrpc.RouteShaderReply
		err = svc.RouteShader(req, &jsonrpc.RouteShaderArg{BuyerID: "2"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, rs.DisableNetworkNext, !reply.RouteShader.DisableNetworkNext)
		assert.Equal(t, rs.AnalysisOnly, !reply.RouteShader.AnalysisOnly)
		assert.Equal(t, rs.SelectionPercent, int(reply.RouteShader.SelectionPercent))
		assert.Equal(t, rs.ABTest, reply.RouteShader.ABTest)
		assert.Equal(t, rs.ProMode, reply.RouteShader.ProMode)
		assert.Equal(t, rs.ReduceLatency, reply.RouteShader.ReduceLatency)
		assert.Equal(t, rs.ReduceJitter, reply.RouteShader.ReduceJitter)
		assert.Equal(t, rs.ReducePacketLoss, reply.RouteShader.ReducePacketLoss)
		assert.Equal(t, rs.Multipath, reply.RouteShader.Multipath)
		assert.Equal(t, rs.AcceptableLatency, int32(reply.RouteShader.AcceptableLatency))
		assert.Equal(t, rs.LatencyThreshold, int32(reply.RouteShader.LatencyThreshold))
		assert.Equal(t, rs.AcceptablePacketLoss, float32(reply.RouteShader.AcceptablePacketLoss))
		assert.Equal(t, rs.BandwidthEnvelopeUpKbps, int32(reply.RouteShader.BandwidthEnvelopeUpKbps))
		assert.Equal(t, rs.BandwidthEnvelopeDownKbps, int32(reply.RouteShader.BandwidthEnvelopeDownKbps))
		assert.Equal(t, len(rs.BannedUsers), len(reply.RouteShader.BannedUsers))
		assert.Equal(t, rs.PacketLossSustained, float32(reply.RouteShader.PacketLossSustained))
	})
}

func TestRemoveRouteShader(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("invalid buyer id", func(t *testing.T) {
		var reply jsonrpc.RemoveRouteShaderReply
		err := svc.RemoveRouteShader(req, &jsonrpc.RemoveRouteShaderArg{BuyerID: "badBuyerID"}, &reply)
		assert.Error(t, err)
	})

	t.Run("unknown buyer id", func(t *testing.T) {
		var reply jsonrpc.RemoveRouteShaderReply
		err := svc.RemoveRouteShader(req, &jsonrpc.RemoveRouteShaderArg{BuyerID: "0"}, &reply)
		assert.Error(t, err)
	})

	t.Run("remove from buyer without route shader", func(t *testing.T) {
		var reply jsonrpc.RemoveRouteShaderReply
		err := svc.RemoveRouteShader(req, &jsonrpc.RemoveRouteShaderArg{BuyerID: "1"}, &reply)
		assert.NoError(t, err)
	})

	err = storer.AddRouteShader(context.Background(), core.NewRouteShader(), 1)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RemoveRouteShaderReply
		err := svc.RemoveRouteShader(req, &jsonrpc.RemoveRouteShaderArg{BuyerID: "1"}, &reply)
		assert.NoError(t, err)
	})
}

func TestUpdateRouteShader(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("invalid hex buyer id", func(t *testing.T) {
		var reply jsonrpc.UpdateRouteShaderReply
		err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{HexBuyerID: "badHexBuyerID"}, &reply)
		assert.Error(t, err)
	})

	t.Run("invalid buyer id", func(t *testing.T) {
		var reply jsonrpc.UpdateRouteShaderReply
		err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{BuyerID: 0, Field: "SelectionPercent", Value: "1"}, &reply)
		assert.Error(t, err)
	})

	err = storer.AddRouteShader(context.Background(), core.NewRouteShader(), 1)
	assert.NoError(t, err)

	intFields := []string{"SelectionPercent"}

	int32Fields := []string{"AcceptableLatency", "LatencyThreshold", "BandwidthEnvelopeUpKbps",
		"BandwidthEnvelopeDownKbps"}

	boolFields := []string{"AnalysisOnly", "DisableNetworkNext", "ABTest", "ProMode", "ReduceLatency",
		"ReduceJitter", "ReducePacketLoss", "Multipath"}

	float32Fields := []string{"AcceptablePacketLoss", "PacketLossSustained"}

	t.Run("invalid int fields", func(t *testing.T) {
		for _, field := range intFields {
			var reply jsonrpc.UpdateRouteShaderReply
			err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{BuyerID: 1, Field: field, Value: "a"}, &reply)
			assert.EqualError(t, fmt.Errorf("BuyersService.UpdateRouteShader Value: %v is not a valid integer type", "a"), err.Error())
		}
	})

	t.Run("invalid int32 fields", func(t *testing.T) {
		for _, field := range int32Fields {
			var reply jsonrpc.UpdateRouteShaderReply
			err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{BuyerID: 1, Field: field, Value: "a"}, &reply)
			assert.EqualError(t, fmt.Errorf("BuyersService.UpdateRouteShader Value: %v is not a valid integer type", "a"), err.Error())
		}
	})

	t.Run("invalid bool fields", func(t *testing.T) {
		for _, field := range boolFields {
			var reply jsonrpc.UpdateRouteShaderReply
			err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{BuyerID: 1, Field: field, Value: "a"}, &reply)
			assert.EqualError(t, fmt.Errorf("BuyersService.UpdateRouteShader Value: %v is not a valid boolean type", "a"), err.Error())
		}
	})

	t.Run("invalid float32 fields", func(t *testing.T) {
		for _, field := range float32Fields {
			var reply jsonrpc.UpdateRouteShaderReply
			err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{BuyerID: 1, Field: field, Value: "a"}, &reply)
			assert.EqualError(t, fmt.Errorf("BuyersService.UpdateRouteShader Value: %v is not a valid float type", "a"), err.Error())
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		var reply jsonrpc.UpdateRouteShaderReply
		err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{BuyerID: 1, Field: "unknown", Value: "a"}, &reply)
		assert.EqualError(t, fmt.Errorf("Field '%v' does not exist on the RouteShader type", "unknown"), err.Error())
	})

	t.Run("success int fields", func(t *testing.T) {
		for _, field := range intFields {
			var reply jsonrpc.UpdateRouteShaderReply
			err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{BuyerID: 1, Field: field, Value: "1"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success int32 fields", func(t *testing.T) {
		for _, field := range int32Fields {
			var reply jsonrpc.UpdateRouteShaderReply
			err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{BuyerID: 1, Field: field, Value: "1"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success bool fields", func(t *testing.T) {
		for _, field := range boolFields {
			var reply jsonrpc.UpdateRouteShaderReply
			err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{BuyerID: 1, Field: field, Value: "true"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success float32 fields", func(t *testing.T) {
		for _, field := range float32Fields {
			var reply jsonrpc.UpdateRouteShaderReply
			err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{BuyerID: 1, Field: field, Value: "1"}, &reply)
			assert.NoError(t, err)
		}
	})
}

func TestGetBannedUsers(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("unknown buyer id", func(t *testing.T) {
		var reply jsonrpc.GetBannedUserReply
		err := svc.GetBannedUsers(req, &jsonrpc.GetBannedUserArg{BuyerID: 0}, &reply)
		assert.Error(t, err)
	})

	t.Run("get from buyer without route shader", func(t *testing.T) {
		var reply jsonrpc.GetBannedUserReply
		err := svc.GetBannedUsers(req, &jsonrpc.GetBannedUserArg{BuyerID: 1}, &reply)
		assert.Error(t, err)
	})

	err = storer.AddRouteShader(context.Background(), core.NewRouteShader(), 1)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.GetBannedUserReply
		err := svc.GetBannedUsers(req, &jsonrpc.GetBannedUserArg{BuyerID: 1}, &reply)
		assert.NoError(t, err)
		assert.Zero(t, len(reply.BannedUsers))
	})
}

func TestAddBannedUser(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("unknown buyer id", func(t *testing.T) {
		var reply jsonrpc.BannedUserReply
		err := svc.AddBannedUser(req, &jsonrpc.BannedUserArgs{BuyerID: 0}, &reply)
		assert.Error(t, err)
	})

	t.Run("add banned user to buyer without route shader", func(t *testing.T) {
		var reply jsonrpc.BannedUserReply
		err := svc.AddBannedUser(req, &jsonrpc.BannedUserArgs{BuyerID: 1, UserID: 0}, &reply)
		assert.Error(t, err)
	})

	err = storer.AddRouteShader(context.Background(), core.NewRouteShader(), 1)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.BannedUserReply
		err := svc.AddBannedUser(req, &jsonrpc.BannedUserArgs{BuyerID: 1, UserID: 0}, &reply)
		assert.NoError(t, err)

		var userReply jsonrpc.GetBannedUserReply
		err = svc.GetBannedUsers(req, &jsonrpc.GetBannedUserArg{BuyerID: 1}, &userReply)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(userReply.BannedUsers))
	})
}

func TestRemoveBannedUser(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("unknown buyer id", func(t *testing.T) {
		var reply jsonrpc.BannedUserReply
		err := svc.RemoveBannedUser(req, &jsonrpc.BannedUserArgs{BuyerID: 0}, &reply)
		assert.Error(t, err)
	})

	t.Run("remove banned user from buyer without route shader", func(t *testing.T) {
		var reply jsonrpc.BannedUserReply
		err := svc.RemoveBannedUser(req, &jsonrpc.BannedUserArgs{BuyerID: 1, UserID: 0}, &reply)
		assert.Error(t, err)
	})

	err = storer.AddRouteShader(context.Background(), core.NewRouteShader(), 1)
	assert.NoError(t, err)

	t.Run("remove banned user that is not banned from buyer", func(t *testing.T) {
		var reply jsonrpc.BannedUserReply
		err := svc.RemoveBannedUser(req, &jsonrpc.BannedUserArgs{BuyerID: 1, UserID: 0}, &reply)
		assert.NoError(t, err)

		var userReply jsonrpc.GetBannedUserReply
		err = svc.GetBannedUsers(req, &jsonrpc.GetBannedUserArg{BuyerID: 1}, &userReply)
		assert.NoError(t, err)
		assert.Zero(t, len(userReply.BannedUsers))
	})

	t.Run("success", func(t *testing.T) {
		var addUserReply jsonrpc.BannedUserReply
		err := svc.AddBannedUser(req, &jsonrpc.BannedUserArgs{BuyerID: 1, UserID: 0}, &addUserReply)
		assert.NoError(t, err)

		var userReply jsonrpc.GetBannedUserReply
		err = svc.GetBannedUsers(req, &jsonrpc.GetBannedUserArg{BuyerID: 1}, &userReply)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(userReply.BannedUsers))

		var reply jsonrpc.BannedUserReply
		err = svc.RemoveBannedUser(req, &jsonrpc.BannedUserArgs{BuyerID: 1, UserID: 0}, &reply)
		assert.NoError(t, err)

		userReply = jsonrpc.GetBannedUserReply{}
		err = svc.GetBannedUsers(req, &jsonrpc.GetBannedUserArg{BuyerID: 1}, &userReply)
		assert.NoError(t, err)
		assert.Zero(t, len(userReply.BannedUsers))
	})
}

func TestBuyer(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("unknown buyer id", func(t *testing.T) {
		var reply jsonrpc.BuyerReply
		err := svc.Buyer(req, &jsonrpc.BuyerArg{BuyerID: 0}, &reply)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.BuyerReply
		err := svc.Buyer(req, &jsonrpc.BuyerArg{BuyerID: 1}, &reply)
		assert.NoError(t, err)
		assert.NotNil(t, reply.Buyer)
		assert.Equal(t, uint64(1), reply.Buyer.ID)
	})
}

func TestUpdateBuyer(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisPool := storage.NewRedisPool(redisServer.Addr(), "", 1, 1)
	var storer = storage.InMemory{}

	pubkey := make([]byte, 4)
	err := storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.BuyersService{
		RedisPoolSessionMap:    redisPool,
		RedisPoolSessionMeta:   redisPool,
		RedisPoolSessionSlices: redisPool,
		RedisPoolTopSessions:   redisPool,
		Storage:                &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.SetIsAnonymous(req, false)

	t.Run("invalid hex buyer id", func(t *testing.T) {
		var reply jsonrpc.UpdateRouteShaderReply
		err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{HexBuyerID: "badHexBuyerID"}, &reply)
		assert.Error(t, err)
	})

	t.Run("invalid buyer id", func(t *testing.T) {
		var reply jsonrpc.UpdateRouteShaderReply
		err := svc.UpdateRouteShader(req, &jsonrpc.UpdateRouteShaderArgs{BuyerID: 0, Field: "SelectionPercent", Value: "1"}, &reply)
		assert.Error(t, err)
	})

	err = storer.AddRouteShader(context.Background(), core.NewRouteShader(), 1)
	assert.NoError(t, err)

	boolFields := []string{"Live", "Debug", "Analytics", "Billing", "Trial"}

	float64Fields := []string{"ExoticLocationFee", "StandardLocationFee"}

	int64Fields := []string{"LookerSeats"}

	t.Run("invalid public key", func(t *testing.T) {
		var reply jsonrpc.UpdateBuyerReply
		err := svc.UpdateBuyer(req, &jsonrpc.UpdateBuyerArgs{BuyerID: 1, Field: "PublicKey", Value: "YFWQjOJfHfOqsCMM/1pd+c5haMhsrE2Gm05bVUQhCnG7YlPUrI/"}, &reply)
		assert.Error(t, err)
	})

	t.Run("invalid bool fields", func(t *testing.T) {
		for _, field := range boolFields {
			var reply jsonrpc.UpdateBuyerReply
			err := svc.UpdateBuyer(req, &jsonrpc.UpdateBuyerArgs{BuyerID: 1, Field: field, Value: "a"}, &reply)
			assert.EqualError(t, fmt.Errorf("BuyersService.UpdateBuyer Value: %v is not a valid boolean type", "a"), err.Error())
		}
	})

	t.Run("invalid float64 fields", func(t *testing.T) {
		for _, field := range float64Fields {
			var reply jsonrpc.UpdateBuyerReply
			err := svc.UpdateBuyer(req, &jsonrpc.UpdateBuyerArgs{BuyerID: 1, Field: field, Value: "a"}, &reply)
			assert.EqualError(t, fmt.Errorf("BuyersService.UpdateBuyer Value: %v is not a valid float64 type", "a"), err.Error())
		}
	})

	t.Run("invalid int64 fields", func(t *testing.T) {
		for _, field := range int64Fields {
			var reply jsonrpc.UpdateBuyerReply
			err := svc.UpdateBuyer(req, &jsonrpc.UpdateBuyerArgs{BuyerID: 1, Field: field, Value: "a"}, &reply)
			assert.EqualError(t, fmt.Errorf("BuyersService.UpdateBuyer Value: %v is not a valid int64 type", "a"), err.Error())
		}
	})

	t.Run("success bool fields", func(t *testing.T) {
		for _, field := range boolFields {
			var reply jsonrpc.UpdateBuyerReply
			err := svc.UpdateBuyer(req, &jsonrpc.UpdateBuyerArgs{BuyerID: 1, Field: field, Value: "true"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success float64 fields", func(t *testing.T) {
		for _, field := range float64Fields {
			var reply jsonrpc.UpdateBuyerReply
			err := svc.UpdateBuyer(req, &jsonrpc.UpdateBuyerArgs{BuyerID: 1, Field: field, Value: "1"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success int64 fields", func(t *testing.T) {
		for _, field := range int64Fields {
			var reply jsonrpc.UpdateBuyerReply
			err := svc.UpdateBuyer(req, &jsonrpc.UpdateBuyerArgs{BuyerID: 1, Field: field, Value: "1"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success short name", func(t *testing.T) {
		var reply jsonrpc.UpdateBuyerReply
		err := svc.UpdateBuyer(req, &jsonrpc.UpdateBuyerArgs{BuyerID: 1, Field: "ShortName", Value: "short-name"}, &reply)
		assert.NoError(t, err)
	})

	t.Run("success public key", func(t *testing.T) {
		var reply jsonrpc.UpdateBuyerReply
		err := svc.UpdateBuyer(req, &jsonrpc.UpdateBuyerArgs{BuyerID: 1, Field: "PublicKey", Value: "YFWQjOJfHfOqsCMM/1pd+c5haMhsrE2Gm05bVUQhCnG7YlPUrI/d1g=="}, &reply)
		assert.NoError(t, err)
	})
}

func TestDatabaseBinGeneration(t *testing.T) {
}
