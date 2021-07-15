package transport_test

/* import (
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

func TestServerInitHandlerBuyerNotLive(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID:   123,
		Live: false,
	})
	assert.NoError(t, err)

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
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
	assert.Equal(t, uint32(transport.InitResponseCustomerNotActive), responsePacket.Response)

	assert.Equal(t, metrics.ServerInitMetrics.SignatureCheckFailed.Value(), 1.0)
}

func TestServerInitHandlerSignatureCheckFailed(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{
		ID:   123,
		Live: true,
	})
	assert.NoError(t, err)

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
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
	assert.Equal(t, uint32(transport.InitResponseSignatureCheckFailed), responsePacket.Response)

	assert.Equal(t, metrics.ServerInitMetrics.SignatureCheckFailed.Value(), 1.0)
}

func TestServerInitHandlerSDKTooOld(t *testing.T) {
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
		Live:      true,
	})
	assert.NoError(t, err)

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
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

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	customerID := binary.LittleEndian.Uint64(privateKey[:8])
	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:        customerID,
		PublicKey: publicKey,
		Live:      true,
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

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     customerID,
		DatacenterID:   crypto.HashID("datacenter.alias"),
		DatacenterName: "datacenter.alias",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

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

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	customerID := binary.LittleEndian.Uint64(privateKey[:8])
	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:        customerID,
		PublicKey: publicKey,
		Live:      true,
	})
	assert.NoError(t, err)

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     customerID,
		DatacenterID:   crypto.HashID("datacenter.alias"),
		DatacenterName: "datacenter.alias",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

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

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	customerID := binary.LittleEndian.Uint64(privateKey[:8])
	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:        customerID,
		PublicKey: publicKey,
		Live:      true,
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
		CustomerID:     customerID,
		DatacenterID:   crypto.HashID("datacenter.name"),
		DatacenterName: "datacenter.name",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

	handler := transport.ServerInitHandlerFunc(logger, storer, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseDataCenterNotEnabled), responsePacket.Response)

	assert.Equal(t, metrics.ServerInitMetrics.DatacenterNotAllowed.Value(), 1.0)
}

func TestServerInitHandlerSuccess(t *testing.T) {
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
		Live:      true,
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
	expectedMetrics := metrics.EmptyServerInitMetrics
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     customerID,
		DatacenterID:   crypto.HashID("datacenter.name"),
		DatacenterName: "datacenter.name",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

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

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	customerID := binary.LittleEndian.Uint64(privateKey[:8])
	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:        customerID,
		PublicKey: publicKey,
		Live:      true,
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
	expectedMetrics := metrics.EmptyServerInitMetrics
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{4, 0, 0},
		CustomerID:     customerID,
		DatacenterID:   crypto.HashID("datacenter.alias"),
		DatacenterName: "datacenter.alias",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

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
*/
