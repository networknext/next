package test

import (
    "net"
    // "time"

    "github.com/networknext/backend/modules/core"
    // "github.com/networknext/backend/modules/crypto"
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
        Version:        sdkVersion,
        BuyerID:        buyerID,
        DatacenterID:   datacenterID,
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
/*
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
    Tags                     [transport.MaxTags]uint64
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
        DirectMinRTT:             config.DirectRTT, // todo: upgrade tests support min/max/prime direct RTT
        DirectMaxRTT:             config.DirectRTT,
        DirectPrimeRTT:           config.DirectRTT,
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

type MatchDataPacketConfig struct {
    Version        transport.SDKVersion
    BuyerID        uint64
    ServerAddress  net.UDPAddr
    DatacenterID   uint64
    UserHash       uint64
    SessionID      uint64
    RetryNumber    uint32
    MatchID        uint64
    NumMatchValues int32
    MatchValues    [transport.MaxMatchValues]float64
    PrivateKey     []byte
}

func (env *TestEnvironment) GenerateMatchDataRequestPacket(config MatchDataPacketConfig) []byte {
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

    requestData, err := transport.MarshalPacket(&requestPacket)
    assert.NoError(env.TestContext, err)

    // If a private key is passed in, sign the packet
    if len(config.PrivateKey) > 0 {
        // We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
        requestDataHeader := append([]byte{transport.PacketTypeMatchDataRequest}, make([]byte, crypto.PacketHashSize)...)
        requestData = append(requestDataHeader, requestData...)
        requestData = crypto.SignPacket(config.PrivateKey, requestData)

        // Once we have the signature, we need to take off the header before passing to the handler
        requestData = requestData[1+crypto.PacketHashSize:]
    }

    return requestData
}
*/
