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

func TestServerUpdateHandlerReadPacketFailure(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, 0, false, nil, logger, metrics.PostSessionMetrics)
	handler := transport.ServerUpdateHandlerFunc(logger, storer, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: nil,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.ServerUpdateMetrics.ReadPacketFailure.Value(), 1.0)
}

func TestServerUpdateHandlerBuyerNotFound(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerUpdatePacket{}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, 0, false, nil, logger, metrics.PostSessionMetrics)
	handler := transport.ServerUpdateHandlerFunc(logger, storer, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, metrics.ServerUpdateMetrics.BuyerNotFound.Value(), 1.0)
}

func TestServerUpdateHandlerSDKTooOld(t *testing.T) {
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

	requestPacket := transport.ServerUpdatePacket{
		Version:    transport.SDKVersion{3, 3, 4},
		CustomerID: 123,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, 0, false, nil, logger, metrics.PostSessionMetrics)
	handler := transport.ServerUpdateHandlerFunc(logger, storer, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, metrics.ServerUpdateMetrics.SDKTooOld.Value(), 1.0)
}

func TestServerUpdateHandlerMisconfiguredDatacenterAlias(t *testing.T) {
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

	requestPacket := transport.ServerUpdatePacket{
		Version:      transport.SDKVersion{4, 0, 0},
		CustomerID:   123,
		DatacenterID: crypto.HashID("datacenter.alias"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, 0, false, nil, logger, metrics.PostSessionMetrics)
	handler := transport.ServerUpdateHandlerFunc(logger, storer, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, metrics.ServerUpdateMetrics.MisconfiguredDatacenterAlias.Value(), 1.0)
}

func TestServerUpdateHandlerDatacenterNotFound(t *testing.T) {
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

	requestPacket := transport.ServerUpdatePacket{
		Version:      transport.SDKVersion{4, 0, 0},
		CustomerID:   123,
		DatacenterID: crypto.HashID("datacenter.alias"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, 0, false, nil, logger, metrics.PostSessionMetrics)
	handler := transport.ServerUpdateHandlerFunc(logger, storer, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, metrics.ServerUpdateMetrics.DatacenterNotFound.Value(), 1.0)
}

func TestServerUpdateHandlerDatacenterNotAllowed(t *testing.T) {
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

	requestPacket := transport.ServerUpdatePacket{
		Version:      transport.SDKVersion{4, 0, 0},
		CustomerID:   123,
		DatacenterID: crypto.HashID("datacenter.name"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, 0, false, nil, logger, metrics.PostSessionMetrics)
	handler := transport.ServerUpdateHandlerFunc(logger, storer, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, metrics.ServerUpdateMetrics.DatacenterNotAllowed.Value(), 1.0)
}

func TestServerUpdateHandlerSuccess(t *testing.T) {
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
	expectedMetrics := metrics.EmptyServerUpdateMetrics
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerUpdatePacket{
		Version:      transport.SDKVersion{4, 0, 0},
		CustomerID:   123,
		DatacenterID: crypto.HashID("datacenter.name"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, 0, false, nil, logger, metrics.PostSessionMetrics)
	handler := transport.ServerUpdateHandlerFunc(logger, storer, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assertAllMetricsEqual(t, expectedMetrics, *metrics.ServerUpdateMetrics)
}

func TestServerUpdateHandlerSuccessDatacenterAliasFound(t *testing.T) {
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
	expectedMetrics := metrics.EmptyServerUpdateMetrics
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerUpdatePacket{
		Version:      transport.SDKVersion{4, 0, 0},
		CustomerID:   123,
		DatacenterID: crypto.HashID("datacenter.alias"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, 0, false, nil, logger, metrics.PostSessionMetrics)
	handler := transport.ServerUpdateHandlerFunc(logger, storer, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assertAllMetricsEqual(t, expectedMetrics, *metrics.ServerUpdateMetrics)
}
