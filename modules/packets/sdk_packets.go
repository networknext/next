package packets

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"errors"
	"fmt"
	"math/rand"
	"net"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
	"github.com/networknext/next/modules/encoding"
)

// ------------------------------------------------------------

func SDK_SignKeypair(publicKey []byte, privateKey []byte) int {
	result := C.crypto_sign_keypair((*C.uchar)(&publicKey[0]), (*C.uchar)(&privateKey[0]))
	return int(result)
}

func SDK_SignPacket(packetData []byte, privateKey []byte) {
	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[18]), C.ulonglong(len(packetData)-18-SDK_CRYPTO_SIGN_BYTES))
	C.crypto_sign_final_create(&state, (*C.uchar)(&packetData[len(packetData)-SDK_CRYPTO_SIGN_BYTES]), nil, (*C.uchar)(&privateKey[0]))
}

func SDK_CheckPacketSignature(packetData []byte, publicKey []byte) bool {
	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[18]), C.ulonglong(len(packetData)-18-SDK_CRYPTO_SIGN_BYTES))
	result := C.crypto_sign_final_verify(&state, (*C.uchar)(&packetData[len(packetData)-SDK_CRYPTO_SIGN_BYTES]), (*C.uchar)(&publicKey[0]))
	if result != 0 {
		core.Error("signed packet did not verify")
		return false
	}
	return true
}

func SDK_WritePacket[P Packet](packet P, packetType int, maxPacketSize int, from *net.UDPAddr, to *net.UDPAddr, privateKey []byte) ([]byte, error) {

	buffer := make([]byte, maxPacketSize)

	writeStream := encoding.CreateWriteStream(buffer[:])

	var dummy [18]byte
	writeStream.SerializeBytes(dummy[:])

	err := packet.Serialize(writeStream)
	if err != nil {
		return nil, fmt.Errorf("failed to write response packet: %v", err)
	}

	writeStream.Flush()

	packetBytes := writeStream.GetBytesProcessed() + SDK_CRYPTO_SIGN_BYTES

	packetData := buffer[:packetBytes]

	packetData[0] = uint8(packetType)

	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[18]), C.ulonglong(len(packetData)-18-SDK_CRYPTO_SIGN_BYTES))
	result := C.crypto_sign_final_create(&state, (*C.uchar)(&packetData[len(packetData)-SDK_CRYPTO_SIGN_BYTES]), nil, (*C.uchar)(&privateKey[0]))

	if result != 0 {
		return nil, fmt.Errorf("failed to sign response packet: %d", result)
	}

	var magic [8]byte

	fromAddressData := core.GetAddressData(from)
	toAddressData := core.GetAddressData(to)

	core.GeneratePittle(packetData[1:3], fromAddressData, toAddressData, packetBytes)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddressData, toAddressData, packetBytes)

	return packetData, nil
}

// ------------------------------------------------------------

type SDK_ServerInitRequestPacket struct {
	Version        SDKVersion
	BuyerId        uint64
	MatchId        uint64
	RequestId      uint64
	DatacenterId   uint64
	DatacenterName string
}

func (packet *SDK_ServerInitRequestPacket) Serialize(stream encoding.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	if core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 1, 2, 11) {
		stream.SerializeUint64(&packet.MatchId)
	}
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeString(&packet.DatacenterName, SDK_MaxDatacenterNameLength)
	return stream.Error()
}

// ------------------------------------------------------------

type SDK_ServerInitResponsePacket struct {
	RequestId     uint64
	Response      uint32
	UpcomingMagic [8]byte
	CurrentMagic  [8]byte
	PreviousMagic [8]byte
}

func (packet *SDK_ServerInitResponsePacket) Serialize(stream encoding.Stream) error {
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeBits(&packet.Response, 8)
	stream.SerializeBytes(packet.UpcomingMagic[:])
	stream.SerializeBytes(packet.CurrentMagic[:])
	stream.SerializeBytes(packet.PreviousMagic[:])
	return stream.Error()
}

// ------------------------------------------------------------

type SDK_ServerUpdateRequestPacket struct {
	Version      SDKVersion
	BuyerId      uint64
	MatchId      uint64
	RequestId    uint64
	ServerId     uint64
	DatacenterId uint64
	NumSessions  uint32
	Uptime       uint64
	DeltaTimeMin float32
	DeltaTimeMax float32
	DeltaTimeAvg float32
}

func (packet *SDK_ServerUpdateRequestPacket) Serialize(stream encoding.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	if core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 1, 2, 11) {
		stream.SerializeUint64(&packet.MatchId)
	}
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint32(&packet.NumSessions)
	if core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 1, 2, 7) {
		stream.SerializeUint64(&packet.ServerId)
	} else if stream.IsReading() {
		var serverAddress net.UDPAddr
		stream.SerializeAddress(&serverAddress)
		packet.ServerId = common.HashString(serverAddress.String())
	}
	stream.SerializeUint64(&packet.Uptime)
	if core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 1, 2, 6) {
		stream.SerializeFloat32(&packet.DeltaTimeMin)
		stream.SerializeFloat32(&packet.DeltaTimeMax)
		stream.SerializeFloat32(&packet.DeltaTimeAvg)
	}
	return stream.Error()
}

// ------------------------------------------------------------

type SDK_ServerUpdateResponsePacket struct {
	RequestId     uint64
	UpcomingMagic [8]byte
	CurrentMagic  [8]byte
	PreviousMagic [8]byte
}

func (packet *SDK_ServerUpdateResponsePacket) Serialize(stream encoding.Stream) error {
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeBytes(packet.UpcomingMagic[:])
	stream.SerializeBytes(packet.CurrentMagic[:])
	stream.SerializeBytes(packet.PreviousMagic[:])
	return stream.Error()
}

// ------------------------------------------------------------

type SDK_ClientRelayRequestPacket struct {
	Version       SDKVersion
	BuyerId       uint64
	RequestId     uint64
	DatacenterId  uint64
	ClientAddress net.UDPAddr
}

func (packet *SDK_ClientRelayRequestPacket) Serialize(stream encoding.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeAddress(&packet.ClientAddress)
	return stream.Error()
}

// ------------------------------------------------------------

type SDK_ClientRelayResponsePacket struct {
	ClientAddress         net.UDPAddr
	RequestId             uint64
	Latitude              float32
	Longitude             float32
	NumClientRelays       int32
	ClientRelayIds        [constants.MaxClientRelays]uint64
	ClientRelayAddresses  [constants.MaxClientRelays]net.UDPAddr
	ClientRelayPingTokens [constants.MaxClientRelays][constants.PingTokenBytes]byte
	ExpireTimestamp       uint64
}

func (packet *SDK_ClientRelayResponsePacket) Serialize(stream encoding.Stream) error {
	stream.SerializeAddress(&packet.ClientAddress)
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeFloat32(&packet.Latitude)
	stream.SerializeFloat32(&packet.Longitude)
	stream.SerializeInteger(&packet.NumClientRelays, 0, constants.MaxClientRelays)
	for i := 0; i < int(packet.NumClientRelays); i++ {
		stream.SerializeUint64(&packet.ClientRelayIds[i])
		stream.SerializeAddress(&packet.ClientRelayAddresses[i])
		stream.SerializeBytes(packet.ClientRelayPingTokens[i][:])
	}
	stream.SerializeUint64(&packet.ExpireTimestamp)
	return stream.Error()
}

// ------------------------------------------------------------

type SDK_ServerRelayRequestPacket struct {
	Version      SDKVersion
	BuyerId      uint64
	RequestId    uint64
	DatacenterId uint64
}

func (packet *SDK_ServerRelayRequestPacket) Serialize(stream encoding.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeUint64(&packet.DatacenterId)
	return stream.Error()
}

// ------------------------------------------------------------

type SDK_ServerRelayResponsePacket struct {
	RequestId             uint64
	NumServerRelays       int32
	ServerRelayIds        [constants.MaxServerRelays]uint64
	ServerRelayAddresses  [constants.MaxServerRelays]net.UDPAddr
	ServerRelayPingTokens [constants.MaxServerRelays][constants.PingTokenBytes]byte
	ExpireTimestamp       uint64
}

func (packet *SDK_ServerRelayResponsePacket) Serialize(stream encoding.Stream) error {
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeInteger(&packet.NumServerRelays, 0, constants.MaxServerRelays)
	for i := 0; i < int(packet.NumServerRelays); i++ {
		stream.SerializeUint64(&packet.ServerRelayIds[i])
		stream.SerializeAddress(&packet.ServerRelayAddresses[i])
		stream.SerializeBytes(packet.ServerRelayPingTokens[i][:])
	}
	stream.SerializeUint64(&packet.ExpireTimestamp)
	return stream.Error()
}

// ------------------------------------------------------------

type SDK_SessionUpdateRequestPacket struct {
	Version                         SDKVersion
	BuyerId                         uint64
	MatchId                         uint64
	DatacenterId                    uint64
	SessionId                       uint64
	SliceNumber                     uint32
	RetryNumber                     int32
	SessionDataBytes                int32
	SessionData                     [SDK_MaxSessionDataSize]byte
	SessionDataSignature            [SDK_SignatureBytes]byte
	ClientAddress                   net.UDPAddr
	ServerAddress                   net.UDPAddr
	ClientRoutePublicKey            [crypto.Box_PublicKeySize]byte
	ServerRoutePublicKey            [crypto.Box_PublicKeySize]byte
	UserHash                        uint64
	PlatformType                    int32
	ConnectionType                  int32
	Next                            bool
	Reported                        bool
	FallbackToDirect                bool
	ClientPingTimedOut              bool
	HasClientRelayPings             bool
	HasServerRelayPings             bool
	ClientRelayPingsHaveChanged     bool
	ServerRelayPingsHaveChanged     bool
	SessionEvents                   uint64
	InternalEvents                  uint64
	DirectRTT                       float32
	DirectJitter                    float32
	DirectPacketLoss                float32
	DirectMaxPacketLossSeen         float32
	NextRTT                         float32
	NextJitter                      float32
	NextPacketLoss                  float32
	NumClientRelays                 int32
	ClientRelayIds                  [SDK_MaxClientRelays]uint64
	ClientRelayRTT                  [SDK_MaxClientRelays]int32
	ClientRelayJitter               [SDK_MaxClientRelays]int32
	ClientRelayPacketLoss           [SDK_MaxClientRelays]float32
	NumServerRelays                 int32
	ServerRelayIds                  [SDK_MaxServerRelays]uint64
	ServerRelayRTT                  [SDK_MaxServerRelays]int32
	ServerRelayJitter               [SDK_MaxServerRelays]int32
	ServerRelayPacketLoss           [SDK_MaxServerRelays]float32
	BandwidthKbpsUp                 uint32
	BandwidthKbpsDown               uint32
	PacketsSentClientToServer       uint64
	PacketsSentServerToClient       uint64
	PacketsLostClientToServer       uint64
	PacketsLostServerToClient       uint64
	PacketsOutOfOrderClientToServer uint64
	PacketsOutOfOrderServerToClient uint64
	JitterClientToServer            float32
	JitterServerToClient            float32
	Latitude                        float32
	Longitude                       float32
	DeltaTimeMin                    float32
	DeltaTimeMax                    float32
	DeltaTimeAvg                    float32
	GameRTT                         float32
	GameJitter                      float32
	GamePacketLoss                  float32
	Flags                           uint32
}

func (packet *SDK_SessionUpdateRequestPacket) Serialize(stream encoding.Stream) error {

	packet.Version.Serialize(stream)

	stream.SerializeUint64(&packet.BuyerId)
	if core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 1, 2, 11) {
		stream.SerializeUint64(&packet.MatchId)
	}
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint64(&packet.SessionId)
	stream.SerializeUint32(&packet.SliceNumber)
	stream.SerializeInteger(&packet.RetryNumber, 0, SDK_MaxSessionUpdateRetries)

	stream.SerializeInteger(&packet.SessionDataBytes, 0, SDK_MaxSessionDataSize)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
		stream.SerializeBytes(packet.SessionDataSignature[:])
	}

	stream.SerializeAddress(&packet.ClientAddress)

	stream.SerializeAddress(&packet.ServerAddress)

	stream.SerializeBytes(packet.ClientRoutePublicKey[:])

	stream.SerializeBytes(packet.ServerRoutePublicKey[:])

	stream.SerializeUint64(&packet.UserHash)

	stream.SerializeInteger(&packet.PlatformType, SDK_PlatformTypeUnknown, SDK_PlatformTypeMax)

	stream.SerializeInteger(&packet.ConnectionType, SDK_ConnectionTypeUnknown, SDK_ConnectionTypeMax)

	stream.SerializeBool(&packet.Next)

	stream.SerializeBool(&packet.Reported)
	stream.SerializeBool(&packet.FallbackToDirect)

	if !core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 1, 2, 9) {
		var clientNextBandwidthOverLimit bool
		var serverNextBandwidthOverLimit bool
		stream.SerializeBool(&clientNextBandwidthOverLimit)
		stream.SerializeBool(&serverNextBandwidthOverLimit)
	}

	stream.SerializeBool(&packet.ClientPingTimedOut)
	stream.SerializeBool(&packet.HasClientRelayPings)
	stream.SerializeBool(&packet.HasServerRelayPings)
	stream.SerializeBool(&packet.ClientRelayPingsHaveChanged)
	stream.SerializeBool(&packet.ServerRelayPingsHaveChanged)

	hasSessionEvents := stream.IsWriting() && packet.SessionEvents != 0
	hasInternalEvents := stream.IsWriting() && packet.InternalEvents != 0
	hasLostPackets := stream.IsWriting() && (packet.PacketsLostClientToServer+packet.PacketsLostServerToClient) > 0
	hasOutOfOrderPackets := stream.IsWriting() && (packet.PacketsOutOfOrderClientToServer+packet.PacketsOutOfOrderServerToClient) > 0

	stream.SerializeBool(&hasSessionEvents)
	stream.SerializeBool(&hasInternalEvents)
	stream.SerializeBool(&hasLostPackets)
	stream.SerializeBool(&hasOutOfOrderPackets)

	if hasSessionEvents {
		stream.SerializeUint64(&packet.SessionEvents)
	}

	if hasInternalEvents {
		stream.SerializeUint64(&packet.InternalEvents)
	}

	stream.SerializeFloat32(&packet.DirectRTT)
	stream.SerializeFloat32(&packet.DirectJitter)
	stream.SerializeFloat32(&packet.DirectPacketLoss)
	stream.SerializeFloat32(&packet.DirectMaxPacketLossSeen)

	if packet.Next {
		stream.SerializeFloat32(&packet.NextRTT)
		stream.SerializeFloat32(&packet.NextJitter)
		stream.SerializeFloat32(&packet.NextPacketLoss)
	}

	if packet.HasClientRelayPings {
		stream.SerializeInteger(&packet.NumClientRelays, 0, int32(SDK_MaxClientRelays))
		for i := int32(0); i < packet.NumClientRelays; i++ {
			stream.SerializeUint64(&packet.ClientRelayIds[i])
			if packet.HasClientRelayPings {
				stream.SerializeInteger(&packet.ClientRelayRTT[i], 0, SDK_MaxRelayRTT)
				stream.SerializeInteger(&packet.ClientRelayJitter[i], 0, SDK_MaxRelayJitter)
				stream.SerializeFloat32(&packet.ClientRelayPacketLoss[i])
			}
		}
	}

	if packet.HasServerRelayPings {
		stream.SerializeInteger(&packet.NumServerRelays, 0, int32(SDK_MaxServerRelays))
		for i := int32(0); i < packet.NumServerRelays; i++ {
			stream.SerializeUint64(&packet.ServerRelayIds[i])
			if packet.HasServerRelayPings {
				stream.SerializeInteger(&packet.ServerRelayRTT[i], 0, SDK_MaxRelayRTT)
				stream.SerializeInteger(&packet.ServerRelayJitter[i], 0, SDK_MaxRelayJitter)
				stream.SerializeFloat32(&packet.ServerRelayPacketLoss[i])
			}
		}
	}

	stream.SerializeUint32(&packet.BandwidthKbpsUp)
	stream.SerializeUint32(&packet.BandwidthKbpsDown)

	if !core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 1, 2, 9) {
		if packet.Next {
			var nextKbpsUp uint32
			var nextKbpsDown uint32
			stream.SerializeUint32(&nextKbpsUp)
			stream.SerializeUint32(&nextKbpsDown)
		}
	}

	stream.SerializeUint64(&packet.PacketsSentClientToServer)
	stream.SerializeUint64(&packet.PacketsSentServerToClient)

	if hasLostPackets {
		stream.SerializeUint64(&packet.PacketsLostClientToServer)
		stream.SerializeUint64(&packet.PacketsLostServerToClient)
	}

	if hasOutOfOrderPackets {
		stream.SerializeUint64(&packet.PacketsOutOfOrderClientToServer)
		stream.SerializeUint64(&packet.PacketsOutOfOrderServerToClient)
	}

	stream.SerializeFloat32(&packet.JitterClientToServer)
	stream.SerializeFloat32(&packet.JitterServerToClient)

	if core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 1, 2, 6) {
		stream.SerializeFloat32(&packet.Latitude)
		stream.SerializeFloat32(&packet.Longitude)
		stream.SerializeFloat32(&packet.DeltaTimeMin)
		stream.SerializeFloat32(&packet.DeltaTimeMax)
		stream.SerializeFloat32(&packet.DeltaTimeAvg)
	}

	if core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 1, 2, 8) {
		stream.SerializeFloat32(&packet.GameRTT)
		stream.SerializeFloat32(&packet.GameJitter)
		stream.SerializeFloat32(&packet.GamePacketLoss)
	}

	if core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 1, 2, 10) {
		stream.SerializeUint32(&packet.Flags)
	}

	return stream.Error()
}

func GenerateRandomSessionData() SDK_SessionData {

	sessionData := SDK_SessionData{
		Version:                        uint32(common.RandomInt(SDK_SessionDataVersion_Min, SDK_SessionDataVersion_Max)),
		SessionId:                      rand.Uint64(),
		SessionVersion:                 uint32(common.RandomInt(0, 255)),
		SliceNumber:                    rand.Uint32(),
		ExpireTimestamp:                rand.Uint64(),
		RouteChanged:                   common.RandomBool(),
		RouteNumRelays:                 int32(common.RandomInt(0, SDK_MaxRelaysPerRoute)),
		RouteCost:                      int32(common.RandomInt(0, SDK_InvalidRouteValue)),
		PrevPacketsSentClientToServer:  rand.Uint64(),
		PrevPacketsSentServerToClient:  rand.Uint64(),
		PrevPacketsLostClientToServer:  rand.Uint64(),
		PrevPacketsLostServerToClient:  rand.Uint64(),
		WriteSummary:                   common.RandomBool(),
		WroteSummary:                   common.RandomBool(),
		ShouldSendClientRelaysToPortal: common.RandomBool(),
		ShouldSendServerRelaysToPortal: common.RandomBool(),
		SentClientRelaysToPortal:       common.RandomBool(),
		SentServerRelaysToPortal:       common.RandomBool(),
		NextEnvelopeBytesUpSum:         rand.Uint64(),
		NextEnvelopeBytesDownSum:       rand.Uint64(),
		StartTimestamp:                 rand.Uint64(),
		DurationOnNext:                 rand.Uint32(),
		Error:                          rand.Uint64(),
		BestScore:                      uint32(common.RandomInt(0, 999)),
		BestDirectRTT:                  uint32(common.RandomInt(0, 500)),
		BestNextRTT:                    uint32(common.RandomInt(0, 500)),
	}

	for i := 0; i < int(sessionData.RouteNumRelays); i++ {
		sessionData.RouteRelayIds[i] = rand.Uint64()
	}

	sessionData.Latitude = rand.Float32()
	sessionData.Longitude = rand.Float32()

	sessionData.RouteState.Next = common.RandomBool()
	sessionData.RouteState.Veto = common.RandomBool()
	sessionData.RouteState.Disabled = common.RandomBool()
	sessionData.RouteState.NotSelected = common.RandomBool()
	sessionData.RouteState.ABTest = common.RandomBool()
	sessionData.RouteState.A = common.RandomBool()
	sessionData.RouteState.B = common.RandomBool()
	sessionData.RouteState.ForcedNext = common.RandomBool()
	sessionData.RouteState.ReduceLatency = common.RandomBool()
	sessionData.RouteState.ReducePacketLoss = common.RandomBool()
	sessionData.RouteState.LatencyWorse = common.RandomBool()
	sessionData.RouteState.NoRoute = common.RandomBool()
	sessionData.RouteState.NextLatencyTooHigh = common.RandomBool()
	sessionData.RouteState.Mispredict = common.RandomBool()
	sessionData.RouteState.RouteLost = common.RandomBool()
	sessionData.RouteState.LackOfDiversity = common.RandomBool()
	sessionData.RouteState.MispredictCounter = uint32(common.RandomInt(0, 3))
	sessionData.RouteState.LatencyWorseCounter = uint32(common.RandomInt(0, 3))

	for i := range sessionData.ExcludeClientRelay {
		sessionData.ExcludeClientRelay[i] = common.RandomBool()
	}

	for i := range sessionData.ExcludeServerRelay {
		sessionData.ExcludeServerRelay[i] = common.RandomBool()
	}

	return sessionData
}

// ------------------------------------------------------------

type SDK_SessionUpdateResponsePacket struct {
	SessionId            uint64
	SliceNumber          uint32
	SessionDataBytes     int32
	SessionData          [SDK_MaxSessionDataSize]byte
	SessionDataSignature [SDK_SignatureBytes]byte
	RouteType            int32
	NumTokens            int32
	Tokens               []byte
	Multipath            bool
}

func (packet *SDK_SessionUpdateResponsePacket) Serialize(stream encoding.Stream) error {

	stream.SerializeUint64(&packet.SessionId)

	stream.SerializeUint32(&packet.SliceNumber)

	stream.SerializeInteger(&packet.SessionDataBytes, 0, SDK_MaxSessionDataSize)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
		stream.SerializeBytes(packet.SessionDataSignature[:])
	}

	stream.SerializeInteger(&packet.RouteType, 0, SDK_RouteTypeContinue)

	if packet.RouteType != SDK_RouteTypeDirect {
		stream.SerializeBool(&packet.Multipath)
		stream.SerializeInteger(&packet.NumTokens, 0, SDK_MaxTokens)
	}

	if packet.RouteType == SDK_RouteTypeNew {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*SDK_EncryptedNextRouteTokenSize)
		}
		stream.SerializeBytes(packet.Tokens)
	}

	if packet.RouteType == SDK_RouteTypeContinue {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*SDK_EncryptedContinueRouteTokenSize)
		}
		stream.SerializeBytes(packet.Tokens)
	}

	return stream.Error()
}

// ------------------------------------------------------------

type SDK_SessionData struct {
	Version                             uint32
	SessionId                           uint64
	SessionVersion                      uint32
	SliceNumber                         uint32
	ExpireTimestamp                     uint64
	Latitude                            float32
	Longitude                           float32
	RouteChanged                        bool
	RouteNumRelays                      int32
	RouteCost                           int32
	RouteRelayIds                       [SDK_MaxRelaysPerRoute]uint64
	RouteState                          core.RouteState
	WriteSummary                        bool
	WroteSummary                        bool
	ShouldSendClientRelaysToPortal      bool
	ShouldSendServerRelaysToPortal      bool
	SentClientRelaysToPortal            bool
	SentServerRelaysToPortal            bool
	PrevPacketsSentClientToServer       uint64
	PrevPacketsSentServerToClient       uint64
	PrevPacketsLostClientToServer       uint64
	PrevPacketsLostServerToClient       uint64
	PrevPacketsOutOfOrderClientToServer uint64
	PrevPacketsOutOfOrderServerToClient uint64
	NextEnvelopeBytesUpSum              uint64
	NextEnvelopeBytesDownSum            uint64
	DurationOnNext                      uint32
	StartTimestamp                      uint64
	Error                               uint64
	BestScore                           uint32
	BestDirectRTT                       uint32
	BestNextRTT                         uint32
	ExcludeClientRelay                  [SDK_MaxClientRelays]bool
	ExcludeServerRelay                  [SDK_MaxServerRelays]bool
	LikelyVPNOrCrossRegion              bool
	NoClientRelays                      bool
	NoServerRelays                      bool
	AllClientRelaysAreZero              bool
}

func (sessionData *SDK_SessionData) Serialize(stream encoding.Stream) error {

	if stream.IsWriting() {
		if sessionData.Version < SDK_SessionDataVersion_Min || sessionData.Version > SDK_SessionDataVersion_Max {
			panic(fmt.Sprintf("invalid session data version: %d", sessionData.Version))
		}
	}

	stream.SerializeBits(&sessionData.Version, 8)

	if stream.IsReading() {
		if sessionData.Version < SDK_SessionDataVersion_Min || sessionData.Version > SDK_SessionDataVersion_Max {
			return errors.New(fmt.Sprintf("invalid session data version: %d", sessionData.Version))
		}
	}

	stream.SerializeUint64(&sessionData.SessionId)
	stream.SerializeBits(&sessionData.SessionVersion, 8)

	stream.SerializeUint32(&sessionData.SliceNumber)

	stream.SerializeUint64(&sessionData.ExpireTimestamp)

	stream.SerializeFloat32(&sessionData.Latitude)
	stream.SerializeFloat32(&sessionData.Longitude)

	stream.SerializeBool(&sessionData.RouteChanged)

	hasRoute := sessionData.RouteNumRelays > 0

	stream.SerializeBool(&hasRoute)

	stream.SerializeInteger(&sessionData.RouteCost, 0, SDK_InvalidRouteValue)

	if hasRoute {
		stream.SerializeInteger(&sessionData.RouteNumRelays, 0, SDK_MaxTokens)
		for i := int32(0); i < sessionData.RouteNumRelays; i++ {
			stream.SerializeUint64(&sessionData.RouteRelayIds[i])
		}
	}

	stream.SerializeBool(&sessionData.RouteState.Next)
	stream.SerializeBool(&sessionData.RouteState.Veto)
	stream.SerializeBool(&sessionData.RouteState.Disabled)
	stream.SerializeBool(&sessionData.RouteState.NotSelected)
	stream.SerializeBool(&sessionData.RouteState.ABTest)
	stream.SerializeBool(&sessionData.RouteState.A)
	stream.SerializeBool(&sessionData.RouteState.B)
	stream.SerializeBool(&sessionData.RouteState.ForcedNext)
	stream.SerializeBool(&sessionData.RouteState.ReduceLatency)
	stream.SerializeBool(&sessionData.RouteState.ReducePacketLoss)

	if sessionData.Version < 6 {
		var multipath bool
		stream.SerializeBool(&multipath)
	}

	stream.SerializeBool(&sessionData.RouteState.LatencyWorse)
	stream.SerializeBool(&sessionData.RouteState.NoRoute)
	stream.SerializeBool(&sessionData.RouteState.NextLatencyTooHigh)
	stream.SerializeBool(&sessionData.RouteState.Mispredict)
	stream.SerializeBool(&sessionData.RouteState.RouteLost)
	stream.SerializeBool(&sessionData.RouteState.LackOfDiversity)
	stream.SerializeBits(&sessionData.RouteState.MispredictCounter, 2)
	stream.SerializeBits(&sessionData.RouteState.LatencyWorseCounter, 2)
	stream.SerializeBits(&sessionData.RouteState.PLSustainedCounter, 2)

	stream.SerializeUint64(&sessionData.PrevPacketsSentClientToServer)
	stream.SerializeUint64(&sessionData.PrevPacketsSentServerToClient)
	stream.SerializeUint64(&sessionData.PrevPacketsLostClientToServer)
	stream.SerializeUint64(&sessionData.PrevPacketsLostServerToClient)
	stream.SerializeBool(&sessionData.WriteSummary)
	stream.SerializeBool(&sessionData.WroteSummary)
	stream.SerializeBool(&sessionData.ShouldSendClientRelaysToPortal)
	stream.SerializeBool(&sessionData.ShouldSendServerRelaysToPortal)
	stream.SerializeBool(&sessionData.SentClientRelaysToPortal)
	stream.SerializeBool(&sessionData.SentServerRelaysToPortal)
	stream.SerializeUint64(&sessionData.NextEnvelopeBytesUpSum)
	stream.SerializeUint64(&sessionData.NextEnvelopeBytesDownSum)
	stream.SerializeUint32(&sessionData.DurationOnNext)
	stream.SerializeUint64(&sessionData.StartTimestamp)
	stream.SerializeUint64(&sessionData.Error)

	stream.SerializeBits(&sessionData.BestScore, 10)
	stream.SerializeBits(&sessionData.BestDirectRTT, 10)
	stream.SerializeBits(&sessionData.BestNextRTT, 10)

	for i := 0; i < SDK_MaxClientRelays; i++ {
		stream.SerializeBool(&sessionData.ExcludeClientRelay[i])
	}

	for i := 0; i < SDK_MaxServerRelays; i++ {
		stream.SerializeBool(&sessionData.ExcludeServerRelay[i])
	}

	if sessionData.Version >= 2 {
		stream.SerializeBool(&sessionData.LikelyVPNOrCrossRegion)
	}

	if sessionData.Version >= 3 {
		stream.SerializeBool(&sessionData.NoClientRelays)
	}

	if sessionData.Version >= 4 {
		stream.SerializeBool(&sessionData.NoServerRelays)
	}

	if sessionData.Version >= 5 {
		stream.SerializeBool(&sessionData.AllClientRelaysAreZero)
	}

	return stream.Error()
}

// ------------------------------------------------------------
