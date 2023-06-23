package packets

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"errors"
	"fmt"
	"math/rand"
	"net"

	"github.com/networknext/accelerate/modules/common"
	"github.com/networknext/accelerate/modules/constants"
	"github.com/networknext/accelerate/modules/core"
	"github.com/networknext/accelerate/modules/crypto"
	"github.com/networknext/accelerate/modules/encoding"
)

// ------------------------------------------------------------

func SDK5_SignKeypair(publicKey []byte, privateKey []byte) int {
	result := C.crypto_sign_keypair((*C.uchar)(&publicKey[0]), (*C.uchar)(&privateKey[0]))
	return int(result)
}

func SDK5_SignPacket(packetData []byte, privateKey []byte) {
	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[16]), C.ulonglong(len(packetData)-16-2-SDK5_CRYPTO_SIGN_BYTES))
	C.crypto_sign_final_create(&state, (*C.uchar)(&packetData[len(packetData)-2-SDK5_CRYPTO_SIGN_BYTES]), nil, (*C.uchar)(&privateKey[0]))
}

func SDK5_CheckPacketSignature(packetData []byte, publicKey []byte) bool {

	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[16]), C.ulonglong(len(packetData)-16-2-SDK5_CRYPTO_SIGN_BYTES))
	result := C.crypto_sign_final_verify(&state, (*C.uchar)(&packetData[len(packetData)-2-SDK5_CRYPTO_SIGN_BYTES]), (*C.uchar)(&publicKey[0]))

	if result != 0 {
		core.Error("signed packet did not verify")
		return false
	}

	return true
}

func SDK5_WritePacket[P Packet](packet P, packetType int, maxPacketSize int, from *net.UDPAddr, to *net.UDPAddr, privateKey []byte) ([]byte, error) {

	buffer := make([]byte, maxPacketSize)

	writeStream := encoding.CreateWriteStream(buffer[:])

	var dummy [16]byte
	writeStream.SerializeBytes(dummy[:])

	err := packet.Serialize(writeStream)
	if err != nil {
		return nil, fmt.Errorf("failed to write response packet: %v", err)
	}

	writeStream.Flush()

	packetBytes := writeStream.GetBytesProcessed() + SDK5_CRYPTO_SIGN_BYTES + 2

	packetData := buffer[:packetBytes]

	packetData[0] = uint8(packetType)

	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[16]), C.ulonglong(len(packetData)-16-2-SDK5_CRYPTO_SIGN_BYTES))
	result := C.crypto_sign_final_create(&state, (*C.uchar)(&packetData[len(packetData)-2-SDK5_CRYPTO_SIGN_BYTES]), nil, (*C.uchar)(&privateKey[0]))

	if result != 0 {
		return nil, fmt.Errorf("failed to sign response packet: %d", result)
	}

	var magic [8]byte
	var fromAddressBuffer [32]byte
	var toAddressBuffer [32]byte

	fromAddressData, fromAddressPort := core.GetAddressData(from, fromAddressBuffer[:])
	toAddressData, toAddressPort := core.GetAddressData(to, toAddressBuffer[:])

	core.GenerateChonkle(packetData[1:16], magic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, packetBytes)

	core.GeneratePittle(packetData[packetBytes-2:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, packetBytes)

	return packetData, nil
}

// ------------------------------------------------------------

type SDK5_ServerInitRequestPacket struct {
	Version        SDKVersion
	BuyerId        uint64
	RequestId      uint64
	DatacenterId   uint64
	DatacenterName string
}

func (packet *SDK5_ServerInitRequestPacket) Serialize(stream encoding.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeString(&packet.DatacenterName, SDK5_MaxDatacenterNameLength)
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_ServerInitResponsePacket struct {
	RequestId     uint64
	Response      uint32
	UpcomingMagic [8]byte
	CurrentMagic  [8]byte
	PreviousMagic [8]byte
}

func (packet *SDK5_ServerInitResponsePacket) Serialize(stream encoding.Stream) error {
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeBits(&packet.Response, 8)
	stream.SerializeBytes(packet.UpcomingMagic[:])
	stream.SerializeBytes(packet.CurrentMagic[:])
	stream.SerializeBytes(packet.PreviousMagic[:])
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_ServerUpdateRequestPacket struct {
	Version       SDKVersion
	BuyerId       uint64
	RequestId     uint64
	DatacenterId  uint64
	MatchId       uint64
	NumSessions   uint32
	ServerAddress net.UDPAddr
}

func (packet *SDK5_ServerUpdateRequestPacket) Serialize(stream encoding.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint64(&packet.MatchId)
	stream.SerializeUint32(&packet.NumSessions)
	stream.SerializeAddress(&packet.ServerAddress)
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_ServerUpdateResponsePacket struct {
	RequestId     uint64
	UpcomingMagic [8]byte
	CurrentMagic  [8]byte
	PreviousMagic [8]byte
}

func (packet *SDK5_ServerUpdateResponsePacket) Serialize(stream encoding.Stream) error {
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeBytes(packet.UpcomingMagic[:])
	stream.SerializeBytes(packet.CurrentMagic[:])
	stream.SerializeBytes(packet.PreviousMagic[:])
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_SessionUpdateRequestPacket struct {
	Version                         SDKVersion
	BuyerId                         uint64
	DatacenterId                    uint64
	SessionId                       uint64
	SliceNumber                     uint32
	RetryNumber                     int32
	SessionDataBytes                int32
	SessionData                     [SDK5_MaxSessionDataSize]byte
	SessionDataSignature            [SDK5_SignatureBytes]byte
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
	ClientNextBandwidthOverLimit    bool
	ServerNextBandwidthOverLimit    bool
	ClientPingTimedOut              bool
	HasNearRelayPings               bool
	SessionEvents                   uint64
	InternalEvents                  uint64
	DirectRTT                       float32
	DirectJitter                    float32
	DirectPacketLoss                float32
	DirectMaxPacketLossSeen         float32
	NextRTT                         float32
	NextJitter                      float32
	NextPacketLoss                  float32
	NumNearRelays                   int32
	NearRelayIds                    [SDK5_MaxNearRelays]uint64
	NearRelayRTT                    [SDK5_MaxNearRelays]int32
	NearRelayJitter                 [SDK5_MaxNearRelays]int32
	NearRelayPacketLoss             [SDK5_MaxNearRelays]float32
	DirectKbpsUp                    uint32
	DirectKbpsDown                  uint32
	NextKbpsUp                      uint32
	NextKbpsDown                    uint32
	PacketsSentClientToServer       uint64
	PacketsSentServerToClient       uint64
	PacketsLostClientToServer       uint64
	PacketsLostServerToClient       uint64
	PacketsOutOfOrderClientToServer uint64
	PacketsOutOfOrderServerToClient uint64
	JitterClientToServer            float32
	JitterServerToClient            float32
}

func (packet *SDK5_SessionUpdateRequestPacket) Serialize(stream encoding.Stream) error {

	packet.Version.Serialize(stream)

	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint64(&packet.SessionId)
	stream.SerializeUint32(&packet.SliceNumber)
	stream.SerializeInteger(&packet.RetryNumber, 0, SDK5_MaxSessionUpdateRetries)

	stream.SerializeInteger(&packet.SessionDataBytes, 0, SDK5_MaxSessionDataSize)
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

	stream.SerializeInteger(&packet.PlatformType, SDK5_PlatformTypeUnknown, SDK5_PlatformTypeMax)

	stream.SerializeInteger(&packet.ConnectionType, SDK5_ConnectionTypeUnknown, SDK5_ConnectionTypeMax)

	stream.SerializeBool(&packet.Next)

	stream.SerializeBool(&packet.Reported)
	stream.SerializeBool(&packet.FallbackToDirect)
	stream.SerializeBool(&packet.ClientNextBandwidthOverLimit)
	stream.SerializeBool(&packet.ServerNextBandwidthOverLimit)
	stream.SerializeBool(&packet.ClientPingTimedOut)
	stream.SerializeBool(&packet.HasNearRelayPings)

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

	if packet.HasNearRelayPings {
		stream.SerializeInteger(&packet.NumNearRelays, 0, int32(SDK5_MaxNearRelays))
		for i := int32(0); i < packet.NumNearRelays; i++ {
			stream.SerializeUint64(&packet.NearRelayIds[i])
			if packet.HasNearRelayPings {
				stream.SerializeInteger(&packet.NearRelayRTT[i], 0, SDK5_MaxNearRelayRTT)
				stream.SerializeInteger(&packet.NearRelayJitter[i], 0, SDK5_MaxNearRelayJitter)
				stream.SerializeFloat32(&packet.NearRelayPacketLoss[i])
			}
		}
	}

	stream.SerializeUint32(&packet.DirectKbpsUp)
	stream.SerializeUint32(&packet.DirectKbpsDown)

	if packet.Next {
		stream.SerializeUint32(&packet.NextKbpsUp)
		stream.SerializeUint32(&packet.NextKbpsDown)
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

	return stream.Error()
}

func GenerateRandomSessionData() SDK5_SessionData {

	sessionData := SDK5_SessionData{
		Version:                       uint32(common.RandomInt(SDK5_SessionDataVersion_Min, SDK5_SessionDataVersion_Max)),
		SessionId:                     rand.Uint64(),
		SessionVersion:                uint32(common.RandomInt(0, 255)),
		SliceNumber:                   rand.Uint32(),
		ExpireTimestamp:               rand.Uint64(),
		RouteChanged:                  common.RandomBool(),
		RouteNumRelays:                int32(common.RandomInt(0, SDK5_MaxRelaysPerRoute)),
		RouteCost:                     int32(common.RandomInt(0, SDK5_InvalidRouteValue)),
		FallbackToDirect:              common.RandomBool(),
		PrevPacketsSentClientToServer: rand.Uint64(),
		PrevPacketsSentServerToClient: rand.Uint64(),
		PrevPacketsLostClientToServer: rand.Uint64(),
		PrevPacketsLostServerToClient: rand.Uint64(),
		WriteSummary:                  common.RandomBool(),
		WroteSummary:                  common.RandomBool(),
		NextEnvelopeBytesUpSum:        rand.Uint64(),
		NextEnvelopeBytesDownSum:      rand.Uint64(),
		SessionDuration:               rand.Uint32(),
		StartTimestamp:                rand.Uint64(),
		DurationOnNext:                rand.Uint32(),
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
	sessionData.RouteState.Multipath = common.RandomBool()
	sessionData.RouteState.LatencyWorse = common.RandomBool()
	sessionData.RouteState.NoRoute = common.RandomBool()
	sessionData.RouteState.NextLatencyTooHigh = common.RandomBool()
	sessionData.RouteState.Mispredict = common.RandomBool()
	sessionData.RouteState.RouteLost = common.RandomBool()
	sessionData.RouteState.LackOfDiversity = common.RandomBool()
	sessionData.RouteState.MispredictCounter = uint32(common.RandomInt(0, 3))
	sessionData.RouteState.LatencyWorseCounter = uint32(common.RandomInt(0, 3))
	sessionData.RouteState.LocationVeto = common.RandomBool()

	return sessionData
}

// ------------------------------------------------------------

type SDK5_SessionUpdateResponsePacket struct {
	SessionId                uint64
	SliceNumber              uint32
	SessionDataBytes         int32
	SessionData              [SDK5_MaxSessionDataSize]byte
	SessionDataSignature     [SDK5_SignatureBytes]byte
	RouteType                int32
	HasNearRelays            bool
	NumNearRelays            int32
	NearRelayIds             [SDK5_MaxNearRelays]uint64
	NearRelayAddresses       [SDK5_MaxNearRelays]net.UDPAddr
	NearRelayPingTokens      [SDK5_MaxNearRelays * constants.PingTokenBytes]byte
	NearRelayExpireTimestamp uint64
	NumTokens                int32
	Tokens                   []byte
	Multipath                bool
	HasDebug                 bool
	Debug                    string
}

func (packet *SDK5_SessionUpdateResponsePacket) Serialize(stream encoding.Stream) error {

	stream.SerializeUint64(&packet.SessionId)

	stream.SerializeUint32(&packet.SliceNumber)

	stream.SerializeInteger(&packet.SessionDataBytes, 0, SDK5_MaxSessionDataSize)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
		stream.SerializeBytes(packet.SessionDataSignature[:])
	}

	stream.SerializeInteger(&packet.RouteType, 0, SDK5_RouteTypeContinue)

	stream.SerializeBool(&packet.HasNearRelays)

	if packet.HasNearRelays {
		stream.SerializeInteger(&packet.NumNearRelays, 0, int32(SDK5_MaxNearRelays))
		for i := int32(0); i < packet.NumNearRelays; i++ {
			stream.SerializeUint64(&packet.NearRelayIds[i])
			stream.SerializeAddress(&packet.NearRelayAddresses[i])
			stream.SerializeBytes(packet.NearRelayPingTokens[i*constants.PingTokenBytes : (i+1)*constants.PingTokenBytes])
		}
		stream.SerializeUint64(&packet.NearRelayExpireTimestamp)
	}

	if packet.RouteType != SDK5_RouteTypeDirect {
		stream.SerializeBool(&packet.Multipath)
		stream.SerializeInteger(&packet.NumTokens, 0, SDK5_MaxTokens)
	}

	if packet.RouteType == SDK5_RouteTypeNew {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*SDK5_EncryptedNextRouteTokenSize)
		}
		stream.SerializeBytes(packet.Tokens)
	}

	if packet.RouteType == SDK5_RouteTypeContinue {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*SDK5_EncryptedContinueRouteTokenSize)
		}
		stream.SerializeBytes(packet.Tokens)
	}

	stream.SerializeBool(&packet.HasDebug)
	stream.SerializeString(&packet.Debug, SDK5_MaxSessionDebug)

	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_SessionData struct {
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
	RouteRelayIds                       [SDK5_MaxRelaysPerRoute]uint64
	RouteState                          core.RouteState
	FallbackToDirect                    bool
	PrevPacketsSentClientToServer       uint64
	PrevPacketsSentServerToClient       uint64
	PrevPacketsLostClientToServer       uint64
	PrevPacketsLostServerToClient       uint64
	PrevPacketsOutOfOrderClientToServer uint64
	PrevPacketsOutOfOrderServerToClient uint64
	WriteSummary                        bool
	WroteSummary                        bool
	NextEnvelopeBytesUpSum              uint64
	NextEnvelopeBytesDownSum            uint64
	DurationOnNext                      uint32
	SessionDuration                     uint32
	StartTimestamp                      uint64
}

func (sessionData *SDK5_SessionData) Serialize(stream encoding.Stream) error {

	if stream.IsWriting() {
		if sessionData.Version < SDK5_SessionDataVersion_Min || sessionData.Version > SDK5_SessionDataVersion_Max {
			panic(fmt.Sprintf("invalid session data version: %d", sessionData.Version))
		}
	}

	stream.SerializeBits(&sessionData.Version, 8)

	if stream.IsReading() {
		if sessionData.Version < SDK5_SessionDataVersion_Min || sessionData.Version > SDK5_SessionDataVersion_Max {
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

	stream.SerializeInteger(&sessionData.RouteCost, 0, SDK5_InvalidRouteValue)

	if hasRoute {
		stream.SerializeInteger(&sessionData.RouteNumRelays, 0, SDK5_MaxTokens)
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
	stream.SerializeBool(&sessionData.RouteState.Multipath)
	stream.SerializeBool(&sessionData.RouteState.LatencyWorse)
	stream.SerializeBool(&sessionData.RouteState.NoRoute)
	stream.SerializeBool(&sessionData.RouteState.NextLatencyTooHigh)
	stream.SerializeBool(&sessionData.RouteState.Mispredict)
	stream.SerializeBool(&sessionData.FallbackToDirect)
	stream.SerializeBool(&sessionData.RouteState.RouteLost)
	stream.SerializeBool(&sessionData.RouteState.LackOfDiversity)
	stream.SerializeBits(&sessionData.RouteState.MispredictCounter, 2)
	stream.SerializeBits(&sessionData.RouteState.LatencyWorseCounter, 2)
	stream.SerializeBits(&sessionData.RouteState.PLSustainedCounter, 2)
	stream.SerializeBool(&sessionData.RouteState.LocationVeto)

	stream.SerializeUint64(&sessionData.PrevPacketsSentClientToServer)
	stream.SerializeUint64(&sessionData.PrevPacketsSentServerToClient)
	stream.SerializeUint64(&sessionData.PrevPacketsLostClientToServer)
	stream.SerializeUint64(&sessionData.PrevPacketsLostServerToClient)
	stream.SerializeBool(&sessionData.WriteSummary)
	stream.SerializeBool(&sessionData.WroteSummary)
	stream.SerializeUint64(&sessionData.NextEnvelopeBytesUpSum)
	stream.SerializeUint64(&sessionData.NextEnvelopeBytesDownSum)
	stream.SerializeUint32(&sessionData.SessionDuration)
	stream.SerializeUint32(&sessionData.DurationOnNext)
	stream.SerializeUint64(&sessionData.StartTimestamp)

	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_MatchDataRequestPacket struct {
	Version        SDKVersion
	BuyerId        uint64
	ServerAddress  net.UDPAddr
	DatacenterId   uint64
	UserHash       uint64
	SessionId      uint64
	RetryNumber    uint32
	MatchId        uint64
	NumMatchValues int32
	MatchValues    [SDK5_MaxMatchValues]float64
}

func (packet *SDK5_MatchDataRequestPacket) Serialize(stream encoding.Stream) error {

	packet.Version.Serialize(stream)

	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeAddress(&packet.ServerAddress)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint64(&packet.UserHash)
	stream.SerializeUint64(&packet.SessionId)
	stream.SerializeUint32(&packet.RetryNumber)
	stream.SerializeUint64(&packet.MatchId)

	hasMatchValues := stream.IsWriting() && packet.NumMatchValues > 0

	stream.SerializeBool(&hasMatchValues)

	if hasMatchValues {
		stream.SerializeInteger(&packet.NumMatchValues, 0, SDK5_MaxMatchValues)
		for i := 0; i < int(packet.NumMatchValues); i++ {
			stream.SerializeFloat64(&packet.MatchValues[i])
		}
	}

	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_MatchDataResponsePacket struct {
	SessionId uint64
}

func (packet *SDK5_MatchDataResponsePacket) Serialize(stream encoding.Stream) error {
	stream.SerializeUint64(&packet.SessionId)
	return stream.Error()
}

// ------------------------------------------------------------
