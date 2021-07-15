package transport

import (
	"errors"
	"math"
	"net"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
)

const (
	DefaultMaxPacketSize = 4096

	MaxDatacenterNameLength = 256
	MaxSessionUpdateRetries = 10

	SessionDataVersion = 13

	MaxSessionDataSize = 511

	MaxTokens = 7

	PacketTypeServerUpdate       = 220
	PacketTypeSessionUpdate      = 221
	PacketTypeSessionResponse    = 222
	PacketTypeServerInitRequest  = 223
	PacketTypeServerInitResponse = 224

	InitResponseOK                   = 0
	InitResponseUnknownBuyer         = 1
	InitResponseUnknownDatacenter    = 2
	InitResponseOldSDKVersion        = 3
	InitResponseSignatureCheckFailed = 4
	InitResponseBuyerNotActive       = 5
	InitResponseDataCenterNotEnabled = 6

	// IMPORTANT: Update Serialize(), Validate(), and ClampEntry() in modules/billing/billing_entry.go when a new connection type is added
	ConnectionTypeUnknown  = 0
	ConnectionTypeWired    = 1
	ConnectionTypeWifi     = 2
	ConnectionTypeCellular = 3
	ConnectionTypeMax      = 3

	// IMPORTANT: Update Serialize(), Validate(), and ClampEntry() in modules/billing/billing_entry.go when a new platform type is added
	PlatformTypeUnknown     = 0
	PlatformTypeWindows     = 1
	PlatformTypeMac         = 2
	PlatformTypeLinux       = 3
	PlatformTypeSwitch      = 4
	PlatformTypePS4         = 5
	PlatformTypeIOS         = 6
	PlatformTypeXBoxOne     = 7
	PlatformTypeMax_404     = 7 // SDK 4.0.4 and older
	PlatformTypeXBoxSeriesX = 8
	PlatformTypePS5         = 9
	PlatformTypeMax_405     = 9 // SDK 4.0.5 and newer
	PlatformTypeGDK         = 10
	PlatformTypeMax_410     = 10 // SDK 4.0.10 and newer
	PlatformTypeMax         = 10

	FallbackFlagsBadRouteToken              = (1 << 0)
	FallbackFlagsNoNextRouteToContinue      = (1 << 1)
	FallbackFlagsPreviousUpdateStillPending = (1 << 2)
	FallbackFlagsBadContinueToken           = (1 << 3)
	FallbackFlagsRouteExpired               = (1 << 4)
	FallbackFlagsRouteRequestTimedOut       = (1 << 5)
	FallbackFlagsContinueRequestTimedOut    = (1 << 6)
	FallbackFlagsClientTimedOut             = (1 << 7)
	FallbackFlagsUpgradeResponseTimedOut    = (1 << 8)
	FallbackFlagsRouteUpdateTimedOut        = (1 << 9)
	FallbackFlagsDirectPongTimedOut         = (1 << 10)
	FallbackFlagsNextPongTimedOut           = (1 << 11)
	FallbackFlagsCount_400                  = 11
	FallbackFlagsCount_401                  = 12

	MaxTags = 8

	NextMaxSessionDebug = 1024
)

// ConnectionTypeText is similar to http.StatusText(int) which converts the code to a readable text format
func ConnectionTypeText(conntype uint8) string {
	switch conntype {
	case ConnectionTypeWired:
		return "wired"
	case ConnectionTypeWifi:
		return "wifi"
	case ConnectionTypeCellular:
		return "cellular"
	default:
		return "unknown"
	}
}

func ParseConnectionType(conntype string) uint8 {
	switch conntype {
	case "wired":
		return ConnectionTypeWired
	case "wifi":
		return ConnectionTypeWifi
	case "cellular":
		return ConnectionTypeCellular
	default:
		return ConnectionTypeUnknown
	}
}

// PlatformTypeText is similar to http.StatusText(int) which converts the code to a readable text format
func PlatformTypeText(platformType uint8) string {
	switch platformType {
	case PlatformTypeWindows:
		return "Windows"
	case PlatformTypeMac:
		return "Mac"
	case PlatformTypeLinux:
		return "Linux"
	case PlatformTypeSwitch:
		return "Switch"
	case PlatformTypePS4:
		return "PS4"
	case PlatformTypeIOS:
		return "IOS"
	case PlatformTypeXBoxOne:
		return "XBox One"
	case PlatformTypeXBoxSeriesX:
		return "XBox Series X"
	case PlatformTypePS5:
		return "PS5"
	case PlatformTypeGDK:
		return "GDK"
	default:
		return "unknown"
	}
}

func ParsePlatformType(conntype string) uint8 {
	switch conntype {
	case "Windows":
		return PlatformTypeWindows
	case "Mac":
		return PlatformTypeMac
	case "Linux":
		return PlatformTypeLinux
	case "Switch":
		return PlatformTypeSwitch
	case "PS4":
		return PlatformTypePS4
	case "IOS":
		return PlatformTypeIOS
	case "XBox One":
		return PlatformTypeXBoxOne
	case "XBox Series X":
		return PlatformTypeXBoxSeriesX
	case "PS5":
		return PlatformTypePS5
	case "GDK":
		return PlatformTypeGDK
	default:
		return PlatformTypeUnknown
	}
}

// FallbackFlagText is similar to http.StatusText(int) which converts the code to a readable text format
func FallbackFlagText(fallbackFlag uint32) string {
	switch fallbackFlag {
	case FallbackFlagsBadRouteToken:
		return "bad route token"
	case FallbackFlagsNoNextRouteToContinue:
		return "no next route to continue"
	case FallbackFlagsPreviousUpdateStillPending:
		return "previous update still pending"
	case FallbackFlagsBadContinueToken:
		return "bad continue token"
	case FallbackFlagsRouteExpired:
		return "route expired"
	case FallbackFlagsRouteRequestTimedOut:
		return "route request timed out"
	case FallbackFlagsContinueRequestTimedOut:
		return "continue request timed out"
	case FallbackFlagsClientTimedOut:
		return "client timed out"
	case FallbackFlagsUpgradeResponseTimedOut:
		return "upgrade response timed out"
	case FallbackFlagsRouteUpdateTimedOut:
		return "route update timed out"
	case FallbackFlagsDirectPongTimedOut:
		return "direct pong timed out"
	default:
		return "unknown"
	}
}

type Packet interface {
	Serialize(stream encoding.Stream) error
}

func UnmarshalPacket(packet Packet, data []byte) error {
	if err := packet.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func MarshalPacket(packet Packet) ([]byte, error) {
	buffer := [DefaultMaxPacketSize]byte{}
	stream, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := packet.Serialize(stream); err != nil {
		return nil, err
	}
	stream.Flush()

	return buffer[:stream.GetBytesProcessed()], nil
}

type ServerInitRequestPacket struct {
	Version        SDKVersion
	BuyerID        uint64
	DatacenterID   uint64
	RequestID      uint64
	DatacenterName string
}

func (packet *ServerInitRequestPacket) Serialize(stream encoding.Stream) error {
	versionMajor := uint32(packet.Version.Major)
	versionMinor := uint32(packet.Version.Minor)
	versionPatch := uint32(packet.Version.Patch)
	stream.SerializeBits(&versionMajor, 8)
	stream.SerializeBits(&versionMinor, 8)
	stream.SerializeBits(&versionPatch, 8)
	packet.Version = SDKVersion{int32(versionMajor), int32(versionMinor), int32(versionPatch)}
	stream.SerializeUint64(&packet.BuyerID)
	stream.SerializeUint64(&packet.DatacenterID)
	stream.SerializeUint64(&packet.RequestID)
	stream.SerializeString(&packet.DatacenterName, MaxDatacenterNameLength)
	return stream.Error()
}

type ServerInitResponsePacket struct {
	RequestID uint64
	Response  uint32
}

func (packet *ServerInitResponsePacket) Serialize(stream encoding.Stream) error {
	stream.SerializeUint64(&packet.RequestID)
	stream.SerializeBits(&packet.Response, 8)
	return stream.Error()
}

type ServerUpdatePacket struct {
	Version       SDKVersion
	BuyerID       uint64
	DatacenterID  uint64
	NumSessions   uint32
	ServerAddress net.UDPAddr
}

func (packet *ServerUpdatePacket) Serialize(stream encoding.Stream) error {
	versionMajor := uint32(packet.Version.Major)
	versionMinor := uint32(packet.Version.Minor)
	versionPatch := uint32(packet.Version.Patch)
	stream.SerializeBits(&versionMajor, 8)
	stream.SerializeBits(&versionMinor, 8)
	stream.SerializeBits(&versionPatch, 8)
	packet.Version = SDKVersion{int32(versionMajor), int32(versionMinor), int32(versionPatch)}
	stream.SerializeUint64(&packet.BuyerID)
	stream.SerializeUint64(&packet.DatacenterID)
	stream.SerializeUint32(&packet.NumSessions)
	stream.SerializeAddress(&packet.ServerAddress)
	return stream.Error()
}

type SessionUpdatePacket struct {
	Version                         SDKVersion
	BuyerID                         uint64
	DatacenterID                    uint64
	SessionID                       uint64
	SliceNumber                     uint32
	RetryNumber                     int32
	SessionDataBytes                int32
	SessionData                     [MaxSessionDataSize]byte
	ClientAddress                   net.UDPAddr
	ServerAddress                   net.UDPAddr
	ClientRoutePublicKey            []byte
	ServerRoutePublicKey            []byte
	UserHash                        uint64
	PlatformType                    int32
	ConnectionType                  int32
	Next                            bool
	Committed                       bool
	Reported                        bool
	FallbackToDirect                bool
	ClientBandwidthOverLimit        bool
	ServerBandwidthOverLimit        bool
	ClientPingTimedOut              bool
	NumTags                         int32
	Tags                            [MaxTags]uint64
	Flags                           uint32
	UserFlags                       uint64
	DirectRTT                       float32
	DirectJitter                    float32
	DirectPacketLoss                float32
	NextRTT                         float32
	NextJitter                      float32
	NextPacketLoss                  float32
	NumNearRelays                   int32
	NearRelayIDs                    []uint64
	NearRelayRTT                    []int32
	NearRelayJitter                 []int32
	NearRelayPacketLoss             []int32
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

func (packet *SessionUpdatePacket) Serialize(stream encoding.Stream) error {

	versionMajor := uint32(packet.Version.Major)
	versionMinor := uint32(packet.Version.Minor)
	versionPatch := uint32(packet.Version.Patch)

	stream.SerializeBits(&versionMajor, 8)
	stream.SerializeBits(&versionMinor, 8)
	stream.SerializeBits(&versionPatch, 8)

	packet.Version = SDKVersion{int32(versionMajor), int32(versionMinor), int32(versionPatch)}

	stream.SerializeUint64(&packet.BuyerID)
	stream.SerializeUint64(&packet.DatacenterID)
	stream.SerializeUint64(&packet.SessionID)
	stream.SerializeUint32(&packet.SliceNumber)
	stream.SerializeInteger(&packet.RetryNumber, 0, MaxSessionUpdateRetries)

	stream.SerializeInteger(&packet.SessionDataBytes, 0, MaxSessionDataSize)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
	}

	stream.SerializeAddress(&packet.ClientAddress)

	stream.SerializeAddress(&packet.ServerAddress)

	if stream.IsReading() {
		packet.ClientRoutePublicKey = make([]byte, crypto.KeySize)
		packet.ServerRoutePublicKey = make([]byte, crypto.KeySize)
	}

	stream.SerializeBytes(packet.ClientRoutePublicKey)

	stream.SerializeBytes(packet.ServerRoutePublicKey)

	stream.SerializeUint64(&packet.UserHash)

	if core.ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 4, 0, 5) {
		if core.ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 4, 0, 10) {
			stream.SerializeInteger(&packet.PlatformType, PlatformTypeUnknown, PlatformTypeMax_410)
		} else {
			stream.SerializeInteger(&packet.PlatformType, PlatformTypeUnknown, PlatformTypeMax_405)
		}
	} else {
		stream.SerializeInteger(&packet.PlatformType, PlatformTypeUnknown, PlatformTypeMax_404)
	}

	stream.SerializeInteger(&packet.ConnectionType, ConnectionTypeUnknown, ConnectionTypeMax)

	stream.SerializeBool(&packet.Next)
	stream.SerializeBool(&packet.Committed)
	stream.SerializeBool(&packet.Reported)
	stream.SerializeBool(&packet.FallbackToDirect)
	stream.SerializeBool(&packet.ClientBandwidthOverLimit)
	stream.SerializeBool(&packet.ServerBandwidthOverLimit)
	if core.ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 4, 0, 2) {
		stream.SerializeBool(&packet.ClientPingTimedOut)
	}

	hasTags := stream.IsWriting() && packet.NumTags > 0
	hasFlags := stream.IsWriting() && packet.Flags != 0
	hasUserFlags := stream.IsWriting() && packet.UserFlags != 0
	hasLostPackets := stream.IsWriting() && (packet.PacketsLostClientToServer+packet.PacketsLostServerToClient) > 0
	hasOutOfOrderPackets := stream.IsWriting() && (packet.PacketsOutOfOrderClientToServer+packet.PacketsOutOfOrderServerToClient) > 0

	stream.SerializeBool(&hasTags)
	stream.SerializeBool(&hasFlags)
	stream.SerializeBool(&hasUserFlags)
	stream.SerializeBool(&hasLostPackets)
	stream.SerializeBool(&hasOutOfOrderPackets)

	if hasTags {
		if core.ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 4, 0, 3) {
			// multiple tags (SDK 4.0.3 and above)
			stream.SerializeInteger(&packet.NumTags, 0, MaxTags)
			for i := 0; i < int(packet.NumTags); i++ {
				stream.SerializeUint64(&packet.Tags[i])
			}
		} else {
			// single tag (< SDK 4.0.3)
			stream.SerializeUint64(&packet.Tags[0])
			if stream.IsWriting() {
				packet.NumTags = 1
			}
		}
	}

	if hasFlags {
		if core.ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 4, 0, 1) {
			// flag added in SDK 4.0.1 for fallback to new direct reason
			stream.SerializeBits(&packet.Flags, FallbackFlagsCount_401)
		} else {
			stream.SerializeBits(&packet.Flags, FallbackFlagsCount_400)
		}
	}

	if hasUserFlags {
		stream.SerializeUint64(&packet.UserFlags)
	}

	stream.SerializeFloat32(&packet.DirectRTT)
	stream.SerializeFloat32(&packet.DirectJitter)
	stream.SerializeFloat32(&packet.DirectPacketLoss)

	if packet.Next {
		stream.SerializeFloat32(&packet.NextRTT)
		stream.SerializeFloat32(&packet.NextJitter)
		stream.SerializeFloat32(&packet.NextPacketLoss)
	}

	stream.SerializeInteger(&packet.NumNearRelays, 0, core.MaxNearRelays)

	if stream.IsReading() {
		packet.NearRelayIDs = make([]uint64, packet.NumNearRelays)
		packet.NearRelayRTT = make([]int32, packet.NumNearRelays)
		packet.NearRelayJitter = make([]int32, packet.NumNearRelays)
		packet.NearRelayPacketLoss = make([]int32, packet.NumNearRelays)
	}

	for i := int32(0); i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIDs[i])
		if core.ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 4, 0, 4) {
			// SDK 4.0.4 optimized transmission of near relay rtt, jitter and packet loss
			stream.SerializeInteger(&packet.NearRelayRTT[i], 0, 255)
			stream.SerializeInteger(&packet.NearRelayJitter[i], 0, 255)
			stream.SerializeInteger(&packet.NearRelayPacketLoss[i], 0, 100)
		} else {
			rtt := float32(packet.NearRelayRTT[i])
			jitter := float32(packet.NearRelayJitter[i])
			packetLoss := float32(packet.NearRelayPacketLoss[i])

			stream.SerializeFloat32(&rtt)
			stream.SerializeFloat32(&jitter)
			stream.SerializeFloat32(&packetLoss)

			packet.NearRelayRTT[i] = int32(math.Ceil(float64(rtt)))
			packet.NearRelayJitter[i] = int32(math.Ceil(float64(jitter)))
			packet.NearRelayPacketLoss[i] = int32(math.Floor(float64(packetLoss + 0.5)))
		}
	}

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

type SessionResponsePacket struct {
	Version            SDKVersion
	SessionID          uint64
	SliceNumber        uint32
	SessionDataBytes   int32
	SessionData        [MaxSessionDataSize]byte
	RouteType          int32
	NearRelaysChanged  bool
	NumNearRelays      int32
	NearRelayIDs       []uint64
	NearRelayAddresses []net.UDPAddr
	NumTokens          int32
	Tokens             []byte
	Multipath          bool
	Committed          bool
	HasDebug           bool
	Debug              string
	ExcludeNearRelays  bool
	NearRelayExcluded  [core.MaxNearRelays]bool
	HighFrequencyPings bool
}

func (packet *SessionResponsePacket) Serialize(stream encoding.Stream) error {

	stream.SerializeUint64(&packet.SessionID)

	stream.SerializeUint32(&packet.SliceNumber)

	stream.SerializeInteger(&packet.SessionDataBytes, 0, MaxSessionDataSize)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
	}

	stream.SerializeInteger(&packet.RouteType, 0, routing.RouteTypeContinue)

	nearRelaysChanged := true
	if core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 4, 0, 4) {
		stream.SerializeBool(&packet.NearRelaysChanged)
		nearRelaysChanged = packet.NearRelaysChanged
	}

	if nearRelaysChanged {
		stream.SerializeInteger(&packet.NumNearRelays, 0, core.MaxNearRelays)
		if stream.IsReading() {
			packet.NearRelayIDs = make([]uint64, packet.NumNearRelays)
			packet.NearRelayAddresses = make([]net.UDPAddr, packet.NumNearRelays)
		}
		for i := int32(0); i < packet.NumNearRelays; i++ {
			stream.SerializeUint64(&packet.NearRelayIDs[i])
			stream.SerializeAddress(&packet.NearRelayAddresses[i])
		}
	}

	if packet.RouteType != routing.RouteTypeDirect {
		stream.SerializeBool(&packet.Multipath)
		stream.SerializeBool(&packet.Committed)
		stream.SerializeInteger(&packet.NumTokens, 0, MaxTokens)
	}

	if packet.RouteType == routing.RouteTypeNew {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*routing.EncryptedNextRouteTokenSize)
		}
		stream.SerializeBytes(packet.Tokens)
	}

	if packet.RouteType == routing.RouteTypeContinue {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*routing.EncryptedContinueRouteTokenSize)
		}
		stream.SerializeBytes(packet.Tokens)
	}

	if core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 4, 0, 4) {
		stream.SerializeBool(&packet.HasDebug)
		stream.SerializeString(&packet.Debug, NextMaxSessionDebug)
	}

	if core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 4, 0, 5) {
		stream.SerializeBool(&packet.ExcludeNearRelays)
		if packet.ExcludeNearRelays {
			for i := range packet.NearRelayExcluded {
				stream.SerializeBool(&packet.NearRelayExcluded[i])
			}
		}
	}

	if core.ProtocolVersionAtLeast(uint32(packet.Version.Major), uint32(packet.Version.Minor), uint32(packet.Version.Patch), 4, 0, 6) {
		stream.SerializeBool(&packet.HighFrequencyPings)
	}

	return stream.Error()
}

type SessionData struct {
	Version                       uint32
	SessionID                     uint64
	SessionVersion                uint32
	SliceNumber                   uint32
	ExpireTimestamp               uint64
	Initial                       bool
	Location                      routing.Location
	RouteChanged                  bool
	RouteNumRelays                int32
	RouteCost                     int32
	RouteRelayIDs                 [core.MaxRelaysPerRoute]uint64
	RouteState                    core.RouteState
	EverOnNext                    bool
	FellBackToDirect              bool
	PrevPacketsSentClientToServer uint64
	PrevPacketsSentServerToClient uint64
	PrevPacketsLostClientToServer uint64
	PrevPacketsLostServerToClient uint64
	HoldNearRelays                bool
	HoldNearRelayRTT              [core.MaxNearRelays]int32
	WroteSummary                  bool
}

func UnmarshalSessionData(sessionData *SessionData, data []byte) error {
	if err := sessionData.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func MarshalSessionData(sessionData *SessionData) ([]byte, error) {
	// If we never got around to setting the session data version, set it here so that we can serialize it properly
	if sessionData.Version == 0 {
		sessionData.Version = SessionDataVersion
	}

	buffer := [DefaultMaxPacketSize]byte{}

	stream, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := sessionData.Serialize(stream); err != nil {
		return nil, err
	}
	stream.Flush()

	return buffer[:stream.GetBytesProcessed()], nil
}

func (sessionData *SessionData) Serialize(stream encoding.Stream) error {

	// IMPORTANT: DO NOT EVER CHANGE CODE IN THIS FUNCTION BELOW HERE.
	// CHANGING CODE BELOW HERE *WILL* BREAK PRODUCTION!!!!

	stream.SerializeBits(&sessionData.Version, 8)

	if sessionData.Version < 8 {
		return errors.New("session data is too old")
	}

	stream.SerializeUint64(&sessionData.SessionID)
	stream.SerializeBits(&sessionData.SessionVersion, 8)

	stream.SerializeUint32(&sessionData.SliceNumber)

	stream.SerializeUint64(&sessionData.ExpireTimestamp)

	stream.SerializeBool(&sessionData.Initial)

	locationSize := uint32(sessionData.Location.Size())
	stream.SerializeUint32(&locationSize)
	if stream.IsReading() {
		locationBytes := make([]byte, locationSize)
		stream.SerializeBytes(locationBytes)
		if err := sessionData.Location.UnmarshalBinary(locationBytes); err != nil {
			return err
		}
	} else {
		locationBytes, err := sessionData.Location.MarshalBinary()
		if err != nil {
			return err
		}
		stream.SerializeBytes(locationBytes)
	}

	stream.SerializeBool(&sessionData.RouteChanged)

	hasRoute := sessionData.RouteNumRelays > 0
	stream.SerializeBool(&hasRoute)

	stream.SerializeInteger(&sessionData.RouteCost, 0, routing.InvalidRouteValue)

	if hasRoute {
		stream.SerializeInteger(&sessionData.RouteNumRelays, 0, core.MaxRelaysPerRoute)
		for i := int32(0); i < sessionData.RouteNumRelays; i++ {
			stream.SerializeUint64(&sessionData.RouteRelayIDs[i])
		}
	}

	stream.SerializeUint64(&sessionData.RouteState.UserID)
	stream.SerializeBool(&sessionData.RouteState.Next)
	stream.SerializeBool(&sessionData.RouteState.Veto)
	stream.SerializeBool(&sessionData.RouteState.Banned)
	stream.SerializeBool(&sessionData.RouteState.Disabled)
	stream.SerializeBool(&sessionData.RouteState.NotSelected)
	stream.SerializeBool(&sessionData.RouteState.ABTest)
	stream.SerializeBool(&sessionData.RouteState.A)
	stream.SerializeBool(&sessionData.RouteState.B)
	stream.SerializeBool(&sessionData.RouteState.ForcedNext)
	stream.SerializeBool(&sessionData.RouteState.ReduceLatency)
	stream.SerializeBool(&sessionData.RouteState.ReducePacketLoss)
	stream.SerializeBool(&sessionData.RouteState.ProMode)
	stream.SerializeBool(&sessionData.RouteState.Multipath)
	stream.SerializeBool(&sessionData.RouteState.Committed)
	stream.SerializeBool(&sessionData.RouteState.CommitVeto)
	stream.SerializeInteger(&sessionData.RouteState.CommitCounter, 0, 4)
	stream.SerializeBool(&sessionData.RouteState.LatencyWorse)
	stream.SerializeBool(&sessionData.RouteState.MultipathOverload)
	stream.SerializeBool(&sessionData.RouteState.NoRoute)
	stream.SerializeBool(&sessionData.RouteState.NextLatencyTooHigh)
	stream.SerializeBool(&sessionData.RouteState.Mispredict)
	stream.SerializeBool(&sessionData.EverOnNext)
	stream.SerializeBool(&sessionData.FellBackToDirect)

	stream.SerializeInteger(&sessionData.RouteState.NumNearRelays, 0, core.MaxNearRelays)

	for i := int32(0); i < sessionData.RouteState.NumNearRelays; i++ {
		stream.SerializeInteger(&sessionData.RouteState.NearRelayRTT[i], 0, 255)
		stream.SerializeInteger(&sessionData.RouteState.NearRelayJitter[i], 0, 255)
		nearRelayPLHistory := int32(sessionData.RouteState.NearRelayPLHistory[i])
		stream.SerializeInteger(&nearRelayPLHistory, 0, 255)
		sessionData.RouteState.NearRelayPLHistory[i] = uint32(nearRelayPLHistory)
	}

	directPLHistory := int32(sessionData.RouteState.DirectPLHistory)
	stream.SerializeInteger(&directPLHistory, 0, 255)
	sessionData.RouteState.DirectPLHistory = uint32(directPLHistory)

	stream.SerializeInteger(&sessionData.RouteState.PLHistoryIndex, 0, 7)
	stream.SerializeInteger(&sessionData.RouteState.PLHistorySamples, 0, 8)

	stream.SerializeBool(&sessionData.RouteState.RelayWentAway)
	stream.SerializeBool(&sessionData.RouteState.RouteLost)
	stream.SerializeInteger(&sessionData.RouteState.DirectJitter, 0, 255)

	stream.SerializeUint32(&sessionData.RouteState.DirectPLCount)

	for i := int32(0); i < sessionData.RouteState.NumNearRelays; i++ {
		stream.SerializeUint32(&sessionData.RouteState.NearRelayPLCount[i])
	}

	stream.SerializeBool(&sessionData.RouteState.LackOfDiversity)

	stream.SerializeBits(&sessionData.RouteState.MispredictCounter, 2)

	stream.SerializeBits(&sessionData.RouteState.LatencyWorseCounter, 2)

	if sessionData.Version >= 9 {
		stream.SerializeBool(&sessionData.RouteState.MultipathRestricted)

		stream.SerializeUint64(&sessionData.PrevPacketsSentClientToServer)
		stream.SerializeUint64(&sessionData.PrevPacketsSentServerToClient)
		stream.SerializeUint64(&sessionData.PrevPacketsLostClientToServer)
		stream.SerializeUint64(&sessionData.PrevPacketsLostServerToClient)
	}

	if sessionData.Version >= 10 {
		stream.SerializeBool(&sessionData.RouteState.LocationVeto)
	}

	if sessionData.Version >= 11 {
		stream.SerializeBool(&sessionData.HoldNearRelays)
		if sessionData.HoldNearRelays {
			for i := 0; i < core.MaxNearRelays; i++ {
				stream.SerializeInteger(&sessionData.HoldNearRelayRTT[i], 0, 255)
			}
		}
	}

	// IMPORTANT: Remove this in the future. We need this to stem fall back to directs 05-27-21
	// Done

	if sessionData.Version >= 12 {
		stream.SerializeInteger(&sessionData.RouteState.PLSustainedCounter, 0, 3)
	}

	if sessionData.Version >= 13 {
		stream.SerializeBool(&sessionData.WroteSummary)
	}

	// IMPORTANT: ADD NEW FIELDS BELOW HERE ONLY.

	// >>> new fields go here <<<

	// IMPORTANT: ADD NEW FIELDS ABOVE HERE ONLY.
	// AFTER YOU ADD NEW FIELDS, UPDATE THE COMMENTS SO FIELDS
	// MAY ONLY BE ADDED *AFTER* YOUR NEW FIELDS.
	// FAILING TO FOLLOW THESE INSRUCTIONS WILL BREAK PRODUCTION!!!!

	return stream.Error()
}
