package transport_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	md "github.com/networknext/backend/modules/match_data"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/test"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

// Server init handler tests

func TestServerInitHandlerSDK5Func_BuyerNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false, false)

	unknownPublicKey, unknownPrivateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	unknownBuyerID := binary.LittleEndian.Uint64(unknownPublicKey[:8])

	unknownPublicKey = unknownPublicKey[8:]
	unknownPrivateKey = unknownPrivateKey[8:]

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacketSDK5(transport.SDKVersion{5, 0, 0}, unknownBuyerID, datacenterID, fromAddr, toAddr, unknownPrivateKey)

	handler := transport.ServerInitHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseUnknownBuyer), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.BuyerNotFound.Value())
}

func TestServerInitHandlerSDK5Func_BuyerNotLive(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", false, false)

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacketSDK5(transport.SDKVersion{5, 0, 0}, buyerID, datacenterID, fromAddr, toAddr, privateKey)

	handler := transport.ServerInitHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseBuyerNotActive), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.BuyerNotActive.Value())
}

func TestServerInitHandlerSDK5Func_SigCheckFail(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacketSDK5(transport.SDKVersion{5, 0, 0}, buyerID, datacenterID, fromAddr, toAddr, privateKey[2:])

	handler := transport.ServerInitHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseSignatureCheckFailed), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.SignatureCheckFailed.Value())
}

func TestServerInitHandlerSDK5Func_SDKTooOld(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacketSDK5(transport.SDKVersion{4, 2, 0}, buyerID, datacenterID, fromAddr, toAddr, privateKey)

	handler := transport.ServerInitHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseOldSDKVersion), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.SDKTooOld.Value())
}

func TestServerInitHandlerSDK5Func_Success_DatacenterNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacketSDK5(transport.SDKVersion{5, 0, 0}, buyerID, datacenterID, fromAddr, toAddr, privateKey)

	handler := transport.ServerInitHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.DatacenterNotFound.Value())
}

func TestServerInitHandlerSDK5Func_Success(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacketSDK5(transport.SDKVersion{5, 0, 0}, buyerID, datacenter.ID, fromAddr, toAddr, privateKey)

	handler := transport.ServerInitHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)
	assert.Zero(t, metrics.ServerInitMetrics.DatacenterNotFound.Value())
}

func TestServerInitHandlerSDK5Func_ServerTracker_DatacenterNotFound_WithoutName(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := ""
	datacenterID := crypto.HashID(datacenterName)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacketSDK5(transport.SDKVersion{5, 0, 0}, buyerID, datacenterID, fromAddr, toAddr, privateKey)

	handler := transport.ServerInitHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.DatacenterNotFound.Value())

	assert.NotEmpty(t, serverTracker.Tracker)

	buyerHexID := fmt.Sprintf("%016x", buyerID)
	for _, serverInfo := range serverTracker.Tracker[buyerHexID] {
		assert.Equal(t, serverInfo.DatacenterID, fmt.Sprintf("%016x", datacenterID))
		assert.Equal(t, serverInfo.DatacenterName, "unknown_init")
	}
}

func TestServerInitHandlerSDK5Func_ServerTracker_DatacenterFound_WithName(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacketSDK5(transport.SDKVersion{5, 0, 0}, buyerID, datacenter.ID, fromAddr, toAddr, privateKey)

	handler := transport.ServerInitHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)
	assert.Zero(t, metrics.ServerInitMetrics.DatacenterNotFound.Value())

	assert.NotEmpty(t, serverTracker.Tracker)

	buyerHexID := fmt.Sprintf("%016x", buyerID)
	for _, serverInfo := range serverTracker.Tracker[buyerHexID] {
		assert.Equal(t, serverInfo.DatacenterID, fmt.Sprintf("%016x", datacenter.ID))
		assert.Equal(t, serverInfo.DatacenterName, datacenter.Name)
	}
}

// Server update handler tests

func TestServerUpdateHandlerSDK5Func_BuyerNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false, false)

	unknownPublicKey, unknownPrivateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	unknownBuyerID := binary.LittleEndian.Uint64(unknownPublicKey[:8])

	unknownPublicKey = unknownPublicKey[8:]
	unknownPrivateKey = unknownPrivateKey[8:]

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerUpdatePacketSDK5(transport.SDKVersion{5, 0, 0}, unknownBuyerID, datacenterID, 10, fromAddr, toAddr, unknownPrivateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, postSessionHandler, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	assert.Zero(t, responseBuffer.Bytes())
	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.BuyerNotFound.Value())
}

func TestServerUpdateHandlerSDK5Func_BuyerNotLive(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", false, false)

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerUpdatePacketSDK5(transport.SDKVersion{5, 0, 0}, buyerID, datacenterID, 10, fromAddr, toAddr, privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, postSessionHandler, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	assert.Zero(t, responseBuffer.Bytes())
	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.BuyerNotLive.Value())
}

func TestServerUpdateHandlerSDK5Func_SigCheckFail(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerUpdatePacketSDK5(transport.SDKVersion{5, 0, 0}, buyerID, datacenterID, 10, fromAddr, toAddr, privateKey[2:])

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, postSessionHandler, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	assert.Zero(t, responseBuffer.Bytes())
	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.SignatureCheckFailed.Value())
}

func TestServerUpdateHandlerSDK5Func_SDKTooOld(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerUpdatePacketSDK5(transport.SDKVersion{4, 20, 0}, buyerID, datacenterID, 10, fromAddr, toAddr, privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, postSessionHandler, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	// var responsePacket transport.ServerResponsePacketSDK5
	// err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	// assert.NoError(t, err)
	assert.Zero(t, responseBuffer.Bytes())
	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.SDKTooOld.Value())
}

func TestServerUpdateHandlerSDK5Func_DatacenterNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerUpdatePacketSDK5(transport.SDKVersion{5, 0, 0}, buyerID, datacenterID, 10, fromAddr, toAddr, privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, postSessionHandler, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	var responsePacket transport.ServerResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.DatacenterNotFound.Value())
}

func TestServerUpdateHandlerSDK5Func_Success(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerUpdatePacketSDK5(transport.SDKVersion{5, 0, 0}, buyerID, datacenter.ID, 10, fromAddr, toAddr, privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, postSessionHandler, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	var responsePacket transport.ServerResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Zero(t, metrics.ServerUpdateMetrics.DatacenterNotFound.Value())
}

func TestServerUpdateHandlerSDK5Func_ServerTracker_DatacenterNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerUpdatePacketSDK5(transport.SDKVersion{5, 0, 0}, buyerID, datacenterID, 10, fromAddr, toAddr, privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, postSessionHandler, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	var responsePacket transport.ServerResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.DatacenterNotFound.Value())

	buyerHexID := fmt.Sprintf("%016x", buyerID)
	for _, serverInfo := range serverTracker.Tracker[buyerHexID] {
		assert.Equal(t, serverInfo.DatacenterID, fmt.Sprintf("%016x", datacenterID))
		assert.Equal(t, serverInfo.DatacenterName, "unknown_update")
	}
}

func TestServerUpdateHandlerSDK5Func_ServerTracker_DatacenterFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerUpdatePacketSDK5(transport.SDKVersion{5, 0, 0}, buyerID, datacenter.ID, 10, fromAddr, toAddr, privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, postSessionHandler, serverTracker, env.GetBackendLoadBalancerIP(), crypto.BackendPrivateKey, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *fromAddr,
		Data: requestData,
	})

	var responsePacket transport.ServerResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Zero(t, metrics.ServerUpdateMetrics.DatacenterNotFound.Value())

	buyerHexID := fmt.Sprintf("%016x", buyerID)
	for _, serverInfo := range serverTracker.Tracker[buyerHexID] {
		assert.Equal(t, serverInfo.DatacenterID, fmt.Sprintf("%016x", datacenter.ID))
		assert.Equal(t, serverInfo.DatacenterName, datacenter.Name)
	}
}

// Session update handler

func TestSessionUpdateHandlerSDK5Func_Pre_BuyerNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false, false)

	unknownPublicKey, _, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	unknownBuyerID := binary.LittleEndian.Uint64(unknownPublicKey[:8])

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = unknownBuyerID

	assert.True(t, transport.SessionPreSDK5(&state))

	assert.True(t, state.BuyerNotFound)
	assert.Equal(t, float64(1), state.Metrics.BuyerNotFound.Value())
}

func TestSessionUpdateHandlerSDK5Func_Pre_BuyerNotLive(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, _ := env.AddBuyer("local", false, false)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID

	assert.True(t, transport.SessionPreSDK5(&state))

	assert.True(t, state.BuyerNotLive)
	assert.Equal(t, float64(1), state.Metrics.BuyerNotLive.Value())
}

func TestSessionUpdateHandlerSDK5Func_Pre_SigCheckFail(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey[2:])
	state.PacketData = requestData

	assert.True(t, transport.SessionPreSDK5(&state))

	assert.True(t, state.SignatureCheckFailed)
	assert.Equal(t, float64(1), state.Metrics.SignatureCheckFailed.Value())
}

func TestSessionUpdateHandlerSDK5Func_Pre_ClientTimedOut(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.Packet.ClientPingTimedOut = true

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	assert.True(t, transport.SessionPreSDK5(&state))

	assert.True(t, state.Packet.ClientPingTimedOut)
	assert.Equal(t, float64(1), state.Metrics.ClientPingTimedOut.Value())
}

func TestSessionUpdateHandlerSDK5Func_Pre_LocationVeto(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.IpLocator = getErrorLocator()

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	assert.True(t, transport.SessionPreSDK5(&state))

	assert.True(t, state.Output.RouteState.LocationVeto)
	assert.Equal(t, float64(1), state.Metrics.ClientLocateFailure.Value())
}

func TestSessionUpdateHandlerSDK5Func_Pre_DatacenterNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = crypto.HashID("unknown.datacenter.name")
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	assert.True(t, transport.SessionPreSDK5(&state))

	assert.True(t, state.UnknownDatacenter)
	assert.Equal(t, float64(1), state.Metrics.DatacenterNotFound.Value())
}

func TestSessionUpdateHandlerSDK5Func_Pre_DatacenterNotFound_AnalysisOnly(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, true)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = crypto.HashID("unknown.datacenter.name")
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.RouteMatrix = env.GetRouteMatrix()
	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	assert.False(t, transport.SessionPreSDK5(&state))

	assert.True(t, state.UnknownDatacenter)
	assert.Equal(t, float64(1), state.Metrics.DatacenterNotFound.Value())
	assert.Equal(t, routing.UnknownDatacenter, state.Datacenter)
}

func TestSessionUpdateHandlerSDK5Func_Pre_DatacenterNotEnabled(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.RouteMatrix = env.GetRouteMatrix()
	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	assert.True(t, transport.SessionPreSDK5(&state))

	assert.True(t, state.DatacenterNotEnabled)
	assert.Equal(t, float64(1), state.Metrics.DatacenterNotEnabled.Value())
}

func TestSessionUpdateHandlerSDK5Func_Pre_DatacenterNotEnabled_AnalysisOnly(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, true)
	datacenter := env.AddDatacenter("datacenter.name")

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.RouteMatrix = env.GetRouteMatrix()
	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	assert.False(t, transport.SessionPreSDK5(&state))

	assert.False(t, state.DatacenterNotEnabled)
	assert.Equal(t, float64(0), state.Metrics.DatacenterNotEnabled.Value())
	assert.Equal(t, datacenter, state.Datacenter)
}

func TestSessionUpdateHandlerSDK5Func_Pre_NoRelaysInDatacenter(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name, true)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.RouteMatrix = env.GetRouteMatrix()
	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	assert.True(t, transport.SessionPreSDK5(&state))

	assert.Equal(t, float64(1), state.Metrics.NoRelaysInDatacenter.Value())
}

func TestSessionUpdateHandlerSDK5Func_Pre_NoRelaysInDatacenter_AnalysisOnly(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, true)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name, true)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.RouteMatrix = env.GetRouteMatrix()
	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	assert.False(t, transport.SessionPreSDK5(&state))

	assert.Equal(t, float64(0), state.Metrics.NoRelaysInDatacenter.Value())
	assert.Equal(t, datacenter, state.Datacenter)
}

func TestSessionUpdateHandlerSDK5Func_Pre_NoRelaysInDatacenter_DatacenterAccelerationDisabled(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name, false)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.RouteMatrix = env.GetRouteMatrix()

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	assert.False(t, transport.SessionPreSDK5(&state))

	assert.Equal(t, float64(0), state.Metrics.NoRelaysInDatacenter.Value())
	assert.Equal(t, datacenter, state.Datacenter)
}

func TestSessionUpdateHandlerSDK5Func_Pre_StaleRouteMatrix(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name, true)
	env.AddRelay("losangeles.1", "10.0.0.2", datacenter.ID)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.StaleDuration = time.Second * 20
	state.RouteMatrix = env.GetRouteMatrix()
	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	// Make the route matrix creation time older by 30 seconds
	state.RouteMatrix.CreatedAt = uint64(time.Now().Add(-(time.Second * 30)).Unix())

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	assert.True(t, transport.SessionPreSDK5(&state))

	assert.True(t, state.StaleRouteMatrix)
	assert.Equal(t, float64(1), state.Metrics.StaleRouteMatrix.Value())
	assert.Equal(t, datacenter, state.Datacenter)
}

func TestSessionUpdateHandlerSDK5Func_Pre_Successs(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name, true)
	env.AddRelay("losangeles.1", "10.0.0.2", datacenter.ID)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.StaleDuration = time.Second * 20
	state.RouteMatrix = env.GetRouteMatrix()
	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	assert.False(t, transport.SessionPreSDK5(&state))
	assert.Equal(t, datacenter, state.Datacenter)
}

func TestSessionUpdateHandlerSDK5Func_Pre_Successs_AnalysisOnly(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, true)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name, true)
	env.AddRelay("losangeles.1", "10.0.0.2", datacenter.ID)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.StaleDuration = time.Second * 20
	state.RouteMatrix = env.GetRouteMatrix()
	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	assert.False(t, transport.SessionPreSDK5(&state))
	assert.False(t, state.UnknownDatacenter)
	assert.False(t, state.DatacenterNotEnabled)
	assert.Equal(t, float64(0), state.Metrics.DatacenterNotEnabled.Value())
	assert.Equal(t, float64(0), state.Metrics.NoRelaysInDatacenter.Value())
	assert.Equal(t, datacenter, state.Datacenter)
}

func TestSessionUpdateHandlerSDK5Func_Pre_Success_DatacenterAccelerationDisabled(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name, false)
	env.AddRelay("losangeles.1", "10.0.0.2", datacenter.ID)

	state := transport.SessionHandlerStateSDK5{}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.StaleDuration = time.Second * 20
	state.RouteMatrix = env.GetRouteMatrix()
	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	requestData := env.GenerateEmptySessionUpdatePacketSDK5(fromAddr, toAddr, privateKey)
	state.PacketData = requestData

	assert.False(t, transport.SessionPreSDK5(&state))
	assert.False(t, state.UnknownDatacenter)
	assert.False(t, state.DatacenterNotEnabled)
	assert.Equal(t, float64(0), state.Metrics.DatacenterNotEnabled.Value())
	assert.Equal(t, float64(0), state.Metrics.NoRelaysInDatacenter.Value())
	assert.Equal(t, datacenter, state.Datacenter)
}

func TestSessionUpdateHandlerSDK5Func_NewSession_Success(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerStateSDK5{
		Metrics: metrics.SessionUpdateMetrics,
		Output: transport.SessionDataSDK5{
			SliceNumber: 0,
		},
	}
	transport.SessionUpdateNewSessionSDK5(&state)

	assert.Equal(t, uint32(1), state.Output.SliceNumber)
	assert.Equal(t, state.Output, state.Input)
}

func TestSessionUpdateHandlerSDK5Func_ExistingSession_BadSessionID(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	sessionData := transport.SessionDataSDK5{
		SessionID: uint64(123456789),
	}

	var sessionDataBytesFixed [511]byte
	sessionDataBytes, err := transport.MarshalSessionDataSDK5(&sessionData)
	assert.NoError(t, err)

	copy(sessionDataBytesFixed[:], sessionDataBytes)

	state := transport.SessionHandlerStateSDK5{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacketSDK5{
			SessionID:   uint64(9876543120),
			SessionData: sessionDataBytesFixed,
		},
	}

	transport.SessionUpdateExistingSessionSDK5(&state)

	assert.Equal(t, float64(1), state.Metrics.BadSessionID.Value())
}

func TestSessionUpdateHandlerSDK5Func_ExistingSession_BadSliceNumber(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	sessionData := transport.SessionDataSDK5{
		SessionID:   uint64(123456789),
		SliceNumber: 23,
	}

	var sessionDataBytesFixed [511]byte
	sessionDataBytes, err := transport.MarshalSessionDataSDK5(&sessionData)
	assert.NoError(t, err)

	copy(sessionDataBytesFixed[:], sessionDataBytes)

	state := transport.SessionHandlerStateSDK5{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacketSDK5{
			SessionID:   uint64(123456789),
			SliceNumber: 199,
			SessionData: sessionDataBytesFixed,
		},
	}

	transport.SessionUpdateExistingSessionSDK5(&state)

	assert.Equal(t, float64(1), state.Metrics.BadSliceNumber.Value())
}

func TestSessionUpdateHandlerSDK5Func_ExistingSession_Success(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	sessionData := transport.SessionDataSDK5{
		SessionID:   uint64(123456789),
		SliceNumber: 1,
		Initial:     true,
	}

	var sessionDataBytesFixed [511]byte
	sessionDataBytes, err := transport.MarshalSessionDataSDK5(&sessionData)
	assert.NoError(t, err)

	copy(sessionDataBytesFixed[:], sessionDataBytes)

	state := transport.SessionHandlerStateSDK5{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacketSDK5{
			SessionID:   uint64(123456789),
			SliceNumber: 1,
			SessionData: sessionDataBytesFixed,
		},
	}

	transport.SessionUpdateExistingSessionSDK5(&state)

	assert.False(t, state.Output.Initial)
	assert.Equal(t, uint32(2), state.Output.SliceNumber)
}

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_NoFallback(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerStateSDK5{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacketSDK5{
			FallbackToDirect: false,
		},
		Output: transport.SessionDataSDK5{
			FellBackToDirect: false,
		},
	}

	assert.False(t, transport.SessionHandleFallbackToDirectSDK5(&state))
}

// TODO: implement these unit tests when SDK5 supports fallback to direct flags
/*

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_BadRouteToken(t *testing.T) {
    t.Parallel()

    metricsHandler := metrics.LocalHandler{}
    metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
    assert.NoError(t, err)

    state := transport.SessionHandlerStateSDK5{
        Metrics: metrics.SessionUpdateMetrics,
        Packet: transport.SessionUpdatePacketSDK5{
            FallbackToDirect: true,
            Flags:            (1 << 0),
        },
        Output: transport.SessionDataSDK5{
            FellBackToDirect: false,
        },
    }

    assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
    assert.Equal(t, float64(1), state.Metrics.FallbackToDirectBadRouteToken.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_NoNextRouteToContinue(t *testing.T) {
    t.Parallel()

    metricsHandler := metrics.LocalHandler{}
    metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
    assert.NoError(t, err)

    state := transport.SessionHandlerStateSDK5{
        Metrics: metrics.SessionUpdateMetrics,
        Packet: transport.SessionUpdatePacketSDK5{
            FallbackToDirect: true,
            Flags:            (1 << 1),
        },
        Output: transport.SessionDataSDK5{
            FellBackToDirect: false,
        },
    }

    assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
    assert.Equal(t, float64(1), state.Metrics.FallbackToDirectNoNextRouteToContinue.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_PreviousUpdateStillPending(t *testing.T) {
    t.Parallel()

    metricsHandler := metrics.LocalHandler{}
    metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
    assert.NoError(t, err)

    state := transport.SessionHandlerStateSDK5{
        Metrics: metrics.SessionUpdateMetrics,
        Packet: transport.SessionUpdatePacketSDK5{
            FallbackToDirect: true,
            Flags:            (1 << 2),
        },
        Output: transport.SessionDataSDK5{
            FellBackToDirect: false,
        },
    }

    assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
    assert.Equal(t, float64(1), state.Metrics.FallbackToDirectPreviousUpdateStillPending.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_BadContinueToken(t *testing.T) {
    t.Parallel()

    metricsHandler := metrics.LocalHandler{}
    metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
    assert.NoError(t, err)

    state := transport.SessionHandlerStateSDK5{
        Metrics: metrics.SessionUpdateMetrics,
        Packet: transport.SessionUpdatePacketSDK5{
            FallbackToDirect: true,
            Flags:            (1 << 3),
        },
        Output: transport.SessionDataSDK5{
            FellBackToDirect: false,
        },
    }

    assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
    assert.Equal(t, float64(1), state.Metrics.FallbackToDirectBadContinueToken.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_RouteExpired(t *testing.T) {
    t.Parallel()

    metricsHandler := metrics.LocalHandler{}
    metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
    assert.NoError(t, err)

    state := transport.SessionHandlerStateSDK5{
        Metrics: metrics.SessionUpdateMetrics,
        Packet: transport.SessionUpdatePacketSDK5{
            FallbackToDirect: true,
            Flags:            (1 << 4),
        },
        Output: transport.SessionDataSDK5{
            FellBackToDirect: false,
        },
    }

    assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
    assert.Equal(t, float64(1), state.Metrics.FallbackToDirectRouteExpired.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_RouteRequestTimedOut(t *testing.T) {
    t.Parallel()

    metricsHandler := metrics.LocalHandler{}
    metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
    assert.NoError(t, err)

    state := transport.SessionHandlerStateSDK5{
        Metrics: metrics.SessionUpdateMetrics,
        Packet: transport.SessionUpdatePacketSDK5{
            FallbackToDirect: true,
            Flags:            (1 << 5),
        },
        Output: transport.SessionDataSDK5{
            FellBackToDirect: false,
        },
    }

    assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
    assert.Equal(t, float64(1), state.Metrics.FallbackToDirectRouteRequestTimedOut.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_ContinueRequestTimedOut(t *testing.T) {
    t.Parallel()

    metricsHandler := metrics.LocalHandler{}
    metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
    assert.NoError(t, err)

    state := transport.SessionHandlerStateSDK5{
        Metrics: metrics.SessionUpdateMetrics,
        Packet: transport.SessionUpdatePacketSDK5{
            FallbackToDirect: true,
            Flags:            (1 << 6),
        },
        Output: transport.SessionDataSDK5{
            FellBackToDirect: false,
        },
    }

    assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
    assert.Equal(t, float64(1), state.Metrics.FallbackToDirectContinueRequestTimedOut.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_ClientTimedOut(t *testing.T) {
    t.Parallel()

    metricsHandler := metrics.LocalHandler{}
    metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
    assert.NoError(t, err)

    state := transport.SessionHandlerStateSDK5{
        Metrics: metrics.SessionUpdateMetrics,
        Packet: transport.SessionUpdatePacketSDK5{
            FallbackToDirect: true,
            Flags:            (1 << 7),
        },
        Output: transport.SessionDataSDK5{
            FellBackToDirect: false,
        },
    }

    assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
    assert.Equal(t, float64(1), state.Metrics.FallbackToDirectClientTimedOut.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_UpgradeResponseTimedOut(t *testing.T) {
    t.Parallel()

    metricsHandler := metrics.LocalHandler{}
    metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
    assert.NoError(t, err)

    state := transport.SessionHandlerStateSDK5{
        Metrics: metrics.SessionUpdateMetrics,
        Packet: transport.SessionUpdatePacketSDK5{
            FallbackToDirect: true,
            Flags:            (1 << 8),
        },
        Output: transport.SessionDataSDK5{
            FellBackToDirect: false,
        },
    }

    assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
    assert.Equal(t, float64(1), state.Metrics.FallbackToDirectUpgradeResponseTimedOut.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_RouteUpdateTimedOut(t *testing.T) {
    t.Parallel()

    metricsHandler := metrics.LocalHandler{}
    metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
    assert.NoError(t, err)

    state := transport.SessionHandlerStateSDK5{
        Metrics: metrics.SessionUpdateMetrics,
        Packet: transport.SessionUpdatePacketSDK5{
            FallbackToDirect: true,
            Flags:            (1 << 9),
        },
        Output: transport.SessionDataSDK5{
            FellBackToDirect: false,
        },
    }

    assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
    assert.Equal(t, float64(1), state.Metrics.FallbackToDirectRouteUpdateTimedOut.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_DirectPongTimedOut(t *testing.T) {
    t.Parallel()

    metricsHandler := metrics.LocalHandler{}
    metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
    assert.NoError(t, err)

    state := transport.SessionHandlerStateSDK5{
        Metrics: metrics.SessionUpdateMetrics,
        Packet: transport.SessionUpdatePacketSDK5{
            FallbackToDirect: true,
            Flags:            (1 << 10),
        },
        Output: transport.SessionDataSDK5{
            FellBackToDirect: false,
        },
    }

    assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
    assert.Equal(t, float64(1), state.Metrics.FallbackToDirectDirectPongTimedOut.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_NextPongTimedOut(t *testing.T) {
    t.Parallel()

    metricsHandler := metrics.LocalHandler{}
    metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
    assert.NoError(t, err)

    state := transport.SessionHandlerStateSDK5{
        Metrics: metrics.SessionUpdateMetrics,
        Packet: transport.SessionUpdatePacketSDK5{
            FallbackToDirect: true,
            Flags:            (1 << 11),
        },
        Output: transport.SessionDataSDK5{
            FellBackToDirect: false,
        },
    }

    assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
    assert.Equal(t, float64(1), state.Metrics.FallbackToDirectNextPongTimedOut.Value())
}

*/

func TestSessionUpdateHandlerSDK5Func_SessionHandleFallbackToDirect_Unknown(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerStateSDK5{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacketSDK5{
			FallbackToDirect: true,
		},
		Output: transport.SessionDataSDK5{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirectSDK5(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectUnknownReason.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionUpdateNearRelayStats_AnalysisOnly(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerStateSDK5{
		Metrics: metrics.SessionUpdateMetrics,
	}

	state.Datacenter = routing.Datacenter{
		ID:        crypto.HashID("datacenter.name"),
		Name:      "datacenter.name",
		AliasName: "datacenter.name",
	}

	state.RouteMatrix = &routing.RouteMatrix{
		RelayDatacenterIDs: []uint64{
			12345,
			123423,
			12351321,
		},
	}

	state.Buyer = routing.Buyer{
		RouteShader: core.RouteShader{
			AnalysisOnly: true,
		},
	}

	assert.False(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
	assert.Zero(t, state.Metrics.NoRelaysInDatacenter.Value())
	assert.Zero(t, len(state.Packet.NearRelayIDs))
}

func TestSessionUpdateHandlerSDK5Func_SessionUpdateNearRelayStats_DatacenterAccelerationDisabled(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerStateSDK5{
		Metrics: metrics.SessionUpdateMetrics,
	}

	state.Datacenter = routing.Datacenter{
		ID:        crypto.HashID("datacenter.name"),
		Name:      "datacenter.name",
		AliasName: "datacenter.name",
	}

	state.RouteMatrix = &routing.RouteMatrix{
		RelayDatacenterIDs: []uint64{
			12345,
			123423,
			12351321,
		},
	}

	state.DatacenterAccelerationEnabled = false

	assert.False(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
	assert.Equal(t, float64(0), state.Metrics.NoRelaysInDatacenter.Value())
	assert.Zero(t, len(state.Packet.NearRelayIDs))
}

func TestSessionUpdateHandlerSDK5Func_SessionUpdateNearRelayStats_NoRelaysInDatacenter(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerStateSDK5{
		Metrics:                       metrics.SessionUpdateMetrics,
		DatacenterAccelerationEnabled: true,
	}

	state.Datacenter = routing.Datacenter{
		ID:        crypto.HashID("datacenter.name"),
		Name:      "datacenter.name",
		AliasName: "datacenter.name",
	}

	state.RouteMatrix = &routing.RouteMatrix{
		RelayDatacenterIDs: []uint64{
			12345,
			123423,
			12351321,
		},
	}

	assert.False(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
	assert.Equal(t, float64(1), state.Metrics.NoRelaysInDatacenter.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionUpdateNearRelayStats_HoldNearRelays(t *testing.T) {
	t.Parallel()

	t.Run("Large Customer Transition false -> true before slice 4", func(t *testing.T) {
		updatePacket := transport.SessionUpdatePacketSDK5{
			SliceNumber: uint32(2),
		}

		metricsHandler := metrics.LocalHandler{}
		metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
		assert.NoError(t, err)

		buyer := routing.Buyer{
			InternalConfig: core.InternalConfig{
				LargeCustomer: false,
			},
		}

		dc := routing.Datacenter{
			ID:        crypto.HashID("datacenter.name"),
			Name:      "datacenter.name",
			AliasName: "datacenter.name",
		}

		routeMatrix := &routing.RouteMatrix{
			RelayIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
			RelayDatacenterIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
		}

		state := transport.SessionHandlerStateSDK5{
			Packet:                        updatePacket,
			Metrics:                       metrics.SessionUpdateMetrics,
			RouteMatrix:                   routeMatrix,
			Datacenter:                    dc,
			Buyer:                         buyer,
			DatacenterAccelerationEnabled: true,
		}

		assert.False(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
		assert.False(t, state.Output.HoldNearRelays)
		assert.False(t, state.Response.DontPingNearRelays)

		state.Packet.SliceNumber++
		state.Buyer.InternalConfig.LargeCustomer = true

		assert.True(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
		assert.False(t, state.Output.HoldNearRelays)
		assert.False(t, state.Response.DontPingNearRelays)
	})

	t.Run("Large Customer Transition true -> false before slice 4", func(t *testing.T) {
		updatePacket := transport.SessionUpdatePacketSDK5{
			SliceNumber: uint32(2),
		}

		metricsHandler := metrics.LocalHandler{}
		metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
		assert.NoError(t, err)

		buyer := routing.Buyer{
			InternalConfig: core.InternalConfig{
				LargeCustomer: true,
			},
		}

		dc := routing.Datacenter{
			ID:        crypto.HashID("datacenter.name"),
			Name:      "datacenter.name",
			AliasName: "datacenter.name",
		}

		routeMatrix := &routing.RouteMatrix{
			RelayIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
			RelayDatacenterIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
		}

		state := transport.SessionHandlerStateSDK5{
			Packet:                        updatePacket,
			Metrics:                       metrics.SessionUpdateMetrics,
			RouteMatrix:                   routeMatrix,
			Datacenter:                    dc,
			Buyer:                         buyer,
			DatacenterAccelerationEnabled: true,
		}

		assert.True(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
		assert.False(t, state.Output.HoldNearRelays)
		assert.False(t, state.Response.DontPingNearRelays)

		state.Packet.SliceNumber++
		state.Buyer.InternalConfig.LargeCustomer = false

		assert.False(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
		assert.False(t, state.Output.HoldNearRelays)
		assert.False(t, state.Response.DontPingNearRelays)
	})

	t.Run("Large Customer Transition false -> true on or after slice 4", func(t *testing.T) {
		rand.Seed(time.Now().Unix())

		updatePacket := transport.SessionUpdatePacketSDK5{
			SliceNumber: uint32(4),
			NearRelayIDs: []uint64{
				rand.Uint64(),
				rand.Uint64(),
				rand.Uint64(),
			},
			NearRelayRTT: []int32{
				rand.Int31n(255),
				rand.Int31n(255),
				rand.Int31n(255),
			},
			NearRelayJitter: []int32{
				rand.Int31n(255),
				rand.Int31n(255),
				rand.Int31n(255),
			},
			NearRelayPacketLoss: []int32{
				rand.Int31n(100),
				rand.Int31n(100),
				rand.Int31n(100),
			},
		}

		metricsHandler := metrics.LocalHandler{}
		metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
		assert.NoError(t, err)

		buyer := routing.Buyer{
			InternalConfig: core.InternalConfig{
				LargeCustomer: false,
			},
		}

		dc := routing.Datacenter{
			ID:        crypto.HashID("datacenter.name"),
			Name:      "datacenter.name",
			AliasName: "datacenter.name",
		}

		routeMatrix := &routing.RouteMatrix{
			RelayIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
			RelayDatacenterIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
		}

		state := transport.SessionHandlerStateSDK5{
			Packet:                        updatePacket,
			Metrics:                       metrics.SessionUpdateMetrics,
			RouteMatrix:                   routeMatrix,
			Datacenter:                    dc,
			Buyer:                         buyer,
			DatacenterAccelerationEnabled: true,
		}

		assert.False(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
		assert.False(t, state.Output.HoldNearRelays)
		assert.False(t, state.Response.DontPingNearRelays)

		state.Packet.SliceNumber++
		state.Buyer.InternalConfig.LargeCustomer = true
		state.Input = state.Output

		assert.True(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
		assert.True(t, state.Output.HoldNearRelays)
		assert.True(t, state.Response.DontPingNearRelays)

		updatePacket = transport.SessionUpdatePacketSDK5{
			SliceNumber: state.Packet.SliceNumber + 1,
			NearRelayIDs: []uint64{
				rand.Uint64(),
				rand.Uint64(),
				rand.Uint64(),
			},
			NearRelayRTT: []int32{
				rand.Int31n(255),
				rand.Int31n(255),
				rand.Int31n(255),
			},
			NearRelayJitter: []int32{
				rand.Int31n(255),
				rand.Int31n(255),
				rand.Int31n(255),
			},
			NearRelayPacketLoss: []int32{
				rand.Int31n(100),
				rand.Int31n(100),
				rand.Int31n(100),
			},
		}

		state.Packet = updatePacket
		state.Input = state.Output

		assert.True(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
		assert.True(t, state.Output.HoldNearRelays)
		assert.True(t, state.Response.DontPingNearRelays)
		assert.Equal(t, state.Packet.NearRelayIDs, updatePacket.NearRelayIDs)
		for i := 0; i < len(state.Packet.NearRelayIDs); i++ {
			assert.Equal(t, state.Packet.NearRelayRTT[i], state.Input.HoldNearRelayRTT[i])
			assert.Equal(t, state.Packet.NearRelayRTT[i], updatePacket.NearRelayRTT[i])
		}
	})

	t.Run("Large Customer Transition true -> false on or after slice 4", func(t *testing.T) {
		rand.Seed(time.Now().Unix())

		updatePacket := transport.SessionUpdatePacketSDK5{
			SliceNumber: uint32(4),
			NearRelayIDs: []uint64{
				rand.Uint64(),
				rand.Uint64(),
				rand.Uint64(),
			},
			NearRelayRTT: []int32{
				rand.Int31n(255),
				rand.Int31n(255),
				rand.Int31n(255),
			},
			NearRelayJitter: []int32{
				rand.Int31n(255),
				rand.Int31n(255),
				rand.Int31n(255),
			},
			NearRelayPacketLoss: []int32{
				rand.Int31n(100),
				rand.Int31n(100),
				rand.Int31n(100),
			},
		}

		metricsHandler := metrics.LocalHandler{}
		metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
		assert.NoError(t, err)

		buyer := routing.Buyer{
			InternalConfig: core.InternalConfig{
				LargeCustomer: true,
			},
		}

		dc := routing.Datacenter{
			ID:        crypto.HashID("datacenter.name"),
			Name:      "datacenter.name",
			AliasName: "datacenter.name",
		}

		routeMatrix := &routing.RouteMatrix{
			RelayIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
			RelayDatacenterIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
		}

		state := transport.SessionHandlerStateSDK5{
			Packet:                        updatePacket,
			Metrics:                       metrics.SessionUpdateMetrics,
			RouteMatrix:                   routeMatrix,
			Datacenter:                    dc,
			Buyer:                         buyer,
			DatacenterAccelerationEnabled: true,
		}

		assert.True(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
		assert.True(t, state.Output.HoldNearRelays)
		assert.True(t, state.Response.DontPingNearRelays)

		state.Packet.SliceNumber++
		state.Buyer.InternalConfig.LargeCustomer = false
		state.Input = state.Output

		assert.False(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
		assert.True(t, state.Output.HoldNearRelays)
		assert.True(t, state.Response.DontPingNearRelays)
		assert.Equal(t, state.Packet.NearRelayIDs, updatePacket.NearRelayIDs)
		for i := 0; i < len(state.Packet.NearRelayIDs); i++ {
			assert.Equal(t, state.Packet.NearRelayRTT[i], state.Input.HoldNearRelayRTT[i])
			assert.Equal(t, state.Packet.NearRelayRTT[i], updatePacket.NearRelayRTT[i])
		}

		state.Packet.SliceNumber++
		state.Input = state.Output

		assert.False(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
		assert.True(t, state.Output.HoldNearRelays)
		assert.True(t, state.Response.DontPingNearRelays)
		assert.Equal(t, state.Packet.NearRelayIDs, updatePacket.NearRelayIDs)
		for i := 0; i < len(state.Packet.NearRelayIDs); i++ {
			assert.Equal(t, state.Packet.NearRelayRTT[i], state.Input.HoldNearRelayRTT[i])
			assert.Equal(t, state.Packet.NearRelayRTT[i], updatePacket.NearRelayRTT[i])
		}
	})
}

func TestSessionUpdateHandlerSDK5Func_SessionUpdateNearRelayStats_RelayNoLongerExists(t *testing.T) {
	t.Parallel()

	updatePacket := transport.SessionUpdatePacketSDK5{
		SliceNumber:  uint32(2),
		NearRelayIDs: []uint64{1234, 12345, 123456},
	}

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	buyer := routing.Buyer{
		InternalConfig: core.InternalConfig{
			LargeCustomer: false,
		},
	}

	dc := routing.Datacenter{
		ID:        crypto.HashID("datacenter.name"),
		Name:      "datacenter.name",
		AliasName: "datacenter.name",
	}

	relayIDsToIndices := make(map[uint64]int32)
	relayIDsToIndices[1234] = 0
	relayIDsToIndices[12345] = 1

	routeMatrix := &routing.RouteMatrix{
		RelayIDsToIndices: relayIDsToIndices,
		RelayIDs: []uint64{
			crypto.HashID("datacenter.name"),
			uint64(1234),
			uint64(12345),
			uint64(123456),
		},
		RelayDatacenterIDs: []uint64{
			crypto.HashID("datacenter.name"),
			uint64(1234),
			uint64(12345),
			uint64(123456),
		},
	}

	state := transport.SessionHandlerStateSDK5{
		Packet:                        updatePacket,
		Metrics:                       metrics.SessionUpdateMetrics,
		RouteMatrix:                   routeMatrix,
		Datacenter:                    dc,
		Buyer:                         buyer,
		DatacenterAccelerationEnabled: true,
	}

	assert.False(t, state.Buyer.InternalConfig.LargeCustomer)
	assert.True(t, transport.SessionUpdateNearRelayStatsSDK5(&state))
	assert.False(t, state.Output.HoldNearRelays)
	assert.False(t, state.Response.DontPingNearRelays)
	assert.Equal(t, int32(0), state.NearRelayIndices[0])
	assert.Equal(t, int32(1), state.NearRelayIndices[1])
	assert.Equal(t, int32(-1), state.NearRelayIndices[2])
}

func TestSessionUpdateHandlerSDK5Func_SessionMakeRouteDecision_NextWithoutRouteRelays(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerStateSDK5{
		Metrics: metrics.SessionUpdateMetrics,
		Input: transport.SessionDataSDK5{
			RouteState: core.RouteState{
				Next: true,
			},
			RouteNumRelays: 0,
		},
		DatacenterAccelerationEnabled: true,
	}

	transport.SessionMakeRouteDecisionSDK5(&state)

	assert.False(t, state.Output.RouteState.Next)
	assert.True(t, state.Output.RouteState.Veto)
	assert.Equal(t, float64(1), state.Metrics.NextWithoutRouteRelays.Value())
}

func TestSessionUpdateHandlerSDK5Func_SessionMakeRouteDecision_SDKAbortedSession(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerStateSDK5{
		Metrics: metrics.SessionUpdateMetrics,
		Input: transport.SessionDataSDK5{
			RouteState: core.RouteState{
				Next: true,
			},
			RouteNumRelays: 5,
		},
		Packet: transport.SessionUpdatePacketSDK5{
			Next: false,
		},
	}

	transport.SessionMakeRouteDecisionSDK5(&state)

	assert.False(t, state.Output.RouteState.Next)
	assert.True(t, state.Output.RouteState.Veto)
	assert.Equal(t, float64(1), state.Metrics.SDKAborted.Value())
}

func TestSessionUpdateHandlerSDK5Func_BuyerNotFound_NoResponse(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false, false)
	datacenter := env.AddDatacenter("datacenter.name")

	unknownPublicKey, unknownPrivateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	unknownBuyerID := binary.LittleEndian.Uint64(unknownPublicKey[:8])

	unknownPublicKey = unknownPublicKey[8:]
	unknownPrivateKey = unknownPrivateKey[8:]

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	getIPLocatorFunc := func() *routing.MaxmindDB {
		return getSuccessLocator(t)
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], unknownPrivateKey)

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", env.GetDatabaseWrapper)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	sessionUpdateConfig := test.SessionUpdatePacketConfigSDK5{
		Version:        transport.SDKVersion{5, 0, 0},
		BuyerID:        unknownBuyerID,
		DatacenterID:   datacenter.ID,
		SessionID:      123456789,
		SliceNumber:    0,
		Next:           false,
		PublicKey:      unknownPublicKey,
		PrivateKey:     unknownPrivateKey,
		ServerAddress:  "127.0.0.1:32202",
		BackendAddress: "127.0.0.1:40000",
	}

	requestData := env.GenerateSessionUpdatePacketSDK5(sessionUpdateConfig)

	handler := transport.SessionUpdateHandlerSDK5Func(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, env.GetDatabaseWrapper, env.GetMagicValues, core.ParseAddress(sessionUpdateConfig.BackendAddress), crypto.BackendPrivateKey, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	// Buyer not found - no response
	assert.Zero(t, len(responseBuffer.Bytes()))
}

func TestSessionUpdateHandlerSDK5Func_SigCheckFailed_NoResponse(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, publicKey, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	getIPLocatorFunc := func() *routing.MaxmindDB {
		return getSuccessLocator(t)
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], privateKey)

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", env.GetDatabaseWrapper)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	sessionUpdateConfig := test.SessionUpdatePacketConfigSDK5{
		Version:        transport.SDKVersion{5, 0, 0},
		BuyerID:        buyerID,
		DatacenterID:   datacenter.ID,
		SessionID:      123456789,
		SliceNumber:    0,
		Next:           false,
		PublicKey:      publicKey,
		PrivateKey:     privateKey[2:],
		ServerAddress:  "127.0.0.1:32202",
		BackendAddress: "127.0.0.1:40000",
	}

	requestData := env.GenerateSessionUpdatePacketSDK5(sessionUpdateConfig)

	handler := transport.SessionUpdateHandlerSDK5Func(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, env.GetDatabaseWrapper, env.GetMagicValues, core.ParseAddress(sessionUpdateConfig.BackendAddress), crypto.BackendPrivateKey, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		From: *core.ParseAddress(sessionUpdateConfig.ServerAddress),
		Data: requestData,
	})

	// SigCheck failed - no response
	assert.Zero(t, len(responseBuffer.Bytes()))
}

func TestSessionUpdateHandlerSDK5Func_DirectResponse(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, publicKey, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("losangeles")
	env.AddRelay("los.angeles.1", "10.0.0.2", datacenter.ID)
	env.AddRelay("los.angeles.2", "10.0.0.3", datacenter.ID)

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	getIPLocatorFunc := func() *routing.MaxmindDB {
		return getSuccessLocator(t)
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], privateKey)

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", env.GetDatabaseWrapper)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	sessionUpdateConfig := test.SessionUpdatePacketConfigSDK5{
		Version:        transport.SDKVersion{5, 0, 0},
		BuyerID:        buyerID,
		DatacenterID:   datacenter.ID,
		SessionID:      123456789,
		SliceNumber:    0,
		Next:           false,
		PublicKey:      publicKey,
		PrivateKey:     privateKey,
		ServerAddress:  "127.0.0.1:32202",
		BackendAddress: "127.0.0.1:40000",
	}

	requestData := env.GenerateSessionUpdatePacketSDK5(sessionUpdateConfig)

	handler := transport.SessionUpdateHandlerSDK5Func(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, env.GetDatabaseWrapper, env.GetMagicValues, core.ParseAddress(sessionUpdateConfig.BackendAddress), crypto.BackendPrivateKey, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		From: *core.ParseAddress(sessionUpdateConfig.ServerAddress),
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacketSDK5
	responsePacket.Version = transport.SDKVersion{5, 0, 0} // Make sure we unmarshal the response the same way we marshaled the request
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, int32(routing.RouteTypeDirect), responsePacket.RouteType)

	var sessionData transport.SessionDataSDK5
	err = transport.UnmarshalSessionDataSDK5(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.False(t, sessionData.RouteState.Next)
}

func TestSessionUpdateHandlerSDK5Func_SessionMakeRouteDecision_NextResponse(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, publicKey, privateKey := env.AddBuyer("local", true, false)
	datacenterLA := env.AddDatacenter("losangeles")
	env.AddRelay("losangeles.1", "10.0.0.2", datacenterLA.ID)
	datacenterChicago := env.AddDatacenter("chicago")
	env.AddRelay("chicago.1", "10.0.0.4", datacenterChicago.ID)
	env.SetCost("losangeles.1", "chicago.1", 10)

	env.DatabaseWrapper.DatacenterMaps[buyerID][datacenterLA.ID] = routing.DatacenterMap{
		BuyerID:            buyerID,
		DatacenterID:       datacenterLA.ID,
		EnableAcceleration: true,
	}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	getIPLocatorFunc := func() *routing.MaxmindDB {
		return getSuccessLocator(t)
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], privateKey)

	relayIDs := env.GetRelayIds()

	for i, id := range relayIDs {
		env.DatabaseWrapper.RelayMap[id] = routing.Relay{
			Addr:      env.GetRelayAddresses()[i],
			PublicKey: publicKey,
		}
	}

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", env.GetDatabaseWrapper)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	sessionDataConfig := test.SessionDataConfigSDK5{
		Version:     transport.SessionDataVersionSDK5,
		Initial:     false,
		SessionID:   123456789,
		SliceNumber: 3,
	}

	sessionDataPacket, sessionDataSize := env.GenerateSessionDataPacketSDK5(sessionDataConfig)

	sessionUpdateConfig := test.SessionUpdatePacketConfigSDK5{
		Version:          transport.SDKVersion{5, 0, 0},
		BuyerID:          buyerID,
		DatacenterID:     datacenterLA.ID,
		SessionID:        sessionDataConfig.SessionID,
		SliceNumber:      sessionDataConfig.SliceNumber,
		PublicKey:        publicKey,
		PrivateKey:       privateKey,
		SessionData:      sessionDataPacket,
		SessionDataBytes: int32(sessionDataSize),
		NearRelayRTT:     10,
		NearRelayJitter:  10,
		NearRelayPL:      1,
		NextRTT:          10,
		NextJitter:       10,
		NextPacketLoss:   1,
		DirectMinRTT:     1000,
		DirectMaxRTT:     1000,
		DirectPrimeRTT:   1000,
		DirectJitter:     1000,
		DirectPacketLoss: 100,
		ClientAddress:    "10.0.0.9",
		ServerAddress:    "10.0.0.10",
		BackendAddress:   "127.0.0.1:40000",
		UserHash:         100,
	}

	requestData := env.GenerateSessionUpdatePacketSDK5(sessionUpdateConfig)

	handler := transport.SessionUpdateHandlerSDK5Func(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, env.GetDatabaseWrapper, env.GetMagicValues, core.ParseAddress(sessionUpdateConfig.BackendAddress), crypto.BackendPrivateKey, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		From: *core.ParseAddress(sessionUpdateConfig.ServerAddress),
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacketSDK5
	responsePacket.Version = transport.SDKVersion{5, 0, 0} // Make sure we unmarshal the response the same way we marshaled the request
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, int32(routing.RouteTypeNew), responsePacket.RouteType)

	var sessionData transport.SessionDataSDK5
	err = transport.UnmarshalSessionDataSDK5(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.True(t, sessionData.RouteState.Next)
	assert.True(t, sessionData.Initial)
}

func TestSessionUpdateHandlerSDK5Func_SessionMakeRouteDecision_ContinueResponse(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, publicKey, privateKey := env.AddBuyer("local", true, false)
	datacenterLA := env.AddDatacenter("losangeles")
	env.AddRelay("losangeles.1", "10.0.0.2", datacenterLA.ID)
	datacenterChicago := env.AddDatacenter("chicago")
	env.AddRelay("chicago.1", "10.0.0.4", datacenterChicago.ID)
	env.SetCost("losangeles.1", "chicago.1", 10)

	env.DatabaseWrapper.DatacenterMaps[buyerID][datacenterLA.ID] = routing.DatacenterMap{
		BuyerID:            buyerID,
		DatacenterID:       datacenterLA.ID,
		EnableAcceleration: true,
	}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	getIPLocatorFunc := func() *routing.MaxmindDB {
		return getSuccessLocator(t)
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], privateKey)

	relayIDs := env.GetRelayIds()

	for i, id := range relayIDs {
		env.DatabaseWrapper.RelayMap[id] = routing.Relay{
			Addr:      env.GetRelayAddresses()[i],
			PublicKey: publicKey,
		}
	}

	env.DatabaseWrapper.DatacenterMaps[buyerID][datacenterLA.ID] = routing.DatacenterMap{
		BuyerID:            buyerID,
		DatacenterID:       datacenterLA.ID,
		EnableAcceleration: true,
	}

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", env.GetDatabaseWrapper)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	sessionDataConfig := test.SessionDataConfigSDK5{
		Version:        transport.SessionDataVersionSDK5,
		Initial:        true,
		SessionID:      123456789,
		SliceNumber:    5,
		RouteNumRelays: 2,
		RouteRelayIDs:  [5]uint64{relayIDs[0], relayIDs[1], 0, 0, 0},
		RouteState: core.RouteState{
			UserID:          100,
			Next:            true,
			NumNearRelays:   2,
			NearRelayRTT:    [32]int32{10, 10},
			NearRelayJitter: [32]int32{10, 10},
		},
	}

	sessionDataPacket, sessionDataSize := env.GenerateSessionDataPacketSDK5(sessionDataConfig)

	sessionUpdateConfig := test.SessionUpdatePacketConfigSDK5{
		Version:          transport.SDKVersion{5, 0, 0},
		BuyerID:          buyerID,
		DatacenterID:     datacenterLA.ID,
		SessionID:        sessionDataConfig.SessionID,
		SliceNumber:      sessionDataConfig.SliceNumber,
		PublicKey:        publicKey,
		PrivateKey:       privateKey,
		SessionData:      sessionDataPacket,
		SessionDataBytes: int32(sessionDataSize),
		Next:             true,
		Committed:        true,
		NearRelayRTT:     10,
		NearRelayJitter:  10,
		NearRelayPL:      1,
		NextRTT:          11,
		NextJitter:       11,
		NextPacketLoss:   1,
		DirectMinRTT:     1000,
		DirectMaxRTT:     1000,
		DirectPrimeRTT:   1000,
		DirectJitter:     1000,
		DirectPacketLoss: 100,
		ClientAddress:    "10.0.0.9",
		ServerAddress:    "10.0.0.10",
		BackendAddress:   "127.0.0.1:40000",
		UserHash:         100,
	}

	requestData := env.GenerateSessionUpdatePacketSDK5(sessionUpdateConfig)

	handler := transport.SessionUpdateHandlerSDK5Func(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, env.GetDatabaseWrapper, env.GetMagicValues, core.ParseAddress(sessionUpdateConfig.BackendAddress), crypto.BackendPrivateKey, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		From: *core.ParseAddress(sessionUpdateConfig.ServerAddress),
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacketSDK5
	responsePacket.Version = transport.SDKVersion{5, 0, 0} // Make sure we unmarshal the response the same way we marshaled the request
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, int32(routing.RouteTypeContinue), responsePacket.RouteType)

	var sessionData transport.SessionDataSDK5
	err = transport.UnmarshalSessionDataSDK5(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.True(t, sessionData.RouteState.Next)
	assert.False(t, sessionData.Initial)
}

// Match data handler tests

func TestMatchDataHandlerSDK5Func_BuyerNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false, false)

	unknownPublicKey, unknownPrivateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	unknownPrivateKey = unknownPrivateKey[8:]

	unknownBuyerID := binary.LittleEndian.Uint64(unknownPublicKey[:8])

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
	assert.NoError(t, err)

	backendAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	matchDataConfig := test.MatchDataPacketConfigSDK5{
		Version:        transport.SDKVersion{5, 0, 0},
		BuyerID:        unknownBuyerID,
		ServerAddress:  *serverAddr,
		BackendAddress: *backendAddr,
		DatacenterID:   rand.Uint64(),
		UserHash:       rand.Uint64(),
		SessionID:      crypto.GenerateSessionID(),
		MatchID:        rand.Uint64(),
		PrivateKey:     unknownPrivateKey,
	}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateMatchDataRequestPacketSDK5(matchDataConfig)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.MatchDataHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, postSessionHandler, backendAddr, crypto.BackendPrivateKey, metrics.MatchDataHandlerMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *serverAddr,
		Data: requestData,
	})

	var responsePacket transport.MatchDataResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.MatchDataResponseUnknownBuyer), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.MatchDataHandlerMetrics.BuyerNotFound.Value())
}

func TestMatchDataHandlerSDK5Func_BuyerNotLive(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", false, false)

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
	assert.NoError(t, err)

	backendAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	matchDataConfig := test.MatchDataPacketConfigSDK5{
		Version:        transport.SDKVersion{5, 0, 0},
		BuyerID:        buyerID,
		ServerAddress:  *serverAddr,
		BackendAddress: *backendAddr,
		DatacenterID:   rand.Uint64(),
		UserHash:       rand.Uint64(),
		SessionID:      crypto.GenerateSessionID(),
		MatchID:        rand.Uint64(),
		PrivateKey:     privateKey,
	}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateMatchDataRequestPacketSDK5(matchDataConfig)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.MatchDataHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, postSessionHandler, backendAddr, crypto.BackendPrivateKey, metrics.MatchDataHandlerMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *serverAddr,
		Data: requestData,
	})

	var responsePacket transport.MatchDataResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.MatchDataResponseBuyerNotActive), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.MatchDataHandlerMetrics.BuyerNotActive.Value())
}

func TestMatchDataHandlerSDK5Func_SigCheckFail(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
	assert.NoError(t, err)

	backendAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	matchDataConfig := test.MatchDataPacketConfigSDK5{
		Version:        transport.SDKVersion{5, 0, 0},
		BuyerID:        buyerID,
		ServerAddress:  *serverAddr,
		BackendAddress: *backendAddr,
		DatacenterID:   rand.Uint64(),
		UserHash:       rand.Uint64(),
		SessionID:      crypto.GenerateSessionID(),
		MatchID:        rand.Uint64(),
		PrivateKey:     privateKey[2:],
	}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateMatchDataRequestPacketSDK5(matchDataConfig)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.MatchDataHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, postSessionHandler, backendAddr, crypto.BackendPrivateKey, metrics.MatchDataHandlerMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *serverAddr,
		Data: requestData,
	})

	var responsePacket transport.MatchDataResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.MatchDataResponseSignatureCheckFailed), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.MatchDataHandlerMetrics.SignatureCheckFailed.Value())
}

func TestMatchDataHandlerSDK5Func_Success(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
	assert.NoError(t, err)

	backendAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	matchDataConfig := test.MatchDataPacketConfigSDK5{
		Version:        transport.SDKVersion{5, 0, 0},
		BuyerID:        buyerID,
		ServerAddress:  *serverAddr,
		BackendAddress: *backendAddr,
		DatacenterID:   rand.Uint64(),
		UserHash:       rand.Uint64(),
		SessionID:      crypto.GenerateSessionID(),
		MatchID:        rand.Uint64(),
		PrivateKey:     privateKey,
	}

	metrics, err := metrics.NewServerBackend5Metrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateMatchDataRequestPacketSDK5(matchDataConfig)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.MatchDataHandlerSDK5Func(env.GetDatabaseWrapper, env.GetMagicValues, postSessionHandler, backendAddr, crypto.BackendPrivateKey, metrics.MatchDataHandlerMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		From: *serverAddr,
		Data: requestData,
	})

	var responsePacket transport.MatchDataResponsePacketSDK5
	err = transport.UnmarshalPacketSDK5(&responsePacket, core.GetPacketDataSDK5(responseBuffer.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.MatchDataResponseOK), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.MatchDataHandlerMetrics.HandlerMetrics.Invocations.Value())
	assert.Zero(t, metrics.MatchDataHandlerMetrics.BuyerNotFound.Value())
	assert.Zero(t, metrics.MatchDataHandlerMetrics.BuyerNotActive.Value())
	assert.Zero(t, metrics.MatchDataHandlerMetrics.SignatureCheckFailed.Value())
}
