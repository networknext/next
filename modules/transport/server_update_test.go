package transport_test

/*
import (
	"bytes"
	"context"
	"encoding/binary"
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

func TestServerUpdateHandlerSignatureCheckFailed(t *testing.T) {
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
		CustomerID: 123,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, 0, false, nil, logger, metrics.PostSessionMetrics)
	handler := transport.ServerUpdateHandlerFunc(logger, storer, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, metrics.ServerUpdateMetrics.SignatureCheckFailed.Value(), 1.0)
}

func TestServerUpdateHandlerSDKTooOld(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	customerID := binary.LittleEndian.Uint64(privateKey[:8])
	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:        customerID,
		PublicKey: publicKey,
	})
	assert.NoError(t, err)

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerUpdatePacket{
		Version:    transport.SDKVersion{3, 3, 4},
		CustomerID: customerID,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

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

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	customerID := binary.LittleEndian.Uint64(privateKey[:8])
	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:        customerID,
		PublicKey: publicKey,
	})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{
		BuyerID:      customerID,
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
		CustomerID:   customerID,
		DatacenterID: crypto.HashID("datacenter.alias"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

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

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	customerID := binary.LittleEndian.Uint64(privateKey[:8])
	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:        customerID,
		PublicKey: publicKey,
	})
	assert.NoError(t, err)

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerUpdatePacket{
		Version:      transport.SDKVersion{4, 0, 0},
		CustomerID:   customerID,
		DatacenterID: crypto.HashID("datacenter.alias"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

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

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	customerID := binary.LittleEndian.Uint64(privateKey[:8])
	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:        customerID,
		PublicKey: publicKey,
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
		CustomerID:   customerID,
		DatacenterID: crypto.HashID("datacenter.name"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

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

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	customerID := binary.LittleEndian.Uint64(privateKey[:8])
	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:        customerID,
		PublicKey: publicKey,
	})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{
		ID:   crypto.HashID("datacenter.name"),
		Name: "datacenter.name",
	})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{
		BuyerID:      customerID,
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
		CustomerID:   customerID,
		DatacenterID: crypto.HashID("datacenter.name"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

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

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	customerID := binary.LittleEndian.Uint64(privateKey[:8])
	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:        customerID,
		PublicKey: publicKey,
	})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{
		ID:   crypto.HashID("datacenter.name"),
		Name: "datacenter.name",
	})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{
		BuyerID:      customerID,
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
		CustomerID:   customerID,
		DatacenterID: crypto.HashID("datacenter.alias"),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, 0, false, nil, logger, metrics.PostSessionMetrics)
	handler := transport.ServerUpdateHandlerFunc(logger, storer, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assertAllMetricsEqual(t, expectedMetrics, *metrics.ServerUpdateMetrics)
}
*/
