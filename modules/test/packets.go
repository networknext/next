package test

import (
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func (env *TestEnvironment) GenerateServerInitRequestPacket(sdkVersion transport.SDKVersion, buyerID uint64, datacenterID uint64, datacenterName string, privateKey []byte) []byte {
	requestPacket := transport.ServerInitRequestPacket{
		Version:        sdkVersion,
		BuyerID:        buyerID,
		DatacenterID:   datacenterID,
		DatacenterName: datacenterName,
	}

	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(env.TestContext, err)

	// If a private key is passed in, sign the packet
	if len(privateKey) > 0 {
		// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
		requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
		requestData = append(requestDataHeader, requestData...)
		requestData = crypto.SignPacket(privateKey, requestData)

		// Once we have the signature, we need to take off the header before passing to the handler
		requestData = requestData[1+crypto.PacketHashSize:]
	}

	return requestData
}

func (env *TestEnvironment) GenerateServerUpdatePacket(sdkVersion transport.SDKVersion, buyerID uint64, datacenterID uint64, datacenterName string, numSessions uint32, serverAddress string, privateKey []byte) []byte {
	requestPacket := transport.ServerUpdatePacket{
		Version:       sdkVersion,
		BuyerID:       buyerID,
		DatacenterID:  datacenterID,
		NumSessions:   numSessions,
		ServerAddress: *core.ParseAddress(serverAddress),
	}

	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(env.TestContext, err)

	// If a private key is passed in, sign the packet
	if len(privateKey) > 0 {
		// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
		requestDataHeader := append([]byte{transport.PacketTypeServerUpdate}, make([]byte, crypto.PacketHashSize)...)
		requestData = append(requestDataHeader, requestData...)
		requestData = crypto.SignPacket(privateKey, requestData)

		// Once we have the signature, we need to take off the header before passing to the handler
		requestData = requestData[1+crypto.PacketHashSize:]
	}

	return requestData
}
func (env *TestEnvironment) GenerateEmptySessionUpdatePacket(privateKey []byte) []byte {
	routePublicKey, _, err := core.GenerateRelayKeyPair()
	assert.NoError(env.TestContext, err)

	requestPacket := transport.SessionUpdatePacket{
		ClientRoutePublicKey: routePublicKey,
		ServerRoutePublicKey: routePublicKey,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(env.TestContext, err)

	if len(privateKey) > 0 {
		// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
		requestDataHeader := append([]byte{transport.PacketTypeSessionUpdate}, make([]byte, crypto.PacketHashSize)...)
		requestData = append(requestDataHeader, requestData...)

		// Break the crypto check by not passing in full privat key
		requestData = crypto.SignPacket(privateKey, requestData)

		// Once we have the signature, we need to take off the header before passing to the handler
		requestData = requestData[1+crypto.PacketHashSize:]
	}

	return requestData
}
