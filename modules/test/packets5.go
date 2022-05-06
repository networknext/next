package test

import (
	"net"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func (env *TestEnvironment) GenerateServerInitRequestPacketSDK5(
	sdkVersion transport.SDKVersion,
	buyerID uint64,
	datacenterID uint64,
	fromAddr *net.UDPAddr,
	toAddr *net.UDPAddr,
	privateKey []byte,
) []byte {
	requestPacket := transport.ServerInitRequestPacketSDK5{
		Version:      sdkVersion,
		BuyerID:      buyerID,
		DatacenterID: datacenterID,
	}

	var emptyMagic [8]byte

	requestData, err := transport.MarshalPacketSDK5(transport.PacketTypeServerInitRequestSDK5, &requestPacket, emptyMagic[:], fromAddr, toAddr, privateKey)
	assert.NoError(env.TestContext, err)

	return requestData
}

func (env *TestEnvironment) GenerateServerUpdatePacketSDK5(
	sdkVersion transport.SDKVersion,
	buyerID uint64,
	datacenterID uint64,
	numSessions uint32,
	fromAddr *net.UDPAddr,
	toAddr *net.UDPAddr,
	privateKey []byte,
) []byte {
	requestPacket := transport.ServerUpdatePacketSDK5{
		Version:       sdkVersion,
		BuyerID:       buyerID,
		DatacenterID:  datacenterID,
		NumSessions:   numSessions,
		ServerAddress: *fromAddr,
	}

	var emptyMagic [8]byte

	requestData, err := transport.MarshalPacketSDK5(transport.PacketTypeServerUpdateSDK5, &requestPacket, emptyMagic[:], fromAddr, toAddr, privateKey)
	assert.NoError(env.TestContext, err)

	return requestData
}

func (env *TestEnvironment) GenerateEmptySessionUpdatePacketSDK5(
	fromAddr *net.UDPAddr,
	toAddr *net.UDPAddr,
	privateKey []byte,
) []byte {
	routePublicKey, _, err := core.GenerateRelayKeyPair()
	assert.NoError(env.TestContext, err)

	requestPacket := transport.SessionUpdatePacketSDK5{
		ClientRoutePublicKey: routePublicKey,
		ServerRoutePublicKey: routePublicKey,
	}

	var emptyMagic [8]byte

	requestData, err := transport.MarshalPacketSDK5(transport.PacketTypeSessionUpdateSDK5, &requestPacket, emptyMagic[:], fromAddr, toAddr, privateKey)
	assert.NoError(env.TestContext, err)

	return requestData
}

type SessionUpdatePacketConfigSDK5 struct {
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
	BackendAddress           string
	UserHash                 uint64
	NearRelayRTT             int32
	NearRelayJitter          int32
	NearRelayPL              int32
	NumTags                  int32
	Tags                     [transport.MaxTags]uint64
	// UserFlags                uint64
	SessionData      [transport.MaxSessionDataSize]byte
	SessionDataBytes int32
	NextRTT          float32
	NextJitter       float32
	NextPacketLoss   float32
	DirectMinRTT     float32
	DirectMaxRTT     float32
	DirectPrimeRTT   float32
	DirectJitter     float32
	DirectPacketLoss float32
}

func (env *TestEnvironment) GenerateSessionUpdatePacketSDK5(config SessionUpdatePacketConfigSDK5) []byte {
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

	requestPacket := transport.SessionUpdatePacketSDK5{
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
		// UserFlags:                config.UserFlags,
		NumNearRelays:       numNearRelays,
		NearRelayIDs:        relayIDs,
		NearRelayRTT:        nearRelayRTT,
		NearRelayJitter:     nearRelayJitter,
		NearRelayPacketLoss: nearRelayPL,
		NextRTT:             config.NextRTT,
		NextJitter:          config.NextJitter,
		NextPacketLoss:      config.NextPacketLoss,
		DirectMinRTT:        config.DirectMinRTT,
		DirectMaxRTT:        config.DirectMaxRTT,
		DirectPrimeRTT:      config.DirectPrimeRTT,
		DirectJitter:        config.DirectJitter,
		DirectPacketLoss:    config.DirectPacketLoss,
	}

	var emptyMagic [8]byte

	requestData, err := transport.MarshalPacketSDK5(transport.PacketTypeSessionUpdateSDK5, &requestPacket, emptyMagic[:], core.ParseAddress(config.ServerAddress), core.ParseAddress(config.BackendAddress), config.PrivateKey)
	assert.NoError(env.TestContext, err)

	return requestData
}

type SessionDataConfigSDK5 struct {
	Version        uint32
	Initial        bool
	SessionID      uint64
	SliceNumber    uint32
	RouteNumRelays int32
	RouteRelayIDs  [5]uint64
	RouteState     core.RouteState
}

func (env *TestEnvironment) GenerateSessionDataPacketSDK5(config SessionDataConfigSDK5) ([511]byte, int) {
	requestSessionData := transport.SessionDataSDK5{
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
	requestSessionDataBytes, err := transport.MarshalSessionDataSDK5(&requestSessionData)
	assert.NoError(env.TestContext, err)

	copy(requestSessionDataBytesFixed[:], requestSessionDataBytes)

	return requestSessionDataBytesFixed, len(requestSessionDataBytes)
}

type MatchDataPacketConfigSDK5 struct {
	Version        transport.SDKVersion
	BuyerID        uint64
	ServerAddress  net.UDPAddr
	BackendAddress net.UDPAddr
	DatacenterID   uint64
	UserHash       uint64
	SessionID      uint64
	RetryNumber    uint32
	MatchID        uint64
	NumMatchValues int32
	MatchValues    [transport.MaxMatchValues]float64
	PrivateKey     []byte
}

func (env *TestEnvironment) GenerateMatchDataRequestPacketSDK5(config MatchDataPacketConfigSDK5) []byte {
	requestPacket := transport.MatchDataRequestPacket{
		Version:        config.Version,
		BuyerID:        config.BuyerID,
		ServerAddress:  config.ServerAddress,
		DatacenterID:   config.DatacenterID,
		UserHash:       config.UserHash,
		SessionID:      config.SessionID,
		RetryNumber:    config.RetryNumber,
		MatchID:        config.MatchID,
		NumMatchValues: config.NumMatchValues,
		MatchValues:    config.MatchValues,
	}

	var emptyMagic [8]byte

	requestData, err := transport.MarshalPacketSDK5(transport.PacketTypeMatchDataRequest, &requestPacket, emptyMagic[:], &config.ServerAddress, &config.BackendAddress, config.PrivateKey)
	assert.NoError(env.TestContext, err)

	return requestData
}
