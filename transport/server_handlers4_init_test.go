package transport_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestServerInitHandler4ReadPacketFailure(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}
	datacenterTracker := transport.NewDatacenterTracker()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerInitMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	handler := transport.ServerInitHandlerFunc4(logger, storer, datacenterTracker, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: nil,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.ErrorMetrics.ReadPacketFailure.Value(), 1.0)
}

func TestServerInitHandler4BuyerNotFound(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}
	datacenterTracker := transport.NewDatacenterTracker()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerInitMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket4{}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc4(logger, storer, datacenterTracker, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseUnknownCustomer), responsePacket.Response)

	assert.Equal(t, metrics.ErrorMetrics.BuyerNotFound.Value(), 1.0)
}

func TestServerInitHandler4SDKTooOld(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID: 123,
	})
	assert.NoError(t, err)

	datacenterTracker := transport.NewDatacenterTracker()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerInitMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket4{
		Version:    transport.SDKVersion{3, 3, 4},
		CustomerID: 123,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc4(logger, storer, datacenterTracker, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseOldSDKVersion), responsePacket.Response)

	assert.Equal(t, metrics.ErrorMetrics.SDKTooOld.Value(), 1.0)
}

func TestServerInitHandler4MisconfiguredDatacenterAlias(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		CompanyCode: "local",
		ID:          123,
	})
	assert.NoError(t, err)

	err = storer.AddCustomer(context.Background(), routing.Customer{
		Code: "local",
		Name: "Local",
	})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{
		ID: crypto.HashID("datacenter.name"),
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
	metrics, err := metrics.NewServerInitMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket4{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     123,
		DatacenterID:   crypto.HashID("datacenter.alias"),
		DatacenterName: "datacenter.alias",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc4(logger, storer, datacenterTracker, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseUnknownDatacenter), responsePacket.Response)

	assert.Equal(t, metrics.ErrorMetrics.DatacenterNotFound.Value(), 1.0)

	unknownDatacenterNames := datacenterTracker.GetUnknownDatacentersNames()
	assert.Equal(t, []string{"datacenter.alias"}, unknownDatacenterNames)
}

func TestServerInitHandler4DatacenterAndAliasNotFound(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID: 123,
	})
	assert.NoError(t, err)

	datacenterTracker := transport.NewDatacenterTracker()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerInitMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket4{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     123,
		DatacenterID:   crypto.HashID("datacenter.alias"),
		DatacenterName: "datacenter.alias",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc4(logger, storer, datacenterTracker, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)

	assert.Equal(t, metrics.ErrorMetrics.DatacenterNotFound.Value(), 1.0)

	unknownDatacenterNames := datacenterTracker.GetUnknownDatacentersNames()
	assert.Equal(t, []string{"datacenter.alias"}, unknownDatacenterNames)
}

func TestServerInitHandler4Success(t *testing.T) {
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
	metrics, err := metrics.NewServerInitMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket4{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     123,
		DatacenterID:   crypto.HashID("datacenter.name"),
		DatacenterName: "datacenter.name",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc4(logger, storer, datacenterTracker, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)

	unknownDatacenterNames := datacenterTracker.GetUnknownDatacentersNames()
	assert.Empty(t, unknownDatacenterNames)
}

func TestServerInitHandler4SuccessDatacenterAliasFound(t *testing.T) {
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
	metrics, err := metrics.NewServerInitMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket4{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     123,
		DatacenterID:   crypto.HashID("datacenter.alias"),
		DatacenterName: "datacenter.alias",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.ServerInitHandlerFunc4(logger, storer, datacenterTracker, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)

	unknownDatacenterNames := datacenterTracker.GetUnknownDatacentersNames()
	assert.Empty(t, unknownDatacenterNames)
}
