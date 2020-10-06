package transport_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestServerUpdateHandler4ReadPacketFailure(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}
	datacenterTracker := transport.NewDatacenterTracker()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	handler := transport.ServerUpdateHandlerFunc4(logger, storer, datacenterTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: nil,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.ServerUpdateMetrics.ReadPacketFailure.Value(), 1.0)
}

func TestServerUpdateHandler4BuyerNotFound(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}
	datacenterTracker := transport.NewDatacenterTracker()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerUpdatePacket4{}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerUpdateHandlerFunc4(logger, storer, datacenterTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, metrics.ServerUpdateMetrics.BuyerNotFound.Value(), 1.0)
}

func TestServerUpdateHandler4SDKTooOld(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID: 123,
	})
	assert.NoError(t, err)

	datacenterTracker := transport.NewDatacenterTracker()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerUpdatePacket4{
		Version:    transport.SDKVersion{3, 3, 4},
		CustomerID: 123,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerUpdateHandlerFunc4(logger, storer, datacenterTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, metrics.ServerUpdateMetrics.SDKTooOld.Value(), 1.0)
}

func TestServerUpdateHandler4MisconfiguredDatacenterAlias(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID: 123,
	})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{
		BuyerID:    123,
		Datacenter: crypto.HashID("datacenter.name"),
		Alias:      "datacenter.alias",
	})
	assert.NoError(t, err)

	datacenterTracker := transport.NewDatacenterTracker()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerUpdatePacket4{
		Version:      transport.SDKVersion{4, 0, 0},
		CustomerID:   123,
		DatacenterID: crypto.HashID("datacenter.alias"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerUpdateHandlerFunc4(logger, storer, datacenterTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, metrics.ServerUpdateMetrics.DatacenterNotFound.Value(), 1.0)

	unknownDatacenters := datacenterTracker.GetUnknownDatacenters()
	assert.Equal(t, []string{fmt.Sprintf("%016x", crypto.HashID("datacenter.alias"))}, unknownDatacenters)
}

func TestServerUpdateHandler4DatacenterAndAliasNotFound(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID: 123,
	})
	assert.NoError(t, err)

	datacenterTracker := transport.NewDatacenterTracker()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerUpdatePacket4{
		Version:      transport.SDKVersion{4, 0, 0},
		CustomerID:   123,
		DatacenterID: crypto.HashID("datacenter.alias"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerUpdateHandlerFunc4(logger, storer, datacenterTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, metrics.ServerUpdateMetrics.DatacenterNotFound.Value(), 1.0)

	unknownDatacenters := datacenterTracker.GetUnknownDatacenters()
	assert.Equal(t, []string{fmt.Sprintf("%016x", crypto.HashID("datacenter.alias"))}, unknownDatacenters)
}

func TestServerUpdateHandler4Success(t *testing.T) {
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

	datacenterTracker := transport.NewDatacenterTracker()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerUpdatePacket4{
		Version:      transport.SDKVersion{4, 0, 0},
		CustomerID:   123,
		DatacenterID: crypto.HashID("datacenter.name"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerUpdateHandlerFunc4(logger, storer, datacenterTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	unknownDatacenters := datacenterTracker.GetUnknownDatacenters()
	assert.Empty(t, unknownDatacenters)
}

func TestServerUpdateHandler4SuccessDatacenterAliasFound(t *testing.T) {
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
		BuyerID:    123,
		Datacenter: crypto.HashID("datacenter.name"),
		Alias:      "datacenter.alias",
	})
	assert.NoError(t, err)

	datacenterTracker := transport.NewDatacenterTracker()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerUpdatePacket4{
		Version:      transport.SDKVersion{4, 0, 0},
		CustomerID:   123,
		DatacenterID: crypto.HashID("datacenter.alias"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerUpdateHandlerFunc4(logger, storer, datacenterTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	unknownDatacenterNames := datacenterTracker.GetUnknownDatacentersNames()
	assert.Empty(t, unknownDatacenterNames)
}
