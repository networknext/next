package transport_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func TestServerInitHandlerReadPacketFailure(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	handler := transport.ServerInitHandlerFunc(logger, storer, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: nil,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.ServerInitMetrics.ReadPacketFailure.Value(), 1.0)
}

func TestServerInitHandlerBuyerNotFound(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc(logger, storer, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseUnknownCustomer), responsePacket.Response)

	assert.Equal(t, metrics.ServerInitMetrics.BuyerNotFound.Value(), 1.0)
}

func TestServerInitHandlerSDKTooOld(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID: 123,
	})
	assert.NoError(t, err)

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:    transport.SDKVersion{3, 3, 4},
		CustomerID: 123,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc(logger, storer, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseOldSDKVersion), responsePacket.Response)

	assert.Equal(t, metrics.ServerInitMetrics.SDKTooOld.Value(), 1.0)
}

func TestServerInitHandlerMisconfiguredDatacenterAlias(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID: 123,
	})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{
		BuyerID:      123,
		DatacenterID: crypto.HashID("datacenter.name"),
		Alias:        "datacenter.alias",
	})
	assert.NoError(t, err)

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     123,
		DatacenterID:   crypto.HashID("datacenter.alias"),
		DatacenterName: "datacenter.alias",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc(logger, storer, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseUnknownDatacenter), responsePacket.Response)

	assert.Equal(t, metrics.ServerInitMetrics.MisconfiguredDatacenterAlias.Value(), 1.0)
}

func TestServerInitHandlerDatacenterNotFound(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID: 123,
	})
	assert.NoError(t, err)

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     123,
		DatacenterID:   crypto.HashID("datacenter.alias"),
		DatacenterName: "datacenter.alias",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc(logger, storer, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseUnknownDatacenter), responsePacket.Response)

	assert.Equal(t, metrics.ServerInitMetrics.DatacenterNotFound.Value(), 1.0)
}

func TestServerInitHandlerDatacenterNotAllowed(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID: 123,
	})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{
		ID: crypto.HashID("datacenter.name"),
	})
	assert.NoError(t, err)

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     123,
		DatacenterID:   crypto.HashID("datacenter.name"),
		DatacenterName: "datacenter.name",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc(logger, storer, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseUnknownDatacenter), responsePacket.Response)

	assert.Equal(t, metrics.ServerInitMetrics.DatacenterNotAllowed.Value(), 1.0)
}

func TestServerInitHandlerSuccess(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID: 123,
	})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{
		ID:   crypto.HashID("datacenter.name"),
		Name: "datacenter.name",
	})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{
		BuyerID:      123,
		DatacenterID: crypto.HashID("datacenter.name"),
	})
	assert.NoError(t, err)

	metricsHandler := metrics.LocalHandler{}
	expectedMetrics := metrics.EmptyServerInitMetrics
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     123,
		DatacenterID:   crypto.HashID("datacenter.name"),
		DatacenterName: "datacenter.name",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc(logger, storer, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)

	assertAllMetricsEqual(t, expectedMetrics, *metrics.ServerInitMetrics)
}

func TestServerInitHandlerSuccessDatacenterAliasFound(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID: 123,
	})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{
		ID:   crypto.HashID("datacenter.name"),
		Name: "datacenter.name",
	})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{
		BuyerID:      123,
		DatacenterID: crypto.HashID("datacenter.name"),
		Alias:        "datacenter.alias",
	})
	assert.NoError(t, err)

	metricsHandler := metrics.LocalHandler{}
	expectedMetrics := metrics.EmptyServerInitMetrics
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     123,
		DatacenterID:   crypto.HashID("datacenter.alias"),
		DatacenterName: "datacenter.alias",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc(logger, storer, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)

	assertAllMetricsEqual(t, expectedMetrics, *metrics.ServerInitMetrics)
}
