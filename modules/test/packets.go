package test

import (
	"time"

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

type SessionUpdatePacketConfig struct {
	Version                  transport.SDKVersion
	BuyerID                  uint64
	DatacenterID             uint64
	SessionID                uint64
	SliceNumber              uint32
	RetryNumber              int32
	Next                     bool
	PublicKey                []byte
	PrivateKey               []byte
	Committed                bool
	Reported                 bool
	FallbackToDirect         bool
	ClientBandwidthOverLimit bool
	ServerBandwidthOverLimit bool
	ClientPingTimedOut       bool
	ClientAddress            string
	ServerAddress            string
	UserHash                 uint64
	NearRelayRTT             int32
	NearRelayJitter          int32
	NearRelayPL              int32
	NumTags                  int32
	Tags                     [8]uint64
	Flags                    uint32
	UserFlags                uint64
	SessionData              [transport.MaxSessionDataSize]byte
	SessionDataBytes         int32
	NextRTT                  float32
	NextJitter               float32
	NextPacketLoss           float32
	DirectRTT                float32
	DirectJitter             float32
	DirectPacketLoss         float32
}

func (env *TestEnvironment) GenerateSessionUpdatePacket(config SessionUpdatePacketConfig) []byte {
	relayIDs := env.GetRelayIds()
	numNearRelays := int32(len(relayIDs))
	nearRelayRTT := make([]int32, numNearRelays)
	nearRelayJitter := make([]int32, numNearRelays)
	nearRelayPL := make([]int32, numNearRelays)

	for i := range relayIDs {
		nearRelayRTT[i] = config.NearRelayRTT
		nearRelayJitter[i] = config.NearRelayJitter
		nearRelayPL[i] = config.NearRelayPL
	}

	requestPacket := transport.SessionUpdatePacket{
		Version:                  config.Version,
		BuyerID:                  config.BuyerID,
		DatacenterID:             config.DatacenterID,
		SessionID:                config.SessionID,
		SliceNumber:              config.SliceNumber,
		RetryNumber:              config.RetryNumber,
		Next:                     config.Next,
		ClientRoutePublicKey:     config.PublicKey,
		ServerRoutePublicKey:     config.PublicKey,
		ClientAddress:            *core.ParseAddress(config.ClientAddress),
		ServerAddress:            *core.ParseAddress(config.ServerAddress),
		UserHash:                 config.UserHash,
		PlatformType:             0,
		ConnectionType:           0,
		SessionDataBytes:         config.SessionDataBytes,
		SessionData:              config.SessionData,
		Committed:                config.Committed,
		Reported:                 config.Reported,
		FallbackToDirect:         config.FallbackToDirect,
		ClientBandwidthOverLimit: config.ClientBandwidthOverLimit,
		ServerBandwidthOverLimit: config.ServerBandwidthOverLimit,
		ClientPingTimedOut:       config.ClientPingTimedOut,
		NumTags:                  config.NumTags,
		Tags:                     config.Tags,
		Flags:                    config.Flags,
		UserFlags:                config.UserFlags,
		NumNearRelays:            numNearRelays,
		NearRelayIDs:             relayIDs,
		NearRelayRTT:             nearRelayRTT,
		NearRelayJitter:          nearRelayJitter,
		NearRelayPacketLoss:      nearRelayPL,
		NextRTT:                  config.NextRTT,
		NextJitter:               config.NextJitter,
		NextPacketLoss:           config.NextPacketLoss,
		DirectRTT:                config.DirectRTT,
		DirectJitter:             config.DirectJitter,
		DirectPacketLoss:         config.DirectPacketLoss,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(env.TestContext, err)

	if len(config.PrivateKey) > 0 {
		// Add the packet type byte and hash bytes to the request data so we can sign it properly
		requestDataHeader := append([]byte{transport.PacketTypeSessionUpdate}, make([]byte, crypto.PacketHashSize)...)
		requestData = append(requestDataHeader, requestData...)

		// Sign the packet
		requestData = crypto.SignPacket(config.PrivateKey, requestData)

		// Once the packet is signed, we need to remove the header before passing to the session update handler
		requestData = requestData[1+crypto.PacketHashSize:]
	}

	return requestData
}

type SessionDataConfig struct {
	Version        uint32
	Initial        bool
	SessionID      uint64
	SliceNumber    uint32
	RouteNumRelays int32
	RouteRelayIDs  [5]uint64
	RouteState     core.RouteState
}

func (env *TestEnvironment) GenerateSessionDataPacket(config SessionDataConfig) ([511]byte, int) {
	requestSessionData := transport.SessionData{
		Version:         config.Version,
		ExpireTimestamp: uint64(time.Now().Add(time.Minute * 1).Unix()),
		SessionID:       config.SessionID,
		SliceNumber:     config.SliceNumber,
		Initial:         config.Initial,
		RouteNumRelays:  config.RouteNumRelays,
		RouteRelayIDs:   config.RouteRelayIDs,
		RouteState:      config.RouteState,
	}

	var requestSessionDataBytesFixed [511]byte
	requestSessionDataBytes, err := transport.MarshalSessionData(&requestSessionData)
	assert.NoError(env.TestContext, err)

	copy(requestSessionDataBytesFixed[:], requestSessionDataBytes)

	return requestSessionDataBytesFixed, len(requestSessionDataBytes)
}
